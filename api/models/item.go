package models

import "time"

// Item is the ERP-managed item master shared with the POS item schema.
type Item struct {
	ID           uint64    `json:"id"`
	CompanyID    uint64    `json:"company_id"`
	CategoryID   uint64    `json:"category_id"`
	CategoryName string    `json:"category_name,omitempty"`
	Name         string    `json:"name"`
	SKU          string    `json:"sku"`
	Description  *string   `json:"description"`
	CostPrice    string    `json:"cost_price"`
	SellingPrice string    `json:"selling_price"`
	Stock        int64     `json:"stock"`
	MinStock     int64     `json:"min_stock"`
	Unit         string    `json:"unit"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ItemCategory struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}
