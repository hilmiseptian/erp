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

type PurchaseRequestHandler struct {
	Service *services.PurchaseRequestService
}

func NewPurchaseRequestHandler(service *services.PurchaseRequestService) *PurchaseRequestHandler {
	return &PurchaseRequestHandler{Service: service}
}

type purchaseRequestPayload struct {
	BranchID   uint64                       `json:"branch_id"`
	SupplierID uint64                       `json:"supplier_id"`
	PRDate     string                       `json:"pr_date"`
	ReqDate    string                       `json:"req_date"`
	Lines      []purchaseRequestLinePayload `json:"lines"`
}
type purchaseRequestLinePayload struct {
	ItemID uint64 `json:"item_id"`
	Qty    string `json:"qty"`
	Price  string `json:"price"`
}

func (p purchaseRequestPayload) model() models.PurchaseRequest {
	pr := models.PurchaseRequest{BranchID: p.BranchID, SupplierID: p.SupplierID, PRDate: p.PRDate, ReqDate: p.ReqDate, Lines: make([]models.PurchaseRequestLine, 0, len(p.Lines))}
	for _, line := range p.Lines {
		pr.Lines = append(pr.Lines, models.PurchaseRequestLine{ItemID: line.ItemID, Qty: line.Qty, Price: line.Price})
	}
	return pr
}

// Branches lists active branches belonging to the authenticated company.
// @Summary List request branches
// @Tags Purchase Requests
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/purchase-requests/branches [get]
func (h *PurchaseRequestHandler) Branches(c *gin.Context) {
	user := prUser(c)
	data, err := h.Service.Branches(user.CompanyID)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to load branches"})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

// Suppliers lists active suppliers for the authenticated company.
// @Summary List request suppliers
// @Tags Purchase Requests
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/purchase-requests/suppliers [get]
func (h *PurchaseRequestHandler) Suppliers(c *gin.Context) {
	data, err := h.Service.Suppliers(prUser(c).CompanyID)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to load suppliers"})
		return
	}
	c.JSON(200, gin.H{"data": data})
}

// List purchase requests for the authenticated company.
// @Summary List purchase requests
// @Tags Purchase Requests
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Page size (max 100)"
// @Param search query string false "PR number, branch, or requester"
// @Param branch_id query int false "Branch ID"
// @Param status query string false "open, partially_ordered, ordered, or cancelled"
// @Success 200 {object} map[string]interface{}
// @Router /api/purchase-requests [get]
func (h *PurchaseRequestHandler) List(c *gin.Context) {
	user := prUser(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	branchID, _ := strconv.ParseUint(c.Query("branch_id"), 10, 64)
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
	case "", "open", "partially_ordered", "ordered", "cancelled":
	default:
		c.JSON(400, gin.H{"message": "Invalid status filter"})
		return
	}
	list, total, err := h.Service.List(user.CompanyID, branchID, status, c.Query("search"), limit, (page-1)*limit)
	if err != nil {
		c.JSON(500, gin.H{"message": "Failed to retrieve purchase requests"})
		return
	}
	c.JSON(200, gin.H{"data": list, "pagination": gin.H{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}

// Get one purchase request and its lines.
// @Summary Get purchase request
// @Tags Purchase Requests
// @Security BearerAuth
// @Param id path int true "Purchase request ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /api/purchase-requests/{id} [get]
func (h *PurchaseRequestHandler) Get(c *gin.Context) {
	id, ok := prID(c)
	if !ok {
		return
	}
	data, err := h.Service.Get(prUser(c).CompanyID, id)
	if err != nil {
		prError(c, err)
		return
	}
	c.JSON(200, gin.H{"data": data})
}

// Create a purchase request without approval.
// @Summary Create purchase request
// @Tags Purchase Requests
// @Security BearerAuth
// @Param request body purchaseRequestPayload true "Purchase request details"
// @Success 201 {object} map[string]interface{}
// @Failure 422 {object} map[string]string
// @Router /api/purchase-requests [post]
func (h *PurchaseRequestHandler) Create(c *gin.Context) {
	var payload purchaseRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"message": "Invalid request body"})
		return
	}
	user := prUser(c)
	id, err := h.Service.Create(user.CompanyID, user.ID, payload.model())
	if err != nil {
		prError(c, err)
		return
	}
	pr, err := h.Service.Get(user.CompanyID, id)
	if err != nil {
		prError(c, err)
		return
	}
	c.JSON(201, gin.H{"message": "Purchase request created successfully", "data": pr})
}

// Update an open purchase request.
// @Summary Update purchase request
// @Tags Purchase Requests
// @Security BearerAuth
// @Param id path int true "Purchase request ID"
// @Param request body purchaseRequestPayload true "Purchase request details"
// @Success 200 {object} map[string]interface{}
// @Router /api/purchase-requests/{id} [put]
func (h *PurchaseRequestHandler) Update(c *gin.Context) {
	id, ok := prID(c)
	if !ok {
		return
	}
	var payload purchaseRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"message": "Invalid request body"})
		return
	}
	if err := h.Service.Update(prUser(c).CompanyID, id, payload.model()); err != nil {
		prError(c, err)
		return
	}
	pr, err := h.Service.Get(prUser(c).CompanyID, id)
	if err != nil {
		prError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Purchase request updated successfully", "data": pr})
}

// Cancel an open or partially ordered purchase request.
// @Summary Cancel purchase request
// @Tags Purchase Requests
// @Security BearerAuth
// @Param id path int true "Purchase request ID"
// @Success 200 {object} map[string]string
// @Router /api/purchase-requests/{id}/cancel [patch]
func (h *PurchaseRequestHandler) Cancel(c *gin.Context) {
	id, ok := prID(c)
	if !ok {
		return
	}
	if err := h.Service.Cancel(prUser(c).CompanyID, id); err != nil {
		prError(c, err)
		return
	}
	c.JSON(200, gin.H{"message": "Purchase request cancelled successfully"})
}

func prUser(c *gin.Context) middleware.AuthUser {
	value, _ := c.Get("authUser")
	user, _ := value.(middleware.AuthUser)
	return user
}
func prID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(400, gin.H{"message": "Invalid purchase request ID"})
		return 0, false
	}
	return id, true
}
func prError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrPRNotFound):
		c.JSON(404, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrPRInvalid):
		c.JSON(422, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrPRLocked):
		c.JSON(409, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
	}
}
