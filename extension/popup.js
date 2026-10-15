const api = globalThis.browser ?? chrome;

const statusEl = document.getElementById("status");

function fail(err) {
  console.error(err);
  statusEl.textContent = "Fehler: " + (err?.message ?? err);
}

document.getElementById("open").addEventListener("click", async () => {
  try {
    const data = await api.storage.local.get(null);
    if (!data.vorname) {
      statusEl.textContent = "Bitte zuerst Daten in den Einstellungen speichern.";
      return;
    }
    if (!globalThis.schufa) throw new Error("letter.js nicht geladen");
    await globalThis.schufa.requestViaMailto(api, data);
  } catch (err) {
    fail(err);
  }
});

document.getElementById("settings").addEventListener("click", () => {
  api.runtime.openOptionsPage();
});

async function init() {
  const data = await api.storage.local.get(["vorname", "last_sent", "next_due"]);
  if (!data.vorname) {
    statusEl.textContent = "Noch keine Daten gespeichert.";
    return;
  }
  statusEl.textContent = data.next_due
    ? "Nächste Anfrage fällig: " + new Date(data.next_due).toLocaleDateString("de-DE")
    : "Noch keine Anfrage gesendet.";
}

init().catch(fail);
