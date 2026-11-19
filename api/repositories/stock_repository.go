package repositories

import (
	"database/sql"
	"errors"
	"strings"

	"erp-api/models"
)

type StockRepository struct{ DB *sql.DB }

func NewStockRepository(db *sql.DB) *StockRepository { return &StockRepository{DB: db} }

func (r *StockRepository) Branches(companyID uint64) ([]models.Branch, error) {
	rows, err := r.DB.Query(`SELECT id,name,code FROM branches WHERE company_id=? ORDER BY name`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	branches := make([]models.Branch, 0)
	for rows.Next() {
		var branch models.Branch
		if err := rows.Scan(&branch.ID, &branch.Name, &branch.Code); err != nil {
			return nil, err
		}
		branches = append(branches, branch)
	}
	return branches, rows.Err()
}

func (r *StockRepository) List(companyID, branchID uint64, search string, limit, offset int) ([]models.StockSummary, int, error) {
	where := []string{"b.company_id=?", "b.id=?", "i.company_id=b.company_id", "i.is_active=1"}
	args := []any{companyID, branchID}
	if search != "" {
		where = append(where, "(i.name LIKE ? OR i.sku LIKE ?)")
		like := "%" + strings.TrimSpace(search) + "%"
		args = append(args, like, like)
	}
	clause := strings.Join(where, " AND ")
	base := ` FROM branches b JOIN items i ON i.company_id=b.company_id LEFT JOIN stock_period sp ON sp.company_id=b.company_id AND sp.branch_id=b.id AND sp.item_id=i.id WHERE ` + clause
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.Query(`SELECT b.company_id,b.id,b.name,i.id,i.sku,i.name,i.unit,COALESCE(sp.stock_date,CURRENT_DATE()),COALESCE(sp.qty,0),i.min_stock`+base+` ORDER BY i.name,i.id LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	list := make([]models.StockSummary, 0)
	for rows.Next() {
		var stock models.StockSummary
		if err := rows.Scan(&stock.CompanyID, &stock.BranchID, &stock.BranchName, &stock.ItemID, &stock.SKU, &stock.ItemName, &stock.Unit, &stock.StockDate, &stock.Qty, &stock.MinStock); err != nil {
			return nil, 0, err
		}
		list = append(list, stock)
	}
	return list, total, rows.Err()
}

func (r *StockRepository) Cards(companyID, branchID, itemID uint64, limit, offset int) ([]models.StockCard, int, error) {
	base := ` FROM stock_card sc JOIN branches b ON b.id=sc.branch_id WHERE sc.company_id=? AND sc.branch_id=? AND sc.item_id=? AND b.company_id=?`
	args := []any{companyID, branchID, itemID, companyID}
	var total int
	if err := r.DB.QueryRow(`SELECT COUNT(*)`+base, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.DB.Query(`SELECT sc.id,sc.company_id,sc.stock_date,sc.ref_num,sc.branch_id,sc.item_id,sc.trans_type,sc.in_qty,sc.out_qty,sc.hpp`+base+` ORDER BY sc.stock_date DESC,sc.id DESC LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	cards := make([]models.StockCard, 0)
	for rows.Next() {
		var card models.StockCard
		if err := rows.Scan(&card.ID, &card.CompanyID, &card.StockDate, &card.RefNum, &card.BranchID, &card.ItemID, &card.TransType, &card.InQty, &card.OutQty, &card.HPP); err != nil {
			return nil, 0, err
		}
		cards = append(cards, card)
	}
	return cards, total, rows.Err()
}

func applyStockMovementTx(tx *sql.Tx, companyID, branchID, itemID uint64, refNum, transType, movementDate string, inQty, outQty int64, hpp string) error {
	if _, err := tx.Exec(`INSERT INTO stock_period (company_id,branch_id,item_id,stock_date,qty) VALUES (?,?,?,?,0)
		ON DUPLICATE KEY UPDATE id=id`, companyID, branchID, itemID, movementDate); err != nil {
		return err
	}
	var qty int64
	if err := tx.QueryRow(`SELECT qty FROM stock_period WHERE company_id=? AND branch_id=? AND item_id=? FOR UPDATE`, companyID, branchID, itemID).Scan(&qty); err != nil {
		return err
	}
	if qty+inQty-outQty < 0 {
		return ErrStockInsufficient
	}
	if _, err := tx.Exec(`UPDATE stock_period SET qty=qty+?-?,stock_date=? WHERE company_id=? AND branch_id=? AND item_id=?`, inQty, outQty, movementDate, companyID, branchID, itemID); err != nil {
		return err
	}
	_, err := tx.Exec(`INSERT INTO stock_card (company_id,stock_date,ref_num,branch_id,item_id,trans_type,in_qty,out_qty,hpp) VALUES (?,CONCAT(?,' 00:00:00'),?,?,?,?,?,?,?)`, companyID, movementDate, refNum, branchID, itemID, transType, inQty, outQty, hpp)
	return err
}

var ErrStockInsufficient = errors.New("insufficient branch stock")
