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
	viewMainMenu viewState = iota
	viewMenuProducts
	viewMenuUsers
	viewMenuOrders
	viewListResults
	viewCreateProduct
	viewDeleteProduct
	viewRenameCategory
	viewCreateUser
	viewConfirmUser
	viewDeleteUser
	viewUpdateOrderStatus
	viewCreateManualOrder
)

type AdminModel struct {
	cursor       int
	status       string
	items        []string
	serverURL    string
	state        viewState
	returnState  viewState
	step         int
	inputBuffer  string
	formName     string
	formCategory string
	formPriceHT  string
	formStock    string
	formOldCat   string
	formNewCat   string
	formEmail    string
	formPassword string
	formRole     string
	formOrderID  string
	formStatus   string
	formReason   string
	formUserID   string
	formProdID   string
	formQuantity string
	resultsTitle string
}

func NewAdminModel(serverURL string) AdminModel {
	return AdminModel{
		cursor:      0,
		status:      "Connecté au serveur Admin : " + serverURL,
		serverURL:   serverURL,
		state:       viewMainMenu,
		returnState: viewMainMenu,
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
			case "esc", "enter", "backspace", " ", "x", "q":
				m.state = m.returnState
				m.cursor = 0
				return m, nil
			}
			return m, nil
		}

		switch m.state {
		case viewCreateProduct:
			return m.updateCreateProduct(msg)
		case viewDeleteProduct:
			return m.updateDeleteProduct(msg)
		case viewRenameCategory:
			return m.updateRenameCategory(msg)
		case viewCreateUser:
			return m.updateCreateUser(msg)
		case viewConfirmUser:
			return m.updateConfirmUser(msg)
		case viewDeleteUser:
			return m.updateDeleteUser(msg)
		case viewUpdateOrderStatus:
			return m.updateOrderStatus(msg)
		case viewCreateManualOrder:
			return m.updateCreateManualOrder(msg)
		case viewMainMenu:
			return m.updateMainMenu(msg)
		case viewMenuProducts:
			return m.updateMenuProducts(msg)
		case viewMenuUsers:
			return m.updateMenuUsers(msg)
		case viewMenuOrders:
			return m.updateMenuOrders(msg)
		}
	}
	return m, nil
}

func (m AdminModel) updateMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuLen := 4
	switch msg.String() {
	case "ctrl+c", "q", "x":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < menuLen-1 {
			m.cursor++
		}
	case "enter", " ":
		selected := m.cursor
		m.cursor = 0
		switch selected {
		case 0:
			m.state = viewMenuProducts
		case 1:
			m.state = viewMenuUsers
		case 2:
			m.state = viewMenuOrders
		case 3:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m AdminModel) updateMenuProducts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuLen := 5
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q", "x":
		m.state = viewMainMenu
		m.cursor = 0
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < menuLen-1 {
			m.cursor++
		}
	case "enter", " ":
		selected := m.cursor
		m.cursor = 0
		m.step = 0
		m.inputBuffer = ""
		switch selected {
		case 0:
			m.resultsTitle = "LISTE DES PRODUITS"
			m.returnState = viewMenuProducts
			m.fetchProducts()
			m.state = viewListResults
		case 1:
			m.state = viewCreateProduct
			m.formName = ""
			m.formCategory = ""
			m.formPriceHT = ""
			m.formStock = ""
		case 2:
			m.state = viewDeleteProduct
		case 3:
			m.state = viewRenameCategory
			m.formOldCat = ""
			m.formNewCat = ""
		case 4:
			m.state = viewMainMenu
		}
	}
	return m, nil
}

