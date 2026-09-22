package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/new-er/schufa-dsgvo-requester/internal/config"
	"github.com/new-er/schufa-dsgvo-requester/internal/mailer"
	"github.com/new-er/schufa-dsgvo-requester/internal/scheduler"
	"github.com/new-er/schufa-dsgvo-requester/internal/template"
	"github.com/new-er/schufa-dsgvo-requester/internal/tui/tabs"
)

type Model struct {
	activeTab   int
	tabs        []string
	sendTab     *tabs.SendTab
	personalTab *tabs.PersonalTab
	recurringTab *tabs.RecurringTab
	messageTab  *tabs.MessageTab
	logsTab     *tabs.LogsTab
	config      *config.Config
	configPath  string
	binaryPath  string
	width       int
}

func InitialModel() (*Model, error) {
	configPath := "config.toml"
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}
	binPath, _ := os.Executable()
	binPath, _ = filepath.Abs(binPath)

	sendTab := tabs.NewSendTab(cfg)
	personalTab := tabs.NewPersonalTab(cfg, configPath)
	recurringTab := tabs.NewRecurringTab(cfg, binPath, configPath)
	messageTab := tabs.NewMessageTab(cfg, configPath)
	logsTab := tabs.NewLogsTab()

	m := &Model{
		activeTab:   0,
		tabs:        []string{"Send", "Personal", "Recurring", "Message", "Logs"},
		sendTab:     sendTab,
		personalTab: personalTab,
		recurringTab: recurringTab,
		messageTab:  messageTab,
		logsTab:     logsTab,
		config:      cfg,
		configPath:  configPath,
		binaryPath:  binPath,
	}
	return m, nil
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.sendTab.Init(),
		m.personalTab.Init(),
		m.recurringTab.Init(),
		m.messageTab.Init(),
		m.logsTab.Init(),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		// Forward to all tabs so inactive tabs get sized too
		for _, tab := range []interface{}{m.sendTab, m.personalTab, m.recurringTab, m.messageTab, m.logsTab} {
			if updatable, ok := tab.(interface{ Update(tea.Msg) (tea.Model, tea.Cmd) }); ok {
				if _, cmd := updatable.Update(msg); cmd != nil {
					cmds = append(cmds, cmd)
				}
			}
		}
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
case "tab", "right":
		if !(m.activeTab == 1 && m.personalTab.Editing()) && !(m.activeTab == 3 && m.messageTab.Editing()) {
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			return m, nil
		}
	case "shift+tab", "left":
		if !(m.activeTab == 1 && m.personalTab.Editing()) && !(m.activeTab == 3 && m.messageTab.Editing()) {
			m.activeTab = (m.activeTab - 1 + len(m.tabs)) % len(m.tabs)
			return m, nil
		}
		}
	}

	// Delegate to active tab
	switch m.activeTab {
	case 0:
		newTab, cmd := m.sendTab.Update(msg)
		m.sendTab = newTab.(*tabs.SendTab)
		cmds = append(cmds, cmd)
	case 1:
		newTab, cmd := m.personalTab.Update(msg)
		m.personalTab = newTab.(*tabs.PersonalTab)
		cmds = append(cmds, cmd)
	case 2:
		newTab, cmd := m.recurringTab.Update(msg)
		m.recurringTab = newTab.(*tabs.RecurringTab)
		cmds = append(cmds, cmd)
	case 3:
		newTab, cmd := m.messageTab.Update(msg)
		m.messageTab = newTab.(*tabs.MessageTab)
		cmds = append(cmds, cmd)
	case 4:
		newTab, cmd := m.logsTab.Update(msg)
		m.logsTab = newTab.(*tabs.LogsTab)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	var content string
	switch m.activeTab {
	case 0:
		content = m.sendTab.View()
	case 1:
		content = m.personalTab.View()
	case 2:
		content = m.recurringTab.View()
	case 3:
		content = m.messageTab.View()
	case 4:
		content = m.logsTab.View()
	}
	// render tab bar
	var renderedTabs []string
	for i, name := range m.tabs {
		style := TabStyle
		if i == m.activeTab {
			style = ActiveTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(name))
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	// content pane with border
	contentPane := ContentStyle.Width(max(m.width-4, 20)).Render(content)

	// combine with a separator line
	sep := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#CCCCCC", Dark: "#444444"}).Render(strings.Repeat("─", max(m.width, 20)))
	return fmt.Sprintf("%s\n%s\n%s", tabBar, sep, contentPane)
}

// Helper functions used by tabs
func sendMail(cfg *config.Config, configPath string, dryRun bool) (string, error) {
	body, err := template.Render(cfg, configPath)
	if err != nil {
		return "", err
	}
	if dryRun {
		return body, nil
	}
	if err := mailer.Send(cfg, "DSGVO Auskunftsersuchen", body); err != nil {
		return "", err
	}
	return "Mail sent successfully", nil
}

func saveConfig(cfg *config.Config, path string) error {
	return config.Save(path, cfg)
}

func generateSystemd(cfg *config.Config, bin, conf, outDir, onCal string) error {
	return scheduler.GenerateSystemd(cfg, bin, conf, outDir, onCal)
}

func installSystemd(cfg *config.Config, bin, conf string) error {
	return scheduler.InstallSystemd(cfg, bin, conf)
}
