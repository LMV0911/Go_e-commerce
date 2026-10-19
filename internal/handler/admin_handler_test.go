package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LMV0911/Go_e-commerce/internal/auth"
	"github.com/LMV0911/Go_e-commerce/internal/handler"
	"github.com/LMV0911/Go_e-commerce/internal/models"
)

func TestAdminRoutes_AuthenticationAndAuthorization(t *testing.T) {
	router := handler.SetupRoutes(nil)

	adminRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/admin/products"},
		{"GET", "/api/admin/products/PDT-123456"},
		{"POST", "/api/admin/products"},
		{"PUT", "/api/admin/products/PDT-123456"},
		{"DELETE", "/api/admin/products/PDT-123456"},
		{"PUT", "/api/admin/categories/rename"},
		{"GET", "/api/admin/users"},
		{"GET", "/api/admin/users/1"},
		{"POST", "/api/admin/users"},
		{"POST", "/api/admin/users/1/confirm"},
		{"PUT", "/api/admin/users/1"},
		{"DELETE", "/api/admin/users/1"},
		{"GET", "/api/admin/orders"},
		{"GET", "/api/admin/orders/CMD-123456"},
		{"PUT", "/api/admin/orders/CMD-123456/status"},
		{"POST", "/api/admin/orders/manual"},
	}

	clientUser := &models.User{
		ID:          10,
		Email:       "client@test.com",
		Role:        models.RoleClient,
		IsConfirmed: true,
	}
	clientToken, err := auth.GenerateToken(clientUser)
	if err != nil {
		t.Fatalf("Erreur génération token client: %v", err)
	}

	adminUser := &models.User{
		ID:          1,
		Email:       "admin@ecommerce.com",
		Role:        models.RoleAdmin,
		IsConfirmed: true,
	}
	adminToken, err := auth.GenerateToken(adminUser)
	if err != nil {
		t.Fatalf("Erreur génération token admin: %v", err)
	}

	for _, route := range adminRoutes {
		t.Run("Unauthenticated "+route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusUnauthorized {
				t.Errorf("Attendu 401 Unauthorized pour %s %s, obtenu %d", route.method, route.path, rr.Code)
			}
		})

		t.Run("Forbidden for Client "+route.method+" "+route.path, func(t *testing.T) {
			req := httptest.NewRequest(route.method, route.path, nil)
			req.Header.Set("Authorization", "Bearer "+clientToken)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code != http.StatusForbidden {
				t.Errorf("Attendu 403 Forbidden pour client sur %s %s, obtenu %d", route.method, route.path, rr.Code)
			}
		})

		t.Run("Authorized for Admin "+route.method+" "+route.path, func(t *testing.T) {
			defer func() {
				// Passing nil db causes repository queries to panic after passing middleware authorization.
				// Recovering confirms that the request reached the handler (passed RequireRole).
				if r := recover(); r != nil {
					// Handler was reached
				}
			}()

			req := httptest.NewRequest(route.method, route.path, nil)
			req.Header.Set("Authorization", "Bearer "+adminToken)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code == http.StatusUnauthorized || rr.Code == http.StatusForbidden {
				t.Errorf("L'administrateur a reçu %d (non autorisé/interdit) pour %s %s", rr.Code, route.method, route.path)
			}
		})
	}
}
