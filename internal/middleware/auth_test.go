package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LMV0911/Go_e-commerce/internal/auth"
	"github.com/LMV0911/Go_e-commerce/internal/middleware"
	"github.com/LMV0911/Go_e-commerce/internal/models"
)

func TestRequireAuthMiddleware(t *testing.T) {
	protectedHandler := middleware.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("autorisé"))
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	rr := httptest.NewRecorder()
	protectedHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Code attendu 401 Unauthorized, obtenu: %d", rr.Code)
	}

	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer faux-token")
	rr = httptest.NewRecorder()
	protectedHandler(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("Code attendu 401 Unauthorized, obtenu: %d", rr.Code)
	}

	user := &models.User{ID: 10, Email: "client@test.com", Role: models.RoleClient}
	token, _ := auth.GenerateToken(user)

	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr = httptest.NewRecorder()
	protectedHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Code attendu 200 OK, obtenu: %d", rr.Code)
	}
}

func TestRequireRoleMiddleware(t *testing.T) {
	adminOnlyHandler := middleware.RequireRole(models.RoleAdmin, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("admin ok"))
	})

	clientUser := &models.User{ID: 10, Email: "client@test.com", Role: models.RoleClient}
	clientToken, _ := auth.GenerateToken(clientUser)

	req := httptest.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+clientToken)
	rr := httptest.NewRecorder()
	adminOnlyHandler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("Code attendu 403 Forbidden pour le client, obtenu: %d", rr.Code)
	}

	adminUser := &models.User{ID: 1, Email: "admin@test.com", Role: models.RoleAdmin}
	adminToken, _ := auth.GenerateToken(adminUser)

	req = httptest.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rr = httptest.NewRecorder()
	adminOnlyHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Code attendu 200 OK pour l'admin, obtenu: %d", rr.Code)
	}
}
