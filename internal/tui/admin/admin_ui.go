package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type viewState int

const (
	viewMenu viewState = iota
	viewListResults
	viewCreateProduct
	viewConfirmUser
)

type AdminModel struct {
	choices      []string
	cursor       int
	status       string
	items        []string
	serverURL    string
	state        viewState
	step         int
	inputBuffer  string
	formName     string
	formCategory string
	formPriceHT  string
	formStock    string
	resultsTitle string
}

func NewAdminModel(serverURL string) AdminModel {
	return AdminModel{
		choices: []string{
			"📦 1. Lister les Produits",
			"➕ 2. Créer un nouveau Produit (Génère PDT-XXXXXX)",
			"👥 3. Lister les Utilisateurs",
			"✅ 4. Valider/Confirmer un compte Utilisateur",
			"📋 5. Lister les Commandes",
			"🚪 6. Quitter",
		},
		status:    "Connecté au serveur Admin : " + serverURL,
		serverURL: serverURL,
		state:     viewMenu,
	}
}

func (m AdminModel) Init() tea.Cmd {
	return nil
}

func (m AdminModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state == viewListResults {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc", "enter", "backspace", "b", " ", "x", "q":
				m.state = viewMenu
				return m, nil
			}
			return m, nil
		}

		if m.state == viewCreateProduct {
			return m.updateCreateProduct(msg)
		}
		if m.state == viewConfirmUser {
			return m.updateConfirmUser(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q", "x":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			switch m.cursor {
			case 0:
				m.resultsTitle = "LISTE DES PRODUITS"
				m.fetchProducts()
				m.state = viewListResults
			case 1:
				m.state = viewCreateProduct
				m.step = 0
				m.inputBuffer = ""
				m.formName = ""
				m.formCategory = ""
				m.formPriceHT = ""
				m.formStock = ""
			case 2:
				m.resultsTitle = "LISTE DES UTILISATEURS"
				m.fetchUsers()
				m.state = viewListResults
			case 3:
				m.state = viewConfirmUser
				m.inputBuffer = ""
			case 4:
				m.resultsTitle = "LISTE DES COMMANDES"
				m.fetchOrders()
				m.state = viewListResults
			case 5:
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m AdminModel) updateCreateProduct(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenu
		m.inputBuffer = ""
		return m, nil
	case tea.KeyBackspace:
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
	case tea.KeyEnter:
		val := strings.TrimSpace(m.inputBuffer)
		m.inputBuffer = ""

		switch m.step {
		case 0:
			if val == "" {
				m.status = "❌ Le nom du produit ne peut pas être vide"
				return m, nil
			}
			m.formName = val
			m.step = 1
		case 1:
			if val == "" {
				val = "Général"
			}
			m.formCategory = val
			m.step = 2
		case 2:
			p, err := strconv.ParseFloat(val, 64)
			if err != nil || p <= 0 {
				m.status = "❌ Le prix HT doit être un nombre positif (ex: 49.99)"
				return m, nil
			}
			m.formPriceHT = val
			m.step = 3
		case 3:
			st, err := strconv.Atoi(val)
			if err != nil || st < 0 {
				m.status = "❌ Le stock doit être un entier positif (ex: 10)"
				return m, nil
			}
			m.formStock = val
			return m.submitCreateProduct()
		}
	default:
		if msg.Type == tea.KeyRunes {
			m.inputBuffer += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			m.inputBuffer += " "
		}
	}
	return m, nil
}

func (m AdminModel) submitCreateProduct() (tea.Model, tea.Cmd) {
	priceHT, _ := strconv.ParseFloat(m.formPriceHT, 64)
	stock, _ := strconv.Atoi(m.formStock)

	payload := map[string]interface{}{
		"name":        m.formName,
		"category":    m.formCategory,
		"price_ht":    priceHT,
		"price_ttc":   priceHT * 1.20,
		"stock":       stock,
		"description": "Créé via CLI Admin",
	}

	bodyBytes, _ := json.Marshal(payload)
	resp, err := http.Post(m.serverURL+"/api/admin/products", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.status = "❌ Erreur de création du produit: " + err.Error()
		m.state = viewMenu
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		m.status = fmt.Sprintf("❌ Erreur serveur lors de la création (code %d)", resp.StatusCode)
		m.state = viewMenu
		return m, nil
	}

	var created struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&created)

	m.status = fmt.Sprintf("✅ Produit [%s] \"%s\" créé avec succès !", created.ID, m.formName)
	m.state = viewMenu
	return m, nil
}

func (m AdminModel) updateConfirmUser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenu
		m.inputBuffer = ""
		return m, nil
	case tea.KeyBackspace:
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
	case tea.KeyEnter:
		idStr := strings.TrimSpace(m.inputBuffer)
		m.inputBuffer = ""
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			m.status = "❌ Veuillez saisir un ID utilisateur valide"
			return m, nil
		}

		resp, err := http.Post(fmt.Sprintf("%s/api/admin/users/%d/confirm", m.serverURL, id), "application/json", nil)
		if err != nil {
			m.status = "❌ Erreur réseau: " + err.Error()
			m.state = viewMenu
			return m, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			m.status = fmt.Sprintf("❌ Impossible de confirmer l'utilisateur #%d", id)
		} else {
			m.status = fmt.Sprintf("✅ Compte utilisateur #%d validé et confirmé avec succès !", id)
		}
		m.state = viewMenu
	default:
		if msg.Type == tea.KeyRunes {
			m.inputBuffer += string(msg.Runes)
		}
	}
	return m, nil
}

