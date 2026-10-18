package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/LMV0911/Go_e-commerce/internal/models"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateMenu state = iota
	stateLogin
	stateRegister
	stateConfirm
	stateReset
	stateProfile
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
}

func NewClientModel(serverURL string) ClientModel {
	return ClientModel{
		serverURL: serverURL,
		state:     stateMenu,
		message:   "Pret",
	}
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

	if m.state != stateMenu {
		return m.updateForm(keyMsg)
	}

	switch keyMsg.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < 5 {
			m.cursor++
		}
	case "enter":
		m.step = 0
		m.input = ""
		switch m.cursor {
		case 0:
			m.state = stateLogin
		case 1:
			m.state = stateRegister
		case 2:
			m.state = stateConfirm
		case 3:
			m.state = stateReset
		case 4:
			m.loadProfile()
			m.state = stateProfile
		case 5:
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m ClientModel) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
	case tea.KeyEnter:
		val := strings.TrimSpace(m.input)
		m.input = ""

		switch m.state {
		case stateLogin:
			if m.step == 0 {
				m.email = val
				m.step = 1
				return m, nil
			}
			m.password = val
			return m.doLogin()

		case stateRegister:
			if m.step == 0 {
				m.email = val
				m.step = 1
				return m, nil
			}
			m.password = val
			return m.doRegister()

		case stateConfirm:
			if m.step == 0 {
				m.email = val
				m.step = 1
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
		if msg.Type == tea.KeyRunes {
			m.input += string(msg.Runes)
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
	m.token = authResp.Token
	m.user = authResp.User
	m.message = "Connexion reussie: " + m.email
	m.state = stateMenu
	return m, nil
}

func (m ClientModel) doRegister() (tea.Model, tea.Cmd) {
	data, _ := json.Marshal(models.RegisterRequest{
		Email:    m.email,
		Password: m.password,
	})

	resp, err := http.Post(m.serverURL+"/api/auth/register", "application/json", bytes.NewBuffer(data))
	if err != nil {
		m.message = "Erreur connexion: " + err.Error()
		m.state = stateMenu
		return m, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errResp map[string]string
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		m.message = "Erreur: " + errResp["error"]
		m.state = stateMenu
		return m, nil
	}

	var authResp models.AuthResponse
	_ = json.NewDecoder(resp.Body).Decode(&authResp)
	m.message = fmt.Sprintf("Compte cree ! Code unique: %s", authResp.ConfirmationCode)
	m.state = stateMenu
	return m, nil
}

func (m ClientModel) doConfirm() (tea.Model, tea.Cmd) {
	data, _ := json.Marshal(models.ConfirmRequest{
		Email: m.email,
		Code:  m.code,
	})

	resp, err := http.Post(m.serverURL+"/api/auth/confirm", "application/json", bytes.NewBuffer(data))
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

	m.message = "Compte confirme avec succes !"
	m.state = stateMenu
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
	m.message = fmt.Sprintf("Code de reset genere: %s", authResp.ResetCode)
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

	m.message = "Mot de passe reinitialise avec succes !"
	m.state = stateMenu
	return m, nil
}

func (m *ClientModel) loadProfile() {
	if m.token == "" {
		m.message = "Non connecte"
		return
	}

	req, err := http.NewRequest("GET", m.serverURL+"/api/auth/me", nil)
	if err != nil {
		m.message = "Erreur requete: " + err.Error()
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
		m.message = "Jeton JWT invalide ou expire"
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
	if m.state == stateProfile {
		if m.user == nil {
			return "Aucune session active. Connectez-vous d'abord.\n\nAppuyez sur [Entree] pour revenir.\n"
		}
		return fmt.Sprintf("Profil utilisateur:\nID: %d\nEmail: %s\nRole: %s\nConfirme: %t\nCree le: %s\nToken: %s\n\nAppuyez sur [Entree] pour revenir au menu.\n",
			m.user.ID, m.user.Email, m.user.Role, m.user.IsConfirmed,
			m.user.CreatedAt.Format("02/01/2006 15:04"), m.token)
	}

	if m.state == stateLogin {
		if m.step == 0 {
			return fmt.Sprintf("Connexion\nEmail : %s\n\n[Entree] continuer, [Echap] annuler\n", m.input)
		}
		return fmt.Sprintf("Connexion\nEmail : %s\nMot de passe : %s\n\n[Entree] valider, [Echap] annuler\n", m.email, strings.Repeat("*", len(m.input)))
	}

	if m.state == stateRegister {
		if m.step == 0 {
			return fmt.Sprintf("Inscription\nEmail : %s\n\n[Entree] continuer, [Echap] annuler\n", m.input)
		}
		return fmt.Sprintf("Inscription\nEmail : %s\nMot de passe : %s\n\n[Entree] valider, [Echap] annuler\n", m.email, strings.Repeat("*", len(m.input)))
	}

	if m.state == stateConfirm {
		if m.step == 0 {
			return fmt.Sprintf("Confirmation d'inscription\nEmail : %s\n\n[Entree] continuer, [Echap] annuler\n", m.input)
		}
		return fmt.Sprintf("Confirmation d'inscription\nEmail : %s\nCode : %s\n\n[Entree] valider, [Echap] annuler\n", m.email, m.input)
	}

	if m.state == stateReset {
		switch m.step {
		case 0:
			return fmt.Sprintf("Reinitialisation mot de passe\nEmail : %s\n\n[Entree] demander un code, [Echap] annuler\n", m.input)
		case 1:
			return fmt.Sprintf("Reinitialisation mot de passe\nEmail : %s\nCode recu : %s\n\n[Entree] continuer, [Echap] annuler\n", m.email, m.input)
		case 2:
			return fmt.Sprintf("Reinitialisation mot de passe\nEmail : %s\nCode : %s\nNouveau mot de passe : %s\n\n[Entree] valider, [Echap] annuler\n", m.email, m.code, strings.Repeat("*", len(m.input)))
		}
	}

	choices := []string{
		"1. Se connecter",
		"2. Creer un compte (inscription)",
		"3. Confirmer mon compte avec un code",
		"4. Reinitialiser mon mot de passe",
		"5. Voir mon profil / session JWT",
		"6. Quitter",
	}

	var b strings.Builder
	b.WriteString("=== E-Commerce Client ===\n")
	if m.user != nil {
		b.WriteString(fmt.Sprintf("Utilisateur: %s (%s)\n", m.user.Email, m.user.Role))
	}
	b.WriteString(fmt.Sprintf("Message: %s\n\n", m.message))

	for i, c := range choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, c))
	}

	b.WriteString("\n[Haut/Bas] naviguer, [Entree] selectionner, 'q' quitter.\n")
	return b.String()
}
