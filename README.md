# SCHUFA DSGVO Requester (Browser Extension)

Browser extension (Firefox + Chrome, Manifest V3) that requests your SCHUFA data copy under Art. 15 DSGVO by opening a prefilled email to `info@schufa.de`, and reminds you when the next request is due.

## Motivation

I love SCHUFA, and as a Programmer, I love automating things.

I also appreciate the classic German way of receiving information: by letter.

This project simply brings these passions together by requesting my SCHUFA Datenkopie automatically every three months.

## What it does

- Stores your personal data (name, birth date, address) in the browser's local storage — nothing leaves the browser
- Slim toolbar popup: due status + "E-Mail öffnen" (opens your default mail client with the letter prefilled)
- Full **settings/options page** (`options.html`, opens in a tab) for personal data and notification settings
- Schedules a recurring reminder (default: every 3 months) via `browser.alarms`; when due, a notification opens a compose page with the ready-to-send letter
- **Notification settings**: reminders can be turned off, and single vs. repeated (hourly/daily/weekly) once due
- **Configurable recipients**: editable To/CC/BCC (comma-separated lists; To defaults to `info@schufa.de`)
- **Optional previous address**: enable it in settings (requires all four fields) and it is added to the letter

## Install

### Firefox

**From Mozilla Add-ons:** [Install extension](https://addons.mozilla.org/en-US/firefox/addon/schufa-dsgvo-requester/)

Or load from this repository:

1. Open `about:debugging#/runtime/this-firefox`
2. Click "Load Temporary Add-on…" and pick `manifest.json`

### Chrome

**From Chrome Web Store:** [Install extension](https://chromewebstore.google.com/detail/schufa-dsgvo-requester/pdbmdihcobjamfelleebhpeicnikpifo)

Or load from this repository:

1. Open `chrome://extensions`
2. Enable "Developer mode"
3. "Load unpacked" and select the repository root folder

## Usage

1. Open the extension popup → **Einstellungen** (or the extension's options page) and fill in your personal data + interval, hit **Speichern**
2. In the popup, hit **E-Mail öffnen** — your mail client opens with the letter to `info@schufa.de`
3. Send it; the extension records the date and schedules the next reminder

## Files

| File | Purpose |
|------|---------|
| `manifest.json` | MV3 manifest, works in Firefox and Chrome (no build step) |
| `popup.html` / `popup.js` | slim quick action: due status, "E-Mail öffnen", settings link |
| `options.html` / `options.js` | full settings page: personal data, recipients + notification settings |
| `letter.js` | letter template + mailto/schedule helpers |
| `compose.html` / `compose.js` | letter view opened by the reminder notification |
| `background.js` | alarm + notification handling (honours notification settings) |
| `tests/` | Node self-checks (not part of the extension) |

## Development

```bash
node tests/letter.test.js       # letter/date/mailto self-checks
node tests/background.test.js   # alarm scheduling self-checks
node tests/options.test.js      # settings persistence self-checks
npx web-ext lint -s .           # validate for Firefox
```

Open the options page with a `?demo` query string to seed empty fields with sample data (e.g. from the extension console: `browser.tabs.create({ url: browser.runtime.getURL("options.html?demo") })`).

