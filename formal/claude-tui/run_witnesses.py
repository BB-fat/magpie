#!/usr/bin/env python3
"""Run intentionally failing safety assertions without leaving production edits."""
from pathlib import Path
import hashlib
import json
import os
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[1]
GATEWAY = REPO / 'internal/gateway'
target = GATEWAY / 'claude_tui_tla_generated_witness_test.go'
if target.exists():
    raise SystemExit('Witness target already exists; refusing to overwrite')
manifest = json.loads((ROOT / 'source-manifest.json').read_text())
for name, digest in manifest['files'].items():
    if hashlib.sha256((REPO / name).read_bytes()).hexdigest() != digest:
        raise SystemExit('Audited source differs from the manifest: ' + name)
changes = {}
# Observation points only; no condition, lock, state, or branch is changed.
patches = {
    'claude_tui_runtime_claude_interactive_points.go': (
        '\t\trun.abort()\n\t}\n\tb.mu.Lock()\n\tdefer b.mu.Unlock()',
        '\t\trun.abort()\n\t}\n\ttlaBeforeSessionReserve()\n\tb.mu.Lock()\n\tdefer b.mu.Unlock()'),
    'claude_tui_runtime_claude_lite.go': (
        '\t\tb.mu.Unlock()\n\t\tselect {\n\t\tcase <-ch:',
        '\t\tb.mu.Unlock()\n\t\ttlaRetryBeforeWait()\n\t\tselect {\n\t\tcase <-ch:'),
}
expected = ['CallerIsolation violated', 'OneExecutionPerRetry violated',
            'NoStartAfterFinish violated',
            'SingleNativeWriter violated']
env = os.environ.copy()
for key in ('GOPATH', 'GOMODCACHE', 'GOCACHE'):
    env[key] = subprocess.check_output(['go', 'env', key], text=True).strip()
env['GOTOOLCHAIN'] = 'go1.27.1'
result = None
try:
    target.write_bytes((ROOT / 'witnesses/claude_tui_tla_witness_test.go.txt').read_bytes())
    for name, (old, new) in patches.items():
        path = GATEWAY / name
        source = path.read_bytes()
        if source.decode().count(old) != 1:
            raise RuntimeError('Observer location ambiguous: ' + name)
        changes[path] = source
        path.write_text(source.decode().replace(old, new))
    with tempfile.TemporaryDirectory(prefix='magpie-tla-home-') as home:
        env['HOME'] = home
        result = subprocess.run(['go', 'test', '-race', '-tags', 'nogui', './internal/gateway',
                                 '-run', '^TestTLAWitness', '-count=20', '-timeout=3m', '-v'],
                                cwd=REPO, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT,
                                text=True, timeout=210)
        (ROOT / 'results/go-witnesses.log').write_text(result.stdout)
finally:
    for path, source in changes.items():
        path.write_bytes(source)
        assert hashlib.sha256(path.read_bytes()).digest() == hashlib.sha256(source).digest()
    target.unlink(missing_ok=True)
if result is None:
    raise SystemExit('Witness run did not start')
print(result.stdout[-10000:])
counts = {message: result.stdout.count(message) for message in expected}
missing = [message for message in expected if counts[message] != 20]
race = 'WARNING: DATA RACE' in result.stdout
(ROOT / 'results/go-witnesses-summary.json').write_text(json.dumps({
    'source_commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=REPO, text=True).strip(),
    'checker': 'go test -race -count=20, Go 1.27.1', 'expected_failures': expected,
    'exit': result.returncode, 'unexpected_counts': missing, 'failure_counts': counts, 'data_race_detected': race,
    'instrumentation': ['retry pre-wait observation hook', 'session pre-reservation observation hook'],
    'production_sources_restored': True}, indent=2) + '\n')
raise SystemExit(0 if result.returncode == 1 and not missing and not race else 1)
