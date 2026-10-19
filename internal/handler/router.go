package handler

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/LMV0911/Go_e-commerce/internal/middleware"
	"github.com/LMV0911/Go_e-commerce/internal/models"
	"github.com/LMV0911/Go_e-commerce/internal/repository"
)

func SetupRoutes(db *sql.DB) http.Handler {
	mux := http.NewServeMux()

	productRepo := repository.NewProductRepository(db)
	userRepo := repository.NewUserRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	cartRepo := repository.NewCartRepository(db)

	adminHandler := NewAdminHandler(productRepo, userRepo, orderRepo)
	authHandler := NewAuthHandler(userRepo)
	clientHandler := NewClientHandler(productRepo, cartRepo, orderRepo)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})


	mux.HandleFunc("GET /api/products", clientHandler.ListProducts)
	mux.HandleFunc("GET /api/categories", clientHandler.ListCategories)
	mux.HandleFunc("GET /api/cart", clientHandler.GetCart)
	mux.HandleFunc("POST /api/cart/items", clientHandler.AddToCart)
	mux.HandleFunc("PUT /api/cart/items", clientHandler.UpdateCartItem)
	mux.HandleFunc("DELETE /api/cart/items", clientHandler.RemoveCartItem)
	mux.HandleFunc("DELETE /api/cart", clientHandler.ClearCart)
	mux.HandleFunc("POST /api/checkout", clientHandler.Checkout)
	mux.HandleFunc("GET /api/orders", clientHandler.ListOrders)

	mux.HandleFunc("POST /api/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/auth/confirm", authHandler.Confirm)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("POST /api/auth/reset-password", authHandler.ResetPassword)
	mux.HandleFunc("GET /api/auth/me", middleware.RequireAuth(authHandler.Me))

	mux.HandleFunc("GET /api/admin/products", middleware.RequireRole(models.RoleAdmin, adminHandler.ListProducts))
	mux.HandleFunc("GET /api/admin/products/{id}", middleware.RequireRole(models.RoleAdmin, adminHandler.GetProduct))
	mux.HandleFunc("POST /api/admin/products", middleware.RequireRole(models.RoleAdmin, adminHandler.CreateProduct))
	mux.HandleFunc("PUT /api/admin/products/{id}", middleware.RequireRole(models.RoleAdmin, adminHandler.UpdateProduct))
	mux.HandleFunc("DELETE /api/admin/products/{id}", middleware.RequireRole(models.RoleAdmin, adminHandler.DeleteProduct))
	mux.HandleFunc("PUT /api/admin/categories/rename", middleware.RequireRole(models.RoleAdmin, adminHandler.RenameCategory))

	mux.HandleFunc("GET /api/admin/users", middleware.RequireRole(models.RoleAdmin, adminHandler.ListUsers))
	mux.HandleFunc("GET /api/admin/users/{id}", middleware.RequireRole(models.RoleAdmin, adminHandler.GetUser))
	mux.HandleFunc("POST /api/admin/users", middleware.RequireRole(models.RoleAdmin, adminHandler.CreateUser))
	mux.HandleFunc("POST /api/admin/users/{id}/confirm", middleware.RequireRole(models.RoleAdmin, adminHandler.ConfirmUser))
	mux.HandleFunc("PUT /api/admin/users/{id}", middleware.RequireRole(models.RoleAdmin, adminHandler.UpdateUser))
	mux.HandleFunc("DELETE /api/admin/users/{id}", middleware.RequireRole(models.RoleAdmin, adminHandler.DeleteUser))

	mux.HandleFunc("GET /api/admin/orders", middleware.RequireRole(models.RoleAdmin, adminHandler.ListOrders))
	mux.HandleFunc("GET /api/admin/orders/{id}", middleware.RequireRole(models.RoleAdmin, adminHandler.GetOrder))
	mux.HandleFunc("PUT /api/admin/orders/{id}/status", middleware.RequireRole(models.RoleAdmin, adminHandler.UpdateOrderStatus))
	mux.HandleFunc("POST /api/admin/orders/manual", middleware.RequireRole(models.RoleAdmin, adminHandler.CreateManualOrder))

	return middleware.AuthMiddleware(mux)
}
