package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LMV0911/Go_e-commerce/internal/handler"
	"github.com/LMV0911/Go_e-commerce/internal/models"
)

func TestAuthHandler_ValidationErrors(t *testing.T) {
	authHandler := handler.NewAuthHandler(nil)

	t.Run("Register with invalid email", func(t *testing.T) {
		payload, _ := json.Marshal(models.RegisterRequest{
			Email:    "pas-un-email",
			Password: "password123",
		})
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		authHandler.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Code attendu 400, obtenu %d", rr.Code)
		}
	})

	t.Run("Register with short password", func(t *testing.T) {
		payload, _ := json.Marshal(models.RegisterRequest{
			Email:    "test@example.com",
			Password: "123",
		})
		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		authHandler.Register(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Code attendu 400, obtenu %d", rr.Code)
		}
	})

	t.Run("Confirm with empty parameters", func(t *testing.T) {
		payload, _ := json.Marshal(models.ConfirmRequest{
			Email: "",
			Code:  "",
		})
		req := httptest.NewRequest("POST", "/api/auth/confirm", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		authHandler.Confirm(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Code attendu 400, obtenu %d", rr.Code)
		}
	})

	t.Run("Login with empty body", func(t *testing.T) {
		payload, _ := json.Marshal(models.LoginRequest{
			Email:    "",
			Password: "",
		})
		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		authHandler.Login(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Code attendu 400, obtenu %d", rr.Code)
		}
	})

	t.Run("ForgotPassword with empty email", func(t *testing.T) {
		payload, _ := json.Marshal(models.ForgotPasswordRequest{
			Email: "",
		})
		req := httptest.NewRequest("POST", "/api/auth/forgot-password", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		authHandler.ForgotPassword(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Code attendu 400, obtenu %d", rr.Code)
		}
	})

	t.Run("ResetPassword with short password", func(t *testing.T) {
		payload, _ := json.Marshal(models.ResetPasswordRequest{
			Email:       "test@example.com",
			Code:        "123456",
			NewPassword: "123",
		})
		req := httptest.NewRequest("POST", "/api/auth/reset-password", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		authHandler.ResetPassword(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Code attendu 400, obtenu %d", rr.Code)
		}
	})

	t.Run("Me without context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/auth/me", nil)
		rr := httptest.NewRecorder()

		authHandler.Me(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Code attendu 401 Unauthorized, obtenu %d", rr.Code)
		}
	})
}
