import { chromium } from 'playwright';
import { fileURLToPath } from 'url';
import { dirname, resolve } from 'path';

const __dirname = dirname(fileURLToPath(import.meta.url));

const browser = await chromium.launch();
const context = await browser.newContext({ deviceScaleFactor: 2 });
const page = await context.newPage();

await page.setViewportSize({ width: 1200, height: 630 });
await page.goto(`file://${resolve(__dirname, 'og-card.html')}`);

// wait for fonts (typekit + local woff2)
await page.waitForTimeout(3000);

await page.screenshot({
  path: resolve(__dirname, '..', 'og.png'),
  omitBackground: false,
});

console.log('saved og.png');
await browser.close();
