const api = globalThis.browser ?? chrome;
const { recipients, renderLetter, requestViaMailto } = schufa;

let data = {};

api.storage.local.get(null).then((stored) => {
  data = stored;
  document.getElementById("letter").textContent = renderLetter(data);
  const r = recipients(data);
  const parts = [`An: ${r.to.join(", ")}`];
  if (r.cc.length) parts.push(`CC: ${r.cc.join(", ")}`);
  if (r.bcc.length) parts.push(`BCC: ${r.bcc.join(", ")}`);
  document.getElementById("recipients").textContent = parts.join(" · ");
});

document.getElementById("open").addEventListener("click", () => {
  requestViaMailto(api, data);
});
