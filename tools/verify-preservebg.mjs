import { chromium } from 'playwright';
const browser = await chromium.launch();
const page = await browser.newPage();
const errors = [];
page.on('pageerror', e => errors.push(String(e)));
await page.goto('http://localhost:8080/api.html', { waitUntil: 'networkidle' });
await page.waitForTimeout(700);
const html = await page.evaluate(() => document.body.innerHTML);
for (const s of ['PreserveBG', 'AttrPreserveBG', 'Apply', 'PreserveBg']) {
  console.log(s, html.includes(s) ? 'IN-DOM' : 'absent');
}
console.log('page errors:', errors.length ? errors : 'none');
await browser.close();
