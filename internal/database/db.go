package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func InitDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "db"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		return nil, fmt.Errorf("variable d'environnement DB_USER manquant")
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("variable d'environnement DB_PASSWORD manquante")
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		return nil, fmt.Errorf("variable d'environnement DB_NAME manquante")
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("erreur d'ouverture de connexion DB: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("impossible de joindre la base de données PostgreSQL: %w", err)
	}

	log.Println("✅ Connexion sécurisée à la base de données PostgreSQL établie !")
	return db, nil
}
