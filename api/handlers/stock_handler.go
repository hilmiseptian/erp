package handlers

import (
	"strconv"

	"erp-api/middleware"
	"erp-api/services"
	"github.com/gin-gonic/gin"
)

type StockHandler struct{ Service *services.StockService }

func NewStockHandler(service *services.StockService) *StockHandler {
	return &StockHandler{Service: service}
}

// Branches lists branches available to the authenticated company.
// @Summary List branches for stock lookup
// @Tags Stock
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/stocks/branches [get]
func (h *StockHandler) Branches(c *gin.Context) {
	data, err := h.Service.Branches(stockUser(c).CompanyID)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to load branches"})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

// List returns current item balances for one branch.
// @Summary List item stock balances for a branch
// @Tags Stock
// @Security BearerAuth
// @Param branch_id query int true "Branch ID"
// @Param search query string false "Item name or SKU"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /api/stocks [get]
func (h *StockHandler) List(c *gin.Context) {
	branchID, err := strconv.ParseUint(c.Query("branch_id"), 10, 64)
	if err != nil || branchID == 0 {
		c.JSON(400, gin.H{"message": "branch_id is required"})
		return
	}
	page, limit := pageLimit(c)
	list, total, err := h.Service.List(stockUser(c).CompanyID, branchID, c.Query("search"), limit, (page-1)*limit)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to retrieve stock"})
		return
	}
	c.JSON(200, gin.H{"data": list, "pagination": gin.H{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}

// Cards returns stock movements for one item and branch.
// @Summary List stock movements for a branch item
// @Tags Stock
// @Security BearerAuth
// @Param branch_id path int true "Branch ID"
// @Param item_id path int true "Item ID"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /api/stocks/{branch_id}/{item_id}/cards [get]
func (h *StockHandler) Cards(c *gin.Context) {
	branchID, branchErr := strconv.ParseUint(c.Param("branch_id"), 10, 64)
	itemID, itemErr := strconv.ParseUint(c.Param("item_id"), 10, 64)
	if branchErr != nil || itemErr != nil || branchID == 0 || itemID == 0 {
		c.JSON(400, gin.H{"message": "Invalid branch or item ID"})
		return
	}
	page, limit := pageLimit(c)
	cards, total, err := h.Service.Cards(stockUser(c).CompanyID, branchID, itemID, limit, (page-1)*limit)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to retrieve stock card"})
		return
	}
	c.JSON(200, gin.H{"data": cards, "pagination": gin.H{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}

func pageLimit(c *gin.Context) (int, int) {
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
	return page, limit
}

func stockUser(c *gin.Context) middleware.AuthUser {
	value, _ := c.Get("authUser")
	user, _ := value.(middleware.AuthUser)
	return user
}
