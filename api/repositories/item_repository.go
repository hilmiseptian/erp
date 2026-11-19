package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"erp-api/models"
)

type ItemRepository struct{ DB *sql.DB }

func NewItemRepository(db *sql.DB) *ItemRepository { return &ItemRepository{DB: db} }

func (r *ItemRepository) List(companyID uint64, search, status string, categoryID uint64, limit, offset int) ([]models.Item, int, error) {
	where := []string{"i.company_id = ?"}
	args := []any{companyID}
	if search != "" {
		where = append(where, "(i.name LIKE ? OR i.sku LIKE ? OR i.description LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}
	if status == "active" || status == "inactive" {
		where = append(where, "i.is_active = ?")
		args = append(args, status == "active")
	}
	if categoryID > 0 {
		where = append(where, "i.category_id = ?")
		args = append(args, categoryID)
	}
	clause := strings.Join(where, " AND ")
	var total int
	if err := r.DB.QueryRow("SELECT COUNT(*) FROM items i WHERE "+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.Query(`SELECT i.id,i.company_id,i.category_id,c.name,i.name,i.sku,i.description,
		i.cost_price,i.selling_price,i.min_stock,i.unit,i.is_active,i.created_at,i.updated_at
		FROM items i LEFT JOIN categories c ON c.id=i.category_id AND (c.company_id=i.company_id OR c.company_id IS NULL)
		WHERE `+clause+` ORDER BY i.name,i.id LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]models.Item, 0)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *ItemRepository) Get(companyID, id uint64) (*models.Item, error) {
	row := r.DB.QueryRow(`SELECT i.id,i.company_id,i.category_id,c.name,i.name,i.sku,i.description,
		i.cost_price,i.selling_price,i.min_stock,i.unit,i.is_active,i.created_at,i.updated_at
		FROM items i LEFT JOIN categories c ON c.id=i.category_id AND (c.company_id=i.company_id OR c.company_id IS NULL)
		WHERE i.company_id=? AND i.id=? LIMIT 1`, companyID, id)
	item, err := scanItem(row)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *ItemRepository) Categories(companyID uint64) ([]models.ItemCategory, error) {
	rows, err := r.DB.Query(`SELECT id,name FROM categories WHERE is_active=1 AND (company_id=? OR company_id IS NULL) ORDER BY sort_order,name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]models.ItemCategory, 0)
	for rows.Next() {
		var c models.ItemCategory
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

func (r *ItemRepository) CategoryBelongsToCompany(companyID, categoryID uint64) (bool, error) {
	var count int
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM categories WHERE id=? AND is_active=1 AND (company_id=? OR company_id IS NULL)`, categoryID, companyID).Scan(&count)
	return count > 0, err
}

func (r *ItemRepository) SKUExists(companyID uint64, sku string, excludeID uint64) (bool, error) {
	var count int
	// The shared POS schema currently enforces SKU as globally unique.
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM items WHERE LOWER(sku)=LOWER(?) AND id<>?`, sku, excludeID).Scan(&count)
	return count > 0, err
}

func (r *ItemRepository) Create(companyID uint64, item models.Item) (uint64, error) {
	result, err := r.DB.Exec(`INSERT INTO items (company_id,category_id,name,sku,description,cost_price,selling_price,min_stock,unit,is_active) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		companyID, item.CategoryID, item.Name, item.SKU, item.Description, item.CostPrice, item.SellingPrice, item.MinStock, item.Unit, item.IsActive)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return uint64(id), err
}

func (r *ItemRepository) Update(companyID, id uint64, item models.Item) error {
	_, err := r.DB.Exec(`UPDATE items SET category_id=?,name=?,sku=?,description=?,cost_price=?,selling_price=?,min_stock=?,unit=?,is_active=? WHERE company_id=? AND id=?`,
		item.CategoryID, item.Name, item.SKU, item.Description, item.CostPrice, item.SellingPrice, item.MinStock, item.Unit, item.IsActive, companyID, id)
	return err
}

func (r *ItemRepository) SetActive(companyID, id uint64, active bool) error {
	result, err := r.DB.Exec(`UPDATE items SET is_active=? WHERE company_id=? AND id=?`, active, companyID, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		_, err = r.Get(companyID, id)
	}
	return err
}

func scanItem(row scanner) (models.Item, error) {
	var i models.Item
	var categoryName sql.NullString
	err := row.Scan(&i.ID, &i.CompanyID, &i.CategoryID, &categoryName, &i.Name, &i.SKU, &i.Description, &i.CostPrice, &i.SellingPrice, &i.MinStock, &i.Unit, &i.IsActive, &i.CreatedAt, &i.UpdatedAt)
	if err != nil {
		return i, fmt.Errorf("scan item: %w", err)
	}
	if categoryName.Valid {
		i.CategoryName = categoryName.String
	}
	return i, nil
}
