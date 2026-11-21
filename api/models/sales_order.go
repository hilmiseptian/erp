package models

import "time"

type SalesOrderLine struct {
	ID             uint64  `json:"id"`
	SOHeaderID     uint64  `json:"so_header_id"`
	ItemID         uint64  `json:"item_id"`
	SKU            string  `json:"sku"`
	ItemName       string  `json:"item_name"`
	Unit           string  `json:"unit"`
	Qty            int64   `json:"qty"`
	UnitPrice      string  `json:"unit_price"`
	DiscountAmount string  `json:"discount_amount"`
	Subtotal       string  `json:"subtotal"`
	Notes          *string `json:"notes"`
}

type SalesOrder struct {
	ID              uint64           `json:"id"`
	CompanyID       uint64           `json:"company_id"`
	BranchID        uint64           `json:"branch_id"`
	BranchName      string           `json:"branch_name,omitempty"`
	CustomerID      uint64           `json:"customer_id"`
	CustomerName    string           `json:"customer_name"`
	SONumber        string           `json:"so_number"`
	SODate          string           `json:"so_date"`
	DeliveryDate    *string          `json:"delivery_date"`
	DeliveryAddress *string          `json:"delivery_address"`
	Status          string           `json:"status"`
	TotalQty        int64            `json:"total_qty"`
	TotalAmount     string           `json:"total_amount"`
	DiscountAmount  string           `json:"discount_amount"`
	GrandTotal      string           `json:"grand_total"`
	Notes           *string          `json:"notes"`
	CreatedBy       uint64           `json:"created_by"`
	UpdatedBy       *uint64          `json:"updated_by"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       *time.Time       `json:"updated_at"`
	Lines           []SalesOrderLine `json:"lines,omitempty"`
}
