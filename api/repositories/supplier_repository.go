package repositories

import (
	"database/sql"

	"erp-api/models"
)

type SupplierRepository struct {
	DB *sql.DB
}

func NewSupplierRepository(db *sql.DB) *SupplierRepository {
	return &SupplierRepository{
		DB: db,
	}
}

// GetAll returns suppliers belonging to a company.
func (r *SupplierRepository) GetAll(companyID uint64) ([]models.Supplier, error) {
	query := `
        SELECT
            id,
            company_id,
            code,
            name,
            contact_person,
            phone,
            email,
            address,
            is_active,
            created_at,
            updated_at
        FROM suppliers
        WHERE company_id = ?
        ORDER BY id DESC
    `

	rows, err := r.DB.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	suppliers := make([]models.Supplier, 0)

	for rows.Next() {
		var supplier models.Supplier

		err := rows.Scan(
			&supplier.ID,
			&supplier.CompanyID,
			&supplier.Code,
			&supplier.Name,
			&supplier.ContactPerson,
			&supplier.Phone,
			&supplier.Email,
			&supplier.Address,
			&supplier.IsActive,
			&supplier.CreatedAt,
			&supplier.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		suppliers = append(suppliers, supplier)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return suppliers, nil
}

// GetByID returns a supplier belonging to a specific company.
func (r *SupplierRepository) GetByID(
	companyID uint64,
	supplierID uint64,
) (*models.Supplier, error) {
	query := `
        SELECT
            id,
            company_id,
            code,
            name,
            contact_person,
            phone,
            email,
            address,
            is_active,
            created_at,
            updated_at
        FROM suppliers
        WHERE company_id = ? AND id = ?
        LIMIT 1
    `

	var supplier models.Supplier

	err := r.DB.QueryRow(query, companyID, supplierID).Scan(
		&supplier.ID,
		&supplier.CompanyID,
		&supplier.Code,
		&supplier.Name,
		&supplier.ContactPerson,
		&supplier.Phone,
		&supplier.Email,
		&supplier.Address,
		&supplier.IsActive,
		&supplier.CreatedAt,
		&supplier.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &supplier, nil
}

// Create inserts a new supplier.
func (r *SupplierRepository) Create(supplier *models.Supplier) (uint64, error) {
	query := `
        INSERT INTO suppliers (
            company_id,
            code,
            name,
            contact_person,
            phone,
            email,
            address,
            is_active
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `

	result, err := r.DB.Exec(
		query,
		supplier.CompanyID,
		supplier.Code,
		supplier.Name,
		supplier.ContactPerson,
		supplier.Phone,
		supplier.Email,
		supplier.Address,
		supplier.IsActive,
	)

	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(id), nil
}

// Update modifies a supplier within a company.
func (r *SupplierRepository) Update(supplier *models.Supplier) error {
	query := `
        UPDATE suppliers
        SET
            code = ?,
            name = ?,
            contact_person = ?,
            phone = ?,
            email = ?,
            address = ?,
            is_active = ?
        WHERE company_id = ? AND id = ?
    `

	_, err := r.DB.Exec(
		query,
		supplier.Code,
		supplier.Name,
		supplier.ContactPerson,
		supplier.Phone,
		supplier.Email,
		supplier.Address,
		supplier.IsActive,
		supplier.CompanyID,
		supplier.ID,
	)

	return err
}

// Delete removes a supplier within a company.
func (r *SupplierRepository) Delete(
	companyID uint64,
	supplierID uint64,
) error {
	query := `
        DELETE FROM suppliers
        WHERE company_id = ? AND id = ?
    `

	_, err := r.DB.Exec(query, companyID, supplierID)

	return err
}

// IsCodeTaken checks whether a supplier code already exists
// within a company. excludeID is used when updating a supplier.
func (r *SupplierRepository) IsCodeTaken(
	companyID uint64,
	code string,
	excludeID uint64,
) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM suppliers
            WHERE company_id = ?
              AND code = ?
              AND id != ?
        )
    `

	var exists bool

	err := r.DB.QueryRow(query, companyID, code, excludeID).Scan(&exists)

	return exists, err
}
