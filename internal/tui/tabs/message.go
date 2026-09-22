package tabs

import (
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/youruser/schufa-dsgvo-requester/internal/config"
	"github.com/youruser/schufa-dsgvo-requester/internal/logger"
	"github.com/youruser/schufa-dsgvo-requester/internal/template"
)

type MessageTab struct {
	cfg        *config.Config
	configPath string
	ti         textarea.Model
	vp         viewport.Model
	editMode   bool
	width      int
	height     int
	status     string
}

func NewMessageTab(cfg *config.Config, configPath string) *MessageTab {
	ti := textarea.New()
	ti.Prompt = ""
	ti.ShowLineNumbers = false
	ti.Blur()
	ti.Cursor.Style = lipgloss.NewStyle()
	ti.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ti.FocusedStyle.Base = lipgloss.NewStyle()
	ti.CharLimit = 0
	vp := viewport.New(80, 24)
	return &MessageTab{
		cfg:        cfg,
		configPath: configPath,
		ti:         ti,
		vp:         vp,
	}
}

func (m *MessageTab) Editing() bool {
	return m.editMode
}

func (m *MessageTab) Init() tea.Cmd {
	// load raw template (custom or embedded) without rendering placeholders
	body, err := template.Raw(m.configPath)
	if err != nil {
		body = "Error loading template: " + err.Error()
	}
	m.ti.SetValue(body)
	return nil
}

func (m *MessageTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ti.SetWidth(max(m.width-8, 40))
		m.ti.SetHeight(max(m.height-10, 12))
		m.vp.Width = max(m.width-8, 40)
		m.vp.Height = max(m.height-10, 12)
		return m, nil

	case tea.KeyMsg:
		if m.editMode {
			switch msg.String() {
			case "esc":
				// save to mail.custom.txt next to config
				customPath := filepath.Join(filepath.Dir(m.configPath), "mail.custom.txt")
				if err := os.WriteFile(customPath, []byte(m.ti.Value()), 0644); err != nil {
					m.status = "Save error: " + err.Error()
				} else {
					m.status = "✔ Saved"
					logger.Log.Info("config saved",
						"event", "config_saved",
						"section", "template",
					)
				}
				m.ti.Blur()
				m.editMode = false
				return m, nil
			default:
				var cmd tea.Cmd
				m.ti, cmd = m.ti.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		} else {
			switch msg.String() {
			case "enter":
				m.editMode = true
				m.ti.Focus()
				return m, nil
			case "up", "k":
				m.vp.LineUp(1)
				return m, nil
			case "down", "j":
				m.vp.LineDown(1)
				return m, nil
			case "pgup":
				m.vp.HalfViewUp()
				return m, nil
			case "pgdown":
				m.vp.HalfViewDown()
				return m, nil
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *MessageTab) View() string {
	if m.editMode {
		style := lipgloss.NewStyle().
			Padding(0, 1)
		return style.Render(m.ti.View())
	}

	// read‑only preview - show all lines via viewport
	preview := m.ti.Value()
	m.vp.SetContent(preview)
	m.vp.GotoTop()
	
	box := lipgloss.NewStyle().
		Padding(0, 1).
		Width(max(m.width-8, 40)).
		Render(m.vp.View())
	hint := lipgloss.NewStyle().Faint(true).Render("Press Enter to edit")
	if m.status != "" {
		hint += "  " + m.status
	}
	return lipgloss.JoinVertical(lipgloss.Left, box, hint)
}
