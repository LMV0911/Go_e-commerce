package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/LMV0911/Go_e-commerce/internal/models"
	"github.com/LMV0911/Go_e-commerce/internal/utils"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) ListAll() ([]models.Order, error) {
	query := `
		SELECT o.id, o.user_id, u.email, o.status, COALESCE(o.cancel_reason, ''), o.total_ttc, o.created_at
		FROM orders o
		JOIN users u ON o.user_id = u.id
		ORDER BY o.created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur de récupération des commandes: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.UserEmail, &o.Status, &o.CancelReason, &o.TotalTTC, &o.CreatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération des commandes: %w", err)
	}
	return orders, nil
}

func (r *OrderRepository) GetByID(id string) (*models.Order, error) {
	query := `
		SELECT o.id, o.user_id, u.email, o.status, COALESCE(o.cancel_reason, ''), o.total_ttc, o.created_at
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE o.id = $1
	`
	var o models.Order
	err := r.db.QueryRow(query, id).Scan(&o.ID, &o.UserID, &o.UserEmail, &o.Status, &o.CancelReason, &o.TotalTTC, &o.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	itemsQuery := `
		SELECT oi.id, oi.order_id, oi.product_id, p.name, oi.quantity, oi.price_unit_ttc
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
	`
	itemRows, err := r.db.Query(itemsQuery, id)
	if err == nil {
		defer itemRows.Close()
		for itemRows.Next() {
			var item models.OrderItem
			if err := itemRows.Scan(&item.ID, &item.OrderID, &item.ProductID, &item.ProductName, &item.Quantity, &item.PriceUnitTTC); err == nil {
				o.Items = append(o.Items, item)
			}
		}
		_ = itemRows.Err()
	}

	return &o, nil
}

func (r *OrderRepository) UpdateStatus(id string, status string, cancelReason string) error {
	cleanStatus := strings.TrimSpace(status)
	if strings.HasPrefix(strings.ToLower(cleanStatus), "annul") {
		cleanStatus = models.StatusCancelled
	}

	if cleanStatus == models.StatusCancelled && strings.TrimSpace(cancelReason) == "" {
		return fmt.Errorf("un motif d'annulation est obligatoire lorsque la commande est annulée")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var currentStatus string
	err = tx.QueryRow(`SELECT status FROM orders WHERE id = $1`, id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("commande %s non trouvée", id)
		}
		return err
	}

	if currentStatus == models.StatusCancelled && cleanStatus != models.StatusCancelled {
		return fmt.Errorf("une commande déjà annulée ne peut plus changer de statut")
	}

	if cleanStatus == models.StatusCancelled && currentStatus != models.StatusCancelled {
		rows, err := tx.Query(`SELECT product_id, quantity FROM order_items WHERE order_id = $1`, id)
		if err != nil {
			return err
		}
		defer rows.Close()

		type itemRestock struct {
			productID string
			quantity  int
		}
		var items []itemRestock
		for rows.Next() {
			var it itemRestock
			if err := rows.Scan(&it.productID, &it.quantity); err != nil {
				return err
			}
			items = append(items, it)
		}
		if err := rows.Err(); err != nil {
			return err
		}

		for _, it := range items {
			_, err := tx.Exec(`UPDATE products SET stock = stock + $1 WHERE id = $2`, it.quantity, it.productID)
			if err != nil {
				return fmt.Errorf("erreur de réintégration du stock pour le produit %s: %w", it.productID, err)
			}
		}
	}

	query := `UPDATE orders SET status = $1, cancel_reason = $2 WHERE id = $3`
	res, err := tx.Exec(query, cleanStatus, cancelReason, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("commande %s non trouvée", id)
	}

	return tx.Commit()
}

func (r *OrderRepository) CreateManualOrder(req models.CreateManualOrderRequest) (*models.Order, error) {
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("la commande doit contenir au moins un produit")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	orderID := utils.GenerateID("CMD")
	var totalTTC float64 = 0

	type tempItem struct {
		productID string
		quantity  int
		unitPrice float64
	}
	var preparedItems []tempItem

	for _, itemReq := range req.Items {
		var pPriceTTC float64
		var stock int
		err := tx.QueryRow(`SELECT price_ttc, stock FROM products WHERE id = $1`, itemReq.ProductID).Scan(&pPriceTTC, &stock)
		if err != nil {
			return nil, fmt.Errorf("produit %s invalide ou inexistant", itemReq.ProductID)
		}
		if stock < itemReq.Quantity {
			return nil, fmt.Errorf("stock insuffisant pour le produit %s (disponible: %d)", itemReq.ProductID, stock)
		}

		itemTotal := pPriceTTC * float64(itemReq.Quantity)
		totalTTC += itemTotal
		preparedItems = append(preparedItems, tempItem{
			productID: itemReq.ProductID,
			quantity:  itemReq.Quantity,
			unitPrice: pPriceTTC,
		})
	}

	orderQuery := `INSERT INTO orders (id, user_id, status, total_ttc) VALUES ($1, $2, $3, $4) RETURNING created_at`
	var order models.Order
	order.ID = orderID
	order.UserID = req.UserID
	order.Status = models.StatusPending
	order.TotalTTC = totalTTC

	err = tx.QueryRow(orderQuery, orderID, req.UserID, models.StatusPending, totalTTC).Scan(&order.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("erreur d'insertion commande: %w", err)
	}

	for _, pi := range preparedItems {
		_, err := tx.Exec(`INSERT INTO order_items (order_id, product_id, quantity, price_unit_ttc) VALUES ($1, $2, $3, $4)`,
			orderID, pi.productID, pi.quantity, pi.unitPrice)
		if err != nil {
			return nil, fmt.Errorf("erreur d'insertion article commande: %w", err)
		}

		_, err = tx.Exec(`UPDATE products SET stock = stock - $1 WHERE id = $2`, pi.quantity, pi.productID)
		if err != nil {
			return nil, fmt.Errorf("erreur de mise à jour du stock: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &order, nil
}
