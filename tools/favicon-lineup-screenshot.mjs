import { chromium } from 'playwright';
import { fileURLToPath } from 'url';
import { dirname, resolve } from 'path';

const __dirname = dirname(fileURLToPath(import.meta.url));

const browser = await chromium.launch();
const page = await browser.newPage();

await page.setViewportSize({ width: 2800, height: 1300 });
await page.goto(`file://${resolve(__dirname, 'favicon-lineup.html')}`);

// wait for fonts to load
await page.waitForTimeout(3000);

await page.screenshot({
  path: resolve(__dirname, '..', 'favicon-lineup.png'),
  omitBackground: false,
});

console.log('saved favicon-lineup.png');
await browser.close();
