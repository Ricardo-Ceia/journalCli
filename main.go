package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"journalCli/utils"
	"log"
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
	senderStyle     lipgloss.Style
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

func checkServer(username, password string) tea.Msg {
	client := &http.Client{Timeout: 10 * time.Second}

	log.Printf("DEBUG: username=%q, password=%q\n", username, password)

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

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {

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
			case tea.KeyTab, tea.KeyDown:
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
				isValid, err := utils.ValidateCredentials(username, password)
				if !isValid {
					m.err = err
				} else {
					u, p := username, password
					return m, func() tea.Msg { return checkServer(u, p) }
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

	userNameStyle := inputBoxStyle
	passwordStyle := inputBoxStyle

	if m.Focused == 0 {
		userNameStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#FF9F1C"))
	}
	if m.Focused == 1 {
		passwordStyle = inputBoxStyle.BorderForeground(lipgloss.Color("#FF9F1C"))
	}

	form := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		userNameStyle.Render(m.username.View()),
		passwordStyle.Render(m.password.View()),
		buttonStyle.Render("Press Enter to Submit"),
	)

	if m.err != nil {
		form += "\n\n" + errorStyle.Render(m.err.Error())
	}

	return lipgloss.Place(
		80,
		20,
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
