"""Render fixed-policy structure-scoped evaluation without promoting unmapped events."""
import json
import sys
from pathlib import Path
before, after = (json.loads(Path(p).read_text()) for p in sys.argv[1:3])
assert before['Policy'] == after['Policy'], 'Comparison requires identical matching policy'
lines = ['# Structure-scoped Wyckoff benchmark', '', after['Policy'], '',
         'Evaluation only: annotations never enter Analyze. Automatic IDs are mapped by frozen boundaries and formation time, not by event-label agreement.',
         'Unattributed, unmapped, low-confidence and ambiguous predictions are not counted as correct or rejected-target emissions. All reviewed positive labels remain in the denominator.',
         'Reviewed candidates are hypotheses; sparse labels cannot establish global precision or investment accuracy.', '',
         '| Metric | Baseline | Candidate |', '|---|---:|---:|']
for k in ['Positive', 'Hits', 'Rejected', 'RejectedEmitted', 'Uncertain', 'MappedPredictions', 'UnattributedPredictions', 'UnmappedPredictions', 'AmbiguousPredictions', 'LowConfidencePredictions']:
    lines.append(f'| {k} | {before[k]} | {after[k]} |')
lines += ['', 'Unlike the legacy score, all reviewed rejections are eligible by their own structure, including same-date contradictory labels on nested ranges. No missing-ID fallback is used.',
          'Prediction counts are unique ticker / predicted structure / event type / date over prefix replay within the review window. They are not false-positive counts.', '',
          '| Stock | Reviewed structure | Mapping coverage | Positive hits / total | Rejected emitted / total | Uncertain |', '|---|---|---|---:|---:|---:|']
for r in after['Structures']:
    lines.append(f"| {r['Ticker']} | {r['Structure']} | {r['Mapped'] if r['DefinitionAvailable'] else 'TARGET_UNDEFINED'} | {r['Hits']}/{r['Positive']} | {r['RejectedEmitted']}/{r['Rejected']} | {r['Uncertain']} |")
lines += ['', '| Stock | Event | Reviewed structure | Kind | Label | Scoped hit | Disposition | First seen |', '|---|---|---|---|---|---|---|---|']
for r in after['Events']:
    lines.append(f"| {r['Ticker']} | {r['EventID']} | {r['Structure']} | {r['Kind']} | {r['Label']} | {r['Hit']} | {r['Disposition']} | {r['FirstSeen']} |")
Path(sys.argv[3]).write_text('\n'.join(lines) + '\n')
