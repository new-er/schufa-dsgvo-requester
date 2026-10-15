const api = globalThis.browser ?? chrome;

const form = document.getElementById("form");
const statusEl = document.getElementById("status");
const notifyEnabled = document.getElementById("notify_enabled");
const notifyRepeat = document.getElementById("notify_repeat");

const FIELDS = [
  "vorname",
  "nachname",
  "geburtsdatum",
  "strasse",
  "plz",
  "wohnort",
  "land",
  "interval_months",
];
const RECIPIENTS = ["to_addr", "cc_addr", "bcc_addr"];
const ALL_FIELDS = FIELDS.concat(RECIPIENTS);
const SETTINGS = ["notify_enabled", "notify_repeat"];

function readForm() {
  const data = {};
  for (const name of ALL_FIELDS) {
    const value = form.elements[name].value;
    data[name] = name === "interval_months" ? Number(value) : value;
  }
  return data;
}

function fillForm(data) {
  for (const name of ALL_FIELDS) {
    if (data[name] !== undefined) form.elements[name].value = data[name];
  }
}

function readSettings() {
  return {
    notify_enabled: notifyEnabled.checked,
    notify_repeat: notifyRepeat.value,
  };
}

function showDue(data) {
  const due =
    data.last_sent && globalThis.schufa
      ? globalThis.schufa.nextDueDate(data.last_sent, data.interval_months)
      : null;
  statusEl.textContent = due
    ? `Letzte Anfrage: ${new Date(data.last_sent).toLocaleDateString("de-DE")} · nächste: ${due.toLocaleDateString("de-DE")}`
    : "Noch keine Anfrage gesendet.";
}

function fail(err) {
  console.error(err);
  statusEl.textContent = "Fehler: " + (err?.message ?? err);
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  try {
    await api.storage.local.set({ ...readForm(), ...readSettings() });
    showDue(await api.storage.local.get(FIELDS.concat(["last_sent"])));
    statusEl.textContent = "Gespeichert ✓ " + statusEl.textContent;
  } catch (err) {
    fail(err);
  }
});

async function init() {
  const data = await api.storage.local.get(
    ALL_FIELDS.concat(["last_sent"]).concat(SETTINGS),
  );
  fillForm(data);
  form.elements.to_addr.value =
    data.to_addr || globalThis.schufa?.DEFAULT_TO || "info@schufa.de";
  showDue(data);
  notifyEnabled.checked = data.notify_enabled ?? true;
  notifyRepeat.value = data.notify_repeat ?? "once";
}

init().catch(fail);
