package repositories

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"erp-api/models"
)

type PurchaseRequestRepository struct{ DB *sql.DB }

var (
	ErrPurchaseRequestLocked = errors.New("purchase request is not editable")
	ErrPurchaseRequestTarget = errors.New("branch, supplier, or item is unavailable")
)

func NewPurchaseRequestRepository(db *sql.DB) *PurchaseRequestRepository {
	return &PurchaseRequestRepository{DB: db}
}

func (r *PurchaseRequestRepository) Branches(companyID uint64) ([]models.Branch, error) {
	rows, err := r.DB.Query(`SELECT id,name,code FROM branches WHERE company_id=? AND is_active=1 ORDER BY name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	branches := make([]models.Branch, 0)
	for rows.Next() {
		var b models.Branch
		if err := rows.Scan(&b.ID, &b.Name, &b.Code); err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	return branches, rows.Err()
}

func (r *PurchaseRequestRepository) Suppliers(companyID uint64) ([]models.SupplierOption, error) {
	rows, err := r.DB.Query(`SELECT id,name,code FROM suppliers WHERE company_id=? AND is_active=1 ORDER BY name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	suppliers := make([]models.SupplierOption, 0)
	for rows.Next() {
		var s models.SupplierOption
		if err := rows.Scan(&s.ID, &s.Name, &s.Code); err != nil {
			return nil, err
		}
		suppliers = append(suppliers, s)
	}
	return suppliers, rows.Err()
}

func (r *PurchaseRequestRepository) List(companyID, branchID uint64, status, search string, limit, offset int) ([]models.PurchaseRequest, int, error) {
	where := []string{"h.company_id=?"}
	args := []any{companyID}
	if branchID > 0 {
		where = append(where, "h.branch_id=?")
		args = append(args, branchID)
	}
	if status != "" {
		where = append(where, "h.status=?")
		args = append(args, status)
	}
	if search != "" {
		where = append(where, "(h.pr_number LIKE ? OR b.name LIKE ? OR s.name LIKE ? OR u.name LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like, like, like)
	}
	clause := strings.Join(where, " AND ")
	base := ` FROM purchase_request_header h JOIN branches b ON b.id=h.branch_id JOIN suppliers s ON s.id=h.supplier_id JOIN users u ON u.id=h.created_by WHERE ` + clause
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.Query(`SELECT h.id,h.company_id,h.branch_id,b.name,h.supplier_id,s.name,h.pr_number,h.pr_date,h.req_date,h.status,h.total_qty,h.total_price,h.created_by,u.name,h.created_at,h.updated_at`+base+` ORDER BY h.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]models.PurchaseRequest, 0)
	for rows.Next() {
		x, err := scanPurchaseRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, x)
	}
	return list, total, rows.Err()
}

func (r *PurchaseRequestRepository) Get(companyID, id uint64) (*models.PurchaseRequest, error) {
	row := r.DB.QueryRow(`SELECT h.id,h.company_id,h.branch_id,b.name,h.supplier_id,s.name,h.pr_number,h.pr_date,h.req_date,h.status,h.total_qty,h.total_price,h.created_by,u.name,h.created_at,h.updated_at
		FROM purchase_request_header h JOIN branches b ON b.id=h.branch_id JOIN suppliers s ON s.id=h.supplier_id JOIN users u ON u.id=h.created_by WHERE h.company_id=? AND h.id=?`, companyID, id)
	x, err := scanPurchaseRequest(row)
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.Query(`SELECT d.id,d.pr_number,d.item_id,i.sku,i.name,i.unit,d.qty,d.price FROM purchase_request_detail d JOIN items i ON i.id=d.item_id WHERE d.pr_number=? ORDER BY d.id`, x.PRNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	x.Lines = make([]models.PurchaseRequestLine, 0)
	for rows.Next() {
		var line models.PurchaseRequestLine
		if err := rows.Scan(&line.ID, &line.PRNumber, &line.ItemID, &line.SKU, &line.ItemName, &line.Unit, &line.Qty, &line.Price); err != nil {
			return nil, err
		}
		x.Lines = append(x.Lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &x, nil
}

func (r *PurchaseRequestRepository) Create(companyID, userID uint64, number, prDate, reqDate string, branchID, supplierID uint64, lines []models.PurchaseRequestLine) (uint64, error) {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := validateBranchTx(tx, companyID, branchID); err != nil {
		return 0, err
	}
	if err := validateSupplierTx(tx, companyID, supplierID); err != nil {
		return 0, err
	}
	result, err := tx.Exec(`INSERT INTO purchase_request_header (company_id,branch_id,supplier_id,pr_number,pr_date,req_date,status,total_qty,total_price,created_by) VALUES (?,?,?,?,?,?,'open',0,0,?)`, companyID, branchID, supplierID, number, prDate, reqDate, userID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := insertPRLines(tx, number, companyID, lines); err != nil {
		return 0, err
	}
	if err := refreshPRTotals(tx, number); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *PurchaseRequestRepository) Update(companyID, id uint64, prDate, reqDate string, branchID, supplierID uint64, lines []models.PurchaseRequestLine) error {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var number, status string
	if err := tx.QueryRow(`SELECT pr_number,status FROM purchase_request_header WHERE company_id=? AND id=? FOR UPDATE`, companyID, id).Scan(&number, &status); err != nil {
		return err
	}
	if status != "open" {
		return ErrPurchaseRequestLocked
	}
	if err := validateBranchTx(tx, companyID, branchID); err != nil {
		return err
	}
	if err := validateSupplierTx(tx, companyID, supplierID); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE purchase_request_header SET branch_id=?,supplier_id=?,pr_date=?,req_date=? WHERE company_id=? AND id=?`, branchID, supplierID, prDate, reqDate, companyID, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM purchase_request_detail WHERE pr_number=?`, number); err != nil {
		return err
	}
	if err := insertPRLines(tx, number, companyID, lines); err != nil {
		return err
	}
	if err := refreshPRTotals(tx, number); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *PurchaseRequestRepository) Cancel(companyID, id uint64) error {
	res, err := r.DB.Exec(`UPDATE purchase_request_header SET status='cancelled' WHERE company_id=? AND id=? AND status IN ('open','partially_ordered')`, companyID, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

type rowScanner interface{ Scan(...any) error }

func scanPurchaseRequest(row rowScanner) (models.PurchaseRequest, error) {
	var x models.PurchaseRequest
	var date, reqDate string
	var updated sql.NullTime
	err := row.Scan(&x.ID, &x.CompanyID, &x.BranchID, &x.BranchName, &x.SupplierID, &x.SupplierName, &x.PRNumber, &date, &reqDate, &x.Status, &x.TotalQty, &x.TotalPrice, &x.CreatedBy, &x.Creator, &x.CreatedAt, &updated)
	if err != nil {
		return x, err
	}
	x.PRDate = date
	x.ReqDate = reqDate
	if updated.Valid {
		x.UpdatedAt = &updated.Time
	}
	return x, nil
}
func validateBranchTx(tx *sql.Tx, companyID, branchID uint64) error {
	var n int
	err := tx.QueryRow(`SELECT COUNT(*) FROM branches WHERE id=? AND company_id=? AND is_active=1`, branchID, companyID).Scan(&n)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrPurchaseRequestTarget
	}
	return nil
}
func validateSupplierTx(tx *sql.Tx, companyID, supplierID uint64) error {
	var n int
	err := tx.QueryRow(`SELECT COUNT(*) FROM suppliers WHERE id=? AND company_id=? AND is_active=1`, supplierID, companyID).Scan(&n)
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrPurchaseRequestTarget
	}
	return nil
}
func insertPRLines(tx *sql.Tx, number string, companyID uint64, lines []models.PurchaseRequestLine) error {
	for _, line := range lines {
		var n int
		err := tx.QueryRow(`SELECT COUNT(*) FROM items WHERE id=? AND company_id=? AND is_active=1`, line.ItemID, companyID).Scan(&n)
		if err != nil {
			return err
		}
		if n == 0 {
			return ErrPurchaseRequestTarget
		}
		if _, err := tx.Exec(`INSERT INTO purchase_request_detail (item_id,pr_number,qty,price) VALUES (?,?,?,?)`, line.ItemID, number, line.Qty, line.Price); err != nil {
			return err
		}
	}
	return nil
}
func refreshPRTotals(tx *sql.Tx, number string) error {
	_, err := tx.Exec(`UPDATE purchase_request_header SET total_qty=(SELECT COALESCE(SUM(qty),0) FROM purchase_request_detail WHERE pr_number=?),total_price=(SELECT COALESCE(SUM(qty*price),0) FROM purchase_request_detail WHERE pr_number=?) WHERE pr_number=?`, number, number, number)
	return err
}
