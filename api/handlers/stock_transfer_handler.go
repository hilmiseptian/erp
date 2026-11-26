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

type StockTransferHandler struct {
	Service *services.StockTransferService
}

func NewStockTransferHandler(service *services.StockTransferService) *StockTransferHandler {
	return &StockTransferHandler{Service: service}
}

type stockTransferPayload struct {
	FromBranchID uint64                     `json:"from_branch_id"`
	ToBranchID   uint64                     `json:"to_branch_id"`
	TransferDate string                     `json:"transfer_date"`
	Notes        *string                    `json:"notes"`
	Lines        []stockTransferLinePayload `json:"lines"`
}

type stockTransferLinePayload struct {
	ItemID uint64 `json:"item_id"`
	Qty    int64  `json:"qty"`
}

type stockTransferMovementPayload struct {
	GDHeaderID uint64                     `json:"gd_header_id"`
	Date       string                     `json:"date"`
	Lines      []stockTransferLinePayload `json:"lines"`
}

type stockTransferReceiptPayload struct {
	GDHeaderID uint64                     `json:"gd_header_id"`
	Date       string                     `json:"date"`
	Lines      []stockTransferLinePayload `json:"lines"`
}

func (p stockTransferPayload) model() models.StockTransfer {
	transfer := models.StockTransfer{FromBranchID: p.FromBranchID, ToBranchID: p.ToBranchID, TransferDate: p.TransferDate, Notes: p.Notes}
	transfer.Lines = make([]models.StockTransferLine, 0, len(p.Lines))
	for _, line := range p.Lines {
		transfer.Lines = append(transfer.Lines, models.StockTransferLine{ItemID: line.ItemID, Qty: line.Qty})
	}
	return transfer
}

func (p stockTransferMovementPayload) lines() []models.StockTransferLine {
	lines := make([]models.StockTransferLine, 0, len(p.Lines))
	for _, line := range p.Lines {
		lines = append(lines, models.StockTransferLine{ItemID: line.ItemID, Qty: line.Qty})
	}
	return lines
}

func (p stockTransferReceiptPayload) lines() []models.StockTransferLine {
	lines := make([]models.StockTransferLine, 0, len(p.Lines))
	for _, line := range p.Lines {
		lines = append(lines, models.StockTransferLine{ItemID: line.ItemID, Qty: line.Qty})
	}
	return lines
}

