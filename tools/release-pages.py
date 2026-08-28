#!/usr/bin/env python3
"""generate changelog-<ver>.html pages from github release notes.

usage: release-pages.py v0.7.0 [v0.6.0 ...]
fetches each release body via `gh release view` and writes a site-styled page
beside the other top-level pages. re-run any time; pages are overwritten.
"""
import html
import re
import subprocess
import sys
from pathlib import Path

REPO = 'kungfusheep/glyph'
ROOT = Path(__file__).resolve().parent.parent

PAGE = '''<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>glyph - {ver} Changes</title>
<meta name="description" content="Changelog for glyph {ver}.">
<meta name="robots" content="index, follow">
<link rel="canonical" href="https://useglyph.sh/changelog-{slug}">
<meta name="theme-color" content="#131311">
<link rel="icon" type="image/png" sizes="32x32" href="favicon-32.png">
<link rel="icon" type="image/png" sizes="16x16" href="favicon-16.png">
<link rel="preload" href="BerkeleyMonoVariable.woff2" as="font" type="font/woff2" crossorigin>
<link rel="stylesheet" href="https://use.typekit.net/rgt4idh.css">
<link rel="stylesheet" href="tokens.css">
<style>
  @font-face {{
    font-family: 'Berkeley Mono';
    src: url('BerkeleyMonoVariable.woff2') format('woff2');
    font-weight: 100 900;
    font-display: swap;
  }}
  * {{ margin: 0; padding: 0; box-sizing: border-box; }}
  body {{
    background: var(--bg-dark);
    color: var(--dk);
    font-family: var(--mono);
    font-size: 13px;
    line-height: 1.7;
    -webkit-font-smoothing: antialiased;
  }}
  ::selection {{ background: var(--dk); color: var(--bg-dark); }}
  .page {{ max-width: 760px; margin: 0 auto; padding: 48px var(--s5) 96px; }}
  .verlinks {{ font-size: 11px; color: var(--dk-6); margin-bottom: var(--s4); }}
  .verlinks a {{ color: var(--dk-4); text-decoration: none; margin-right: 10px; }}
  .verlinks a:hover {{ color: var(--dk-hi); }}
  h1 {{
    font-family: var(--akz);
    font-weight: 700;
    font-size: 34px;
    color: var(--dk-hi);
    letter-spacing: -0.02em;
    line-height: 1.1;
    margin-bottom: var(--s4);
  }}
  h2 {{
    font-family: var(--akz);
    font-weight: 700;
    font-size: 21px;
    color: var(--dk-hi);
    margin: var(--s6) 0 var(--s3);
    padding-top: var(--s4);
    border-top: 1px solid var(--rule-dk);
  }}
  h3 {{
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.1em;
    color: var(--dk-6);
    margin: var(--s4) 0 var(--s2);
  }}
  p {{ margin-bottom: var(--s3); }}
  ul {{ list-style: none; margin-bottom: var(--s3); }}
  li {{ padding-left: 18px; position: relative; margin-bottom: var(--s2); }}
  li::before {{ content: '-'; position: absolute; left: 0; color: var(--dk-6); }}
  code {{
    color: var(--dk-hi);
    background: rgba(255,255,255,0.05);
    padding: 1px 5px;
    border-radius: 3px;
    font-family: var(--mono);
    font-size: 12px;
  }}
  pre {{
    background: rgba(255,255,255,0.04);
    border: 1px solid var(--rule-dk);
    border-radius: 8px;
    padding: var(--s3);
    overflow-x: auto;
    margin-bottom: var(--s3);
  }}
  pre code {{ background: none; padding: 0; color: var(--dk); }}
  a {{ color: var(--dk-hi); }}
</style>
</head>
<body>
<main class="page">
  <nav class="verlinks">{verlinks}</nav>
{body}
</main>
</body>
</html>
'''


def inline(s):
    s = html.escape(s, quote=False)
    s = re.sub(r'`([^`]+)`', r'<code>\1</code>', s)
    s = re.sub(r'\*\*([^*]+)\*\*', r'<strong>\1</strong>', s)
    s = re.sub(r'\[([^\]]+)\]\(([^)]+)\)', r'<a href="\2">\1</a>', s)
    return s


def md_to_html(md):
    out, in_code, in_list, code = [], False, False, []
    for line in md.splitlines():
        if line.strip().startswith('```'):
            if in_code:
                out.append('<pre><code>%s</code></pre>' % html.escape('\n'.join(code)))
                code, in_code = [], False
            else:
                if in_list:
                    out.append('</ul>')
                    in_list = False
                in_code = True
            continue
        if in_code:
            code.append(line)
            continue
        m = re.match(r'^(#{1,4})\s+(.*)', line)
        if m:
            if in_list:
                out.append('</ul>')
                in_list = False
            level = min(len(m.group(1)), 3)
            out.append('<h%d>%s</h%d>' % (level, inline(m.group(2)), level))
            continue
        m = re.match(r'^\s*[-*]\s+(.*)', line)
        if m:
            if not in_list:
                out.append('<ul>')
                in_list = True
            out.append('<li>%s</li>' % inline(m.group(1)))
            continue
        if not line.strip():
            if in_list:
                out.append('</ul>')
                in_list = False
            continue
        out.append('<p>%s</p>' % inline(line))
    if in_list:
        out.append('</ul>')
    if in_code:
        out.append('<pre><code>%s</code></pre>' % html.escape('\n'.join(code)))
    return '\n'.join('  ' + l for l in out)


def normalise(md, ver):
    """uniform hierarchy across releases: one h1 (added by us), the compare
    link at the foot, and an h2 above any body that has none of its own."""
    lines = md.splitlines()
    compare = None
    kept = []
    for line in lines:
        if re.match(r'^#\s+', line):
            continue  # the page supplies the h1
        m = re.match(r'^\s*(?:\*\*)?full changelog(?:\*\*)?:\s*(\S+)', line, re.I)
        if m:
            compare = m.group(1)
            continue
        kept.append(line)
    body = '\n'.join(kept).strip('\n')
    if not re.search(r'^##\s+', body, re.M):
        body = '## Changes\n\n' + body
    body = '# glyph %s\n\n%s' % (ver.lstrip('v'), body)
    if compare:
        body += '\n\n[Full changelog on GitHub](%s)' % compare
    return body


def main(versions):
    slugs = [v.lstrip('v').rsplit('.0', 1)[0] if v.endswith('.0') else v.lstrip('v') for v in versions]
    for ver, slug in zip(versions, slugs):
        body = subprocess.run(
            ['gh', 'release', 'view', ver, '--repo', REPO, '--json', 'body', '-q', '.body'],
            check=True, capture_output=True, text=True).stdout
        links = ' '.join(
            '<a href="changelog-%s.html">%s</a>' % (s, v) for v, s in zip(versions, slugs) if s != slug)
        page = PAGE.format(ver=ver, slug=slug, verlinks=links, body=md_to_html(normalise(body, ver)))
        out = ROOT / ('changelog-%s.html' % slug)
        out.write_text(page)
        print('wrote', out.name)


if __name__ == '__main__':
    if len(sys.argv) < 2:
        sys.exit(__doc__)
    main(sys.argv[1:])
