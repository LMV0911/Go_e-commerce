package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/LMV0911/Go_e-commerce/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

type viewState int

const (
	viewLogin viewState = iota
	viewMainMenu
	viewMenuProducts
	viewMenuUsers
	viewMenuOrders
	viewListResults
	viewCreateProduct
	viewUpdateProduct
	viewDeleteProduct
	viewRenameCategory
	viewCreateUser
	viewUpdateUser
	viewConfirmUser
	viewDeleteUser
	viewOrderDetail
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
	token        string
	user         *models.User
	authEmail    string
	authPassword string
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
		status:      "Veuillez vous identifier pour accéder au back-office",
		serverURL:   serverURL,
		state:       viewLogin,
		returnState: viewMainMenu,
	}
}

func (m AdminModel) Init() tea.Cmd {
	return nil
}

func (m AdminModel) doRequest(method, urlStr string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if m.token != "" {
		req.Header.Set("Authorization", "Bearer "+m.token)
	}
	return http.DefaultClient.Do(req)
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
		case viewLogin:
			return m.updateLogin(msg)
		case viewCreateProduct:
			return m.updateCreateProduct(msg)
		case viewUpdateProduct:
			return m.updateUpdateProduct(msg)
		case viewDeleteProduct:
			return m.updateDeleteProduct(msg)
		case viewRenameCategory:
			return m.updateRenameCategory(msg)
		case viewCreateUser:
			return m.updateCreateUser(msg)
		case viewUpdateUser:
			return m.updateUpdateUser(msg)
		case viewConfirmUser:
			return m.updateConfirmUser(msg)
		case viewDeleteUser:
			return m.updateDeleteUser(msg)
		case viewOrderDetail:
			return m.updateOrderDetail(msg)
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

func (m AdminModel) updateLogin(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		if m.step == 1 {
			m.step = 0
			m.inputBuffer = m.authEmail
			m.status = ""
			return m, nil
		}
		return m, tea.Quit
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
		return m, nil
	case tea.KeyEnter:
		val := strings.TrimSpace(m.inputBuffer)
		m.inputBuffer = ""

		if m.step == 0 {
			if val == "" {
				m.status = "❌ L'adresse email est requise"
				return m, nil
			}
			m.authEmail = val
			m.step = 1
			m.status = ""
			return m, nil
		}

		if val == "" {
			m.status = "❌ Le mot de passe est requis"
			return m, nil
		}
		m.authPassword = val
		return m.doLogin()
	default:
		s := msg.String()
		if m.step == 0 && (s == "q" || s == "ctrl+c") && m.inputBuffer == "" {
			return m, tea.Quit
		}
		if strings.HasSuffix(s, "@") {
			m.inputBuffer += "@"
		} else if strings.HasPrefix(s, "alt+") && len(s) > 4 {
			m.inputBuffer += s[4:]
		} else if strings.HasPrefix(s, "ctrl+alt+") && len(s) > 9 {
			m.inputBuffer += s[9:]
		} else if len(msg.Runes) > 0 {
			m.inputBuffer += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			m.inputBuffer += " "
		} else if len(s) == 1 {
			m.inputBuffer += s
		}
	}
	return m, nil
}

func (m AdminModel) doLogin() (tea.Model, tea.Cmd) {
	payload := models.LoginRequest{
		Email:    m.authEmail,
		Password: m.authPassword,
	}
	bodyBytes, _ := json.Marshal(payload)

	resp, err := http.Post(m.serverURL+"/api/auth/login", "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.status = "❌ Erreur de connexion au serveur: " + err.Error()
		m.step = 0
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		errMsg := errResp["error"]
		if errMsg == "" {
			errMsg = fmt.Sprintf("Code HTTP %d", resp.StatusCode)
		}
		m.status = "❌ Échec connexion : " + errMsg
		m.step = 1
		return m, nil
	}

	var authResp models.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		m.status = "❌ Erreur décodage de la réponse"
		m.step = 0
		return m, nil
	}

	if authResp.User != nil && authResp.User.Role != models.RoleAdmin {
		m.status = fmt.Sprintf("❌ Accès refusé : compte non administrateur (rôle : %s)", authResp.User.Role)
		m.token = ""
		m.user = nil
		m.authPassword = ""
		m.step = 0
		return m, nil
	}

	m.token = authResp.Token
	m.user = authResp.User
	m.authPassword = ""
	m.status = fmt.Sprintf("Connecté en tant que %s", authResp.User.Email)
	m.state = viewMainMenu
	m.cursor = 0
	m.step = 0
	return m, nil
}

