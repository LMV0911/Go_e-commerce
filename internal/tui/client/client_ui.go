package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/LMV0911/Go_e-commerce/internal/models"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateMenu state = iota
	stateLogin
	stateRegister
	stateConfirm
	stateReset
	stateProfile

	stateCatalog
	stateCart
	stateCheckout
	stateOrders
)

var (
	colorBrand     = lipgloss.Color("#6366F1") 
	colorCyan      = lipgloss.Color("#06B6D4") 
	colorGreen     = lipgloss.Color("#10B981") 
	colorOrange    = lipgloss.Color("#F59E0B") 
	colorRed       = lipgloss.Color("#EF4444") 
	colorMuted     = lipgloss.Color("#6B7280") 
	colorTextLight = lipgloss.Color("#F3F4F6")

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorBrand).
			Padding(0, 2)

	badgeProduct = lipgloss.NewStyle().
			Foreground(colorCyan).
			Bold(true)

	badgeBasket = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EC4899")).
			Bold(true)

	badgeOrder = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A855F7")).
			Bold(true)

	statusPaid = lipgloss.NewStyle().
			Background(colorGreen).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	statusShipping = lipgloss.NewStyle().
			Background(colorOrange).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1).
			Bold(true)

	statusDelivered = lipgloss.NewStyle().
			Background(lipgloss.Color("#3B82F6")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	statusCancelled = lipgloss.NewStyle().
			Background(colorRed).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 1).
			Bold(true)

	statusPending = lipgloss.NewStyle().
			Background(lipgloss.Color("#9CA3AF")).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBrand).
			Padding(1, 2)

	summaryBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorGreen).
			Padding(1, 2)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#374151"))

	normalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB"))
)

type ClientModel struct {
	serverURL string
	state     state
	cursor    int
	message   string
	token     string
	user      *models.User

	step     int
	input    string
	email    string
	password string
	code     string

	products            []models.Product
	categories          []string
	selectedCategoryIdx int
	selectedProductIdx  int
	searchQuery         string
	isSearching         bool
	priceFilterType     string 
	sortIdx             int    

	cart            *models.Cart
	selectedCartIdx int

	cardNumber   string
	expDate      string
	cvc          string
	cardholder   string
	focusedField int 
	lastOrder    *models.Order

	orders           []models.Order
	selectedOrderIdx int
}

func NewClientModel(serverURL string) ClientModel {
	m := ClientModel{
		serverURL:       serverURL,
		state:           stateMenu,
		message:         "Prêt",
		categories:      []string{"Toutes"},
		priceFilterType: "ttc",
		cardNumber:      "4532 7512 8901 2345",
		expDate:         "06/28",
		cvc:             "456",
		cardholder:      "Acheteur Test",
	}
	m.loadCart()
	return m
}

func (m ClientModel) Init() tea.Cmd {
	return nil
}

func (m ClientModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if keyMsg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	if m.state == stateProfile {
		if keyMsg.String() == "esc" || keyMsg.String() == "enter" || keyMsg.String() == "q" {
			m.state = stateMenu
		}
		return m, nil
	}


	if m.state == stateLogin || m.state == stateRegister || m.state == stateConfirm || m.state == stateReset {
		return m.updateAuthForm(keyMsg)
	}
	switch m.state {
	case stateCatalog:
		return m.updateCatalog(keyMsg)
	case stateCart:
		return m.updateCart(keyMsg)
	case stateCheckout:
		return m.updateCheckout(keyMsg)
	case stateOrders:
		return m.updateOrders(keyMsg)
	}


	maxCursor := 8
	if m.user != nil {
		maxCursor = 6
	}

	switch keyMsg.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < maxCursor {
			m.cursor++
		}
	case "enter":
		m.step = 0
		m.input = ""

		if m.user == nil {
			switch m.cursor {
			case 0:
				m.loadProducts()
				m.loadCategories()
				m.loadCart()
				m.message = ""
				m.state = stateCatalog
			case 1:
				m.loadCart()
				m.message = ""
				m.state = stateCart
			case 2:
				m.loadCart()
				m.lastOrder = nil
				m.message = ""
				m.state = stateCheckout
			case 3:
				m.loadOrders()
				m.message = ""
				m.state = stateOrders
			case 4:
				m.message = ""
				m.state = stateLogin
			case 5:
				m.message = ""
				m.state = stateRegister
			case 6:
				m.message = ""
				m.state = stateConfirm
			case 7:
				m.message = ""
				m.state = stateReset
			case 8:
				return m, tea.Quit
			}
		} else {
			switch m.cursor {
			case 0:
				m.loadProducts()
				m.loadCategories()
				m.loadCart()
				m.message = ""
				m.state = stateCatalog
			case 1:
				m.loadCart()
				m.message = ""
				m.state = stateCart
			case 2:
				m.loadCart()
				m.lastOrder = nil
				m.message = ""
				m.state = stateCheckout
			case 3:
				m.loadOrders()
				m.message = ""
				m.state = stateOrders
			case 4:
				m.loadProfile()
				m.state = stateProfile
			case 5: // Déconnexion
				m.user = nil
				m.token = ""
				m.message = "Déconnecté"
				m.cursor = 0
			case 6:
				return m, tea.Quit
			}
		}
	}

	return m, nil
}



