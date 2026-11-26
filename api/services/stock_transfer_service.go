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
	ErrStockTransferNotFound          = errors.New("stock transfer not found")
	ErrStockTransferInvalid           = errors.New("invalid stock transfer")
	ErrStockTransferLocked            = errors.New("stock transfer is no longer editable")
	ErrStockTransferInsufficientStock = errors.New("insufficient stock in the source branch")
	ErrStockTransferQuantityExceeded  = errors.New("stock transfer quantity exceeds the remaining quantity")
)

type StockTransferService struct {
	Repo *repositories.StockTransferRepository
}

func NewStockTransferService(repo *repositories.StockTransferRepository) *StockTransferService {
	return &StockTransferService{Repo: repo}
}

func (s *StockTransferService) Branches(companyID uint64) ([]models.StockTransferBranch, error) {
	return s.Repo.Branches(companyID)
}

func (s *StockTransferService) Items(companyID, branchID uint64, search string) ([]models.StockTransferItem, error) {
	if branchID == 0 {
		return nil, fmt.Errorf("%w: source branch is required", ErrStockTransferInvalid)
	}
	return s.Repo.Items(companyID, branchID, strings.TrimSpace(search))
}

func (s *StockTransferService) List(companyID uint64, status, search string, limit, offset int) ([]models.StockTransfer, int, error) {
	return s.Repo.List(companyID, status, strings.TrimSpace(search), limit, offset)
}

func (s *StockTransferService) Get(companyID, id uint64) (*models.StockTransfer, error) {
	transfer, err := s.Repo.Get(companyID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrStockTransferNotFound
	}
	return transfer, err
}

func (s *StockTransferService) Create(companyID, userID uint64, transfer models.StockTransfer) (uint64, error) {
	if err := validateStockTransfer(transfer); err != nil {
		return 0, err
	}
	number, err := stockTransferNumber("ST")
	if err != nil {
		return 0, err
	}
	id, err := s.Repo.CreateDraft(companyID, userID, number, transfer)
	return id, stockTransferRepoError(err)
}

func (s *StockTransferService) Update(companyID, userID, id uint64, transfer models.StockTransfer) error {
	if err := validateStockTransfer(transfer); err != nil {
		return err
	}
	return stockTransferRepoError(s.Repo.UpdateDraft(companyID, userID, id, transfer))
}

func (s *StockTransferService) Cancel(companyID, userID, id uint64) error {
	return stockTransferRepoError(s.Repo.Cancel(companyID, userID, id))
}

func (s *StockTransferService) Dispatch(companyID, userID, id uint64, date string, lines []models.StockTransferLine) (uint64, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return 0, fmt.Errorf("%w: dispatch_date must use YYYY-MM-DD", ErrStockTransferInvalid)
	}
	if err := validateTransferLines(lines); err != nil {
		return 0, err
	}
	number, err := stockTransferNumber("GD")
	if err != nil {
		return 0, err
	}
	gdID, err := s.Repo.Dispatch(companyID, userID, id, number, date, lines)
	return gdID, stockTransferRepoError(err)
}

func (s *StockTransferService) Receive(companyID, userID, id, gdHeaderID uint64, date string, lines []models.StockTransferLine) (uint64, error) {
	if gdHeaderID == 0 {
		return 0, fmt.Errorf("%w: select a goods delivery", ErrStockTransferInvalid)
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return 0, fmt.Errorf("%w: receipt_date must use YYYY-MM-DD", ErrStockTransferInvalid)
	}
	if err := validateTransferLines(lines); err != nil {
		return 0, err
	}
	number, err := stockTransferNumber("TR")
	if err != nil {
		return 0, err
	}
	receiptID, err := s.Repo.Receive(companyID, userID, id, gdHeaderID, number, date, lines)
	return receiptID, stockTransferRepoError(err)
}

func validateStockTransfer(transfer models.StockTransfer) error {
	if transfer.FromBranchID == 0 || transfer.ToBranchID == 0 || transfer.FromBranchID == transfer.ToBranchID {
		return fmt.Errorf("%w: select two different branches", ErrStockTransferInvalid)
	}
	if _, err := time.Parse("2006-01-02", transfer.TransferDate); err != nil {
		return fmt.Errorf("%w: transfer_date must use YYYY-MM-DD", ErrStockTransferInvalid)
	}
	if err := validateTransferLines(transfer.Lines); err != nil {
		return err
	}
	if transfer.Notes != nil && len(strings.TrimSpace(*transfer.Notes)) > 2000 {
		return fmt.Errorf("%w: notes must not exceed 2000 characters", ErrStockTransferInvalid)
	}
	return nil
}

func validateTransferLines(lines []models.StockTransferLine) error {
	if len(lines) == 0 {
		return fmt.Errorf("%w: add at least one item", ErrStockTransferInvalid)
	}
	seen := make(map[uint64]struct{}, len(lines))
	for _, line := range lines {
		if line.ItemID == 0 || line.Qty < 1 {
			return fmt.Errorf("%w: item and positive whole-number quantity are required", ErrStockTransferInvalid)
		}
		if _, exists := seen[line.ItemID]; exists {
			return fmt.Errorf("%w: an item can appear only once", ErrStockTransferInvalid)
		}
		seen[line.ItemID] = struct{}{}
	}
	return nil
}

func stockTransferNumber(prefix string) (string, error) {
	suffix := make([]byte, 5)
	if _, err := rand.Read(suffix); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%s", prefix, time.Now().Format("20060102-150405"), strings.ToUpper(hex.EncodeToString(suffix))), nil
}

func stockTransferRepoError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return ErrStockTransferNotFound
	case errors.Is(err, repositories.ErrStockTransferBranch), errors.Is(err, repositories.ErrStockTransferItem):
		return fmt.Errorf("%w: branch or item is unavailable", ErrStockTransferInvalid)
	case errors.Is(err, repositories.ErrStockTransferLocked):
		return ErrStockTransferLocked
	case errors.Is(err, repositories.ErrStockTransferQty):
		return ErrStockTransferQuantityExceeded
	case errors.Is(err, repositories.ErrStockTransferDispatch):
		return fmt.Errorf("%w: goods delivery is not available for receipt", ErrStockTransferInvalid)
	case errors.Is(err, repositories.ErrStockInsufficient):
		return ErrStockTransferInsufficientStock
	default:
		return err
	}
}
