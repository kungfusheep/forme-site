import { chromium } from 'playwright';
const browser = await chromium.launch();
const page = await browser.newPage({ deviceScaleFactor: 2 });
await page.setViewportSize({ width: 1200, height: 800 });
const errors = [];
page.on('pageerror', e => errors.push(String(e)));
await page.goto('http://localhost:8080/index.html', { waitUntil: 'networkidle' });
await page.waitForTimeout(300);

const sample = async () => page.evaluate(() => {
  const el = document.querySelector('.hero-title');
  const cs = getComputedStyle(el);
  const r = el.getBoundingClientRect();
  return { fvs: cs.fontVariationSettings, w: Math.round(r.width*100)/100, h: Math.round(r.height*100)/100 };
});

// sample across the 7s cycle: rest(0), inhale-peak(~2.94s=42%), exhale-dip(~5.04s=72%)
const t0 = await sample();
await page.waitForTimeout(2900); const tPeak = await sample();
await page.waitForTimeout(2100); const tDip = await sample();

console.log('rest :', JSON.stringify(t0));
console.log('peak :', JSON.stringify(tPeak));
console.log('dip  :', JSON.stringify(tDip));
const widths = [t0.w, tPeak.w, tDip.w];
console.log('width stable (no reflow):', Math.max(...widths) - Math.min(...widths) < 0.5, '(spread', (Math.max(...widths)-Math.min(...widths)).toFixed(2)+'px)');
console.log('weight actually moves:', new Set([t0.fvs, tPeak.fvs, tDip.fvs]).size > 1);
console.log('page errors:', errors.length ? errors : 'none');

// reduced-motion check
const page2 = await browser.newPage();
await page2.emulateMedia({ reducedMotion: 'reduce' });
await page2.goto('http://localhost:8080/index.html', { waitUntil: 'networkidle' });
const rmA = await page2.evaluate(() => getComputedStyle(document.querySelector('.hero-title')).fontVariationSettings);
await page2.waitForTimeout(1500);
const rmB = await page2.evaluate(() => getComputedStyle(document.querySelector('.hero-title')).fontVariationSettings);
console.log('reduced-motion still (no change):', rmA === rmB, `(${rmA})`);

await browser.close();
