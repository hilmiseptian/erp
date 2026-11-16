package services

import (
	"fmt"
	"time"

	"erp-api/models"
	"erp-api/utils"
)

func validatePR(pr models.PurchaseRequest) error {
	if _, err := time.Parse("2006-01-02", pr.PRDate); err != nil {
		return fmt.Errorf("%w: pr_date must use YYYY-MM-DD", ErrPRInvalid)
	}
	if _, err := time.Parse("2006-01-02", pr.ReqDate); err != nil {
		return fmt.Errorf("%w: req_date must use YYYY-MM-DD", ErrPRInvalid)
	}
	if pr.BranchID == 0 || pr.SupplierID == 0 || len(pr.Lines) == 0 {
		return fmt.Errorf("%w: branch, supplier, and at least one detail line are required", ErrPRInvalid)
	}
	for _, line := range pr.Lines {
		qty, qtyOK := utils.ParseDecimal(line.Qty, 12, 3)
		price, priceOK := utils.ParseDecimal(line.Price, 10, 2)
		if line.ItemID == 0 || !qtyOK || qty.Sign() <= 0 || !priceOK || price.Sign() < 0 {
			return fmt.Errorf("%w: item, positive quantity (max 3 decimals), and price (max 2 decimals) are required", ErrPRInvalid)
		}
	}
	return nil
}
