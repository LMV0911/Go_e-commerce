package repository_test

import (
	"database/sql"
	"testing"

	"github.com/LMV0911/Go_e-commerce/internal/repository"
	_ "github.com/lib/pq"
)

func TestCartRepo_AddItem(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://admin:password@localhost:5432/ecommerce?sslmode=disable")
	if err != nil {
		t.Skip("Postgres non accessible")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skip("Postgres non accessible")
	}

	cartRepo := repository.NewCartRepository(db)
	cart, err := cartRepo.AddItem(1, "PDT-7D2K8N", 1)
	if err != nil {
		t.Fatalf("AddItem a échoué: %v", err)
	}

	if cart == nil {
		t.Fatal("Cart est nil")
	}

	if len(cart.Items) == 0 {
		t.Fatal("Cart ne contient aucun article")
	}

	t.Logf("Panier créé avec succès: ID=%s, Articles=%d, Total=%.2f €", cart.ID, len(cart.Items), cart.TotalTTC)
}