func (m AdminModel) updateMenuUsers(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuLen := 5
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q", "x":
		m.state = viewMainMenu
		m.cursor = 1
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < menuLen-1 {
			m.cursor++
		}
	case "enter", " ":
		selected := m.cursor
		m.cursor = 0
		m.step = 0
		m.inputBuffer = ""
		switch selected {
		case 0:
			m.resultsTitle = "LISTE DES UTILISATEURS"
			m.returnState = viewMenuUsers
			m.fetchUsers()
			m.state = viewListResults
		case 1:
			m.state = viewCreateUser
			m.formEmail = ""
			m.formPassword = ""
			m.formRole = ""
		case 2:
			m.state = viewConfirmUser
		case 3:
			m.state = viewDeleteUser
		case 4:
			m.state = viewMainMenu
			m.cursor = 1
		}
	}
	return m, nil
}

func (m AdminModel) updateMenuOrders(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuLen := 4
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q", "x":
		m.state = viewMainMenu
		m.cursor = 2
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < menuLen-1 {
			m.cursor++
		}
	case "enter", " ":
		selected := m.cursor
		m.cursor = 0
		m.step = 0
		m.inputBuffer = ""
		switch selected {
		case 0:
			m.resultsTitle = "LISTE DES COMMANDES"
			m.returnState = viewMenuOrders
			m.fetchOrders()
			m.state = viewListResults
		case 1:
			m.state = viewUpdateOrderStatus
			m.formOrderID = ""
			m.formStatus = ""
			m.formReason = ""
		case 2:
			m.state = viewCreateManualOrder
			m.formUserID = ""
			m.formProdID = ""
			m.formQuantity = ""
		case 3:
			m.state = viewMainMenu
			m.cursor = 2
		}
	}
	return m, nil
}

func (m AdminModel) updateCreateProduct(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuProducts
		m.inputBuffer = ""
		m.step = 0
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
		m.state = viewMenuProducts
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		m.status = fmt.Sprintf("❌ Erreur serveur lors de la création (code %d)", resp.StatusCode)
		m.state = viewMenuProducts
		return m, nil
	}

	var created struct {
		ID string `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&created)

	m.status = fmt.Sprintf("✅ Produit [%s] \"%s\" créé avec succès !", created.ID, m.formName)
	m.state = viewMenuProducts
	return m, nil
}

func (m AdminModel) updateDeleteProduct(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuProducts
		m.inputBuffer = ""
		m.step = 0
		return m, nil
	case tea.KeyBackspace:
		if len(m.inputBuffer) > 0 {
			m.inputBuffer = m.inputBuffer[:len(m.inputBuffer)-1]
		}
	case tea.KeyEnter:
		prodID := strings.TrimSpace(m.inputBuffer)
		m.inputBuffer = ""
		if prodID == "" {
			m.status = "❌ ID produit requis"
			return m, nil
		}

		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/admin/products/%s", m.serverURL, prodID), nil)
		if err != nil {
			m.status = "❌ Erreur requête: " + err.Error()
			m.state = viewMenuProducts
			return m, nil
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			m.status = "❌ Erreur réseau: " + err.Error()
			m.state = viewMenuProducts
			return m, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			m.status = fmt.Sprintf("❌ Impossible de supprimer le produit %s (inexistant)", prodID)
		} else {
			m.status = fmt.Sprintf("✅ Produit %s supprimé avec succès !", prodID)
		}
		m.state = viewMenuProducts
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			m.inputBuffer += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			m.inputBuffer += " "
		}
	}
	return m, nil
}

func (m AdminModel) updateRenameCategory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuProducts
		m.inputBuffer = ""
		m.step = 0
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
				m.status = "❌ L'ancienne catégorie ne peut pas être vide"
				return m, nil
			}
			m.formOldCat = val
			m.step = 1
		case 1:
			if val == "" {
				m.status = "❌ La nouvelle catégorie ne peut pas être vide"
				return m, nil
			}
			m.formNewCat = val
			return m.submitRenameCategory()
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

func (m AdminModel) submitRenameCategory() (tea.Model, tea.Cmd) {
	payload := map[string]string{
		"old_category": m.formOldCat,
		"new_category": m.formNewCat,
	}
	bodyBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPut, m.serverURL+"/api/admin/categories/rename", bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.status = "❌ Erreur requête: " + err.Error()
		m.state = viewMenuProducts
		return m, nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.status = "❌ Erreur réseau: " + err.Error()
		m.state = viewMenuProducts
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.status = "❌ Erreur lors du renommage de la catégorie"
	} else {
		var res struct {
			RowsAffected int64 `json:"rows_affected"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		m.status = fmt.Sprintf("✅ Catégorie \"%s\" renommée en \"%s\" sur %d produit(s) !", m.formOldCat, m.formNewCat, res.RowsAffected)
	}
	m.state = viewMenuProducts
	return m, nil
}

