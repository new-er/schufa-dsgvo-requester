package tabs

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/new-er/schufa-dsgvo-requester/internal/config"
	"github.com/new-er/schufa-dsgvo-requester/internal/logger"
	"github.com/new-er/schufa-dsgvo-requester/internal/mailer"
	"github.com/new-er/schufa-dsgvo-requester/internal/template"
)

type sendFocus int

const (
	focusList sendFocus = iota
	focusFull
)

type sendResultMsg struct {
	msg string
	err error
}

type spinnerTickMsg struct{}

func spinnerTick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

// local color palette (mirrors internal/tui/styles)
var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	green     = lipgloss.Color("42")
	red       = lipgloss.Color("196")
	spinnerFrames = []string{"⠋","⠙","⠹","⠸","⠼","⠴","⠦","⠧","⠇","⠏"}
)

type SendTab struct {
	cfg        *config.Config
	smallVP    viewport.Model // small preview (fixed height)
	fullVP     viewport.Model // full scrollable view
	selected   int            // 0 preview, 1 dry-run, 2 send
	focus      sendFocus
	dryRun     bool
	status     string
	sending    bool
	sendOk     bool
	sendErr    error
	spinnerIdx int
	width      int
	height     int
}

func NewSendTab(cfg *config.Config) *SendTab {
	return &SendTab{
		cfg:    cfg,
		dryRun: true,
	}
}

func (s *SendTab) Init() tea.Cmd {
	body, err := template.Render(s.cfg, "config.toml")
	if err != nil {
		body = "Error rendering template: " + err.Error()
	}
	// small preview fixed 5 lines
	s.smallVP = viewport.New(0, 5)
	s.smallVP.SetContent(body)
	// full view default size (will be resized on WindowSizeMsg)
	s.fullVP = viewport.New(80, 20)
	s.fullVP.SetContent(body)
	return nil
}

func (s *SendTab) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		contentWidth := max(s.width-4, 20)
		// small preview width
		s.smallVP.Width = contentWidth
		// full view uses most of the content height (leave a couple lines for hint)
		fullHeight := max(s.height-6, 10)
		s.fullVP.Width = contentWidth
		s.fullVP.Height = fullHeight
		return s, nil

	case spinnerTickMsg:
		if s.sending {
			s.spinnerIdx = (s.spinnerIdx + 1) % len(spinnerFrames)
			return s, spinnerTick()
		}

	case tea.KeyMsg:
		switch s.focus {
		case focusList:
			switch msg.String() {
			case "up", "k":
				s.selected = (s.selected - 1 + 3) % 3
			case "down", "j":
				s.selected = (s.selected + 1) % 3
			case "enter":
				if s.selected == 0 {
					s.focus = focusFull
				} else if s.selected == 2 {
					s.sending = true
					s.sendOk = false
					s.sendErr = nil
					s.spinnerIdx = 0
					s.status = ""
					dry := s.dryRun
					return s, tea.Batch(
						func() tea.Msg {
							start := time.Now()
							body, err := template.Render(s.cfg, "config.toml")
							if err != nil {
								return sendResultMsg{err: err}
							}
							if dry {
								duration := time.Since(start)
								logger.Log.Info("mail dry-run",
									"event", "mail_dry_run",
									"dry_run", true,
									"to", s.cfg.Mail.ToAddr,
									"subject", "DSGVO Auskunftsersuchen",
									"duration_ms", duration.Milliseconds(),
								)
								return sendResultMsg{msg: "Dry‑run ok"}
							}
							if err := mailer.Send(s.cfg, "DSGVO Auskunftsersuchen", body); err != nil {
								duration := time.Since(start)
								logger.Log.Error("mail send failed",
									"event", "mail_send_failed",
									"dry_run", false,
									"to", s.cfg.Mail.ToAddr,
									"subject", "DSGVO Auskunftsersuchen",
									"error", err.Error(),
									"duration_ms", duration.Milliseconds(),
								)
								return sendResultMsg{err: err}
							}
							duration := time.Since(start)
							logger.Log.Info("mail sent",
								"event", "mail_sent",
								"dry_run", false,
								"to", s.cfg.Mail.ToAddr,
								"subject", "DSGVO Auskunftsersuchen",
								"duration_ms", duration.Milliseconds(),
							)
							return sendResultMsg{msg: "Mail sent successfully"}
						},
						spinnerTick(),
					)
				}
			case " ":
				if s.selected == 1 {
					s.dryRun = !s.dryRun
				}
			}
		case focusFull:
			switch msg.String() {
			case "esc":
				s.focus = focusList
			default:
				var cmd tea.Cmd
				s.fullVP, cmd = s.fullVP.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case sendResultMsg:
		s.sending = false
		if msg.err != nil {
			s.sendErr = msg.err
			s.status = "Error: " + msg.err.Error()
		} else {
			s.sendOk = true
			s.status = msg.msg
			// persist last sent timestamp for recurring tab (only on real send)
			if !s.dryRun {
				s.cfg.Schedule.LastSent = time.Now().Format(time.RFC3339)
				_ = config.Save("config.toml", s.cfg)
			}
		}
	}

	return s, tea.Batch(cmds...)
}

func (s *SendTab) View() string {
	if s.focus == focusFull {
		hint := lipgloss.NewStyle().Faint(true).Render("Esc to close")
		return lipgloss.JoinVertical(lipgloss.Center, s.fullVP.View(), hint)
	}

	// normal list view
	contentWidth := max(s.width-4, 20)

	previewStyle := lipgloss.NewStyle().
		Width(contentWidth).
		Height(5).
		Padding(0, 1)

	activePreviewStyle := previewStyle.Copy().
		Bold(true)

	chosenPreviewStyle := previewStyle
	if s.focus == focusList && s.selected == 0 {
		chosenPreviewStyle = activePreviewStyle
	}
	preview := chosenPreviewStyle.Render(s.smallVP.View())

	cbMark := "[ ]"
	if s.dryRun {
		cbMark = "[x]"
	}
	cbLine := fmt.Sprintf("%s Dry run", cbMark)
	if s.focus == focusList && s.selected == 1 {
		cbLine = "> " + cbLine + " <"
	}

	btnLabel := "[ Send ]"
	if s.focus == focusList && s.selected == 2 {
		btnLabel = "> Send <"
	}
	btnStyle := lipgloss.NewStyle().Bold(true)
	btnView := btnStyle.Render(btnLabel)

	var actionLine string
	if s.sending {
		frame := spinnerFrames[s.spinnerIdx]
		actionLine = lipgloss.NewStyle().Foreground(highlight).Render(frame + " Sending…")
	} else if s.sendOk {
		actionLine = lipgloss.NewStyle().Foreground(green).Render("✔ " + s.status)
	} else if s.sendErr != nil {
		actionLine = lipgloss.NewStyle().Foreground(red).Render("✘ " + s.status)
	} else {
		actionLine = s.status
	}

	parts := []string{preview, "", cbLine, "", btnView, "", actionLine}
	body := lipgloss.JoinVertical(lipgloss.Left, parts...)
	return body
}