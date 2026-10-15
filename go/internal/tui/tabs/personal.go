package tabs

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/new-er/schufa-dsgvo-requester/internal/config"
	"github.com/new-er/schufa-dsgvo-requester/internal/logger"
)

// fieldDef describes an editable personal field.
type fieldDef struct {
	label string
	get   func() string
	set   func(string)
}

type PersonalTab struct {
	cfg        *config.Config
	configPath string
	fields     []fieldDef
	focusIdx   int
	editMode   bool
	ti         textinput.Model
	origVal    string
	width      int
}

func NewPersonalTab(cfg *config.Config, configPath string) *PersonalTab {
	fields := []fieldDef{
		{"Vorname", func() string { return cfg.Personal.Vorname }, func(v string) { cfg.Personal.Vorname = v }},
		{"Nachname", func() string { return cfg.Personal.Nachname }, func(v string) { cfg.Personal.Nachname = v }},
		{"Geburtsdatum", func() string { return cfg.Personal.Geburtsdatum }, func(v string) { cfg.Personal.Geburtsdatum = v }},
		{"Straße", func() string { return cfg.Personal.Strasse }, func(v string) { cfg.Personal.Strasse = v }},
		{"PLZ", func() string { return cfg.Personal.PLZ }, func(v string) { cfg.Personal.PLZ = v }},
		{"Wohnort", func() string { return cfg.Personal.Wohnort }, func(v string) { cfg.Personal.Wohnort = v }},
		{"Land", func() string { return cfg.Personal.Land }, func(v string) { cfg.Personal.Land = v }},
	}
	ti := textinput.New()
	ti.Prompt = ""
	ti.Blur()
	return &PersonalTab{
		cfg:        cfg,
		configPath: configPath,
		fields:     fields,
		ti:         ti,
	}
}

func (p *PersonalTab) Init() tea.Cmd {
	return nil
}

func (p *PersonalTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		// set textinput width to content width minus some margin
		p.ti.Width = max(p.width-8, 20)
		return p, nil

	case tea.KeyMsg:
		if p.editMode {
			switch msg.String() {
			case "enter":
				// save
				p.fields[p.focusIdx].set(p.ti.Value())
				// persist config
				if err := config.Save(p.configPath, p.cfg); err != nil {
					// could set an error status; for now ignore
				} else {
					logger.Log.Info("config saved",
						"event", "config_saved",
						"section", "personal",
					)
				}
				p.ti.Blur()
				p.editMode = false
				return p, nil
			case "esc":
				// discard
				p.ti.SetValue(p.origVal)
				p.ti.Blur()
				p.editMode = false
				return p, nil
			default:
				var cmd tea.Cmd
				p.ti, cmd = p.ti.Update(msg)
				cmds = append(cmds, cmd)
				return p, tea.Batch(cmds...)
			}
		} else {
			switch msg.String() {
			case "up", "k":
				p.focusIdx = (p.focusIdx - 1 + len(p.fields)) % len(p.fields)
			case "down", "j":
				p.focusIdx = (p.focusIdx + 1) % len(p.fields)
			case "enter":
				// start editing
				p.origVal = p.fields[p.focusIdx].get()
				p.ti.SetValue(p.origVal)
				p.ti.Focus()
				p.editMode = true
				return p, nil
			}
		}
	}

	return p, tea.Batch(cmds...)
}

func (p *PersonalTab) Editing() bool {
	return p.editMode
}

func (p *PersonalTab) View() string {
	// use highlight color from styles (mirrored here)
	highlight := lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}

	var lines []string
	for i, f := range p.fields {
		val := f.get()
		line := fmt.Sprintf("%s: %s", f.label, val)
		style := lipgloss.NewStyle()
		if i == p.focusIdx && !p.editMode {
			style = style.Foreground(highlight).Bold(true)
		}
		if p.editMode && i == p.focusIdx {
			// show textinput
			line = fmt.Sprintf("%s: %s", f.label, p.ti.View())
			style = style.Foreground(highlight).Bold(true)
		}
		lines = append(lines, style.Render(line))
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}