func (m ClientModel) updateCatalog(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.isSearching {
		switch msg.Type {
		case tea.KeyEnter, tea.KeyEsc:
			m.isSearching = false
			m.loadProducts()
			return m, nil
		case tea.KeyBackspace:
			if len(m.searchQuery) > 0 {
				m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				m.loadProducts()
			}
			return m, nil
		default:
			if len(msg.String()) == 1 {
				m.searchQuery += msg.String()
				m.loadProducts()
			}
			return m, nil
		}
	}

	switch msg.String() {
	case "esc":
		m.state = stateMenu
		return m, nil
	case "/":
		m.isSearching = true
		return m, nil
	case "c":
		if len(m.categories) > 0 {
			m.selectedCategoryIdx = (m.selectedCategoryIdx + 1) % len(m.categories)
			m.loadProducts()
		}
	case "t":
		if m.priceFilterType == "ttc" {
			m.priceFilterType = "ht"
		} else {
			m.priceFilterType = "ttc"
		}
		m.loadProducts()
	case "s":
		m.sortIdx = (m.sortIdx + 1) % 4
		m.loadProducts()
	case "up", "k":
		if m.selectedProductIdx > 0 {
			m.selectedProductIdx--
		}
	case "down", "j":
		if m.selectedProductIdx < len(m.products)-1 {
			m.selectedProductIdx++
		}
	case "enter", "+", "a", "A":
		if len(m.products) > 0 && m.selectedProductIdx < len(m.products) {
			p := m.products[m.selectedProductIdx]
			m.addToCart(p.ID, 1)
		}
	case "x", "X", "v", "V":
		m.clearCart()
	case "2":
		m.loadCart()
		m.message = ""
		m.state = stateCart
	case "3":
		m.loadCart()
		m.lastOrder = nil
		m.message = ""
		m.state = stateCheckout
	case "4":
		m.loadOrders()
		m.message = ""
		m.state = stateOrders
	}

	return m, nil
}

func (m ClientModel) updateCart(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateMenu
		return m, nil
	case "up", "k":
		if m.selectedCartIdx > 0 {
			m.selectedCartIdx--
		}
	case "down", "j":
		if m.cart != nil && m.selectedCartIdx < len(m.cart.Items)-1 {
			m.selectedCartIdx++
		}
	case "+":
		if m.cart != nil && m.selectedCartIdx < len(m.cart.Items) {
			it := m.cart.Items[m.selectedCartIdx]
			m.updateCartQty(it.ProductID, it.Quantity+1)
		}
	case "-":
		if m.cart != nil && m.selectedCartIdx < len(m.cart.Items) {
			it := m.cart.Items[m.selectedCartIdx]
			m.updateCartQty(it.ProductID, it.Quantity-1)
		}
	case "d", "backspace", "delete":
		if m.cart != nil && m.selectedCartIdx < len(m.cart.Items) {
			it := m.cart.Items[m.selectedCartIdx]
			m.removeFromCart(it.ProductID)
		}
	case "x", "X", "v", "V":
		m.clearCart()
	case "p", "enter":
		m.loadCart()
		m.lastOrder = nil
		m.state = stateCheckout
	case "1":
		m.loadProducts()
		m.state = stateCatalog
	case "4":
		m.loadOrders()
		m.state = stateOrders
	}
	return m, nil
}

func (m ClientModel) updateCheckout(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.lastOrder != nil {
		switch msg.String() {
		case "esc", "enter":
			m.state = stateMenu
			return m, nil
		case "4":
			m.loadOrders()
			m.state = stateOrders
			return m, nil
		}
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		m.state = stateMenu
		return m, nil
	case tea.KeyTab, tea.KeyDown:
		m.focusedField = (m.focusedField + 1) % 5
		return m, nil
	case tea.KeyShiftTab, tea.KeyUp:
		m.focusedField = (m.focusedField + 4) % 5
		return m, nil
	case tea.KeyEnter:
		if m.focusedField == 4 {
			m.doCheckout()
			return m, nil
		}
		m.focusedField = (m.focusedField + 1) % 5
		return m, nil
	case tea.KeyBackspace:
		switch m.focusedField {
		case 0:
			if len(m.cardNumber) > 0 {
				m.cardNumber = m.cardNumber[:len(m.cardNumber)-1]
			}
		case 1:
			if len(m.expDate) > 0 {
				m.expDate = m.expDate[:len(m.expDate)-1]
			}
		case 2:
			if len(m.cvc) > 0 {
				m.cvc = m.cvc[:len(m.cvc)-1]
			}
		case 3:
			if len(m.cardholder) > 0 {
				m.cardholder = m.cardholder[:len(m.cardholder)-1]
			}
		}
		return m, nil
	default:
		s := msg.String()
		if len(s) == 1 {
			switch m.focusedField {
			case 0:
				if len(m.cardNumber) < 19 {
					m.cardNumber += s
				}
			case 1:
				if len(m.expDate) < 5 {
					m.expDate += s
				}
			case 2:
				if len(m.cvc) < 3 {
					m.cvc += s
				}
			case 3:
				m.cardholder += s
			}
		}
	}
	return m, nil
}

func (m ClientModel) updateOrders(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateMenu
		return m, nil
	case "up", "k":
		if m.selectedOrderIdx > 0 {
			m.selectedOrderIdx--
		}
	case "down", "j":
		if m.selectedOrderIdx < len(m.orders)-1 {
			m.selectedOrderIdx++
		}
	case "1":
		m.loadProducts()
		m.state = stateCatalog
	case "2":
		m.loadCart()
		m.state = stateCart
	}
	return m, nil
}