func (m AdminModel) updateCreateUser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuUsers
		m.inputBuffer = ""
		m.step = 0
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
			if val == "" || !strings.Contains(val, "@") {
				m.status = "❌ Email valide requis"
				return m, nil
			}
			m.formEmail = val
			m.step = 1
		case 1:
			if len(val) < 4 {
				m.status = "❌ Le mot de passe doit comporter au moins 4 caractères"
				return m, nil
			}
			m.formPassword = val
			m.step = 2
		case 2:
			role := "client"
			if val == "admin" {
				role = "admin"
			}
			m.formRole = role
			return m.submitCreateUser()
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

func (m AdminModel) submitCreateUser() (tea.Model, tea.Cmd) {
	payload := map[string]string{
		"email":    m.formEmail,
		"password": m.formPassword,
		"role":     m.formRole,
	}
	bodyBytes, _ := json.Marshal(payload)
	resp, err := http.Post(m.serverURL+"/api/admin/users", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.status = "❌ Erreur réseau: " + err.Error()
		m.state = viewMenuUsers
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		m.status = "❌ Erreur lors de la création de l'utilisateur"
	} else {
		m.status = fmt.Sprintf("✅ Utilisateur %s créé avec succès (rôle: %s) !", m.formEmail, m.formRole)
	}
	m.state = viewMenuUsers
	return m, nil
}

func (m AdminModel) updateConfirmUser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuUsers
		m.inputBuffer = ""
		m.step = 0
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
			m.state = viewMenuUsers
			return m, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			m.status = fmt.Sprintf("❌ Impossible de confirmer l'utilisateur #%d", id)
		} else {
			m.status = fmt.Sprintf("✅ Utilisateur #%d confirmé manuellement avec succès !", id)
		}
		m.state = viewMenuUsers
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			m.inputBuffer += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			m.inputBuffer += " "
		}
	}
	return m, nil
}

func (m AdminModel) updateDeleteUser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuUsers
		m.inputBuffer = ""
		m.step = 0
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

		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/admin/users/%d", m.serverURL, id), nil)
		if err != nil {
			m.status = "❌ Erreur: " + err.Error()
			m.state = viewMenuUsers
			return m, nil
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			m.status = "❌ Erreur réseau: " + err.Error()
			m.state = viewMenuUsers
			return m, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			m.status = fmt.Sprintf("❌ Impossible de supprimer l'utilisateur #%d", id)
		} else {
			m.status = fmt.Sprintf("✅ Utilisateur #%d supprimé avec succès !", id)
		}
		m.state = viewMenuUsers
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			m.inputBuffer += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			m.inputBuffer += " "
		}
	}
	return m, nil
}

func (m AdminModel) updateOrderStatus(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuOrders
		m.inputBuffer = ""
		m.step = 0
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
				m.status = "❌ ID de commande requis"
				return m, nil
			}
			m.formOrderID = val
			m.step = 1
		case 1:
			if val == "" {
				val = "expédiée"
			}
			m.formStatus = val
			if val == "annulé" || val == "annulée" {
				m.step = 2
			} else {
				return m.submitUpdateOrderStatus()
			}
		case 2:
			if val == "" {
				m.status = "❌ Motif d'annulation obligatoire pour le statut annulé"
				return m, nil
			}
			m.formReason = val
			return m.submitUpdateOrderStatus()
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

