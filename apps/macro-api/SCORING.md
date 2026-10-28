# Macro support score v1

Experimental, deterministic 0–100 support index for the equity-investment context.
It is not a probability, predicted return, or buy/sell recommendation. Thresholds
and weights are hypotheses, not calibrated or backtested coefficients. The UI
shows coverage independently of score; coverage is not statistical confidence.

## Model and interpretation

`internal/scoring/model.json` is embedded into the API binary. Its exact SHA-256
and semantic version are persisted in every snapshot. Change the version when
changing a rule or interpretation, and redeploy. Historical versions remain in
storage; the API serves only the active configuration hash, so chart lines do
not silently splice different models. Historical replay/backtesting is not
implemented in this version.

| Indicator | Weight | Group | Economic interpretation |
| --- | ---: | --- | --- |
| VN_DEPOSIT_RATE | 20 | Liquidity | Domestic funding conditions; transmission has lags |
| VN_PMI | 20 | Production | Early survey of manufacturing activity |
| VN_CPI_YOY | 15 | Inflation | Realized consumer inflation; often lagging |
| VN_CREDIT_GROWTH | 15 | Liquidity | Credit transmission, published with delays; YTD |
| VN_IIP_YOY | 10 | Production | Realized industrial activity, roughly coincident |
| US_EFFR | 10 | Liquidity | International monetary conditions; effective rate, not target range |
| VN_PPI_YOY | 5 | Inflation | Producer price pressure; can precede consumer prices |
| VN_M2_GROWTH | 5 | Liquidity | Monetary/liquidity conditions; compatible-method YoY only |

Leading/coincident/lagging describe relationships to economic activity and are
context dependent. They do not guarantee a lead over equity prices. They do not
create extra weight buckets. Production has 30%, inflation 20%, liquidity 50%.

## Exact rules

Let `C(x)=max(0,min(100,x))`; changes are percentage **points**, except PMI
index-point changes. `d(x,[a,b])=max(a-x,0,x-b)` is distance outside a band.
Unless stated otherwise, compare latest period with exactly three calendar
months earlier. Endpoints must be present and use the same source-series ID;
missing intermediate months do not synthesize observations. IIP and credit
require complete consecutive monthly windows. No default 0/50 substitutes for
missing history.

- Deposit: average `C(50−25×change_lower)` and `C(50−25×change_upper)`.
  This averages scores, **never** the underlying interest-rate bounds.
- PMI: `0.7×C(50+10×(PMI−50)) + 0.3×C(50+10×change)`.
- CPI: band [2,4.5] YoY, slope25. This is an experimental reference band,
  **not** an official inflation target.
- PPI: band [−2,2] YoY, slope15; compare previous quarter (three months).
- M2: band [6,18] YoY, slope6; compare three months. Both observations must
  have the same methodology after stripping article/parser provenance tags.
  No crossing the provider's October2025 methodology break.
- For those three band rules, with slope `k`, score is
  `0.7×C(100−k×d(current,band)) + 0.3×C(50+k×(d(previous,band)−d(current,band)))`.
  Deep negative inflation is penalized; lower CPI/PPI is not always better.
- IIP: six consecutive monthly YoY values. Let A=mean latest3 and B=mean prior3;
  score `0.7×C(50+5×A) + 0.3×C(50+5×(A−B))`. This smooths noise and does
  not claim to seasonally adjust IIP or eliminate base effects.
- Credit: derive monthly growth from YTD as
  `100×((1+YTD_current/100)/(1+YTD_previous/100)−1)` within a calendar year.
  January monthly growth equals January YTD; December's YTD is not subtracted
  from January. Obtain six consecutive monthly growth values (usually seven
  source rows). Score `C(50+25×(mean_latest3−mean_prior3))`, an acceleration
  measure, not a claim that faster credit is always desirable. Different source
  series, including manual gap fills, cannot silently form a trend together.
- EFFR: `C(50−25×change)`. Calendar-date anchor three months earlier, clamped
  at month end, using the last available observation on/before anchor within
  seven days. This handles weekends without interpolating a rate.

