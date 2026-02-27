// fzf.js — shared fuzzy search scorer
// used by api.html inline search and omni.js popover

function fzfFuzzy(needle, haystack) {
  const nl = needle.length, hl = haystack.length;
  if (nl === 0) return 1;
  if (nl > hl) return 0;
  const nLow = needle.toLowerCase(), hLow = haystack.toLowerCase();
  for (let i = 0, from = 0; i < nl; i++) {
    const idx = hLow.indexOf(nLow[i], from);
    if (idx < 0) return 0;
    from = idx + 1;
  }
  let score = 0, consecutive = 0, hi = 0;
  for (let ni = 0; ni < nl; ni++) {
    const nc = nLow[ni];
    let found = false;
    while (hi < hl) {
      if (hLow[hi] === nc) {
        score += 1;
        if (consecutive > 0) score += consecutive * 2;
        consecutive++;
        if (needle[ni] === haystack[hi]) score += 1;
        if (hi === 0) { score += 8; }
        else {
          const prev = haystack[hi - 1], cur = haystack[hi];
          if (prev === '.' || prev === '_' || prev === '-' || prev === ' ' || prev === '(') score += 6;
          else if (prev >= 'a' && prev <= 'z' && cur >= 'A' && cur <= 'Z') score += 5;
        }
        hi++; found = true; break;
      }
      consecutive = 0; hi++;
    }
    if (!found) return 0;
  }
  score += Math.round((nl / hl) * 10);
  if (hLow.startsWith(nLow)) score += 15;
  return score;
}

function isWordBoundary(s, i) {
  if (i <= 0) return true;
  const prev = s[i - 1], cur = s[i];
  if (prev === '.' || prev === '_' || prev === '-' || prev === ' ' || prev === '(') return true;
  if (prev >= 'a' && prev <= 'z' && cur >= 'A' && cur <= 'Z') return true;
  return false;
}

function fzfExact(needle, haystack) {
  const nLow = needle.toLowerCase(), hLow = haystack.toLowerCase();
  let from = 0;
  while (from <= hLow.length - nLow.length) {
    const idx = hLow.indexOf(nLow, from);
    if (idx < 0) return 0;
    if (isWordBoundary(haystack, idx)) {
      let score = 10 + needle.length;
      const end = idx + needle.length;
      if (end >= haystack.length || isWordBoundary(haystack, end)) score += 20;
      if (nLow === hLow) score += 15;
      if (idx === 0) score += 8;
      return score;
    }
    from = idx + 1;
  }
  return 0;
}

function fzfPrefix(needle, haystack) {
  return haystack.toLowerCase().startsWith(needle.toLowerCase()) ? 15 + needle.length : 0;
}

function fzfSuffix(needle, haystack) {
  return haystack.toLowerCase().endsWith(needle.toLowerCase()) ? 15 + needle.length : 0;
}

function parseFzfQuery(raw) {
  const tokens = [];
  for (const part of raw.split(/\s+/)) {
    if (!part) continue;
    let term = part, inverse = false;
    if (term.startsWith('!')) { inverse = true; term = term.slice(1); }
    let type = 'fuzzy';
    if (term.startsWith("'")) { type = 'exact'; term = term.slice(1); }
    else if (term.startsWith('^')) { type = 'prefix'; term = term.slice(1); }
    else if (term.endsWith('$')) { type = 'suffix'; term = term.slice(0, -1); }
    if (term) tokens.push({ term, type, inverse });
  }
  return tokens;
}

function matchToken(token, haystack) {
  switch (token.type) {
    case 'exact':  return fzfExact(token.term, haystack);
    case 'prefix': return fzfPrefix(token.term, haystack);
    case 'suffix': return fzfSuffix(token.term, haystack);
    default:       return fzfFuzzy(token.term, haystack);
  }
}
