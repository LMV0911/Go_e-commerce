package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LMV0911/Go_e-commerce/internal/handler"
)

func TestClientHandler_Validation(t *testing.T) {
	clientHandler := handler.NewClientHandler(nil, nil, nil)

	t.Run("Checkout with invalid card number", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]string{
			"card_number":     "1234",
			"exp_date":        "12/28",
			"cvc":             "123",
			"cardholder_name": "Test",
		})
		req := httptest.NewRequest("POST", "/api/checkout", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		clientHandler.Checkout(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Code attendu 400, obtenu %d", rr.Code)
		}
	})

	t.Run("Checkout with invalid CVC", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]string{
			"card_number":     "4532123456789012",
			"exp_date":        "12/28",
			"cvc":             "1", 
			"cardholder_name": "Test",
		})
		req := httptest.NewRequest("POST", "/api/checkout", bytes.NewBuffer(payload))
		rr := httptest.NewRecorder()

		clientHandler.Checkout(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Code attendu 400, obtenu %d", rr.Code)
		}
	})

	t.Run("AddToCart with invalid JSON body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/cart/items", bytes.NewBufferString("ceci-n-est-pas-du-json"))
		rr := httptest.NewRecorder()

		clientHandler.AddToCart(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Code attendu 400, obtenu %d", rr.Code)
		}
	})

	t.Run("RemoveCartItem without product_id parameter", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/cart/items", nil)
		rr := httptest.NewRecorder()

		clientHandler.RemoveCartItem(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Code attendu 400, obtenu %d", rr.Code)
		}
	})
}
