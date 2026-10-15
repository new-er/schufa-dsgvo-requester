# SCHUFA DSGVO Requester (Browser Extension)

Browser extension (Firefox + Chrome, Manifest V3) that requests your SCHUFA data copy under Art. 15 DSGVO by opening a prefilled email to `info@schufa.de`, and reminds you when the next request is due.

## What it does

- Stores your personal data (name, birth date, address) in the browser's local storage — nothing leaves the browser
- Slim toolbar popup: due status + "E-Mail öffnen" (opens your default mail client with the letter prefilled)
- Full **settings/options page** (`options.html`, opens in a tab) for personal data and notification settings
- Schedules a recurring reminder (default: every 3 months) via `browser.alarms`; when due, a notification opens a compose page with the ready-to-send letter
- **Notification settings**: reminders can be turned off, and single vs. repeated (hourly/daily/weekly) once due

## Install

### Firefox

1. Open `about:debugging#/runtime/this-firefox`
2. Click "Load Temporary Add-on…" and pick `extension/manifest.json`

### Chrome

1. Open `chrome://extensions`
2. Enable "Developer mode"
3. "Load unpacked" and select the `extension/` folder

## Usage

1. Open the extension popup → **Einstellungen** (or the extension's options page) and fill in your personal data + interval, hit **Speichern**
2. In the popup, hit **E-Mail öffnen** — your mail client opens with the letter to `info@schufa.de`
3. Send it; the extension records the date and schedules the next reminder

## Files

| File | Purpose |
|------|---------|
| `manifest.json` | MV3 manifest, works in Firefox and Chrome (no build step) |
| `popup.html` / `popup.js` | slim quick action: due status, "E-Mail öffnen", settings link |
| `options.html` / `options.js` | full settings page: personal data + notification settings |
| `letter.js` | letter template + mailto/schedule helpers |
| `compose.html` / `compose.js` | letter view opened by the reminder notification |
| `background.js` | alarm + notification handling (honours notification settings) |
| `letter.test.js`, `background.test.js` | self-checks |

## Development

```bash
node letter.test.js           # letter/date/mailto self-checks
node background.test.js       # alarm scheduling self-checks
npx web-ext lint -s .         # validate for Firefox (run inside extension/)
```

The old Go CLI lives in `go/` during the refactor.
