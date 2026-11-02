#!/usr/bin/env python3
"""Check the repaired Go guarantees without instrumentation or source edits."""
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[1]
CASES = ('CompletedReplyIsCallerScoped', 'SessionReservation',
         'ClosedStartupCannotLaunch', 'NativeStartFenceWaitsForPublication',
         'QueuedRetryRechecksCompletedReply')
env = os.environ.copy()
for key in ('GOPATH', 'GOMODCACHE', 'GOCACHE'):
    env[key] = subprocess.check_output(['go', 'env', key], text=True).strip()
env['GOTOOLCHAIN'] = 'go1.27.1'
with tempfile.TemporaryDirectory(prefix='magpie-tui-fixed-home-') as home:
    env['HOME'] = home
    run = subprocess.run(['go', 'test', '-race', '-tags', 'nogui', './internal/gateway',
                          '-run', '^TestClaudeTUIRuntime(' + '|'.join(CASES) + ')',
                          '-count=20', '-timeout=5m'], cwd=REPO, env=env,
                         stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True,
                         timeout=330)
(ROOT / 'results/fixed-go-regressions.log').write_text(run.stdout)
(ROOT / 'results/fixed-go-regressions-summary.json').write_text(json.dumps({
    'source_commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=REPO, text=True).strip(),
    'checker': 'Go 1.27.1, -race -count=20', 'production_instrumentation': False,
    'test_families': CASES, 'exit': run.returncode}, indent=2) + '\n')
print(run.stdout)
raise SystemExit(run.returncode)
