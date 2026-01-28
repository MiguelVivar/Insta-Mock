// Package tui provides a terminal dashboard for Insta-Mock.
package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Base styles
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("51")).
			Padding(0, 1)

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Padding(1, 0)

	// Method colors
	methodColors = map[string]lipgloss.Color{
		"GET":    lipgloss.Color("42"),  // Green
		"POST":   lipgloss.Color("214"), // Orange
		"PUT":    lipgloss.Color("39"),  // Blue
		"PATCH":  lipgloss.Color("141"), // Purple
		"DELETE": lipgloss.Color("196"), // Red
	}

	// Status colors
	statusColors = map[int]lipgloss.Color{
		2: lipgloss.Color("42"),  // 2xx Green
		3: lipgloss.Color("214"), // 3xx Yellow
		4: lipgloss.Color("196"), // 4xx Red
		5: lipgloss.Color("196"), // 5xx Red
	}
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
	PatchCount    int
	DeleteCount   int
	ErrorCount    int
}

// Model is the Bubbletea model for the dashboard.
type Model struct {
	table    table.Model
	rows     []table.Row
	stats    Stats
	port     string
	width    int
	height   int
	quitting bool
}

// NewModel creates a new dashboard model.
func NewModel(port string) Model {
	columns := []table.Column{
		{Title: "Time", Width: 10},
		{Title: "Method", Width: 8},
		{Title: "Path", Width: 30},
		{Title: "Status", Width: 8},
		{Title: "Latency", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(15),
	)

	// Table styling
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("229"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	s.Cell = s.Cell.Padding(0, 1)

	t.SetStyles(s)

	return Model{
		table: t,
		rows:  make([]table.Row, 0),
		port:  port,
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
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "c":
			// Clear logs
			m.rows = make([]table.Row, 0)
			m.table.SetRows(m.rows)
			m.stats = Stats{}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Adjust table height to fit window
		tableHeight := msg.Height - 10
		if tableHeight < 5 {
			tableHeight = 5
		}
		m.table.SetHeight(tableHeight)
		m.table.SetWidth(msg.Width - 4)

	case RequestLog:
		// Update stats
		m.stats.TotalRequests++
		switch msg.Method {
		case "GET":
			m.stats.GetCount++
		case "POST":
			m.stats.PostCount++
		case "PUT":
			m.stats.PutCount++
		case "PATCH":
			m.stats.PatchCount++
		case "DELETE":
			m.stats.DeleteCount++
		}
		if msg.StatusCode >= 400 {
			m.stats.ErrorCount++
		}

		// Create styled row
		row := m.createRow(msg)
		m.rows = append(m.rows, row)

		// Keep only last 100 rows to prevent memory issues
		if len(m.rows) > 100 {
			m.rows = m.rows[len(m.rows)-100:]
		}

		m.table.SetRows(m.rows)
		// Auto-scroll to bottom
		m.table.GotoBottom()
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// createRow creates a styled table row from a request log.
func (m Model) createRow(log RequestLog) table.Row {
	// Format time
	timeStr := log.Time.Format("15:04:05")

	// Style method with color
	methodStyle := lipgloss.NewStyle()
	if color, ok := methodColors[log.Method]; ok {
		methodStyle = methodStyle.Foreground(color).Bold(true)
	}
	method := methodStyle.Render(log.Method)

	// Style status with color
	statusCategory := log.StatusCode / 100
	statusStyle := lipgloss.NewStyle()
	if color, ok := statusColors[statusCategory]; ok {
		statusStyle = statusStyle.Foreground(color)
	}
	status := statusStyle.Render(fmt.Sprintf("%d", log.StatusCode))

	// Format latency
	latency := log.Latency.Round(time.Microsecond).String()

	return table.Row{timeStr, method, log.Path, status, latency}
}

// View renders the dashboard.
func (m Model) View() string {
	if m.quitting {
		return "👋 Goodbye!\n"
	}

	// Header
	header := titleStyle.Render("🚀 Insta-Mock Dashboard")
	serverInfo := statsStyle.Render(fmt.Sprintf("http://localhost:%s", m.port))

	// Stats bar
	statsBar := m.renderStats()

	// Table
	tableView := baseStyle.Render(m.table.View())

	// Help
	help := helpStyle.Render("↑/↓: scroll • c: clear • q: quit")

	return fmt.Sprintf(
		"%s  %s\n\n%s\n\n%s\n\n%s",
		header,
		serverInfo,
		statsBar,
		tableView,
		help,
	)
}

// renderStats renders the stats bar.
func (m Model) renderStats() string {
	total := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", m.stats.TotalRequests))

	get := lipgloss.NewStyle().Foreground(methodColors["GET"]).Render(fmt.Sprintf("%d", m.stats.GetCount))
	post := lipgloss.NewStyle().Foreground(methodColors["POST"]).Render(fmt.Sprintf("%d", m.stats.PostCount))
	put := lipgloss.NewStyle().Foreground(methodColors["PUT"]).Render(fmt.Sprintf("%d", m.stats.PutCount))
	patch := lipgloss.NewStyle().Foreground(methodColors["PATCH"]).Render(fmt.Sprintf("%d", m.stats.PatchCount))
	del := lipgloss.NewStyle().Foreground(methodColors["DELETE"]).Render(fmt.Sprintf("%d", m.stats.DeleteCount))

	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	errors := errorStyle.Render(fmt.Sprintf("%d", m.stats.ErrorCount))

	return fmt.Sprintf(
		"📊 Total: %s  │  GET: %s  POST: %s  PUT: %s  PATCH: %s  DEL: %s  │  ❌ Errors: %s",
		total, get, post, put, patch, del, errors,
	)
}

// SendLog is a helper to send a log message to the model.
func SendLog(method, path string, status int, latency time.Duration) tea.Msg {
	return RequestLog{
		Time:       time.Now(),
		Method:     method,
		Path:       path,
		StatusCode: status,
		Latency:    latency,
	}
}
