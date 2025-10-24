package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"journalCli/db"
	"journalCli/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var url string = "http://localhost:8080"

var (
	// Titles and section headers
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#6C63FF")). // indigo
			PaddingBottom(1)

	// Input boxes
	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#5EEAD4")). // cyan border
			Padding(0, 1).
			Width(30)

	// Buttons
	buttonStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#0D0D1A")). // dark text
			Background(lipgloss.Color("#6C63FF")). // indigo background
			Padding(0, 2).
			MarginTop(1)

	// Error messages
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")) // red

	// Muted / secondary info
	helpTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94A3B8")).
			Italic(true)

	// Background & text (for future use)
	bgStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#0D0D1A")).
		Foreground(lipgloss.Color("#E6E6E6"))
)

type Page int

type Model struct {
	page            Page
	msg             string
	user            User
	err             error
	inputing        bool
	textarea        textarea.Model
	username        textinput.Model
	password        textinput.Model
	email           textinput.Model
	confirmPassword textinput.Model
	Focused         int
	width           int
	height          int
	senderStyle     lipgloss.Style
	Client          *http.Client
}

type User struct {
	Id       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type LoginSuccessMsg struct {
	User User
}

type SignupSuccessMsg struct {
	User User
}

type ErrMsg struct {
	err error
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

const (
	PageLogin = iota
	PageSignup
	PageMenu
	PageJournal
	PageRead
	PageSettings
	PageHelp
)

func initialModel() Model {
	username := textinput.New()
	username.Placeholder = "Username"
	username.Focus()
	username.CharLimit = 32
	username.Width = 30

	email := textinput.New()
	email.Placeholder = "Email"
	email.CharLimit = 32
	email.Width = 30

	password := textinput.New()
	password.Placeholder = "Password"
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = '•'
	password.CharLimit = 32
	password.Width = 30

	confirmPassword := textinput.New()
	confirmPassword.Placeholder = "Confirm Password"
	confirmPassword.EchoMode = textinput.EchoPassword
	confirmPassword.EchoCharacter = '•'
	confirmPassword.CharLimit = 32
	confirmPassword.Width = 30

	return Model{
		page:            PageLogin,
		senderStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")),
		username:        username,
		email:           email,
		password:        password,
		confirmPassword: confirmPassword,
		inputing:        true,
		Client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func checkServerLogin(email, password string, client *http.Client) tea.Msg {
	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	body, err := json.Marshal(loginReq)

	if err != nil {
		return ErrMsg{err}
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/login", url), bytes.NewBuffer(body))

	if err != nil {
		return ErrMsg{err}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", " application/json")

	//send the request
	res, err := client.Do(req)

	if err != nil {
		return ErrMsg{err}
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)

	if err != nil {
		return ErrMsg{err}
	}

	if res.StatusCode != http.StatusOK {
		return ErrMsg{fmt.Errorf("server returned status: %s, message: %s", res.Status, string(b))}
	}

	var user User
	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
		return ErrMsg{err}
	}

	return LoginSuccessMsg{User: user}
}

func checkServerSignup(username, email, password string, client *http.Client) tea.Msg {
	signupReq := SignupRequest{
		Username: username,
		Email:    email,
		Password: password,
	}
	body, err := json.Marshal(signupReq)
	if err != nil {
		return ErrMsg{err}
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/signup", url), bytes.NewBuffer(body))

	if err != nil {
		return ErrMsg{err}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", " application/json")

	res, err := client.Do(req)

	if err != nil {
		return ErrMsg{err}
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)

	if err != nil {
		return ErrMsg{err}
	}

	if res.StatusCode != http.StatusCreated {
		return ErrMsg{fmt.Errorf("server returned status: %s, message: %s", res.Status, string(b))}
	}

	var user User

	if err := json.NewDecoder(bytes.NewBuffer(b)).Decode(&user); err != nil {
		return ErrMsg{err}
	}

	return SignupSuccessMsg{User: user}
}

func (m *Model) updateFocusSignup() {
	switch m.Focused {
	case 0:
		m.username.Focus()
		m.email.Blur()
		m.password.Blur()
		m.confirmPassword.Blur()
	case 1:
		m.username.Blur()
		m.email.Focus()
		m.password.Blur()
		m.confirmPassword.Blur()
	case 2:
		m.username.Blur()
		m.email.Blur()
		m.password.Focus()
		m.confirmPassword.Blur()
	case 3:
		m.username.Blur()
		m.email.Blur()
		m.password.Blur()
		m.confirmPassword.Focus()
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// ----------- SERVER RESPONSES -----------
	case LoginSuccessMsg:
		m.user = msg.User
		m.page = PageMenu
		m.inputing = false

	case SignupSuccessMsg:
		m.page = PageMenu
		m.inputing = false

	case ErrMsg:
		m.err = msg.err

	// ----------- KEY EVENTS -----------
	case tea.KeyMsg:
		switch m.page {

		// ----------- LOGIN PAGE -----------
		case PageLogin:
			m.username, cmd = m.username.Update(msg)
			cmds = append(cmds, cmd)
			m.password, cmd = m.password.Update(msg)
			cmds = append(cmds, cmd)

			switch msg.Type {
			case tea.KeyCtrlS:
				m.page = PageSignup
				m.Focused = 0
				m.username.SetValue("")
				m.password.SetValue("")
				m.username.Focus()
				m.password.Blur()
			case tea.KeyTab, tea.KeyDown, tea.KeyUp:
				m.Focused = (m.Focused + 1) % 2
				if m.Focused == 0 {
					m.username.Focus()
					m.password.Blur()
				} else {
					m.password.Focus()
					m.username.Blur()
				}
			case tea.KeyEnter:
				username := m.username.Value()
				password := m.password.Value()
				return m, func() tea.Msg { return checkServerLogin(username, password, m.Client) }
			case tea.KeyCtrlC:
				return m, tea.Quit
			}
			return m, tea.Batch(cmds...)

		// ----------- SIGNUP PAGE -----------
		case PageSignup:
			m.username, cmd = m.username.Update(msg)
			cmds = append(cmds, cmd)

			m.email, cmd = m.email.Update(msg)
			cmds = append(cmds, cmd)

			m.password, cmd = m.password.Update(msg)
			cmds = append(cmds, cmd)

			m.confirmPassword, cmd = m.confirmPassword.Update(msg)
			cmds = append(cmds, cmd)

			switch msg.Type {
			case tea.KeyCtrlL:
				m.page = PageLogin
				m.Focused = 0
				m.username.SetValue("")
				m.password.SetValue("")

				m.confirmPassword.SetValue("")
				m.username.Focus()
				m.password.Blur()
				m.confirmPassword.Blur()

			case tea.KeyTab, tea.KeyDown:
				m.Focused = (m.Focused + 1) % 4
				m.updateFocusSignup()
			case tea.KeyUp:
				m.Focused = (m.Focused + 3) % 4
				m.updateFocusSignup()
			case tea.KeyEnter:
				username := m.username.Value()
				password := m.password.Value()
				email := m.email.Value()
				confirmPassword := m.confirmPassword.Value()
				isValid, err := utils.ValidateCredentials(username, password, confirmPassword)
				if !isValid {
					m.err = err
				} else {
					return m, func() tea.Msg { return checkServerSignup(username, email, password, m.Client) }
				}

			case tea.KeyCtrlC:
				return m, tea.Quit
			}
			return m, tea.Batch(cmds...)

		// ----------- MENU PAGE -----------
		case PageMenu:
			switch msg.String() {
			case "1":
				m.page = PageJournal
			case "2":
				m.page = PageRead
			case "3":
				m.page = PageSettings
			case "4":
				m.page = PageHelp
			case "q", "ctrl+c":
				return m, tea.Quit
			}

		// ----------- JOURNAL PAGE -----------
		case PageJournal:
			if msg.String() == "b" {
				m.page = PageMenu
			}

		// ----------- READ PAGE -----------
		case PageRead:
			if msg.String() == "b" {
				m.page = PageMenu
			}

		// ----------- SETTINGS PAGE -----------
		case PageSettings:
			if msg.String() == "b" {
				m.page = PageMenu
			}

		// ----------- HELP PAGE -----------
		case PageHelp:
			if msg.String() == "b" {
				m.page = PageMenu
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	switch m.page {
	case PageLogin:
		return renderLoginPage(m)
	case PageSignup:
		return renderSignupPage(m)
	case PageMenu:
		return renderWelcomeMsg(m) + "Menu Page\n\n1. Journal\n2. Read\n3. Settings\n4. Help\nq. Quit"
	case PageJournal:
		return "Journal Page\n\n[Journal Entries Here]\nb. Back to Menu"
	case PageRead:
		return "Read Page\n\n[Read Entries Here]\nb. Back to Menu"
	case PageSettings:
		return "Settings Page\n\n[Settings Here]\nb. Back to Menu"
	case PageHelp:
		return "Help Page\n\n[Help Info Here]\nb. Back to Menu"
	default:
		return "Unknown Page"
	}
}

func renderGoogleLogo() string {
	coloredGoogleLogo := lipgloss.NewStyle().Foreground(lipgloss.Color("#4285F4")).Render("G")
	coloredGoogleLogo += lipgloss.NewStyle().Foreground(lipgloss.Color("#DB4437")).Render("o")
	coloredGoogleLogo += lipgloss.NewStyle().Foreground(lipgloss.Color("#F4B400")).Render("o")
	coloredGoogleLogo += lipgloss.NewStyle().Foreground(lipgloss.Color("#4285F4")).Render("g")
	coloredGoogleLogo += lipgloss.NewStyle().Foreground(lipgloss.Color("#0F9D58")).Render("l")
	coloredGoogleLogo += lipgloss.NewStyle().Foreground(lipgloss.Color("#DB4437")).Render("e")
	return coloredGoogleLogo
}

func renderLoginPage(m Model) string {
	title := titleStyle.Render("🔐 Login")
	userNameStyle := inputBoxStyle
	passwordStyle := inputBoxStyle
	googleLoginLink := lipgloss.NewStyle().Italic(true).Underline(true).Render("Press Ctrl+g to Login with")
	googleLogo := renderGoogleLogo()
	baseFooter := lipgloss.NewStyle().Italic(true).Bold(true).Render("\nFirst time? ")
	underlineFooter := lipgloss.NewStyle().Italic(true).Underline(true).Render("Press Ctrl+s to go to SignUp Page")
	if m.Focused == 0 {
		userNameStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#A78BFA"))
	}
	if m.Focused == 1 {
		passwordStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#A78BFA"))
	}

	form := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		userNameStyle.Render(m.username.View()),
		passwordStyle.Render(m.password.View()),
		googleLoginLink+" "+googleLogo,
		buttonStyle.Render("Press Enter to Submit"),
		baseFooter+underlineFooter,
	)

	if m.err != nil {
		form = lipgloss.JoinVertical(
			lipgloss.Center,
			form,
			errorStyle.Render(fmt.Sprintf("Error: %v", m.err)),
		)
	}

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		form,
	)
}

func renderSignupPage(m Model) string {
	title := titleStyle.Render("🔐 SignUp")
	googleSignupLink := lipgloss.NewStyle().Italic(true).Underline(true).Render("Press Ctrl+g to SignUp with Google")
	googleLogo := renderGoogleLogo()
	baseFooter := lipgloss.NewStyle().Italic(true).Bold(true).Render("\nAlready have an account? ")
	underlineFooter := lipgloss.NewStyle().Italic(true).Underline(true).Render("Press Ctrl+l to go to Login Page")

	userNameStyle := inputBoxStyle
	emailStyle := inputBoxStyle
	passwordStyle := inputBoxStyle
	confirmPasswordStyle := inputBoxStyle

	if m.Focused == 0 {
		userNameStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#A78BFA"))
	}
	if m.Focused == 1 {
		emailStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#A78BFA"))
	}
	if m.Focused == 2 {
		passwordStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#A78BFA"))
	}
	if m.Focused == 3 {
		confirmPasswordStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#A78BFA"))
	}

	form := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		userNameStyle.Render(m.username.View()),
		emailStyle.Render(m.email.View()),
		passwordStyle.Render(m.password.View()),
		confirmPasswordStyle.Render(m.confirmPassword.View()),
		googleSignupLink+" "+googleLogo,
		buttonStyle.Render("Press Enter to Submit"),
		baseFooter+underlineFooter,
	)

	if m.err != nil {
		form = lipgloss.JoinVertical(
			lipgloss.Center,
			form,
			errorStyle.Render(fmt.Sprintf("Error: %v", m.err)),
		)
	}
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		bgStyle.Render(form),
	)
}

func renderWelcomeMsg(m Model) string {

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#5EEAD4")).PaddingTop(1).PaddingBottom(2)
	welcomeMsg := style.Render(fmt.Sprintf("Welcome, %s", m.user.Username))

	centeredWelcomeMsg := lipgloss.Place(
		m.width,
		1,
		lipgloss.Center,
		lipgloss.Center,
		welcomeMsg,
	)
	return centeredWelcomeMsg
}

var debugFile *os.File

func main() {
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	debugFile = f
	defer f.Close()
	configDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}

	dbPath := filepath.Join(configDir, "journalCli", "journal.db")

	os.Mkdir(filepath.Dir(dbPath), 0755)

	database := db.InitDB(dbPath)

	defer db.CloseDB(database)

	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
	}
}
