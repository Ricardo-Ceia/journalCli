package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
	senderStyle     lipgloss.Style
}

type NormalMsg struct {
	msg string
}

type ErrMsg struct {
	err error
}

const (
	PageLogin = iota
	PageMenu
	PageJournal
	PageRead
	PageSettings
	PageHelp
)

var style = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	PaddingTop(2).
	PaddingLeft(4).
	Width(22)

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

	return Model{
		page:        PageLogin,
		senderStyle: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")),
		username:    username,
		password:    password,
		inputing:    true,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func checkServer(userId string) tea.Msg {
	//Create an HTTP client and make a GET request to the server
	c := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/user?userId=%s", url, userId)
	f, _ := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	fmt.Fprintln(f, "Requesting URL:", url) // Debug print
	res, err := c.Get(url)

	if err != nil {
		return ErrMsg{err}
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)

	if err != nil {
		return ErrMsg{err}
	}

	if res.StatusCode != http.StatusOK {
		return ErrMsg{fmt.Errorf("server returned status code %d", res.StatusCode)}
	}

	return NormalMsg{string(b)}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NormalMsg:
		m.msg = msg.msg
		// After successful login, go to menu
		if m.page == PageLogin {
			m.page = PageMenu
			m.inputing = false
		}
	case ErrMsg:
		m.err = msg.err

	case tea.KeyMsg:
		switch m.page {
		// ------------- LOGIN PAGE -------------
		case PageLogin:
			if m.inputing {
				switch msg.Type {
				case tea.KeyEnter:
					m.inputing = false
					return m, func() tea.Msg { return checkServer(m.userId) }
				case tea.KeyBackspace:
					if len(m.userId) > 0 {
						m.userId = m.userId[:len(m.userId)-1]
					}
				case tea.KeyRunes:
					m.userId += string(msg.Runes)
				}
			}
		// ------------- MENU PAGE -------------
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

		// ------------- FEATURE A PAGE -------------
		case PageJournal:
			if msg.String() == "b" {
				m.page = PageMenu
			}

		// ------------- FEATURE B PAGE -------------
		case PageRead:
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
	return ""
}

func renderLoginPage(m Model) string {
	title := titleStyle.Render("🔐 Login/Signup")

	userNameInputBox := inputBoxStyle.Render(m.userId + "_")
	passwordInputBox := inputBoxStyle.Render("********")
	button := buttonStyle.Render("Press Enter to Submit")

	errMsg := ""
	if m.err != nil {
		errMsg = errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	return lipgloss.Place(
		50,
		10,
		lipgloss.Center,
		lipgloss.Center,
		fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s", title, userNameInputBox, passwordInputBox, button, errMsg),
	)
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
	}
}
