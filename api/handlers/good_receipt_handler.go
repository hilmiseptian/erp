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

type GoodReceiptHandler struct{ Service *services.GoodReceiptService }

func NewGoodReceiptHandler(service *services.GoodReceiptService) *GoodReceiptHandler {
	return &GoodReceiptHandler{Service: service}
}

type goodReceiptPayload struct {
	PONumber string                 `json:"po_number"`
	GRDate   string                 `json:"gr_date"`
	Lines    []goodReceiptLineInput `json:"lines"`
}
type goodReceiptLineInput struct {
	ItemID uint64 `json:"item_id"`
	Qty    int64  `json:"qty"`
}

func (p goodReceiptPayload) model() models.GoodReceipt {
	gr := models.GoodReceipt{PONumber: p.PONumber, GRDate: p.GRDate, Lines: make([]models.GoodReceiptLine, 0, len(p.Lines))}
	for _, l := range p.Lines {
		gr.Lines = append(gr.Lines, models.GoodReceiptLine{ItemID: l.ItemID, Qty: l.Qty})
	}
	return gr
}

// @Summary List purchase orders available for receipt
// @Tags Good Receipts
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/good-receipts/purchase-orders [get]
func (h *GoodReceiptHandler) Sources(c *gin.Context) {
	data, err := h.Service.Sources(grUser(c).CompanyID)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to load purchase orders"})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

// @Summary Get remaining quantities for a purchase order
// @Tags Good Receipts
// @Security BearerAuth
// @Param po_number path string true "Purchase order number"
// @Success 200 {object} map[string]interface{}
// @Router /api/good-receipts/purchase-orders/{po_number} [get]
func (h *GoodReceiptHandler) Source(c *gin.Context) {
	data, err := h.Service.Source(grUser(c).CompanyID, c.Param("po_number"))
	if err != nil {
		grError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": data})
}

// @Summary List good receipts
// @Tags Good Receipts
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param status query string false "draft, posted, or cancelled"
// @Success 200 {object} map[string]interface{}
// @Router /api/good-receipts [get]
func (h *GoodReceiptHandler) List(c *gin.Context) {
	page, limit := grPageLimit(c)
	status := c.Query("status")
	switch status {
	case "", "draft", "posted", "cancelled":
	default:
		c.JSON(400, gin.H{"message": "Invalid status filter"})
		return
	}
	list, total, err := h.Service.List(grUser(c).CompanyID, status, c.Query("search"), limit, (page-1)*limit)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to retrieve good receipts"})
		return
	}
	c.JSON(200, gin.H{"data": list, "pagination": gin.H{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}

// @Summary Get good receipt details
// @Tags Good Receipts
// @Security BearerAuth
// @Param id path int true "Good receipt ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/good-receipts/{id} [get]
func (h *GoodReceiptHandler) Get(c *gin.Context) {
	id, ok := grID(c)
	if !ok {
		return
	}
	gr, err := h.Service.Get(grUser(c).CompanyID, id)
	if err != nil {
		grError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": gr})
}

// @Summary Post a good receipt and add its quantities to branch stock
// @Tags Good Receipts
// @Security BearerAuth
// @Param request body goodReceiptPayload true "Receipt data"
// @Success 201 {object} map[string]interface{}
// @Failure 422 {object} map[string]string
// @Router /api/good-receipts [post]
func (h *GoodReceiptHandler) Create(c *gin.Context) {
	var payload goodReceiptPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"message": "Invalid request body"})
		return
	}
	user := grUser(c)
	id, err := h.Service.Create(user.CompanyID, user.ID, payload.model())
	if err != nil {
		grError(c, err)
		return
	}
	gr, err := h.Service.Get(user.CompanyID, id)
	if err != nil {
		grError(c, err)
		return
	}
	c.JSON(201, gin.H{"message": "Good receipt posted successfully", "data": gr})
}

func grUser(c *gin.Context) middleware.AuthUser {
	v, _ := c.Get("authUser")
	u, _ := v.(middleware.AuthUser)
	return u
}
func grPageLimit(c *gin.Context) (int, int) {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	l, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if p < 1 {
		p = 1
	}
	if l < 1 {
		l = 20
	}
	if l > 100 {
		l = 100
	}
	return p, l
}
func grID(c *gin.Context) (uint64, bool) {
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil || id == 0 {
		c.JSON(400, gin.H{"message": "Invalid good receipt ID"})
		return 0, false
	}
	return id, true
}
func grError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrGRNotFound):
		c.JSON(404, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrGRInvalid):
		c.JSON(422, gin.H{"message": err.Error()})
	case errors.Is(err, repositories.ErrGoodReceiptSource):
		c.JSON(409, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}
