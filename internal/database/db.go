package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	if host == "" || host == "db" {
		// En local (hors réseau interne docker), on se connecte sur localhost
		if os.Getenv("DB_HOST") == "" {
			host = "localhost"
		}
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "admin"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "ecommerce"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("erreur d'ouverture de connexion DB: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("impossible de joindre la base de données PostgreSQL: %w", err)
	}

	_, _ = db.Exec(`ALTER TABLE users ADD COLUMN IF NOT EXISTS reset_code VARCHAR(10);`)
	_, _ = db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`)

	log.Println("Connexion a la base de donnees PostgreSQL etablie")
	return db, nil
}
