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

function mailtoUrl(data) {
  const body = renderLetter(data);
  return `mailto:info@schufa.de?subject=${encodeURIComponent("DSGVO Auskunftsersuchen")}&body=${encodeURIComponent(body)}`;
}

async function requestViaMailto(api, data) {
  const now = new Date().toISOString();
  const due = nextDueDate(now, data.interval_months || 3);
  await api.storage.local.set({ last_sent: now, next_due: due.getTime() });
  location.href = mailtoUrl(data);
}

globalThis.schufa = { nextDueDate, renderLetter, mailtoUrl, requestViaMailto };
