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
	ErrGDNotFound          = errors.New("good delivery not found")
	ErrGDInvalid           = errors.New("invalid good delivery")
	ErrGDInsufficientStock = errors.New("insufficient branch stock")
)

type GoodDeliveryService struct {
	Repo *repositories.GoodDeliveryRepository
}

func NewGoodDeliveryService(repo *repositories.GoodDeliveryRepository) *GoodDeliveryService {
	return &GoodDeliveryService{Repo: repo}
}

func (s *GoodDeliveryService) Sources(companyID uint64, refType string) ([]models.GoodDeliverySource, error) {
	refType = strings.TrimSpace(refType)
	if refType != "sales_order" {
		return nil, fmt.Errorf("%w: unsupported reference type", ErrGDInvalid)
	}
	return s.Repo.Sources(companyID, refType)
}

func (s *GoodDeliveryService) Source(companyID uint64, refType, refNum string) (*models.GoodDeliverySource, error) {
	refType, refNum = strings.TrimSpace(refType), strings.TrimSpace(refNum)
	if refType != "sales_order" || refNum == "" || len(refNum) > 50 {
		return nil, fmt.Errorf("%w: source type or reference number is invalid", ErrGDInvalid)
	}
	source, err := s.Repo.Source(companyID, refType, refNum)
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, repositories.ErrGoodDeliverySource) {
		return nil, repositories.ErrGoodDeliverySource
	}
	return source, err
}

func (s *GoodDeliveryService) List(companyID uint64, search string, limit, offset int) ([]models.GoodDelivery, int, error) {
	return s.Repo.List(companyID, search, limit, offset)
}

func (s *GoodDeliveryService) Get(companyID, id uint64) (*models.GoodDelivery, error) {
	item, err := s.Repo.Get(companyID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrGDNotFound
	}
	return item, err
}

func (s *GoodDeliveryService) Create(companyID, userID uint64, delivery models.GoodDelivery) (uint64, error) {
	delivery.RefType = strings.TrimSpace(delivery.RefType)
	delivery.RefNum = strings.TrimSpace(delivery.RefNum)
	if delivery.RefType != "sales_order" || delivery.RefNum == "" || len(delivery.RefNum) > 50 {
		return 0, fmt.Errorf("%w: select a valid source document", ErrGDInvalid)
	}
	if _, err := time.Parse("2006-01-02", delivery.GDDate); err != nil {
		return 0, fmt.Errorf("%w: gd_date must use YYYY-MM-DD", ErrGDInvalid)
	}
	if len(delivery.Lines) == 0 {
		return 0, fmt.Errorf("%w: enter a delivery quantity for at least one item", ErrGDInvalid)
	}
	for _, line := range delivery.Lines {
		if line.ItemID == 0 || line.Qty < 1 {
			return 0, fmt.Errorf("%w: item and positive whole-number quantity are required", ErrGDInvalid)
		}
	}
	suffix := make([]byte, 6)
	if _, err := rand.Read(suffix); err != nil {
		return 0, err
	}
	number := fmt.Sprintf("GD-%s-%s", time.Now().Format("20060102-150405"), strings.ToUpper(hex.EncodeToString(suffix)))
	id, err := s.Repo.Create(companyID, userID, number, delivery)
	if errors.Is(err, repositories.ErrGoodDeliverySource) {
		return 0, fmt.Errorf("%w: source sales order is unavailable for delivery", ErrGDInvalid)
	}
	if errors.Is(err, repositories.ErrGoodDeliveryQty) {
		return 0, fmt.Errorf("%w: quantity exceeds the remaining sales order quantity", ErrGDInvalid)
	}
	if errors.Is(err, repositories.ErrStockInsufficient) {
		return 0, ErrGDInsufficientStock
	}
	return id, err
}
