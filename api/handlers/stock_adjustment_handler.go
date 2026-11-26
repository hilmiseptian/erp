package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"erp-api/middleware"
	"erp-api/models"
	"erp-api/repositories"
	"erp-api/services"

	"github.com/gin-gonic/gin"
)

type StockAdjustmentHandler struct {
	Service *services.StockAdjustmentService
}

func NewStockAdjustmentHandler(service *services.StockAdjustmentService) *StockAdjustmentHandler {
	return &StockAdjustmentHandler{Service: service}
}

type stockAdjustmentPayload struct {
	BranchID       uint64                       `json:"branch_id"`
	AdjustmentDate string                       `json:"adjustment_date"`
	AdjustmentType string                       `json:"adjustment_type"`
	Reason         *string                      `json:"reason"`
	Lines          []stockAdjustmentLinePayload `json:"lines"`
}

type stockAdjustmentLinePayload struct {
	ItemID    uint64 `json:"item_id"`
	ActualQty *int64 `json:"actual_qty"`
}

func (p stockAdjustmentPayload) model() models.StockAdjustment {
	adjustment := models.StockAdjustment{
		BranchID:       p.BranchID,
		AdjustmentDate: p.AdjustmentDate,
		AdjustmentType: p.AdjustmentType,
		Reason:         p.Reason,
		Lines:          make([]models.StockAdjustmentLine, 0, len(p.Lines)),
	}
	for _, line := range p.Lines {
		actualQty := int64(-1)
		if line.ActualQty != nil {
			actualQty = *line.ActualQty
		}
		adjustment.Lines = append(adjustment.Lines, models.StockAdjustmentLine{ItemID: line.ItemID, ActualQty: actualQty})
	}
	return adjustment
}

// List returns stock adjustment documents for the authenticated company.
// @Summary List stock adjustments
// @Tags Stock Adjustments
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size (max 100)"
// @Param branch_id query int false "Branch ID"
// @Param status query string false "draft, posted, or cancelled"
// @Param adjustment_type query string false "opening_balance or adjustment"
// @Param search query string false "Document number, branch, or reason"
// @Success 200 {object} map[string]interface{}
// @Router /api/stock-adjustments [get]
func (h *StockAdjustmentHandler) List(c *gin.Context) {
	user := stockAdjustmentUser(c)
	page, limit := pageLimit(c)
	branchID := uint64(0)
	if raw := c.Query("branch_id"); raw != "" {
		parsed, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || parsed == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid branch ID"})
			return
		}
		branchID = parsed
	}
	status := c.Query("status")
	switch status {
	case "", "draft", "posted", "cancelled":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid status filter"})
		return
	}
	adjustmentType := c.Query("adjustment_type")
	switch adjustmentType {
	case "", "opening_balance", "adjustment":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid adjustment type filter"})
		return
	}
	list, total, err := h.Service.List(user.CompanyID, branchID, status, adjustmentType, c.Query("search"), limit, (page-1)*limit)
	if err != nil {
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to retrieve stock adjustments"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "pagination": gin.H{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}

// Get returns one stock adjustment with its item lines.
// @Summary Get stock adjustment
// @Tags Stock Adjustments
// @Security BearerAuth
// @Param id path int true "Stock adjustment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /api/stock-adjustments/{id} [get]
func (h *StockAdjustmentHandler) Get(c *gin.Context) {
	id, ok := stockAdjustmentID(c)
	if !ok {
		return
	}
	adjustment, err := h.Service.Get(stockAdjustmentUser(c).CompanyID, id)
	if err != nil {
		stockAdjustmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": adjustment})
}

// Create saves a stock adjustment draft and snapshots the current branch stock.
// @Summary Create stock adjustment draft
// @Tags Stock Adjustments
// @Security BearerAuth
// @Param request body stockAdjustmentPayload true "Stock adjustment details"
// @Success 201 {object} map[string]interface{}
// @Failure 422 {object} map[string]string
// @Router /api/stock-adjustments [post]
func (h *StockAdjustmentHandler) Create(c *gin.Context) {
	var payload stockAdjustmentPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	user := stockAdjustmentUser(c)
	id, err := h.Service.Create(user.CompanyID, user.ID, payload.model())
	if err != nil {
		stockAdjustmentError(c, err)
		return
	}
	adjustment, err := h.Service.Get(user.CompanyID, id)
	if err != nil {
		stockAdjustmentError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Stock adjustment draft created successfully", "data": adjustment})
}

// Update changes a draft and refreshes its system quantity snapshots.
// @Summary Update stock adjustment draft
// @Tags Stock Adjustments
// @Security BearerAuth
// @Param id path int true "Stock adjustment ID"
// @Param request body stockAdjustmentPayload true "Stock adjustment details"
// @Success 200 {object} map[string]interface{}
// @Router /api/stock-adjustments/{id} [put]
func (h *StockAdjustmentHandler) Update(c *gin.Context) {
	id, ok := stockAdjustmentID(c)
	if !ok {
		return
	}
	var payload stockAdjustmentPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	user := stockAdjustmentUser(c)
	if err := h.Service.Update(user.CompanyID, user.ID, id, payload.model()); err != nil {
		stockAdjustmentError(c, err)
		return
	}
	adjustment, err := h.Service.Get(user.CompanyID, id)
	if err != nil {
		stockAdjustmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stock adjustment draft updated successfully", "data": adjustment})
}

// Post applies adjustment differences to the branch stock and stock card atomically.
// @Summary Post a stock adjustment
// @Tags Stock Adjustments
// @Security BearerAuth
// @Param id path int true "Stock adjustment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 409 {object} map[string]string
// @Router /api/stock-adjustments/{id}/post [post]
func (h *StockAdjustmentHandler) Post(c *gin.Context) {
	id, ok := stockAdjustmentID(c)
	if !ok {
		return
	}
	user := stockAdjustmentUser(c)
	if err := h.Service.Post(user.CompanyID, user.ID, id); err != nil {
		stockAdjustmentError(c, err)
		return
	}
	adjustment, err := h.Service.Get(user.CompanyID, id)
	if err != nil {
		stockAdjustmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stock adjustment posted successfully", "data": adjustment})
}

// Cancel a draft stock adjustment.
// @Summary Cancel stock adjustment draft
// @Tags Stock Adjustments
// @Security BearerAuth
// @Param id path int true "Stock adjustment ID"
// @Success 200 {object} map[string]string
// @Router /api/stock-adjustments/{id}/cancel [patch]
func (h *StockAdjustmentHandler) Cancel(c *gin.Context) {
	id, ok := stockAdjustmentID(c)
	if !ok {
		return
	}
	user := stockAdjustmentUser(c)
	if err := h.Service.Cancel(user.CompanyID, user.ID, id); err != nil {
		stockAdjustmentError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stock adjustment cancelled successfully"})
}

func stockAdjustmentUser(c *gin.Context) middleware.AuthUser {
	value, _ := c.Get("authUser")
	user, _ := value.(middleware.AuthUser)
	return user
}

func stockAdjustmentID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid stock adjustment ID"})
		return 0, false
	}
	return id, true
}

func stockAdjustmentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrStockAdjustmentNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrStockAdjustmentInvalid):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
	case errors.Is(err, repositories.ErrStockAdjustmentLocked),
		errors.Is(err, repositories.ErrStockAdjustmentStale),
		errors.Is(err, repositories.ErrStockAdjustmentOpeningBalance),
		errors.Is(err, repositories.ErrStockInsufficient):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	default:
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}
