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
)

var (
	ErrStockAdjustmentNotFound = errors.New("stock adjustment not found")
	ErrStockAdjustmentInvalid  = errors.New("invalid stock adjustment")
)

type StockAdjustmentService struct {
	Repo *repositories.StockAdjustmentRepository
}

func NewStockAdjustmentService(repo *repositories.StockAdjustmentRepository) *StockAdjustmentService {
	return &StockAdjustmentService{Repo: repo}
}

func (s *StockAdjustmentService) List(companyID, branchID uint64, status, adjustmentType, search string, limit, offset int) ([]models.StockAdjustment, int, error) {
	return s.Repo.List(companyID, branchID, status, adjustmentType, strings.TrimSpace(search), limit, offset)
}

func (s *StockAdjustmentService) Get(companyID, id uint64) (*models.StockAdjustment, error) {
	adjustment, err := s.Repo.Get(companyID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrStockAdjustmentNotFound
	}
	return adjustment, err
}

func (s *StockAdjustmentService) Create(companyID, userID uint64, adjustment models.StockAdjustment) (uint64, error) {
	adjustment.Reason = normalizedReason(adjustment.Reason)
	if err := validateStockAdjustment(adjustment); err != nil {
		return 0, err
	}
	number, err := newStockAdjustmentNumber()
	if err != nil {
		return 0, err
	}
	id, err := s.Repo.CreateDraft(companyID, userID, number, adjustment)
	return id, stockAdjustmentRepoError(err)
}

func (s *StockAdjustmentService) Update(companyID, userID, id uint64, adjustment models.StockAdjustment) error {
	adjustment.Reason = normalizedReason(adjustment.Reason)
	if err := validateStockAdjustment(adjustment); err != nil {
		return err
	}
	return stockAdjustmentRepoError(s.Repo.UpdateDraft(companyID, userID, id, adjustment))
}

func normalizedReason(reason *string) *string {
	if reason == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*reason)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (s *StockAdjustmentService) Post(companyID, userID, id uint64) error {
	return stockAdjustmentRepoError(s.Repo.Post(companyID, userID, id))
}

func (s *StockAdjustmentService) Cancel(companyID, userID, id uint64) error {
	return stockAdjustmentRepoError(s.Repo.Cancel(companyID, userID, id))
}

func validateStockAdjustment(adjustment models.StockAdjustment) error {
	if adjustment.BranchID == 0 {
		return fmt.Errorf("%w: branch is required", ErrStockAdjustmentInvalid)
	}
	if _, err := time.Parse("2006-01-02", adjustment.AdjustmentDate); err != nil {
		return fmt.Errorf("%w: adjustment_date must use YYYY-MM-DD", ErrStockAdjustmentInvalid)
	}
	if adjustment.AdjustmentType != "opening_balance" && adjustment.AdjustmentType != "adjustment" {
		return fmt.Errorf("%w: adjustment_type must be opening_balance or adjustment", ErrStockAdjustmentInvalid)
	}
	if len(adjustment.Lines) == 0 {
		return fmt.Errorf("%w: at least one item is required", ErrStockAdjustmentInvalid)
	}
	seen := make(map[uint64]struct{}, len(adjustment.Lines))
	for _, line := range adjustment.Lines {
		if line.ItemID == 0 || line.ActualQty < 0 {
			return fmt.Errorf("%w: item and non-negative whole-number actual quantity are required", ErrStockAdjustmentInvalid)
		}
		if _, exists := seen[line.ItemID]; exists {
			return fmt.Errorf("%w: an item can appear only once", ErrStockAdjustmentInvalid)
		}
		seen[line.ItemID] = struct{}{}
	}
	if adjustment.Reason != nil {
		if len(*adjustment.Reason) > 255 {
			return fmt.Errorf("%w: reason must not exceed 255 characters", ErrStockAdjustmentInvalid)
		}
	}
	return nil
}

func newStockAdjustmentNumber() (string, error) {
	suffix := make([]byte, 3)
	if _, err := rand.Read(suffix); err != nil {
		return "", err
	}
	return fmt.Sprintf("SA-%s-%s", time.Now().Format("20060102-150405"), strings.ToUpper(hex.EncodeToString(suffix))), nil
}

func stockAdjustmentRepoError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return ErrStockAdjustmentNotFound
	case errors.Is(err, repositories.ErrStockAdjustmentTarget):
		return fmt.Errorf("%w: branch or item is unavailable", ErrStockAdjustmentInvalid)
	case errors.Is(err, repositories.ErrStockAdjustmentTotal):
		return fmt.Errorf("%w: total difference quantity exceeds the supported range", ErrStockAdjustmentInvalid)
	case errors.Is(err, repositories.ErrStockAdjustmentLocked):
		return err
	case errors.Is(err, repositories.ErrStockAdjustmentStale):
		return err
	case errors.Is(err, repositories.ErrStockAdjustmentOpeningBalance):
		return err
	default:
		return err
	}
}