func (m *AdminModel) fetchProducts() {
	m.status = "Chargement des produits..."
	m.items = nil

	resp, err := http.Get(m.serverURL + "/api/admin/products")
	if err != nil {
		m.status = "❌ Erreur connexion serveur: " + err.Error()
		return
	}
	defer resp.Body.Close()

	var products []struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		Category string  `json:"category"`
		PriceTTC float64 `json:"price_ttc"`
		Stock    int     `json:"stock"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&products); err != nil {
		m.status = "❌ Erreur décodage produits"
		return
	}

	m.status = fmt.Sprintf("✅ %d produits trouvés", len(products))
	for _, p := range products {
		m.items = append(m.items, fmt.Sprintf("[%s] %s (%s) - %.2f€ TTC - Stock: %d", p.ID, p.Name, p.Category, p.PriceTTC, p.Stock))
	}
}

func (m *AdminModel) fetchUsers() {
	m.status = "Chargement des utilisateurs..."
	m.items = nil

	resp, err := http.Get(m.serverURL + "/api/admin/users")
	if err != nil {
		m.status = "❌ Erreur connexion serveur: " + err.Error()
		return
	}
	defer resp.Body.Close()

	var users []struct {
		ID          int    `json:"id"`
		Email       string `json:"email"`
		Role        string `json:"role"`
		IsConfirmed bool   `json:"is_confirmed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		m.status = "❌ Erreur décodage utilisateurs"
		return
	}

	m.status = fmt.Sprintf("✅ %d utilisateurs trouvés", len(users))
	for _, u := range users {
		confirmed := "❌ Non confirmé"
		if u.IsConfirmed {
			confirmed = "✅ Confirmé"
		}
		m.items = append(m.items, fmt.Sprintf("ID #%d | %s | Rôle: %s | Status: %s", u.ID, u.Email, u.Role, confirmed))
	}
}

func (m *AdminModel) fetchOrders() {
	m.status = "Chargement des commandes..."
	m.items = nil

	resp, err := http.Get(m.serverURL + "/api/admin/orders")
	if err != nil {
		m.status = "❌ Erreur connexion serveur: " + err.Error()
		return
	}
	defer resp.Body.Close()

	var orders []struct {
		ID        string  `json:"id"`
		UserEmail string  `json:"user_email"`
		Status    string  `json:"status"`
		TotalTTC  float64 `json:"total_ttc"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		m.status = "❌ Erreur décodage commandes"
		return
	}

	m.status = fmt.Sprintf("✅ %d commandes trouvées", len(orders))
	for _, o := range orders {
		m.items = append(m.items, fmt.Sprintf("[%s] Client: %s | Statut: %s | Total: %.2f€ TTC", o.ID, o.UserEmail, o.Status, o.TotalTTC))
	}
}

func (m AdminModel) View() string {
	var s strings.Builder

	s.WriteString("===============================================\n")
	s.WriteString("       🛠️  BACK-OFFICE ADMIN E-COMMERCE CLI     \n")
	s.WriteString("===============================================\n\n")

	if m.state == viewListResults {
		s.WriteString(fmt.Sprintf("📋 %s (%s) :\n\n", m.resultsTitle, m.status))
		if len(m.items) == 0 {
			s.WriteString("  (Aucun résultat trouvé)\n")
		} else {
			for _, item := range m.items {
				s.WriteString("  • " + item + "\n")
			}
		}
		s.WriteString("\n-----------------------------------------------\n")
		s.WriteString("👉 Appuyez sur [Entrée], [Échap], [x] ou [b] pour revenir au Menu Principal\n")
		return s.String()
	}

	if m.state == viewCreateProduct {
		s.WriteString("➕ FORMULAIRE DE CRÉATION DE PRODUIT :\n\n")
		switch m.step {
		case 0:
			s.WriteString("👉 Étape 1/4 - Entrez le NOM du produit :\n")
		case 1:
			s.WriteString(fmt.Sprintf("✓ Nom : %s\n", m.formName))
			s.WriteString("👉 Étape 2/4 - Entrez la CATÉGORIE (ex: Périphériques, Écrans, Audio) :\n")
		case 2:
			s.WriteString(fmt.Sprintf("✓ Nom : %s\n", m.formName))
			s.WriteString(fmt.Sprintf("✓ Catégorie : %s\n", m.formCategory))
			s.WriteString("👉 Étape 3/4 - Entrez le PRIX HT en € (ex: 45.50) :\n")
		case 3:
			s.WriteString(fmt.Sprintf("✓ Nom : %s\n", m.formName))
			s.WriteString(fmt.Sprintf("✓ Catégorie : %s\n", m.formCategory))
			s.WriteString(fmt.Sprintf("✓ Prix HT : %s €\n", m.formPriceHT))
			s.WriteString("👉 Étape 4/4 - Entrez la quantité en STOCK (ex: 20) :\n")
		}
		s.WriteString(fmt.Sprintf("\n> %s █\n\n", m.inputBuffer))
		s.WriteString("Appuyez sur [Entrée] pour valider, ou [Échap] pour annuler.\n")
		return s.String()
	}

	if m.state == viewConfirmUser {
		s.WriteString("✅ CONFIRMATION MANUELLE D'UN UTILISATEUR :\n\n")
		s.WriteString("👉 Entrez l'ID numérique de l'utilisateur à valider (ex: 1) :\n")
		s.WriteString(fmt.Sprintf("\n> %s █\n\n", m.inputBuffer))
		s.WriteString("Appuyez sur [Entrée] pour confirmer, ou [Échap] pour annuler.\n")
		return s.String()
	}

	s.WriteString("Statut : " + m.status + "\n\n")
	s.WriteString("Faites votre choix avec les flèches [↑/↓] puis Appuyez sur [Entrée] :\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = "👉"
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
	}

	s.WriteString("\nAppuyez sur 'x', 'q' ou 'Ctrl+C' pour quitter.\n")

	return s.String()
}
