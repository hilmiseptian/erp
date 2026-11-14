package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"erp-api/middleware"
	"erp-api/models"
	"erp-api/services"
)

type SupplierHandler struct {
	Service *services.SupplierService
}

func NewSupplierHandler(service *services.SupplierService) *SupplierHandler {
	return &SupplierHandler{
		Service: service,
	}
}

type supplierRequest struct {
	Code          string  `json:"code"`
	Name          string  `json:"name"`
	ContactPerson *string `json:"contact_person"`
	Phone         *string `json:"phone"`
	Email         *string `json:"email"`
	Address       *string `json:"address"`
	IsActive      *bool   `json:"is_active"`
}

// GetAll handles GET /api/suppliers
func (h *SupplierHandler) GetAll(c *gin.Context) {
	user := getAuthUser(c)

	suppliers, err := h.Service.GetAll(user.CompanyID)
	if err != nil {
		internalServerError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Suppliers retrieved successfully",
		"data":    suppliers,
	})
}

// GetByID handles GET /api/suppliers/:id
func (h *SupplierHandler) GetByID(c *gin.Context) {
	user := getAuthUser(c)

	supplierID, ok := parseSupplierID(c)
	if !ok {
		return
	}

	supplier, err := h.Service.GetByID(user.CompanyID, supplierID)
	if err != nil {
		handleSupplierError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Supplier retrieved successfully",
		"data":    supplier,
	})
}

// Create handles POST /api/suppliers
func (h *SupplierHandler) Create(c *gin.Context) {
	user := getAuthUser(c)

	var request supplierRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	isActive := true
	if request.IsActive != nil {
		isActive = *request.IsActive
	}

	supplier := &models.Supplier{
		Code:          request.Code,
		Name:          request.Name,
		ContactPerson: request.ContactPerson,
		Phone:         request.Phone,
		Email:         request.Email,
		Address:       request.Address,
		IsActive:      isActive,
	}

	id, err := h.Service.Create(user.CompanyID, supplier)
	if err != nil {
		handleSupplierError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Supplier created successfully",
		"data": gin.H{
			"id": id,
		},
	})
}

// Update handles PUT /api/suppliers/:id
func (h *SupplierHandler) Update(c *gin.Context) {
	user := getAuthUser(c)

	supplierID, ok := parseSupplierID(c)
	if !ok {
		return
	}

	var request supplierRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	// Preserve the existing status if is_active is omitted.
	isActive := true

	if request.IsActive != nil {
		isActive = *request.IsActive
	} else {
		existing, err := h.Service.GetByID(user.CompanyID, supplierID)
		if err != nil {
			handleSupplierError(c, err)
			return
		}

		isActive = existing.IsActive
	}

	supplier := &models.Supplier{
		Code:          request.Code,
		Name:          request.Name,
		ContactPerson: request.ContactPerson,
		Phone:         request.Phone,
		Email:         request.Email,
		Address:       request.Address,
		IsActive:      isActive,
	}

	if err := h.Service.Update(user.CompanyID, supplierID, supplier); err != nil {
		handleSupplierError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Supplier updated successfully",
	})
}

// Delete handles DELETE /api/suppliers/:id
func (h *SupplierHandler) Delete(c *gin.Context) {
	user := getAuthUser(c)

	supplierID, ok := parseSupplierID(c)
	if !ok {
		return
	}

	if err := h.Service.Delete(user.CompanyID, supplierID); err != nil {
		handleSupplierError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Supplier deleted successfully",
	})
}

func getAuthUser(c *gin.Context) middleware.AuthUser {
	value, exists := c.Get("authUser")
	if !exists {
		return middleware.AuthUser{}
	}

	user, ok := value.(middleware.AuthUser)
	if !ok {
		return middleware.AuthUser{}
	}

	return user
}

func parseSupplierID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)

	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid supplier ID",
		})
		return 0, false
	}

	return id, true
}

func handleSupplierError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrSupplierNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"message": err.Error(),
		})

	case errors.Is(err, services.ErrDuplicateCode):
		c.JSON(http.StatusConflict, gin.H{
			"message": err.Error(),
		})

	case errors.Is(err, services.ErrInvalidInput):
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": err.Error(),
		})

	default:
		internalServerError(c, err)
	}
}

func internalServerError(c *gin.Context, err error) {
	// Log the actual error on the server, not in the API response.
	c.Error(err)

	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "Internal server error",
	})
}
