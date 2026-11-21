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

type CustomerHandler struct{ Service *services.CustomerService }

func NewCustomerHandler(service *services.CustomerService) *CustomerHandler {
	return &CustomerHandler{Service: service}
}

type customerPayload struct {
	CustomerCode    string  `json:"customer_code"`
	Name            string  `json:"name"`
	TaxNumber       *string `json:"tax_number"`
	ContactPerson   *string `json:"contact_person"`
	Phone           *string `json:"phone"`
	Email           *string `json:"email"`
	BillingAddress  *string `json:"billing_address"`
	ShippingAddress *string `json:"shipping_address"`
}

type customerStatusPayload struct {
	IsActive *bool `json:"is_active"`
}

// List customers for the authenticated company.
// @Summary List customers
// @Tags Customers
// @Security BearerAuth
// @Param search query string false "Search customer code, name, tax number, or contact"
// @Param status query string false "active or inactive"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /api/customers [get]
func (h *CustomerHandler) List(c *gin.Context) {
	page, limit := pageLimit(c)
	status := c.Query("status")
	if c.Query("active") == "true" {
		status = "active"
	}
	user := customerUser(c)
	items, total, err := h.Service.List(user.CompanyID, c.Query("search"), status, limit, (page-1)*limit)
	if err != nil {
		internalServerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "pagination": gin.H{"page": page, "limit": limit, "total": total, "total_pages": (total + limit - 1) / limit}})
}

// Get a customer in the authenticated company.
// @Summary Get customer
// @Tags Customers
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/customers/{id} [get]
func (h *CustomerHandler) Get(c *gin.Context) {
	id, ok := customerID(c)
	if !ok {
		return
	}
	item, err := h.Service.Get(customerUser(c).CompanyID, id)
	if err != nil {
		customerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

// Create a customer for the authenticated company.
// @Summary Create customer
// @Tags Customers
// @Security BearerAuth
// @Param customer body customerPayload true "Customer details"
// @Success 201 {object} map[string]interface{}
// @Router /api/customers [post]
func (h *CustomerHandler) Create(c *gin.Context) {
	var payload customerPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	user := customerUser(c)
	id, err := h.Service.Save(user.CompanyID, user.ID, 0, payload.model())
	if err != nil {
		customerError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Customer created successfully", "data": gin.H{"id": id}})
}

// Update a customer in the authenticated company.
// @Summary Update customer
// @Tags Customers
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Param customer body customerPayload true "Customer details"
// @Success 200 {object} map[string]interface{}
// @Router /api/customers/{id} [put]
func (h *CustomerHandler) Update(c *gin.Context) {
	id, ok := customerID(c)
	if !ok {
		return
	}
	var payload customerPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request body"})
		return
	}
	user := customerUser(c)
	_, err := h.Service.Save(user.CompanyID, user.ID, id, payload.model())
	if err != nil {
		customerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Customer updated successfully"})
}

// SetActive changes the customer's active state.
// @Summary Activate or deactivate customer
// @Tags Customers
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Param status body customerStatusPayload true "Active state"
// @Success 200 {object} map[string]interface{}
// @Router /api/customers/{id}/status [patch]
func (h *CustomerHandler) SetActive(c *gin.Context) {
	id, ok := customerID(c)
	if !ok {
		return
	}
	var request customerStatusPayload
	if err := c.ShouldBindJSON(&request); err != nil || request.IsActive == nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "is_active is required"})
		return
	}
	user := customerUser(c)
	if err := h.Service.SetActive(user.CompanyID, id, user.ID, *request.IsActive); err != nil {
		customerError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Customer status updated successfully"})
}

func (p customerPayload) model() models.Customer {
	return models.Customer{CustomerCode: p.CustomerCode, Name: p.Name, TaxNumber: p.TaxNumber, ContactPerson: p.ContactPerson, Phone: p.Phone, Email: p.Email, BillingAddress: p.BillingAddress, ShippingAddress: p.ShippingAddress}
}

func customerUser(c *gin.Context) middleware.AuthUser {
	value, _ := c.Get("authUser")
	user, _ := value.(middleware.AuthUser)
	return user
}
func customerID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(400, gin.H{"message": "Invalid customer ID"})
		return 0, false
	}
	return id, true
}
func customerError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, services.ErrCustomerNotFound):
		c.JSON(404, gin.H{"message": err.Error()})
	case errors.Is(err, services.ErrCustomerInvalid), errors.Is(err, services.ErrCustomerCodeTaken):
		c.JSON(422, gin.H{"message": err.Error()})
	default:
		internalServerError(c, err)
	}
}
