const assert = require("node:assert");
require("../letter.js");

const { DEFAULT_TO, nextDueDate, renderLetter, recipients, mailtoUrl } = globalThis.schufa;

assert.strictEqual(DEFAULT_TO, "info@schufa.de");

const last = "2026-01-15T09:00:00Z";
const due = nextDueDate(last, 3);
assert.strictEqual((due.getMonth() - new Date(last).getMonth() + 12) % 12, 3);

assert.strictEqual(nextDueDate("", 3), null);

const letter = renderLetter({
  vorname: "Max",
  nachname: "Mustermann",
  geburtsdatum: "04.04.2000",
  strasse: "Musterstraße 12",
  plz: "12345",
  wohnort: "Musterstadt",
  land: "Deutschland",
});
assert.match(letter, /Art\. 15 DSGVO/);
assert.match(letter, /Max Mustermann$/);
assert.ok(!letter.includes("Vorherige Anschrift"), "no previous address by default");

const base = { vorname: "Max", nachname: "Mustermann" };
const prev = {
  prev_strasse: "Alte Str. 5",
  prev_plz: "54321",
  prev_wohnort: "Altstadt",
  prev_land: "Deutschland",
};

const withPrev = renderLetter({ ...base, prev_enabled: true, ...prev });
assert.ok(withPrev.includes("Vorherige Anschrift"), "previous address included");
assert.ok(withPrev.includes("Alte Str. 5"));
assert.ok(
  withPrev.indexOf("Vorherige Anschrift") < withPrev.indexOf("Bitte senden Sie mir"),
  "previous address before the request line",
);

const missingField = renderLetter({ ...base, prev_enabled: true, ...prev, prev_plz: "" });
assert.ok(!missingField.includes("Vorherige Anschrift"), "incomplete previous address omitted");

const notEnabled = renderLetter({ ...base, ...prev });
assert.ok(!notEnabled.includes("Vorherige Anschrift"), "previous address omitted when disabled");

const url = mailtoUrl({ vorname: "Max", nachname: "Mustermann" });
assert.ok(url.startsWith("mailto:info@schufa.de?"), "default recipient");
assert.ok(url.includes("body="));
assert.ok(!url.includes("cc="), "no cc by default");
assert.ok(!url.includes("bcc="), "no bcc by default");

assert.deepStrictEqual(recipients({}), { to: ["info@schufa.de"], cc: [], bcc: [] });
assert.deepStrictEqual(recipients({ to_addr: " a@b.de , c@d.de " }), {
  to: ["a@b.de", "c@d.de"],
  cc: [],
  bcc: [],
});

const custom = mailtoUrl({
  vorname: "Max",
  to_addr: "schufa@example.de, second@example.de",
  cc_addr: "me@example.com",
  bcc_addr: "archive@example.com",
});
assert.ok(custom.startsWith("mailto:schufa@example.de,second@example.de?"), "custom to list");
assert.ok(custom.includes("cc=me@example.com"), "cc kept");
assert.ok(custom.includes("bcc=archive@example.com"), "bcc kept");

const spaced = mailtoUrl({ vorname: "Max", to_addr: "Max Mustermann <max@example.com>" });
assert.ok(spaced.includes("%20"), "spaces encoded as %20");
assert.ok(!spaced.includes("+subject"), "no plus-encoding for spaces");

console.log("letter.test.js: all assertions passed");
