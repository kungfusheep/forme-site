import { chromium } from 'playwright';

// renders each SVG in a browser context with Berkeley Mono loaded,
// captures as 2x PNG for crisp display on retina
const browser = await chromium.launch();
const page = await browser.newPage({ deviceScaleFactor: 2 });
await page.setViewportSize({ width: 1200, height: 900 });

// load a page that has the Berkeley Mono font-face defined
await page.goto('http://localhost:8080/poc-svg.html', { waitUntil: 'networkidle' });
await page.waitForTimeout(500);

const names = ['hero', 'file-browser', 'process-monitor', 'deploy-log', 'fuzzy-finder', 'live-dashboard', 'registration-form'];

for (const name of names) {
  // inject each SVG into the page inside a sized container, screenshot it
  const svgPath = `http://localhost:8080/img/${name}.svg`;
  const resp = await page.evaluate(async (url) => {
    const r = await fetch(url);
    return r.text();
  }, svgPath);

  // extract viewBox dimensions to set explicit width/height
  const vbMatch = resp.match(/viewBox="0 0 ([\d.]+) ([\d.]+)"/);
  const svgW = vbMatch ? vbMatch[1] : '600';
  const svgH = vbMatch ? vbMatch[2] : '400';
  const sized = resp.replace('<svg ', `<svg width="${svgW}" height="${svgH}" `);

  await page.setContent(`
    <!DOCTYPE html>
    <html>
    <head>
    <style>
      @font-face {
        font-family: 'Berkeley Mono';
        src: url('http://localhost:8080/BerkeleyMonoVariable.woff2') format('woff2');
        font-weight: 100 900;
        font-display: swap;
      }
      * { margin: 0; padding: 0; }
      body { background: transparent; display: inline-block; }
    </style>
    </head>
    <body><div class="wrap">${sized}</div></body>
    </html>
  `, { waitUntil: 'networkidle' });

  await page.waitForTimeout(300);

  const wrap = await page.$('.wrap');
  await wrap.screenshot({
    path: `img/${name}.png`,
    omitBackground: true,
  });
  console.log(`captured img/${name}.png`);
}

await browser.close();
console.log('done');
