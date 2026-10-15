const api = globalThis.browser ?? chrome;
const { renderLetter, requestViaMailto } = schufa;

let data = {};

api.storage.local.get(null).then((stored) => {
  data = stored;
  document.getElementById("letter").textContent = renderLetter(data);
});

document.getElementById("open").addEventListener("click", () => {
  requestViaMailto(api, data);
});
