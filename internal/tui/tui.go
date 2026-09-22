package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/new-er/schufa-dsgvo-requester/internal/logger"
)

// Run starts the Bubble Tea program.
func Run() error {
	logDir := logger.DefaultLogDir()
	if err := logger.Init(logDir); err != nil {
		return err
	}
	logger.Log.Info("application start",
		"event", "app_start",
	)
	defer func() {
		logger.Log.Info("application stop",
			"event", "app_stop",
		)
		logger.Close()
	}()

	m, err := InitialModel()
	if err != nil {
		return err
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}