package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"strings"

	"github.com/LMV0911/Go_e-commerce/internal/auth"
	"github.com/LMV0911/Go_e-commerce/internal/middleware"
	"github.com/LMV0911/Go_e-commerce/internal/models"
	"github.com/LMV0911/Go_e-commerce/internal/repository"
	"github.com/LMV0911/Go_e-commerce/internal/utils"
)

type AuthHandler struct {
	userRepo *repository.UserRepository
}

func NewAuthHandler(uRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		userRepo: uRepo,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	defer r.Body.Close()

	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format JSON invalide"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if _, err := mail.ParseAddress(req.Email); err != nil || !strings.Contains(req.Email, ".") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format d'adresse email invalide"})
		return
	}

	if len(req.Password) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Le mot de passe doit comporter au moins 6 caractères"})
		return
	}

	existingUser, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Erreur lors de la vérification de l'utilisateur: %v", err)})
		return
	}
	if existingUser != nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{"error": "Un compte avec cette adresse email existe déjà"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors du chiffrement du mot de passe"})
		return
	}

	code := utils.GenerateConfirmationCode()
	newUser := &models.User{
		Email:            req.Email,
		PasswordHash:     hash,
		Role:             models.RoleClient,
		IsConfirmed:      false,
		ConfirmationCode: code,
	}

	if err := h.userRepo.Create(newUser); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Erreur lors de l'enregistrement: %v", err)})
		return
	}

	log.Printf("code confirmation (%s): %s\n", newUser.Email, code)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(models.AuthResponse{
		Message:          "Inscription reussie. Code de confirmation envoye.",
		ConfirmationCode: code,
		User: &models.User{
			ID:          newUser.ID,
			Email:       newUser.Email,
			Role:        newUser.Role,
			IsConfirmed: newUser.IsConfirmed,
			CreatedAt:   newUser.CreatedAt,
		},
	})
}

func (h *AuthHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	defer r.Body.Close()

	var req models.ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format JSON invalide"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Code = strings.TrimSpace(req.Code)
	if req.Email == "" || req.Code == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "L'email et le code de confirmation sont obligatoires"})
		return
	}

	if err := h.userRepo.ConfirmByCode(req.Email, req.Code); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Compte confirmé avec succès ! Vous pouvez maintenant vous connecter.",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	defer r.Body.Close()

	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format JSON invalide"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email et mot de passe requis"})
		return
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur serveur lors de la recherche du compte"})
		return
	}
	if user == nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email ou mot de passe incorrect"})
		return
	}

	if !user.IsConfirmed {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Votre compte n'est pas encore confirmé. Utilisez votre code de confirmation avant de vous connecter.",
		})
		return
	}

	token, err := auth.GenerateToken(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors de la génération du jeton JWT"})
		return
	}

	json.NewEncoder(w).Encode(models.AuthResponse{
		Token:   token,
		Message: "Connexion réussie",
		User: &models.User{
			ID:          user.ID,
			Email:       user.Email,
			Role:        user.Role,
			IsConfirmed: user.IsConfirmed,
			CreatedAt:   user.CreatedAt,
		},
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	defer r.Body.Close()

	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format JSON invalide"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Adresse email obligatoire"})
		return
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur serveur lors de la recherche"})
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Aucun compte associé à cette adresse email"})
		return
	}

	resetCode := utils.GenerateRandomCode(6)
	if err := h.userRepo.SetResetCode(user.Email, resetCode); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Impossible d'enregistrer le code de réinitialisation"})
		return
	}

	log.Printf("code reset (%s): %s\n", user.Email, resetCode)

	json.NewEncoder(w).Encode(models.AuthResponse{
		Message:   "Code de reinitialisation genere.",
		ResetCode: resetCode,
		User: &models.User{
			ID:    user.ID,
			Email: user.Email,
		},
	})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	defer r.Body.Close()

	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Format JSON invalide"})
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	req.Code = strings.TrimSpace(req.Code)
	if req.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "L'adresse email est requise"})
		return
	}

	if len(req.NewPassword) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Le nouveau mot de passe doit comporter au moins 6 caractères"})
		return
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors du hachage du nouveau mot de passe"})
		return
	}

	if err := h.userRepo.ResetPassword(req.Email, req.Code, hash); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Mot de passe réinitialisé avec succès ! Vous pouvez maintenant vous connecter.",
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	userID, ok := middleware.GetUserID(r.Context())
	if !ok || userID <= 0 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Non authentifié"})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Erreur lors de la récupération du profil"})
		return
	}
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "Utilisateur non trouvé"})
		return
	}

	json.NewEncoder(w).Encode(user)
}
