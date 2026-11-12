const api = globalThis.browser ?? chrome;

const statusEl = document.getElementById("status");
const openBtn = document.getElementById("open");

const REQUIRED = [
  "vorname",
  "nachname",
  "geburtsdatum",
  "strasse",
  "plz",
  "wohnort",
  "land",
];

function fail(err) {
  console.error(err);
  statusEl.textContent = "Fehler: " + (err?.message ?? err);
}

function renderStatus(data) {
  if (!REQUIRED.every((key) => data[key])) {
    openBtn.disabled = true;
    statusEl.textContent = "Bitte zuerst Name, Adresse usw. in den Einstellungen speichern.";
    return;
  }
  openBtn.disabled = false;
  statusEl.textContent = data.next_due
    ? "Nächste Anfrage fällig: " + new Date(data.next_due).toLocaleDateString("de-DE")
    : "Noch keine Anfrage gesendet.";
}

openBtn.addEventListener("click", async () => {
  try {
    const data = await api.storage.local.get(null);
    if (!REQUIRED.every((key) => data[key])) {
      statusEl.textContent = "Bitte zuerst Name, Adresse usw. in den Einstellungen speichern.";
      return;
    }
    await api.tabs.create({ url: api.runtime.getURL("compose.html") });
  } catch (err) {
    fail(err);
  }
});

document.getElementById("settings").addEventListener("click", () => {
  api.runtime.openOptionsPage();
});

async function init() {
  const data = await api.storage.local.get(REQUIRED.concat(["last_sent", "next_due"]));
  renderStatus(data);
}

init().catch(fail);

api.storage.onChanged.addListener((changes, area) => {
  if (area !== "local") return;
  if (!("last_sent" in changes || "next_due" in changes)) return;
  api.storage.local.get(REQUIRED.concat(["last_sent", "next_due"])).then(renderStatus).catch(fail);
});
