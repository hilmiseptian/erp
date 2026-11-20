package repositories

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"sort"
	"strings"

	"erp-api/models"
)

type StockAdjustmentRepository struct{ DB *sql.DB }

var (
	ErrStockAdjustmentTarget         = errors.New("branch or item is unavailable")
	ErrStockAdjustmentLocked         = errors.New("stock adjustment is no longer editable")
	ErrStockAdjustmentStale          = errors.New("stock changed after this draft was saved; update the draft before posting")
	ErrStockAdjustmentOpeningBalance = errors.New("an opening balance is already posted for this branch and item")
	ErrStockAdjustmentTotal          = errors.New("stock adjustment total quantity is too large")
)

func NewStockAdjustmentRepository(db *sql.DB) *StockAdjustmentRepository {
	return &StockAdjustmentRepository{DB: db}
}

func (r *StockAdjustmentRepository) List(companyID, branchID uint64, status, adjustmentType, search string, limit, offset int) ([]models.StockAdjustment, int, error) {
	where := []string{"b.company_id=?"}
	args := []any{companyID}
	if branchID > 0 {
		where = append(where, "h.branch_id=?")
		args = append(args, branchID)
	}
	if status != "" {
		where = append(where, "h.status=?")
		args = append(args, status)
	}
	if adjustmentType != "" {
		where = append(where, "h.adjustment_type=?")
		args = append(args, adjustmentType)
	}
	if search != "" {
		where = append(where, "(h.adjustment_number LIKE ? OR b.name LIKE ? OR h.reason LIKE ?)")
		like := "%" + strings.TrimSpace(search) + "%"
		args = append(args, like, like, like)
	}
	base := ` FROM stock_adjustments_head h JOIN branches b ON b.id=h.branch_id WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.Query(`SELECT h.id,b.company_id,h.branch_id,b.name,h.adjustment_number,DATE_FORMAT(h.adjustment_date,'%Y-%m-%d'),h.adjustment_type,h.reason,h.status,CAST(h.total_qty AS SIGNED),h.created_by,h.updated_by,h.created_at,h.updated_at`+base+` ORDER BY h.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]models.StockAdjustment, 0)
	for rows.Next() {
		adjustment, err := scanStockAdjustment(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, adjustment)
	}
	return list, total, rows.Err()
}

