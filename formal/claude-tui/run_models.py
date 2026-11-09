#!/usr/bin/env python3
"""Run bounded TLC cases. Hypotheses/mutants are not implementation proofs."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
import time

from public_evidence import redact_tlc_output

ROOT = Path(__file__).resolve().parent
SHA256 = '936a262061c914694dfd669a543be24573c45d5aa0ff20a8b96b23d01e050e88'
parser = argparse.ArgumentParser()
parser.add_argument('--jar', default=os.environ.get('TLA2TOOLS_JAR', '/tmp/magpie-tla-tools/tla2tools-1.7.4.jar'))
parser.add_argument('--case', action='append')
parser.add_argument('--timeout', type=int, default=180)
args = parser.parse_args()
jar = Path(args.jar).resolve()
if hashlib.sha256(jar.read_bytes()).hexdigest() != SHA256:
    raise SystemExit('Unexpected tla2tools 1.7.4 JAR checksum; refusing an unpinned checker')
cases = json.loads((ROOT / 'cases.json').read_text())
if args.case:
    unknown = set(args.case) - {case['name'] for case in cases}
    if unknown:
        raise SystemExit('Unknown cases: ' + ', '.join(sorted(unknown)))
results = []
failed = False
for case in cases:
    if args.case and case['name'] not in args.case:
        continue
    name = case['name']
    log = ROOT / 'results' / (name + '.log')
    started = time.monotonic()
    with tempfile.TemporaryDirectory(prefix='magpie-tlc-') as state:
        cmd = ['java', '-XX:+UseParallelGC', '-Xmx1g', '-cp', str(jar), 'tlc2.TLC',
               '-workers', '1', '-fp', '0', '-seed', '1',
               '-metadir', state, '-config', str(ROOT / 'models' / (name + '.cfg')),
               str(ROOT / 'models' / (case['module'] + '.tla'))]
        try:
            run = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                                 timeout=args.timeout, text=True)
            output, code = run.stdout, run.returncode
        except subprocess.TimeoutExpired as err:
            output, code = (err.stdout or b'').decode(errors='replace'), 124
    captured_digest = hashlib.sha256(output.encode()).hexdigest()
    public_output = redact_tlc_output(output)
    log.write_text(public_output)
    error = re.search(r'Error: Invariant (\w+) is violated', output)
    violated = error.group(1) if error else None
    expected = case['expected']
    ok = (code == 12 and violated == expected) if expected else (code == 0 and 'No error has been found' in output)
    states = re.search(r'([\d,]+) states generated, ([\d,]+) distinct states found', output)
    result = {**case, 'exit': code, 'violated': violated, 'as_expected': ok,
              'seconds': round(time.monotonic() - started, 3),
              'generated': int(states[1].replace(',', '')) if states else None,
              'distinct': int(states[2].replace(',', '')) if states else None,
              'log_sha256': hashlib.sha256(log.read_bytes()).hexdigest(),
              'captured_log_sha256': captured_digest,
              'log_redaction': 'Local machine details and absolute paths omitted; checker results unchanged.'}
    results.append(result)
    failed |= not ok
    print(name, 'OK' if ok else 'UNEXPECTED', 'exit', code, 'invariant', violated,
          'distinct', result['distinct'], flush=True)
    if not ok:
        print(public_output[-3500:], flush=True)
(ROOT / 'results' / ('summary-selected.json' if args.case else 'summary.json')).write_text(json.dumps({
    'source_commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip(),
    'checker': 'TLC 1.7.4', 'jar_sha256': SHA256, 'cases': results}, indent=2) + '\n')
raise SystemExit(1 if failed else 0)
