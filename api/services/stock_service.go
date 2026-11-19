package services

import (
	"erp-api/models"
	"erp-api/repositories"
)

type StockService struct{ Repo *repositories.StockRepository }

func NewStockService(repo *repositories.StockRepository) *StockService {
	return &StockService{Repo: repo}
}

func (s *StockService) Branches(companyID uint64) ([]models.Branch, error) {
	return s.Repo.Branches(companyID)
}

func (s *StockService) List(companyID, branchID uint64, search string, limit, offset int) ([]models.StockSummary, int, error) {
	return s.Repo.List(companyID, branchID, search, limit, offset)
}

func (s *StockService) Cards(companyID, branchID, itemID uint64, limit, offset int) ([]models.StockCard, int, error) {
	return s.Repo.Cards(companyID, branchID, itemID, limit, offset)
}
