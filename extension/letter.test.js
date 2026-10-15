const assert = require("node:assert");
require("./letter.js");

const { nextDueDate, renderLetter, mailtoUrl } = globalThis.schufa;

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

const url = mailtoUrl({ vorname: "Max", nachname: "Mustermann" });
assert.ok(url.startsWith("mailto:info@schufa.de?"));
assert.ok(url.includes("body="));

console.log("letter.test.js: all assertions passed");
