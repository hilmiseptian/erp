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
	ErrGoodDeliverySource = errors.New("source document is unavailable for delivery")
	ErrGoodDeliveryQty    = errors.New("delivery quantity exceeds the remaining source quantity")
)

type GoodDeliveryRepository struct{ DB *sql.DB }

func NewGoodDeliveryRepository(db *sql.DB) *GoodDeliveryRepository {
	return &GoodDeliveryRepository{DB: db}
}

func (r *GoodDeliveryRepository) Sources(companyID uint64, refType string) ([]models.GoodDeliverySource, error) {
	if refType != "sales_order" {
		return nil, ErrGoodDeliverySource
	}
	rows, err := r.DB.Query(`SELECT so.so_number,so.branch_id,b.name,so.customer_name
		FROM sales_order_header so JOIN branches b ON b.id=so.branch_id
		WHERE so.company_id=? AND b.is_active=1 AND so.status IN ('open','partially_delivered')
		ORDER BY so.so_date,so.id`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]models.GoodDeliverySource, 0)
	for rows.Next() {
		var source models.GoodDeliverySource
		source.RefType = refType
		if err := rows.Scan(&source.RefNum, &source.BranchID, &source.BranchName, &source.CustomerName); err != nil {
			return nil, err
		}
		list = append(list, source)
	}
	return list, rows.Err()
}

func (r *GoodDeliveryRepository) Source(companyID uint64, refType, refNum string) (*models.GoodDeliverySource, error) {
	if refType != "sales_order" {
		return nil, ErrGoodDeliverySource
	}
	var source models.GoodDeliverySource
	source.RefType = refType
	var deliveryAddress sql.NullString
	err := r.DB.QueryRow(`SELECT so.so_number,so.branch_id,b.name,so.customer_name,so.delivery_address
		FROM sales_order_header so JOIN branches b ON b.id=so.branch_id
		WHERE so.company_id=? AND b.is_active=1 AND so.so_number=?
		AND so.status IN ('open','partially_delivered')`, companyID, refNum).
		Scan(&source.RefNum, &source.BranchID, &source.BranchName, &source.CustomerName, &deliveryAddress)
	if err != nil {
		return nil, err
	}
	if deliveryAddress.Valid {
		source.DeliveryAddress = &deliveryAddress.String
	}
	source.Lines, err = remainingSalesOrderLines(r.DB, companyID, source.RefNum)
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *GoodDeliveryRepository) List(companyID uint64, search string, limit, offset int) ([]models.GoodDelivery, int, error) {
	where := []string{"h.company_id=?"}
	args := []any{companyID}
	if search != "" {
		where = append(where, "(h.gd_number LIKE ? OR h.ref_num LIKE ? OR b.name LIKE ?)")
		like := "%" + strings.TrimSpace(search) + "%"
		args = append(args, like, like, like)
	}
	base := ` FROM goods_delivery_header h JOIN branches b ON b.id=h.branch_id WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.Query(`SELECT h.id,h.company_id,h.branch_id,b.name,h.ref_type,h.ref_num,h.gd_number,DATE_FORMAT(h.gd_date,'%Y-%m-%d'),h.status,h.total_qty,h.created_by,h.updated_by,h.created_at,h.updated_at`+base+` ORDER BY h.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]models.GoodDelivery, 0)
	for rows.Next() {
		item, err := scanGoodDelivery(rows)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, item)
	}
	return list, total, rows.Err()
}

