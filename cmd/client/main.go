package main

import (
	"fmt"
	"os"

	"github.com/LMV0911/Go_e-commerce/internal/tui/client"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	p := tea.NewProgram(client.NewClientModel(serverURL))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Erreur d'exécution TUI Client: %v\n", err)
		os.Exit(1)
	}
}
