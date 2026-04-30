#!/usr/bin/env python3
import re, sys
from pathlib import Path

if len(sys.argv) != 2 or not sys.argv[1].startswith('v'):
    sys.exit('usage: bump-version.py vX.Y.Z')

v = sys.argv[1]
root = Path(__file__).resolve().parent.parent
files = ['index.html', 'concepts.html', 'api.html', 'start.html', 'styleguide.html']

# matches `glyph@latest` or `glyph@vX.Y.Z` in install commands
pattern = re.compile(r'(github\.com/kungfusheep/glyph@)(?:latest|v[\d][\w.\-]*)')

for name in files:
    f = root / name
    if not f.exists():
        continue
    before = f.read_text()
    after = pattern.sub(rf'\g<1>{v}', before)
    if after != before:
        f.write_text(after)
        print(f'updated {name}')
