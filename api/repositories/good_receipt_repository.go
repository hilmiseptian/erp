package repositories

import (
	"context"
	"database/sql"
	"errors"
	"sort"
	"strings"

	"erp-api/models"
)

type GoodReceiptRepository struct{ DB *sql.DB }

var (
	ErrGoodReceiptSource = errors.New("purchase order is unavailable for receipt")
	ErrGoodReceiptQty    = errors.New("receipt quantity exceeds the remaining purchase order quantity")
)

func NewGoodReceiptRepository(db *sql.DB) *GoodReceiptRepository {
	return &GoodReceiptRepository{DB: db}
}

func (r *GoodReceiptRepository) Sources(companyID uint64) ([]models.GoodReceiptSource, error) {
	rows, err := r.DB.Query(`SELECT h.po_number,h.branch_id,b.name,h.supplier_id,s.name
		FROM purchase_order_header h JOIN branches b ON b.id=h.branch_id JOIN suppliers s ON s.id=h.supplier_id
		WHERE b.company_id=? AND h.status IN ('ordered','partially_received') ORDER BY h.due_date,h.id`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]models.GoodReceiptSource, 0)
	for rows.Next() {
		var source models.GoodReceiptSource
		if err := rows.Scan(&source.PONumber, &source.BranchID, &source.BranchName, &source.SupplierID, &source.SupplierName); err != nil {
			return nil, err
		}
		list = append(list, source)
	}
	return list, rows.Err()
}

func (r *GoodReceiptRepository) Source(companyID uint64, poNumber string) (*models.GoodReceiptSource, error) {
	var source models.GoodReceiptSource
	err := r.DB.QueryRow(`SELECT h.po_number,h.branch_id,b.name,h.supplier_id,s.name FROM purchase_order_header h
		JOIN branches b ON b.id=h.branch_id JOIN suppliers s ON s.id=h.supplier_id
		WHERE b.company_id=? AND h.po_number=? AND h.status IN ('ordered','partially_received')`, companyID, poNumber).
		Scan(&source.PONumber, &source.BranchID, &source.BranchName, &source.SupplierID, &source.SupplierName)
	if err != nil {
		return nil, err
	}
	source.Lines, err = remainingPOLines(r.DB, poNumber)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *GoodReceiptRepository) List(companyID uint64, status, search string, limit, offset int) ([]models.GoodReceipt, int, error) {
	where := []string{"b.company_id=?"}
	args := []any{companyID}
	if status != "" {
		where = append(where, "h.status=?")
		args = append(args, status)
	}
	if search != "" {
		where = append(where, "(h.gr_number LIKE ? OR h.po_number LIKE ? OR s.name LIKE ? OR b.name LIKE ?)")
		like := "%" + strings.TrimSpace(search) + "%"
		args = append(args, like, like, like, like)
	}
	base := ` FROM good_receipts_head h JOIN branches b ON b.id=h.branch_id JOIN suppliers s ON s.id=h.supplier_id WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.Query(`SELECT h.id,b.company_id,h.branch_id,b.name,h.supplier_id,s.name,h.po_number,h.gr_number,h.gr_date,h.status,h.total_qty,h.total_price,h.created_by,h.updated_by,h.created_at,h.updated_at`+base+` ORDER BY h.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]models.GoodReceipt, 0)
	for rows.Next() {
		gr, err := scanGoodReceipt(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, gr)
	}
	return list, total, rows.Err()
}

