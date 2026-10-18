package handler

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/LMV0911/Go_e-commerce/internal/middleware"
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

	mux.HandleFunc("GET /api/admin/products", adminHandler.ListProducts)
	mux.HandleFunc("GET /api/admin/products/{id}", adminHandler.GetProduct)
	mux.HandleFunc("POST /api/admin/products", adminHandler.CreateProduct)
	mux.HandleFunc("PUT /api/admin/products/{id}", adminHandler.UpdateProduct)
	mux.HandleFunc("DELETE /api/admin/products/{id}", adminHandler.DeleteProduct)

	mux.HandleFunc("GET /api/admin/users", adminHandler.ListUsers)
	mux.HandleFunc("GET /api/admin/users/{id}", adminHandler.GetUser)
	mux.HandleFunc("POST /api/admin/users/{id}/confirm", adminHandler.ConfirmUser)
	mux.HandleFunc("PUT /api/admin/users/{id}", adminHandler.UpdateUser)
	mux.HandleFunc("DELETE /api/admin/users/{id}", adminHandler.DeleteUser)

	mux.HandleFunc("GET /api/admin/orders", adminHandler.ListOrders)
	mux.HandleFunc("GET /api/admin/orders/{id}", adminHandler.GetOrder)
	mux.HandleFunc("PUT /api/admin/orders/{id}/status", adminHandler.UpdateOrderStatus)
	mux.HandleFunc("POST /api/admin/orders/manual", adminHandler.CreateManualOrder)

	return middleware.AuthMiddleware(mux)
}
