package services

import (
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"erp-api/models"
	"erp-api/repositories"

	mysql "github.com/go-sql-driver/mysql"
)

var (
	ErrCustomerNotFound  = errors.New("customer not found")
	ErrCustomerInvalid   = errors.New("invalid customer")
	ErrCustomerCodeTaken = errors.New("customer code already exists")
)

type CustomerService struct {
	Repo *repositories.CustomerRepository
}

func NewCustomerService(repo *repositories.CustomerRepository) *CustomerService {
	return &CustomerService{Repo: repo}
}

func (s *CustomerService) List(companyID uint64, search, status string, limit, offset int) ([]models.Customer, int, error) {
	return s.Repo.List(companyID, strings.TrimSpace(search), status, limit, offset)
}

func (s *CustomerService) Get(companyID, id uint64) (*models.Customer, error) {
	item, err := s.Repo.Get(companyID, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCustomerNotFound
	}
	return item, err
}

func (s *CustomerService) Save(companyID, userID, id uint64, item models.Customer) (uint64, error) {
	item.CustomerCode = strings.TrimSpace(item.CustomerCode)
	item.Name = strings.TrimSpace(item.Name)
	if companyID == 0 || item.CustomerCode == "" || item.Name == "" || len(item.CustomerCode) > 40 || len(item.Name) > 150 {
		return 0, fmt.Errorf("%w: customer code and name are required", ErrCustomerInvalid)
	}
	if item.Email != nil {
		value := strings.TrimSpace(*item.Email)
		if value != "" {
			if _, err := mail.ParseAddress(value); err != nil || len(value) > 150 {
				return 0, fmt.Errorf("%w: enter a valid email address", ErrCustomerInvalid)
			}
		}
		if value == "" {
			item.Email = nil
		} else {
			item.Email = &value
		}
	}
	for label, value := range map[string]*string{"tax number": item.TaxNumber, "contact person": item.ContactPerson, "phone": item.Phone} {
		max := map[string]int{"tax number": 50, "contact person": 100, "phone": 30}[label]
		if value != nil && len(strings.TrimSpace(*value)) > max {
			return 0, fmt.Errorf("%w: %s is too long", ErrCustomerInvalid, label)
		}
	}
	if item.BillingAddress != nil && len(*item.BillingAddress) > 65535 || item.ShippingAddress != nil && len(*item.ShippingAddress) > 65535 {
		return 0, fmt.Errorf("%w: address is too long", ErrCustomerInvalid)
	}
	item.CompanyID, item.IsActive = companyID, true
	if id == 0 {
		item.CreatedBy = userID
	} else {
		item.ID, item.UpdatedBy = id, &userID
		existing, err := s.Get(companyID, id)
		if err != nil {
			return 0, err
		}
		item.IsActive = existing.IsActive
	}
	taken, err := s.Repo.CodeTaken(companyID, item.CustomerCode, id)
	if err != nil {
		return 0, err
	}
	if taken {
		return 0, ErrCustomerCodeTaken
	}
	if id == 0 {
		newID, err := s.Repo.Create(&item)
		if isCustomerDuplicate(err) {
			return 0, ErrCustomerCodeTaken
		}
		return newID, err
	}
	err = s.Repo.Update(&item)
	if isCustomerDuplicate(err) {
		return 0, ErrCustomerCodeTaken
	}
	return id, err
}

func (s *CustomerService) SetActive(companyID, id, userID uint64, active bool) error {
	if _, err := s.Get(companyID, id); err != nil {
		return err
	}
	return s.Repo.SetActive(companyID, id, userID, active)
}

func isCustomerDuplicate(err error) bool {
	var mysqlErr *mysql.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
