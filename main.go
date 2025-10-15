package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var url string = "http://localhost:8080"

type model struct {
	msg      string
	err      error
	userId   string
	inputing bool
}

type NormalMsg struct {
	msg string
}

type errMsg struct {
	err error
}

func initialModel() model {
	return model{inputing: true}
}

func (m model) Init() tea.Cmd {
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
		return errMsg{err}
	}

	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)

	if err != nil {
		return errMsg{err}
	}

	if res.StatusCode != http.StatusOK {
		return errMsg{fmt.Errorf("server returned status code %d", res.StatusCode)}
	}

	return NormalMsg{string(b)}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case NormalMsg:
		m.msg = msg.msg
	case errMsg:
		m.err = msg.err
	case tea.KeyMsg:
		if m.inputing {
			switch msg.Type {
			case tea.KeyEnter:
				m.inputing = false
				return m, func() tea.Msg {
					return checkServer(m.userId)
				}
			case tea.KeyBackspace, tea.KeyDelete:
				if len(m.userId) > 0 {
					m.userId = m.userId[:len(m.userId)-1]
				}
			case tea.KeyRunes:
				m.userId += string(msg.Runes[0])
			}
			return m, nil
		}
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nWe had some trouble :%v\n\n", m.err)
	}

	if m.inputing {
		return fmt.Sprintf(
			"\nEnter userId and press Enter: %s\n\nPress Ctrl+C to quit.\n",
			m.userId)
	}

	s := fmt.Sprintf("\nReciving messages from the Server (%s)...", url)
	s += fmt.Sprintf("\n\nMessage: %s", m.msg)
	return "" + s + "\n\nPress Ctrl+C to quit.\n"
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
	}
}
