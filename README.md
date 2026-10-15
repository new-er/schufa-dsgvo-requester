# SCHUFA DSGVO Requester (Browser Extension)

Browser extension (Firefox + Chrome, Manifest V3) that requests your SCHUFA data copy under Art. 15 DSGVO by opening a prefilled email to `info@schufa.de`, and reminds you when the next request is due.

## What it does

- Stores your personal data (name, birth date, address) in the browser's local storage — nothing leaves the browser
- "E-Mail öffnen" opens your default mail client with subject and letter body prefilled
- Schedules a recurring reminder (default: every 3 months) via `browser.alarms`; when due, a notification opens a compose page with the ready-to-send letter

## Install

### Firefox

1. Open `about:debugging#/runtime/this-firefox`
2. Click "Load Temporary Add-on…" and pick `extension/manifest.json`

### Chrome

1. Open `chrome://extensions`
2. Enable "Developer mode"
3. "Load unpacked" and select the `extension/` folder

## Usage

1. Open the extension popup and fill in your personal data + interval, hit **Speichern**
2. Hit **E-Mail öffnen** — your mail client opens with the letter to `info@schufa.de`
3. Send it; the extension records the date and schedules the next reminder

## Files

| File | Purpose |
|------|---------|
| `manifest.json` | MV3 manifest, works in Firefox and Chrome (no build step) |
| `popup.html` / `popup.js` | config form + "open email" action |
| `letter.js` | letter template + mailto/schedule helpers |
| `compose.html` / `compose.js` | letter view opened by the reminder notification |
| `background.js` | alarm + notification handling |
| `letter.test.js` | self-check: `node letter.test.js` |

## Development

```bash
node letter.test.js           # run self-checks
npx web-ext lint -s .         # validate for Firefox
```

The old Go CLI lives in `go/` during the refactor.
