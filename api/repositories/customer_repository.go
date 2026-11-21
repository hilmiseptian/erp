package repositories

import (
	"database/sql"
	"strings"

	"erp-api/models"
)

type CustomerRepository struct{ DB *sql.DB }

func NewCustomerRepository(db *sql.DB) *CustomerRepository { return &CustomerRepository{DB: db} }

func (r *CustomerRepository) List(companyID uint64, search, status string, limit, offset int) ([]models.Customer, int, error) {
	where := []string{"company_id=?"}
	args := []any{companyID}
	if search != "" {
		where = append(where, "(customer_code LIKE ? OR name LIKE ? OR tax_number LIKE ? OR contact_person LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like, like, like)
	}
	if status == "active" || status == "inactive" {
		where = append(where, "is_active=?")
		args = append(args, status == "active")
	}
	clause := strings.Join(where, " AND ")
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*) FROM customers WHERE `+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.Query(`SELECT id,company_id,customer_code,name,tax_number,contact_person,phone,email,billing_address,shipping_address,is_active,created_by,updated_by,created_at,updated_at FROM customers WHERE `+clause+` ORDER BY name,id LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]models.Customer, 0)
	for rows.Next() {
		var item models.Customer
		if err := rows.Scan(&item.ID, &item.CompanyID, &item.CustomerCode, &item.Name, &item.TaxNumber, &item.ContactPerson, &item.Phone, &item.Email, &item.BillingAddress, &item.ShippingAddress, &item.IsActive, &item.CreatedBy, &item.UpdatedBy, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *CustomerRepository) Get(companyID, id uint64) (*models.Customer, error) {
	var item models.Customer
	err := r.DB.QueryRow(`SELECT id,company_id,customer_code,name,tax_number,contact_person,phone,email,billing_address,shipping_address,is_active,created_by,updated_by,created_at,updated_at FROM customers WHERE company_id=? AND id=?`, companyID, id).
		Scan(&item.ID, &item.CompanyID, &item.CustomerCode, &item.Name, &item.TaxNumber, &item.ContactPerson, &item.Phone, &item.Email, &item.BillingAddress, &item.ShippingAddress, &item.IsActive, &item.CreatedBy, &item.UpdatedBy, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CustomerRepository) CodeTaken(companyID uint64, code string, excludeID uint64) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM customers WHERE company_id=? AND customer_code=? AND id<>?)`, companyID, code, excludeID).Scan(&exists)
	return exists, err
}

func (r *CustomerRepository) Create(item *models.Customer) (uint64, error) {
	result, err := r.DB.Exec(`INSERT INTO customers(company_id,customer_code,name,tax_number,contact_person,phone,email,billing_address,shipping_address,is_active,created_by) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, item.CompanyID, item.CustomerCode, item.Name, item.TaxNumber, item.ContactPerson, item.Phone, item.Email, item.BillingAddress, item.ShippingAddress, item.IsActive, item.CreatedBy)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return uint64(id), err
}

func (r *CustomerRepository) Update(item *models.Customer) error {
	_, err := r.DB.Exec(`UPDATE customers SET customer_code=?,name=?,tax_number=?,contact_person=?,phone=?,email=?,billing_address=?,shipping_address=?,is_active=?,updated_by=? WHERE company_id=? AND id=?`, item.CustomerCode, item.Name, item.TaxNumber, item.ContactPerson, item.Phone, item.Email, item.BillingAddress, item.ShippingAddress, item.IsActive, item.UpdatedBy, item.CompanyID, item.ID)
	return err
}

func (r *CustomerRepository) SetActive(companyID, id, userID uint64, active bool) error {
	result, err := r.DB.Exec(`UPDATE customers SET is_active=?,updated_by=? WHERE company_id=? AND id=?`, active, userID, companyID, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err == nil && count == 0 {
		_, err = r.Get(companyID, id)
	}
	return err
}
