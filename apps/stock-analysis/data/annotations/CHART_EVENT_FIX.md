# Chart event correctness and duplicate-marker fix

Cause: all multi-range lifecycle events and legacy candidates were merged into one report.events list. The chart plotted every item on one displayed support/resistance pair, stacking identical SOS labels and presenting smaller-range Springs as if they belonged to the larger range.

The standardized report now uses a conservative primary-range event/evidence timeline and preserves raw hypotheses under eventCandidates. Foreign structure IDs are excluded. Boundary/sequence checks reject non-penetrating Springs, stale TESTs and distant pullbacks. Each range has one consolidated SOS and one qualifying LPS with subsequent rebound. Later independent Spring attempts can re-arm after support is re-established. SC selection distinguishes strongest stopping selloffs from high-volume absorption climaxes.

The frontend deduplicates event type/date, rejects another range's attributed markers, validates Spring penetration, groups different same-day labels into one marker and uses that same list for clicks.

Regression fixture results (through 2026-10-30):

- MSB: SPRING 2026-08-26; SOS 2026-09-23; LPS 2026-10-05. No July TEST or repeated markup SOS/LPS.
- BCM: PS 2026-07-23; SC 2026-07-27; AR 2026-08-13; ST 2026-10-07 and 2026-10-14. No smaller-range SPRING on the 36.5–45.5 primary range.
- NVL: preserved high-volume absorption SC 2026-09-30; no regression to an earlier limit-down day.

Go Tests: https://github.com/trungvdn/investment-ai/actions/runs/41459339785 . Dashboard tests and production build: https://github.com/trungvdn/investment-ai/actions/runs/41459339627 . Tests cover candidate leakage into reports/evidence, explicit empty primary timelines, event-marker deduplication/grouping and the reviewed real data. No live deployment or browser visual verification was performed; live input may differ from the reviewed fixture cutoff. All labels remain heuristic hypotheses.
