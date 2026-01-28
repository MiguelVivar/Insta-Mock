// Package tui provides a terminal dashboard for Insta-Mock.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("51")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	methodStyles = map[string]lipgloss.Style{
		"GET":    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")),  // Green
		"POST":   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214")), // Orange
		"PUT":    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")),  // Blue
		"PATCH":  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("141")), // Purple
		"DELETE": lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")), // Red
	}

	statusStyles = map[int]lipgloss.Style{
		2: lipgloss.NewStyle().Foreground(lipgloss.Color("42")),  // 2xx Green
		3: lipgloss.NewStyle().Foreground(lipgloss.Color("214")), // 3xx Yellow
		4: lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // 4xx Red
		5: lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // 5xx Red
	}

	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	borderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("236"))
)

// RequestLog represents a single logged request.
type RequestLog struct {
	Time       time.Time
	Method     string
	Path       string
	StatusCode int
	Latency    time.Duration
}

// Stats holds request statistics.
type Stats struct {
	TotalRequests int
	GetCount      int
	PostCount     int
	PutCount      int
	DeleteCount   int
	ErrorCount    int
}

// Model is the Bubbletea model for the dashboard.
type Model struct {
	viewport viewport.Model
	logs     []RequestLog
	stats    Stats
	port     string
	width    int
	height   int
	ready    bool
}

// NewModel creates a new dashboard model.
func NewModel(port string) Model {
	return Model{
		logs: make([]RequestLog, 0),
		port: port,
	}
}

// Init initializes the model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "c":
			m.logs = make([]RequestLog, 0)
			m.stats = Stats{}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-8)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 8
		}

	case RequestLog:
		m.logs = append(m.logs, msg)
		m.stats.TotalRequests++
		switch msg.Method {
		case "GET":
			m.stats.GetCount++
		case "POST":
			m.stats.PostCount++
		case "PUT", "PATCH":
			m.stats.PutCount++
		case "DELETE":
			m.stats.DeleteCount++
		}
		if msg.StatusCode >= 400 {
			m.stats.ErrorCount++
		}
		m.viewport.SetContent(m.renderLogs())
		m.viewport.GotoBottom()
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the dashboard.
func (m Model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	var b strings.Builder

	// Header
	header := titleStyle.Render("🚀 Insta-Mock Dashboard")
	b.WriteString(header)
	b.WriteString(infoStyle.Render(fmt.Sprintf("  http://localhost:%s", m.port)))
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")

	// Stats bar
	stats := fmt.Sprintf(
		"  📊 Total: %s  │  GET: %s  POST: %s  PUT: %s  DEL: %s  │  Errors: %s",
		lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", m.stats.TotalRequests)),
		methodStyles["GET"].Render(fmt.Sprintf("%d", m.stats.GetCount)),
		methodStyles["POST"].Render(fmt.Sprintf("%d", m.stats.PostCount)),
		methodStyles["PUT"].Render(fmt.Sprintf("%d", m.stats.PutCount)),
		methodStyles["DELETE"].Render(fmt.Sprintf("%d", m.stats.DeleteCount)),
		statusStyles[4].Render(fmt.Sprintf("%d", m.stats.ErrorCount)),
	)
	b.WriteString(stats)
	b.WriteString("\n")
	b.WriteString(borderStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")

	// Logs viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Footer
	b.WriteString(borderStyle.Render(strings.Repeat("─", m.width)))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  q: quit  │  c: clear logs  │  ↑/↓: scroll"))

	return b.String()
}

// renderLogs formats all logs for display.
func (m Model) renderLogs() string {
	var b strings.Builder

	for _, log := range m.logs {
		// Time
		timeStr := dimStyle.Render(log.Time.Format("15:04:05"))

		// Method with color
		methodStyle := methodStyles["GET"]
		if s, ok := methodStyles[log.Method]; ok {
			methodStyle = s
		}
		method := methodStyle.Render(fmt.Sprintf("%-6s", log.Method))

		// Status with color
		statusCategory := log.StatusCode / 100
		statusStyle := statusStyles[statusCategory]
		if statusStyle.Value() == "" {
			statusStyle = infoStyle
		}
		status := statusStyle.Render(fmt.Sprintf("%d", log.StatusCode))

		// Latency
		latency := dimStyle.Render(fmt.Sprintf("%6s", log.Latency.Round(time.Millisecond)))

		// Path
		path := infoStyle.Render(log.Path)

		b.WriteString(fmt.Sprintf("  %s │ %s │ %s │ %s │ %s\n", timeStr, method, status, latency, path))
	}

	return b.String()
}

// AddLog adds a new log entry to the dashboard.
func (m *Model) AddLog(log RequestLog) {
	m.logs = append(m.logs, log)
}