func (r *GoodDeliveryRepository) Get(companyID, id uint64) (*models.GoodDelivery, error) {
	row := r.DB.QueryRow(`SELECT h.id,h.company_id,h.branch_id,b.name,h.ref_type,h.ref_num,h.gd_number,DATE_FORMAT(h.gd_date,'%Y-%m-%d'),h.status,h.total_qty,h.created_by,h.updated_by,h.created_at,h.updated_at
		FROM goods_delivery_header h JOIN branches b ON b.id=h.branch_id WHERE h.company_id=? AND h.id=?`, companyID, id)
	item, err := scanGoodDelivery(row)
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.Query(`SELECT d.id,d.gd_header_id,d.item_id,i.sku,i.name,i.unit,d.qty
		FROM goods_delivery_detail d JOIN items i ON i.id=d.item_id WHERE d.gd_header_id=? ORDER BY d.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	item.Lines = make([]models.GoodDeliveryLine, 0)
	for rows.Next() {
		var line models.GoodDeliveryLine
		if err := rows.Scan(&line.ID, &line.GDHeaderID, &line.ItemID, &line.SKU, &line.ItemName, &line.Unit, &line.Qty); err != nil {
			return nil, err
		}
		item.Lines = append(item.Lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *GoodDeliveryRepository) Create(companyID, userID uint64, number string, gd models.GoodDelivery) (uint64, error) {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if gd.RefType != "sales_order" {
		return 0, ErrGoodDeliverySource
	}
	var soID, branchID uint64
	var status string
	err = tx.QueryRow(`SELECT so.id,so.branch_id,so.status FROM sales_order_header so
		JOIN branches b ON b.id=so.branch_id
		WHERE so.company_id=? AND b.is_active=1 AND so.so_number=? FOR UPDATE`, companyID, gd.RefNum).
		Scan(&soID, &branchID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrGoodDeliverySource
	}
	if err != nil {
		return 0, err
	}
	if status != "open" && status != "partially_delivered" {
		return 0, ErrGoodDeliverySource
	}
	availableLines, err := remainingSalesOrderLines(tx, companyID, gd.RefNum)
	if err != nil {
		return 0, err
	}
	available := make(map[uint64]int64, len(availableLines))
	for _, line := range availableLines {
		available[line.ItemID] = line.Remaining
	}
	requested := make(map[uint64]int64)
	var totalQty int64
	for _, line := range gd.Lines {
		if line.ItemID == 0 || line.Qty < 1 || requested[line.ItemID] > math.MaxInt64-line.Qty || totalQty > math.MaxInt64-line.Qty {
			return 0, ErrGoodDeliveryQty
		}
		requested[line.ItemID] += line.Qty
		totalQty += line.Qty
	}
	if len(requested) == 0 {
		return 0, ErrGoodDeliveryQty
	}
	for itemID, qty := range requested {
		remaining, ok := available[itemID]
		if !ok || qty > remaining {
			return 0, ErrGoodDeliveryQty
		}
	}
	result, err := tx.Exec(`INSERT INTO goods_delivery_header(company_id,branch_id,ref_type,ref_num,gd_number,gd_date,status,total_qty,created_by)
		VALUES(?,?,?, ?,?,?,'posted',?,?)`, companyID, branchID, gd.RefType, gd.RefNum, number, gd.GDDate, totalQty, userID)
	if err != nil {
		return 0, err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	gdID := uint64(lastID)
	itemIDs := make([]uint64, 0, len(requested))
	for itemID := range requested {
		itemIDs = append(itemIDs, itemID)
	}
	sort.Slice(itemIDs, func(i, j int) bool { return itemIDs[i] < itemIDs[j] })
	for _, itemID := range itemIDs {
		qty := requested[itemID]
		if err := lockDeliveryStock(tx, companyID, branchID, itemID, gd.GDDate, qty); err != nil {
			return 0, err
		}
		hpp, err := currentBranchHPP(tx, companyID, branchID, itemID)
		if err != nil {
			return 0, err
		}
		if _, err := tx.Exec(`INSERT INTO goods_delivery_detail(gd_header_id,item_id,qty) VALUES(?,?,?)`, gdID, itemID, qty); err != nil {
			return 0, err
		}
		if err := applyStockMovementTx(tx, companyID, branchID, itemID, number, "good_delivery", gd.GDDate, 0, qty, hpp); err != nil {
			return 0, err
		}
	}
	if err := refreshSalesOrderDeliveryStatus(tx, companyID, soID, gd.RefNum, userID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return gdID, nil
}

func scanGoodDelivery(row scanner) (models.GoodDelivery, error) {
	var item models.GoodDelivery
	var updatedBy sql.NullInt64
	var updatedAt sql.NullTime
	err := row.Scan(&item.ID, &item.CompanyID, &item.BranchID, &item.BranchName, &item.RefType, &item.RefNum, &item.GDNumber, &item.GDDate, &item.Status, &item.TotalQty, &item.CreatedBy, &updatedBy, &item.CreatedAt, &updatedAt)
	if err != nil {
		return item, err
	}
	if updatedBy.Valid {
		value := uint64(updatedBy.Int64)
		item.UpdatedBy = &value
	}
	if updatedAt.Valid {
		item.UpdatedAt = &updatedAt.Time
	}
	return item, nil
}

type gdQueryer interface {
	Query(string, ...any) (*sql.Rows, error)
}

func remainingSalesOrderLines(q gdQueryer, companyID uint64, soNumber string) ([]models.GoodDeliveryLine, error) {
	rows, err := q.Query(`SELECT d.item_id,i.sku,i.name,i.unit,CAST(SUM(d.qty) AS SIGNED),
		CAST(COALESCE(MAX(g.delivered_qty),0) AS SIGNED),
		CAST(SUM(d.qty)-COALESCE(MAX(g.delivered_qty),0) AS SIGNED)
		FROM sales_order_detail d
		JOIN sales_order_header so ON so.id=d.so_header_id
		JOIN items i ON i.id=d.item_id AND i.company_id=so.company_id
	LEFT JOIN (SELECT h.company_id,h.ref_num,gd.item_id,SUM(gd.qty) delivered_qty
			FROM goods_delivery_detail gd JOIN goods_delivery_header h ON h.id=gd.gd_header_id
			WHERE h.ref_type='sales_order' AND h.status='posted' GROUP BY h.company_id,h.ref_num,gd.item_id) g
			ON g.company_id=so.company_id AND g.ref_num=so.so_number AND g.item_id=d.item_id
		WHERE so.company_id=? AND so.so_number=? GROUP BY d.item_id,i.sku,i.name,i.unit
		HAVING SUM(d.qty)-COALESCE(MAX(g.delivered_qty),0)>0 ORDER BY i.name`, companyID, soNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	lines := make([]models.GoodDeliveryLine, 0)
	for rows.Next() {
		var line models.GoodDeliveryLine
		if err := rows.Scan(&line.ItemID, &line.SKU, &line.ItemName, &line.Unit, &line.Ordered, &line.Delivered, &line.Remaining); err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return lines, rows.Err()
}

func lockDeliveryStock(tx *sql.Tx, companyID, branchID, itemID uint64, deliveryDate string, qty int64) error {
	if _, err := tx.Exec(`INSERT INTO stock_period(company_id,branch_id,item_id,stock_date,qty) VALUES(?,?,?,?,0)
		ON DUPLICATE KEY UPDATE id=id`, companyID, branchID, itemID, deliveryDate); err != nil {
		return err
	}
	var current int64
	if err := tx.QueryRow(`SELECT qty FROM stock_period WHERE company_id=? AND branch_id=? AND item_id=? FOR UPDATE`, companyID, branchID, itemID).Scan(&current); err != nil {
		return err
	}
	if current < qty {
		return ErrStockInsufficient
	}
	return nil
}

func currentBranchHPP(tx *sql.Tx, companyID, branchID, itemID uint64) (string, error) {
	var hpp string
	err := tx.QueryRow(`SELECT CAST(COALESCE(ROUND((SUM(in_qty*hpp)-SUM(out_qty*hpp)) /
		NULLIF(SUM(in_qty)-SUM(out_qty),0),4),0) AS CHAR) FROM stock_card
		WHERE company_id=? AND branch_id=? AND item_id=?`, companyID, branchID, itemID).Scan(&hpp)
	return hpp, err
}

func refreshSalesOrderDeliveryStatus(tx *sql.Tx, companyID, soID uint64, soNumber string, userID uint64) error {
	var remainingItems int
	err := tx.QueryRow(`SELECT COUNT(*) FROM (
		SELECT d.item_id FROM sales_order_detail d LEFT JOIN (
			SELECT gd.item_id,SUM(gd.qty) qty FROM goods_delivery_detail gd
			JOIN goods_delivery_header h ON h.id=gd.gd_header_id
			WHERE h.company_id=? AND h.ref_type='sales_order' AND h.ref_num=? AND h.status='posted' GROUP BY gd.item_id
		) shipped ON shipped.item_id=d.item_id
		WHERE d.so_header_id=? GROUP BY d.item_id HAVING SUM(d.qty)>COALESCE(MAX(shipped.qty),0)
	) outstanding_items`, companyID, soNumber, soID).Scan(&remainingItems)
	if err != nil {
		return err
	}
	status := "delivered"
	if remainingItems > 0 {
		var delivered int64
		if err := tx.QueryRow(`SELECT COALESCE(SUM(gd.qty),0) FROM goods_delivery_detail gd
			JOIN goods_delivery_header h ON h.id=gd.gd_header_id
			WHERE h.company_id=? AND h.ref_type='sales_order' AND h.ref_num=? AND h.status='posted'`, companyID, soNumber).Scan(&delivered); err != nil {
			return err
		}
		status = "open"
		if delivered > 0 {
			status = "partially_delivered"
		}
	}
	_, err = tx.Exec(`UPDATE sales_order_header SET status=?,updated_by=? WHERE id=? AND status IN ('open','partially_delivered')`, status, userID, soID)
	return err
}
