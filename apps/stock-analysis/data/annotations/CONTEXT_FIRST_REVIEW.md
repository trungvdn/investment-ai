# Context-first review — 2026-11-02

This is the current result snapshot for PR #20. Earlier benchmark reports describe earlier implementations; this file and docs/wyckoff/analyzer-method.md describe the replacement default analyzer.

## Method and validation

Start with the latest 40 bars, expand for stopping context and retain parent campaigns before narrowing the event range. Select current campaign and forming phase before publishing its events. Confirm swing pivots with two following bars; expose HH/HL and LH/LL independently from Wyckoff phase. Keep original SC/AR boundaries for Spring/SOS and preserve failed campaigns.

Go Tests [41468900807](https://github.com/trungvdn/investment-ai/actions/runs/41468900807) and Dashboard tests/build [41468900695](https://github.com/trungvdn/investment-ai/actions/runs/41468900695) passed. The Go workflow formats source before testing; formatter commit 5b560a4 is the resulting PR head. Synthetic tests cover accumulation and mirrored distribution, UTAD, PS/PSY, event availability, 40-bar minimum, phase progression at successive cutoffs, and confirmed swing direction. Regression tests include SZC, MSB, HDG and nested NLG structures. No deployment or browser visual QA is claimed.

## Reviewed CSV snapshots through 2026-10-30

These are inferred outputs, not newly approved human labels.

| Stock | Current hypothesis | Swing pattern | Range | Bars used |
|---|---|---|---|---|
| SZC | Accumulation D, breakout testing | Mixed | 16.031–19.588 | 107 |
| MSB | Markup E after reaccumulation | HH/HL | 12.808–13.726 | 134 |
| LPB | Markdown E | LH/LL | 50.1–58.5 | 130 |
| NLG | Inner accumulation D | Mixed | 21.3–22.5 | 107 |
| HDG | Markdown after failed accumulation | Mixed | 15.2–17.2 | 107 |
| BCM | Accumulation B | HH/HL inside range | 33.9–47.0 | 107 |
| CTD | Accumulation D | HH/HL | 54.7–65.0 | 110 |
| GEX | Range B, accumulation/distribution bias unresolved | LH/LL inside range | 19.7–27.85 | 118 |
| NVL | Forming accumulation A | LH/LL from preceding decline | 9.6–11.3 | 63 |

SZC primary timeline: PS 23 July, SC 27 July, AR 12 August, ST 6 October, SOS 23 October. No late-October BC or LPS is asserted. MSB has Spring 26 August (available 27 August), SOS 23 September, LPS 5 October (available 7 October). HDG preserves its failed accumulation rather than inventing a successful new Spring.

## Same-fixture benchmark

Baseline: 263e389c1c9d83a4d2f6354aeaf2bee44d154bb7. Both versions replay the same nine CSV/JSON fixtures using the same evaluation harness and unchanged human annotations. Complete per-label tables and replay artifacts are attached to the Go run.

| Metric | Baseline | Current |
|---|---:|---:|
| Raw reviewed positive date matches | 22/36 | 25/36 |
| Raw rejected-label date matches | 3/20 | 1/20 |
| Structure-scoped positive matches | 8/36 | 19/36 |
| Structure-scoped rejected matches | 0/22 | 0/22 |
| Unscoped emitted predictions | 97 | 0 |
| Unmapped structure predictions | 72 | 34 |

Raw output scope changed: baseline events included multi-campaign hypotheses; current events contain the selected campaign, with other campaign hypotheses retained separately in structures. Thus this is a contract-and-method comparison, **not a measured gain in investment precision**. Scoped matching requires compatible geometry and timing; unmapped/ambiguous predictions do not count as correct. Three LPB rejected targets lack range geometry (TARGET_UNDEFINED); zero scoped rejected hits does not establish zero false positives. Uncertain labels remain excluded from positive/rejected scoring.

## Unresolved differences

- 17/36 reviewed positives still fail strict scoped matching. Thirty-four emitted predictions remain unmapped; do not hide these behind a zero rejected-hit count.
- NLG Test 20 October and LPS 27 October remain model hypotheses corresponding to uncertain human review. Subsequent re-entry into the inner range is counter-evidence; do not promote these to human-confirmed labels.
- BCM first ST is dated 6 October, available 7 October; human annotation dates it 7 October. OHLC-extreme range 33.9–47.0 differs from manually selected operational range 36.5–45.5.
- A rejected raw SZC SC is still matched on a historical replay (23 June); it is not in the current primary timeline. Historical replay mistakes remain visible.
- Event availability is rule availability, not proof of a tradable signal. Recomputing campaigns can revise a developing interpretation; immutable per-day vintage persistence is not implemented.
- Nine reviewed samples and synthetic tests do not establish out-of-sample accuracy. Heuristic confidence is not a calibrated probability. Corporate-action adjustment has not been independently verified.

See [method](../../../../docs/wyckoff/analyzer-method.md) for concrete rules and limitations.
