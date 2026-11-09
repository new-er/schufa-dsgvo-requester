#!/usr/bin/env node
"use strict";

// Publishes a zipped extension to the Chrome Web Store using the Chrome Web
// Store API. No external dependencies: signs a JWT with the built-in `crypto`
// module to obtain an OAuth2 access token from a Google service account.
//
// Required env vars:
//   CHROME_WEBSTORE_SERVICE_ACCOUNT  full JSON service-account key
//   CHROME_WEBSTORE_EXTENSION_ID     Chrome Web Store item id (from the store URL)
//   CHROME_WEBSTORE_ZIP              path to the extension zip (default: dist/schufa-dsgvo-requester.zip)
//   CHROME_WEBSTORE_PUBLISH_TARGET   default | group (default: default)

const { createSign } = require("node:crypto");
const { readFileSync } = require("node:fs");
const { request } = require("node:https");

const SERVICE_ACCOUNT = JSON.parse(process.env.CHROME_WEBSTORE_SERVICE_ACCOUNT || "{}");
const EXTENSION_ID = process.env.CHROME_WEBSTORE_EXTENSION_ID;
const ZIP_PATH = process.env.CHROME_WEBSTORE_ZIP || "dist/schufa-dsgvo-requester.zip";
const PUBLISH_TARGET = process.env.CHROME_WEBSTORE_PUBLISH_TARGET || "default";

if (!SERVICE_ACCOUNT.private_key || !SERVICE_ACCOUNT.client_email) {
  console.error("Missing/invalid CHROME_WEBSTORE_SERVICE_ACCOUNT");
  process.exit(1);
}
if (!EXTENSION_ID) {
  console.error("Missing CHROME_WEBSTORE_EXTENSION_ID");
  process.exit(1);
}

function base64url(input) {
  return Buffer.from(input)
    .toString("base64")
    .replace(/\+/g, "-")
    .replace(/\//g, "_")
    .replace(/=+$/, "");
}

function signJwt() {
  const now = Math.floor(Date.now() / 1000);
  const header = base64url(JSON.stringify({ alg: "RS256", typ: "JWT" }));
  const payload = base64url(
    JSON.stringify({
      iss: SERVICE_ACCOUNT.client_email,
      scope: "https://www.googleapis.com/auth/chromewebstore",
      aud: "https://oauth2.googleapis.com/token",
      exp: now + 3600,
      iat: now,
    })
  );
  const signingInput = `${header}.${payload}`;
  const signature = createSign("RSA-SHA256")
    .update(signingInput)
    .sign(SERVICE_ACCOUNT.private_key, "base64");
  return `${signingInput}.${base64url(signature)}`;
}

function fetchJson(options, body) {
  return new Promise((resolve, reject) => {
    const req = request(options, (res) => {
      let data = "";
      res.on("data", (c) => (data += c));
      res.on("end", () => resolve({ status: res.statusCode, body: data }));
    });
    req.on("error", reject);
    req.end(body);
  });
}

async function getAccessToken(jwt) {
  const res = await fetchJson(
    {
      hostname: "oauth2.googleapis.com",
      path: "/token",
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
    },
    `grant_type=urn:ietf:params:oauth:grant-type:jwt-bearer&assertion=${encodeURIComponent(jwt)}`
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
  const jwt = signJwt();
  const token = await getAccessToken(jwt);
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
