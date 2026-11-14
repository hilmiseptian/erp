package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"erp-api/middleware"
	"erp-api/models"
	"erp-api/services"
	"github.com/gin-gonic/gin"
)

type ItemHandler struct{ Service *services.ItemService }

func NewItemHandler(service *services.ItemService) *ItemHandler {
	return &ItemHandler{Service: service}
}

type itemRequest struct {
	CategoryID   uint64  `json:"category_id"`
	Name         string  `json:"name"`
	SKU          string  `json:"sku"`
	Description  *string `json:"description"`
	CostPrice    string  `json:"cost_price"`
	SellingPrice string  `json:"selling_price"`
	MinStock     int64   `json:"min_stock"`
	Unit         string  `json:"unit"`
	IsActive     *bool   `json:"is_active"`
}

func (r itemRequest) model() models.Item {
	active := true
	if r.IsActive != nil {
		active = *r.IsActive
	}
	return models.Item{CategoryID: r.CategoryID, Name: r.Name, SKU: r.SKU, Description: r.Description, CostPrice: r.CostPrice, SellingPrice: r.SellingPrice, MinStock: r.MinStock, Unit: r.Unit, IsActive: active}
}
func itemCompany(c *gin.Context) uint64 {
	u, ok := c.Get("authUser")
	if !ok {
		return 0
	}
	user, ok := u.(middleware.AuthUser)
	if !ok {
		return 0
	}
	return user.CompanyID
}

// List returns the authenticated company's items.
// @Summary List items
// @Tags Items
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size (max 100)"
// @Param search query string false "Search name, SKU, or description"
// @Param status query string false "active or inactive"
// @Param category_id query int false "Category ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/items [get]
func (h *ItemHandler) List(c *gin.Context) {
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
	categoryID, _ := strconv.ParseUint(c.Query("category_id"), 10, 64)
	items, total, err := h.Service.List(itemCompany(c), c.Query("search"), c.Query("status"), categoryID, limit, (page-1)*limit)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to retrieve items"})
		return
	}
	c.JSON(200, gin.H{"data": items, "pagination": gin.H{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}
// Categories lists active categories available to the company.
// @Summary List item categories
// @Tags Items
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/items/categories [get]
func (h *ItemHandler) Categories(c *gin.Context) {
	items, err := h.Service.Categories(itemCompany(c))
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to retrieve categories"})
		return
	}
	c.JSON(200, gin.H{"data": items})
}
// Get returns one item in the authenticated company.
// @Summary Get item
// @Tags Items
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /api/items/{id} [get]
func (h *ItemHandler) Get(c *gin.Context) {
	id, ok := itemID(c)
	if !ok {
		return
	}
	item, err := h.Service.Get(itemCompany(c), id)
	if err != nil {
		itemError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": item})
}
// Create adds an item with zero opening stock.
// @Summary Create item
// @Tags Items
// @Security BearerAuth
// @Param item body itemRequest true "Item details"
// @Success 201 {object} map[string]interface{}
// @Failure 422 {object} map[string]string
// @Router /api/items [post]
func (h *ItemHandler) Create(c *gin.Context) {
	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid request body"})
		return
	}
	if err := h.Service.Save(itemCompany(c), 0, req.model()); err != nil {
		itemError(c, err)
		return
	}
	c.JSON(201, gin.H{"message": "Item created successfully"})
}
// Update changes item master data without changing stock.
// @Summary Update item
// @Tags Items
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Param item body itemRequest true "Item details"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /api/items/{id} [put]
func (h *ItemHandler) Update(c *gin.Context) {
	id, ok := itemID(c)
	if !ok {
		return
	}
	var req itemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"message": "Invalid request body"})
		return
	}
	if err := h.Service.Save(itemCompany(c), id, req.model()); err != nil {
		itemError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Item updated successfully"})
}
// SetActive changes item availability without deleting its history.
// @Summary Set item status
// @Tags Items
// @Security BearerAuth
// @Param id path int true "Item ID"
// @Param status body object true "Status payload with is_active"
// @Success 200 {object} map[string]string
// @Router /api/items/{id}/status [patch]
func (h *ItemHandler) SetActive(c *gin.Context) {
	id, ok := itemID(c)
	if !ok {
		return
	}
	var req struct {
		IsActive *bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.IsActive == nil {
		c.JSON(400, gin.H{"message": "is_active is required"})
		return
	}
	if err := h.Service.SetActive(itemCompany(c), id, *req.IsActive); err != nil {
		itemError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Item status updated successfully"})
}
func itemID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(400, gin.H{"message": "Invalid item ID"})
		return 0, false
	}
	return id, true
}
func itemError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrItemNotFound):
		c.JSON(404, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrItemDuplicateSKU):
		c.JSON(409, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrItemInvalid):
		c.JSON(422, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}