func (r *GoodReceiptRepository) Get(companyID, id uint64) (*models.GoodReceipt, error) {
	row := r.DB.QueryRow(`SELECT h.id,b.company_id,h.branch_id,b.name,h.supplier_id,s.name,h.po_number,h.gr_number,h.gr_date,h.status,h.total_qty,h.total_price,h.created_by,h.updated_by,h.created_at,h.updated_at
		FROM good_receipts_head h JOIN branches b ON b.id=h.branch_id JOIN suppliers s ON s.id=h.supplier_id WHERE b.company_id=? AND h.id=?`, companyID, id)
	gr, err := scanGoodReceipt(row)
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.Query(`SELECT d.id,d.item_id,i.sku,i.name,i.unit,d.qty,d.price FROM good_receipts_detail d JOIN items i ON i.id=d.item_id WHERE d.gr_number=? ORDER BY d.id`, gr.GRNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	gr.Lines = make([]models.GoodReceiptLine, 0)
	for rows.Next() {
		var line models.GoodReceiptLine
		if err := rows.Scan(&line.ID, &line.ItemID, &line.SKU, &line.ItemName, &line.Unit, &line.Qty, &line.Price); err != nil {
			return nil, err
		}
		gr.Lines = append(gr.Lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &gr, nil
}

func (r *GoodReceiptRepository) Create(companyID, userID uint64, number, poNumber, grDate string, lines []models.GoodReceiptLine) (uint64, error) {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var branchID, supplierID uint64
	var status string
	err = tx.QueryRow(`SELECT h.branch_id,h.supplier_id,h.status FROM purchase_order_header h JOIN branches b ON b.id=h.branch_id WHERE b.company_id=? AND h.po_number=? FOR UPDATE`, companyID, poNumber).Scan(&branchID, &supplierID, &status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrGoodReceiptSource
		}
		return 0, err
	}
	if status != "ordered" && status != "partially_received" {
		return 0, ErrGoodReceiptSource
	}
	remaining, err := remainingPOLines(tx, poNumber)
	if err != nil {
		return 0, err
	}
	available := make(map[uint64]models.GoodReceiptLine, len(remaining))
	for _, line := range remaining {
		available[line.ItemID] = line
	}
	requested := make(map[uint64]int64)
	for _, line := range lines {
		requested[line.ItemID] += line.Qty
	}
	for itemID, qty := range requested {
		line, ok := available[itemID]
		if !ok || qty > line.Remaining {
			return 0, ErrGoodReceiptQty
		}
	}
	res, err := tx.Exec(`INSERT INTO good_receipts_head (branch_id,supplier_id,po_number,gr_number,gr_date,status,total_qty,total_price,created_by) VALUES (?,?,?,?,?,'posted',0,0,?)`, branchID, supplierID, poNumber, number, grDate, userID)
	if err != nil {
		return 0, err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	grID := uint64(lastID)
	itemIDs := make([]uint64, 0, len(requested))
	for itemID := range requested {
		itemIDs = append(itemIDs, itemID)
	}
	sort.Slice(itemIDs, func(i, j int) bool { return itemIDs[i] < itemIDs[j] })
	for _, itemID := range itemIDs {
		qty := requested[itemID]
		line := available[itemID]
		if _, err := tx.Exec(`INSERT INTO good_receipts_detail (item_id,gr_number,qty,price) VALUES (?,?,?,?)`, itemID, number, qty, line.Price); err != nil {
			return 0, err
		}
		if err := applyStockMovementTx(tx, companyID, branchID, itemID, number, "good_receipt", grDate, qty, 0, line.Price); err != nil {
			return 0, err
		}
	}
	if _, err := tx.Exec(`UPDATE good_receipts_head SET total_qty=(SELECT COALESCE(SUM(qty),0) FROM good_receipts_detail WHERE gr_number=?),total_price=(SELECT COALESCE(SUM(qty*price),0) FROM good_receipts_detail WHERE gr_number=?) WHERE id=?`, number, number, grID); err != nil {
		return 0, err
	}
	if err := refreshPOReceiptStatus(tx, poNumber); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return grID, nil
}

type grScanner interface{ Scan(...any) error }

func scanGoodReceipt(row grScanner) (models.GoodReceipt, error) {
	var gr models.GoodReceipt
	var updatedBy sql.NullInt64
	var updatedAt sql.NullTime
	err := row.Scan(&gr.ID, &gr.CompanyID, &gr.BranchID, &gr.BranchName, &gr.SupplierID, &gr.SupplierName, &gr.PONumber, &gr.GRNumber, &gr.GRDate, &gr.Status, &gr.TotalQty, &gr.TotalPrice, &gr.CreatedBy, &updatedBy, &gr.CreatedAt, &updatedAt)
	if err != nil {
		return gr, err
	}
	if updatedBy.Valid {
		v := uint64(updatedBy.Int64)
		gr.UpdatedBy = &v
	}
	if updatedAt.Valid {
		gr.UpdatedAt = &updatedAt.Time
	}
	return gr, nil
}

type grQueryer interface {
	Query(string, ...any) (*sql.Rows, error)
}

func remainingPOLines(q grQueryer, poNumber string) ([]models.GoodReceiptLine, error) {
	rows, err := q.Query(`SELECT d.item_id,i.sku,i.name,i.unit,CAST(SUM(d.qty) AS SIGNED),CAST(COALESCE(r.received_qty,0) AS SIGNED),CAST(SUM(d.qty)-COALESCE(r.received_qty,0) AS SIGNED),CAST(ROUND(SUM(d.qty*d.price)/NULLIF(SUM(d.qty),0),2) AS CHAR)
		FROM purchase_order_detail d JOIN purchase_order_header h ON h.id=d.po_header_id JOIN items i ON i.id=d.item_id
		LEFT JOIN (SELECT gd.item_id,SUM(gd.qty) received_qty FROM good_receipts_detail gd JOIN good_receipts_head gh ON gh.gr_number=gd.gr_number WHERE gh.po_number=? AND gh.status='posted' GROUP BY gd.item_id) r ON r.item_id=d.item_id
		WHERE h.po_number=? GROUP BY d.item_id,i.sku,i.name,i.unit,r.received_qty HAVING SUM(d.qty)-COALESCE(r.received_qty,0)>0 ORDER BY i.name`, poNumber, poNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lines := make([]models.GoodReceiptLine, 0)
	for rows.Next() {
		var line models.GoodReceiptLine
		if err := rows.Scan(&line.ItemID, &line.SKU, &line.ItemName, &line.Unit, &line.Ordered, &line.Received, &line.Remaining, &line.Price); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, rows.Err()
}

func refreshPOReceiptStatus(tx *sql.Tx, poNumber string) error {
	var remaining int
	err := tx.QueryRow(`SELECT COUNT(*) FROM (SELECT d.item_id FROM purchase_order_detail d LEFT JOIN (SELECT gd.item_id,SUM(gd.qty) qty FROM good_receipts_detail gd JOIN good_receipts_head gh ON gh.gr_number=gd.gr_number WHERE gh.po_number=? AND gh.status='posted' GROUP BY gd.item_id) r ON r.item_id=d.item_id JOIN purchase_order_header h ON h.id=d.po_header_id WHERE h.po_number=? GROUP BY d.item_id HAVING SUM(d.qty)>COALESCE(MAX(r.qty),0)) x`, poNumber, poNumber).Scan(&remaining)
	if err != nil {
		return err
	}
	status := "received"
	if remaining > 0 {
		status = "partially_received"
	}
	_, err = tx.Exec(`UPDATE purchase_order_header SET status=? WHERE po_number=? AND status IN ('ordered','partially_received')`, status, poNumber)
	return err
}
