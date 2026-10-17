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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erreur d'itération sur les utilisateurs: %w", err)
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

func (r *UserRepository) Create(u *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, role, is_confirmed, confirmation_code)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	err := r.db.QueryRow(query, u.Email, u.PasswordHash, u.Role, u.IsConfirmed, u.ConfirmationCode).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return fmt.Errorf("erreur lors de la création de l'utilisateur: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, role, is_confirmed, 
		       COALESCE(confirmation_code, ''), COALESCE(reset_code, ''), created_at 
		FROM users 
		WHERE LOWER(email) = LOWER($1)
	`
	var u models.User
	err := r.db.QueryRow(query, email).Scan(
		&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.IsConfirmed,
		&u.ConfirmationCode, &u.ResetCode, &u.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur lors de la recherche de l'utilisateur: %w", err)
	}
	return &u, nil
}

func (r *UserRepository) ConfirmByCode(email, code string) error {
	u, err := r.GetByEmail(email)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("aucun utilisateur trouvé avec l'email %s", email)
	}
	if u.IsConfirmed {
		return fmt.Errorf("ce compte est déjà confirmé")
	}
	if u.ConfirmationCode == "" || u.ConfirmationCode != code {
		return fmt.Errorf("code de confirmation incorrect")
	}

	query := `UPDATE users SET is_confirmed = TRUE, confirmation_code = NULL WHERE id = $1`
	_, err = r.db.Exec(query, u.ID)
	if err != nil {
		return fmt.Errorf("erreur lors de la confirmation du compte: %w", err)
	}
	return nil
}

func (r *UserRepository) SetResetCode(email, resetCode string) error {
	u, err := r.GetByEmail(email)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("aucun utilisateur trouvé avec l'email %s", email)
	}

	query := `UPDATE users SET reset_code = $1 WHERE id = $2`
	_, err = r.db.Exec(query, resetCode, u.ID)
	if err != nil {
		return fmt.Errorf("erreur lors de l'enregistrement du code de réinitialisation: %w", err)
	}
	return nil
}

func (r *UserRepository) ResetPassword(email, code, newPasswordHash string) error {
	u, err := r.GetByEmail(email)
	if err != nil {
		return err
	}
	if u == nil {
		return fmt.Errorf("aucun utilisateur trouvé avec l'email %s", email)
	}

	if code != "" {
		if u.ResetCode == "" || u.ResetCode != code {
			return fmt.Errorf("code de réinitialisation incorrect ou expiré")
		}
	}

	query := `UPDATE users SET password_hash = $1, reset_code = NULL WHERE id = $2`
	_, err = r.db.Exec(query, newPasswordHash, u.ID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du mot de passe: %w", err)
	}
	return nil
}

func (r *UserRepository) UpdatePassword(id int, newPasswordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	res, err := r.db.Exec(query, newPasswordHash, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur ID %d non trouvé", id)
	}
	return nil
}
