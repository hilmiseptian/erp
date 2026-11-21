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

type SalesOrderHandler struct{ Service *services.SalesOrderService }

func NewSalesOrderHandler(service *services.SalesOrderService) *SalesOrderHandler {
	return &SalesOrderHandler{Service: service}
}

type salesOrderPayload struct {
	BranchID        uint64                  `json:"branch_id"`
	CustomerID      uint64                  `json:"customer_id"`
	SODate          string                  `json:"so_date"`
	DeliveryDate    *string                 `json:"delivery_date"`
	DeliveryAddress *string                 `json:"delivery_address"`
	DiscountAmount  string                  `json:"discount_amount"`
	Notes           *string                 `json:"notes"`
	Lines           []salesOrderLinePayload `json:"lines"`
}
type salesOrderLinePayload struct {
	ID             uint64  `json:"id"`
	ItemID         uint64  `json:"item_id"`
	Qty            int64   `json:"qty"`
	DiscountAmount string  `json:"discount_amount"`
	Notes          *string `json:"notes"`
}

func (p salesOrderPayload) model() models.SalesOrder {
	item := models.SalesOrder{BranchID: p.BranchID, CustomerID: p.CustomerID, SODate: p.SODate, DeliveryDate: p.DeliveryDate, DeliveryAddress: p.DeliveryAddress, DiscountAmount: p.DiscountAmount, Notes: p.Notes, Lines: make([]models.SalesOrderLine, 0, len(p.Lines))}
	for _, line := range p.Lines {
		item.Lines = append(item.Lines, models.SalesOrderLine{ID: line.ID, ItemID: line.ItemID, Qty: line.Qty, DiscountAmount: line.DiscountAmount, Notes: line.Notes})
	}
	return item
}

// Branches lists branches available to the authenticated company.
// @Summary List branches for sales orders
// @Tags Sales Orders
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/sales-orders/branches [get]
func (h *SalesOrderHandler) Branches(c *gin.Context) {
	items, err := h.Service.Branches(salesOrderUser(c).CompanyID)
	if err != nil {
		internalServerError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": items})
}

// List sales orders in the authenticated company.
// @Summary List sales orders
// @Tags Sales Orders
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param search query string false "SO number, customer, or branch"
// @Param status query string false "open, partially_delivered, delivered, or cancelled"
// @Success 200 {object} map[string]interface{}
// @Router /api/sales-orders [get]
func (h *SalesOrderHandler) List(c *gin.Context) {
	page, limit := pageLimit(c)
	items, total, err := h.Service.List(salesOrderUser(c).CompanyID, c.Query("status"), c.Query("search"), limit, (page-1)*limit)
	if err != nil {
		salesOrderError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": items, "pagination": gin.H{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}

// Get a sales order in the authenticated company.
// @Summary Get sales order
// @Tags Sales Orders
// @Security BearerAuth
// @Param id path int true "Sales order ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/sales-orders/{id} [get]
func (h *SalesOrderHandler) Get(c *gin.Context) {
	id, ok := salesOrderID(c)
	if !ok {
		return
	}
	item, err := h.Service.Get(salesOrderUser(c).CompanyID, id)
	if err != nil {
		salesOrderError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": item})
}

// Create a B2B sales order. The API uses the current item selling price and does not move stock.
// @Summary Create sales order
// @Tags Sales Orders
// @Security BearerAuth
// @Param request body salesOrderPayload true "Sales order details"
// @Success 201 {object} map[string]interface{}
// @Router /api/sales-orders [post]
func (h *SalesOrderHandler) Create(c *gin.Context) {
	var payload salesOrderPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"message": "Invalid request body"})
		return
	}
	user := salesOrderUser(c)
	id, err := h.Service.Create(user.CompanyID, user.ID, payload.model())
	if err != nil {
		salesOrderError(c, err)
		return
	}
	item, err := h.Service.Get(user.CompanyID, id)
	if err != nil {
		salesOrderError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Sales order created successfully", "data": item})
}

// Update an open sales order. Unit prices remain the original snapshots.
// @Summary Update open sales order
// @Tags Sales Orders
// @Security BearerAuth
// @Param id path int true "Sales order ID"
// @Param request body salesOrderPayload true "Sales order details"
// @Success 200 {object} map[string]interface{}
// @Router /api/sales-orders/{id} [put]
func (h *SalesOrderHandler) Update(c *gin.Context) {
	id, ok := salesOrderID(c)
	if !ok {
		return
	}
	var payload salesOrderPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"message": "Invalid request body"})
		return
	}
	user := salesOrderUser(c)
	if err := h.Service.Update(user.CompanyID, user.ID, id, payload.model()); err != nil {
		salesOrderError(c, err)
		return
	}
	item, err := h.Service.Get(user.CompanyID, id)
	if err != nil {
		salesOrderError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Sales order updated successfully", "data": item})
}

// Cancel an open or partially delivered sales order.
// @Summary Cancel sales order
// @Tags Sales Orders
// @Security BearerAuth
// @Param id path int true "Sales order ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/sales-orders/{id}/cancel [patch]
func (h *SalesOrderHandler) Cancel(c *gin.Context) {
	id, ok := salesOrderID(c)
	if !ok {
		return
	}
	user := salesOrderUser(c)
	if err := h.Service.Cancel(user.CompanyID, user.ID, id); err != nil {
		salesOrderError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Sales order cancelled successfully"})
}

func salesOrderUser(c *gin.Context) middleware.AuthUser {
	value, _ := c.Get("authUser")
	user, _ := value.(middleware.AuthUser)
	return user
}
func salesOrderID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(400, gin.H{"message": "Invalid sales order ID"})
		return 0, false
	}
	return id, true
}
func salesOrderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrSalesOrderNotFound), errors.Is(err, repositories.ErrSalesOrderNotFound):
		c.JSON(404, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrSalesOrderInvalid), errors.Is(err, repositories.ErrSalesOrderInvalid):
		c.JSON(422, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrSalesOrderLocked), errors.Is(err, repositories.ErrSalesOrderLocked):
		c.JSON(409, gin.H{"message": err.Error()})
	default:
		internalServerError(c, err)
	}
}