func (m *ClientModel) setAuthHeader(req *http.Request) {
	if m.token != "" {
		req.Header.Set("Authorization", "Bearer "+m.token)
	}
}

func (m *ClientModel) loadProducts() {
	params := url.Values{}
	if m.searchQuery != "" {
		params.Set("q", m.searchQuery)
	}
	if m.selectedCategoryIdx > 0 && m.selectedCategoryIdx < len(m.categories) {
		params.Set("category", m.categories[m.selectedCategoryIdx])
	}
	if m.priceFilterType != "" {
		params.Set("price_type", m.priceFilterType)
	}
	switch m.sortIdx {
	case 1:
		params.Set("sort", "name_asc")
	case 2:
		params.Set("sort", "price_asc")
	case 3:
		params.Set("sort", "price_desc")
	}

	req, _ := http.NewRequest("GET", m.serverURL+"/api/products?"+params.Encode(), nil)
	m.setAuthHeader(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.message = "Erreur catalogue: " + err.Error()
		return
	}
	defer resp.Body.Close()

	var prods []models.Product
	if err := json.NewDecoder(resp.Body).Decode(&prods); err == nil {
		m.products = prods
		if m.selectedProductIdx >= len(m.products) {
			m.selectedProductIdx = 0
		}
	}
}

func (m *ClientModel) loadCategories() {
	req, _ := http.NewRequest("GET", m.serverURL+"/api/categories", nil)
	m.setAuthHeader(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var cats []string
	if err := json.NewDecoder(resp.Body).Decode(&cats); err == nil {
		m.categories = append([]string{"Toutes"}, cats...)
	}
}

func (m *ClientModel) loadCart() {
	req, _ := http.NewRequest("GET", m.serverURL+"/api/cart", nil)
	m.setAuthHeader(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.message = "Erreur panier: " + err.Error()
		return
	}
	defer resp.Body.Close()

	var cart models.Cart
	if err := json.NewDecoder(resp.Body).Decode(&cart); err == nil {
		m.cart = &cart
		if m.selectedCartIdx >= len(m.cart.Items) {
			m.selectedCartIdx = 0
		}
	}
}

func (m *ClientModel) addToCart(productID string, qty int) {
	data, _ := json.Marshal(map[string]any{
		"product_id": productID,
		"quantity":   qty,
	})
	req, _ := http.NewRequest("POST", m.serverURL+"/api/cart/items", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	m.setAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.message = "❌ Erreur ajout panier: " + err.Error()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp["error"] != "" {
			m.message = "❌ " + errResp["error"]
		} else {
			m.message = fmt.Sprintf("❌ Erreur serveur (%d)", resp.StatusCode)
		}
		return
	}

	var cart models.Cart
	_ = json.NewDecoder(resp.Body).Decode(&cart)
	m.cart = &cart

	prodName := productID
	for _, it := range cart.Items {
		if it.ProductID == productID {
			prodName = it.Product.Name
			break
		}
	}
	m.message = fmt.Sprintf("✨ « %s » ajouté au panier ! (Total panier: %d article(s) - %.2f € TTC)", prodName, len(cart.Items), cart.TotalTTC)
}

func (m *ClientModel) updateCartQty(productID string, qty int) {
	if qty <= 0 {
		m.removeFromCart(productID)
		return
	}
	data, _ := json.Marshal(map[string]any{
		"product_id": productID,
		"quantity":   qty,
	})
	req, _ := http.NewRequest("PUT", m.serverURL+"/api/cart/items", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	m.setAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.message = "❌ Erreur mise à jour: " + err.Error()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		m.message = "❌ " + errResp["error"]
		return
	}

	var cart models.Cart
	_ = json.NewDecoder(resp.Body).Decode(&cart)
	m.cart = &cart
	m.message = fmt.Sprintf("Quantité mise à jour : %d", qty)
}

func (m *ClientModel) removeFromCart(productID string) {
	req, _ := http.NewRequest("DELETE", m.serverURL+"/api/cart/items?product_id="+productID, nil)
	m.setAuthHeader(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.message = "❌ Erreur suppression: " + err.Error()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		m.message = "❌ " + errResp["error"]
		return
	}

	var cart models.Cart
	_ = json.NewDecoder(resp.Body).Decode(&cart)
	m.cart = &cart
	if m.selectedCartIdx >= len(m.cart.Items) && m.selectedCartIdx > 0 {
		m.selectedCartIdx--
	}
	m.message = "🗑️ Article supprimé du panier"
}

func (m *ClientModel) clearCart() {
	req, _ := http.NewRequest("DELETE", m.serverURL+"/api/cart", nil)
	m.setAuthHeader(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.message = "❌ Erreur vidage panier: " + err.Error()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		m.message = "❌ " + errResp["error"]
		return
	}

	var cart models.Cart
	_ = json.NewDecoder(resp.Body).Decode(&cart)
	m.cart = &cart
	m.selectedCartIdx = 0
	m.message = "🗑️ Votre panier a été entièrement vidé !"
}

func (m *ClientModel) doCheckout() {
	if m.cart == nil || len(m.cart.Items) == 0 {
		m.message = "Votre panier est vide"
		return
	}

	data, _ := json.Marshal(map[string]string{
		"card_number":     m.cardNumber,
		"exp_date":        m.expDate,
		"cvc":             m.cvc,
		"cardholder_name": m.cardholder,
	})

	req, _ := http.NewRequest("POST", m.serverURL+"/api/checkout", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	m.setAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.message = "Erreur paiement: " + err.Error()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		m.message = "❌ Paiement refusé: " + errResp["error"]
		return
	}

	var order models.Order
	_ = json.NewDecoder(resp.Body).Decode(&order)
	m.lastOrder = &order
	m.cart = nil
	m.message = fmt.Sprintf("✅ Commande %s payée avec succès !", order.ID)
}

func (m *ClientModel) loadOrders() {
	req, _ := http.NewRequest("GET", m.serverURL+"/api/orders", nil)
	m.setAuthHeader(req)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.message = "Erreur commandes: " + err.Error()
		return
	}
	defer resp.Body.Close()

	var ords []models.Order
	if err := json.NewDecoder(resp.Body).Decode(&ords); err == nil {
		m.orders = ords
		if m.selectedOrderIdx >= len(m.orders) {
			m.selectedOrderIdx = 0
		}
	}
}



