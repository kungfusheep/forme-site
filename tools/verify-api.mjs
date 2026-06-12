import { chromium } from 'playwright';
const browser = await chromium.launch();
const page = await browser.newPage();
const errors = [];
page.on('pageerror', e => errors.push(String(e)));
await page.goto('http://localhost:8080/api.html', { waitUntil: 'networkidle' });
await page.waitForTimeout(800);
const html = await page.evaluate(() => document.body.innerHTML);
for (const sym of ['Osc', 'Trigger', 'RequestAnimation', 'OnFilterChange', 'JumpItem', 'CharWrap', 'Textf', 'Fps', 'Sine', 'Steps']) {
  console.log(sym, html.includes(sym) ? 'IN-DOM' : 'MISSING');
}
console.log('page errors:', errors.length ? errors : 'none');
await browser.close();
