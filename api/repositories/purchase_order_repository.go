package repositories

import (
	"context"
	"database/sql"
	"errors"
	"math/big"
	"strings"

	"erp-api/models"
)

type PurchaseOrderRepository struct{ DB *sql.DB }

var (
	ErrPurchaseOrderSource = errors.New("purchase request is unavailable for ordering")
	ErrPurchaseOrderQty    = errors.New("purchase order quantity exceeds the remaining purchase request quantity")
)

func NewPurchaseOrderRepository(db *sql.DB) *PurchaseOrderRepository {
	return &PurchaseOrderRepository{DB: db}
}

func (r *PurchaseOrderRepository) Sources(companyID uint64) ([]models.PurchaseOrderSourceRequest, error) {
	rows, err := r.DB.Query(`SELECT h.pr_number,h.branch_id,b.name,h.supplier_id,s.name,h.pr_date,h.req_date
		FROM purchase_request_header h JOIN branches b ON b.id=h.branch_id JOIN suppliers s ON s.id=h.supplier_id
		WHERE b.company_id=? AND h.status IN ('open','partially_ordered')
		ORDER BY h.pr_date DESC,h.id DESC`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]models.PurchaseOrderSourceRequest, 0)
	for rows.Next() {
		var source models.PurchaseOrderSourceRequest
		if err := rows.Scan(&source.PRNumber, &source.BranchID, &source.BranchName, &source.SupplierID, &source.SupplierName, &source.PRDate, &source.ReqDate); err != nil {
			return nil, err
		}
		result = append(result, source)
	}
	return result, rows.Err()
}

func (r *PurchaseOrderRepository) Source(companyID uint64, prNumber string) (*models.PurchaseOrderSourceRequest, error) {
	var source models.PurchaseOrderSourceRequest
	err := r.DB.QueryRow(`SELECT h.pr_number,h.branch_id,b.name,h.supplier_id,s.name,h.pr_date,h.req_date
		FROM purchase_request_header h JOIN branches b ON b.id=h.branch_id JOIN suppliers s ON s.id=h.supplier_id
		WHERE b.company_id=? AND h.pr_number=? AND h.status IN ('open','partially_ordered')`, companyID, prNumber).
		Scan(&source.PRNumber, &source.BranchID, &source.BranchName, &source.SupplierID, &source.SupplierName, &source.PRDate, &source.ReqDate)
	if err != nil {
		return nil, err
	}
	source.Lines, err = remainingPRLines(r.DB, prNumber)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *PurchaseOrderRepository) List(companyID uint64, status, search string, limit, offset int) ([]models.PurchaseOrder, int, error) {
	where := []string{"b.company_id=?"}
	args := []any{companyID}
	if status != "" {
		where = append(where, "h.status=?")
		args = append(args, status)
	}
	if search != "" {
		where = append(where, "(h.po_number LIKE ? OR h.ref_num LIKE ? OR s.name LIKE ? OR b.name LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like, like, like)
	}
	base := ` FROM purchase_order_header h JOIN branches b ON b.id=h.branch_id JOIN suppliers s ON s.id=h.supplier_id WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.Query(`SELECT h.id,h.branch_id,b.name,h.supplier_id,s.name,h.ref_num,h.po_number,h.po_date,h.due_date,h.status,h.total_qty,h.total_price,h.created_by,h.updated_by,h.created_at,h.updated_at`+base+` ORDER BY h.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]models.PurchaseOrder, 0)
	for rows.Next() {
		po, err := scanPurchaseOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, po)
	}
	return list, total, rows.Err()
}

