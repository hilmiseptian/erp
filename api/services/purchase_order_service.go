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
	ErrPONotFound = errors.New("purchase order not found")
	ErrPOInvalid  = errors.New("invalid purchase order")
	ErrPOLocked   = errors.New("purchase order cannot be cancelled in its current state")
)

type PurchaseOrderService struct {
	Repo *repositories.PurchaseOrderRepository
}

func NewPurchaseOrderService(repo *repositories.PurchaseOrderRepository) *PurchaseOrderService {
	return &PurchaseOrderService{Repo: repo}
}

func (s *PurchaseOrderService) Sources(companyID uint64) ([]models.PurchaseOrderSourceRequest, error) {
	return s.Repo.Sources(companyID)
}

func (s *PurchaseOrderService) Source(companyID uint64, prNumber string) (*models.PurchaseOrderSourceRequest, error) {
	source, err := s.Repo.Source(companyID, strings.TrimSpace(prNumber))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repositories.ErrPurchaseOrderSource
	}
	return source, err
}

func (s *PurchaseOrderService) List(companyID uint64, status, search string, limit, offset int) ([]models.PurchaseOrder, int, error) {
	return s.Repo.List(companyID, status, strings.TrimSpace(search), limit, offset)
}

func (s *PurchaseOrderService) Get(companyID, id uint64) (*models.PurchaseOrder, error) {
	po, err := s.Repo.Get(companyID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPONotFound
	}
	return po, err
}

func (s *PurchaseOrderService) Create(companyID, userID uint64, po models.PurchaseOrder) (uint64, error) {
	if err := validatePO(po); err != nil {
		return 0, err
	}
	suffix := make([]byte, 3)
	if _, err := rand.Read(suffix); err != nil {
		return 0, err
	}
	number := fmt.Sprintf("PO-%s-%s", time.Now().Format("20060102-150405"), strings.ToUpper(hex.EncodeToString(suffix)))
	id, err := s.Repo.Create(companyID, userID, number, po.PRNumber, po.PODate, po.DueDate, po.Lines)
	if errors.Is(err, repositories.ErrPurchaseOrderSource) {
		return 0, fmt.Errorf("%w: source purchase request is unavailable", ErrPOInvalid)
	}
	if errors.Is(err, repositories.ErrPurchaseOrderQty) {
		return 0, fmt.Errorf("%w: quantity exceeds the remaining quantity on the source purchase request", ErrPOInvalid)
	}
	return id, err
}

func (s *PurchaseOrderService) Cancel(companyID, id, userID uint64) error {
	err := s.Repo.Cancel(companyID, id, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrPONotFound
	}
	if errors.Is(err, repositories.ErrPurchaseOrderSource) {
		return ErrPOLocked
	}
	return err
}

func validatePO(po models.PurchaseOrder) error {
	poDate, err := time.Parse("2006-01-02", po.PODate)
	if err != nil {
		return fmt.Errorf("%w: po_date must use YYYY-MM-DD", ErrPOInvalid)
	}
	dueDate, err := time.Parse("2006-01-02", po.DueDate)
	if err != nil || dueDate.Before(poDate) {
		return fmt.Errorf("%w: due_date must use YYYY-MM-DD and cannot be before po_date", ErrPOInvalid)
	}
	if strings.TrimSpace(po.PRNumber) == "" || len(po.Lines) == 0 {
		return fmt.Errorf("%w: source PR and at least one detail line are required", ErrPOInvalid)
	}
	for _, line := range po.Lines {
		qty, qtyOK := utils.ParseDecimal(line.Qty, 12, 3)
		price, priceOK := utils.ParseDecimal(line.Price, 10, 2)
		if line.ItemID == 0 || !qtyOK || qty.Sign() <= 0 || !priceOK || price.Sign() < 0 {
			return fmt.Errorf("%w: item, positive quantity (max 3 decimals), and non-negative price (max 2 decimals) are required", ErrPOInvalid)
		}
	}
	return nil
}
