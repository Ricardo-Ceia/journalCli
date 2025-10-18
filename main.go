package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"journalCli/utils"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var url string = "http://localhost:8080"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			PaddingBottom(1)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Width(30)

	buttonStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 2).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F5F"))
)

type Page int

type Model struct {
	page            Page
	msg             string
	err             error
	userId          string
	inputing        bool
	textarea        textarea.Model
	username        textinput.Model
	password        textinput.Model
	confirmPassword textinput.Model
	Focused         int
	width           int
	height          int
	senderStyle     lipgloss.Style
	Client          *http.Client
}

type NormalMsg struct {
	msg string
}

type ErrMsg struct {
	err error
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SignupRequest struct {
	Username string `json:"username"`
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

func checkServerLogin(username, password string, client *http.Client) tea.Msg {
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	body, err := json.Marshal(loginReq)

	if err != nil {
		return ErrMsg{err}
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/auth", url), bytes.NewBuffer(body))

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

	return NormalMsg{string(b)}
}

func checkServerSignup(username, password string, client *http.Client) tea.Msg {
	signupReq := SignupRequest{
		Username: username,
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

	if res.StatusCode != http.StatusOK {
		return ErrMsg{fmt.Errorf("server returned status: %s, message: %s", res.Status, string(b))}
	}
	return NormalMsg{string(b)}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// ----------- SERVER RESPONSES -----------
	case NormalMsg:
		m.msg = msg.msg
		if m.page == PageLogin {
			m.page = PageMenu
			m.inputing = false
		}

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

			case tea.KeyTab, tea.KeyDown, tea.KeyUp:
				m.Focused = (m.Focused + 1) % 3
				switch m.Focused {
				case 0:
					m.username.Focus()
					m.password.Blur()
					m.confirmPassword.Blur()
				case 1:
					m.password.Focus()
					m.username.Blur()
					m.confirmPassword.Blur()
				case 2:
					m.confirmPassword.Focus()
					m.username.Blur()
					m.password.Blur()
				}
			case tea.KeyEnter:
				username := m.username.Value()
				password := m.password.Value()
				confirmPassword := m.confirmPassword.Value()
				isValid, err := utils.ValidateCredentials(username, password, confirmPassword)
				if !isValid {
					m.err = err
				} else {
					return m, func() tea.Msg { return checkServerSignup(username, password, m.Client) }
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
		return "Menu Page\n\n1. Journal\n2. Read\n3. Settings\n4. Help\nq. Quit"
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
		userNameStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#FF9F1C"))
	}
	if m.Focused == 1 {
		passwordStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#FF9F1C"))
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
	passwordStyle := inputBoxStyle
	confirmPasswordStyle := inputBoxStyle

	if m.Focused == 0 {
		userNameStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#FF9F1C"))
	}
	if m.Focused == 1 {
		passwordStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#FF9F1C"))
	}
	if m.Focused == 2 {
		confirmPasswordStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#FF9F1C"))
	}

	form := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		userNameStyle.Render(m.username.View()),
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
		form,
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
	}
}