func (r *PurchaseOrderRepository) Get(companyID, id uint64) (*models.PurchaseOrder, error) {
	row := r.DB.QueryRow(`SELECT h.id,h.branch_id,b.name,h.supplier_id,s.name,h.ref_num,h.po_number,h.po_date,h.due_date,h.status,h.total_qty,h.total_price,h.created_by,h.updated_by,h.created_at,h.updated_at
		FROM purchase_order_header h JOIN branches b ON b.id=h.branch_id JOIN suppliers s ON s.id=h.supplier_id
		WHERE b.company_id=? AND h.id=?`, companyID, id)
	po, err := scanPurchaseOrder(row)
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.Query(`SELECT d.id,d.po_header_id,d.item_id,i.sku,i.name,i.unit,d.qty,d.price
		FROM purchase_order_detail d JOIN items i ON i.id=d.item_id WHERE d.po_header_id=? ORDER BY d.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	po.Lines = make([]models.PurchaseOrderLine, 0)
	for rows.Next() {
		var line models.PurchaseOrderLine
		if err := rows.Scan(&line.ID, &line.POHeaderID, &line.ItemID, &line.SKU, &line.ItemName, &line.Unit, &line.Qty, &line.Price); err != nil {
			return nil, err
		}
		po.Lines = append(po.Lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &po, nil
}

func (r *PurchaseOrderRepository) Create(companyID, userID uint64, number, prNumber, poDate, dueDate string, lines []models.PurchaseOrderLine) (uint64, error) {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var branchID, supplierID uint64
	var status string
	err = tx.QueryRow(`SELECT h.branch_id,h.supplier_id,h.status FROM purchase_request_header h
		JOIN branches b ON b.id=h.branch_id WHERE b.company_id=? AND h.pr_number=? FOR UPDATE`, companyID, prNumber).
		Scan(&branchID, &supplierID, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrPurchaseOrderSource
		}
		return 0, err
	}
	if status != "open" && status != "partially_ordered" {
		return 0, ErrPurchaseOrderSource
	}
	remaining, err := remainingPRLines(tx, prNumber)
	if err != nil {
		return 0, err
	}
	available := make(map[uint64]string, len(remaining))
	for _, line := range remaining {
		available[line.ItemID] = line.RemainingQty
	}
	requested := make(map[uint64]string)
	for _, line := range lines {
		requested[line.ItemID] = addDecimalStrings(requested[line.ItemID], line.Qty)
	}
	for itemID, qty := range requested {
		if remainingQty, ok := available[itemID]; !ok || decimalExceeds(qty, remainingQty) {
			return 0, ErrPurchaseOrderQty
		}
	}
	result, err := tx.Exec(`INSERT INTO purchase_order_header (branch_id,supplier_id,po_number,ref_num,po_date,due_date,status,total_qty,total_price,created_by) VALUES (?,?,?,?,?,?,'ordered',0,0,?)`, branchID, supplierID, number, prNumber, poDate, dueDate, userID)
	if err != nil {
		return 0, err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	poID := uint64(lastID)
	for _, line := range lines {
		if _, ok := available[line.ItemID]; !ok {
			return 0, ErrPurchaseOrderQty
		}
		if _, err := tx.Exec(`INSERT INTO purchase_order_detail (po_header_id,item_id,qty,price) VALUES (?,?,?,?)`, poID, line.ItemID, line.Qty, line.Price); err != nil {
			return 0, err
		}
	}
	if err := refreshPOTotals(tx, poID); err != nil {
		return 0, err
	}
	if err := refreshPRStatus(tx, prNumber); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return poID, nil
}

func (r *PurchaseOrderRepository) Cancel(companyID, id, userID uint64) error {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var prNumber, status string
	err = tx.QueryRow(`SELECT h.ref_num,h.status FROM purchase_order_header h JOIN branches b ON b.id=h.branch_id
		WHERE b.company_id=? AND h.id=? FOR UPDATE`, companyID, id).Scan(&prNumber, &status)
	if err != nil {
		return err
	}
	if status == "cancelled" || status == "partially_received" || status == "received" {
		return ErrPurchaseOrderSource
	}
	var lockedPR string
	if err := tx.QueryRow(`SELECT pr_number FROM purchase_request_header WHERE pr_number=? FOR UPDATE`, prNumber).Scan(&lockedPR); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE purchase_order_header SET status='cancelled',updated_by=? WHERE id=?`, userID, id); err != nil {
		return err
	}
	if err := refreshPRStatus(tx, prNumber); err != nil {
		return err
	}
	return tx.Commit()
}

func scanPurchaseOrder(row scanner) (models.PurchaseOrder, error) {
	var po models.PurchaseOrder
	var updatedBy sql.NullInt64
	var updatedAt sql.NullTime
	err := row.Scan(&po.ID, &po.BranchID, &po.BranchName, &po.SupplierID, &po.SupplierName, &po.PRNumber, &po.PONumber, &po.PODate, &po.DueDate, &po.Status, &po.TotalQty, &po.TotalPrice, &po.CreatedBy, &updatedBy, &po.CreatedAt, &updatedAt)
	if err != nil {
		return po, err
	}
	if updatedBy.Valid {
		value := uint64(updatedBy.Int64)
		po.UpdatedBy = &value
	}
	if updatedAt.Valid {
		po.UpdatedAt = &updatedAt.Time
	}
	return po, nil
}

