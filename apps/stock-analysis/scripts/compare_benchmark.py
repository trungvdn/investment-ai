"""Compare identical reviewed annotations on baseline and candidate analyzers."""
import json
import sys
from pathlib import Path
before, after = (json.loads(Path(p).read_text()) for p in sys.argv[1:3])
lines = ['# Reviewed Wyckoff benchmark', '',
         'Sparse reviewed hypotheses; not investment accuracy or calibrated probabilities.',
         'Replay feeds only OHLCV through each cutoff; labels and future outcomes never enter Analyze.',
         'Hit = event date lies inside the reviewed interval and is emitted at some cutoff from interval end through analysis_as_of.',
         'This measures retrospective discoverability, not earliest real-time confirmation. Recent-event pruning makes final-cutoff recall a separate observation.', '',
         '| Metric | Before | After |', '|---|---:|---:|']
for key in ['Positive','Hits','Rejected','RejectedEmitted','Uncertain','Unsupported','StructureAmbiguous']:
    lines.append(f'| {key} | {before[key]} | {after[key]} |')
lines += ['', 'RejectedEmitted counts rejected proposed labels only on reviewed intervals; it is not global precision.',
          'Uncertain, unsupported taxonomy and structure-ambiguous negatives are excluded from positive/rejected metrics.',
          'Correct candidates are hypothesis labels, not confirmed events. Corrected replacement labels are not added as extra positives without their own correct record.', '',
          '| Stock | Event / structure | Kind | Label | Before hit | After hit | First seen after | At final cutoff after |',
          '|---|---|---|---|---|---|---|---|']
old={(r['Ticker'],r['ID']):r for r in before['Events']}
for r in after['Events']:
    p=old[(r['Ticker'],r['ID'])]
    lines.append(f"| {r['Ticker']} | {r['ID']} / {r['Structure']} | {r['Kind']} | {r['Label']} | {p['Hit']} | {r['Hit']} | {r['FirstSeen']} | {r['AtCutoff']} |")
Path(sys.argv[3]).write_text('\n'.join(lines)+'\n')
