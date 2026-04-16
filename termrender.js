// terminal canvas renderer — parses ANSI escape codes and paints to <canvas>
// usage: <canvas data-term="img/example.json"></canvas>
(function () {
  const BG = '#1a1a18'
  const CHROME_BG = '#222220'
  const DEFAULT_FG = '#c8c4b8'
  const BORDER_COLOR = '#2a2a28'
  const FONT = "'Berkeley Mono', monospace"
  const FONT_SIZE = 14
  const PAD = 16
  const CHROME_H = 40
  const DOTS = [['#c44040', 20], ['#c4a040', 38], ['#6a9a50', 56]]
  const SHADOW_PAD = 32
  const shadowCache = new Map()

  // standard ANSI 256 color palette (first 16)
  const ANSI = [
    '#131311', '#d47080', '#8a9a78', '#c4a06a',
    '#7a9aba', '#a088b8', '#7a9a9a', '#c8c4b8',
    '#444444', '#d47080', '#9aaa88', '#d4b87a',
    '#8aaaca', '#b098c8', '#8aaaba', '#e8e4d8',
  ]

  function ansiToHex(n) {
    if (n < 16) return ANSI[n]
    if (n < 232) {
      n -= 16
      const b = (n % 6) * 51
      const g = ((n / 6 | 0) % 6) * 51
      const r = ((n / 36 | 0) % 6) * 51
      return '#' + [r, g, b].map(v => v.toString(16).padStart(2, '0')).join('')
    }
    const v = (n - 232) * 10 + 8
    return '#' + v.toString(16).padStart(2, '0').repeat(3)
  }

  // parse a single ANSI-styled line into [{ch, fg, bg, bold, dim, inverse}, ...]
  function parseLine(str) {
    const cells = []
    let fg = null, bg = null, bold = false, dim = false, italic = false, underline = false, inverse = false
    let i = 0

    while (i < str.length) {
      if (str[i] === '\x1b' && str[i + 1] === '[') {
        let j = i + 2
        while (j < str.length && str[j] !== 'm') j++
        const params = str.slice(i + 2, j).split(';').map(Number)
        i = j + 1

        for (let p = 0; p < params.length; p++) {
          const v = params[p]
          if (v === 0) { fg = null; bg = null; bold = false; dim = false; italic = false; underline = false; inverse = false }
          else if (v === 1) bold = true
          else if (v === 2) dim = true
          else if (v === 3) italic = true
          else if (v === 4) underline = true
          else if (v === 7) inverse = true
          else if (v === 23) italic = false
          else if (v === 24) underline = false
          else if (v === 27) inverse = false
          else if (v >= 30 && v <= 37) fg = ansiToHex(v - 30)
          else if (v >= 40 && v <= 47) bg = ansiToHex(v - 40)
          else if (v >= 90 && v <= 97) fg = ansiToHex(v - 90 + 8)
          else if (v >= 100 && v <= 107) bg = ansiToHex(v - 100 + 8)
          else if (v === 38 && params[p + 1] === 5) { fg = ansiToHex(params[p + 2]); p += 2 }
          else if (v === 48 && params[p + 1] === 5) { bg = ansiToHex(params[p + 2]); p += 2 }
          else if (v === 38 && params[p + 1] === 2) {
            fg = '#' + [params[p+2], params[p+3], params[p+4]].map(x => x.toString(16).padStart(2, '0')).join('')
            p += 4
          }
          else if (v === 48 && params[p + 1] === 2) {
            bg = '#' + [params[p+2], params[p+3], params[p+4]].map(x => x.toString(16).padStart(2, '0')).join('')
            p += 4
          }
          else if (v === 39) fg = null
          else if (v === 49) bg = null
        }
        continue
      }

      // apply inverse: swap fg/bg
      let cellFG = fg, cellBG = bg
      if (inverse) {
        cellFG = bg || BG
        cellBG = fg || DEFAULT_FG
      }

      cells.push({ ch: str[i], fg: cellFG, bg: cellBG, bold, dim, italic, underline })
      i++
    }

    return cells
  }

  function render(canvas, data) {
    const ctx = canvas.getContext('2d')
    const dpr = window.devicePixelRatio || 1

    // measure character width with the actual font
    ctx.font = FONT_SIZE + 'px ' + FONT
    const charW = ctx.measureText('M').width
    const charH = FONT_SIZE * 1.5

    const totalW = data.w * charW + PAD * 2
    const totalH = data.h * charH + PAD * 2 + CHROME_H
    const canvasW = totalW + SHADOW_PAD * 2
    const canvasH = totalH + SHADOW_PAD * 2

    canvas.width = canvasW * dpr
    canvas.height = canvasH * dpr
    canvas.style.maxWidth = canvasW + 'px'
    canvas.style.width = '100%'
    canvas.style.aspectRatio = (canvasW / canvasH).toFixed(4)
    ctx.scale(dpr, dpr)
    ctx.translate(SHADOW_PAD, SHADOW_PAD)

    // macOS-style drop shadow — single pass, cached per size
    const shadowKey = totalW + 'x' + totalH
    let shadowCanvas = shadowCache.get(shadowKey)
    if (!shadowCanvas) {
      shadowCanvas = document.createElement('canvas')
      const pad = SHADOW_PAD * dpr
      shadowCanvas.width = canvasW * dpr
      shadowCanvas.height = canvasH * dpr
      const sctx = shadowCanvas.getContext('2d')
      sctx.scale(dpr, dpr)
      sctx.translate(SHADOW_PAD, SHADOW_PAD)
      // layer 1: contact
      sctx.save()
      sctx.shadowColor = 'rgba(0,0,0,0.3)'
      sctx.shadowBlur = 6
      sctx.shadowOffsetY = 2
      roundRect(sctx, 0, 0, totalW, totalH, 10)
      sctx.fillStyle = BG
      sctx.fill()
      sctx.restore()
      // layer 2: diffuse
      sctx.save()
      sctx.shadowColor = 'rgba(0,0,0,0.15)'
      sctx.shadowBlur = 20
      sctx.shadowOffsetY = 8
      roundRect(sctx, 0, 0, totalW, totalH, 10)
      sctx.fillStyle = BG
      sctx.fill()
      sctx.restore()
      // layer 3: ambient
      sctx.save()
      sctx.shadowColor = 'rgba(0,0,0,0.08)'
      sctx.shadowBlur = 40
      sctx.shadowOffsetY = 12
      roundRect(sctx, 0, 0, totalW, totalH, 10)
      sctx.fillStyle = BG
      sctx.fill()
      sctx.restore()
      shadowCache.set(shadowKey, shadowCanvas)
    }
    ctx.save()
    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.drawImage(shadowCanvas, 0, 0)
    ctx.restore()

    // chrome bar
    roundRect(ctx, 0, 0, totalW, CHROME_H, 10)
    ctx.fillStyle = CHROME_BG
    ctx.fill()
    ctx.fillRect(0, 30, totalW, 10)

    // dots
    for (const [color, cx] of DOTS) {
      ctx.fillStyle = color
      ctx.beginPath()
      ctx.arc(cx, 20, 6, 0, Math.PI * 2)
      ctx.fill()
    }

    // separator
    ctx.strokeStyle = BORDER_COLOR
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(0, CHROME_H)
    ctx.lineTo(totalW, CHROME_H)
    ctx.stroke()

    // render lines
    for (let y = 0; y < data.lines.length; y++) {
      const cells = parseLine(data.lines[y])

      for (let x = 0; x < cells.length; x++) {
        const cell = cells[x]
        const px = PAD + x * charW
        const py = PAD + CHROME_H + y * charH

        if (cell.bg) {
          ctx.fillStyle = cell.bg
          ctx.fillRect(px, py, charW + 0.5, charH)
        }

        if (cell.ch && cell.ch !== ' ') {
          const color = cell.fg || DEFAULT_FG
          if (cell.dim) ctx.globalAlpha = 0.5

          if (drawBoxChar(ctx, cell.ch, px, py, charW, charH, color)) {
            // drawn via primitives
          } else {
            ctx.font = (cell.italic ? 'italic ' : '') + (cell.bold ? 'bold ' : '') + FONT_SIZE + 'px ' + FONT
            ctx.fillStyle = color
            ctx.fillText(cell.ch, px, py + charH * 0.75)
          }

          if (cell.underline) {
            ctx.fillStyle = cell.fg || DEFAULT_FG
            ctx.fillRect(px, py + charH * 0.85, charW, 1)
          }

          if (cell.dim) ctx.globalAlpha = 1
        }
      }
    }

    // CSS mask fades the canvas to transparent at the bottom of the terminal window
    const fadeStart = ((SHADOW_PAD + totalH * 0.75) / canvasH * 100).toFixed(1)
    const fadeEnd = ((SHADOW_PAD + totalH * 0.98) / canvasH * 100).toFixed(1)
    canvas.style.maskImage = `linear-gradient(to bottom, black ${fadeStart}%, transparent ${fadeEnd}%)`
    canvas.style.webkitMaskImage = `linear-gradient(to bottom, black ${fadeStart}%, transparent ${fadeEnd}%)`
  }

  // draw box-drawing characters with canvas primitives for pixel-perfect alignment
  function drawBoxChar(ctx, ch, x, y, w, h, color) {
    const cx = Math.round(x + w / 2) + 0.5
    const cy = Math.round(y + h / 2) + 0.5
    const lw = 1

    ctx.strokeStyle = color
    ctx.fillStyle = color
    ctx.lineWidth = lw
    ctx.lineCap = 'square'

    switch (ch) {
      // straight lines
      case '─': line(ctx, x, cy, x + w, cy); return true
      case '│': line(ctx, cx, y, cx, y + h); return true
      case '━': ctx.lineWidth = 2; line(ctx, x, cy, x + w, cy); return true
      case '┃': ctx.lineWidth = 2; line(ctx, cx, y, cx, y + h); return true

      // rounded corners — connect midpoints of edges with a quarter-circle
      case '╭': corner(ctx, cx, cy, x + w, cy, cx, y + h, 3); return true
      case '╮': corner(ctx, cx, cy, x, cy, cx, y + h, 3); return true
      case '╰': corner(ctx, cx, cy, x + w, cy, cx, y, 3); return true
      case '╯': corner(ctx, cx, cy, x, cy, cx, y, 3); return true

      // sharp corners
      case '┌': line(ctx, cx, cy, x + w, cy); line(ctx, cx, cy, cx, y + h); return true
      case '┐': line(ctx, x, cy, cx, cy); line(ctx, cx, cy, cx, y + h); return true
      case '└': line(ctx, cx, cy, x + w, cy); line(ctx, cx, y, cx, cy); return true
      case '┘': line(ctx, x, cy, cx, cy); line(ctx, cx, y, cx, cy); return true

      // T-junctions
      case '├': line(ctx, cx, y, cx, y + h); line(ctx, cx, cy, x + w, cy); return true
      case '┤': line(ctx, cx, y, cx, y + h); line(ctx, x, cy, cx, cy); return true
      case '┬': line(ctx, x, cy, x + w, cy); line(ctx, cx, cy, cx, y + h); return true
      case '┴': line(ctx, x, cy, x + w, cy); line(ctx, cx, y, cx, cy); return true
      case '┼': line(ctx, x, cy, x + w, cy); line(ctx, cx, y, cx, y + h); return true

      // double lines
      case '═': ctx.lineWidth = 1; line(ctx, x, cy - 2, x + w, cy - 2); line(ctx, x, cy + 2, x + w, cy + 2); return true
      case '║': ctx.lineWidth = 1; line(ctx, cx - 2, y, cx - 2, y + h); line(ctx, cx + 2, y, cx + 2, y + h); return true
      case '╔': line(ctx, cx - 2, cy - 2, x + w, cy - 2); line(ctx, cx + 2, cy + 2, x + w, cy + 2); line(ctx, cx - 2, cy - 2, cx - 2, y + h); line(ctx, cx + 2, cy + 2, cx + 2, y + h); return true
      case '╗': line(ctx, x, cy - 2, cx + 2, cy - 2); line(ctx, x, cy + 2, cx - 2, cy + 2); line(ctx, cx + 2, cy - 2, cx + 2, y + h); line(ctx, cx - 2, cy + 2, cx - 2, y + h); return true
      case '╚': line(ctx, cx - 2, y, cx - 2, cy + 2); line(ctx, cx + 2, y, cx + 2, cy - 2); line(ctx, cx - 2, cy + 2, x + w, cy + 2); line(ctx, cx + 2, cy - 2, x + w, cy - 2); return true
      case '╝': line(ctx, cx + 2, y, cx + 2, cy + 2); line(ctx, cx - 2, y, cx - 2, cy - 2); line(ctx, x, cy + 2, cx + 2, cy + 2); line(ctx, x, cy - 2, cx - 2, cy - 2); return true

      // horizontal rule
      case '—': line(ctx, x, cy, x + w, cy); return true
      case '–': line(ctx, x, cy, x + w, cy); return true

      // block elements — vertical fills (bottom up)
      case '▁': ctx.fillStyle = color; ctx.fillRect(x, y + h * 7/8, w + 0.5, h * 1/8 + 0.5); return true
      case '▂': ctx.fillStyle = color; ctx.fillRect(x, y + h * 3/4, w + 0.5, h * 1/4 + 0.5); return true
      case '▃': ctx.fillStyle = color; ctx.fillRect(x, y + h * 5/8, w + 0.5, h * 3/8 + 0.5); return true
      case '▄': ctx.fillStyle = color; ctx.fillRect(x, y + h / 2, w + 0.5, h / 2 + 0.5); return true
      case '▅': ctx.fillStyle = color; ctx.fillRect(x, y + h * 3/8, w + 0.5, h * 5/8 + 0.5); return true
      case '▆': ctx.fillStyle = color; ctx.fillRect(x, y + h * 1/4, w + 0.5, h * 3/4 + 0.5); return true
      case '▇': ctx.fillStyle = color; ctx.fillRect(x, y + h * 1/8, w + 0.5, h * 7/8 + 0.5); return true
      case '█': ctx.fillStyle = color; ctx.fillRect(x, y, w + 0.5, h + 0.5); return true

      // block elements — top fills (top down)
      case '▀': ctx.fillStyle = color; ctx.fillRect(x, y, w + 0.5, h / 2 + 0.5); return true

      // block elements — horizontal fills (left to right)
      case '▏': ctx.fillStyle = color; ctx.fillRect(x, y, w * 1/8 + 0.5, h + 0.5); return true
      case '▎': ctx.fillStyle = color; ctx.fillRect(x, y, w * 1/4 + 0.5, h + 0.5); return true
      case '▍': ctx.fillStyle = color; ctx.fillRect(x, y, w * 3/8 + 0.5, h + 0.5); return true
      case '▌': ctx.fillStyle = color; ctx.fillRect(x, y, w / 2 + 0.5, h + 0.5); return true
      case '▋': ctx.fillStyle = color; ctx.fillRect(x, y, w * 5/8 + 0.5, h + 0.5); return true
      case '▊': ctx.fillStyle = color; ctx.fillRect(x, y, w * 3/4 + 0.5, h + 0.5); return true
      case '▉': ctx.fillStyle = color; ctx.fillRect(x, y, w * 7/8 + 0.5, h + 0.5); return true

      // block elements — right fills
      case '▐': ctx.fillStyle = color; ctx.fillRect(x + w / 2, y, w / 2 + 0.5, h + 0.5); return true

      // shade blocks
      case '░': ctx.fillStyle = color; ctx.globalAlpha = 0.25; ctx.fillRect(x, y, w + 0.5, h + 0.5); ctx.globalAlpha = 1; return true
      case '▒': ctx.fillStyle = color; ctx.globalAlpha = 0.5; ctx.fillRect(x, y, w + 0.5, h + 0.5); ctx.globalAlpha = 1; return true
      case '▓': ctx.fillStyle = color; ctx.globalAlpha = 0.75; ctx.fillRect(x, y, w + 0.5, h + 0.5); ctx.globalAlpha = 1; return true

      default: return false
    }
  }

  function line(ctx, x1, y1, x2, y2) {
    ctx.beginPath()
    ctx.moveTo(x1, y1)
    ctx.lineTo(x2, y2)
    ctx.stroke()
  }

  // draw a rounded corner: line from hx,hy to cx,cy (horizontal), then cx,cy to vx,vy (vertical)
  // with a rounded join at cx,cy
  function corner(ctx, cx, cy, hx, hy, vx, vy, r) {
    ctx.beginPath()
    ctx.moveTo(hx, hy)
    ctx.arcTo(cx, cy, vx, vy, r)
    ctx.lineTo(vx, vy)
    ctx.stroke()
  }

  function roundRect(ctx, x, y, w, h, r) {
    ctx.beginPath()
    ctx.moveTo(x + r, y)
    ctx.lineTo(x + w - r, y)
    ctx.arcTo(x + w, y, x + w, y + r, r)
    ctx.lineTo(x + w, y + h - r)
    ctx.arcTo(x + w, y + h, x + w - r, y + h, r)
    ctx.lineTo(x + r, y + h)
    ctx.arcTo(x, y + h, x, y + h - r, r)
    ctx.lineTo(x, y + r)
    ctx.arcTo(x, y, x + r, y, r)
    ctx.closePath()
  }

  // cache loaded data so we can re-render on DPR change
  const loaded = new Map()

  function init() {
    const canvases = document.querySelectorAll('canvas[data-termdata]')
    for (const canvas of canvases) {
      const data = JSON.parse(canvas.getAttribute('data-termdata'))
      loaded.set(canvas, data)
      render(canvas, data)
    }
  }

  // re-render all canvases when pixel density changes (pinch zoom, moving between displays)
  function watchDPR() {
    const mq = matchMedia(`(resolution: ${devicePixelRatio}dppx)`)
    mq.addEventListener('change', () => {
      shadowCache.clear()
      for (const [canvas, data] of loaded) {
        render(canvas, data)
      }
      watchDPR()
    }, { once: true })
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => { init(); watchDPR() })
  } else {
    init(); watchDPR()
  }

  function renderOne(canvas) {
    const data = JSON.parse(canvas.getAttribute('data-termdata'))
    loaded.set(canvas, data)
    render(canvas, data)
  }

  window.termRender = { init, renderOne }
})()
