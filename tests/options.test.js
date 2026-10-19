const assert = require("node:assert");
const path = require("node:path");

require("../letter.js");

const OPTIONS = path.join(__dirname, "..", "options.js");

const FIELD_NAMES = [
  "vorname",
  "nachname",
  "geburtsdatum",
  "strasse",
  "plz",
  "wohnort",
  "land",
  "interval_months",
  "prev_strasse",
  "prev_plz",
  "prev_wohnort",
  "prev_land",
  "to_addr",
  "cc_addr",
  "bcc_addr",
];

function mount(stored, search = "") {
  const els = {};
  for (const name of FIELD_NAMES) {
    els[name] = { value: name === "interval_months" ? "3" : "", disabled: false, required: false };
  }
  const statusEl = { textContent: "" };
  const dueEl = { textContent: "" };
  const listeners = {};
  const form = { elements: els, addEventListener: (type, fn) => (listeners.form = fn) };
  const notifyEnabled = { checked: true };
  const notifyRepeat = { value: "once" };
  const prevEnabled = { checked: false, addEventListener: (type, fn) => (listeners.prev = fn) };
  globalThis.location = { search };
  globalThis.document = {
    getElementById: (id) =>
      ({ form, status: statusEl, due: dueEl, notify_enabled: notifyEnabled, notify_repeat: notifyRepeat, prev_enabled: prevEnabled }[id]),
  };
  globalThis.browser = {
    storage: {
      local: {
        get: async (keys) =>
          Object.fromEntries(keys.filter((k) => k in stored).map((k) => [k, stored[k]])),
        set: async (obj) => Object.assign(stored, obj),
      },
      onChanged: { addListener: (fn) => (listeners.storage = fn) },
    },
  };
  delete require.cache[OPTIONS];
  require(OPTIONS);
  return { els, statusEl, dueEl, listeners, notifyEnabled, notifyRepeat, prevEnabled, stored };
}

const tick = () => new Promise((r) => setTimeout(r, 20));

(async () => {
  const prev = {
    prev_strasse: "Alte Str. 5",
    prev_plz: "54321",
    prev_wohnort: "Altstadt",
    prev_land: "Deutschland",
  };

  // 1. stored settings are restored on load (regression: prev_enabled was not read)
  {
    const ui = mount({ notify_enabled: false, notify_repeat: "weekly", prev_enabled: true, ...prev });
    await tick();
    assert.strictEqual(ui.prevEnabled.checked, true, "prev_enabled restored");
    assert.strictEqual(ui.notifyEnabled.checked, false, "notify_enabled restored");
    assert.strictEqual(ui.notifyRepeat.value, "weekly", "notify_repeat restored");
    assert.strictEqual(ui.els.prev_strasse.disabled, false, "prev fields enabled when on");
    assert.strictEqual(ui.els.prev_strasse.required, true, "prev fields required when on");
    assert.strictEqual(ui.els.prev_strasse.value, "Alte Str. 5", "prev field restored");
  }

  // 2. disabling the checkbox disables the fields
  {
    const ui = mount({ prev_enabled: true, ...prev });
    await tick();
    ui.prevEnabled.checked = false;
    ui.listeners.prev();
    assert.strictEqual(ui.els.prev_plz.disabled, true, "fields disabled when off");
    assert.strictEqual(ui.els.prev_plz.required, false, "fields not required when off");
  }

  // 2b. ?demo seeds only empty fields, never overwriting stored data
  {
    const empty = mount({}, "?demo");
    await tick();
    assert.strictEqual(empty.els.vorname.value, "Max", "demo seeds empty field");
    assert.strictEqual(empty.els.wohnort.value, "Musterstadt", "demo seeds city");

    const partial = mount({ vorname: "Erika" }, "?demo");
    await tick();
    assert.strictEqual(partial.els.vorname.value, "Erika", "stored value not overwritten");
    assert.strictEqual(partial.els.nachname.value, "Mustermann", "other fields still seeded");

    const noFlag = mount({}, "");
    await tick();
    assert.strictEqual(noFlag.els.vorname.value, "", "no seeding without ?demo");
  }

  // 3. saving persists prev_enabled and the prev address
  {
    const ui = mount({});
    await tick();
    ui.prevEnabled.checked = true;
    ui.listeners.prev();
    for (const k of Object.keys(prev)) ui.els[k].value = prev[k];
    ui.els.vorname.value = "Max";
    ui.els.to_addr.value = "info@schufa.de";
    await ui.listeners.form({ preventDefault: () => {} });
    assert.strictEqual(ui.stored.prev_enabled, true, "prev_enabled persisted");
    assert.strictEqual(ui.stored.prev_strasse, "Alte Str. 5", "prev field persisted");
  }

  // 4. a live last_sent change refreshes the due line without reload
  {
    const ui = mount({ interval_months: 3 });
    await tick();
    assert.strictEqual(ui.dueEl.textContent, "Noch keine Anfrage gesendet.", "no sent -> empty due");
    const iso = new Date().toISOString();
    ui.stored.last_sent = iso;
    ui.listeners.storage({ last_sent: { newValue: iso } }, "local");
    await tick();
    assert.ok(ui.dueEl.textContent.startsWith("Letzte Anfrage:"), "due line refreshed on change");
  }

  console.log("options.test.js: all assertions passed");
})();