type queryer interface {
	Query(string, ...any) (*sql.Rows, error)
	QueryRow(string, ...any) *sql.Row
	Exec(string, ...any) (sql.Result, error)
}

func remainingPRLines(q queryer, prNumber string) ([]models.PurchaseOrderLine, error) {
	rows, err := q.Query(`SELECT d.item_id,i.sku,i.name,i.unit,CAST(SUM(d.qty) AS CHAR),
		CAST(COALESCE(o.ordered_qty,0) AS CHAR),CAST(SUM(d.qty)-COALESCE(o.ordered_qty,0) AS CHAR),
		CAST(ROUND(SUM(d.qty*d.price)/NULLIF(SUM(d.qty),0),2) AS CHAR)
		FROM purchase_request_detail d JOIN items i ON i.id=d.item_id
		LEFT JOIN (SELECT pd.item_id,SUM(pd.qty) AS ordered_qty FROM purchase_order_detail pd
			JOIN purchase_order_header ph ON ph.id=pd.po_header_id
			WHERE ph.ref_num=? AND ph.status IN ('ordered','partially_received','received') GROUP BY pd.item_id) o ON o.item_id=d.item_id
		WHERE d.pr_number=? GROUP BY d.item_id,i.sku,i.name,i.unit,o.ordered_qty
		HAVING SUM(d.qty)-COALESCE(o.ordered_qty,0)>0 ORDER BY i.name`, prNumber, prNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lines := make([]models.PurchaseOrderLine, 0)
	for rows.Next() {
		var line models.PurchaseOrderLine
		if err := rows.Scan(&line.ItemID, &line.SKU, &line.ItemName, &line.Unit, &line.RequestedQty, &line.OrderedQty, &line.RemainingQty, &line.Price); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, rows.Err()
}

func refreshPOTotals(tx *sql.Tx, id uint64) error {
	_, err := tx.Exec(`UPDATE purchase_order_header SET total_qty=(SELECT COALESCE(SUM(qty),0) FROM purchase_order_detail WHERE po_header_id=?),
		total_price=(SELECT COALESCE(SUM(qty*price),0) FROM purchase_order_detail WHERE po_header_id=?) WHERE id=?`, id, id, id)
	return err
}

func refreshPRStatus(tx *sql.Tx, prNumber string) error {
	var remaining int
	err := tx.QueryRow(`SELECT COUNT(*) FROM (
		SELECT d.item_id FROM purchase_request_detail d LEFT JOIN (
			SELECT pd.item_id,SUM(pd.qty) qty FROM purchase_order_detail pd JOIN purchase_order_header ph ON ph.id=pd.po_header_id
			WHERE ph.ref_num=? AND ph.status IN ('ordered','partially_received','received') GROUP BY pd.item_id
		) o ON o.item_id=d.item_id WHERE d.pr_number=? GROUP BY d.item_id HAVING SUM(d.qty)>COALESCE(MAX(o.qty),0)
	) remaining_items`, prNumber, prNumber).Scan(&remaining)
	if err != nil {
		return err
	}
	status := "ordered"
	if remaining > 0 {
		var ordered int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM purchase_order_header WHERE ref_num=? AND status IN ('ordered','partially_received','received')`, prNumber).Scan(&ordered); err != nil {
			return err
		}
		status = "open"
		if ordered > 0 {
			status = "partially_ordered"
		}
	}
	_, err = tx.Exec(`UPDATE purchase_request_header SET status=? WHERE pr_number=? AND status IN ('open','partially_ordered','ordered')`, status, prNumber)
	return err
}

// These helpers accumulate repeated item rows using exact rational arithmetic.
func addDecimalStrings(left, right string) string {
	if left == "" {
		return right
	}
	a, aOK := new(big.Rat).SetString(left)
	b, bOK := new(big.Rat).SetString(right)
	if !aOK || !bOK {
		return right
	}
	return new(big.Rat).Add(a, b).FloatString(3)
}

func decimalExceeds(value, maximum string) bool {
	a, aOK := new(big.Rat).SetString(value)
	b, bOK := new(big.Rat).SetString(maximum)
	return !aOK || !bOK || a.Cmp(b) > 0
}
