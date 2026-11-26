package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const maxLoggedRequestPayloadBytes = 32 * 1024

// SlogRequestLogger writes one structured access record per request. It includes
// a sanitized, size-limited JSON payload only when the response is a server error.
func SlogRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := newRequestID()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		payload := captureJSONPayload(c.Request)
		startedAt := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		attrs := []any{
			"request_id", requestID,
			"method", c.Request.Method,
			"route", route,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(startedAt).Milliseconds(),
		}
		if value, exists := c.Get("authUser"); exists {
			if user, ok := value.(AuthUser); ok {
				attrs = append(attrs, "user_id", user.ID, "company_id", user.CompanyID)
			}
		}

		if c.Writer.Status() >= http.StatusInternalServerError {
			if errText := c.Errors.String(); errText != "" {
				attrs = append(attrs, "error", errText)
			}
			if payload != nil {
				attrs = append(attrs, "payload", payload)
			}
			slog.ErrorContext(c.Request.Context(), "HTTP request failed", attrs...)
			return
		}

		slog.InfoContext(c.Request.Context(), "HTTP request completed", attrs...)
	}
}

func newRequestID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err == nil {
		return hex.EncodeToString(value)
	}
	return time.Now().UTC().Format("20060102T150405.000000000")
}

func captureJSONPayload(request *http.Request) any {
	if request.Body == nil || !strings.Contains(strings.ToLower(request.Header.Get("Content-Type")), "json") {
		return nil
	}

	originalBody := request.Body
	raw, err := io.ReadAll(io.LimitReader(originalBody, maxLoggedRequestPayloadBytes+1))
	request.Body = io.NopCloser(io.MultiReader(bytes.NewReader(raw), originalBody))
	if err != nil {
		return map[string]any{"unavailable": "request body could not be read for logging"}
	}
	if len(raw) > maxLoggedRequestPayloadBytes {
		return map[string]any{
			"truncated":   true,
			"size_limit":  maxLoggedRequestPayloadBytes,
			"body_logged": false,
		}
	}

	var payload any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return map[string]any{"unavailable": "request body is not valid JSON"}
	}
	return redactRequestPayload(payload)
}

func redactRequestPayload(value any) any {
	switch value := value.(type) {
	case map[string]any:
		redacted := make(map[string]any, len(value))
		for key, item := range value {
			if sensitivePayloadKey(key) {
				redacted[key] = "[REDACTED]"
				continue
			}
			redacted[key] = redactRequestPayload(item)
		}
		return redacted
	case []any:
		redacted := make([]any, len(value))
		for index, item := range value {
			redacted[index] = redactRequestPayload(item)
		}
		return redacted
	default:
		return value
	}
}

func sensitivePayloadKey(key string) bool {
	normalized := strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(key))
	for _, fragment := range []string{
		"password", "passwd", "token", "secret", "authorization", "apikey",
		"email", "phone", "address", "taxnumber",
	} {
		if strings.Contains(normalized, fragment) {
			return true
		}
	}
	return false
}
