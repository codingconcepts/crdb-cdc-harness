package views

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/codingconcepts/crdb-cdc-harness/pkg/models"
	"github.com/codingconcepts/crdb-cdc-harness/pkg/styles"
)

// Model holds the configuration for the terminal UI.
type Model struct {
	countWritten uint64
	countUnread  int
	avgLatency   time.Duration
	dbWriting    bool
	statusChan   chan bool
	messages     chan models.KafkaMessage
	spinner      spinner.Model
	help         help.Model
	keys         models.KeyMap
	width        int
	height       int
}

// NewModel initializes and returns a new instance of Model.
func NewModel(messages chan models.KafkaMessage, statusChan chan bool) Model {
	s := spinner.New()
	s.Spinner = spinner.Points
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(styles.WritingAreadActiveColour))

	return Model{
		avgLatency: time.Duration(0),
		dbWriting:  true,
		statusChan: statusChan,
		messages:   messages,
		spinner:    s,
		help:       help.New(),
		keys:       models.Keys,
		width:      80,
		height:     24,
	}
}

// UpdateStatsMsg holds the variables published during status updates.
type UpdateStatsMsg struct {
	CountUnread  int
	CountWritten uint64
}

// LatencyUpdateMsgs holds the variables published during latency updates.
type LatencyUpdateMsg struct {
	AvgLatency time.Duration
}

func (m Model) View() string {
	var statusContent string
	var borderColor string
	if m.dbWriting {
		statusContent = m.spinner.View() + styles.ActiveStyle.Render(" WRITING ACTIVE")
		borderColor = styles.WritingBorderActiveColour
	} else {
		statusContent = styles.InactiveStyle.Render("⏸ WRITING PAUSED")
		borderColor = styles.WritingBorderInactiveColour
	}
	status := styles.BoxStyle.
		BorderForeground(lipgloss.Color(borderColor)).
		Render(statusContent)

	statsContent := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.StatsLabelStyle.Render("Written: "),
			styles.StatsValueStyle.Render(fmt.Sprintf("%04d", m.countWritten))),
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.StatsLabelStyle.Render("Unread:  "),
			styles.StatsValueStyle.Render(fmt.Sprintf("%04d", m.countUnread))),
		lipgloss.JoinHorizontal(lipgloss.Left,
			styles.StatsLabelStyle.Render("Latency: "),
			styles.StatsValueStyle.Render(m.avgLatency.String())),
	)
	stats := styles.BoxStyle.Render(statsContent)

	help := styles.BoxStyle.
		BorderForeground(lipgloss.Color(styles.HelpAreaBorderColour)).
		Render(m.help.View(m.keys))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		status,
		stats,
		help,
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Toggle):
			m.dbWriting = !m.dbWriting
			m.statusChan <- m.dbWriting
			return m, nil
		}

	case UpdateStatsMsg:
		m.countUnread = msg.CountUnread
		m.countWritten = msg.CountWritten

	case LatencyUpdateMsg:
		m.avgLatency = msg.AvgLatency

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		width := min(msg.Width-4, 80)
		styles.BoxStyle = styles.BoxStyle.Width(width)
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		tea.Every(time.Second, func(time.Time) tea.Msg {
			return UpdateStatsMsg{
				CountUnread: m.countUnread,
			}
		}),
	)
}
