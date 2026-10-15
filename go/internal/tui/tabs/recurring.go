package tabs

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/new-er/schufa-dsgvo-requester/internal/config"
	"github.com/new-er/schufa-dsgvo-requester/internal/logger"
	"github.com/new-er/schufa-dsgvo-requester/internal/scheduler"
)

// RecurringTab allows setting interval months and installing/uninstalling systemd timer.
type RecurringTab struct {
	cfg        *config.Config
	configPath string
	binaryPath string

	focusIdx int       // 0 = interval, 1 = install/uninstall
	editMode bool
	ti       textinput.Model
	origVal  string
	status   string
	width    int
}

func NewRecurringTab(cfg *config.Config, binaryPath, configPath string) *RecurringTab {
	ti := textinput.New()
	ti.Prompt = ""
	ti.Blur()
	return &RecurringTab{
		cfg:        cfg,
		configPath: configPath,
		binaryPath: binaryPath,
		ti:         ti,
	}
}

func (r *RecurringTab) Init() tea.Cmd {
	return nil
}

func (r *RecurringTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		r.width = msg.Width
		r.ti.Width = max(r.width-8, 20)
		return r, nil

	case tea.KeyMsg:
		if r.editMode {
			switch msg.String() {
			case "enter":
				// save interval
				val := r.ti.Value()
				if n, err := strconv.Atoi(val); err == nil && n > 0 {
					r.cfg.Schedule.IntervalMonths = n
					if err := config.Save(r.configPath, r.cfg); err == nil {
						logger.Log.Info("config saved",
							"event", "config_saved",
							"section", "schedule",
						)
					}
				}
				r.ti.Blur()
				r.editMode = false
				return r, nil
			case "esc":
				r.ti.SetValue(r.origVal)
				r.ti.Blur()
				r.editMode = false
				return r, nil
			default:
				var cmd tea.Cmd
				r.ti, cmd = r.ti.Update(msg)
				cmds = append(cmds, cmd)
				return r, tea.Batch(cmds...)
			}
		} else {
			switch msg.String() {
			case "up", "k":
				r.focusIdx = (r.focusIdx - 1 + 2) % 2
			case "down", "j":
				r.focusIdx = (r.focusIdx + 1) % 2
			case "enter":
				if r.focusIdx == 0 {
					// edit interval
					r.origVal = strconv.Itoa(r.cfg.Schedule.IntervalMonths)
					r.ti.SetValue(r.origVal)
					r.ti.Focus()
					r.editMode = true
					return r, nil
				} else if r.focusIdx == 1 {
					// install or uninstall via CLI subcommand
					var subcmd string
					if scheduler.IsSystemdInstalled() {
						subcmd = "uninstall-systemd"
					} else {
						subcmd = "install-systemd"
					}
					cmd := exec.Command(os.Args[0], subcmd)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err := cmd.Run()
					if err != nil {
						r.status = "Error: " + err.Error()
					} else {
						if subcmd == "install-systemd" {
							r.status = "✔ Installed"
						} else {
							r.status = "✔ Uninstalled"
						}
					}
					return r, nil
				}
			}
		}
	}

	return r, tea.Batch(cmds...)
}

func (r *RecurringTab) View() string {
	highlight := lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	green := lipgloss.Color("42")
	redc := lipgloss.Color("196")

	// Last sent
	lastSent := r.cfg.Schedule.LastSent
	if lastSent == "" {
		lastSent = "Never"
	} else {
		if t, err := time.Parse(time.RFC3339, lastSent); err == nil {
			lastSent = t.Format("2006-01-02 15:04")
		}
	}

	// Next send
	nextStr := "–"
	if r.cfg.Schedule.IntervalMonths > 0 {
		next, err := scheduler.NextRun(r.cfg, time.Now())
		if err == nil {
			nextStr = next.Format("2006-01-02 15:04")
		}
	}

	var lines []string
	// Last sent line
	lines = append(lines, fmt.Sprintf("Last sent: %s", lastSent))
	// Next send line
	lines = append(lines, fmt.Sprintf("Next send: %s", nextStr))
	lines = append(lines, "")

	// Interval line
	intervalLabel := fmt.Sprintf("Interval months: %d", r.cfg.Schedule.IntervalMonths)
	if r.editMode {
		intervalLabel = fmt.Sprintf("Interval months: %s", r.ti.View())
	}
	style := lipgloss.NewStyle()
	if r.focusIdx == 0 && !r.editMode {
		style = style.Foreground(highlight).Bold(true)
	} else if r.editMode {
		style = style.Foreground(highlight).Bold(true)
	}
	lines = append(lines, style.Render(intervalLabel))

	// Install/Uninstall button line
	btnLabel := "Install"
	if scheduler.IsSystemdInstalled() {
		btnLabel = "Uninstall"
	}
	btnLine := fmt.Sprintf("[ %s ]", btnLabel)
	if r.focusIdx == 1 && !r.editMode {
		btnLine = "> " + btnLine + " <"
	}
	btnStyle := lipgloss.NewStyle().Bold(true)
	if r.focusIdx == 1 && !r.editMode {
		btnStyle = btnStyle.Foreground(highlight)
	}
	lines = append(lines, btnStyle.Render(btnLine))

	// Status line
	if r.status != "" {
		st := r.status
		if strings.HasPrefix(st, "✔") {
			st = lipgloss.NewStyle().Foreground(green).Render(st)
		} else if strings.HasPrefix(st, "✘") {
			st = lipgloss.NewStyle().Foreground(redc).Render(st)
		}
		lines = append(lines, "", st)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}