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

type GoodDeliveryHandler struct{ Service *services.GoodDeliveryService }

func NewGoodDeliveryHandler(service *services.GoodDeliveryService) *GoodDeliveryHandler {
	return &GoodDeliveryHandler{Service: service}
}

type goodDeliveryPayload struct {
	RefType string                  `json:"ref_type"`
	RefNum  string                  `json:"ref_num"`
	GDDate  string                  `json:"gd_date"`
	Lines   []goodDeliveryLineInput `json:"lines"`
}

type goodDeliveryLineInput struct {
	ItemID uint64 `json:"item_id"`
	Qty    int64  `json:"qty"`
}

func (p goodDeliveryPayload) model() models.GoodDelivery {
	delivery := models.GoodDelivery{RefType: p.RefType, RefNum: p.RefNum, GDDate: p.GDDate, Lines: make([]models.GoodDeliveryLine, 0, len(p.Lines))}
	for _, line := range p.Lines {
		delivery.Lines = append(delivery.Lines, models.GoodDeliveryLine{ItemID: line.ItemID, Qty: line.Qty})
	}
	return delivery
}

// Sources lists open sales orders that can be delivered.
// @Summary List sources for a goods delivery
// @Tags Goods Deliveries
// @Security BearerAuth
// @Param ref_type query string true "Currently sales_order"
// @Success 200 {object} map[string]interface{}
// @Router /api/good-deliveries/sources [get]
func (h *GoodDeliveryHandler) Sources(c *gin.Context) {
	items, err := h.Service.Sources(gdUser(c).CompanyID, c.Query("ref_type"))
	if err != nil {
		gdError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

// Source returns remaining item quantities from a source document.
// @Summary Get remaining quantities for a goods delivery source
// @Tags Goods Deliveries
// @Security BearerAuth
// @Param ref_type path string true "Currently sales_order"
// @Param ref_num path string true "Source document number"
// @Success 200 {object} map[string]interface{}
// @Router /api/good-deliveries/sources/{ref_type}/{ref_num} [get]
func (h *GoodDeliveryHandler) Source(c *gin.Context) {
	item, err := h.Service.Source(gdUser(c).CompanyID, c.Param("ref_type"), c.Param("ref_num"))
	if err != nil {
		gdError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// List posted goods deliveries for the authenticated company.
// @Summary List goods deliveries
// @Tags Goods Deliveries
// @Security BearerAuth
// @Param search query string false "GD number, source number, or branch"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /api/good-deliveries [get]
func (h *GoodDeliveryHandler) List(c *gin.Context) {
	page, limit := pageLimit(c)
	items, total, err := h.Service.List(gdUser(c).CompanyID, c.Query("search"), limit, (page-1)*limit)
	if err != nil {
		internalServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "pagination": gin.H{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}

// Get returns one posted goods delivery with detail lines.
// @Summary Get goods delivery details
// @Tags Goods Deliveries
// @Security BearerAuth
// @Param id path int true "Goods delivery ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/good-deliveries/{id} [get]
func (h *GoodDeliveryHandler) Get(c *gin.Context) {
	id, ok := gdID(c)
	if !ok {
		return
	}
	item, err := h.Service.Get(gdUser(c).CompanyID, id)
	if err != nil {
		gdError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// Create immediately posts a goods delivery, decrements stock, and updates the source order.
// @Summary Post a goods delivery
// @Tags Goods Deliveries
// @Security BearerAuth
// @Param request body goodDeliveryPayload true "Goods delivery details"
// @Success 201 {object} map[string]interface{}
// @Failure 422 {object} map[string]string
// @Router /api/good-deliveries [post]
func (h *GoodDeliveryHandler) Create(c *gin.Context) {
	var payload goodDeliveryPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	user := gdUser(c)
	id, err := h.Service.Create(user.CompanyID, user.ID, payload.model())
	if err != nil {
		gdError(c, err)
		return
	}
	item, err := h.Service.Get(user.CompanyID, id)
	if err != nil {
		gdError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Goods delivery posted and branch stock updated", "data": item})
}

func gdUser(c *gin.Context) middleware.AuthUser {
	value, _ := c.Get("authUser")
	user, _ := value.(middleware.AuthUser)
	return user
}
func gdID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid goods delivery ID"})
		return 0, false
	}
	return id, true
}
func gdError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrGDNotFound):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrGDInvalid):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrGDInsufficientStock):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case errors.Is(err, repositories.ErrGoodDeliverySource):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	default:
		internalServerError(c, err)
	}
}
