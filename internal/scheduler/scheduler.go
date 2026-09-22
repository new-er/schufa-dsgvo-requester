package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
	"time"

	"github.com/youruser/schufa-dsgvo-requester/internal/config"
)

const systemdServiceTmpl = `[Unit]
Description=SCHUFA DSGVO Requester
After=network-online.target

[Service]
Type=oneshot
ExecStart={{.BinaryPath}} --config {{.ConfigPath}}
Environment=CONFIG_PATH={{.ConfigPath}}
StandardOutput=journal
StandardError=journal
User={{.User}}
`

const systemdTimerTmpl = `[Unit]
Description=Quarterly SCHUFA DSGVO Request

[Timer]
OnCalendar={{.OnCalendar}}
Persistent=true
RandomizedDelaySec=1h

[Install]
WantedBy=timers.target
`

type SchedulerData struct {
	BinaryPath string
	ConfigPath string
	OnCalendar string
	User       string
}

// GenerateSystemd writes service and timer files to outDir.
func GenerateSystemd(cfg *config.Config, binaryPath, configPath, outDir, onCalendar string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}
	user := os.Getenv("USER")
	if user == "" {
		user = "root"
	}
	data := SchedulerData{
		BinaryPath: binaryPath,
		ConfigPath: configPath,
		OnCalendar: onCalendar,
		User:       user,
	}
	svcPath := filepath.Join(outDir, "schufa-dsgvo.service")
	timerPath := filepath.Join(outDir, "schufa-dsgvo.timer")
	if err := writeTemplate(svcPath, systemdServiceTmpl, data); err != nil {
		return err
	}
	if err := writeTemplate(timerPath, systemdTimerTmpl, data); err != nil {
		return err
	}
	return nil
}

func writeTemplate(path, tmplStr string, data SchedulerData) error {
	tmpl, err := template.New("sched").Parse(tmplStr)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, data)
}

// InstallSystemd generates unit files, copies to /etc/systemd/system and enables timer.
func InstallSystemd(cfg *config.Config, binaryPath, configPath string) error {
	tmpDir, err := os.MkdirTemp("", "schufa-dsgvo-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	onCalendar := fmt.Sprintf("*-*-01 09:00:00")
	if cfg.Schedule.IntervalMonths > 0 {
		onCalendar = fmt.Sprintf("*-01/%d-01 09:00:00", cfg.Schedule.IntervalMonths)
	}

	if err := GenerateSystemd(cfg, binaryPath, configPath, tmpDir, onCalendar); err != nil {
		return err
	}

	script := fmt.Sprintf(`set -e
cp %s/schufa-dsgvo.service /etc/systemd/system/
cp %s/schufa-dsgvo.timer /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now schufa-dsgvo.timer
`, tmpDir, tmpDir)

	cmd := exec.Command("pkexec", "sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func IsSystemdInstalled() bool {
	svc := "/etc/systemd/system/schufa-dsgvo.service"
	timer := "/etc/systemd/system/schufa-dsgvo.timer"
	_, err1 := os.Stat(svc)
	_, err2 := os.Stat(timer)
	return err1 == nil && err2 == nil
}

func UninstallSystemd() error {
	script := `set -e
systemctl disable --now schufa-dsgvo.timer
rm -f /etc/systemd/system/schufa-dsgvo.service /etc/systemd/system/schufa-dsgvo.timer
systemctl daemon-reload
`
	cmd := exec.Command("pkexec", "sh", "-c", script)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// NextRun returns the next execution time based on the systemd timer OnCalendar format.
func NextRun(cfg *config.Config, now time.Time) (time.Time, error) {
	interval := cfg.Schedule.IntervalMonths
	if interval <= 0 {
		interval = 3
	}
	// Timer runs on 1st of month at 09:00 every N months
	year, month, _ := now.Date()
	candidate := time.Date(year, month, 1, 9, 0, 0, 0, now.Location())
	if candidate.Before(now) || candidate.Equal(now) {
		candidate = candidate.AddDate(0, interval, 0)
	}
	// Find the next month that matches the interval
	for candidate.Month()%time.Month(interval) != 1%time.Month(interval) {
		candidate = candidate.AddDate(0, 1, 0)
	}
	return candidate, nil
}