func (m ClientModel) updateAuthForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit
	case tea.KeyEsc:
		m.state = stateMenu
		m.input = ""
		return m, nil
	case tea.KeyBackspace:
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil
	case tea.KeyTab:
		if (m.state == stateRegister || m.state == stateLogin) && m.step == 0 && m.input == "" {
			m.input = "acheteur@test.fr"
			return m, nil
		}
	case tea.KeyEnter:
		val := strings.TrimSpace(m.input)
		m.input = ""

		switch m.state {
		case stateLogin:
			if m.step == 0 {
				if val == "" {
					m.message = "❌ L'adresse email est requise"
					return m, nil
				}
				m.email = val
				m.step = 1
				m.message = ""
				return m, nil
			}
			if val == "" {
				m.message = "❌ Le mot de passe est requis"
				return m, nil
			}
			m.password = val
			return m.doLogin()

		case stateRegister:
			if m.step == 0 {
				if val == "" {
					m.message = "❌ L'adresse email est requise"
					return m, nil
				}
				if !strings.Contains(val, "@") || !strings.Contains(val, ".") {
					m.message = fmt.Sprintf("❌ Format email invalide : '@' ou '.' manquant (Saisi : « %s »)", val)
					m.input = val
					return m, nil
				}
				m.email = val
				m.step = 1
				m.message = ""
				return m, nil
			}
			if len(val) < 6 {
				m.message = "❌ Le mot de passe doit comporter au moins 6 caractères"
				return m, nil
			}
			m.password = val
			return m.doRegister()

		case stateConfirm:
			if m.step == 0 {
				if val == "" {
					m.message = "❌ L'adresse email est requise"
					return m, nil
				}
				m.email = val
				m.step = 1
				m.message = ""
				return m, nil
			}
			if val == "" {
				m.message = "❌ Le code de confirmation est requis"
				return m, nil
			}
			m.code = val
			return m.doConfirm()

		case stateReset:
			if m.step == 0 {
				m.email = val
				return m.doRequestReset()
			}
			if m.step == 1 {
				m.code = val
				m.step = 2
				return m, nil
			}
			m.password = val
			return m.doResetPassword()
		}
	default:
		s := msg.String()
		if strings.HasSuffix(s, "@") {
			m.input += "@"
		} else if strings.HasPrefix(s, "alt+") && len(s) > 4 {
			m.input += s[4:]
		} else if strings.HasPrefix(s, "ctrl+alt+") && len(s) > 9 {
			m.input += s[9:]
		} else if len(msg.Runes) > 0 {
			m.input += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			m.input += " "
		} else if len(s) == 1 {
			m.input += s
		}
	}
	return m, nil
}

func (m ClientModel) doLogin() (tea.Model, tea.Cmd) {
	data, _ := json.Marshal(models.LoginRequest{
		Email:    m.email,
		Password: m.password,
	})

	resp, err := http.Post(m.serverURL+"/api/auth/login", "application/json", bytes.NewBuffer(data))
	if err != nil {
		m.message = "❌ Erreur connexion: " + err.Error()
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		m.message = "❌ " + errResp["error"]
		m.step = 1
		return m, nil
	}

	var authResp models.AuthResponse
	_ = json.NewDecoder(resp.Body).Decode(&authResp)
	m.token = authResp.Token
	m.user = authResp.User
	m.message = "✅ Connexion réussie : " + m.email
	m.state = stateMenu
	m.cursor = 0
	return m, nil
}

func (m ClientModel) doRegister() (tea.Model, tea.Cmd) {
	data, _ := json.Marshal(models.RegisterRequest{
		Email:    m.email,
		Password: m.password,
	})

	resp, err := http.Post(m.serverURL+"/api/auth/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		m.message = "❌ Erreur connexion: " + err.Error()
		m.step = 0
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		m.message = "❌ " + errResp["error"]
		m.step = 0
		m.input = m.email
		return m, nil
	}

	var authResp models.AuthResponse
	_ = json.NewDecoder(resp.Body).Decode(&authResp)
	m.code = authResp.ConfirmationCode
	m.state = stateConfirm
	m.step = 1
	m.input = authResp.ConfirmationCode
	m.message = fmt.Sprintf("🎉 Compte créé ! Votre code unique est : %s\n(Appuyez sur [Entrée] pour confirmer votre compte)", authResp.ConfirmationCode)
	return m, nil
}

