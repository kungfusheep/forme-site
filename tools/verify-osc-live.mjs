import { chromium } from 'playwright';
const browser = await chromium.launch();
const page = await browser.newPage();
const errors = [];
page.on('pageerror', e => errors.push(String(e)));
page.on('console', m => { if (m.type() === 'error') errors.push(m.text()); });
await page.goto('http://localhost:8080/index.html', { waitUntil: 'networkidle' });
const osc = page.locator('canvas[data-termsrc]');
await osc.scrollIntoViewIfNeeded();
await page.waitForTimeout(600);
const snap = () => osc.evaluate(c => c.toDataURL());
const a = await snap();
await page.waitForTimeout(250); // > 1 frame at 15fps
const b = await snap();
const reg = await page.locator('canvas[data-term="registration-form"]').evaluate(c => c.width > 0 && c.height > 0);
console.log('osc canvas painted:', a.length > 5000);
console.log('osc canvas animating:', a !== b);
console.log('registration canvas ok:', reg);
// scroll away and confirm the ticker stops
await page.evaluate(() => window.scrollTo(0, 0));
await page.waitForTimeout(400);
const c1 = await snap();
await page.waitForTimeout(250);
const c2 = await snap();
console.log('stops offscreen:', c1 === c2);
console.log('page errors:', errors.length ? errors : 'none');
await browser.close();
