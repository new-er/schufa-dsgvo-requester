package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/new-er/schufa-dsgvo-requester/internal/config"
	"github.com/new-er/schufa-dsgvo-requester/internal/logger"
	"github.com/new-er/schufa-dsgvo-requester/internal/mailer"
	"github.com/new-er/schufa-dsgvo-requester/internal/scheduler"
	"github.com/new-er/schufa-dsgvo-requester/internal/template"
	"github.com/new-er/schufa-dsgvo-requester/internal/tui"
)

func main() {
	// Check for subcommands first
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install-systemd":
			runInstallSystemd()
			return
		case "uninstall-systemd":
			runUninstallSystemd()
			return
		}
	}

	dryRun := flag.Bool("dry-run", false, "Print mail instead of sending")
	flag.Parse()

	// Ensure config exists (creates from embedded example if missing)
	cfgPath := configPath()
	if _, err := config.LoadOrCreate(cfgPath); err != nil {
		log.Fatalf("config error: %v", err)
	}

	// Default to TUI when no explicit flags given
	if !*dryRun && len(os.Args) == 1 {
		if err := tui.Run(cfgPath); err != nil {
			log.Fatalf("tui error: %v", err)
		}
		return
	}
	cfg, err := config.LoadOrCreate(cfgPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logDir := logger.DefaultLogDir()
	if err := logger.Init(logDir); err != nil {
		log.Fatalf("logger init error: %v", err)
	}
	defer logger.Close()

	body, err := template.Render(cfg, cfgPath)
	if err != nil {
		log.Fatalf("template error: %v", err)
	}

	if *dryRun {
		fmt.Print(body)
		logger.Close()
		os.Exit(0)
	}

	// CLI send (could be scheduled via cron)
	logger.Log.Info("scheduled run start",
		"event", "scheduled_run_start",
		"trigger", "cron",
	)
	if err := mailer.Send(cfg, "DSGVO Auskunftsersuchen", body); err != nil {
		logger.Log.Error("scheduled run finished with error",
			"event", "scheduled_run_finished",
			"trigger", "cron",
			"result", "error",
			"error", err.Error(),
		)
		log.Fatalf("send error: %v", err)
	}
	logger.Log.Info("scheduled run finished",
		"event", "scheduled_run_finished",
		"trigger", "cron",
		"result", "ok",
	)
}

func configPath() string {
	bin, _ := os.Executable()
	return filepath.Join(filepath.Dir(bin), "config.toml")
}

func runInstallSystemd() {
	cfgPath := configPath()
	cfg, err := config.LoadOrCreate(cfgPath)
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	binPath, _ := os.Executable()
	if err := scheduler.InstallSystemd(cfg, binPath, cfgPath); err != nil {
		log.Fatalf("install error: %v", err)
	}
	fmt.Println("Systemd timer installed")
}

func runUninstallSystemd() {
	if err := scheduler.UninstallSystemd(); err != nil {
		log.Fatalf("uninstall error: %v", err)
	}
	fmt.Println("Systemd timer uninstalled")
}