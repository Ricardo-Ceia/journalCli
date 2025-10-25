package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"journalCli/db"
	"journalCli/utils"
	"net/http"
	"os"
	"strings"
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
			Foreground(lipgloss.Color("#ba8f95")). // indigo
			PaddingBottom(1)

	// Input boxes
	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#897c80")). // cyan border
			Width(30)

	// Buttons
	buttonStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#0D0D1A")). // dark text
			Background(lipgloss.Color("#CFBCDF")).
			Margin(1).
			Align(lipgloss.Center)

	// Error messages
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")) // red
)

type Page int

type tickMsg time.Time

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
	journal         textarea.Model
	currentTime     time.Time
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

	journal := textarea.New()
	journal.Placeholder = "Write your thoughts here..."
	journal.ShowLineNumbers = true
	journal.CharLimit = -1

	journal.FocusedStyle = textarea.Style{
		Base: lipgloss.NewStyle(),
	}

	journal.BlurredStyle = textarea.Style{
		Base: lipgloss.NewStyle(),
	}

	return Model{
		page:            PageLogin,
		journal:         journal,
		senderStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")),
		username:        username,
		email:           email,
		password:        password,
		confirmPassword: confirmPassword,
		inputing:        true,
		currentTime:     time.Now(),
		Client: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Init() tea.Cmd {
	return tickEverySecond()
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

	if err != nil {
		return ErrMsg{err}
	}

	if res.StatusCode != http.StatusOK {
		return ErrMsg{fmt.Errorf("server returned status: %s", res.Status)}
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

	if err != nil {
		return ErrMsg{err}
	}

	if res.StatusCode != http.StatusCreated {
		return ErrMsg{fmt.Errorf("server returned status: %s", res.Status)}
	}

	var user User

	if err := json.NewDecoder(res.Body).Decode(&user); err != nil {
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

	case tickMsg:
		m.currentTime = time.Time(msg)
		return m, tickEverySecond()
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// ----------- SERVER RESPONSES -----------
	case LoginSuccessMsg:
		m.user = msg.User
		m.page = PageMenu
		m.inputing = false

	case SignupSuccessMsg:
		m.user = msg.User
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

			if !m.inputing {
				m.journal.Focus()
				m.inputing = true
			}
			m.journal, cmd = m.journal.Update(msg)
			cmds = append(cmds, cmd)

			switch msg.Type {
			case tea.KeyCtrlS:
			//TODO
			case tea.KeyEsc:
				m.page = PageMenu
				m.journal.SetValue("")
				m.journal.Blur()
				m.inputing = false
			case tea.KeyCtrlC:
				return m, tea.Quit
			}
			return m, tea.Batch(cmds...)

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
		return renderJournal(m)
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
	userEmailStyle := inputBoxStyle
	passwordStyle := inputBoxStyle
	googleLoginLink := lipgloss.NewStyle().Italic(true).Underline(true).Render("Press Ctrl+g to Login with")
	googleLogo := renderGoogleLogo()
	baseFooter := lipgloss.NewStyle().Italic(true).Bold(true).PaddingTop(1).Render("First time? ")
	underlineFooter := lipgloss.NewStyle().Italic(true).Underline(true).Render("Press Ctrl+s to go to SignUp Page")
	if m.Focused == 0 {
		userEmailStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#A78BFA"))
	}
	if m.Focused == 1 {
		passwordStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#A78BFA"))
	}

	form := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		userEmailStyle.Render(m.email.View()),
		passwordStyle.Render(m.password.View()),
		buttonStyle.Render("Press Enter to Submit"),
		" ",
		googleLoginLink+" "+googleLogo,
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
	baseFooter := lipgloss.NewStyle().Italic(true).Bold(true).Render("Already have an account? ")
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
		buttonStyle.Render("Press Enter to Submit"),
		googleSignupLink+" "+googleLogo,
		"",
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

func renderWelcomeMsg(m Model) string {

	baseStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E6E6E6")).PaddingTop(1).PaddingBottom(1)
	usernameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C63FF")).Bold(true)
	welcomeMsg := fmt.Sprintf("Welcome, %s💜", usernameStyle.Render(m.user.Username))
	styledMsg := baseStyle.Render(welcomeMsg)

	border := lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA")).Render(strings.Repeat("─", len(welcomeMsg)))

	return lipgloss.Place(
		m.width,
		3, // height: 2 for message + 1 for border
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, styledMsg, border),
	)
}

func renderJournal(m Model) string {
	currentTime := m.currentTime.Format("2006/01/02 15:04:05")
	clock := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#E0AfA0")).Render("🕰️ " + currentTime)

	header := lipgloss.Place(
		m.width,
		1,
		lipgloss.Center,
		lipgloss.Center,
		clock,
	)

	instructions := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("#A78BFA")).
		Render("\nCtrl+S to Save | Esc to Back | Ctrl+C to Quit")

	centeredInstructions := lipgloss.Place(
		m.width,
		1,
		lipgloss.Center,
		lipgloss.Center,
		instructions,
	)

	m.journal.SetWidth(m.width)
	m.journal.SetHeight(m.height - 6)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		centeredInstructions,
		"",
		"",
		m.journal.View(),
	)

	if m.err != nil {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			errorStyle.Render(fmt.Sprintf("Error: %v", m.err)),
		)
	}

	return content
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

	database := db.GetDB()

	defer db.CloseDB(database)

	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
	}
}
