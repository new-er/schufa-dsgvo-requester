function monthsFrom(iso, months) {
  const d = new Date(iso);
  d.setMonth(d.getMonth() + months);
  return d;
}

function nextDueDate(lastSent, intervalMonths) {
  if (!lastSent) return null;
  return monthsFrom(lastSent, intervalMonths);
}

function renderLetter(data) {
  return `Sehr geehrte Damen und Herren,

hiermit beantrage ich gemäß Art. 15 DSGVO Auskunft über die zu meiner Person gespeicherten Daten.

Persönliche Daten
Vorname(n)
${data.vorname}

Nachname(n)
${data.nachname}

Geburtsdatum
${data.geburtsdatum}

Anschrift
Straße & Nr.
${data.strasse}

PLZ
${data.plz}

Wohnort
${data.wohnort}

Land
${data.land}

Bitte senden Sie mir die Auskunft an die oben genannte Anschrift.

Mit freundlichen Grüßen
${data.vorname} ${data.nachname}`;
}

const DEFAULT_TO = "info@schufa.de";

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
  return {
    to: to.length ? to : [DEFAULT_TO],
    cc: addressList(data.cc_addr),
    bcc: addressList(data.bcc_addr),
  };
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

async function requestViaMailto(api, data) {
  const now = new Date().toISOString();
  const due = nextDueDate(now, data.interval_months || 3);
  await api.storage.local.set({ last_sent: now, next_due: due.getTime() });
  location.href = mailtoUrl(data);
}

globalThis.schufa = { DEFAULT_TO, nextDueDate, renderLetter, recipients, mailtoUrl, requestViaMailto };
