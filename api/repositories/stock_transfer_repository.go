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

var (
	ErrStockTransferBranch   = errors.New("stock transfer branch is unavailable")
	ErrStockTransferItem     = errors.New("stock transfer item is unavailable")
	ErrStockTransferLocked   = errors.New("stock transfer is no longer editable")
	ErrStockTransferQty      = errors.New("stock transfer quantity exceeds the remaining quantity")
	ErrStockTransferDispatch = errors.New("goods delivery is unavailable for transfer receipt")
)

type StockTransferRepository struct{ DB *sql.DB }

func NewStockTransferRepository(db *sql.DB) *StockTransferRepository {
	return &StockTransferRepository{DB: db}
}

func (r *StockTransferRepository) Branches(companyID uint64) ([]models.StockTransferBranch, error) {
	rows, err := r.DB.Query(`SELECT id,name,code FROM branches WHERE company_id=? AND is_active=1 ORDER BY name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	branches := make([]models.StockTransferBranch, 0)
	for rows.Next() {
		var branch models.StockTransferBranch
		if err := rows.Scan(&branch.ID, &branch.Name, &branch.Code); err != nil {
			return nil, err
		}
		branches = append(branches, branch)
	}
	return branches, rows.Err()
}

func (r *StockTransferRepository) Items(companyID, branchID uint64, search string) ([]models.StockTransferItem, error) {
	args := []any{branchID, companyID, companyID}
	filter := ""
	if search != "" {
		filter = " AND (i.name LIKE ? OR i.sku LIKE ?)"
		like := "%" + search + "%"
		args = append(args, like, like)
	}
	rows, err := r.DB.Query(`SELECT i.id,i.sku,i.name,i.unit,CAST(COALESCE(sp.qty,0) AS SIGNED)
		FROM items i JOIN branches b ON b.id=? AND b.company_id=? AND b.is_active=1
		LEFT JOIN stock_period sp ON sp.company_id=b.company_id AND sp.branch_id=b.id AND sp.item_id=i.id
		WHERE i.company_id=? AND i.is_active=1 AND COALESCE(sp.qty,0)>0`+filter+` ORDER BY i.name`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.StockTransferItem, 0)
	for rows.Next() {
		var item models.StockTransferItem
		if err := rows.Scan(&item.ID, &item.SKU, &item.Name, &item.Unit, &item.StockQty); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *StockTransferRepository) List(companyID uint64, status, search string, limit, offset int) ([]models.StockTransfer, int, error) {
	where := []string{"h.company_id=?"}
	args := []any{companyID}
	if status != "" {
		where = append(where, "h.status=?")
		args = append(args, status)
	}
	if search != "" {
		where = append(where, "(h.transfer_number LIKE ? OR source.name LIKE ? OR destination.name LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like, like)
	}
	base := ` FROM stock_transfer_header h JOIN branches source ON source.id=h.from_branch_id
		JOIN branches destination ON destination.id=h.to_branch_id WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.Query(`SELECT h.id,h.company_id,h.from_branch_id,source.name,h.to_branch_id,destination.name,
		h.transfer_number,DATE_FORMAT(h.transfer_date,'%Y-%m-%d'),h.status,h.total_qty,h.notes,h.created_by,h.updated_by,h.created_at,h.updated_at`+
		base+` ORDER BY h.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	transfers := make([]models.StockTransfer, 0)
	for rows.Next() {
		transfer, err := scanStockTransfer(rows)
		if err != nil {
			return nil, 0, err
		}
		transfers = append(transfers, transfer)
	}
	return transfers, total, rows.Err()
}

func (r *StockTransferRepository) Get(companyID, id uint64) (*models.StockTransfer, error) {
	row := r.DB.QueryRow(`SELECT h.id,h.company_id,h.from_branch_id,source.name,h.to_branch_id,destination.name,
		h.transfer_number,DATE_FORMAT(h.transfer_date,'%Y-%m-%d'),h.status,h.total_qty,h.notes,h.created_by,h.updated_by,h.created_at,h.updated_at
		FROM stock_transfer_header h JOIN branches source ON source.id=h.from_branch_id
		JOIN branches destination ON destination.id=h.to_branch_id WHERE h.company_id=? AND h.id=?`, companyID, id)
	transfer, err := scanStockTransfer(row)
	if err != nil {
		return nil, err
	}
	transfer.Lines, err = r.transferLines(companyID, transfer)
	if err != nil {
		return nil, err
	}
	transfer.Dispatches, err = r.dispatches(companyID, transfer)
	if err != nil {
		return nil, err
	}
	transfer.Receipts, err = r.receipts(companyID, transfer.ID)
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (r *StockTransferRepository) CreateDraft(companyID, userID uint64, number string, transfer models.StockTransfer) (uint64, error) {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if err := validateTransferBranches(tx, companyID, transfer.FromBranchID, transfer.ToBranchID); err != nil {
		return 0, err
	}
	if err := validateTransferItems(tx, companyID, transfer.Lines); err != nil {
		return 0, err
	}
	total, err := transferQtyTotal(transfer.Lines)
	if err != nil {
		return 0, err
	}
	result, err := tx.Exec(`INSERT INTO stock_transfer_header(company_id,from_branch_id,to_branch_id,transfer_number,transfer_date,status,total_qty,notes,created_by)
		VALUES(?,?,?,?,?,'draft',?,?,?)`, companyID, transfer.FromBranchID, transfer.ToBranchID, number, transfer.TransferDate, total, transfer.Notes, userID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	if err := insertTransferLines(tx, uint64(id), transfer.Lines); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *StockTransferRepository) UpdateDraft(companyID, userID, id uint64, transfer models.StockTransfer) error {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	err = tx.QueryRow(`SELECT status FROM stock_transfer_header WHERE company_id=? AND id=? FOR UPDATE`, companyID, id).Scan(&status)
	if err != nil {
		return err
	}
	if status != "draft" {
		return ErrStockTransferLocked
	}
	if err := validateTransferBranches(tx, companyID, transfer.FromBranchID, transfer.ToBranchID); err != nil {
		return err
	}
	if err := validateTransferItems(tx, companyID, transfer.Lines); err != nil {
		return err
	}
	total, err := transferQtyTotal(transfer.Lines)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE stock_transfer_header SET from_branch_id=?,to_branch_id=?,transfer_date=?,total_qty=?,notes=?,updated_by=? WHERE id=?`,
		transfer.FromBranchID, transfer.ToBranchID, transfer.TransferDate, total, transfer.Notes, userID, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM stock_transfer_detail WHERE transfer_header_id=?`, id); err != nil {
		return err
	}
	if err := insertTransferLines(tx, id, transfer.Lines); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *StockTransferRepository) Cancel(companyID, userID, id uint64) error {
	result, err := r.DB.Exec(`UPDATE stock_transfer_header SET status='cancelled',updated_by=? WHERE company_id=? AND id=? AND status='draft'`, userID, companyID, id)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		var status string
		if err := r.DB.QueryRow(`SELECT status FROM stock_transfer_header WHERE company_id=? AND id=?`, companyID, id).Scan(&status); err != nil {
			return err
		}
		return ErrStockTransferLocked
	}
	return nil
}

func (r *StockTransferRepository) Dispatch(companyID, userID, id uint64, gdNumber, dispatchDate string, requested []models.StockTransferLine) (uint64, error) {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var transferNumber, status string
	var branchID uint64
	err = tx.QueryRow(`SELECT transfer_number,from_branch_id,status FROM stock_transfer_header WHERE company_id=? AND id=? FOR UPDATE`, companyID, id).
		Scan(&transferNumber, &branchID, &status)
	if err != nil {
		return 0, err
	}
	if status != "draft" && status != "in_progress" {
		return 0, ErrStockTransferLocked
	}
	remaining, err := remainingTransferLines(tx, companyID, id, transferNumber)
	if err != nil {
		return 0, err
	}
	if err := validateRequestedTransferQty(remaining, requested); err != nil {
		return 0, err
	}
	total, err := transferQtyTotal(requested)
	if err != nil {
		return 0, err
	}
	result, err := tx.Exec(`INSERT INTO goods_delivery_header(company_id,branch_id,ref_type,ref_num,gd_number,gd_date,status,total_qty,created_by)
		VALUES(?,?, 'stock_transfer', ?,?,?,'posted',?,?)`, companyID, branchID, transferNumber, gdNumber, dispatchDate, total, userID)
	if err != nil {
		return 0, err
	}
	gdID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	itemIDs := sortedTransferItemIDs(requested)
	for _, itemID := range itemIDs {
		qty := findTransferQty(requested, itemID)
		if err := lockDeliveryStock(tx, companyID, branchID, itemID, dispatchDate, qty); err != nil {
			return 0, err
		}
		hpp, err := currentBranchHPP(tx, companyID, branchID, itemID)
		if err != nil {
			return 0, err
		}
		if _, err := tx.Exec(`INSERT INTO goods_delivery_detail(gd_header_id,item_id,qty) VALUES(?,?,?)`, gdID, itemID, qty); err != nil {
			return 0, err
		}
		if err := applyStockMovementTx(tx, companyID, branchID, itemID, gdNumber, "good_delivery", dispatchDate, 0, qty, hpp); err != nil {
			return 0, err
		}
	}
	if _, err := tx.Exec(`UPDATE stock_transfer_header SET status='in_progress',updated_by=? WHERE id=?`, userID, id); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return uint64(gdID), nil
}

func (r *StockTransferRepository) Receive(companyID, userID, transferID, gdID uint64, receiptNumber, receiptDate string, requested []models.StockTransferLine) (uint64, error) {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var transferNumber, status string
	var fromBranchID, toBranchID uint64
	err = tx.QueryRow(`SELECT transfer_number,from_branch_id,to_branch_id,status FROM stock_transfer_header WHERE company_id=? AND id=? FOR UPDATE`, companyID, transferID).
		Scan(&transferNumber, &fromBranchID, &toBranchID, &status)
	if err != nil {
		return 0, err
	}
	if status != "in_progress" {
		return 0, ErrStockTransferLocked
	}
	var gdNumber string
	var gdStatus string
	err = tx.QueryRow(`SELECT gd_number,status FROM goods_delivery_header WHERE id=? AND company_id=? AND branch_id=? AND ref_type='stock_transfer' AND ref_num=? FOR UPDATE`, gdID, companyID, fromBranchID, transferNumber).
		Scan(&gdNumber, &gdStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrStockTransferDispatch
	}
	if err != nil {
		return 0, err
	}
	if gdStatus != "posted" {
		return 0, ErrStockTransferDispatch
	}
	available, err := remainingDispatchLines(tx, companyID, gdID)
	if err != nil {
		return 0, err
	}
	if err := validateRequestedDispatchQty(available, requested); err != nil {
		return 0, err
	}
	total, err := transferQtyTotal(requested)
	if err != nil {
		return 0, err
	}
	result, err := tx.Exec(`INSERT INTO stock_transfer_receipt_header(company_id,branch_id,transfer_header_id,gd_header_id,receipt_number,receipt_date,status,total_qty,received_by)
		VALUES(?,?,?,?,?,?,'posted',?,?)`, companyID, toBranchID, transferID, gdID, receiptNumber, receiptDate, total, userID)
	if err != nil {
		return 0, err
	}
	receiptID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	itemIDs := sortedTransferItemIDs(requested)
	for _, itemID := range itemIDs {
		qty := findTransferQty(requested, itemID)
		var hpp string
		if err := tx.QueryRow(`SELECT CAST(hpp AS CHAR) FROM stock_card WHERE company_id=? AND branch_id=? AND item_id=? AND ref_num=? AND trans_type='good_delivery' AND out_qty>0 ORDER BY id DESC LIMIT 1`, companyID, fromBranchID, itemID, gdNumber).Scan(&hpp); err != nil {
			return 0, err
		}
		if _, err := tx.Exec(`INSERT INTO stock_transfer_receipt_detail(receipt_header_id,item_id,qty) VALUES(?,?,?)`, receiptID, itemID, qty); err != nil {
			return 0, err
		}
		if err := applyStockMovementTx(tx, companyID, toBranchID, itemID, receiptNumber, "stock_transfer_in", receiptDate, qty, 0, hpp); err != nil {
			return 0, err
		}
	}
	var received, planned int64
	if err := tx.QueryRow(`SELECT COALESCE(SUM(rd.qty),0) FROM stock_transfer_receipt_detail rd JOIN stock_transfer_receipt_header rh ON rh.id=rd.receipt_header_id WHERE rh.transfer_header_id=? AND rh.status='posted'`, transferID).Scan(&received); err != nil {
		return 0, err
	}
	if err := tx.QueryRow(`SELECT total_qty FROM stock_transfer_header WHERE id=?`, transferID).Scan(&planned); err != nil {
		return 0, err
	}
	newStatus := "in_progress"
	if received == planned {
		newStatus = "completed"
	}
	if _, err := tx.Exec(`UPDATE stock_transfer_header SET status=?,updated_by=? WHERE id=?`, newStatus, userID, transferID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return uint64(receiptID), nil
}

func (r *StockTransferRepository) transferLines(companyID uint64, transfer models.StockTransfer) ([]models.StockTransferLine, error) {
	rows, err := r.DB.Query(`SELECT d.id,d.item_id,i.sku,i.name,i.unit,CAST(d.qty AS SIGNED),
		CAST(COALESCE(g.qty,0) AS SIGNED),CAST(COALESCE(r.qty,0) AS SIGNED)
		FROM stock_transfer_detail d JOIN items i ON i.id=d.item_id
		LEFT JOIN (SELECT h.ref_num,gd.item_id,SUM(gd.qty) qty FROM goods_delivery_header h JOIN goods_delivery_detail gd ON gd.gd_header_id=h.id
			WHERE h.company_id=? AND h.ref_type='stock_transfer' AND h.status='posted' GROUP BY h.ref_num,gd.item_id) g ON g.ref_num=? AND g.item_id=d.item_id
		LEFT JOIN (SELECT rh.transfer_header_id,rd.item_id,SUM(rd.qty) qty FROM stock_transfer_receipt_header rh JOIN stock_transfer_receipt_detail rd ON rd.receipt_header_id=rh.id
			WHERE rh.company_id=? AND rh.status='posted' GROUP BY rh.transfer_header_id,rd.item_id) r ON r.transfer_header_id=d.transfer_header_id AND r.item_id=d.item_id
		WHERE d.transfer_header_id=? ORDER BY i.name`, companyID, transfer.TransferNumber, companyID, transfer.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lines := make([]models.StockTransferLine, 0)
	for rows.Next() {
		var line models.StockTransferLine
		if err := rows.Scan(&line.ID, &line.ItemID, &line.SKU, &line.ItemName, &line.Unit, &line.Qty, &line.DispatchedQty, &line.ReceivedQty); err != nil {
			return nil, err
		}
		line.InTransitQty = line.DispatchedQty - line.ReceivedQty
		line.RemainingDispatchQty = line.Qty - line.DispatchedQty
		lines = append(lines, line)
	}
	return lines, rows.Err()
}

func (r *StockTransferRepository) dispatches(companyID uint64, transfer models.StockTransfer) ([]models.StockTransferDispatch, error) {
	rows, err := r.DB.Query(`SELECT h.id,h.gd_number,DATE_FORMAT(h.gd_date,'%Y-%m-%d'),h.total_qty,d.item_id,i.sku,i.name,i.unit,CAST(d.qty AS SIGNED),CAST(COALESCE(SUM(rd.qty),0) AS SIGNED)
		FROM goods_delivery_header h JOIN goods_delivery_detail d ON d.gd_header_id=h.id JOIN items i ON i.id=d.item_id
		LEFT JOIN stock_transfer_receipt_header rh ON rh.gd_header_id=h.id AND rh.status='posted'
		LEFT JOIN stock_transfer_receipt_detail rd ON rd.receipt_header_id=rh.id AND rd.item_id=d.item_id
		WHERE h.company_id=? AND h.ref_type='stock_transfer' AND h.ref_num=? AND h.status='posted'
		GROUP BY h.id,h.gd_number,h.gd_date,h.total_qty,d.item_id,i.sku,i.name,i.unit,d.qty ORDER BY h.id,d.item_id`, companyID, transfer.TransferNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	dispatchMap := make(map[uint64]*models.StockTransferDispatch)
	order := make([]uint64, 0)
	for rows.Next() {
		var gdID uint64
		var dispatch models.StockTransferDispatch
		var line models.StockTransferDispatchLine
		if err := rows.Scan(&gdID, &dispatch.GDNumber, &dispatch.GDDate, &dispatch.TotalQty, &line.ItemID, &line.SKU, &line.ItemName, &line.Unit, &line.Qty, &line.ReceivedQty); err != nil {
			return nil, err
		}
		dispatch.GDHeaderID = gdID
		line.Remaining = line.Qty - line.ReceivedQty
		if dispatchMap[gdID] == nil {
			dispatch.Lines = make([]models.StockTransferDispatchLine, 0)
			dispatchMap[gdID] = &dispatch
			order = append(order, gdID)
		}
		dispatchMap[gdID].Lines = append(dispatchMap[gdID].Lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	dispatches := make([]models.StockTransferDispatch, 0, len(order))
	for _, id := range order {
		dispatches = append(dispatches, *dispatchMap[id])
	}
	return dispatches, nil
}

func (r *StockTransferRepository) receipts(companyID, transferID uint64) ([]models.StockTransferReceipt, error) {
	rows, err := r.DB.Query(`SELECT rh.id,rh.receipt_number,rh.gd_header_id,gh.gd_number,DATE_FORMAT(rh.receipt_date,'%Y-%m-%d'),rh.total_qty
		FROM stock_transfer_receipt_header rh JOIN goods_delivery_header gh ON gh.id=rh.gd_header_id
		WHERE rh.company_id=? AND rh.transfer_header_id=? AND rh.status='posted' ORDER BY rh.id`, companyID, transferID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	receipts := make([]models.StockTransferReceipt, 0)
	for rows.Next() {
		var receipt models.StockTransferReceipt
		if err := rows.Scan(&receipt.ID, &receipt.ReceiptNumber, &receipt.GDHeaderID, &receipt.GDNumber, &receipt.ReceiptDate, &receipt.TotalQty); err != nil {
			return nil, err
		}
		receipts = append(receipts, receipt)
	}
	return receipts, rows.Err()
}

func remainingTransferLines(q interface {
	Query(string, ...any) (*sql.Rows, error)
}, companyID, transferID uint64, transferNumber string) (map[uint64]int64, error) {
	rows, err := q.Query(`SELECT d.item_id,CAST(d.qty-COALESCE(g.dispatched_qty,0) AS SIGNED)
		FROM stock_transfer_detail d LEFT JOIN (
			SELECT gd.item_id,SUM(gd.qty) dispatched_qty FROM goods_delivery_header h JOIN goods_delivery_detail gd ON gd.gd_header_id=h.id
			WHERE h.company_id=? AND h.ref_type='stock_transfer' AND h.ref_num=? AND h.status='posted' GROUP BY gd.item_id
		) g ON g.item_id=d.item_id WHERE d.transfer_header_id=?`, companyID, transferNumber, transferID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	remaining := make(map[uint64]int64)
	for rows.Next() {
		var itemID uint64
		var qty int64
		if err := rows.Scan(&itemID, &qty); err != nil {
			return nil, err
		}
		remaining[itemID] = qty
	}
	return remaining, rows.Err()
}

func remainingDispatchLines(q interface {
	Query(string, ...any) (*sql.Rows, error)
}, companyID, gdID uint64) (map[uint64]int64, error) {
	rows, err := q.Query(`SELECT d.item_id,CAST(d.qty-COALESCE(r.received_qty,0) AS SIGNED)
		FROM goods_delivery_detail d LEFT JOIN (
			SELECT rd.item_id,SUM(rd.qty) received_qty FROM stock_transfer_receipt_header rh JOIN stock_transfer_receipt_detail rd ON rd.receipt_header_id=rh.id
			WHERE rh.company_id=? AND rh.gd_header_id=? AND rh.status='posted' GROUP BY rd.item_id
		) r ON r.item_id=d.item_id WHERE d.gd_header_id=?`, companyID, gdID, gdID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	remaining := make(map[uint64]int64)
	for rows.Next() {
		var itemID uint64
		var qty int64
		if err := rows.Scan(&itemID, &qty); err != nil {
			return nil, err
		}
		remaining[itemID] = qty
	}
	return remaining, rows.Err()
}

func scanStockTransfer(row interface{ Scan(...any) error }) (models.StockTransfer, error) {
	var transfer models.StockTransfer
	var notes sql.NullString
	var updatedBy sql.NullInt64
	var updatedAt sql.NullTime
	err := row.Scan(&transfer.ID, &transfer.CompanyID, &transfer.FromBranchID, &transfer.FromBranchName,
		&transfer.ToBranchID, &transfer.ToBranchName, &transfer.TransferNumber, &transfer.TransferDate,
		&transfer.Status, &transfer.TotalQty, &notes, &transfer.CreatedBy, &updatedBy, &transfer.CreatedAt, &updatedAt)
	if err != nil {
		return transfer, err
	}
	if notes.Valid {
		transfer.Notes = &notes.String
	}
	if updatedBy.Valid {
		value := uint64(updatedBy.Int64)
		transfer.UpdatedBy = &value
	}
	if updatedAt.Valid {
		transfer.UpdatedAt = &updatedAt.Time
	}
	return transfer, nil
}

func validateTransferBranches(tx *sql.Tx, companyID, fromBranchID, toBranchID uint64) error {
	if fromBranchID == 0 || toBranchID == 0 || fromBranchID == toBranchID {
		return ErrStockTransferBranch
	}
	var count int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM branches WHERE company_id=? AND is_active=1 AND id IN (?,?)`, companyID, fromBranchID, toBranchID).Scan(&count); err != nil {
		return err
	}
	if count != 2 {
		return ErrStockTransferBranch
	}
	return nil
}

func validateTransferItems(tx *sql.Tx, companyID uint64, lines []models.StockTransferLine) error {
	for _, line := range lines {
		var count int
		if err := tx.QueryRow(`SELECT COUNT(*) FROM items WHERE company_id=? AND id=? AND is_active=1`, companyID, line.ItemID).Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			return ErrStockTransferItem
		}
	}
	return nil
}

