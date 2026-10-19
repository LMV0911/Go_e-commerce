package repository

import (
	"database/sql"
	"fmt"
	"math"

	"github.com/LMV0911/Go_e-commerce/internal/models"
	"github.com/LMV0911/Go_e-commerce/internal/utils"
)

type CartRepository struct {
	db *sql.DB
}

func NewCartRepository(db *sql.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) GetOrCreate(userID int) (*models.Cart, error) {
	cart := &models.Cart{
		UserID: userID,
		Items:  make([]models.CartItem, 0),
	}

	query := `SELECT id, user_id, created_at, updated_at FROM carts WHERE user_id = $1 ORDER BY created_at DESC LIMIT 1`
	err := r.db.QueryRow(query, userID).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt)

	if err == sql.ErrNoRows {
		cart.ID = utils.GenerateID("BSK")
		insertQuery := `INSERT INTO carts (id, user_id) VALUES ($1, $2) RETURNING created_at, updated_at`
		err = r.db.QueryRow(insertQuery, cart.ID, userID).Scan(&cart.CreatedAt, &cart.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("erreur lors de la création du panier: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("erreur de récupération du panier: %w", err)
	}

	itemsQuery := `
		SELECT ci.id, ci.product_id, ci.quantity, p.name, p.description, p.category, p.price_ht, p.price_ttc, p.stock
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = $1
		ORDER BY ci.id ASC
	`
	rows, err := r.db.Query(itemsQuery, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des articles du panier: %w", err)
	}
	defer rows.Close()

	var totalTTC float64
	for rows.Next() {
		var it models.CartItem
		var p models.Product
		if err := rows.Scan(&it.ID, &it.ProductID, &it.Quantity, &p.Name, &p.Description, &p.Category, &p.PriceHT, &p.PriceTTC, &p.Stock); err != nil {
			return nil, err
		}
		p.ID = it.ProductID
		it.Product = p
		totalTTC += p.PriceTTC * float64(it.Quantity)
		cart.Items = append(cart.Items, it)
	}

	cart.TotalTTC = math.Round(totalTTC*100) / 100
	return cart, nil
}

func (r *CartRepository) AddItem(userID int, productID string, quantity int) (*models.Cart, error) {
	if quantity <= 0 {
		quantity = 1
	}

	var stock int
	err := r.db.QueryRow(`SELECT stock FROM products WHERE id = $1`, productID).Scan(&stock)
	if err != nil {
		return nil, fmt.Errorf("produit %s introuvable", productID)
	}
	if stock < quantity {
		return nil, fmt.Errorf("stock insuffisant (disponible: %d)", stock)
	}

	cart, err := r.GetOrCreate(userID)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO cart_items (cart_id, product_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (cart_id, product_id)
		DO UPDATE SET quantity = cart_items.quantity + $3
	`
	_, err = r.db.Exec(query, cart.ID, productID, quantity)
	if err != nil {
		return nil, fmt.Errorf("erreur ajout article au panier: %w", err)
	}

	_, _ = r.db.Exec(`UPDATE carts SET updated_at = NOW() WHERE id = $1`, cart.ID)
	return r.GetOrCreate(userID)
}

func (r *CartRepository) UpdateItem(userID int, productID string, quantity int) (*models.Cart, error) {
	cart, err := r.GetOrCreate(userID)
	if err != nil {
		return nil, err
	}

	if quantity <= 0 {
		return r.RemoveItem(userID, productID)
	}

	var stock int
	err = r.db.QueryRow(`SELECT stock FROM products WHERE id = $1`, productID).Scan(&stock)
	if err != nil {
		return nil, fmt.Errorf("produit %s introuvable", productID)
	}
	if stock < quantity {
		return nil, fmt.Errorf("stock insuffisant (disponible: %d)", stock)
	}

	_, err = r.db.Exec(`UPDATE cart_items SET quantity = $1 WHERE cart_id = $2 AND product_id = $3`, quantity, cart.ID, productID)
	if err != nil {
		return nil, fmt.Errorf("erreur modification quantité: %w", err)
	}

	_, _ = r.db.Exec(`UPDATE carts SET updated_at = NOW() WHERE id = $1`, cart.ID)
	return r.GetOrCreate(userID)
}

func (r *CartRepository) RemoveItem(userID int, productID string) (*models.Cart, error) {
	cart, err := r.GetOrCreate(userID)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec(`DELETE FROM cart_items WHERE cart_id = $1 AND product_id = $2`, cart.ID, productID)
	if err != nil {
		return nil, fmt.Errorf("erreur suppression article: %w", err)
	}

	_, _ = r.db.Exec(`UPDATE carts SET updated_at = NOW() WHERE id = $1`, cart.ID)
	return r.GetOrCreate(userID)
}

func (r *CartRepository) Clear(userID int) (*models.Cart, error) {
	cart, err := r.GetOrCreate(userID)
	if err != nil {
		return nil, err
	}

	_, err = r.db.Exec(`DELETE FROM cart_items WHERE cart_id = $1`, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("erreur vidage panier: %w", err)
	}

	_, _ = r.db.Exec(`UPDATE carts SET updated_at = NOW() WHERE id = $1`, cart.ID)
	return r.GetOrCreate(userID)
}
