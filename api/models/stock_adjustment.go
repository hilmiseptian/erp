package models

import "time"

type StockAdjustmentLine struct {
	ID            uint64 `json:"id"`
	ItemID        uint64 `json:"item_id"`
	SKU           string `json:"sku"`
	ItemName      string `json:"item_name"`
	Unit          string `json:"unit"`
	SystemQty     int64  `json:"system_qty"`
	ActualQty     int64  `json:"actual_qty"`
	DifferenceQty int64  `json:"difference_qty"`
}

type StockAdjustment struct {
	ID               uint64                `json:"id"`
	CompanyID        uint64                `json:"company_id"`
	BranchID         uint64                `json:"branch_id"`
	BranchName       string                `json:"branch_name"`
	AdjustmentNumber string                `json:"adjustment_number"`
	AdjustmentDate   string                `json:"adjustment_date"`
	AdjustmentType   string                `json:"adjustment_type"`
	Reason           *string               `json:"reason"`
	Status           string                `json:"status"`
	TotalQty         int64                 `json:"total_qty"`
	CreatedBy        uint64                `json:"created_by"`
	UpdatedBy        *uint64               `json:"updated_by"`
	CreatedAt        time.Time             `json:"created_at"`
	UpdatedAt        *time.Time            `json:"updated_at"`
	Lines            []StockAdjustmentLine `json:"lines,omitempty"`
}