func (m ClientModel) doConfirm() (tea.Model, tea.Cmd) {
	data, _ := json.Marshal(models.ConfirmRequest{
		Email: m.email,
		Code:  m.code,
	})

	resp, err := http.Post(m.serverURL+"/api/auth/confirm", "application/json", bytes.NewBuffer(data))
	if err != nil {
		m.message = "❌ Erreur connexion: " + err.Error()
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		m.message = "❌ " + errResp["error"]
		return m, nil
	}

	m.message = "✅ Compte confirmé avec succès ! Vous pouvez maintenant vous connecter."
	m.state = stateLogin
	m.step = 0
	m.input = m.email
	return m, nil
}

func (m ClientModel) doRequestReset() (tea.Model, tea.Cmd) {
	data, _ := json.Marshal(models.ForgotPasswordRequest{Email: m.email})
	resp, err := http.Post(m.serverURL+"/api/auth/forgot-password", "application/json", bytes.NewBuffer(data))
	if err != nil {
		m.message = "Erreur connexion: " + err.Error()
		m.state = stateMenu
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		m.message = "Erreur: " + errResp["error"]
		m.state = stateMenu
		return m, nil
	}

	var authResp models.AuthResponse
	_ = json.NewDecoder(resp.Body).Decode(&authResp)
	m.message = fmt.Sprintf("Code de reset généré: %s", authResp.ResetCode)
	m.step = 1
	return m, nil
}

func (m ClientModel) doResetPassword() (tea.Model, tea.Cmd) {
	data, _ := json.Marshal(models.ResetPasswordRequest{
		Email:       m.email,
		Code:        m.code,
		NewPassword: m.password,
	})

	resp, err := http.Post(m.serverURL+"/api/auth/reset-password", "application/json", bytes.NewBuffer(data))
	if err != nil {
		m.message = "Erreur connexion: " + err.Error()
		m.state = stateMenu
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		m.message = "Erreur: " + errResp["error"]
		m.state = stateMenu
		return m, nil
	}

	m.message = "Mot de passe réinitialisé avec succès !"
	m.state = stateMenu
	return m, nil
}

func (m *ClientModel) loadProfile() {
	if m.token == "" {
		m.message = "Non connecté"
		return
	}

	req, err := http.NewRequest("GET", m.serverURL+"/api/auth/me", nil)
	if err != nil {
		m.message = "Erreur requête: " + err.Error()
		return
	}
	req.Header.Set("Authorization", "Bearer "+m.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		m.message = "Erreur connexion: " + err.Error()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.message = "Jeton JWT invalide ou expiré"
		m.user = nil
		m.token = ""
		return
	}

	var u models.User
	if err := json.NewDecoder(resp.Body).Decode(&u); err == nil {
		m.user = &u
	}
}



