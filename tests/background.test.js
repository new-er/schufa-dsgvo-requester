const assert = require("node:assert");

const store = {};
let alarm = null;
const listeners = {};
const icons = [];

globalThis.browser = {
  storage: {
    local: { get: async (keys) => Object.fromEntries(keys.filter((k) => k in store).map((k) => [k, store[k]])) },
    onChanged: { addListener: (fn) => (listeners.storage = fn) },
  },
  alarms: {
    clear: async () => (alarm = null),
    create: (name, options) => (alarm = { name, options }),
    onAlarm: { addListener: () => {} },
  },
  runtime: {
    onInstalled: { addListener: () => {} },
    onStartup: { addListener: () => {} },
    getURL: (p) => p,
  },
  notifications: { create: () => {}, onClicked: { addListener: () => {} } },
  tabs: { create: () => {} },
  action: { setIcon: (o) => icons.push(o) },
};

require("../background.js");

async function resync(changes) {
  listeners.storage(changes, "local");
  await new Promise((r) => setTimeout(r, 0));
  return alarm;
}

(async () => {
  const due = Date.now() + 1000;

  assert.strictEqual(await resync({ next_due: { newValue: due } }), null, "no next_due -> no alarm");

  store.next_due = due;
  store.notify_enabled = true;
  store.notify_repeat = "once";
  assert.deepStrictEqual((await resync({ notify_enabled: {} })).options, { when: due }, "once -> single fire at due");

  store.notify_repeat = "daily";
  assert.deepStrictEqual((await resync({ notify_repeat: {} })).options, { when: due, periodInMinutes: 1440 }, "daily -> period after due");

  store.notify_repeat = "hourly";
  assert.deepStrictEqual((await resync({ notify_repeat: {} })).options, { when: due, periodInMinutes: 60 }, "hourly -> period after due");

  store.notify_enabled = false;
  assert.strictEqual(await resync({ notify_enabled: {} }), null, "disabled -> no alarm");

  assert.strictEqual(icons.at(-1).path, "icon-gray.png", "future due -> gray icon");
  store.next_due = Date.now() - 1000;
  await resync({ next_due: {} });
  assert.strictEqual(icons.at(-1).path, "icon-blue.png", "past due -> blue icon");

  console.log("background.test.js: all assertions passed");
})();
