# SCHUFA DSGVO Requester

## Motivation

I love SCHUFA, and as a Programmer, I love automating things.

I also appreciate the classic German way of receiving information: by letter.

This project simply brings these passions together by requesting my SCHUFA Datenkopie automatically every three months.

---

## What it does

Sends a GDPR (DSGVO) Art. 15 data access request to SCHUFA via email on a recurring schedule. The request includes your personal details and asks SCHUFA to send the data copy by post to your address.

## Features

- **Automated scheduling** — runs every 3 months (configurable) via systemd timer
- **Interactive TUI** — runs by default (no flags), configure and send from terminal
- **Dry-run mode** — preview the email before sending (`--dry-run`)
- **Systemd integration** — `install-systemd` / `uninstall-systemd` commands
- **Auto-config creation** — generates `config.toml` from embedded example on first run
- **Secure config** — validates file permissions (must be `0600`)
- **Structured logging** — JSON logs to `~/.local/share/schufa-dsgvo-requester/logs/`

## Install

```bash
go install github.com/new-er/schufa-dsgvo-requester@latest
```

## Usage

```bash
schufa-dsgvo-requester              # opens TUI → configure → send once or install recurring schedule
schufa-dsgvo-requester --dry-run    # preview email without sending
schufa-dsgvo-requester install-systemd
schufa-dsgvo-requester uninstall-systemd
```

## Config (auto-created on first run next to binary)

Edit `config.toml` (must be `chmod 600`):

```toml
[mail]
smtp_host = "smtp.example.com"
smtp_port = 587
username  = "your@email.com"
password  = "app-password"
from_addr = "your@email.com"
to_addr   = "info@schufa.de"

[personal]
vorname      = "Max"
nachname     = "Mustermann"
geburtsdatum = "04.04.2000"
strasse      = "Musterstraße 12"
plz          = "12345"
wohnort      = "Musterstadt"
land         = "Deutschland"

[schedule]
interval_months = 3
# last_sent = "2026-01-15T09:00:00Z"  # set automatically after first send
```

| Section | Key | Description |
|---------|-----|-------------|
| `mail` | `smtp_host` | SMTP server hostname |
| `mail` | `smtp_port` | SMTP port (usually 587 for STARTTLS) |
| `mail` | `username` | SMTP username |
| `mail` | `password` | SMTP password / app password |
| `mail` | `from_addr` | Sender email address |
| `mail` | `to_addr` | Recipient (SCHUFA: `info@schufa.de`) |
| `personal` | `vorname` | First name |
| `personal` | `nachname` | Last name |
| `personal` | `geburtsdatum` | Birth date (DD.MM.YYYY) |
| `personal` | `strasse` | Street + number |
| `personal` | `plz` | Postal code |
| `personal` | `wohnort` | City |
| `personal` | `land` | Country |
| `schedule` | `interval_months` | Interval in months (default: 3) |

## Recurring Schedule

The timer runs on the 1st of every N months at 09:00 (with ±1h randomization):

```bash
# Install (creates /etc/systemd/system/schufa-dsgvo.{service,timer})
schufa-dsgvo-requester install-systemd

# Check status
systemctl status schufa-dsgvo.timer
journalctl -u schufa-dsgvo.service -f

# Uninstall
schufa-dsgvo-requester uninstall-systemd
```

## Logs

Structured JSON logs at `~/.local/share/schufa-dsgvo-requester/logs/schufa.log`:

```json
{"level":"info","event":"scheduled_run_start","trigger":"cron","time":"2026-01-01T09:00:00Z"}
{"level":"info","event":"scheduled_run_finished","trigger":"cron","result":"ok","time":"2026-01-01T09:00:02Z"}
```

## Email Template

The sent email uses `internal/template/mail.txt` (Go template):

```
Sehr geehrte Damen und Herren,

hiermit beantrage ich gemäß Art. 15 DSGVO Auskunft über die zu meiner Person gespeicherten Daten.

Persönliche Daten
Vorname(n)
{{ .Vorname }}
...
```

## Requirements

- Linux with systemd (for recurring schedule)
- SMTP with STARTTLS (port 587)
- `pkexec` or `sudo` for install-systemd
- Go 1.24+ (only for building from source)

## License

MIT