func insertTransferLines(tx *sql.Tx, transferID uint64, lines []models.StockTransferLine) error {
	for _, line := range lines {
		if _, err := tx.Exec(`INSERT INTO stock_transfer_detail(transfer_header_id,item_id,qty) VALUES(?,?,?)`, transferID, line.ItemID, line.Qty); err != nil {
			return err
		}
	}
	return nil
}

func transferQtyTotal(lines []models.StockTransferLine) (int64, error) {
	var total int64
	for _, line := range lines {
		if line.Qty < 1 || total > math.MaxInt64-line.Qty {
			return 0, ErrStockTransferQty
		}
		total += line.Qty
	}
	return total, nil
}

func validateRequestedTransferQty(remaining map[uint64]int64, requested []models.StockTransferLine) error {
	return validateRequestedQty(remaining, requested)
}

func validateRequestedDispatchQty(remaining map[uint64]int64, requested []models.StockTransferLine) error {
	return validateRequestedQty(remaining, requested)
}

func validateRequestedQty(remaining map[uint64]int64, requested []models.StockTransferLine) error {
	for _, line := range requested {
		available, ok := remaining[line.ItemID]
		if !ok || line.Qty > available {
			return ErrStockTransferQty
		}
	}
	return nil
}

func sortedTransferItemIDs(lines []models.StockTransferLine) []uint64 {
	itemIDs := make([]uint64, 0, len(lines))
	for _, line := range lines {
		itemIDs = append(itemIDs, line.ItemID)
	}
	sort.Slice(itemIDs, func(i, j int) bool { return itemIDs[i] < itemIDs[j] })
	return itemIDs
}

func findTransferQty(lines []models.StockTransferLine, itemID uint64) int64 {
	for _, line := range lines {
		if line.ItemID == itemID {
			return line.Qty
		}
	}
	return 0
}
