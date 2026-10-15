package auth_test

import (
	"testing"

	"github.com/LMV0911/Go_e-commerce/internal/auth"
	"github.com/LMV0911/Go_e-commerce/internal/models"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "SecretPass123"

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword a échoué: %v", err)
	}

	if hash == password {
		t.Fatal("Le hash ne doit pas être égal au mot de passe en clair")
	}

	if !auth.CheckPassword(hash, password) {
		t.Fatal("CheckPassword devrait retourner true pour le mot de passe valide")
	}

	if auth.CheckPassword(hash, "MauvaisMotDePasse") {
		t.Fatal("CheckPassword devrait retourner false pour un mot de passe erroné")
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	user := &models.User{
		ID:    42,
		Email: "client@test.com",
		Role:  models.RoleClient,
	}

	token, err := auth.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken a échoué: %v", err)
	}
	if token == "" {
		t.Fatal("Le token généré ne doit pas être vide")
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken a échoué: %v", err)
	}

	if claims.UserID != 42 {
		t.Errorf("claims.UserID attendu: 42, obtenu: %d", claims.UserID)
	}
	if claims.Email != "client@test.com" {
		t.Errorf("claims.Email attendu: client@test.com, obtenu: %s", claims.Email)
	}
	if claims.Role != models.RoleClient {
		t.Errorf("claims.Role attendu: %s, obtenu: %s", models.RoleClient, claims.Role)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	_, err := auth.ValidateToken("")
	if err == nil {
		t.Fatal("ValidateToken devrait échouer avec une chaîne vide")
	}

	_, err = auth.ValidateToken("invalid.token.string")
	if err == nil {
		t.Fatal("ValidateToken devrait échouer avec un faux token")
	}
}
