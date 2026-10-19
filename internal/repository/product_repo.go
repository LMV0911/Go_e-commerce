package repository

import (
	"database/sql"
	"fmt"

	"github.com/LMV0911/Go_e-commerce/internal/models"
	"github.com/LMV0911/Go_e-commerce/internal/utils"
)

type ProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) *ProductRepository {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Create(p *models.Product) error {
	if p.ID == "" {
		p.ID = utils.GenerateID("PDT")
	}

	query := `
		INSERT INTO products (id, name, description, category, price_ht, price_ttc, stock)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`
	err := r.db.QueryRow(query, p.ID, p.Name, p.Description, p.Category, p.PriceHT, p.PriceTTC, p.Stock).Scan(&p.CreatedAt)
	if err != nil {
		return fmt.Errorf("erreur lors de la création du produit: %w", err)
	}
	return nil
}

func (r *ProductRepository) List() ([]models.Product, error) {
	return r.Search("", "", "", "", "", nil, nil, "")
}

func (r *ProductRepository) Search(q, name, category, description, priceType string, minPrice, maxPrice *float64, sortBy string) ([]models.Product, error) {
	query := `SELECT id, name, description, category, price_ht, price_ttc, stock, created_at FROM products WHERE 1=1`
	var args []any
	argID := 1

	if q != "" {
		query += fmt.Sprintf(" AND (name ILIKE $%d OR description ILIKE $%d OR category ILIKE $%d)", argID, argID, argID)
		args = append(args, "%"+q+"%")
		argID++
	}
	if name != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argID)
		args = append(args, "%"+name+"%")
		argID++
	}
	if category != "" && category != "Toutes" && category != "Tous" {
		query += fmt.Sprintf(" AND category ILIKE $%d", argID)
		args = append(args, category)
		argID++
	}
	if description != "" {
		query += fmt.Sprintf(" AND description ILIKE $%d", argID)
		args = append(args, "%"+description+"%")
		argID++
	}

	priceCol := "price_ttc"
	if priceType == "ht" {
		priceCol = "price_ht"
	}

	if minPrice != nil {
		query += fmt.Sprintf(" AND %s >= $%d", priceCol, argID)
		args = append(args, *minPrice)
		argID++
	}
	if maxPrice != nil {
		query += fmt.Sprintf(" AND %s <= $%d", priceCol, argID)
		args = append(args, *maxPrice)
		argID++
	}

	switch sortBy {
	case "name_asc":
		query += " ORDER BY name ASC"
	case "name_desc":
		query += " ORDER BY name DESC"
	case "price_asc":
		query += fmt.Sprintf(" ORDER BY %s ASC", priceCol)
	case "price_desc":
		query += fmt.Sprintf(" ORDER BY %s DESC", priceCol)
	default:
		query += " ORDER BY created_at DESC"
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("erreur de récupération des produits: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.PriceHT, &p.PriceTTC, &p.Stock, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur lors de l'itération des produits: %w", err)
	}
	return products, nil
}

func (r *ProductRepository) ListCategories() ([]string, error) {
	query := `SELECT DISTINCT category FROM products ORDER BY category ASC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération catégories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var cat string
		if err := rows.Scan(&cat); err == nil {
			categories = append(categories, cat)
		}
	}
	return categories, nil
}

func (r *ProductRepository) GetByID(id string) (*models.Product, error) {
	query := `SELECT id, name, description, category, price_ht, price_ttc, stock, created_at FROM products WHERE id = $1`
	var p models.Product
	err := r.db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Description, &p.Category, &p.PriceHT, &p.PriceTTC, &p.Stock, &p.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) Update(p *models.Product) error {
	query := `
		UPDATE products
		SET name = $1, description = $2, category = $3, price_ht = $4, price_ttc = $5, stock = $6
		WHERE id = $7
	`
	res, err := r.db.Exec(query, p.Name, p.Description, p.Category, p.PriceHT, p.PriceTTC, p.Stock, p.ID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("produit %s non trouvé", p.ID)
	}
	return nil
}

func (r *ProductRepository) Delete(id string) error {
	query := `DELETE FROM products WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("produit %s non trouvé", id)
	}
	return nil
}

func (r *ProductRepository) RenameCategory(oldCategory, newCategory string) (int64, error) {
	query := `UPDATE products SET category = $1 WHERE category ILIKE $2 OR unaccent(category) ILIKE unaccent($2)`
	res, err := r.db.Exec(query, newCategory, oldCategory)
	if err != nil {
		res, err = r.db.Exec(`UPDATE products SET category = $1 WHERE category ILIKE $2`, newCategory, oldCategory)
		if err != nil {
			return 0, err
		}
	}
	return res.RowsAffected()
}

