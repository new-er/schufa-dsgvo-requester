const api = globalThis.browser ?? chrome;

function addressList(value) {
  return (value || "")
    .split(",")
    .map((addr) => addr.trim())
    .filter(Boolean);
}

function encodeAddress(addr) {
  return encodeURIComponent(addr).replace(/%40/g, "@");
}

function recipients(data) {
  const to = addressList(data.to_addr);
  const fallbackTo = globalThis.schufa?.DEFAULT_TO ?? "info@schufa.de";
  return {
    to: to.length ? to : [fallbackTo],
    cc: addressList(data.cc_addr),
    bcc: addressList(data.bcc_addr),
  };
}

function previousAddress(data) {
  const fields = ["prev_strasse", "prev_plz", "prev_wohnort", "prev_land"];
  if (!data.prev_enabled || !fields.every((key) => data[key])) return "";
  return `\n\nVorherige Anschrift\nStraße & Nr.\n${data.prev_strasse}\n\nPLZ\n${data.prev_plz}\n\nWohnort\n${data.prev_wohnort}\n\nLand\n${data.prev_land}`;
}

function renderLetter(data) {
  return `Sehr geehrte Damen und Herren,

hiermit beantrage ich gemäß Art. 15 DSGVO Auskunft über die zu meiner Person gespeicherten Daten.

Persönliche Daten
Vorname(n)
${data.vorname || "(nicht gesetzt)"}

Nachname(n)
${data.nachname || "(nicht gesetzt)"}

Geburtsdatum
${data.geburtsdatum || "(nicht gesetzt)"}

Anschrift
Straße & Nr.
${data.strasse || "(nicht gesetzt)"}

PLZ
${data.plz || "(nicht gesetzt)"}

Wohnort
${data.wohnort || "(nicht gesetzt)"}

Land
${data.land || "(nicht gesetzt)"}${previousAddress(data)}

Bitte senden Sie mir die Auskunft an die oben genannte Anschrift.

Mit freundlichen Grüßen
${data.vorname || ""} ${data.nachname || ""}`;
}

function mailtoUrl(data) {
  const body = renderLetter(data);
  const { to, cc, bcc } = recipients(data);
  const params = [
    `subject=${encodeURIComponent("DSGVO Auskunftsersuchen")}`,
    `body=${encodeURIComponent(body)}`,
  ];
  if (cc.length) params.push(`cc=${cc.map(encodeAddress).join(",")}`);
  if (bcc.length) params.push(`bcc=${bcc.map(encodeAddress).join(",")}`);
  return `mailto:${to.map(encodeAddress).join(",")}?${params.join("&")}`;
}

let data = {};

function show(msg) {
  const el = document.getElementById("letter");
  if (el) el.textContent = msg;
  console.log(msg);
}

function fail(msg) {
  const el = document.getElementById("letter");
  if (el) el.textContent = "FEHLER: " + msg;
  console.error(msg);
  const r = document.getElementById("recipients");
  if (r) r.textContent = "";
}

api.storage.local.get(null).then((stored) => {
  data = stored || {};
  try {
    document.getElementById("letter").textContent = renderLetter(data);
    const r = recipients(data);
    const parts = [`An: ${r.to.join(", ")}`];
    if (r.cc.length) parts.push(`CC: ${r.cc.join(", ")}`);
    if (r.bcc.length) parts.push(`BCC: ${r.bcc.join(", ")}`);
    document.getElementById("recipients").textContent = parts.join(" · ");
    console.log("compose.js loaded OK");
  } catch (err) {
    fail("Rendern: " + err);
  }
}).catch((err) => {
  fail("Storage: " + err);
});

document.getElementById("open").addEventListener("click", () => {
  try {
    const now = new Date().toISOString();
    const interval = data.interval_months || 3;
    const due = new Date(now);
    due.setMonth(due.getMonth() + interval);
    api.storage.local.set({ last_sent: now, next_due: due.getTime() }).then(() => {
      location.href = mailtoUrl(data);
    }).catch((err) => {
      fail("Speichern: " + err);
    });
  } catch (err) {
    fail("Mailto: " + err);
  }
});
