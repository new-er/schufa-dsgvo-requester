const api = globalThis.browser ?? chrome;
const { nextDueDate, requestViaMailto } = schufa;

const form = document.getElementById("form");
const status = document.getElementById("status");

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
  const due = data.last_sent ? nextDueDate(data.last_sent, data.interval_months) : null;
  status.textContent = due
    ? `Letzte Anfrage: ${new Date(data.last_sent).toLocaleDateString("de-DE")} · nächste: ${due.toLocaleDateString("de-DE")}`
    : "Noch keine Anfrage gesendet.";
}

api.storage.local.get(FIELDS.concat(["last_sent"])).then((data) => {
  fillForm(data);
  showDue(data);
});

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  await api.storage.local.set(readForm());
  const data = await api.storage.local.get(FIELDS.concat(["last_sent"]));
  showDue(data);
  status.textContent = "Gespeichert. " + status.textContent;
});

document.getElementById("open").addEventListener("click", async () => {
  const data = readForm();
  await api.storage.local.set(data);
  await requestViaMailto(api, data);
});
