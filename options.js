const api = globalThis.browser ?? chrome;

const form = document.getElementById("form");
const statusEl = document.getElementById("status");
const dueEl = document.getElementById("due");
const notifyEnabled = document.getElementById("notify_enabled");
const notifyRepeat = document.getElementById("notify_repeat");
const prevEnabled = document.getElementById("prev_enabled");

const DEMO = new URLSearchParams(location.search).has("demo");
const DEMO_DATA = {
  vorname: "Max",
  nachname: "Mustermann",
  geburtsdatum: "04.04.2000",
  strasse: "Musterstraße 12",
  plz: "12345",
  wohnort: "Musterstadt",
  land: "Deutschland",
};

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
const PREV_FIELDS = ["prev_strasse", "prev_plz", "prev_wohnort", "prev_land"];
const RECIPIENTS = ["to_addr", "cc_addr", "bcc_addr"];
const ALL_FIELDS = FIELDS.concat(PREV_FIELDS, RECIPIENTS);
const SETTINGS = ["notify_enabled", "notify_repeat", "prev_enabled"];

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
    prev_enabled: prevEnabled.checked,
  };
}

function togglePrev() {
  for (const name of PREV_FIELDS) {
    form.elements[name].disabled = !prevEnabled.checked;
    form.elements[name].required = prevEnabled.checked;
  }
}

prevEnabled.addEventListener("change", togglePrev);

function showDue(data) {
  const due =
    data.last_sent && globalThis.schufa
      ? globalThis.schufa.nextDueDate(data.last_sent, data.interval_months)
      : null;
  dueEl.textContent = due
    ? `Letzte Anfrage: ${new Date(data.last_sent).toLocaleDateString("de-DE")} · nächste: ${due.toLocaleDateString("de-DE")}`
    : "Noch keine Anfrage gesendet.";
}

api.storage.onChanged.addListener((changes, area) => {
  if (area !== "local") return;
  if (!("last_sent" in changes || "next_due" in changes)) return;
  api.storage.local.get(FIELDS.concat(["last_sent"])).then(showDue).catch(fail);
});

function fail(err) {
  console.error(err);
  statusEl.textContent = "Fehler: " + (err?.message ?? err);
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  try {
    await api.storage.local.set({ ...readForm(), ...readSettings() });
    statusEl.textContent = "Gespeichert ✓";
  } catch (err) {
    fail(err);
  }
});

async function init() {
  const data = await api.storage.local.get(
    ALL_FIELDS.concat(["last_sent"]).concat(SETTINGS),
  );
  fillForm(data);
  if (DEMO) {
    for (const [name, value] of Object.entries(DEMO_DATA)) {
      if (!form.elements[name].value) form.elements[name].value = value;
    }
  }
  form.elements.to_addr.value =
    data.to_addr || globalThis.schufa?.DEFAULT_TO || "info@schufa.de";
  showDue(data);
  notifyEnabled.checked = data.notify_enabled ?? true;
  notifyRepeat.value = data.notify_repeat ?? "once";
  prevEnabled.checked = data.prev_enabled ?? false;
  togglePrev();
}

init().catch(fail);
