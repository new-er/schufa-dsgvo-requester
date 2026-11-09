# SCHUFA DSGVO Requester (Browser Extension)

Browser extension (Firefox + Chrome, Manifest V3) that requests your SCHUFA data copy under Art. 15 DSGVO by opening a prefilled email to `info@schufa.de`, and reminds you when the next request is due.

## Motivation

I love SCHUFA, and as a Programmer, I love automating things.

I also appreciate the classic German way of receiving information: by letter.

This project simply brings these passions together by requesting my SCHUFA Datenkopie automatically every three months.

## What it does

- Stores your personal data (name, birth date, address) in the browser's local storage — nothing leaves the browser
- Slim toolbar popup: due status + "E-Mail öffnen" (opens your default mail client with the letter prefilled)
- **Toolbar icon shows status**: grey = nothing due, blue = a new request is due and ready to send
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

## Publishing (CI/CD)

Publishing is triggered by pushing a version tag (e.g. `git tag v0.1.1 && git push origin v0.1.1`). The [`publish.yml`](.github/workflows/publish.yml) workflow zips the extension, runs the tests, then uploads to:

- **Firefox AMO** via `web-ext sign` (`web-ext sign --help`) — needs `AMO_JWT_ISSUER` + `AMO_JWT_SECRET`.
- **Chrome Web Store** via `tools/publish-chrome.js` (Chrome Web Store API) — needs `CHROME_WEBSTORE_CLIENT_ID`, `CHROME_WEBSTORE_CLIENT_SECRET`, `CHROME_WEBSTORE_REFRESH_TOKEN` + `CHROME_WEBSTORE_EXTENSION_ID`.

Both stores review submissions, so "uploaded" does not mean "live immediately". Bump `version` in `manifest.json` for every tag — neither store accepts a duplicate version.

### CI/CD secret variables

Add these in your repository **Settings → Secrets and variables → Actions**:

| Secret | Where to get it |
|--------|-----------------|
| `AMO_JWT_ISSUER` | addons.mozilla.org → your account → API keys → create a JWT key → "JWT issuer" |
| `AMO_JWT_SECRET` | same JWT key → "JWT secret" |
| `AMO_ADDON_ID` | (not needed) — the add-on ID comes from `manifest.json` (`browser_specific_settings.gecko.id`) |
| `CHROME_WEBSTORE_CLIENT_ID` | Google Cloud Console → APIs & Services → OAuth2 client (web app) → client id |
| `CHROME_WEBSTORE_CLIENT_SECRET` | same OAuth2 client → client secret |
| `CHROME_WEBSTORE_REFRESH_TOKEN` | one-time OAuth2 flow with scope `https://www.googleapis.com/auth/chromewebstore` (see below) |
| `CHROME_WEBSTORE_EXTENSION_ID` | Chrome Web Store item id from your extension's store URL (for the current listing: `pdbmdihcobjamfelleebhpeicnikpifo`) |

### Getting a refresh token (one-time)

1. In Google Cloud Console, create an OAuth2 **web app** client and note its `client_id` and `client_secret`.
2. Open this URL (replace `YOUR_CLIENT_ID`) and sign in with the Google Account that is a Chrome Web Store developer for your extension:

   ```
   https://accounts.google.com/o/oauth2/v2/auth?client_id=YOUR_CLIENT_ID&redirect_uri=urn:ietf:wg:oauth:2.0:oob&scope=https://www.googleapis.com/auth/chromewebstore&response_type=code&access_type=offline&prompt=consent
   ```

3. Exchange the shown code for a refresh token:

   ```bash
   curl -X POST https://oauth2.googleapis.com/token \
     -d code=THE_CODE \
     -d client_id=YOUR_CLIENT_ID \
     -d client_secret=YOUR_CLIENT_SECRET \
     -d redirect_uri=urn:ietf:wg:oauth:2.0:oob \
     -d grant_type=authorization_code
   ```

4. Save the response's `refresh_token` as the `CHROME_WEBSTORE_REFRESH_TOKEN` secret (it does not expire).

The first Chrome upload must be seeded manually once via the Chrome Web Store developer dashboard to create the item; after that CI can update and publish it.

Open the options page with a `?demo` query string to seed empty fields with sample data (e.g. from the extension console: `browser.tabs.create({ url: browser.runtime.getURL("options.html?demo") })`).

