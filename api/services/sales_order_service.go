package services

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"erp-api/models"
	"erp-api/repositories"
	"erp-api/utils"
)

var (
	ErrSalesOrderNotFound = repositories.ErrSalesOrderNotFound
	ErrSalesOrderInvalid  = repositories.ErrSalesOrderInvalid
	ErrSalesOrderLocked   = repositories.ErrSalesOrderLocked
)

type SalesOrderService struct {
	Repo *repositories.SalesOrderRepository
}

func NewSalesOrderService(repo *repositories.SalesOrderRepository) *SalesOrderService {
	return &SalesOrderService{Repo: repo}
}

func (s *SalesOrderService) Branches(companyID uint64) ([]models.Branch, error) {
	return s.Repo.Branches(companyID)
}
func (s *SalesOrderService) List(companyID uint64, status, search string, limit, offset int) ([]models.SalesOrder, int, error) {
	if status != "" && status != "open" && status != "partially_delivered" && status != "delivered" && status != "cancelled" {
		return nil, 0, fmt.Errorf("%w: invalid status filter", ErrSalesOrderInvalid)
	}
	return s.Repo.List(companyID, status, strings.TrimSpace(search), limit, offset)
}
func (s *SalesOrderService) Get(companyID, id uint64) (*models.SalesOrder, error) {
	item, err := s.Repo.Get(companyID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSalesOrderNotFound
	}
	return item, err
}
func (s *SalesOrderService) Create(companyID, userID uint64, item models.SalesOrder) (uint64, error) {
	if err := validateSalesOrder(item); err != nil {
		return 0, err
	}
	random := make([]byte, 6)
	if _, err := rand.Read(random); err != nil {
		return 0, err
	}
	number := fmt.Sprintf("SO-%s-%s", time.Now().Format("20060102-150405"), strings.ToUpper(hex.EncodeToString(random)))
	return s.Repo.Create(companyID, userID, number, item)
}
func (s *SalesOrderService) Update(companyID, userID, id uint64, item models.SalesOrder) error {
	if err := validateSalesOrder(item); err != nil {
		return err
	}
	return s.Repo.Update(companyID, userID, id, item)
}
func (s *SalesOrderService) Cancel(companyID, userID, id uint64) error {
	return s.Repo.Cancel(companyID, userID, id)
}

func validateSalesOrder(item models.SalesOrder) error {
	orderDate, err := time.Parse("2006-01-02", item.SODate)
	if item.BranchID == 0 || item.CustomerID == 0 || err != nil {
		return fmt.Errorf("%w: branch, customer, and a valid YYYY-MM-DD order date are required", ErrSalesOrderInvalid)
	}
	if item.DeliveryDate != nil && strings.TrimSpace(*item.DeliveryDate) != "" {
		delivery, err := time.Parse("2006-01-02", *item.DeliveryDate)
		if err != nil || delivery.Before(orderDate) {
			return fmt.Errorf("%w: delivery date must use YYYY-MM-DD and cannot be before the order date", ErrSalesOrderInvalid)
		}
	}
	if item.DiscountAmount != "" {
		discount, ok := utils.ParseDecimal(item.DiscountAmount, 10, 2)
		if !ok || discount.Sign() < 0 {
			return fmt.Errorf("%w: order discount must be a non-negative amount with up to two decimals", ErrSalesOrderInvalid)
		}
	}
	if len(item.Lines) == 0 {
		return fmt.Errorf("%w: add at least one item", ErrSalesOrderInvalid)
	}
	for _, line := range item.Lines {
		if line.ItemID == 0 || line.Qty < 1 {
			return fmt.Errorf("%w: each line needs an item and a positive whole-number quantity", ErrSalesOrderInvalid)
		}
		if line.DiscountAmount != "" {
			discount, ok := utils.ParseDecimal(line.DiscountAmount, 10, 2)
			if !ok || discount.Sign() < 0 {
				return fmt.Errorf("%w: line discount must be a non-negative amount with up to two decimals", ErrSalesOrderInvalid)
			}
		}
	}
	return nil
}
