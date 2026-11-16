package models

import "time"

type PurchaseRequestLine struct {
	ID       uint64 `json:"id"`
	PRNumber string `json:"pr_number"`
	ItemID   uint64 `json:"item_id"`
	SKU      string `json:"sku"`
	ItemName string `json:"item_name"`
	Unit     string `json:"unit"`
	Qty      string `json:"qty"`
	Price    string `json:"price"`
}

type PurchaseRequest struct {
	ID           uint64                `json:"id"`
	CompanyID    uint64                `json:"company_id"`
	BranchID     uint64                `json:"branch_id"`
	BranchName   string                `json:"branch_name"`
	SupplierID   uint64                `json:"supplier_id"`
	SupplierName string                `json:"supplier_name"`
	PRNumber     string                `json:"pr_number"`
	PRDate       string                `json:"pr_date"`
	ReqDate      string                `json:"req_date"`
	Status       string                `json:"status"`
	TotalQty     string                `json:"total_qty"`
	TotalPrice   string                `json:"total_price"`
	CreatedBy    uint64                `json:"created_by"`
	Creator      string                `json:"creator"`
	Lines        []PurchaseRequestLine `json:"lines,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    *time.Time            `json:"updated_at"`
}

type Branch struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type SupplierOption struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}