func (m AdminModel) submitUpdateOrderStatus() (tea.Model, tea.Cmd) {
	payload := map[string]string{
		"status":        m.formStatus,
		"cancel_reason": m.formReason,
	}
	bodyBytes, _ := json.Marshal(payload)

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/admin/orders/%s/status", m.serverURL, m.formOrderID), bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.status = "❌ Erreur requête: " + err.Error()
		m.state = viewMenuOrders
		return m, nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.status = "❌ Erreur réseau: " + err.Error()
		m.state = viewMenuOrders
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		m.status = fmt.Sprintf("❌ Erreur: %s", errResp.Error)
	} else {
		m.status = fmt.Sprintf("✅ Statut de la commande %s mis à jour vers \"%s\" !", m.formOrderID, m.formStatus)
	}
	m.state = viewMenuOrders
	return m, nil
}

func (m AdminModel) updateCreateManualOrder(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuOrders
		m.inputBuffer = ""
		m.step = 0
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
			uid, err := strconv.Atoi(val)
			if err != nil || uid <= 0 {
				m.status = "❌ ID utilisateur numérique valide requis"
				return m, nil
			}
			m.formUserID = val
			m.step = 1
		case 1:
			if val == "" {
				m.status = "❌ ID produit requis (ex: PDT-XXXXXX)"
				return m, nil
			}
			m.formProdID = val
			m.step = 2
		case 2:
			qty, err := strconv.Atoi(val)
			if err != nil || qty <= 0 {
				m.status = "❌ Quantité positive requise"
				return m, nil
			}
			m.formQuantity = val
			return m.submitCreateManualOrder()
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

func (m AdminModel) submitCreateManualOrder() (tea.Model, tea.Cmd) {
	uid, _ := strconv.Atoi(m.formUserID)
	qty, _ := strconv.Atoi(m.formQuantity)

	payload := map[string]interface{}{
		"user_id": uid,
		"items": []map[string]interface{}{
			{
				"product_id": m.formProdID,
				"quantity":   qty,
			},
		},
	}

	bodyBytes, _ := json.Marshal(payload)
	resp, err := http.Post(m.serverURL+"/api/admin/orders/manual", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.status = "❌ Erreur réseau: " + err.Error()
		m.state = viewMenuOrders
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		m.status = fmt.Sprintf("❌ Erreur création commande: %s", errResp.Error)
	} else {
		var order struct {
			ID       string  `json:"id"`
			TotalTTC float64 `json:"total_ttc"`
		}
		json.NewDecoder(resp.Body).Decode(&order)
		m.status = fmt.Sprintf("✅ Commande manuelle [%s] créée avec succès ! Total: %.2f€ TTC", order.ID, order.TotalTTC)
	}
	m.state = viewMenuOrders
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
	s.WriteString("       BACK-OFFICE ADMIN E-COMMERCE CLI     \n")
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
		s.WriteString("Appuyez sur [Entrée], [Échap] ou [x] pour revenir au menu\n")
		return s.String()
	}

	if m.state == viewCreateProduct {
		s.WriteString("CRÉATION D'UN PRODUIT (Génère PDT-XXXXXX) :\n\n")
		switch m.step {
		case 0:
			s.WriteString("Étape 1/4 - Entrez le NOM du produit :\n")
		case 1:
			s.WriteString(fmt.Sprintf("✓ Nom : %s\n", m.formName))
			s.WriteString("Étape 2/4 - Entrez la CATÉGORIE (ex: Informatique, Audio, Gaming) :\n")
		case 2:
			s.WriteString(fmt.Sprintf("✓ Nom : %s\n", m.formName))
			s.WriteString(fmt.Sprintf("✓ Catégorie : %s\n", m.formCategory))
			s.WriteString("Étape 3/4 - Entrez le PRIX HT en € (ex: 45.50) :\n")
		case 3:
			s.WriteString(fmt.Sprintf("✓ Nom : %s\n", m.formName))
			s.WriteString(fmt.Sprintf("✓ Catégorie : %s\n", m.formCategory))
			s.WriteString(fmt.Sprintf("✓ Prix HT : %s €\n", m.formPriceHT))
			s.WriteString("Étape 4/4 - Entrez la quantité en STOCK (ex: 20) :\n")
		}
		s.WriteString(fmt.Sprintf("\n> %s █\n\n", m.inputBuffer))
		s.WriteString("Appuyez sur [Entrée] pour valider, ou [Échap] pour annuler.\n")
		return s.String()
	}

	if m.state == viewDeleteProduct {
		s.WriteString("SUPPRESSION D'UN PRODUIT :\n\n")
		s.WriteString("Entrez l'identifiant du produit à supprimer (ex: PDT-SIL1BB) :\n")
		s.WriteString(fmt.Sprintf("\n> %s █\n\n", m.inputBuffer))
		s.WriteString("Appuyez sur [Entrée] pour valider, ou [Échap] pour annuler.\n")
		return s.String()
	}

	if m.state == viewRenameCategory {
		s.WriteString("RENOMMER UNE CATÉGORIE SUR TOUS LES PRODUITS :\n\n")
		switch m.step {
		case 0:
			s.WriteString("Étape 1/2 - Entrez l'ANCIENNE catégorie (ex: imformatique) :\n")
		case 1:
			s.WriteString(fmt.Sprintf("✓ Ancienne catégorie : %s\n", m.formOldCat))
			s.WriteString("Étape 2/2 - Entrez la NOUVELLE catégorie (ex: Informatique) :\n")
		}
		s.WriteString(fmt.Sprintf("\n> %s █\n\n", m.inputBuffer))
		s.WriteString("Appuyez sur [Entrée] pour valider, ou [Échap] pour annuler.\n")
		return s.String()
	}

	if m.state == viewCreateUser {
		s.WriteString("CRÉATION D'UN NOUVEL UTILISATEUR :\n\n")
		switch m.step {
		case 0:
			s.WriteString("Étape 1/3 - Entrez l'EMAIL de l'utilisateur :\n")
		case 1:
			s.WriteString(fmt.Sprintf("✓ Email : %s\n", m.formEmail))
			s.WriteString("Étape 2/3 - Entrez le MOT DE PASSE :\n")
		case 2:
			s.WriteString(fmt.Sprintf("✓ Email : %s\n", m.formEmail))
			s.WriteString("✓ Mot de passe configuré\n")
			s.WriteString("Étape 3/3 - Entrez le RÔLE ('client' ou 'admin', défaut: client) :\n")
		}
		s.WriteString(fmt.Sprintf("\n> %s █\n\n", m.inputBuffer))
		s.WriteString("Appuyez sur [Entrée] pour valider, ou [Échap] pour annuler.\n")
		return s.String()
	}

	if m.state == viewConfirmUser {
		s.WriteString("CONFIRMATION MANUELLE D'UN UTILISATEUR :\n\n")
		s.WriteString("Entrez l'ID numérique de l'utilisateur à valider (ex: 1) :\n")
		s.WriteString(fmt.Sprintf("\n> %s █\n\n", m.inputBuffer))
		s.WriteString("Appuyez sur [Entrée] pour valider, ou [Échap] pour annuler.\n")
		return s.String()
	}

	if m.state == viewDeleteUser {
		s.WriteString("SUPPRESSION D'UN UTILISATEUR :\n\n")
		s.WriteString("Entrez l'ID numérique de l'utilisateur à supprimer (ex: 2) :\n")
		s.WriteString(fmt.Sprintf("\n> %s █\n\n", m.inputBuffer))
		s.WriteString("Appuyez sur [Entrée] pour valider, ou [Échap] pour annuler.\n")
		return s.String()
	}

	if m.state == viewUpdateOrderStatus {
		s.WriteString("MODIFICATION DU STATUT D'UNE COMMANDE :\n\n")
		switch m.step {
		case 0:
			s.WriteString("Étape 1 - Entrez l'ID de la commande (ex: CMD-XXXXXX) :\n")
		case 1:
			s.WriteString(fmt.Sprintf("✓ Commande : %s\n", m.formOrderID))
			s.WriteString("Étape 2 - Entrez le NOUVEAU STATUT ('payé', 'expédié', 'livré', 'annulé') :\n")
		case 2:
			s.WriteString(fmt.Sprintf("✓ Commande : %s\n", m.formOrderID))
			s.WriteString(fmt.Sprintf("✓ Statut : %s\n", m.formStatus))
			s.WriteString("Étape 3 - Entrez le MOTIF D'ANNULATION OBLIGATOIRE :\n")
		}
		s.WriteString(fmt.Sprintf("\n> %s █\n\n", m.inputBuffer))
		s.WriteString("Appuyez sur [Entrée] pour valider, ou [Échap] pour annuler.\n")
		return s.String()
	}

	if m.state == viewCreateManualOrder {
		s.WriteString("CRÉATION D'UNE COMMANDE MANUELLE (ADMIN) :\n\n")
		switch m.step {
		case 0:
			s.WriteString("Étape 1/3 - Entrez l'ID numérique du client (ex: 1) :\n")
		case 1:
			s.WriteString(fmt.Sprintf("✓ Client ID : %s\n", m.formUserID))
			s.WriteString("Étape 2/3 - Entrez l'ID du produit (ex: PDT-A3G3XD) :\n")
		case 2:
			s.WriteString(fmt.Sprintf("✓ Client ID : %s\n", m.formUserID))
			s.WriteString(fmt.Sprintf("✓ Produit ID : %s\n", m.formProdID))
			s.WriteString("Étape 3/3 - Entrez la QUANTITÉ souhaitée (ex: 2) :\n")
		}
		s.WriteString(fmt.Sprintf("\n> %s █\n\n", m.inputBuffer))
		s.WriteString("Appuyez sur [Entrée] pour valider, ou [Échap] pour annuler.\n")
		return s.String()
	}

	var choices []string
	var menuTitle string

	switch m.state {
	case viewMainMenu:
		menuTitle = "MENU PRINCIPAL"
		choices = []string{
			"1. Gestion des Produits",
			"2. Gestion des Utilisateurs",
			"3. Gestion des Commandes",
			"4. Quitter",
		}
	case viewMenuProducts:
		menuTitle = "GESTION DES PRODUITS"
		choices = []string{
			"1. Lister les Produits",
			"2. Ajouter un nouveau Produit",
			"3. Supprimer un Produit",
			"4. Renommer une Catégorie sur tous les produits",
			"5. Retour au Menu Principal",
		}
	case viewMenuUsers:
		menuTitle = "GESTION DES UTILISATEURS"
		choices = []string{
			"1. Lister les Utilisateurs",
			"2. Créer un Utilisateur",
			"3. Confirmer manuellement un compte",
			"4. Supprimer un Utilisateur",
			"5. Retour au Menu Principal",
		}
	case viewMenuOrders:
		menuTitle = "GESTION DES COMMANDES"
		choices = []string{
			"1. Lister les Commandes",
			"2. Modifier le statut d'une Commande (Statut / Motif)",
			"3. Créer une Commande (Client + Produits)",
			"4. Retour au Menu Principal",
		}
	}

	s.WriteString(menuTitle + " :\n\n")
	s.WriteString("Statut : " + m.status + "\n\n")
	s.WriteString("Naviguez avec les flèches [↑/↓] puis validez avec [Entrée] :\n\n")

	for i, choice := range choices {
		cursor := " "
		if m.cursor == i {
			cursor = "👉"
		}
		s.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
	}

	s.WriteString("\nAppuyez sur [Échap] ou 'x' pour revenir en arrière, 'q' pour quitter.\n")

	return s.String()
}
