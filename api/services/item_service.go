package services

import (
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"erp-api/models"
	"erp-api/repositories"
)

var (
	ErrItemNotFound     = errors.New("item not found")
	ErrItemInvalid      = errors.New("invalid item")
	ErrItemDuplicateSKU = errors.New("SKU already exists")
)

type ItemService struct{ Repo *repositories.ItemRepository }

func NewItemService(repo *repositories.ItemRepository) *ItemService { return &ItemService{Repo: repo} }

func (s *ItemService) List(companyID uint64, search, status string, categoryID uint64, limit, offset int) ([]models.Item, int, error) {
	return s.Repo.List(companyID, strings.TrimSpace(search), status, categoryID, limit, offset)
}
func (s *ItemService) Get(companyID, id uint64) (*models.Item, error) {
	item, err := s.Repo.Get(companyID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrItemNotFound
	}
	return item, err
}
func (s *ItemService) Categories(companyID uint64) ([]models.ItemCategory, error) {
	return s.Repo.Categories(companyID)
}

func (s *ItemService) Save(companyID, id uint64, item models.Item) error {
	item.Name = strings.TrimSpace(item.Name)
	item.SKU = strings.TrimSpace(item.SKU)
	item.Unit = strings.TrimSpace(item.Unit)
	if companyID == 0 || item.CategoryID == 0 || item.Name == "" || item.SKU == "" || len(item.Name) > 255 || len(item.SKU) > 100 || len(item.Unit) > 30 {
		return fmt.Errorf("%w: name, SKU, category, and unit are required", ErrItemInvalid)
	}
	if item.MinStock < 0 {
		return fmt.Errorf("%w: minimum stock cannot be negative", ErrItemInvalid)
	}
	cost, ok := new(big.Rat).SetString(item.CostPrice)
	if !ok || !validDecimal(item.CostPrice) || cost.Sign() < 0 {
		return fmt.Errorf("%w: invalid cost price", ErrItemInvalid)
	}
	selling, ok := new(big.Rat).SetString(item.SellingPrice)
	if !ok || !validDecimal(item.SellingPrice) || selling.Sign() < 0 {
		return fmt.Errorf("%w: invalid selling price", ErrItemInvalid)
	}
	valid, err := s.Repo.CategoryBelongsToCompany(companyID, item.CategoryID)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("%w: category is unavailable", ErrItemInvalid)
	}
	exists, err := s.Repo.SKUExists(companyID, item.SKU, id)
	if err != nil {
		return err
	}
	if exists {
		return ErrItemDuplicateSKU
	}
	if id == 0 {
		_, err = s.Repo.Create(companyID, item)
		return err
	}
	if _, err = s.Get(companyID, id); err != nil {
		return err
	}
	return s.Repo.Update(companyID, id, item)
}

func validDecimal(value string) bool {
	if value == "" {
		return false
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return false
	}
	if len(parts[0]) > 10 {
		return false
	}
	if len(parts) == 2 && len(parts[1]) > 2 {
		return false
	}
	return true
}

func (s *ItemService) SetActive(companyID, id uint64, active bool) error {
	if _, err := s.Get(companyID, id); err != nil {
		return err
	}
	return s.Repo.SetActive(companyID, id, active)
}
