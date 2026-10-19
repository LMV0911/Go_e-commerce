package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/LMV0911/Go_e-commerce/internal/middleware"
	"github.com/LMV0911/Go_e-commerce/internal/repository"
)

type ClientHandler struct {
	productRepo *repository.ProductRepository
	cartRepo    *repository.CartRepository
	orderRepo   *repository.OrderRepository
}

func NewClientHandler(productRepo *repository.ProductRepository, cartRepo *repository.CartRepository, orderRepo *repository.OrderRepository) *ClientHandler {
	return &ClientHandler{
		productRepo: productRepo,
		cartRepo:    cartRepo,
		orderRepo:   orderRepo,
	}
}

func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *ClientHandler) getUserID(r *http.Request) int {
	if uid, ok := middleware.GetUserID(r.Context()); ok && uid > 0 {
		return uid
	}
	if uidStr := r.URL.Query().Get("user_id"); uidStr != "" {
		if uid, err := strconv.Atoi(uidStr); err == nil && uid > 0 {
			return uid
		}
	}
	return 1 
}

func (h *ClientHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	name := r.URL.Query().Get("name")
	category := r.URL.Query().Get("category")
	description := r.URL.Query().Get("description")
	priceType := strings.ToLower(r.URL.Query().Get("price_type"))
	sortBy := r.URL.Query().Get("sort")

	var minPrice, maxPrice *float64
	if minStr := r.URL.Query().Get("min_price"); minStr != "" {
		if v, err := strconv.ParseFloat(minStr, 64); err == nil {
			minPrice = &v
		}
	}
	if maxStr := r.URL.Query().Get("max_price"); maxStr != "" {
		if v, err := strconv.ParseFloat(maxStr, 64); err == nil {
			maxPrice = &v
		}
	}

	products, err := h.productRepo.Search(q, name, category, description, priceType, minPrice, maxPrice, sortBy)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, products)
}

func (h *ClientHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.productRepo.ListCategories()
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, categories)
}

func (h *ClientHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	cart, err := h.cartRepo.GetOrCreate(userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, cart)
}

func (h *ClientHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	var body struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Corps de requête invalide"})
		return
	}

	if body.Quantity <= 0 {
		body.Quantity = 1
	}

	cart, err := h.cartRepo.AddItem(userID, body.ProductID, body.Quantity)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, cart)
}

func (h *ClientHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	var body struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Corps de requête invalide"})
		return
	}

	cart, err := h.cartRepo.UpdateItem(userID, body.ProductID, body.Quantity)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, cart)
}

func (h *ClientHandler) RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	productID := r.URL.Query().Get("product_id")
	if productID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "product_id requis"})
		return
	}

	cart, err := h.cartRepo.RemoveItem(userID, productID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusOK, cart)
}

func (h *ClientHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	cart, err := h.cartRepo.Clear(userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, cart)
}

func (h *ClientHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)

	var payload struct {
		CardNumber     string `json:"card_number"`
		ExpDate        string `json:"exp_date"`
		CVC            string `json:"cvc"`
		CardholderName string `json:"cardholder_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Informations de paiement invalides"})
		return
	}

	cleanCard := strings.ReplaceAll(payload.CardNumber, " ", "")
	cleanCard = strings.ReplaceAll(cleanCard, "-", "")
	if len(cleanCard) != 16 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Numéro de carte invalide (16 chiffres requis)"})
		return
	}
	if len(payload.CVC) != 3 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Code CVC invalide (3 chiffres requis)"})
		return
	}
	if len(payload.ExpDate) < 4 {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "Date d'expiration invalide (format MM/YY requis)"})
		return
	}

	cart, err := h.cartRepo.GetOrCreate(userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	order, err := h.orderRepo.Checkout(userID, cart, payload.CardNumber, payload.ExpDate, payload.CVC, payload.CardholderName)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	respondJSON(w, http.StatusCreated, order)
}

func (h *ClientHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserID(r)
	orders, err := h.orderRepo.ListByUserID(userID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	respondJSON(w, http.StatusOK, orders)
}
