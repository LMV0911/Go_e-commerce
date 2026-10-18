package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LMV0911/Go_e-commerce/internal/tui/admin"
	"github.com/LMV0911/Go_e-commerce/internal/tui/client"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	wishbm "github.com/charmbracelet/wish/bubbletea"
	wishlog "github.com/charmbracelet/wish/logging"
)

const (
	host = "0.0.0.0"
	port = 2222
)

func main() {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = "http://server:8080"
	}

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, fmt.Sprintf("%d", port))),
		wish.WithHostKeyPath(".ssh/term_info_ed25519"),
		wish.WithMiddleware(
			wishbm.Middleware(teaHandler(serverURL)),
			wishlog.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln("Impossible de démarrer le serveur SSH:", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Printf("🚀 Serveur SSH Charm Wish démarré sur %s:%d\n", host, port)

	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatalln("Erreur pendant l'exécution du serveur SSH:", err)
		}
	}()

	<-done
	log.Println("Arrêt du serveur SSH...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln("Erreur lors de l'arrêt du serveur SSH:", err)
	}
}

func teaHandler(serverURL string) func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	return func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
		if strings.EqualFold(s.User(), "client") {
			m := client.NewClientModel(serverURL)
			return m, []tea.ProgramOption{tea.WithAltScreen()}
		}
		m := admin.NewAdminModel(serverURL)
		return m, []tea.ProgramOption{tea.WithAltScreen()}
	}
}