func (m AdminModel) updateMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuLen := 5
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
			m.token = ""
			m.user = nil
			m.authEmail = ""
			m.authPassword = ""
			m.inputBuffer = ""
			m.step = 0
			m.status = "Déconnecté avec succès."
			m.state = viewLogin
			m.cursor = 0
		case 4:
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m AdminModel) updateMenuProducts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuLen := 6
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
			m.state = viewUpdateProduct
			m.formProdID = ""
			m.formName = ""
			m.formCategory = ""
			m.formPriceHT = ""
			m.formStock = ""
		case 3:
			m.state = viewDeleteProduct
		case 4:
			m.state = viewRenameCategory
			m.formOldCat = ""
			m.formNewCat = ""
		case 5:
			m.state = viewMainMenu
		}
	}
	return m, nil
}

func (m AdminModel) updateMenuUsers(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuLen := 6
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
			m.state = viewUpdateUser
			m.formUserID = ""
			m.formEmail = ""
			m.formRole = ""
		case 3:
			m.state = viewConfirmUser
		case 4:
			m.state = viewDeleteUser
		case 5:
			m.state = viewMainMenu
			m.cursor = 1
		}
	}
	return m, nil
}

func (m AdminModel) updateMenuOrders(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuLen := 5
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
			m.state = viewOrderDetail
			m.formOrderID = ""
		case 2:
			m.state = viewUpdateOrderStatus
			m.formOrderID = ""
			m.formStatus = ""
			m.formReason = ""
		case 3:
			m.state = viewCreateManualOrder
			m.formUserID = ""
			m.formProdID = ""
			m.formQuantity = ""
		case 4:
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
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
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
	resp, err := m.doRequest(http.MethodPost, m.serverURL+"/api/admin/products", bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.status = "❌ Erreur de création du produit: " + err.Error()
		m.state = viewMenuProducts
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		m.resultsTitle = "❌ ÉCHEC DE LA CRÉATION DU PRODUIT"
		m.status = "Erreur"
		m.items = []string{
			fmt.Sprintf("Nom       : %s", m.formName),
			fmt.Sprintf("Catégorie : %s", m.formCategory),
			"",
			"Raison de l'échec :",
			"  " + errResp.Error,
		}
		m.returnState = viewMenuProducts
		m.state = viewListResults
		return m, nil
	}

	var created struct {
		ID       string  `json:"id"`
		Name     string  `json:"name"`
		PriceHT  float64 `json:"price_ht"`
		PriceTTC float64 `json:"price_ttc"`
		Stock    int     `json:"stock"`
	}
	json.NewDecoder(resp.Body).Decode(&created)

	m.resultsTitle = "✅ PRODUIT CRÉÉ AVEC SUCCÈS"
	m.status = created.ID
	m.items = []string{
		fmt.Sprintf("Identifiant : %s", created.ID),
		fmt.Sprintf("Nom         : %s", m.formName),
		fmt.Sprintf("Catégorie   : %s", m.formCategory),
		fmt.Sprintf("Prix HT     : %s € (Prix TTC: %.2f €)", m.formPriceHT, created.PriceTTC),
		fmt.Sprintf("Stock       : %s", m.formStock),
	}
	m.returnState = viewMenuProducts
	m.state = viewListResults
	return m, nil
}