Band rules combine level and distance-to-band improvement. Unchanged values
inside their reference band score85, not100; the trend score is neutral50.
A neutral rate change scores50. These are intentional model definitions and
must be validated empirically before interpreting score thresholds as regimes.

## Eligibility, coverage and provenance

Production `macro_observations` and active canonical source policies supply one
row per indicator/period, using the same priority/revision selection as the
indicator monitor. Inputs require valid definition/unit/period, finite values,
raw provenance, revision and first-seen time. Future vintages/publications are
excluded. Reference-period age uses catalog MaxAgeDays: daily EFFR7,
monthly CPI/PMI/IIP60, deposit/credit/M2 75, quarterly PPI120 days.
A source currently returning403 does not invalidate an already saved fresh row.
Manual, unverified production gap fills are explicitly flagged in the output.

Coverage is the sum of nominal weights whose full rule can be evaluated.
Effective weight = nominal weight / eligible coverage ×100. Contributions use
that effective weight and sum to the displayed total. Group scores normalize
within their own eligible subset; group coverage is relative to group weight.
No eligible components: score=null, status=INSUFFICIENT. Otherwise status is
PROVISIONAL unless coverage≥80 **and** deposit, CPI and PMI are all eligible.
Only then status=READY; this means data eligibility, not investment certainty.
The UI warns that a change in coverage changes the scoring basket. Contribution
deltas include that normalization effect and are not causal attribution.

## Operation and storage

Migration011 adds only `macro_score_snapshots` for derived output. Raw historical
indicator data still lives in `macro_observations`; no parallel indicator-history
table is introduced. Starting the rebuilt Go API applies migrations automatically.

A separate worker polls production rows every minute, with a30second timeout,
including writes made directly by Python, manual SQL, revisions and canonical
selection changes. It does not wait for network collection jobs. Each poll
reads its canonical input set in one SQL statement. A transaction advisory lock
serializes replicas. Snapshots retain the exact input observations, raw IDs,
source/method, first-seen time, revision, weights and configuration hash.

EVENT snapshots append only when the input/eligibility fingerprint changes.
A collection timestamp refresh alone is not a new information vintage. Expiring
observations changes eligibility and triggers recalculation even without a new
source value. Daily/weekly snapshots append once per bucket after18:00 Vietnam
(UTC+7); weekly buckets are Friday only. They use **actual execution time**,
not a pretended historical close. Unique keys and conflict-ignore enforce
idempotence. Downtime gaps are not backfilled. Source data backfill can update
today's score but cannot create a historical score that the system never knew.

`GET /api/macro/score?frequency=daily|weekly` returns model metadata, latest
snapshot and up to365 daily/weekly points in ascending order. GET is read-only.
Before first worker run: current=null. Storage failure:503. Invalid frequency:400.
Legacy `/dashboard.score` remains null for compatibility; the React panel uses
the explicit versioned score endpoint and fails independently of indicator data.
Charts use step lines, with no invented points. Snapshot components provide
the reproducible explanation and the UI compares contributions with the previous
available snapshot for the chosen frequency.

Deploy using the existing compose workflow, for example:

```sh
docker compose -f docker-compose.macro.yml up -d --build macro-api
curl -fsS 'http://localhost:3000/api/macro/score?frequency=daily'
```

Rebuild/restart the frontend using your normal deployment as well. Allow up to
one minute for the first score. Historical daily/weekly charts accumulate from
deployment; an initially empty chart is expected, even if raw history exists.

## Verification

```sh
cd apps/macro-api
go test -race ./...
go vet ./...
# Set MACRO_TEST_DATABASE_URL to exercise PostgreSQL tests (also run in CI).
```

Tests cover coverage/critical gates, contribution arithmetic, freshness, future
vintages, missing trends, upper-bound revisions, credit January reset, PMI
momentum, deflation penalties, methodology changes, deduplication across workers,
immutable old vintages, Friday/day buckets and no retrospective gap fabrication.
Frontend tests cover schema consistency, missing storage, frequency selection,
and keeping indicator data independent of score failures.
