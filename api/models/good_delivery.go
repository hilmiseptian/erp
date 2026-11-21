package models

import "time"

type GoodDeliveryLine struct {
	ID         uint64 `json:"id"`
	GDHeaderID uint64 `json:"gd_header_id"`
	ItemID     uint64 `json:"item_id"`
	SKU        string `json:"sku"`
	ItemName   string `json:"item_name"`
	Unit       string `json:"unit"`
	Qty        int64  `json:"qty"`
	Ordered    int64  `json:"ordered_qty,omitempty"`
	Delivered  int64  `json:"delivered_qty,omitempty"`
	Remaining  int64  `json:"remaining_qty,omitempty"`
}

type GoodDelivery struct {
	ID         uint64             `json:"id"`
	CompanyID  uint64             `json:"company_id"`
	BranchID   uint64             `json:"branch_id"`
	BranchName string             `json:"branch_name,omitempty"`
	RefType    string             `json:"ref_type"`
	RefNum     string             `json:"ref_num"`
	GDNumber   string             `json:"gd_number"`
	GDDate     string             `json:"gd_date"`
	Status     string             `json:"status"`
	TotalQty   int64              `json:"total_qty"`
	CreatedBy  uint64             `json:"created_by"`
	UpdatedBy  *uint64            `json:"updated_by"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  *time.Time         `json:"updated_at"`
	Lines      []GoodDeliveryLine `json:"lines,omitempty"`
}

type GoodDeliverySource struct {
	RefType         string             `json:"ref_type"`
	RefNum          string             `json:"ref_num"`
	BranchID        uint64             `json:"branch_id"`
	BranchName      string             `json:"branch_name"`
	CustomerName    string             `json:"customer_name"`
	DeliveryAddress *string            `json:"delivery_address"`
	Lines           []GoodDeliveryLine `json:"lines"`
}
