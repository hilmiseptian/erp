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
	ErrPRNotFound = errors.New("purchase request not found")
	ErrPRInvalid  = errors.New("invalid purchase request")
	ErrPRLocked   = errors.New("purchase request can no longer be edited")
)

type PurchaseRequestService struct {
	Repo *repositories.PurchaseRequestRepository
}

func NewPurchaseRequestService(repo *repositories.PurchaseRequestRepository) *PurchaseRequestService {
	return &PurchaseRequestService{Repo: repo}
}
func (s *PurchaseRequestService) Branches(companyID uint64) ([]models.Branch, error) {
	return s.Repo.Branches(companyID)
}
func (s *PurchaseRequestService) Suppliers(companyID uint64) ([]models.SupplierOption, error) {
	return s.Repo.Suppliers(companyID)
}
func (s *PurchaseRequestService) List(companyID, branchID uint64, status, search string, limit, offset int) ([]models.PurchaseRequest, int, error) {
	return s.Repo.List(companyID, branchID, status, strings.TrimSpace(search), limit, offset)
}
func (s *PurchaseRequestService) Get(companyID, id uint64) (*models.PurchaseRequest, error) {
	pr, err := s.Repo.Get(companyID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPRNotFound
	}
	return pr, err
}

func (s *PurchaseRequestService) Create(companyID, userID uint64, pr models.PurchaseRequest) (uint64, error) {
	if err := validatePR(pr); err != nil {
		return 0, err
	}
	suffix := make([]byte, 3)
	if _, err := rand.Read(suffix); err != nil {
		return 0, err
	}
	number := fmt.Sprintf("PR-%s-%s", time.Now().Format("20060102-150405"), strings.ToUpper(hex.EncodeToString(suffix)))
	id, err := s.Repo.Create(companyID, userID, number, pr.PRDate, pr.ReqDate, pr.BranchID, pr.SupplierID, pr.Lines)
	if errors.Is(err, repositories.ErrPurchaseRequestTarget) {
		return 0, fmt.Errorf("%w: branch, supplier, or item is unavailable", ErrPRInvalid)
	}
	return id, err
}

func (s *PurchaseRequestService) Update(companyID, id uint64, pr models.PurchaseRequest) error {
	if err := validatePR(pr); err != nil {
		return err
	}
	existing, err := s.Get(companyID, id)
	if err != nil {
		return err
	}
	if existing.Status != "open" {
		return ErrPRLocked
	}
	err = s.Repo.Update(companyID, id, pr.PRDate, pr.ReqDate, pr.BranchID, pr.SupplierID, pr.Lines)
	if errors.Is(err, repositories.ErrPurchaseRequestLocked) {
		return ErrPRLocked
	}
	if errors.Is(err, repositories.ErrPurchaseRequestTarget) {
		return fmt.Errorf("%w: branch, supplier, or item is unavailable", ErrPRInvalid)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrPRNotFound
	}
	return err
}

func (s *PurchaseRequestService) Cancel(companyID, id uint64) error {
	if _, err := s.Get(companyID, id); err != nil {
		return err
	}
	err := s.Repo.Cancel(companyID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrPRLocked
	}
	return err
}
