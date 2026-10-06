package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/LMV0911/Go_e-commerce/internal/models"
	"github.com/LMV0911/Go_e-commerce/internal/repository"
)

type AdminHandler struct {
	productRepo *repository.ProductRepository
	userRepo    *repository.UserRepository
	orderRepo   *repository.OrderRepository
}

func NewAdminHandler(pRepo *repository.ProductRepository, uRepo *repository.UserRepository, oRepo *repository.OrderRepository) *AdminHandler {
	return &AdminHandler{
		productRepo: pRepo,
		userRepo:    uRepo,
		orderRepo:   oRepo,
	}
}

func (h *AdminHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	products, err := h.productRepo.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Erreur récupération produits: %v", err)})
		return
	}
	if products == nil {
		products = []models.Product{}
	}
	json.NewEncoder(w).Encode(products)
}

func (h *AdminHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	id := r.PathValue("id")

	product, err := h.productRepo.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if product == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Produit introuvable"})
		return
	}
	json.NewEncoder(w).Encode(product)
}

func (h *AdminHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	defer r.Body.Close()

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format JSON invalide"})
		return
	}

	if p.Name == "" || p.PriceHT <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Le nom et un prix HT supérieur à zéro sont obligatoires"})
		return
	}

	if p.PriceTTC <= 0 {
		p.PriceTTC = p.PriceHT * 1.20
	}

	if err := h.productRepo.Create(&p); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Erreur création produit: %v", err)})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *AdminHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	id := r.PathValue("id")
	defer r.Body.Close()

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format JSON invalide"})
		return
	}

	p.ID = id
	if err := h.productRepo.Update(&p); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Produit mis à jour avec succès"})
}

func (h *AdminHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	id := r.PathValue("id")

	if err := h.productRepo.Delete(id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Produit supprimé avec succès"})
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	users, err := h.userRepo.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Erreur récupération utilisateurs: %v", err)})
		return
	}
	if users == nil {
		users = []models.User{}
	}
	json.NewEncoder(w).Encode(users)
}

func (h *AdminHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Identifiant utilisateur invalide"})
		return
	}

	user, err := h.userRepo.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Utilisateur introuvable"})
		return
	}
	json.NewEncoder(w).Encode(user)
}

func (h *AdminHandler) ConfirmUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Identifiant utilisateur invalide"})
		return
	}

	if err := h.userRepo.ConfirmAccount(id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Compte utilisateur confirmé avec succès"})
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Identifiant utilisateur invalide"})
		return
	}
	defer r.Body.Close()

	var u models.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format JSON invalide"})
		return
	}

	u.ID = id
	if err := h.userRepo.Update(&u); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Utilisateur mis à jour avec succès"})
}

func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Identifiant utilisateur invalide"})
		return
	}

	if err := h.userRepo.Delete(id); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": "Utilisateur supprimé avec succès"})
}

func (h *AdminHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	orders, err := h.orderRepo.ListAll()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Erreur récupération commandes: %v", err)})
		return
	}
	if orders == nil {
		orders = []models.Order{}
	}
	json.NewEncoder(w).Encode(orders)
}

func (h *AdminHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	id := r.PathValue("id")

	order, err := h.orderRepo.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	if order == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Commande introuvable"})
		return
	}
	json.NewEncoder(w).Encode(order)
}

func (h *AdminHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	id := r.PathValue("id")
	defer r.Body.Close()

	var req models.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format JSON invalide"})
		return
	}

	if err := h.orderRepo.UpdateStatus(id, req.Status, req.CancelReason); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Statut de commande mis à jour"})
}

func (h *AdminHandler) CreateManualOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur de lecture de la requête"})
		return
	}

	var req models.CreateManualOrderRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format JSON invalide"})
		return
	}

	order, err := h.orderRepo.CreateManualOrder(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
