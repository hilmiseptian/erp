package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"erp-api/models"
	"erp-api/repositories"

	mysql "github.com/go-sql-driver/mysql"
)

var (
	ErrSupplierNotFound = errors.New("supplier not found")
	ErrDuplicateCode    = errors.New("supplier code already exists")
	ErrInvalidInput     = errors.New("invalid supplier input")
)

type SupplierService struct {
	Repository *repositories.SupplierRepository
}

func NewSupplierService(
	repository *repositories.SupplierRepository,
) *SupplierService {
	return &SupplierService{
		Repository: repository,
	}
}

func (s *SupplierService) GetAll(companyID uint64) ([]models.Supplier, error) {
	return s.Repository.GetAll(companyID)
}

func (s *SupplierService) GetByID(
	companyID uint64,
	supplierID uint64,
) (*models.Supplier, error) {
	supplier, err := s.Repository.GetByID(companyID, supplierID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrSupplierNotFound
	}

	if err != nil {
		return nil, err
	}

	return supplier, nil
}

func (s *SupplierService) Create(
	companyID uint64,
	supplier *models.Supplier,
) (uint64, error) {
	if err := validateSupplier(supplier); err != nil {
		return 0, err
	}

	supplier.CompanyID = companyID
	supplier.Code = strings.TrimSpace(supplier.Code)
	supplier.Name = strings.TrimSpace(supplier.Name)

	exists, err := s.Repository.IsCodeTaken(
		companyID,
		supplier.Code,
		0,
	)

	if err != nil {
		return 0, err
	}

	if exists {
		return 0, ErrDuplicateCode
	}

	id, err := s.Repository.Create(supplier)

	if err != nil {
		if isDuplicateKey(err) {
			return 0, ErrDuplicateCode
		}

		return 0, err
	}

	return id, nil
}

func (s *SupplierService) Update(
	companyID uint64,
	supplierID uint64,
	supplier *models.Supplier,
) error {
	if err := validateSupplier(supplier); err != nil {
		return err
	}

	// Ensure supplier belongs to the authenticated company.
	_, err := s.GetByID(companyID, supplierID)

	if err != nil {
		return err
	}

	supplier.ID = supplierID
	supplier.CompanyID = companyID
	supplier.Code = strings.TrimSpace(supplier.Code)
	supplier.Name = strings.TrimSpace(supplier.Name)

	exists, err := s.Repository.IsCodeTaken(
		companyID,
		supplier.Code,
		supplierID,
	)

	if err != nil {
		return err
	}

	if exists {
		return ErrDuplicateCode
	}

	err = s.Repository.Update(supplier)

	if err != nil {
		if isDuplicateKey(err) {
			return ErrDuplicateCode
		}

		return err
	}

	return nil
}

func (s *SupplierService) Delete(
	companyID uint64,
	supplierID uint64,
) error {
	// Check ownership and existence before deleting.
	_, err := s.GetByID(companyID, supplierID)

	if err != nil {
		return err
	}

	return s.Repository.Delete(companyID, supplierID)
}

func validateSupplier(supplier *models.Supplier) error {
	if strings.TrimSpace(supplier.Code) == "" {
		return fmt.Errorf("%w: code is required", ErrInvalidInput)
	}

	if strings.TrimSpace(supplier.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}

	if len(strings.TrimSpace(supplier.Code)) > 30 {
		return fmt.Errorf("%w: code maximum length is 30", ErrInvalidInput)
	}

	if len(strings.TrimSpace(supplier.Name)) > 150 {
		return fmt.Errorf("%w: name maximum length is 150", ErrInvalidInput)
	}

	return nil
}

func isDuplicateKey(err error) bool {
	var mysqlErr *mysql.MySQLError

	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
