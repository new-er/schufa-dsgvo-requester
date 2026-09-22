package tabs

import (
	"encoding/json"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/youruser/schufa-dsgvo-requester/internal/logger"
)

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type LogsTab struct {
	vp   viewport.Model
	width  int
	height int
}

func NewLogsTab() *LogsTab {
	vp := viewport.New(80, 24)
	return &LogsTab{vp: vp}
}

func (l *LogsTab) Init() tea.Cmd {
	l.refresh()
	return tick()
}

func (l *LogsTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height
		l.vp.Width = max(l.width-8, 40)
		l.vp.Height = max(l.height-10, 10)
		return l, nil
	case tea.KeyMsg:
		var cmd tea.Cmd
		l.vp, cmd = l.vp.Update(msg)
		cmds = append(cmds, cmd)
		return l, tea.Batch(cmds...)
	case tickMsg:
		l.refresh()
		cmds = append(cmds, tick())
		return l, tea.Batch(cmds...)
	}
	return l, tea.Batch(cmds...)
}

func (l *LogsTab) refresh() {
	lines := logger.GetLogs()
	if lines == nil {
		l.vp.SetContent("(no logs)")
		return
	}
	var formatted []string
	for _, line := range lines {
		formatted = append(formatted, formatLogLine(line))
	}
	l.vp.SetContent(strings.Join(formatted, "\n"))
	l.vp.GotoTop()
}

func formatLogLine(line string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(line), &m); err != nil {
		return line
	}
	timeStr := ""
	if t, ok := m["time"].(string); ok {
		if len(t) > 19 {
			timeStr = t[:19]
		} else {
			timeStr = t
		}
	}
	level := ""
	if l, ok := m["level"].(string); ok {
		level = l
	}
	msg := ""
	if msgVal, ok := m["msg"].(string); ok {
		msg = msgVal
	}
	event := ""
	if e, ok := m["event"].(string); ok {
		event = " [" + e + "]"
	}
	return timeStr + " " + level + event + " " + msg
}

func (l *LogsTab) View() string {
	return l.vp.View()
}