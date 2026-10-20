const api = globalThis.browser ?? chrome;
const ALARM = "schufa-due";
const NOTIFICATION = "schufa-due";

const REPEAT_PERIODS = { hourly: 60, daily: 1440, weekly: 10080 };

async function syncAlarm() {
  const { next_due, notify_enabled = true, notify_repeat = "once" } =
    await api.storage.local.get(["next_due", "notify_enabled", "notify_repeat"]);
  await api.alarms.clear(ALARM);
  if (!next_due || !notify_enabled) return;
  const options = { when: next_due };
  if (REPEAT_PERIODS[notify_repeat]) {
    options.periodInMinutes = REPEAT_PERIODS[notify_repeat];
  }
  api.alarms.create(ALARM, options);
}

async function syncIcon() {
  const { next_due } = await api.storage.local.get(["next_due"]);
  const due = next_due && Date.now() >= next_due;
  await api.action.setIcon({ path: due ? "icon-blue.png" : "icon-gray.png" });
}

const refresh = () => {
  syncAlarm();
  syncIcon();
};
api.runtime.onInstalled.addListener(refresh);
api.runtime.onStartup.addListener(refresh);
api.storage.onChanged.addListener((changes, area) => {
  if (area !== "local") return;
  if (["next_due", "last_sent"].some((key) => key in changes)) {
    syncIcon();
  }
  if (["next_due", "notify_enabled", "notify_repeat"].some((key) => key in changes)) {
    syncAlarm();
  }
});

api.alarms.onAlarm.addListener(async (alarm) => {
  if (alarm.name !== ALARM) return;
  await syncIcon();
  await api.notifications.create(NOTIFICATION, {
    type: "basic",
    iconUrl: "icon-blue.png",
    title: "SCHUFA DSGVO Auskunft fällig",
    message: "Zeit für ein neues Auskunftsersuchen nach Art. 15 DSGVO.",
  });
});

api.notifications.onClicked.addListener((id) => {
  if (id !== NOTIFICATION) return;
  api.tabs.create({ url: api.runtime.getURL("compose.html") });
});
