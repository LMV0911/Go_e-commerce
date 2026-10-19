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
	query := `SELECT id, name, description, category, price_ht, price_ttc, stock, created_at FROM products ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
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
	return products, nil
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
	query := `UPDATE products SET category = $1 WHERE category ILIKE $2`
	res, err := r.db.Exec(query, newCategory, oldCategory)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

