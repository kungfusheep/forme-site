// animated wave effect behind section headers
// auto-discovers .section-head elements inside .zone-dark or .concept containers
(function () {
  const instances = []
  let dpr = window.devicePixelRatio || 1

  const darkWaves = [
    { freq: 0.003, amp: 16, speed: 0.15,  color: 'rgba(100, 160, 130, 0.07)' },
    { freq: 0.005, amp: 12, speed: -0.1,  color: 'rgba(120, 170, 150, 0.05)' },
    { freq: 0.004, amp: 14, speed: 0.08,  color: 'rgba(90, 150, 140, 0.06)' },
  ]

  const headers = document.querySelectorAll('.zone-dark .section-head, .concept .section-head')
  if (!headers.length) return

  for (const header of headers) {
    header.style.position = 'relative'
    header.style.overflow = 'visible'
    // use a wrapper div to hold position in the flow, canvas is fixed to viewport
    const wrapper = document.createElement('div')
    wrapper.style.cssText = 'position:absolute;top:50%;left:0;width:0;height:0;overflow:visible'
    const canvas = document.createElement('canvas')
    canvas.className = 'bg-waves'
    canvas.style.cssText = 'position:fixed;left:0;height:120px;pointer-events:none;z-index:0;opacity:0;transition:opacity 0.8s'
    wrapper.appendChild(canvas)
    header.appendChild(wrapper)
    instances.push({ canvas, ctx: canvas.getContext('2d'), waves: darkWaves, h: 120, visible: false, header, wrapper })
  }

  const observer = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      const inst = instances.find(i => i.canvas === entry.target)
      if (inst) {
        inst.visible = entry.isIntersecting
        inst.canvas.style.opacity = entry.isIntersecting ? '1' : '0'
      }
    }
  }, { threshold: 0.1 })

  for (const inst of instances) observer.observe(inst.canvas)

  function resize() {
    dpr = window.devicePixelRatio || 1
    for (const inst of instances) {
      const vw = window.innerWidth
      inst.canvas.width = vw * dpr
      inst.canvas.height = inst.h * dpr
      inst.canvas.style.width = vw + 'px'
    }
  }

  function updatePositions() {
    for (const inst of instances) {
      const rect = inst.header.getBoundingClientRect()
      const midY = rect.top + rect.height / 2
      inst.canvas.style.top = (midY - inst.h / 2) + 'px'
    }
  }

  resize()
  updatePositions()
  window.addEventListener('resize', () => { resize(); updatePositions() })
  window.addEventListener('scroll', updatePositions, { passive: true })

  const features = {
    ripple: false,
    breathingWeight: true,
    dash: true,
    echo: true,
    pulse: true,
    mist: false,
    thread: false,
    fog: false,
    refract: true,
  }

  const palette = {
    dark: [[100,160,130], [120,170,150], [90,150,140]],
  }

  function waveY(x, wave, baseY, t, n, i) {
    return baseY +
      Math.sin(x * wave.freq + t * wave.speed * 0.01 + n + i) * wave.amp * dpr +
      Math.sin(x * wave.freq * 0.6 + t * wave.speed * 0.007 + n * 3 + i * 2) * wave.amp * 0.5 * dpr
  }

  // keyboard toggles (dev only — remove before shipping)
  const keyMap = {
    a: 'ripple', s: 'breathingWeight', d: 'dash', f: 'echo', g: 'pulse',
    h: 'mist', j: 'thread', k: 'fog', l: 'refract',
  }
  document.addEventListener('keydown', (e) => {
    const feat = keyMap[e.key]
    if (feat && !e.ctrlKey && !e.metaKey && !e.altKey) {
      features[feat] = !features[feat]
      console.log(Object.entries(features).map(([k,v]) => (v ? '●' : '○') + ' ' + k).join('  '))
    }
  })

  function draw(t) {
    for (let n = 0; n < instances.length; n++) {
      const inst = instances[n]
      if (!inst.visible) continue
      const { ctx, canvas, waves } = inst
      const w = canvas.width
      const h = canvas.height
      ctx.clearRect(0, 0, w, h)

      const points = []
      for (let i = 0; i < waves.length; i++) {
        const wave = waves[i]
        const baseY = h / 2 + (i - 1) * 18 * dpr
        const pts = []
        for (let x = 0; x <= w; x += 2) {
          pts.push({ x, y: waveY(x, wave, baseY, t, n, i) })
        }
        points.push(pts)
      }

      if (features.ripple) {
        for (let i = 0; i < waves.length; i++) {
          const wave = waves[i]
          const baseY = h / 2 + (i - 1) * 18 * dpr
          const c = palette.dark[i % 3]
          const color = `rgba(${c},0.05)`
          for (let r = -1; r <= 1; r += 2) {
            ctx.beginPath()
            ctx.strokeStyle = color
            ctx.lineWidth = 0.5 * dpr
            for (let x = 0; x <= w; x += 2) {
              const mainY = waveY(x, wave, baseY, t, n, i)
              const micro = Math.sin(x * 0.03 + t * 0.003 * r + i) * 3 * dpr * r
              if (x === 0) ctx.moveTo(x, mainY + micro)
              else ctx.lineTo(x, mainY + micro)
            }
            ctx.stroke()
          }
        }
      }

      if (features.echo) {
        for (let e = 1; e <= 3; e++) {
          const offsetY = e * 4 * dpr
          const alpha = (0.04 - e * 0.01)
          for (let i = 0; i < waves.length; i++) {
            const pts = points[i]
            const c = palette.dark[i % 3]
            const color = `rgba(${c},${alpha})`
            ctx.beginPath()
            ctx.strokeStyle = color
            ctx.lineWidth = 0.5 * dpr
            for (let j = 0; j < pts.length; j++) {
              if (j === 0) ctx.moveTo(pts[j].x, pts[j].y + offsetY)
              else ctx.lineTo(pts[j].x, pts[j].y + offsetY)
            }
            ctx.stroke()
          }
        }
      }

      for (let i = 0; i < waves.length; i++) {
        const wave = waves[i]
        const pts = points[i]

        if (features.breathingWeight) {
          const segLen = 4
          for (let j = 0; j < pts.length - 1; j += segLen) {
            const breath = 1 + 1.5 * Math.sin(j * 0.02 + t * 0.001 + i)
            ctx.beginPath()
            ctx.strokeStyle = wave.color
            ctx.lineWidth = Math.max(0.5, breath) * dpr
            if (features.dash) {
              const dashLen = 6 * dpr
              const dashOff = (t * 0.02 + i * 50) % (dashLen * 2)
              ctx.setLineDash([dashLen, dashLen])
              ctx.lineDashOffset = -dashOff
            } else {
              ctx.setLineDash([])
            }
            const end = Math.min(j + segLen + 1, pts.length)
            ctx.moveTo(pts[j].x, pts[j].y)
            for (let k = j + 1; k < end; k++) ctx.lineTo(pts[k].x, pts[k].y)
            ctx.stroke()
          }
        } else {
          ctx.beginPath()
          ctx.strokeStyle = wave.color
          ctx.lineWidth = 1 * dpr
          if (features.dash) {
            const dashLen = 6 * dpr
            const dashOff = (t * 0.02 + i * 50) % (dashLen * 2)
            ctx.setLineDash([dashLen, dashLen])
            ctx.lineDashOffset = -dashOff
          } else {
            ctx.setLineDash([])
          }
          for (let j = 0; j < pts.length; j++) {
            if (j === 0) ctx.moveTo(pts[j].x, pts[j].y)
            else ctx.lineTo(pts[j].x, pts[j].y)
          }
          ctx.stroke()
        }
        ctx.setLineDash([])
      }

      if (features.pulse) {
        const pulseWidth = 0.08
        const segLen = 4
        for (let i = 0; i < waves.length; i++) {
          const pts = points[i]
          const pulsePos = ((t * 0.0001 * (1 + i * 0.2)) % 8.0) - 0.2
          for (let j = 0; j < pts.length - 1; j += segLen) {
            const progress = j / pts.length
            const dist = Math.abs(progress - pulsePos)
            const intensity = Math.max(0, 1 - dist / pulseWidth)
            const alpha = intensity * 0.2
            if (alpha < 0.005) continue
            const c = palette.dark[i % 3]
            const color = `rgba(${c},${alpha.toFixed(3)})`
            ctx.beginPath()
            ctx.strokeStyle = color
            ctx.lineWidth = (1 + intensity * 2) * dpr
            const end = Math.min(j + segLen + 1, pts.length)
            ctx.moveTo(pts[j].x, pts[j].y)
            for (let k = j + 1; k < end; k++) ctx.lineTo(pts[k].x, pts[k].y)
            ctx.stroke()
          }
        }
      }

      if (features.mist) {
        const colors = palette.dark
        const spotX = (0.5 + 0.45 * Math.sin(t * 0.00005)) * w
        const spotRadius = w * 0.2
        for (let i = 0; i < waves.length; i++) {
          const pts = points[i]
          const [r,g,bl] = colors[i % colors.length]
          const segLen = 6
          for (let j = 0; j < pts.length - 1; j += segLen) {
            const dist = Math.abs(pts[j].x - spotX)
            if (dist > spotRadius) continue
            const proximity = 1 - dist / spotRadius
            const alpha = proximity * proximity * 0.12
            ctx.beginPath()
            ctx.strokeStyle = `rgba(${r},${g},${bl},${alpha.toFixed(3)})`
            ctx.lineWidth = (1 + proximity) * dpr
            const end = Math.min(j + segLen + 1, pts.length)
            ctx.moveTo(pts[j].x, pts[j].y)
            for (let k = j + 1; k < end; k++) ctx.lineTo(pts[k].x, pts[k].y)
            ctx.stroke()
          }
        }
      }

      if (features.thread) {
        const colors = palette.dark
        const segLen = 20
        for (let i = 0; i < waves.length; i++) {
          const pts = points[i]
          const [r,g,bl] = colors[i % colors.length]
          for (let j = 0; j < pts.length - segLen; j += segLen) {
            const v = Math.sin(t * 0.0008 + j * 0.004 + i * 4) *
                      Math.sin(t * 0.0003 + j * 0.01 + i * 7)
            if (v < 0.4) continue
            const intensity = (v - 0.4) / 0.6
            const alpha = intensity * 0.18
            ctx.beginPath()
            ctx.strokeStyle = `rgba(${r},${g},${bl},${alpha.toFixed(3)})`
            ctx.lineWidth = (1 + intensity * 0.5) * dpr
            const end = Math.min(j + segLen + 1, pts.length)
            ctx.moveTo(pts[j].x, pts[j].y)
            for (let k = j + 1; k < end; k++) ctx.lineTo(pts[k].x, pts[k].y)
            ctx.stroke()
          }
        }
      }

      if (features.fog) {
        const colors = palette.dark
        const segLen = 8
        for (let i = 0; i < waves.length; i++) {
          const pts = points[i]
          const [r,g,bl] = colors[i % colors.length]
          for (let j = 0; j < pts.length - 1; j += segLen) {
            const nx = j * 0.002 + t * 0.00004
            const ny = i * 2 + t * 0.00003
            const v = Math.sin(nx * 3 + Math.sin(ny * 2)) * 0.5 + 0.5
            const alpha = v * 0.1
            if (alpha < 0.01) continue
            ctx.beginPath()
            ctx.strokeStyle = `rgba(${r},${g},${bl},${alpha.toFixed(3)})`
            ctx.lineWidth = 1 * dpr
            const end = Math.min(j + segLen + 1, pts.length)
            ctx.moveTo(pts[j].x, pts[j].y)
            for (let k = j + 1; k < end; k++) ctx.lineTo(pts[k].x, pts[k].y)
            ctx.stroke()
          }
        }
      }

      if (features.refract && points.length >= 2) {
        const colors = palette.dark
        const threshold = 16 * dpr
        const segLen = 6
        for (let i = 0; i < points.length - 1; i++) {
          const top = points[i], bot = points[i + 1]
          for (let w2 = 0; w2 < 2; w2++) {
            const pts = w2 === 0 ? top : bot
            const [r,g,bl] = colors[(i + w2) % colors.length]
            for (let j = 0; j < pts.length - 1; j += segLen) {
              const gap = Math.abs(top[j].y - bot[j].y)
              if (gap > threshold) continue
              const closeness = 1 - gap / threshold
              const alpha = closeness * closeness * 0.08
              if (alpha < 0.005) continue
              ctx.beginPath()
              ctx.strokeStyle = `rgba(${r},${g},${bl},${alpha.toFixed(3)})`
              ctx.lineWidth = (1 + closeness) * dpr
              const end = Math.min(j + segLen + 1, pts.length)
              ctx.moveTo(pts[j].x, pts[j].y)
              for (let k = j + 1; k < end; k++) ctx.lineTo(pts[k].x, pts[k].y)
              ctx.stroke()
            }
          }
        }
      }
    }
    requestAnimationFrame(draw)
  }
  requestAnimationFrame(draw)
})()
