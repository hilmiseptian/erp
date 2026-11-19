package models

import "time"

type StockSummary struct {
	CompanyID  uint64 `json:"company_id"`
	BranchID   uint64 `json:"branch_id"`
	BranchName string `json:"branch_name"`
	ItemID     uint64 `json:"item_id"`
	SKU        string `json:"sku"`
	ItemName   string `json:"item_name"`
	Unit       string `json:"unit"`
	StockDate  string `json:"stock_date"`
	Qty        int64  `json:"qty"`
	MinStock   int64  `json:"min_stock"`
}

type StockCard struct {
	ID        uint64    `json:"id"`
	CompanyID uint64    `json:"company_id"`
	StockDate time.Time `json:"stock_date"`
	RefNum    string    `json:"ref_num"`
	BranchID  uint64    `json:"branch_id"`
	ItemID    uint64    `json:"item_id"`
	TransType string    `json:"trans_type"`
	InQty     int64     `json:"in_qty"`
	OutQty    int64     `json:"out_qty"`
	HPP       string    `json:"hpp"`
}

type GoodReceiptLine struct {
	ID        uint64 `json:"id"`
	ItemID    uint64 `json:"item_id"`
	SKU       string `json:"sku"`
	ItemName  string `json:"item_name"`
	Unit      string `json:"unit"`
	Qty       int64  `json:"qty"`
	Price     string `json:"price"`
	Ordered   int64  `json:"ordered_qty,omitempty"`
	Received  int64  `json:"received_qty,omitempty"`
	Remaining int64  `json:"remaining_qty,omitempty"`
}

type GoodReceipt struct {
	ID           uint64            `json:"id"`
	CompanyID    uint64            `json:"company_id"`
	BranchID     uint64            `json:"branch_id"`
	BranchName   string            `json:"branch_name,omitempty"`
	SupplierID   uint64            `json:"supplier_id"`
	SupplierName string            `json:"supplier_name,omitempty"`
	PONumber     string            `json:"po_number"`
	GRNumber     string            `json:"gr_number"`
	GRDate       string            `json:"gr_date"`
	Status       string            `json:"status"`
	TotalQty     int64             `json:"total_qty"`
	TotalPrice   string            `json:"total_price"`
	CreatedBy    uint64            `json:"created_by"`
	UpdatedBy    *uint64           `json:"updated_by"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    *time.Time        `json:"updated_at"`
	Lines        []GoodReceiptLine `json:"lines,omitempty"`
}

type GoodReceiptSource struct {
	PONumber     string            `json:"po_number"`
	BranchID     uint64            `json:"branch_id"`
	BranchName   string            `json:"branch_name"`
	SupplierID   uint64            `json:"supplier_id"`
	SupplierName string            `json:"supplier_name"`
	Lines        []GoodReceiptLine `json:"lines"`
}