func (m AdminModel) updateUpdateProduct(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuProducts
		m.inputBuffer = ""
		m.step = 0
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
	case tea.KeyEnter:
		val := strings.TrimSpace(m.inputBuffer)
		m.inputBuffer = ""

		switch m.step {
		case 0:
			if val == "" {
				m.status = "❌ L'ID du produit ne peut pas être vide"
				return m, nil
			}
			resp, err := m.doRequest(http.MethodGet, m.serverURL+"/api/admin/products/"+val, nil)
			if err != nil {
				m.status = "❌ Erreur réseau: " + err.Error()
				return m, nil
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusNotFound {
				m.status = "❌ Produit introuvable: " + val
				return m, nil
			}
			var p struct {
				ID       string  `json:"id"`
				Name     string  `json:"name"`
				Category string  `json:"category"`
				PriceHT  float64 `json:"price_ht"`
				Stock    int     `json:"stock"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
				m.status = "❌ Erreur décodage produit"
				return m, nil
			}
			m.formProdID = p.ID
			m.formName = p.Name
			m.formCategory = p.Category
			m.formPriceHT = fmt.Sprintf("%.2f", p.PriceHT)
			m.formStock = fmt.Sprintf("%d", p.Stock)
			m.status = fmt.Sprintf("Produit [%s] chargé. [Entrée] pour conserver chaque valeur.", p.ID)
			m.step = 1
		case 1:
			if val != "" {
				m.formName = val
			}
			m.step = 2
		case 2:
			if val != "" {
				m.formCategory = val
			}
			m.step = 3
		case 3:
			if val != "" {
				p, err := strconv.ParseFloat(val, 64)
				if err != nil || p <= 0 {
					m.status = "❌ Le prix HT doit être un nombre positif (ex: 49.99)"
					return m, nil
				}
				m.formPriceHT = val
			}
			m.step = 4
		case 4:
			if val != "" {
				st, err := strconv.Atoi(val)
				if err != nil || st < 0 {
					m.status = "❌ Le stock doit être un entier positif (ex: 10)"
					return m, nil
				}
				m.formStock = val
			}
			return m.submitUpdateProduct()
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

func (m AdminModel) submitUpdateProduct() (tea.Model, tea.Cmd) {
	priceHT, _ := strconv.ParseFloat(m.formPriceHT, 64)
	stock, _ := strconv.Atoi(m.formStock)

	payload := map[string]interface{}{
		"name":        m.formName,
		"category":    m.formCategory,
		"price_ht":    priceHT,
		"price_ttc":   priceHT * 1.20,
		"stock":       stock,
		"description": "Mis à jour via CLI Admin",
	}

	bodyBytes, _ := json.Marshal(payload)
	resp, err := m.doRequest(http.MethodPut, m.serverURL+"/api/admin/products/"+m.formProdID, bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.status = "❌ Erreur de mise à jour du produit: " + err.Error()
		m.state = viewMenuProducts
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		m.resultsTitle = "❌ ÉCHEC DE LA MODIFICATION"
		m.status = "Erreur"
		m.items = []string{
			fmt.Sprintf("Produit ID : %s", m.formProdID),
			"",
			"Raison de l'échec :",
			"  " + errResp.Error,
		}
		m.returnState = viewMenuProducts
		m.state = viewListResults
		return m, nil
	} else {
		m.resultsTitle = "✅ PRODUIT MIS À JOUR AVEC SUCCÈS"
		m.status = "Enregistré"
		m.items = []string{
			fmt.Sprintf("Produit ID : %s", m.formProdID),
			fmt.Sprintf("Nom        : %s", m.formName),
			fmt.Sprintf("Catégorie  : %s", m.formCategory),
			fmt.Sprintf("Prix HT    : %s € (Prix TTC: %.2f €)", m.formPriceHT, priceHT*1.20),
			fmt.Sprintf("Stock      : %d", stock),
		}
		m.returnState = viewMenuProducts
		m.state = viewListResults
		return m, nil
	}
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
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
	case tea.KeyEnter:
		prodID := strings.TrimSpace(m.inputBuffer)
		m.inputBuffer = ""
		if prodID == "" {
			m.status = "❌ ID produit requis"
			return m, nil
		}

		resp, err := m.doRequest(http.MethodDelete, fmt.Sprintf("%s/api/admin/products/%s", m.serverURL, prodID), nil)
		if err != nil {
			m.status = "❌ Erreur réseau: " + err.Error()
			m.state = viewMenuProducts
			return m, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			m.resultsTitle = "❌ ÉCHEC DE LA SUPPRESSION"
			m.status = "Erreur"
			m.items = []string{
				fmt.Sprintf("Produit ID : %s", prodID),
				"",
				"Impossible de supprimer le produit (identifiant inexistant ou erreur serveur).",
			}
			m.returnState = viewMenuProducts
			m.state = viewListResults
			return m, nil
		} else {
			m.resultsTitle = "✅ PRODUIT SUPPRIMÉ AVEC SUCCÈS"
			m.status = prodID
			m.items = []string{
				fmt.Sprintf("Le produit %s a été supprimé définitivement du catalogue.", prodID),
			}
			m.returnState = viewMenuProducts
			m.state = viewListResults
			return m, nil
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

func (m AdminModel) updateRenameCategory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuProducts
		m.inputBuffer = ""
		m.step = 0
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
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

	resp, err := m.doRequest(http.MethodPut, m.serverURL+"/api/admin/categories/rename", bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.status = "❌ Erreur réseau: " + err.Error()
		m.state = viewMenuProducts
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.resultsTitle = "❌ ÉCHEC DU RENOMMAGE"
		m.status = "Erreur"
		m.items = []string{
			fmt.Sprintf("Ancienne catégorie : %s", m.formOldCat),
			fmt.Sprintf("Nouvelle catégorie : %s", m.formNewCat),
			"",
			"Erreur lors du renommage en base de données.",
		}
		m.returnState = viewMenuProducts
		m.state = viewListResults
		return m, nil
	} else {
		var res struct {
			RowsAffected int64 `json:"rows_affected"`
		}
		json.NewDecoder(resp.Body).Decode(&res)
		m.resultsTitle = "✅ CATÉGORIE RENOMMÉE AVEC SUCCÈS"
		m.status = fmt.Sprintf("%d produit(s) modifié(s)", res.RowsAffected)
		m.items = []string{
			fmt.Sprintf("Ancienne catégorie : %s", m.formOldCat),
			fmt.Sprintf("Nouvelle catégorie : %s", m.formNewCat),
			fmt.Sprintf("Nombre de produits mis à jour : %d", res.RowsAffected),
		}
		m.returnState = viewMenuProducts
		m.state = viewListResults
		return m, nil
	}
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
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
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
	resp, err := m.doRequest(http.MethodPost, m.serverURL+"/api/admin/users", bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.status = "❌ Erreur réseau: " + err.Error()
		m.state = viewMenuUsers
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		m.resultsTitle = "❌ ÉCHEC DE LA CRÉATION UTILISATEUR"
		m.status = "Erreur"
		m.items = []string{
			fmt.Sprintf("Email : %s", m.formEmail),
			"",
			"Raison de l'échec :",
			"  " + errResp.Error,
		}
		m.returnState = viewMenuUsers
		m.state = viewListResults
		return m, nil
	} else {
		var u struct {
			ID int `json:"id"`
		}
		json.NewDecoder(resp.Body).Decode(&u)
		m.resultsTitle = "✅ UTILISATEUR CRÉÉ AVEC SUCCÈS"
		m.status = fmt.Sprintf("ID #%d", u.ID)
		m.items = []string{
			fmt.Sprintf("Identifiant : #%d", u.ID),
			fmt.Sprintf("Email       : %s", m.formEmail),
			fmt.Sprintf("Rôle        : %s", m.formRole),
			"Statut      : Confirmé par défaut",
		}
		m.returnState = viewMenuUsers
		m.state = viewListResults
		return m, nil
	}
}

func (m AdminModel) updateUpdateUser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuUsers
		m.inputBuffer = ""
		m.step = 0
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
	case tea.KeyEnter:
		val := strings.TrimSpace(m.inputBuffer)
		m.inputBuffer = ""

		switch m.step {
		case 0:
			if val == "" {
				m.status = "❌ L'ID de l'utilisateur ne peut pas être vide"
				return m, nil
			}
			id, err := strconv.Atoi(val)
			if err != nil || id <= 0 {
				m.status = "❌ ID utilisateur invalide"
				return m, nil
			}
			resp, err := m.doRequest(http.MethodGet, fmt.Sprintf("%s/api/admin/users/%d", m.serverURL, id), nil)
			if err != nil {
				m.status = "❌ Erreur réseau: " + err.Error()
				return m, nil
			}
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusNotFound {
				m.status = fmt.Sprintf("❌ Utilisateur #%d introuvable", id)
				return m, nil
			}
			var u struct {
				ID    int    `json:"id"`
				Email string `json:"email"`
				Role  string `json:"role"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
				m.status = "❌ Erreur décodage utilisateur"
				return m, nil
			}
			m.formUserID = val
			m.formEmail = u.Email
			m.formRole = u.Role
			m.status = fmt.Sprintf("Utilisateur #%d chargé. [Entrée] pour conserver chaque valeur.", id)
			m.step = 1
		case 1:
			if val != "" {
				if !strings.Contains(val, "@") {
					m.status = "❌ Email valide requis"
					return m, nil
				}
				m.formEmail = val
			}
			m.step = 2
		case 2:
			if val != "" {
				role := "client"
				if val == "admin" {
					role = "admin"
				}
				m.formRole = role
			}
			return m.submitUpdateUser()
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

func (m AdminModel) submitUpdateUser() (tea.Model, tea.Cmd) {
	payload := map[string]interface{}{
		"email":        m.formEmail,
		"role":         m.formRole,
		"is_confirmed": true,
	}
	bodyBytes, _ := json.Marshal(payload)
	resp, err := m.doRequest(http.MethodPut, m.serverURL+"/api/admin/users/"+m.formUserID, bytes.NewBuffer(bodyBytes))
	if err != nil {
		m.status = "❌ Erreur réseau: " + err.Error()
		m.state = viewMenuUsers
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		m.resultsTitle = "❌ ÉCHEC DE LA MODIFICATION UTILISATEUR"
		m.status = "Erreur"
		m.items = []string{
			fmt.Sprintf("Utilisateur ID : %s", m.formUserID),
			"",
			"Raison de l'échec :",
			"  " + errResp.Error,
		}
		m.returnState = viewMenuUsers
		m.state = viewListResults
		return m, nil
	} else {
		m.resultsTitle = "✅ UTILISATEUR MIS À JOUR AVEC SUCCÈS"
		m.status = fmt.Sprintf("ID #%s", m.formUserID)
		m.items = []string{
			fmt.Sprintf("Identifiant : #%s", m.formUserID),
			fmt.Sprintf("Nouvel email: %s", m.formEmail),
			fmt.Sprintf("Nouveau rôle: %s", m.formRole),
		}
		m.returnState = viewMenuUsers
		m.state = viewListResults
		return m, nil
	}
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
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
	case tea.KeyEnter:
		idStr := strings.TrimSpace(m.inputBuffer)
		m.inputBuffer = ""
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			m.status = "❌ Veuillez saisir un ID utilisateur valide"
			return m, nil
		}

		resp, err := m.doRequest(http.MethodPost, fmt.Sprintf("%s/api/admin/users/%d/confirm", m.serverURL, id), nil)
		if err != nil {
			m.status = "❌ Erreur réseau: " + err.Error()
			m.state = viewMenuUsers
			return m, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			m.resultsTitle = "❌ ÉCHEC DE LA CONFIRMATION"
			m.status = "Erreur"
			m.items = []string{
				fmt.Sprintf("Utilisateur ID : #%d", id),
				"",
				"Impossible de confirmer le compte (ID introuvable ou erreur serveur).",
			}
			m.returnState = viewMenuUsers
			m.state = viewListResults
			return m, nil
		} else {
			m.resultsTitle = "✅ UTILISATEUR CONFIRMÉ AVEC SUCCÈS"
			m.status = fmt.Sprintf("ID #%d", id)
			m.items = []string{
				fmt.Sprintf("Le compte utilisateur #%d a été validé manuellement.", id),
				"L'utilisateur peut maintenant se connecter à l'application.",
			}
			m.returnState = viewMenuUsers
			m.state = viewListResults
			return m, nil
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

func (m AdminModel) updateDeleteUser(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuUsers
		m.inputBuffer = ""
		m.step = 0
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
	case tea.KeyEnter:
		idStr := strings.TrimSpace(m.inputBuffer)
		m.inputBuffer = ""
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			m.status = "❌ Veuillez saisir un ID utilisateur valide"
			return m, nil
		}

		resp, err := m.doRequest(http.MethodDelete, fmt.Sprintf("%s/api/admin/users/%d", m.serverURL, id), nil)
		if err != nil {
			m.status = "❌ Erreur réseau: " + err.Error()
			m.state = viewMenuUsers
			return m, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			m.resultsTitle = "❌ ÉCHEC DE LA SUPPRESSION"
			m.status = "Erreur"
			m.items = []string{
				fmt.Sprintf("Utilisateur ID : #%d", id),
				"",
				"Impossible de supprimer l'utilisateur (ID introuvable ou erreur serveur).",
			}
			m.returnState = viewMenuUsers
			m.state = viewListResults
			return m, nil
		} else {
			m.resultsTitle = "✅ UTILISATEUR SUPPRIMÉ AVEC SUCCÈS"
			m.status = fmt.Sprintf("ID #%d", id)
			m.items = []string{
				fmt.Sprintf("Le compte utilisateur #%d a été supprimé définitivement.", id),
			}
			m.returnState = viewMenuUsers
			m.state = viewListResults
			return m, nil
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

func (m AdminModel) updateOrderDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = viewMenuOrders
		m.inputBuffer = ""
		m.step = 0
		return m, nil
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
	case tea.KeyEnter:
		orderID := strings.TrimSpace(m.inputBuffer)
		m.inputBuffer = ""
		if orderID == "" {
			m.status = "❌ ID de commande requis"
			return m, nil
		}
		return m.fetchOrderDetail(orderID)
	default:
		if msg.Type == tea.KeyRunes {
			m.inputBuffer += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			m.inputBuffer += " "
		}
	}
	return m, nil
}

func (m AdminModel) fetchOrderDetail(orderID string) (tea.Model, tea.Cmd) {
	resp, err := m.doRequest(http.MethodGet, m.serverURL+"/api/admin/orders/"+orderID, nil)
	if err != nil {
		m.status = "❌ Erreur réseau: " + err.Error()
		m.state = viewMenuOrders
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		m.status = "❌ Commande introuvable: " + orderID
		m.state = viewMenuOrders
		return m, nil
	}
	if resp.StatusCode != http.StatusOK {
		m.status = fmt.Sprintf("❌ Erreur serveur (Code %d)", resp.StatusCode)
		m.state = viewMenuOrders
		return m, nil
	}

	var order struct {
		ID           string  `json:"id"`
		UserID       int     `json:"user_id"`
		UserEmail    string  `json:"user_email"`
		Status       string  `json:"status"`
		CancelReason string  `json:"cancel_reason"`
		TotalTTC     float64 `json:"total_ttc"`
		Items        []struct {
			ProductID    string  `json:"product_id"`
			ProductName  string  `json:"product_name"`
			Quantity     int     `json:"quantity"`
			PriceUnitTTC float64 `json:"price_unit_ttc"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		m.status = "❌ Erreur décodage commande"
		m.state = viewMenuOrders
		return m, nil
	}

	m.resultsTitle = fmt.Sprintf("DÉTAIL COMMANDE %s", order.ID)
	m.items = nil
	m.items = append(m.items, fmt.Sprintf("Client : %s (ID: %d)", order.UserEmail, order.UserID))
	m.items = append(m.items, fmt.Sprintf("Statut : %s", order.Status))
	if order.CancelReason != "" {
		m.items = append(m.items, fmt.Sprintf("Motif d'annulation : %s", order.CancelReason))
	}
	m.items = append(m.items, "--- Articles commandés ---")
	if len(order.Items) == 0 {
		m.items = append(m.items, "  (Aucun article enregistré)")
	} else {
		for _, it := range order.Items {
			m.items = append(m.items, fmt.Sprintf("  • %s (ID: %s) x%d à %.2f€ TTC = %.2f€ TTC", it.ProductName, it.ProductID, it.Quantity, it.PriceUnitTTC, float64(it.Quantity)*it.PriceUnitTTC))
		}
	}
	m.items = append(m.items, fmt.Sprintf("TOTAL : %.2f€ TTC", order.TotalTTC))
	m.returnState = viewMenuOrders
	m.state = viewListResults
	m.status = "✅ Détails de la commande chargés"
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
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
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
			if strings.HasPrefix(strings.ToLower(val), "annul") {
				m.formStatus = "annulé"
				m.step = 2
			} else {
				m.formStatus = val
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

	resp, err := m.doRequest(http.MethodPut, fmt.Sprintf("%s/api/admin/orders/%s/status", m.serverURL, m.formOrderID), bytes.NewBuffer(bodyBytes))
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
		m.resultsTitle = "❌ ÉCHEC DU CHANGEMENT DE STATUT"
		m.status = "Refusé par le serveur"
		m.items = []string{
			fmt.Sprintf("Commande ID : %s", m.formOrderID),
			"",
			"Raison du refus :",
			"  " + errResp.Error,
		}
		m.returnState = viewMenuOrders
		m.state = viewListResults
		return m, nil
	} else {
		m.resultsTitle = "✅ STATUT MIS À JOUR AVEC SUCCÈS"
		m.status = m.formStatus
		m.items = []string{
			fmt.Sprintf("Commande ID : %s", m.formOrderID),
			fmt.Sprintf("Nouveau statut : %s", m.formStatus),
		}
		if m.formReason != "" {
			m.items = append(m.items, fmt.Sprintf("Motif : %s", m.formReason))
		}
		if m.formStatus == "annulé" {
			m.items = append(m.items, "📦 Stock automatiquement réintégré dans le catalogue.")
		}
		m.returnState = viewMenuOrders
		m.state = viewListResults
		return m, nil
	}
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
	case tea.KeyBackspace, tea.KeyDelete:
		m.inputBuffer = deleteLastRune(m.inputBuffer)
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
	resp, err := m.doRequest(http.MethodPost, m.serverURL+"/api/admin/orders/manual", bytes.NewBuffer(bodyBytes))
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
		m.resultsTitle = "❌ ÉCHEC DE LA COMMANDE"
		m.status = "Refusée par le serveur"
		m.items = []string{
			fmt.Sprintf("Client ID   : %d", uid),
			fmt.Sprintf("Produit ID  : %s", m.formProdID),
			fmt.Sprintf("Quantité    : %d", qty),
			"",
			"Raison du refus :",
			"  " + errResp.Error,
		}
		m.returnState = viewMenuOrders
		m.state = viewListResults
		return m, nil
	} else {
		var order struct {
			ID       string  `json:"id"`
			TotalTTC float64 `json:"total_ttc"`
		}
		json.NewDecoder(resp.Body).Decode(&order)
		m.resultsTitle = "✅ COMMANDE CRÉÉE AVEC SUCCÈS"
		m.status = fmt.Sprintf("ID: %s", order.ID)
		m.items = []string{
			fmt.Sprintf("Commande ID : %s", order.ID),
			fmt.Sprintf("Client ID   : %d", uid),
			fmt.Sprintf("Produit ID  : %s", m.formProdID),
			fmt.Sprintf("Quantité    : %d", qty),
			fmt.Sprintf("Total payé  : %.2f € TTC", order.TotalTTC),
			"",
			"Le stock a été décrémenté automatiquement en BDD.",
		}
		m.returnState = viewMenuOrders
		m.state = viewListResults
		return m, nil
	}
}

func (m *AdminModel) fetchProducts() {
	m.status = "Chargement des produits..."
	m.items = nil

	resp, err := m.doRequest(http.MethodGet, m.serverURL+"/api/admin/products", nil)
	if err != nil {
		m.status = "❌ Erreur connexion serveur: " + err.Error()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if msg, ok := errResp["error"]; ok {
			m.status = "❌ " + msg
		} else {
			m.status = fmt.Sprintf("❌ Erreur serveur (Code %d)", resp.StatusCode)
		}
		return
	}

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

	resp, err := m.doRequest(http.MethodGet, m.serverURL+"/api/admin/users", nil)
	if err != nil {
		m.status = "❌ Erreur connexion serveur: " + err.Error()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if msg, ok := errResp["error"]; ok {
			m.status = "❌ " + msg
		} else {
			m.status = fmt.Sprintf("❌ Erreur serveur (Code %d)", resp.StatusCode)
		}
		return
	}

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

	resp, err := m.doRequest(http.MethodGet, m.serverURL+"/api/admin/orders", nil)
	if err != nil {
		m.status = "❌ Erreur connexion serveur: " + err.Error()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if msg, ok := errResp["error"]; ok {
			m.status = "❌ " + msg
		} else {
			m.status = fmt.Sprintf("❌ Erreur serveur (Code %d)", resp.StatusCode)
		}
		return
	}

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

	if m.user != nil {
		s.WriteString(fmt.Sprintf("👤 Connecté : %s (Rôle : %s)\n", m.user.Email, m.user.Role))
	}

	if m.status != "" && m.state != viewListResults {
		s.WriteString("Statut : " + m.status + "\n\n")
	}

	if m.state == viewLogin {
		s.WriteString("🔐 ESPACE ADMINISTRATION - CONNEXION REQUISE :\n\n")
		switch m.step {
		case 0:
			s.WriteString("Étape 1/2 - Entrez votre EMAIL administrateur :\n")
			s.WriteString(fmt.Sprintf("\n> %s █\n\n", m.inputBuffer))
		case 1:
			s.WriteString(fmt.Sprintf("✓ Email : %s\n", m.authEmail))
			s.WriteString("Étape 2/2 - Entrez votre MOT DE PASSE :\n")
			masked := strings.Repeat("*", len(m.inputBuffer))
			s.WriteString(fmt.Sprintf("\n> %s █\n\n", masked))
		}
		s.WriteString("Appuyez sur [Entrée] pour valider, ou [Échap] / [q] pour quitter.\n")
		return s.String()
	}

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

	if m.state == viewUpdateProduct {
		s.WriteString("MODIFICATION D'UN PRODUIT :\n\n")
		switch m.step {
		case 0:
			s.WriteString("Étape 1/5 - Entrez l'ID du produit à modifier (ex: PDT-SIL1BB) :\n")
		case 1:
			s.WriteString(fmt.Sprintf("✓ Produit ID : %s\n", m.formProdID))
			s.WriteString(fmt.Sprintf("Étape 2/5 - Nom du produit [Actuel : %s] (Entrée pour conserver) :\n", m.formName))
		case 2:
			s.WriteString(fmt.Sprintf("✓ Produit ID : %s\n", m.formProdID))
			s.WriteString(fmt.Sprintf("✓ Nom : %s\n", m.formName))
			s.WriteString(fmt.Sprintf("Étape 3/5 - Catégorie [Actuelle : %s] (Entrée pour conserver) :\n", m.formCategory))
		case 3:
			s.WriteString(fmt.Sprintf("✓ Produit ID : %s\n", m.formProdID))
			s.WriteString(fmt.Sprintf("✓ Nom : %s\n", m.formName))
			s.WriteString(fmt.Sprintf("✓ Catégorie : %s\n", m.formCategory))
			s.WriteString(fmt.Sprintf("Étape 4/5 - Prix HT en € [Actuel : %s €] (Entrée pour conserver) :\n", m.formPriceHT))
		case 4:
			s.WriteString(fmt.Sprintf("✓ Produit ID : %s\n", m.formProdID))
			s.WriteString(fmt.Sprintf("✓ Nom : %s\n", m.formName))
			s.WriteString(fmt.Sprintf("✓ Catégorie : %s\n", m.formCategory))
			s.WriteString(fmt.Sprintf("✓ Prix HT : %s €\n", m.formPriceHT))
			s.WriteString(fmt.Sprintf("Étape 5/5 - Quantité en stock [Actuel : %s] (Entrée pour conserver) :\n", m.formStock))
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

	if m.state == viewUpdateUser {
		s.WriteString("MODIFICATION D'UN UTILISATEUR :\n\n")
		switch m.step {
		case 0:
			s.WriteString("Étape 1/3 - Entrez l'ID numérique de l'utilisateur (ex: 2) :\n")
		case 1:
			s.WriteString(fmt.Sprintf("✓ Utilisateur ID : %s\n", m.formUserID))
			s.WriteString(fmt.Sprintf("Étape 2/3 - Email de l'utilisateur [Actuel : %s] (Entrée pour conserver) :\n", m.formEmail))
		case 2:
			s.WriteString(fmt.Sprintf("✓ Utilisateur ID : %s\n", m.formUserID))
			s.WriteString(fmt.Sprintf("✓ Email : %s\n", m.formEmail))
			s.WriteString(fmt.Sprintf("Étape 3/3 - Rôle ('client' ou 'admin') [Actuel : %s] (Entrée pour conserver) :\n", m.formRole))
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

	if m.state == viewOrderDetail {
		s.WriteString("CONSULTATION D'UNE COMMANDE :\n\n")
		s.WriteString("Entrez l'ID de la commande à consulter (ex: CMD-1F2S8B) :\n")
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
			"4. Déconnexion",
			"5. Quitter",
		}
	case viewMenuProducts:
		menuTitle = "GESTION DES PRODUITS"
		choices = []string{
			"1. Lister les Produits",
			"2. Ajouter un nouveau Produit",
			"3. Modifier un Produit",
			"4. Supprimer un Produit",
			"5. Renommer une Catégorie sur tous les produits",
			"6. Retour au Menu Principal",
		}
	case viewMenuUsers:
		menuTitle = "GESTION DES UTILISATEURS"
		choices = []string{
			"1. Lister les Utilisateurs",
			"2. Créer un Utilisateur",
			"3. Modifier un Utilisateur (Email / Rôle)",
			"4. Confirmer manuellement un compte",
			"5. Supprimer un Utilisateur",
			"6. Retour au Menu Principal",
		}
	case viewMenuOrders:
		menuTitle = "GESTION DES COMMANDES"
		choices = []string{
			"1. Lister les Commandes",
			"2. Consulter le détail d'une Commande (Articles)",
			"3. Modifier le statut d'une Commande (Statut / Motif)",
			"4. Créer une Commande (Client + Produits)",
			"5. Retour au Menu Principal",
		}
	}

	s.WriteString(menuTitle + " :\n\n")
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

func deleteLastRune(s string) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return ""
	}
	return string(runes[:len(runes)-1])
}
