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
	ErrGRNotFound = errors.New("good receipt not found")
	ErrGRInvalid  = errors.New("invalid good receipt")
)

type GoodReceiptService struct {
	Repo *repositories.GoodReceiptRepository
}

func NewGoodReceiptService(repo *repositories.GoodReceiptRepository) *GoodReceiptService {
	return &GoodReceiptService{Repo: repo}
}
func (s *GoodReceiptService) Sources(companyID uint64) ([]models.GoodReceiptSource, error) {
	return s.Repo.Sources(companyID)
}
func (s *GoodReceiptService) Source(companyID uint64, poNumber string) (*models.GoodReceiptSource, error) {
	source, err := s.Repo.Source(companyID, strings.TrimSpace(poNumber))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repositories.ErrGoodReceiptSource
	}
	return source, err
}
func (s *GoodReceiptService) List(companyID uint64, status, search string, limit, offset int) ([]models.GoodReceipt, int, error) {
	return s.Repo.List(companyID, status, search, limit, offset)
}
func (s *GoodReceiptService) Get(companyID, id uint64) (*models.GoodReceipt, error) {
	gr, err := s.Repo.Get(companyID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrGRNotFound
	}
	return gr, err
}
func (s *GoodReceiptService) Create(companyID, userID uint64, gr models.GoodReceipt) (uint64, error) {
	if _, err := time.Parse("2006-01-02", gr.GRDate); err != nil {
		return 0, fmt.Errorf("%w: gr_date must use YYYY-MM-DD", ErrGRInvalid)
	}
	if strings.TrimSpace(gr.PONumber) == "" || len(gr.Lines) == 0 {
		return 0, fmt.Errorf("%w: purchase order and at least one item are required", ErrGRInvalid)
	}
	for _, line := range gr.Lines {
		if line.ItemID == 0 || line.Qty < 1 {
			return 0, fmt.Errorf("%w: item and positive whole-number quantity are required", ErrGRInvalid)
		}
	}
	suffix := make([]byte, 3)
	if _, err := rand.Read(suffix); err != nil {
		return 0, err
	}
	number := fmt.Sprintf("GR-%s-%s", time.Now().Format("20060102-150405"), strings.ToUpper(hex.EncodeToString(suffix)))
	id, err := s.Repo.Create(companyID, userID, number, gr.PONumber, gr.GRDate, gr.Lines)
	if errors.Is(err, repositories.ErrGoodReceiptSource) {
		return 0, fmt.Errorf("%w: purchase order is not open for receipt", ErrGRInvalid)
	}
	if errors.Is(err, repositories.ErrGoodReceiptQty) {
		return 0, fmt.Errorf("%w: quantity exceeds the remaining purchase order quantity", ErrGRInvalid)
	}
	return id, err
}
