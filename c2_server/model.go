package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wrap"
	"strconv"
	"strings"
)

type (
	errMsg error
)

type Responses struct {
	current string
	rest    chan string
}

func (r Responses) awaitNext() Responses {
	return Responses{current: <-r.rest, rest: r.rest}
}

var ResultChannel = make(chan string)

type Model struct {
	currentResponse string
	viewport        viewport.Model
	ready           bool
	history         []string
	header          string
	textarea        textarea.Model
	commandStyle    lipgloss.Style
	outputStyle     lipgloss.Style
	infoStyle       lipgloss.Style
	successStyle    lipgloss.Style
	err             error
}

func InitialModel() Model {
	header := "C2 server console"
	ta := textarea.New()
	ta.Placeholder = "Send a command..."
	ta.Focus()
	ta.Prompt = "👉 "
	ta.CharLimit = 280
	ta.SetWidth(150)
	ta.SetHeight(1)
	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Prompt.Foreground(lipgloss.Color("#20C20E"))
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	return Model{
		header:       header,
		textarea:     ta,
		commandStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#20C20E")).Bold(true),
		outputStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Italic(true),
		infoStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#8C92AC")).Italic(true),
		successStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#20C20E")).Bold(true),
		err:          nil,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(func() tea.Msg {
		return Responses{<-ResultChannel, ResultChannel}
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)
	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case Responses:
		m.currentResponse = msg.current
		resp := m.currentResponse
		switch {
		case resp == "beacon":
			if !strings.Contains(m.header, "Online") {
				m.header += m.infoStyle.Render(" | Status: ") + m.successStyle.Render("Online")
			}
		case strings.HasPrefix(resp, "initC2_"):
			hostname := strings.TrimPrefix(m.currentResponse, "initC2_")
			m.header = "Controlling impl@" + hostname
			m.history = append(m.history, wrap.String(m.outputStyle.Render("Received beacon from impl@"+hostname), m.viewport.Width))
		case strings.HasPrefix(resp, "sendLong_"):
			size, _ := strconv.Atoi(strings.TrimPrefix(resp, "sendLong_"))
			m.history = append(m.history, wrap.String(m.infoStyle.Render("Long payload detected: "+strconv.Itoa(size*48)+" bytes"), m.viewport.Width))
		default:
			m.history = append(m.history, wrap.String(m.outputStyle.Render("(impl_output)> ")+m.currentResponse, m.viewport.Width))
		}

		m.viewport.SetContent(strings.Join(m.history, "\n"))
		m.viewport.GotoBottom()
		return m, func() tea.Msg { return msg.awaitNext() }
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			m.history = append(m.history, wrap.String(m.commandStyle.Render("(command)> ")+m.textarea.Value(), m.viewport.Width))
			RequestChannel <- m.textarea.Value()
			m.viewport.SetContent(strings.Join(m.history, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView(m.header))
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight + 2
		if !m.ready {
			m.viewport = viewport.New(msg.Width-2, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = false
			m.viewport.SetContent(wrap.String(`Welcome to the C2 server console!
Type a command and press Enter to send.`, m.viewport.Width))
			m.viewport.SetContent(strings.Join(m.history, "\n"))
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}