func (r *StockAdjustmentRepository) Get(companyID, id uint64) (*models.StockAdjustment, error) {
	row := r.DB.QueryRow(`SELECT h.id,b.company_id,h.branch_id,b.name,h.adjustment_number,DATE_FORMAT(h.adjustment_date,'%Y-%m-%d'),h.adjustment_type,h.reason,h.status,CAST(h.total_qty AS SIGNED),h.created_by,h.updated_by,h.created_at,h.updated_at
		FROM stock_adjustments_head h JOIN branches b ON b.id=h.branch_id WHERE b.company_id=? AND h.id=?`, companyID, id)
	adjustment, err := scanStockAdjustment(row)
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.Query(`SELECT d.id,d.item_id,i.sku,i.name,i.unit,CAST(d.system_qty AS SIGNED),CAST(d.actual_qty AS SIGNED),d.difference_qty
		FROM stock_adjustments_detail d JOIN items i ON i.id=d.item_id WHERE d.adjustment_number=? ORDER BY d.id`, adjustment.AdjustmentNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	adjustment.Lines = make([]models.StockAdjustmentLine, 0)
	for rows.Next() {
		var line models.StockAdjustmentLine
		if err := rows.Scan(&line.ID, &line.ItemID, &line.SKU, &line.ItemName, &line.Unit, &line.SystemQty, &line.ActualQty, &line.DifferenceQty); err != nil {
			return nil, err
		}
		adjustment.Lines = append(adjustment.Lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &adjustment, nil
}

func (r *StockAdjustmentRepository) CreateDraft(companyID, userID uint64, number string, adjustment models.StockAdjustment) (uint64, error) {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := branchBelongsToCompany(tx, companyID, adjustment.BranchID); err != nil {
		return 0, err
	}
	lines, totalQty, err := snapshotAdjustmentLines(tx, companyID, adjustment.BranchID, adjustment.Lines)
	if err != nil {
		return 0, err
	}
	result, err := tx.Exec(`INSERT INTO stock_adjustments_head (branch_id,adjustment_number,adjustment_date,adjustment_type,reason,status,total_qty,created_by)
		VALUES (?,?,?,?,?,'draft',?,?)`, adjustment.BranchID, number, adjustment.AdjustmentDate, adjustment.AdjustmentType, adjustment.Reason, totalQty, userID)
	if err != nil {
		return 0, err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := insertAdjustmentLines(tx, number, lines); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return uint64(lastID), nil
}

func (r *StockAdjustmentRepository) UpdateDraft(companyID, userID, id uint64, adjustment models.StockAdjustment) error {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var number, status string
	err = tx.QueryRow(`SELECT h.adjustment_number,h.status FROM stock_adjustments_head h JOIN branches b ON b.id=h.branch_id WHERE b.company_id=? AND h.id=? FOR UPDATE`, companyID, id).Scan(&number, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return sql.ErrNoRows
	}
	if err != nil {
		return err
	}
	if status != "draft" {
		return ErrStockAdjustmentLocked
	}
	if err := branchBelongsToCompany(tx, companyID, adjustment.BranchID); err != nil {
		return err
	}
	lines, totalQty, err := snapshotAdjustmentLines(tx, companyID, adjustment.BranchID, adjustment.Lines)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE stock_adjustments_head SET branch_id=?,adjustment_date=?,adjustment_type=?,reason=?,total_qty=?,updated_by=? WHERE id=?`,
		adjustment.BranchID, adjustment.AdjustmentDate, adjustment.AdjustmentType, adjustment.Reason, totalQty, userID, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM stock_adjustments_detail WHERE adjustment_number=?`, number); err != nil {
		return err
	}
	if err := insertAdjustmentLines(tx, number, lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *StockAdjustmentRepository) Post(companyID, userID, id uint64) error {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var adjustment models.StockAdjustment
	err = tx.QueryRow(`SELECT h.branch_id,h.adjustment_number,DATE_FORMAT(h.adjustment_date,'%Y-%m-%d'),h.adjustment_type,h.status
		FROM stock_adjustments_head h JOIN branches b ON b.id=h.branch_id WHERE b.company_id=? AND h.id=? FOR UPDATE`, companyID, id).
		Scan(&adjustment.BranchID, &adjustment.AdjustmentNumber, &adjustment.AdjustmentDate, &adjustment.AdjustmentType, &adjustment.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return sql.ErrNoRows
	}
	if err != nil {
		return err
	}
	if adjustment.Status != "draft" {
		return ErrStockAdjustmentLocked
	}
	rows, err := tx.Query(`SELECT item_id,CAST(system_qty AS SIGNED),CAST(actual_qty AS SIGNED),difference_qty
		FROM stock_adjustments_detail WHERE adjustment_number=? ORDER BY item_id FOR UPDATE`, adjustment.AdjustmentNumber)
	if err != nil {
		return err
	}
	lines := make([]models.StockAdjustmentLine, 0)
	for rows.Next() {
		var line models.StockAdjustmentLine
		if err := rows.Scan(&line.ItemID, &line.SystemQty, &line.ActualQty, &line.DifferenceQty); err != nil {
			rows.Close()
			return err
		}
		lines = append(lines, line)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, line := range lines {
		currentQty, err := lockStockQtyTx(tx, companyID, adjustment.BranchID, line.ItemID, adjustment.AdjustmentDate)
		if err != nil {
			return err
		}
		if currentQty != line.SystemQty {
			return ErrStockAdjustmentStale
		}
		if adjustment.AdjustmentType == "opening_balance" {
			var posted int
			if err := tx.QueryRow(`SELECT COUNT(*) FROM stock_adjustments_detail d JOIN stock_adjustments_head h ON h.adjustment_number=d.adjustment_number
				WHERE h.branch_id=? AND h.adjustment_type='opening_balance' AND h.status='posted' AND d.item_id=?`, adjustment.BranchID, line.ItemID).Scan(&posted); err != nil {
				return err
			}
			if posted > 0 {
				return ErrStockAdjustmentOpeningBalance
			}
		}
		if line.DifferenceQty == 0 {
			continue
		}
		var hpp string
		if err := tx.QueryRow(`SELECT CAST(cost_price AS CHAR) FROM items WHERE company_id=? AND id=?`, companyID, line.ItemID).Scan(&hpp); err != nil {
			return err
		}
		inQty, outQty := int64(0), int64(0)
		if line.DifferenceQty > 0 {
			inQty = line.DifferenceQty
		} else {
			outQty = -line.DifferenceQty
		}
		transType := "stock_adjustment"
		if adjustment.AdjustmentType == "opening_balance" {
			transType = "opening_balance"
		}
		if err := applyStockMovementTx(tx, companyID, adjustment.BranchID, line.ItemID, adjustment.AdjustmentNumber, transType, adjustment.AdjustmentDate, inQty, outQty, hpp); err != nil {
			return err
		}
	}
	_, err = tx.Exec(`UPDATE stock_adjustments_head SET status='posted',updated_by=? WHERE id=? AND status='draft'`, userID, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *StockAdjustmentRepository) Cancel(companyID, userID, id uint64) error {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	err = tx.QueryRow(`SELECT h.status FROM stock_adjustments_head h JOIN branches b ON b.id=h.branch_id WHERE b.company_id=? AND h.id=? FOR UPDATE`, companyID, id).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return sql.ErrNoRows
	}
	if err != nil {
		return err
	}
	if status != "draft" {
		return ErrStockAdjustmentLocked
	}
	if _, err := tx.Exec(`UPDATE stock_adjustments_head SET status='cancelled',updated_by=? WHERE id=?`, userID, id); err != nil {
		return err
	}
	return tx.Commit()
}

func branchBelongsToCompany(q interface{ QueryRow(string, ...any) *sql.Row }, companyID, branchID uint64) error {
	var count int
	if err := q.QueryRow(`SELECT COUNT(*) FROM branches WHERE id=? AND company_id=?`, branchID, companyID).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return ErrStockAdjustmentTarget
	}
	return nil
}

func snapshotAdjustmentLines(tx *sql.Tx, companyID, branchID uint64, requested []models.StockAdjustmentLine) ([]models.StockAdjustmentLine, int64, error) {
	lines := append([]models.StockAdjustmentLine(nil), requested...)
	sort.Slice(lines, func(i, j int) bool { return lines[i].ItemID < lines[j].ItemID })
	var totalQty int64
	for index := range lines {
		line := &lines[index]
		var systemQty int64
		err := tx.QueryRow(`SELECT CAST(COALESCE(sp.qty,0) AS SIGNED) FROM items i LEFT JOIN stock_period sp ON sp.company_id=i.company_id AND sp.branch_id=? AND sp.item_id=i.id
			WHERE i.company_id=? AND i.id=? AND i.is_active=1`, branchID, companyID, line.ItemID).Scan(&systemQty)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, ErrStockAdjustmentTarget
		}
		if err != nil {
			return nil, 0, err
		}
		line.SystemQty = systemQty
		line.DifferenceQty = line.ActualQty - systemQty
		absolute := line.DifferenceQty
		if absolute < 0 {
			absolute = -absolute
		}
		if totalQty > math.MaxInt64-absolute {
			return nil, 0, ErrStockAdjustmentTotal
		}
		totalQty += absolute
	}
	return lines, totalQty, nil
}

func insertAdjustmentLines(tx *sql.Tx, number string, lines []models.StockAdjustmentLine) error {
	for _, line := range lines {
		if _, err := tx.Exec(`INSERT INTO stock_adjustments_detail (adjustment_number,item_id,system_qty,actual_qty,difference_qty) VALUES (?,?,?,?,?)`,
			number, line.ItemID, line.SystemQty, line.ActualQty, line.DifferenceQty); err != nil {
			return err
		}
	}
	return nil
}

func lockStockQtyTx(tx *sql.Tx, companyID, branchID, itemID uint64, movementDate string) (int64, error) {
	if _, err := tx.Exec(`INSERT INTO stock_period (company_id,branch_id,item_id,stock_date,qty) VALUES (?,?,?,?,0) ON DUPLICATE KEY UPDATE id=id`, companyID, branchID, itemID, movementDate); err != nil {
		return 0, err
	}
	var qty int64
	err := tx.QueryRow(`SELECT CAST(qty AS SIGNED) FROM stock_period WHERE company_id=? AND branch_id=? AND item_id=? FOR UPDATE`, companyID, branchID, itemID).Scan(&qty)
	return qty, err
}

func scanStockAdjustment(row scanner) (models.StockAdjustment, error) {
	var adjustment models.StockAdjustment
	var reason sql.NullString
	var updatedBy sql.NullInt64
	var updatedAt sql.NullTime
	err := row.Scan(&adjustment.ID, &adjustment.CompanyID, &adjustment.BranchID, &adjustment.BranchName, &adjustment.AdjustmentNumber, &adjustment.AdjustmentDate, &adjustment.AdjustmentType, &reason, &adjustment.Status, &adjustment.TotalQty, &adjustment.CreatedBy, &updatedBy, &adjustment.CreatedAt, &updatedAt)
	if err != nil {
		return adjustment, err
	}
	if reason.Valid {
		adjustment.Reason = &reason.String
	}
	if updatedBy.Valid {
		value := uint64(updatedBy.Int64)
		adjustment.UpdatedBy = &value
	}
	if updatedAt.Valid {
		adjustment.UpdatedAt = &updatedAt.Time
	}
	return adjustment, nil
}