func (m ClientModel) View() string {
	switch m.state {
	case stateProfile:
		if m.user == nil {
			return "Aucune session active. Connectez-vous d'abord.\n\nAppuyez sur [Entrée] pour revenir.\n"
		}
		return fmt.Sprintf("👤 Profil utilisateur:\nID: %d\nEmail: %s\nRole: %s\nConfirmé: %t\nCréé le: %s\n\nAppuyez sur [Entrée] ou [Echap] pour revenir au menu.\n",
			m.user.ID, m.user.Email, m.user.Role, m.user.IsConfirmed,
			m.user.CreatedAt.Format("02/01/2006 15:04"))

	case stateLogin:
		msgBox := ""
		if m.message != "" {
			msgBox = "\n" + m.message + "\n"
		}
		if m.step == 0 {
			return fmt.Sprintf("🔐 Connexion%s\nEmail : %s█\n\n[Entrée] continuer | [Tab] pré-remplir 'acheteur@test.fr' | [Echap] annuler\n", msgBox, m.input)
		}
		return fmt.Sprintf("🔐 Connexion%s\nEmail : %s\nMot de passe : %s█\n\n[Entrée] valider, [Echap] annuler\n", msgBox, m.email, strings.Repeat("*", len(m.input)))

	case stateRegister:
		msgBox := ""
		if m.message != "" {
			msgBox = "\n" + m.message + "\n"
		}
		if m.step == 0 {
			return fmt.Sprintf("📝 Inscription%s\nEmail : %s█\n\n[Entrée] continuer | [Tab] pré-remplir 'acheteur@test.fr' | [Echap] annuler\n", msgBox, m.input)
		}
		return fmt.Sprintf("📝 Inscription%s\nEmail : %s\nMot de passe : %s█ (min. 6 caractères)\n\n[Entrée] valider, [Echap] annuler\n", msgBox, m.email, strings.Repeat("*", len(m.input)))

	case stateConfirm:
		msgBox := ""
		if m.message != "" {
			msgBox = "\n" + m.message + "\n"
		}
		if m.step == 0 {
			return fmt.Sprintf("✉️ Confirmation d'inscription%s\nEmail : %s█\n\n[Entrée] continuer, [Echap] annuler\n", msgBox, m.input)
		}
		return fmt.Sprintf("✉️ Confirmation d'inscription%s\nEmail : %s\nCode reçu : %s█\n\n[Entrée] valider, [Echap] annuler\n", msgBox, m.email, m.input)

	case stateReset:
		switch m.step {
		case 0:
			return fmt.Sprintf("🔄 Réinitialisation mot de passe\nEmail : %s█\n\n[Entrée] demander un code, [Echap] annuler\n", m.input)
		case 1:
			return fmt.Sprintf("🔄 Réinitialisation mot de passe\nEmail : %s\nCode reçu : %s█\n\n[Entrée] continuer, [Echap] annuler\n", m.email, m.input)
		case 2:
			return fmt.Sprintf("🔄 Réinitialisation mot de passe\nEmail : %s\nCode : %s\nNouveau mot de passe : %s█\n\n[Entrée] valider, [Echap] annuler\n", m.email, m.code, strings.Repeat("*", len(m.input)))
		}

	case stateCatalog:
		return m.renderCatalog()
	case stateCart:
		return m.renderCart()
	case stateCheckout:
		return m.renderCheckout()
	case stateOrders:
		return m.renderOrders()
	}

	// Menu Principal
	var b strings.Builder
	b.WriteString(headerStyle.Render("🛒 E-COMMERCE CLI — ESPACE CLIENT"))
	b.WriteString("\n\n")

	if m.user != nil {
		b.WriteString(fmt.Sprintf("Connecté en tant que: %s (Rôle: %s)\n",
			lipgloss.NewStyle().Bold(true).Foreground(colorCyan).Render(m.user.Email), m.user.Role))
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render("Mode Invité — Non connecté\n"))
	}

	if m.message != "" {
		b.WriteString(fmt.Sprintf("Message : %s\n\n", m.message))
	}

	cartCount := 0
	var cartTotal float64
	if m.cart != nil {
		cartCount = len(m.cart.Items)
		cartTotal = m.cart.TotalTTC
	}

	var choices []string
	if m.user == nil {
		choices = []string{
			"1. 📦 Parcourir le Catalogue & Rechercher",
			fmt.Sprintf("2. 🛒 Mon Panier (%d articles - %.2f €)", cartCount, cartTotal),
			"3. 💳 Passer au Paiement & Commander",
			"4. 📋 Mes Commandes & Suivi",
			"5. 🔐 Se connecter",
			"6. 📝 Créer un compte (inscription)",
			"7. ✉️ Confirmer mon compte avec un code",
			"8. 🔄 Réinitialiser mon mot de passe",
			"9. 🚪 Quitter",
		}
	} else {
		choices = []string{
			"1. 📦 Parcourir le Catalogue & Rechercher",
			fmt.Sprintf("2. 🛒 Mon Panier (%d articles - %.2f €)", cartCount, cartTotal),
			"3. 💳 Passer au Paiement & Commander",
			"4. 📋 Mes Commandes & Suivi",
			"5. 👤 Voir mon Profil / Session",
			"6. 🚪 Déconnexion",
			"7. 🚪 Quitter",
		}
	}

	for i, c := range choices {
		cursor := "  "
		style := normalRowStyle
		if m.cursor == i {
			cursor = "▶ "
			style = selectedRowStyle
		}
		b.WriteString(style.Render(cursor+c) + "\n")
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(colorMuted).Render("[Haut/Bas] naviguer, [Entrée] sélectionner, 'q' quitter"))
	return b.String()
}

