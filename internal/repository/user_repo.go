package repository

import (
	"database/sql"
	"fmt"

	"github.com/LMV0911/Go_e-commerce/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) List() ([]models.User, error) {
	query := `SELECT id, email, role, is_confirmed, created_at FROM users ORDER BY created_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("erreur de récupération des utilisateurs: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Email, &u.Role, &u.IsConfirmed, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	query := `SELECT id, email, role, is_confirmed, created_at FROM users WHERE id = $1`
	var u models.User
	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Email, &u.Role, &u.IsConfirmed, &u.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Update(u *models.User) error {
	query := `UPDATE users SET email = $1, role = $2, is_confirmed = $3 WHERE id = $4`
	res, err := r.db.Exec(query, u.Email, u.Role, u.IsConfirmed, u.ID)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur ID %d non trouvé", u.ID)
	}
	return nil
}

func (r *UserRepository) Delete(id int) error {
	query := `DELETE FROM users WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur ID %d non trouvé", id)
	}
	return nil
}

func (r *UserRepository) ConfirmAccount(id int) error {
	query := `UPDATE users SET is_confirmed = TRUE WHERE id = $1`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur ID %d non trouvé", id)
	}
	return nil
}
