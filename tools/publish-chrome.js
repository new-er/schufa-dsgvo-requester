#!/usr/bin/env node
"use strict";

// Publishes a zipped extension to the Chrome Web Store using the Chrome Web
// Store API. No external dependencies: obtains an OAuth2 access token via the
// refresh-token grant, then uploads the zip and publishes it.
//
// Required env vars:
//   CHROME_WEBSTORE_CLIENT_ID        OAuth2 web-app client id
//   CHROME_WEBSTORE_CLIENT_SECRET    OAuth2 web-app client secret
//   CHROME_WEBSTORE_REFRESH_TOKEN    refresh token from one-time OAuth2 flow
//   CHROME_WEBSTORE_EXTENSION_ID     Chrome Web Store item id (from the store URL)
//   CHROME_WEBSTORE_ZIP              path to the extension zip (default: dist/schufa-dsgvo-requester.zip)
//   CHROME_WEBSTORE_PUBLISH_TARGET   default | group (default: default)

const { readFileSync } = require("node:fs");
const { request } = require("node:https");

const CLIENT_ID = process.env.CHROME_WEBSTORE_CLIENT_ID;
const CLIENT_SECRET = process.env.CHROME_WEBSTORE_CLIENT_SECRET;
const REFRESH_TOKEN = process.env.CHROME_WEBSTORE_REFRESH_TOKEN;
const EXTENSION_ID = process.env.CHROME_WEBSTORE_EXTENSION_ID;
const ZIP_PATH = process.env.CHROME_WEBSTORE_ZIP || "dist/schufa-dsgvo-requester.zip";
const PUBLISH_TARGET = process.env.CHROME_WEBSTORE_PUBLISH_TARGET || "default";

if (!CLIENT_ID || !CLIENT_SECRET || !REFRESH_TOKEN) {
  console.error("Missing/invalid CHROME_WEBSTORE_CLIENT_ID / CLIENT_SECRET / REFRESH_TOKEN");
  process.exit(1);
}
if (!EXTENSION_ID) {
  console.error("Missing CHROME_WEBSTORE_EXTENSION_ID");
  process.exit(1);
}

function fetchJson(options, body) {
  return new Promise((resolve, reject) => {
    const req = request(options, (res) => {
      let data = "";
      res.on("data", (c) => (data += c));
      res.on("end", () => resolve({ status: res.statusCode, body: data }));
    });
    req.on("error", reject);
    if (body !== undefined) req.write(body);
    req.end();
  });
}

async function getAccessToken() {
  const res = await fetchJson(
    {
      hostname: "oauth2.googleapis.com",
      path: "/token",
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
    },
    `grant_type=refresh_token&client_id=${encodeURIComponent(CLIENT_ID)}&client_secret=${encodeURIComponent(CLIENT_SECRET)}&refresh_token=${encodeURIComponent(REFRESH_TOKEN)}`
  );

  if (res.status !== 200) {
    throw new Error(`Failed to get access token: ${res.status} ${res.body}`);
  }
  return JSON.parse(res.body).access_token;
}

async function uploadZip(token, zip) {
  const res = await fetchJson(
    {
      hostname: "www.googleapis.com",
      path: `/upload/chromewebstore/v1.1/items/${EXTENSION_ID}`,
      method: "PUT",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/zip",
        "Content-Length": zip.length,
      },
    },
    zip
  );

  if (res.status !== 200) {
    throw new Error(`Upload failed: ${res.status} ${res.body}`);
  }
  return JSON.parse(res.body);
}

async function publish(token, target) {
  const res = await fetchJson(
    {
      hostname: "www.googleapis.com",
      path: `/chromewebstore/v1.1/items/${EXTENSION_ID}/publish`,
      method: "POST",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
    },
    JSON.stringify({ publishTarget: target })
  );

  if (res.status !== 200) {
    throw new Error(`Publish failed: ${res.status} ${res.body}`);
  }
  return JSON.parse(res.body);
}

(async () => {
  const token = await getAccessToken();
  const zip = readFileSync(ZIP_PATH);
  await uploadZip(token, zip);
  await publish(token, PUBLISH_TARGET);
  console.log(
    `Published ${ZIP_PATH} to Chrome Web Store item ${EXTENSION_ID} (target: ${PUBLISH_TARGET})`
  );
})().catch((err) => {
  console.error(err);
  process.exit(1);
});
