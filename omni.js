// omni.js — site-wide search popover
// self-contained: injects CSS, lazily loads api.json, builds UI
(function() {
  // skip on the api page — it has its own inline search
  if (location.pathname.endsWith('api.html') || location.pathname.endsWith('api')) return;

  // fzf scorer functions loaded from fzf.js (global scope)

  // ============================================================
  // DISPLAY NAME MAP (matches api.html)
  // ============================================================

  const displayNames = {
    VBoxFn: 'VBox', HBoxFn: 'HBox', OverlayFn: 'Overlay',
    TextC: 'Text', InputC: 'Input', ListC: 'List',
    CheckListC: 'CheckList', TabsC: 'Tabs',
    FilterListC: 'FilterList', FilterLogC: 'FilterLog', LogC: 'Log',
    LayerViewC: 'LayerView', JumpC: 'Jump',
    RadioC: 'Radio', CheckboxC: 'Checkbox',
    ScrollbarC: 'Scroll',
    HRuleC: 'HRule', VRuleC: 'VRule', LeaderC: 'Leader',
    SpacerC: 'Spacer', AutoTableC: 'AutoTable',
    SparklineC: 'Sparkline', ProgressC: 'Progress', SpinnerC: 'Spinner',
    ForEachC: 'ForEach', Condition: 'If', OrdCondition: 'IfOrd',
    SwitchBuilder: 'Switch', RichTextNode: 'Rich',
    Custom: 'Widget', ThemeEx: 'Theme',
  };

  const hidden = new Set([
    'VBoxC', 'HBoxC', 'OverlayC',
    'VBoxNode', 'HBoxNode', 'OverlayNode', 'SpacerNode',
    'TextNode', 'TabsNode', 'HRuleNode', 'VRuleNode', 'LeaderNode',
    'JumpNode', 'ScrollbarNode', 'SparklineNode', 'ProgressNode', 'SpinnerNode',
    'IfNode', 'ElseNode', 'SwitchNode', 'ForEachNode', 'LayerViewNode',
    'ChildSize', 'Size', 'Rect', 'Component', 'TextInput', 'Renderer',
    'LayoutFunc', 'Layout', 'FocusGroup',
  ]);

  function getDisplayName(type) {
    if (displayNames[type.name]) return displayNames[type.name];
    if (type.constructors?.length === 1) return type.constructors[0].name;
    if (type.vars?.length) return type.vars[0].names[0];
    return type.name;
  }

  // ============================================================
  // DATA
  // ============================================================

  let entries = null; // cached search entries
  let loading = false;

  function loadData() {
    if (entries || loading) return Promise.resolve();
    loading = true;
    return fetch('api.json')
      .then(r => { if (!r.ok) throw new Error(r.status); return r.json(); })
      .then(data => {
        entries = [];
        for (const t of (data.types || [])) {
          if (hidden.has(t.name)) continue;
          const dname = getDisplayName(t);
          const methods = (t.methods || []).map(m => m.name);
          const kind = extractKind(t.decl);
          entries.push({ displayName: dname, typeName: t.name, methods, kind });
        }
      })
      .catch(() => { loading = false; });
  }

  function extractKind(decl) {
    if (!decl) return '';
    const m = decl.match(/^type \w+(?:\[.*?\])? ([\s\S]+)/);
    if (!m) return '';
    const raw = m[1].replace(/\{[\s\S]*\}/, '').replace(/struct\s*$/, '').trim();
    if (!raw || raw === 'struct' || raw.length > 30) return '';
    return raw;
  }

  function scoreEntry(entry, tokens) {
    let total = 0;
    for (const tok of tokens) {
      let best = matchToken(tok, entry.displayName) * 3;
      best = Math.max(best, matchToken(tok, entry.typeName) * 3);
      for (const m of entry.methods) {
        best = Math.max(best, matchToken(tok, m) * 2);
      }
      if (tok.inverse) {
        if (best > 0) return 0;
      } else {
        if (best === 0) return 0;
        total += best;
      }
    }
    return total || 1;
  }

  // which method matched best (for display)
  function bestMethodMatch(entry, tokens) {
    let best = null, bestScore = 0;
    for (const tok of tokens) {
      if (tok.inverse) continue;
      for (const m of entry.methods) {
        const s = matchToken(tok, m);
        if (s > bestScore) { bestScore = s; best = m; }
      }
    }
    // only show if the method scored higher than the name itself
    if (!best) return null;
    for (const tok of tokens) {
      if (tok.inverse) continue;
      const nameScore = Math.max(matchToken(tok, entry.displayName), matchToken(tok, entry.typeName));
      if (nameScore * 1.5 >= bestScore) return null;
    }
    return best;
  }

  // ============================================================
  // CSS
  // ============================================================

  const style = document.createElement('style');
  style.textContent = `
    @property --omni-angle { syntax: '<angle>'; initial-value: 135deg; inherits: false; }
    @keyframes omni-spin { from { --omni-angle: 135deg; } to { --omni-angle: 495deg; } }

    .omni-backdrop {
      position: fixed; inset: 0; z-index: 9998;
      background: rgba(0,0,0,0.6);
      backdrop-filter: blur(2px);
      -webkit-backdrop-filter: blur(2px);
      opacity: 0; transition: opacity 0.2s;
      pointer-events: none;
    }
    .omni-backdrop.open { opacity: 1; pointer-events: auto; }

    .omni-panel {
      position: fixed;
      top: 22vh; left: 50%;
      transform: translateX(-50%) scale(0.96);
      z-index: 9999;
      width: min(520px, calc(100vw - 32px));
      opacity: 0;
      transition: opacity 0.2s, transform 0.2s;
      pointer-events: none;
    }
    .omni-panel.open {
      opacity: 1; transform: translateX(-50%) scale(1);
      pointer-events: auto;
    }

    .omni-inner {
      position: relative;
      isolation: isolate;
      border-radius: 10px;
      box-shadow:
        0 0 30px 16px var(--bg-dark, #131311),
        0 4px 45px 12px rgba(0,0,0,0.4),
        -5px -2px 25px 8px var(--bg-dark, #131311);
    }
    .omni-inner::before {
      content: '';
      position: absolute;
      inset: -1px;
      border-radius: 11px;
      padding: 1px;
      background: conic-gradient(from var(--omni-angle), #da5050, #f86868, #a04050, #6c3058, #f86868, #da5050);
      -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
      -webkit-mask-composite: xor;
      mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
      mask-composite: exclude;
      animation: omni-spin 3s linear infinite;
      opacity: 0.8;
      pointer-events: none;
    }
    .omni-inner::after {
      content: '';
      position: absolute;
      inset: -8px;
      border-radius: 16px;
      background: conic-gradient(from var(--omni-angle), #c44040, #e85050, #983848, #682850, #e85050, #c44040);
      filter: blur(20px);
      animation: omni-spin 3s linear infinite;
      opacity: 0.2;
      pointer-events: none;
      z-index: -1;
    }

    .omni-input {
      display: block;
      width: 100%;
      box-sizing: border-box;
      background: var(--bg-dark3, #0e0e0c);
      border: 1px solid transparent;
      border-radius: 10px 10px 0 0;
      color: #c8c4b8;
      font-family: var(--mono, 'Berkeley Mono', monospace);
      font-size: 16px;
      padding: 14px 16px;
      outline: none;
    }
    .omni-input::placeholder { color: #444; }
    .omni-results:empty ~ .omni-input,
    .omni-input:only-child {
      border-radius: 10px;
    }

    .omni-results {
      background: var(--bg-dark3, #0e0e0c);
      border-top: 1px solid var(--rule-dk, #262624);
      border-radius: 0;
      max-height: 360px;
      overflow-y: auto;
      scrollbar-width: thin;
      scrollbar-color: var(--rule-dk, #262624) transparent;
    }
    .omni-results:empty { display: none; }

    .omni-row {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 10px 16px;
      cursor: pointer;
      color: #8a8880;
      transition: background 0.1s;
    }
    .omni-row:hover { background: rgba(255,255,255,0.03); }
    .omni-row.selected { background: rgba(218,80,80,0.12); }

    .omni-row-name {
      font-family: var(--mono, 'Berkeley Mono', monospace);
      font-weight: 600;
      color: #c8c4b8;
      font-size: 14px;
    }
    .omni-row-method {
      font-family: var(--mono, 'Berkeley Mono', monospace);
      color: #666;
      font-size: 13px;
    }
    .omni-row-kind {
      font-size: 11px;
      color: #555;
      margin-left: auto;
      white-space: nowrap;
    }

    .omni-hint {
      display: flex;
      align-items: center;
      gap: 12px;
      padding: 8px 16px;
      font-size: 11px;
      color: #333;
      font-family: var(--mono, 'Berkeley Mono', monospace);
      background: var(--bg-dark3, #0e0e0c);
      border-top: 1px solid var(--rule-dk, #262624);
      border-radius: 0 0 10px 10px;
    }
    .omni-results:not(:empty) + .omni-hint {
      border-radius: 0 0 10px 10px;
    }
    .omni-results:empty + .omni-hint {
      border-top: 1px solid var(--rule-dk, #262624);
    }
    .omni-hint kbd {
      background: #1a1a18;
      border: 1px solid #333;
      border-radius: 3px;
      padding: 1px 5px;
      font-size: 10px;
    }

  `;
  document.head.appendChild(style);

  // ============================================================
  // DOM
  // ============================================================

  const backdrop = document.createElement('div');
  backdrop.className = 'omni-backdrop';
  document.body.appendChild(backdrop);

  const panel = document.createElement('div');
  panel.className = 'omni-panel';
  panel.innerHTML = `
    <div class="omni-inner">
      <input class="omni-input" type="text" placeholder="Search API..." autocomplete="off" spellcheck="false">
      <div class="omni-results"></div>
      <div class="omni-hint">
        <span><kbd>↑↓</kbd> navigate</span>
        <span><kbd>↵</kbd> open</span>
        <span><kbd>esc</kbd> close</span>
      </div>
    </div>
  `;
  document.body.appendChild(panel);

  const input = panel.querySelector('.omni-input');
  const results = panel.querySelector('.omni-results');
  let selectedIdx = 0;
  let visibleEntries = [];

  // ============================================================
  // OPEN / CLOSE
  // ============================================================

  let isOpen = false;

  function open() {
    if (isOpen) return;
    isOpen = true;
    loadData().then(() => {
      backdrop.classList.add('open');
      panel.classList.add('open');
      input.value = '';
      results.innerHTML = '';
      selectedIdx = 0;
      visibleEntries = [];
      input.focus({ preventScroll: true });
    });
  }

  function close() {
    if (!isOpen) return;
    isOpen = false;
    backdrop.classList.remove('open');
    panel.classList.remove('open');
    input.blur();
  }

  window.openOmni = open;

  backdrop.addEventListener('click', close);

  // ============================================================
  // SEARCH
  // ============================================================

  function search() {
    const raw = input.value.trim();
    if (!raw || !entries) {
      results.innerHTML = '';
      visibleEntries = [];
      selectedIdx = 0;
      return;
    }

    const tokens = parseFzfQuery(raw);
    const scored = entries
      .map(e => ({ entry: e, score: scoreEntry(e, tokens) }))
      .filter(s => s.score > 0)
      .sort((a, b) => b.score - a.score)
      .slice(0, 15);

    visibleEntries = scored;
    selectedIdx = 0;

    results.innerHTML = scored.map(({ entry }, i) => {
      const method = bestMethodMatch(entry, tokens);
      return `<div class="omni-row${i === 0 ? ' selected' : ''}" data-idx="${i}">
        <span class="omni-row-name">${esc(entry.displayName)}</span>
        ${method ? `<span class="omni-row-method">.${esc(method)}</span>` : ''}
        ${entry.kind ? `<span class="omni-row-kind">${esc(entry.kind)}</span>` : ''}
      </div>`;
    }).join('');
  }

  function esc(s) {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  function updateSelection() {
    results.querySelectorAll('.omni-row').forEach((row, i) => {
      row.classList.toggle('selected', i === selectedIdx);
      if (i === selectedIdx) row.scrollIntoView({ block: 'nearest' });
    });
  }

  function navigate() {
    if (!visibleEntries.length) return;
    const entry = visibleEntries[selectedIdx]?.entry;
    if (!entry) return;
    const tokens = parseFzfQuery(input.value.trim());
    const method = bestMethodMatch(entry, tokens);
    const hash = method ? entry.displayName + '.' + method : entry.displayName;
    close();
    window.location = 'api.html#' + hash;
  }

  let searchTimeout;
  input.addEventListener('input', () => {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(search, 60);
  });

  results.addEventListener('click', e => {
    const row = e.target.closest('.omni-row');
    if (!row) return;
    selectedIdx = parseInt(row.dataset.idx, 10);
    navigate();
  });

  // ============================================================
  // KEYBOARD
  // ============================================================

  document.addEventListener('keydown', e => {
    if (isOpen) {
      if (e.key === 'Escape') { e.preventDefault(); close(); }
      else if (e.key === 'ArrowDown') {
        e.preventDefault();
        if (selectedIdx < visibleEntries.length - 1) { selectedIdx++; updateSelection(); }
      }
      else if (e.key === 'ArrowUp') {
        e.preventDefault();
        if (selectedIdx > 0) { selectedIdx--; updateSelection(); }
      }
      else if (e.key === 'Enter') { e.preventDefault(); navigate(); }
      return;
    }

    if (e.key === '/' && !e.ctrlKey && !e.metaKey) {
      const tag = document.activeElement?.tagName;
      if (tag === 'INPUT' || tag === 'TEXTAREA' || document.activeElement?.isContentEditable) return;
      e.preventDefault();
      open();
    }
  });
})();
