const api = globalThis.browser ?? chrome;
const ALARM = "schufa-due";
const NOTIFICATION = "schufa-due";

async function syncAlarm() {
  const { next_due } = await api.storage.local.get("next_due");
  await api.alarms.clear(ALARM);
  if (next_due) {
    api.alarms.create(ALARM, { when: next_due });
  }
}

api.runtime.onInstalled.addListener(syncAlarm);
api.runtime.onStartup.addListener(syncAlarm);
api.storage.onChanged.addListener((changes, area) => {
  if (area === "local" && changes.next_due) syncAlarm();
});

api.alarms.onAlarm.addListener(async (alarm) => {
  if (alarm.name !== ALARM) return;
  await api.notifications.create(NOTIFICATION, {
    type: "basic",
    iconUrl: "icon.png",
    title: "SCHUFA DSGVO Auskunft fällig",
    message: "Zeit für ein neues Auskunftsersuchen nach Art. 15 DSGVO.",
  });
});

api.notifications.onClicked.addListener((id) => {
  if (id !== NOTIFICATION) return;
  api.tabs.create({ url: api.runtime.getURL("compose.html") });
});
