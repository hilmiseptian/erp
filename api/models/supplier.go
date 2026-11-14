package models

import "time"

type Supplier struct {
	ID            uint64    `json:"id"`
	CompanyID     uint64    `json:"company_id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	ContactPerson *string   `json:"contact_person"`
	Phone         *string   `json:"phone"`
	Email         *string   `json:"email"`
	Address       *string   `json:"address"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}
