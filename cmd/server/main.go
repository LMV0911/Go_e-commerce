package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	fmt.Println("Démarrage du serveur e-commerce sur le port 8080...")

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "OK")
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Erreur du serveur: %v\n", err)
	}
}