func (m ClientModel) renderCatalog() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("📦 CATALOGUE & RECHERCHE DE PRODUITS"))
	b.WriteString("\n\n")

	if m.message != "" {
		if strings.HasPrefix(m.message, "❌") {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444")).Render(m.message) + "\n\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981")).Render(m.message) + "\n\n")
		}
	}

	cartInfo := "🛒 Panier actuel : 0 article (0.00 €)"
	if m.cart != nil && len(m.cart.Items) > 0 {
		cartInfo = fmt.Sprintf("🛒 Panier actuel [%s] : %d article(s)  |  Total: %s  (Appuyez sur [2] pour afficher)",
			badgeBasket.Render(m.cart.ID),
			len(m.cart.Items),
			lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render(fmt.Sprintf("%.2f € TTC", m.cart.TotalTTC)),
		)
	}
	b.WriteString(boxStyle.Render(cartInfo))
	b.WriteString("\n\n")

	queryTxt := m.searchQuery
	if m.isSearching {
		queryTxt = lipgloss.NewStyle().Background(colorBrand).Foreground(colorTextLight).Render(m.searchQuery + "█")
	} else if queryTxt == "" {
		queryTxt = lipgloss.NewStyle().Foreground(colorMuted).Render("(Appuyez sur '/' pour rechercher)")
	}

	catName := "Toutes"
	if m.selectedCategoryIdx < len(m.categories) {
		catName = m.categories[m.selectedCategoryIdx]
	}

	sortNames := []string{"Date (récent)", "Nom (A-Z)", "Prix croissant", "Prix décroissant"}
	filterTxt := fmt.Sprintf("🔍 Recherche : %s  |  📂 Catégorie [c] : %s  |  🏷️ Affichage [t] : %s  |  🔄 Tri [s] : %s",
		queryTxt,
		lipgloss.NewStyle().Bold(true).Foreground(colorCyan).Render(catName),
		lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render(strings.ToUpper(m.priceFilterType)),
		lipgloss.NewStyle().Bold(true).Foreground(colorOrange).Render(sortNames[m.sortIdx]),
	)
	b.WriteString(boxStyle.Render(filterTxt))
	b.WriteString("\n\n")

	if len(m.products) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render("Aucun produit ne correspond à vos critères de recherche.\n"))
		b.WriteString("\n[Echap] Retour au menu")
		return b.String()
	}

	header := fmt.Sprintf("%-14s %-30s %-16s %-12s %-12s %-8s", "RÉFÉRENCE", "NOM", "CATÉGORIE", "PRIX HT", "PRIX TTC", "STOCK")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorBrand).Render(header))
	b.WriteString("\n" + strings.Repeat("─", 96) + "\n")

	for i, p := range m.products {
		refStr := badgeProduct.Render(p.ID)
		nameStr := p.Name
		if len(nameStr) > 28 {
			nameStr = nameStr[:25] + "..."
		}
		catStr := p.Category
		if len(catStr) > 15 {
			catStr = catStr[:13] + ".."
		}
		htStr := fmt.Sprintf("%.2f €", p.PriceHT)
		ttcStr := fmt.Sprintf("%.2f €", p.PriceTTC)
		stockStr := fmt.Sprintf("%d", p.Stock)

		line := fmt.Sprintf("%-14s %-30s %-16s %-12s %-12s %-8s", refStr, nameStr, catStr, htStr, ttcStr, stockStr)

		if i == m.selectedProductIdx {
			b.WriteString(selectedRowStyle.Render("▶ " + line))
		} else {
			b.WriteString(normalRowStyle.Render("  " + line))
		}
		b.WriteString("\n")
	}

	if m.selectedProductIdx < len(m.products) {
		p := m.products[m.selectedProductIdx]
		desc := fmt.Sprintf("📦 %s — %s\n📝 Description : %s\n💡 Touche [Entrée] ou [+] pour ajouter au panier",
			p.Name, p.ID, p.Description)
		b.WriteString("\n" + summaryBoxStyle.Render(desc))
	}

	searchHint := ""
	if m.isSearching {
		searchHint = lipgloss.NewStyle().Bold(true).Foreground(colorOrange).Render("🔍 MODE RECHERCHE : tapez votre mot-clé puis appuyez sur [Entrée] pour valider ou [Echap] pour quitter la recherche") + "\n"
	}
	b.WriteString("\n\n" + searchHint + lipgloss.NewStyle().Foreground(colorMuted).Render("[Entrée]/[+] Ajouter panier | [x] Vider Panier | [2] Voir Panier | [/] Recherche | [c] Catégorie | [t] HT/TTC | [s] Tri | [Echap] Menu"))
	return b.String()
}

func (m ClientModel) renderCart() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("🛒 MON PANIER CLIENT"))
	b.WriteString("\n\n")

	if m.message != "" {
		if strings.HasPrefix(m.message, "❌") {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444")).Render(m.message) + "\n\n")
		} else {
			b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981")).Render(m.message) + "\n\n")
		}
	}

	if m.cart == nil || len(m.cart.Items) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render("Votre panier est actuellement vide.\nVisitez le catalogue pour ajouter des articles !\n\n[1] Catalogue | [Echap] Menu"))
		return b.String()
	}

	ref := badgeBasket.Render(m.cart.ID)
	b.WriteString(fmt.Sprintf("Identifiant métier du panier : %s (Compte client #%d)\n\n", ref, m.cart.UserID))

	header := fmt.Sprintf("%-14s %-32s %-8s %-12s %-12s", "RÉF", "PRODUIT", "QTÉ", "UNIT. TTC", "TOTAL TTC")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorBrand).Render(header))
	b.WriteString("\n" + strings.Repeat("─", 84) + "\n")

	for i, it := range m.cart.Items {
		refStr := badgeProduct.Render(it.ProductID)
		nameStr := it.Product.Name
		if len(nameStr) > 30 {
			nameStr = nameStr[:27] + "..."
		}
		line := fmt.Sprintf("%-14s %-32s %-8d %-12.2f € %-12.2f €",
			refStr, nameStr, it.Quantity, it.Product.PriceTTC, it.Product.PriceTTC*float64(it.Quantity))

		if i == m.selectedCartIdx {
			b.WriteString(selectedRowStyle.Render("▶ " + line))
		} else {
			b.WriteString(normalRowStyle.Render("  " + line))
		}
		b.WriteString("\n")
	}

	summary := fmt.Sprintf(
		"Frais de livraison : %10s (Toujours gratuits)\n"+
			"────────────────────────────────────────\n"+
			"TOTAL TTC DU PANIER: %10.2f €",
		"0.00 €", m.cart.TotalTTC,
	)
	b.WriteString("\n" + summaryBoxStyle.Render(summary))
	b.WriteString("\n\n" + lipgloss.NewStyle().Foreground(colorMuted).Render("[+] Augmenter | [-] Diminuer | [d] Supprimer | [x] Vider | [p] Payer la commande | [Echap] Menu"))

	return b.String()
}

