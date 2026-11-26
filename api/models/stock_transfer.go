package models

import "time"

type StockTransferLine struct {
	ID                   uint64 `json:"id"`
	ItemID               uint64 `json:"item_id"`
	SKU                  string `json:"sku"`
	ItemName             string `json:"item_name"`
	Unit                 string `json:"unit"`
	Qty                  int64  `json:"qty"`
	DispatchedQty        int64  `json:"dispatched_qty"`
	ReceivedQty          int64  `json:"received_qty"`
	InTransitQty         int64  `json:"in_transit_qty"`
	RemainingDispatchQty int64  `json:"remaining_dispatch_qty"`
}

type StockTransferDispatchLine struct {
	ItemID      uint64 `json:"item_id"`
	SKU         string `json:"sku"`
	ItemName    string `json:"item_name"`
	Unit        string `json:"unit"`
	Qty         int64  `json:"qty"`
	ReceivedQty int64  `json:"received_qty"`
	Remaining   int64  `json:"remaining_qty"`
}

type StockTransferDispatch struct {
	GDHeaderID uint64                      `json:"gd_header_id"`
	GDNumber   string                      `json:"gd_number"`
	GDDate     string                      `json:"gd_date"`
	TotalQty   int64                       `json:"total_qty"`
	Lines      []StockTransferDispatchLine `json:"lines"`
}

type StockTransfer struct {
	ID             uint64                  `json:"id"`
	CompanyID      uint64                  `json:"company_id"`
	FromBranchID   uint64                  `json:"from_branch_id"`
	FromBranchName string                  `json:"from_branch_name"`
	ToBranchID     uint64                  `json:"to_branch_id"`
	ToBranchName   string                  `json:"to_branch_name"`
	TransferNumber string                  `json:"transfer_number"`
	TransferDate   string                  `json:"transfer_date"`
	Status         string                  `json:"status"`
	TotalQty       int64                   `json:"total_qty"`
	Notes          *string                 `json:"notes,omitempty"`
	CreatedBy      uint64                  `json:"created_by"`
	UpdatedBy      *uint64                 `json:"updated_by"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      *time.Time              `json:"updated_at"`
	Lines          []StockTransferLine     `json:"lines,omitempty"`
	Dispatches     []StockTransferDispatch `json:"dispatches,omitempty"`
	Receipts       []StockTransferReceipt  `json:"receipts,omitempty"`
}

type StockTransferBranch struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type StockTransferItem struct {
	ID       uint64 `json:"id"`
	SKU      string `json:"sku"`
	Name     string `json:"name"`
	Unit     string `json:"unit"`
	StockQty int64  `json:"stock_qty"`
}

type StockTransferReceipt struct {
	ID            uint64 `json:"id"`
	ReceiptNumber string `json:"receipt_number"`
	GDHeaderID    uint64 `json:"gd_header_id"`
	GDNumber      string `json:"gd_number"`
	ReceiptDate   string `json:"receipt_date"`
	TotalQty      int64  `json:"total_qty"`
}
