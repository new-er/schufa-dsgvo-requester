const api = globalThis.browser ?? chrome;

const form = document.getElementById("form");
const statusEl = document.getElementById("status");

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

function readForm() {
  const data = {};
  for (const name of FIELDS) {
    const value = form.elements[name].value;
    data[name] = name === "interval_months" ? Number(value) : value;
  }
  return data;
}

function fillForm(data) {
  for (const name of FIELDS) {
    if (data[name] !== undefined) form.elements[name].value = data[name];
  }
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
    const data = readForm();
    await api.storage.local.set(data);
    showDue(await api.storage.local.get(FIELDS.concat(["last_sent"])));
    statusEl.textContent = "Gespeichert ✓ " + statusEl.textContent;
  } catch (err) {
    fail(err);
  }
});

document.getElementById("open").addEventListener("click", async () => {
  try {
    if (!globalThis.schufa) throw new Error("letter.js nicht geladen");
    const data = readForm();
    await api.storage.local.set(data);
    await globalThis.schufa.requestViaMailto(api, data);
  } catch (err) {
    fail(err);
  }
});

async function init() {
  const data = await api.storage.local.get(FIELDS.concat(["last_sent"]));
  fillForm(data);
  showDue(data);
}

init().catch(fail);
