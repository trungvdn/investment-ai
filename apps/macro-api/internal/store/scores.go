package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/scoring"
)

// RefreshMacroScore polls production observations, including direct Python/SQL
// writes. An advisory lock serializes replicas. No network ingestion is involved.
func (p *Postgres) RefreshMacroScore(ctx context.Context, now time.Time) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var locked bool
	if err = tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock(614074121)`).Scan(&locked); err != nil {
		return err
	}
	if !locked {
		return nil
	}
	model := scoring.DefaultModel()
	codes := []string{}
	for _, r := range model.Rules {
		codes = append(codes, r.Code)
	}
	// One statement provides a consistent canonical input set. Source-policy changes
	// are respected exactly as on the indicator monitor. Manual gap fills compete
	// by their deliberately low priority and never override a higher-priority row.
	rows, err := tx.Query(ctx, `SELECT DISTINCT ON (indicator_code,period_end) `+observationColumns+`
 FROM macro_observations o WHERE indicator_code=ANY($1::text[]) AND first_seen_at <= $2
 AND period_end >= ($2::timestamptz - interval '2 years')
 AND (series_id='MANUAL_USER_20261027' OR NOT EXISTS
 (SELECT 1 FROM macro_source_policies p WHERE p.indicator_code=o.indicator_code AND p.active_batch_id IS NOT NULL)
 OR id IN (SELECT l.observation_id FROM macro_canonical_links l JOIN macro_source_policies p ON p.active_batch_id=l.batch_id WHERE p.indicator_code=o.indicator_code))
 ORDER BY indicator_code,period_end,priority DESC,revision DESC,collected_at DESC`, codes, now)
	if err != nil {
		return err
	}
	data := map[string][]macro.Observation{}
	for rows.Next() {
		var o macro.Observation
		if err = rows.Scan(&o.IndicatorCode, &o.Period, &o.PeriodEnd, &o.Value, &o.Unit, &o.Source, &o.SourceURL, &o.SeriesID, &o.Method, &o.Priority, &o.PublishedAt, &o.FirstSeenAt, &o.CollectedAt, &o.Revision, &o.RawID, &o.UpperValue); err != nil {
			rows.Close()
			return err
		}
		data[o.IndicatorCode] = append(data[o.IndicatorCode], o)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	snapshot := scoring.Calculate(model, data, now)
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}
	var lastHash string
	var lastAsOf time.Time
	err = tx.QueryRow(ctx, `SELECT input_hash,as_of FROM macro_score_snapshots WHERE model_hash=$1 ORDER BY as_of DESC,id DESC LIMIT 1`, model.Hash).Scan(&lastHash, &lastAsOf)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if now.Before(lastAsOf) {
		return nil
	} // A delayed replica must not append an older calculation.
	insert := func(kind, bucket string) error {
		_, err := tx.Exec(ctx, `INSERT INTO macro_score_snapshots(model_version,model_hash,kind,bucket,as_of,input_hash,payload) VALUES($1,$2,$3,$4,$5,$6,$7) ON CONFLICT(model_hash,kind,bucket) DO NOTHING`, model.Version, model.Hash, kind, bucket, now, snapshot.InputHash, payload)
		return err
	}
	if lastHash != snapshot.InputHash {
		if err = insert("EVENT", now.UTC().Format(time.RFC3339Nano)); err != nil {
			return err
		}
	}
	local := now.In(time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60))
	// Actual execution time is retained; missed days/weeks are never fabricated.
	if local.Hour() >= 18 {
		if err = insert("DAILY", local.Format("2006-01-02")); err != nil {
			return err
		}
		if local.Weekday() == time.Friday {
			if err = insert("WEEKLY", local.Format("2006-01-02")); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (p *Postgres) MacroScores(ctx context.Context, frequency string) (scoring.Response, error) {
	model := scoring.DefaultModel()
	out := scoring.Response{Model: model, Frequency: frequency, History: []scoring.Point{}}
	// One SQL statement keeps current and history on the same database snapshot.
	rows, err := p.Pool.Query(ctx, `WITH latest AS (
 SELECT id FROM macro_score_snapshots WHERE model_hash=$1 ORDER BY as_of DESC,id DESC LIMIT 1
 ), history AS (
 SELECT id FROM macro_score_snapshots WHERE model_hash=$1 AND kind=$2 ORDER BY as_of DESC,id DESC LIMIT 365
 ) SELECT kind,payload,id IN (SELECT id FROM latest) AS current
 FROM macro_score_snapshots WHERE id IN (SELECT id FROM latest UNION SELECT id FROM history) ORDER BY as_of,id`, model.Hash, map[string]string{"daily": "DAILY", "weekly": "WEEKLY"}[frequency])
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var kind string
		var raw []byte
		var current bool
		var s scoring.Snapshot
		if err = rows.Scan(&kind, &raw, &current); err != nil {
			return out, err
		}
		if err = json.Unmarshal(raw, &s); err != nil {
			return out, err
		}
		if current {
			copy := s
			out.Current = &copy
		}
		if kind == map[string]string{"daily": "DAILY", "weekly": "WEEKLY"}[frequency] {
			out.History = append(out.History, scoring.Point{Kind: kind, Snapshot: s})
		}
	}
	return out, rows.Err()
}
