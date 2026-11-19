package models

import "time"

type PurchaseOrderLine struct {
	ID           uint64 `json:"id"`
	POHeaderID   uint64 `json:"po_header_id"`
	ItemID       uint64 `json:"item_id"`
	SKU          string `json:"sku"`
	ItemName     string `json:"item_name"`
	Unit         string `json:"unit"`
	Qty          int64  `json:"qty"`
	Price        string `json:"price"`
	RequestedQty int64  `json:"requested_qty,omitempty"`
	OrderedQty   int64  `json:"ordered_qty,omitempty"`
	RemainingQty int64  `json:"remaining_qty,omitempty"`
}

type PurchaseOrder struct {
	ID           uint64              `json:"id"`
	BranchID     uint64              `json:"branch_id"`
	BranchName   string              `json:"branch_name,omitempty"`
	SupplierID   uint64              `json:"supplier_id"`
	SupplierName string              `json:"supplier_name,omitempty"`
	PRNumber     string              `json:"ref_num"`
	PONumber     string              `json:"po_number"`
	PODate       string              `json:"po_date"`
	DueDate      string              `json:"due_date"`
	Status       string              `json:"status"`
	TotalQty     int64               `json:"total_qty"`
	TotalPrice   string              `json:"total_price"`
	CreatedBy    uint64              `json:"created_by"`
	UpdatedBy    *uint64             `json:"updated_by"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    *time.Time          `json:"updated_at"`
	Lines        []PurchaseOrderLine `json:"lines,omitempty"`
}

type PurchaseOrderSourceRequest struct {
	PRNumber     string              `json:"pr_number"`
	BranchID     uint64              `json:"branch_id"`
	BranchName   string              `json:"branch_name"`
	SupplierID   uint64              `json:"supplier_id"`
	SupplierName string              `json:"supplier_name"`
	PRDate       string              `json:"pr_date"`
	ReqDate      string              `json:"req_date"`
	Lines        []PurchaseOrderLine `json:"lines"`
}