// @Summary List branches available for stock transfer
// @Tags Stock Transfers
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/stock-transfers/branches [get]
func (h *StockTransferHandler) Branches(c *gin.Context) {
	branches, err := h.Service.Branches(transferUser(c).CompanyID)
	if err != nil {
		internalServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": branches})
}

// @Summary List stocked items for a source branch
// @Tags Stock Transfers
// @Security BearerAuth
// @Param branch_id query int true "Source branch ID"
// @Param search query string false "Item SKU or name"
// @Success 200 {object} map[string]interface{}
// @Router /api/stock-transfers/items [get]
func (h *StockTransferHandler) Items(c *gin.Context) {
	branchID, err := strconv.ParseUint(c.Query("branch_id"), 10, 64)
	if err != nil || branchID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "A valid source branch is required"})
		return
	}
	items, err := h.Service.Items(transferUser(c).CompanyID, branchID, c.Query("search"))
	if err != nil {
		stockTransferError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// @Summary List stock transfers
// @Tags Stock Transfers
// @Security BearerAuth
// @Param status query string false "draft, in_progress, completed, or cancelled"
// @Param search query string false "Transfer number or branch name"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /api/stock-transfers [get]
func (h *StockTransferHandler) List(c *gin.Context) {
	status := c.Query("status")
	switch status {
	case "", "draft", "in_progress", "completed", "cancelled":
	default:
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid status filter"})
		return
	}
	page, limit := pageLimit(c)
	list, total, err := h.Service.List(transferUser(c).CompanyID, status, c.Query("search"), limit, (page-1)*limit)
	if err != nil {
		internalServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": list, "pagination": gin.H{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}

// @Summary Get a stock transfer, dispatches, and received quantities
// @Tags Stock Transfers
// @Security BearerAuth
// @Param id path int true "Stock transfer ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/stock-transfers/{id} [get]
func (h *StockTransferHandler) Get(c *gin.Context) {
	id, ok := transferID(c)
	if !ok {
		return
	}
	transfer, err := h.Service.Get(transferUser(c).CompanyID, id)
	if err != nil {
		stockTransferError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": transfer})
}

// @Summary Create a stock transfer draft
// @Tags Stock Transfers
// @Security BearerAuth
// @Param request body stockTransferPayload true "Stock transfer details"
// @Success 201 {object} map[string]interface{}
// @Failure 422 {object} map[string]string
// @Router /api/stock-transfers [post]
func (h *StockTransferHandler) Create(c *gin.Context) {
	var payload stockTransferPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	user := transferUser(c)
	id, err := h.Service.Create(user.CompanyID, user.ID, payload.model())
	if err != nil {
		stockTransferError(c, err)
		return
	}
	transfer, err := h.Service.Get(user.CompanyID, id)
	if err != nil {
		internalServerError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Stock transfer draft created", "data": transfer})
}

// @Summary Update a stock transfer draft
// @Tags Stock Transfers
// @Security BearerAuth
// @Param id path int true "Stock transfer ID"
// @Param request body stockTransferPayload true "Stock transfer details"
// @Success 200 {object} map[string]interface{}
// @Router /api/stock-transfers/{id} [put]
func (h *StockTransferHandler) Update(c *gin.Context) {
	id, ok := transferID(c)
	if !ok {
		return
	}
	var payload stockTransferPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	user := transferUser(c)
	if err := h.Service.Update(user.CompanyID, user.ID, id, payload.model()); err != nil {
		stockTransferError(c, err)
		return
	}
	transfer, err := h.Service.Get(user.CompanyID, id)
	if err != nil {
		internalServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stock transfer draft updated", "data": transfer})
}

// @Summary Dispatch part of a stock transfer
// @Tags Stock Transfers
// @Security BearerAuth
// @Param id path int true "Stock transfer ID"
// @Param request body stockTransferMovementPayload true "Dispatch quantities"
// @Success 201 {object} map[string]interface{}
// @Router /api/stock-transfers/{id}/dispatches [post]
func (h *StockTransferHandler) Dispatch(c *gin.Context) {
	id, ok := transferID(c)
	if !ok {
		return
	}
	var payload stockTransferMovementPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	gdID, err := h.Service.Dispatch(transferUser(c).CompanyID, transferUser(c).ID, id, payload.Date, payload.lines())
	if err != nil {
		stockTransferError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Transfer dispatched and source stock updated", "data": gin.H{"gd_header_id": gdID}})
}

// @Summary Receive part of a dispatched stock transfer
// @Tags Stock Transfers
// @Security BearerAuth
// @Param id path int true "Stock transfer ID"
// @Param request body stockTransferReceiptPayload true "Receipt quantities and goods delivery"
// @Success 201 {object} map[string]interface{}
// @Router /api/stock-transfers/{id}/receipts [post]
func (h *StockTransferHandler) Receive(c *gin.Context) {
	id, ok := transferID(c)
	if !ok {
		return
	}
	var payload stockTransferReceiptPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	user := transferUser(c)
	receiptID, err := h.Service.Receive(user.CompanyID, user.ID, id, payload.GDHeaderID, payload.Date, payload.lines())
	if err != nil {
		stockTransferError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Transfer receipt posted and destination stock updated", "data": gin.H{"receipt_id": receiptID}})
}

// @Summary Cancel a stock transfer draft
// @Tags Stock Transfers
// @Security BearerAuth
// @Param id path int true "Stock transfer ID"
// @Success 200 {object} map[string]string
// @Router /api/stock-transfers/{id}/cancel [patch]
func (h *StockTransferHandler) Cancel(c *gin.Context) {
	id, ok := transferID(c)
	if !ok {
		return
	}
	user := transferUser(c)
	if err := h.Service.Cancel(user.CompanyID, user.ID, id); err != nil {
		stockTransferError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Stock transfer draft cancelled"})
}

func transferUser(c *gin.Context) middleware.AuthUser {
	value, _ := c.Get("authUser")
	user, _ := value.(middleware.AuthUser)
	return user
}

func transferID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid stock transfer ID"})
		return 0, false
	}
	return id, true
}

func stockTransferError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrStockTransferNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrStockTransferInvalid):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrStockTransferLocked), errors.Is(err, services.ErrStockTransferInsufficientStock), errors.Is(err, services.ErrStockTransferQuantityExceeded), errors.Is(err, repositories.ErrStockTransferQty):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	default:
		internalServerError(c, err)
	}
}