func (m ClientModel) renderCheckout() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("💳 TUNNEL DE PAIEMENT & VALIDATION"))
	b.WriteString("\n\n")

	if m.lastOrder != nil {
		orderBadge := badgeOrder.Render(m.lastOrder.ID)
		statusBadge := statusPaid.Render(" PAYÉ ")

		msg := fmt.Sprintf(
			"🎉 PAIEMENT ACCEPTÉ & COMMANDE VALIDÉE 🎉\n\n"+
				"Identifiant unique : %s\n"+
				"Date               : %s\n"+
				"Statut             : %s\n"+
				"Total réglé        : %.2f € TTC (Frais de livraison gratuits)\n"+
				"Carte fictive      : %s\n\n"+
				"👉 Appuyez sur [4] pour consulter vos Commandes ou [Echap] pour revenir au menu.",
			orderBadge,
			m.lastOrder.CreatedAt.Format("02/01/2006 15:04"),
			statusBadge,
			m.lastOrder.TotalTTC,
			m.cardNumber,
		)
		b.WriteString(boxStyle.Render(msg))
		return b.String()
	}

	if m.cart == nil || len(m.cart.Items) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render("Votre panier est vide. Impossible de procéder au paiement.\n\n[1] Catalogue | [Echap] Menu"))
		return b.String()
	}

	b.WriteString(fmt.Sprintf("Règlement du Panier %s  |  Montant Total TTC : %.2f €\n\n",
		badgeBasket.Render(m.cart.ID), m.cart.TotalTTC))

	fields := []struct {
		Label string
		Value string
	}{
		{"1. Numéro de Carte (16 chiffres)", m.cardNumber},
		{"2. Date d'Expiration (MM/YY)    ", m.expDate},
		{"3. Code CVC (3 chiffres)        ", m.cvc},
		{"4. Nom du Titulaire             ", m.cardholder},
	}

	for i, f := range fields {
		cursor := "  "
		valStyle := normalRowStyle
		if i == m.focusedField {
			cursor = "▶ "
			valStyle = lipgloss.NewStyle().Bold(true).Foreground(colorCyan)
		}
		b.WriteString(fmt.Sprintf("%s%s : %s\n", cursor, f.Label, valStyle.Render("["+f.Value+"]")))
	}

	b.WriteString("\n")
	btnText := " [ VALIDER LE PAIEMENT FICTIF ] "
	if m.focusedField == 4 {
		b.WriteString(lipgloss.NewStyle().Bold(true).Background(colorGreen).Foreground(lipgloss.Color("#FFFFFF")).Render(btnText))
	} else {
		b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#374151")).Foreground(colorMuted).Render(btnText))
	}

	b.WriteString("\n\n" + lipgloss.NewStyle().Foreground(colorMuted).Render("(Navigation : Tab / Entrée / Flèches Haut & Bas | [Echap] Menu)"))
	return b.String()
}

func (m ClientModel) renderOrders() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("📋 HISTORIQUE DES COMMANDES CLIENT"))
	b.WriteString("\n\n")

	if len(m.orders) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorMuted).Render("Aucune commande enregistrée pour le moment.\nValidez un panier dans [3] Paiement pour créer une commande !\n\n[Echap] Menu"))
		return b.String()
	}

	header := fmt.Sprintf("%-16s %-20s %-14s %-24s", "N° COMMANDE", "DATE", "TOTAL TTC", "STATUT")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorBrand).Render(header))
	b.WriteString("\n" + strings.Repeat("─", 80) + "\n")

	for i, o := range m.orders {
		refStr := badgeOrder.Render(o.ID)
		dateStr := o.CreatedAt.Format("02/01/2006 15:04")
		totStr := fmt.Sprintf("%.2f €", o.TotalTTC)

		var statusStr string
		switch o.Status {
		case models.StatusPaid:
			statusStr = statusPaid.Render(" PAYÉ ")
		case models.StatusDelivering:
			statusStr = statusShipping.Render(" EN COURS LIVRAISON ")
		case models.StatusDelivered:
			statusStr = statusDelivered.Render(" LIVRÉ ")
		case models.StatusCancelled:
			statusStr = statusCancelled.Render(" ANNULÉ ")
		default:
			statusStr = statusPending.Render(" EN ATTENTE ")
		}

		line := fmt.Sprintf("%-16s %-20s %-14s %-24s", refStr, dateStr, totStr, statusStr)

		if i == m.selectedOrderIdx {
			b.WriteString(selectedRowStyle.Render("▶ " + line))
		} else {
			b.WriteString(normalRowStyle.Render("  " + line))
		}
		b.WriteString("\n")
	}

	if m.selectedOrderIdx < len(m.orders) {
		selected := m.orders[m.selectedOrderIdx]
		b.WriteString("\n")

		var details strings.Builder
		details.WriteString(fmt.Sprintf("Détail de la commande %s\n", badgeOrder.Render(selected.ID)))

		if selected.Status == models.StatusCancelled {
			reason := selected.CancelReason
			if reason == "" {
				reason = "Non spécifiée"
			}
			details.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorRed).Render(fmt.Sprintf("⚠️ RAISON D'ANNULATION : %s\n", reason)))
		}

		details.WriteString("Articles achetés :\n")
		for _, it := range selected.Items {
			details.WriteString(fmt.Sprintf("  • %-12s %-28s x%-2d (%.2f € TTC)\n",
				it.ProductID, it.ProductName, it.Quantity, it.PriceUnitTTC*float64(it.Quantity)))
		}
		details.WriteString("Frais de livraison : 0.00 € (GRATUITS)")

		b.WriteString(summaryBoxStyle.Render(details.String()))
	}

	b.WriteString("\n\n" + lipgloss.NewStyle().Foreground(colorMuted).Render("[Haut/Bas] Naviguer | [1] Catalogue | [2] Panier | [Echap] Menu"))
	return b.String()
}
