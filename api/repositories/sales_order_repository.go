package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	"erp-api/models"
	"erp-api/utils"
)

var (
	ErrSalesOrderNotFound = errors.New("sales order not found")
	ErrSalesOrderInvalid  = errors.New("invalid sales order")
	ErrSalesOrderLocked   = errors.New("sales order cannot be changed in its current state")
)

type SalesOrderRepository struct{ DB *sql.DB }

func NewSalesOrderRepository(db *sql.DB) *SalesOrderRepository { return &SalesOrderRepository{DB: db} }

func (r *SalesOrderRepository) Branches(companyID uint64) ([]models.Branch, error) {
	rows, err := r.DB.Query(`SELECT id,name,code FROM branches WHERE company_id=? AND is_active=1 ORDER BY name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]models.Branch, 0)
	for rows.Next() {
		var item models.Branch
		if err := rows.Scan(&item.ID, &item.Name, &item.Code); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *SalesOrderRepository) List(companyID uint64, status, search string, limit, offset int) ([]models.SalesOrder, int, error) {
	where := []string{"h.company_id=?"}
	args := []any{companyID}
	if status != "" {
		where = append(where, "h.status=?")
		args = append(args, status)
	}
	if search != "" {
		where = append(where, "(h.so_number LIKE ? OR h.customer_name LIKE ? OR c.customer_code LIKE ? OR b.name LIKE ?)")
		like := "%" + search + "%"
		args = append(args, like, like, like, like)
	}
	base := ` FROM sales_order_header h JOIN branches b ON b.id=h.branch_id JOIN customers c ON c.id=h.customer_id WHERE ` + strings.Join(where, " AND ")
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.Query(`SELECT h.id,h.company_id,h.branch_id,b.name,h.customer_id,h.customer_name,h.so_number,DATE_FORMAT(h.so_date,'%Y-%m-%d'),DATE_FORMAT(h.delivery_date,'%Y-%m-%d'),h.delivery_address,h.status,h.total_qty,CAST(h.total_amount AS CHAR),CAST(h.discount_amount AS CHAR),CAST(h.grand_total AS CHAR),h.notes,h.created_by,h.updated_by,h.created_at,h.updated_at`+base+` ORDER BY h.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]models.SalesOrder, 0)
	for rows.Next() {
		item, err := scanSalesOrder(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *SalesOrderRepository) Get(companyID, id uint64) (*models.SalesOrder, error) {
	row := r.DB.QueryRow(`SELECT h.id,h.company_id,h.branch_id,b.name,h.customer_id,h.customer_name,h.so_number,DATE_FORMAT(h.so_date,'%Y-%m-%d'),DATE_FORMAT(h.delivery_date,'%Y-%m-%d'),h.delivery_address,h.status,h.total_qty,CAST(h.total_amount AS CHAR),CAST(h.discount_amount AS CHAR),CAST(h.grand_total AS CHAR),h.notes,h.created_by,h.updated_by,h.created_at,h.updated_at FROM sales_order_header h JOIN branches b ON b.id=h.branch_id WHERE h.company_id=? AND h.id=?`, companyID, id)
	item, err := scanSalesOrder(row)
	if err != nil {
		return nil, err
	}
	rows, err := r.DB.Query(`SELECT d.id,d.so_header_id,d.item_id,i.sku,i.name,i.unit,d.qty,CAST(d.unit_price AS CHAR),CAST(d.discount_amount AS CHAR),CAST(d.subtotal AS CHAR),d.notes FROM sales_order_detail d JOIN items i ON i.id=d.item_id WHERE d.so_header_id=? ORDER BY d.id`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	item.Lines = make([]models.SalesOrderLine, 0)
	for rows.Next() {
		var line models.SalesOrderLine
		if err := rows.Scan(&line.ID, &line.SOHeaderID, &line.ItemID, &line.SKU, &line.ItemName, &line.Unit, &line.Qty, &line.UnitPrice, &line.DiscountAmount, &line.Subtotal, &line.Notes); err != nil {
			return nil, err
		}
		item.Lines = append(item.Lines, line)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *SalesOrderRepository) Create(companyID, userID uint64, number string, order models.SalesOrder) (uint64, error) {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	customerName, customerAddress, err := lockSalesOrderScope(tx, companyID, order.BranchID, order.CustomerID)
	if err != nil {
		return 0, err
	}
	address := customerAddress
	if order.DeliveryAddress != nil {
		address = order.DeliveryAddress
	}
	totalQty, totalAmount, grandTotal, prepared, err := prepareSalesOrderLines(tx, companyID, order.Lines, nil, order.DiscountAmount)
	if err != nil {
		return 0, err
	}
	discount, _ := utils.ParseDecimal(defaultMoney(order.DiscountAmount), 10, 2)
	result, err := tx.Exec(`INSERT INTO sales_order_header(company_id,branch_id,customer_id,so_number,so_date,delivery_date,customer_name,delivery_address,status,total_qty,total_amount,discount_amount,grand_total,notes,created_by) VALUES(?,?,?,?,?,?,?,?,'open',?,?,?,?,?,?)`, companyID, order.BranchID, order.CustomerID, number, order.SODate, nullableString(order.DeliveryDate), customerName, address, totalQty, totalAmount, discount.FloatString(2), grandTotal, order.Notes, userID)
	if err != nil {
		return 0, err
	}
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	id := uint64(lastID)
	if err := insertSalesOrderLines(tx, id, prepared); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func (r *SalesOrderRepository) Update(companyID, userID, id uint64, order models.SalesOrder) error {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	err = tx.QueryRow(`SELECT status FROM sales_order_header WHERE company_id=? AND id=? FOR UPDATE`, companyID, id).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSalesOrderNotFound
	}
	if err != nil {
		return err
	}
	if status != "open" {
		return ErrSalesOrderLocked
	}
	customerName, customerAddress, err := lockSalesOrderScope(tx, companyID, order.BranchID, order.CustomerID)
	if err != nil {
		return err
	}
	address := customerAddress
	if order.DeliveryAddress != nil {
		address = order.DeliveryAddress
	}
	previous, err := loadExistingSalesOrderLines(tx, id)
	if err != nil {
		return err
	}
	totalQty, totalAmount, grandTotal, prepared, err := prepareSalesOrderLines(tx, companyID, order.Lines, previous, order.DiscountAmount)
	if err != nil {
		return err
	}
	discount, _ := utils.ParseDecimal(defaultMoney(order.DiscountAmount), 10, 2)
	_, err = tx.Exec(`UPDATE sales_order_header SET branch_id=?,customer_id=?,so_date=?,delivery_date=?,customer_name=?,delivery_address=?,total_qty=?,total_amount=?,discount_amount=?,grand_total=?,notes=?,updated_by=? WHERE company_id=? AND id=?`, order.BranchID, order.CustomerID, order.SODate, nullableString(order.DeliveryDate), customerName, address, totalQty, totalAmount, discount.FloatString(2), grandTotal, order.Notes, userID, companyID, id)
	if err != nil {
		return err
	}
	kept := make(map[uint64]bool, len(prepared))
	for _, line := range prepared {
		if line.ExistingID == 0 {
			if err := insertSalesOrderLine(tx, id, line); err != nil {
				return err
			}
			continue
		}
		kept[line.ExistingID] = true
		if _, err := tx.Exec(`UPDATE sales_order_detail SET qty=?,discount_amount=?,subtotal=?,notes=? WHERE id=? AND so_header_id=?`, line.Qty, line.Discount, line.Subtotal, line.Notes, line.ExistingID, id); err != nil {
			return err
		}
	}
	for lineID := range previous {
		if !kept[lineID] {
			if _, err := tx.Exec(`DELETE FROM sales_order_detail WHERE id=? AND so_header_id=?`, lineID, id); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (r *SalesOrderRepository) Cancel(companyID, userID, id uint64) error {
	tx, err := r.DB.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var status string
	err = tx.QueryRow(`SELECT status FROM sales_order_header WHERE company_id=? AND id=? FOR UPDATE`, companyID, id).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrSalesOrderNotFound
	}
	if err != nil {
		return err
	}
	if status != "open" && status != "partially_delivered" {
		return ErrSalesOrderLocked
	}
	if _, err := tx.Exec(`UPDATE sales_order_header SET status='cancelled',updated_by=? WHERE company_id=? AND id=?`, userID, companyID, id); err != nil {
		return err
	}
	return tx.Commit()
}

func scanSalesOrder(row scanner) (models.SalesOrder, error) {
	var item models.SalesOrder
	var deliveryDate, deliveryAddress, notes sql.NullString
	var updatedUser sql.NullInt64
	var updatedAt sql.NullTime
	err := row.Scan(&item.ID, &item.CompanyID, &item.BranchID, &item.BranchName, &item.CustomerID, &item.CustomerName, &item.SONumber, &item.SODate, &deliveryDate, &deliveryAddress, &item.Status, &item.TotalQty, &item.TotalAmount, &item.DiscountAmount, &item.GrandTotal, &notes, &item.CreatedBy, &updatedUser, &item.CreatedAt, &updatedAt)
	if err != nil {
		return item, err
	}
	if deliveryDate.Valid {
		item.DeliveryDate = &deliveryDate.String
	}
	if deliveryAddress.Valid {
		item.DeliveryAddress = &deliveryAddress.String
	}
	if notes.Valid {
		item.Notes = &notes.String
	}
	if updatedUser.Valid {
		value := uint64(updatedUser.Int64)
		item.UpdatedBy = &value
	}
	if updatedAt.Valid {
		item.UpdatedAt = &updatedAt.Time
	}
	return item, nil
}

func lockSalesOrderScope(tx *sql.Tx, companyID, branchID, customerID uint64) (string, *string, error) {
	var branch uint64
	if err := tx.QueryRow(`SELECT id FROM branches WHERE id=? AND company_id=? AND is_active=1 FOR UPDATE`, branchID, companyID).Scan(&branch); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil, fmt.Errorf("%w: branch is unavailable", ErrSalesOrderInvalid)
		}
		return "", nil, err
	}
	var name string
	var address sql.NullString
	var active bool
	err := tx.QueryRow(`SELECT name,shipping_address,is_active FROM customers WHERE id=? AND company_id=? FOR UPDATE`, customerID, companyID).Scan(&name, &address, &active)
	if errors.Is(err, sql.ErrNoRows) || err == nil && !active {
		return "", nil, fmt.Errorf("%w: customer is unavailable", ErrSalesOrderInvalid)
	}
	if err != nil {
		return "", nil, err
	}
	if address.Valid {
		return name, &address.String, nil
	}
	return name, nil, nil
}

type storedLine struct {
	ItemID    uint64
	UnitPrice string
}

func loadExistingSalesOrderLines(tx *sql.Tx, id uint64) (map[uint64]storedLine, error) {
	rows, err := tx.Query(`SELECT id,item_id,CAST(unit_price AS CHAR) FROM sales_order_detail WHERE so_header_id=? FOR UPDATE`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[uint64]storedLine{}
	for rows.Next() {
		var lineID, itemID uint64
		var price string
		if err := rows.Scan(&lineID, &itemID, &price); err != nil {
			return nil, err
		}
		result[lineID] = storedLine{ItemID: itemID, UnitPrice: price}
	}
	return result, rows.Err()
}

func prepareSalesOrderLines(tx *sql.Tx, companyID uint64, lines []models.SalesOrderLine, existing map[uint64]storedLine, headerDiscount string) (int64, string, string, []preparedOrderLine, error) {
	if len(lines) == 0 {
		return 0, "", "", nil, fmt.Errorf("%w: add at least one item", ErrSalesOrderInvalid)
	}
	if headerDiscount == "" {
		headerDiscount = "0"
	}
	discount, ok := utils.ParseDecimal(headerDiscount, 10, 2)
	if !ok || discount.Sign() < 0 {
		return 0, "", "", nil, fmt.Errorf("%w: document discount must be a non-negative amount with up to two decimals", ErrSalesOrderInvalid)
	}
	seen := map[uint64]bool{}
	prepared := make([]preparedOrderLine, 0, len(lines))
	var qtyTotal int64
	sum := new(big.Rat)
	for _, line := range lines {
		if line.ItemID == 0 || line.Qty < 1 || seen[line.ItemID] {
			return 0, "", "", nil, fmt.Errorf("%w: each item must be unique and have a positive whole-number quantity", ErrSalesOrderInvalid)
		}
		seen[line.ItemID] = true
		price := ""
		if line.ID > 0 {
			previous, ok := existing[line.ID]
			if !ok || previous.ItemID != line.ItemID {
				return 0, "", "", nil, fmt.Errorf("%w: sales order line is invalid", ErrSalesOrderInvalid)
			}
			price = previous.UnitPrice
		} else {
			err := tx.QueryRow(`SELECT CAST(selling_price AS CHAR) FROM items WHERE id=? AND company_id=? AND is_active=1 FOR UPDATE`, line.ItemID, companyID).Scan(&price)
			if errors.Is(err, sql.ErrNoRows) {
				return 0, "", "", nil, fmt.Errorf("%w: item is unavailable", ErrSalesOrderInvalid)
			}
			if err != nil {
				return 0, "", "", nil, err
			}
		}
		priceRat, valid := utils.ParseDecimal(price, 10, 2)
		if !valid || priceRat.Sign() < 0 {
			return 0, "", "", nil, fmt.Errorf("%w: item has an invalid selling price", ErrSalesOrderInvalid)
		}
		lineDiscountValue := line.DiscountAmount
		if lineDiscountValue == "" {
			lineDiscountValue = "0"
		}
		lineDiscount, valid := utils.ParseDecimal(lineDiscountValue, 10, 2)
		if !valid || lineDiscount.Sign() < 0 {
			return 0, "", "", nil, fmt.Errorf("%w: line discount must be a non-negative amount with up to two decimals", ErrSalesOrderInvalid)
		}
		gross := new(big.Rat).Mul(priceRat, big.NewRat(line.Qty, 1))
		if lineDiscount.Cmp(gross) > 0 {
			return 0, "", "", nil, fmt.Errorf("%w: line discount cannot exceed the line amount", ErrSalesOrderInvalid)
		}
		subtotal := new(big.Rat).Sub(gross, lineDiscount)
		sum.Add(sum, subtotal)
		if qtyTotal > math.MaxInt64-line.Qty {
			return 0, "", "", nil, fmt.Errorf("%w: total quantity is too large", ErrSalesOrderInvalid)
		}
		qtyTotal += line.Qty
		prepared = append(prepared, preparedOrderLine{ExistingID: line.ID, ItemID: line.ItemID, Qty: line.Qty, UnitPrice: price, Discount: lineDiscount.FloatString(2), Subtotal: subtotal.FloatString(2), Notes: line.Notes})
	}
	if discount.Cmp(sum) > 0 {
		return 0, "", "", nil, fmt.Errorf("%w: order discount cannot exceed the amount after line discounts", ErrSalesOrderInvalid)
	}
	if len(sum.FloatString(2)) > 13 || len(new(big.Rat).Sub(sum, discount).FloatString(2)) > 13 {
		return 0, "", "", nil, fmt.Errorf("%w: order amount exceeds the supported limit", ErrSalesOrderInvalid)
	}
	return qtyTotal, sum.FloatString(2), new(big.Rat).Sub(sum, discount).FloatString(2), prepared, nil
}

type preparedOrderLine struct {
	ExistingID                    uint64
	ItemID                        uint64
	Qty                           int64
	UnitPrice, Discount, Subtotal string
	Notes                         *string
}

func insertSalesOrderLines(tx *sql.Tx, id uint64, lines []preparedOrderLine) error {
	for _, line := range lines {
		if err := insertSalesOrderLine(tx, id, line); err != nil {
			return err
		}
	}
	return nil
}

func insertSalesOrderLine(tx *sql.Tx, id uint64, line preparedOrderLine) error {
	_, err := tx.Exec(`INSERT INTO sales_order_detail(so_header_id,item_id,qty,unit_price,discount_amount,subtotal,notes) VALUES(?,?,?,?,?,?,?)`, id, line.ItemID, line.Qty, line.UnitPrice, line.Discount, line.Subtotal, line.Notes)
	return err
}
func nullableString(value *string) any {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil
	}
	return strings.TrimSpace(*value)
}
func defaultMoney(value string) string {
	if strings.TrimSpace(value) == "" {
		return "0"
	}
	return value
}
