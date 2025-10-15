package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const url = "http://localhost:8080"

type model struct {
	msg string
	err error
}

type NormalMsg struct {
	msg string
}

type errMsg struct {
	err error
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return checkServer
}

func checkServer() tea.Msg {
	//Create an HTTP client and make a GET request to the server
	c := &http.Client{Timeout: 10 * time.Second}
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
