package main

import (
	"log"
	"net/http"
	"os"

	"github.com/LMV0911/Go_e-commerce/internal/database"
	"github.com/LMV0911/Go_e-commerce/internal/handler"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Démarrage du serveur E-Commerce sur le port :%s...\n", port)

	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("❌ Erreur fatale de connexion BDD: %v\n", err)
	}
	defer db.Close()

	router := handler.SetupRoutes(db)

	log.Printf("✅ Serveur prêt et écoute sur http://0.0.0.0:%s\n", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("❌ Erreur lors du fonctionnement du serveur: %v\n", err)
	}
}
