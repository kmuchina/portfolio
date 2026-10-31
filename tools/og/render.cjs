#!/usr/bin/env node
// render.cjs: re-renders the committed bitmap assets with a headless
// Chromium (Playwright). Not part of the Go build, which stays Node-free:
// run it by hand (`make og`) after changing design/og.html or design/icons.html.
//
//   design/og.html     → static/og.png             (1200 × 630)
//   design/icons.html  → static/apple-touch-icon.png (180 × 180)
//                        static/favicon.ico         (32 × 32, PNG in an ICO)
//
// Playwright is looked up as `playwright`, then `playwright-core`, or at the
// path in $PLAYWRIGHT_MODULE (e.g. another project's node_modules/playwright-core).
"use strict";

const fs = require("fs");
const path = require("path");
const { pathToFileURL } = require("url");

function loadPlaywright() {
  for (const name of [process.env.PLAYWRIGHT_MODULE, "playwright", "playwright-core"]) {
    if (!name) continue;
    try { return require(name); } catch {}
  }
  return null;
}

// An ICO file holding one PNG image (supported by every current browser).
function ico(png, size) {
  const header = Buffer.alloc(6 + 16);
  header.writeUInt16LE(0, 0); // reserved
  header.writeUInt16LE(1, 2); // type: icon
  header.writeUInt16LE(1, 4); // one image
  header.writeUInt8(size % 256, 6); // width (0 = 256)
  header.writeUInt8(size % 256, 7); // height
  header.writeUInt8(0, 8); // palette
  header.writeUInt8(0, 9); // reserved
  header.writeUInt16LE(1, 10); // colour planes
  header.writeUInt16LE(32, 12); // bits per pixel
  header.writeUInt32LE(png.length, 14);
  header.writeUInt32LE(22, 18); // offset of the image data
  return Buffer.concat([header, png]);
}

async function main() {
  const pw = loadPlaywright();
  if (!pw) {
    console.error(`make og needs Playwright with Chromium, which isn't installed here.
Either:
  npm install -g playwright && npx playwright install chromium
  make og
or point at an existing copy:
  PLAYWRIGHT_MODULE=/path/to/node_modules/playwright-core make og
Nothing was changed: static/og.png and the icons stay as committed.`);
    process.exit(1);
  }
  const root = path.resolve(__dirname, "..", "..");
  const url = (f) => pathToFileURL(path.join(root, "design", f)).href;
  const out = (f) => path.join(root, "static", f);

  const browser = await pw.chromium.launch();
  try {
    const og = await browser.newPage({ viewport: { width: 1200, height: 630 }, deviceScaleFactor: 1 });
    await og.goto(url("og.html"));
    await og.evaluate(() => document.fonts.ready);
    await og.screenshot({ path: out("og.png") });

    const icons = await browser.newPage({ viewport: { width: 400, height: 300 }, deviceScaleFactor: 1 });
    await icons.goto(url("icons.html"));
    await icons.evaluate(() => document.fonts.ready);
    await icons.locator("#touch").screenshot({ path: out("apple-touch-icon.png") });
    const png32 = await icons.locator("#ico").screenshot({ omitBackground: true });
    fs.writeFileSync(out("favicon.ico"), ico(png32, 32));
  } finally {
    await browser.close();
  }
  for (const f of ["og.png", "apple-touch-icon.png", "favicon.ico"]) {
    console.log(`static/${f}: ${(fs.statSync(out(f)).size / 1024).toFixed(1)} KB`);
  }
}

main().catch((err) => { console.error(err); process.exit(1); });
