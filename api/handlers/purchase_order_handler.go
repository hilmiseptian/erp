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

type PurchaseOrderHandler struct {
	Service *services.PurchaseOrderService
}

func NewPurchaseOrderHandler(service *services.PurchaseOrderService) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{Service: service}
}

type purchaseOrderPayload struct {
	PRNumber string                   `json:"ref_num"`
	PODate   string                   `json:"po_date"`
	DueDate  string                   `json:"due_date"`
	Lines    []purchaseOrderLineInput `json:"lines"`
}

type purchaseOrderLineInput struct {
	ItemID uint64 `json:"item_id"`
	Qty    int64  `json:"qty"`
	Price  string `json:"price"`
}

func (p purchaseOrderPayload) model() models.PurchaseOrder {
	po := models.PurchaseOrder{PRNumber: p.PRNumber, PODate: p.PODate, DueDate: p.DueDate, Lines: make([]models.PurchaseOrderLine, 0, len(p.Lines))}
	for _, line := range p.Lines {
		po.Lines = append(po.Lines, models.PurchaseOrderLine{ItemID: line.ItemID, Qty: line.Qty, Price: line.Price})
	}
	return po
}

// Sources lists purchase requests that still have quantities available to order.
// @Summary List purchase requests available for ordering
// @Tags Purchase Orders
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/purchase-orders/purchase-requests [get]
func (h *PurchaseOrderHandler) Sources(c *gin.Context) {
	data, err := h.Service.Sources(poUser(c).CompanyID)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to load purchase request options"})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

// Source returns remaining quantities for a purchase request.
// @Summary Get purchase request quantities available for ordering
// @Tags Purchase Orders
// @Security BearerAuth
// @Param pr_number path string true "Source PR number"
// @Success 200 {object} map[string]interface{}
// @Router /api/purchase-orders/purchase-requests/{pr_number} [get]
func (h *PurchaseOrderHandler) Source(c *gin.Context) {
	data, err := h.Service.Source(poUser(c).CompanyID, c.Param("pr_number"))
	if err != nil {
		poError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": data})
}

// List purchase orders for the authenticated company.
// @Summary List purchase orders
// @Tags Purchase Orders
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size (max 100)"
// @Param search query string false "PO number, PR number, supplier, or branch"
// @Param status query string false "draft, ordered, partially_received, received, or cancelled"
// @Success 200 {object} map[string]interface{}
// @Router /api/purchase-orders [get]
func (h *PurchaseOrderHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	status := c.Query("status")
	switch status {
	case "", "draft", "ordered", "partially_received", "received", "cancelled":
	default:
		c.JSON(400, gin.H{"message": "Invalid status filter"})
		return
	}
	list, total, err := h.Service.List(poUser(c).CompanyID, status, c.Query("search"), limit, (page-1)*limit)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to retrieve purchase orders"})
		return
	}
	c.JSON(200, gin.H{"data": list, "pagination": gin.H{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}

// Get one purchase order with its detail lines.
// @Summary Get purchase order
// @Tags Purchase Orders
// @Security BearerAuth
// @Param id path int true "Purchase order ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/purchase-orders/{id} [get]
func (h *PurchaseOrderHandler) Get(c *gin.Context) {
	id, ok := poID(c)
	if !ok {
		return
	}
	data, err := h.Service.Get(poUser(c).CompanyID, id)
	if err != nil {
		poError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": data})
}

// Create a purchase order from one purchase request.
// @Summary Create purchase order
// @Tags Purchase Orders
// @Security BearerAuth
// @Param request body purchaseOrderPayload true "Purchase order details"
// @Success 201 {object} map[string]interface{}
// @Failure 422 {object} map[string]string
// @Router /api/purchase-orders [post]
func (h *PurchaseOrderHandler) Create(c *gin.Context) {
	var payload purchaseOrderPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"message": "Invalid request body"})
		return
	}
	user := poUser(c)
	id, err := h.Service.Create(user.CompanyID, user.ID, payload.model())
	if err != nil {
		poError(c, err)
		return
	}
	data, err := h.Service.Get(user.CompanyID, id)
	if err != nil {
		poError(c, err)
		return
	}
	c.JSON(201, gin.H{"message": "Purchase order created successfully", "data": data})
}

// Cancel an ordered purchase order that has not been received.
// @Summary Cancel purchase order
// @Tags Purchase Orders
// @Security BearerAuth
// @Param id path int true "Purchase order ID"
// @Success 200 {object} map[string]string
// @Router /api/purchase-orders/{id}/cancel [patch]
func (h *PurchaseOrderHandler) Cancel(c *gin.Context) {
	id, ok := poID(c)
	if !ok {
		return
	}
	user := poUser(c)
	if err := h.Service.Cancel(user.CompanyID, id, user.ID); err != nil {
		poError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Purchase order cancelled successfully"})
}

func poUser(c *gin.Context) middleware.AuthUser {
	value, _ := c.Get("authUser")
	user, _ := value.(middleware.AuthUser)
	return user
}

func poID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(400, gin.H{"message": "Invalid purchase order ID"})
		return 0, false
	}
	return id, true
}

func poError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrPONotFound):
		c.JSON(404, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrPOInvalid):
		c.JSON(422, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrPOLocked), errors.Is(err, repositories.ErrPurchaseOrderSource):
		c.JSON(409, gin.H{"message": err.Error()})
	default:
		_ = c.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}
