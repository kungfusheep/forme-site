import { chromium } from 'playwright';
import { mkdirSync } from 'fs';

const URL = 'http://localhost:8765';
const OUT = '/Users/petegriffiths/code/go/src/glyph-site/audit-screenshots';
mkdirSync(OUT, { recursive: true });

const viewports = [
  { name: 'mobile',  width: 390,  height: 844 },
  { name: 'tablet',  width: 768,  height: 1024 },
  { name: 'desktop', width: 1440, height: 900 },
];

(async () => {
  const browser = await chromium.launch();

  for (const vp of viewports) {
    const ctx = await browser.newContext({
      viewport: { width: vp.width, height: vp.height },
      deviceScaleFactor: 2,
    });
    const page = await ctx.newPage();

    // wait for fonts
    await page.goto(URL, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1500);

    // full page screenshot
    await page.screenshot({
      path: `${OUT}/${vp.name}-full.png`,
      fullPage: true,
    });

    // above-the-fold screenshot
    await page.screenshot({
      path: `${OUT}/${vp.name}-fold.png`,
      fullPage: false,
    });

    await ctx.close();
    console.log(`${vp.name}: done`);
  }

  await browser.close();
  console.log(`\nScreenshots saved to ${OUT}`);
})();
