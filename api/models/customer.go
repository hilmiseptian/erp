package models

import "time"

type Customer struct {
	ID              uint64     `json:"id"`
	CompanyID       uint64     `json:"company_id"`
	CustomerCode    string     `json:"customer_code"`
	Name            string     `json:"name"`
	TaxNumber       *string    `json:"tax_number"`
	ContactPerson   *string    `json:"contact_person"`
	Phone           *string    `json:"phone"`
	Email           *string    `json:"email"`
	BillingAddress  *string    `json:"billing_address"`
	ShippingAddress *string    `json:"shipping_address"`
	IsActive        bool       `json:"is_active"`
	CreatedBy       uint64     `json:"created_by"`
	UpdatedBy       *uint64    `json:"updated_by"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       *time.Time `json:"updated_at"`
}
