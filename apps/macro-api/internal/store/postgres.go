package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"github.com/trungvdn/investment-ai/apps/macro-api/migrations"
	"time"
)

var schema = migrations.Schema

type Postgres struct{ Pool *pgxpool.Pool }

func Open(ctx context.Context, databaseURL string) (*Postgres, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid DATABASE_URL")
	}
	config.MaxConns = 8
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("database configuration failed")
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database connection failed")
	}
	return &Postgres{pool}, nil
}
func (p *Postgres) Migrate(ctx context.Context) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT pg_advisory_xact_lock(614074120)"); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `CREATE TABLE IF NOT EXISTS macro_schema_migrations(version INTEGER PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT now())`); err != nil {
		return err
	}
	var applied bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM macro_schema_migrations WHERE version=1)`).Scan(&applied); err != nil {
		return err
	}
	if !applied {
		if _, err = tx.Exec(ctx, schema); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO macro_schema_migrations(version) VALUES(1)`); err != nil {
			return err
		}
	}
	if _, err = tx.Exec(ctx, migrations.Operations); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO macro_schema_migrations(version) VALUES(2) ON CONFLICT DO NOTHING`); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, migrations.IngestionPilot); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO macro_schema_migrations(version) VALUES(3) ON CONFLICT DO NOTHING`); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, migrations.CPIReports); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO macro_schema_migrations(version) VALUES(4) ON CONFLICT DO NOTHING`); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, migrations.TradingEconomicsPMI); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO macro_schema_migrations(version) VALUES(5) ON CONFLICT DO NOTHING`); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, migrations.SourceSelection); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO macro_schema_migrations(version) VALUES(6) ON CONFLICT DO NOTHING`); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, migrations.PMIHistory); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO macro_schema_migrations(version) VALUES(7) ON CONFLICT DO NOTHING`); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (p *Postgres) Close() { p.Pool.Close() }
func (p *Postgres) SaveDocument(ctx context.Context, d macro.RawDocument) (int64, error) {
	sum := sha256.Sum256(d.Body)
	var id int64
	err := p.Pool.QueryRow(ctx, `INSERT INTO macro_raw_documents(source,source_url,content_type,checksum,payload,first_fetched_at,last_fetched_at) VALUES($1,$2,$3,$4,$5,$6,$6) ON CONFLICT(source,source_url,checksum) DO UPDATE SET last_fetched_at=GREATEST(macro_raw_documents.last_fetched_at,EXCLUDED.last_fetched_at) RETURNING id`, d.Source, d.URL, d.ContentType, hex.EncodeToString(sum[:]), d.Body, d.FetchedAt).Scan(&id)
	return id, err
}
func (p *Postgres) Save(ctx context.Context, o macro.Observation) error {
	if o.RawID <= 0 || o.CollectedAt.IsZero() {
		return fmt.Errorf("raw provenance and collection time are required")
	}
	if err := macro.Validate(o, time.Now()); err != nil {
		return err
	}
	_, err := p.Pool.Exec(ctx, `SELECT * FROM macro_save_observation_v1($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		o.IndicatorCode, o.Period, o.PeriodEnd, o.Value, o.Unit, o.Source, o.SourceURL, o.SeriesID, o.Method, o.Priority, o.PublishedAt, o.CollectedAt, o.RawID)
	return err
}
func (p *Postgres) RecordAttempt(ctx context.Context, provider string, a macro.Attempt) error {
	_, err := p.Pool.Exec(ctx, `INSERT INTO macro_attempts(indicator_code,provider_id,checked_at,success,message) VALUES($1,$2,$3,$4,$5) ON CONFLICT(indicator_code,provider_id) DO UPDATE SET checked_at=EXCLUDED.checked_at,success=EXCLUDED.success,message=EXCLUDED.message`, a.Code, provider, a.CheckedAt, a.Success, a.Message)
	return err
}

const observationColumns = `indicator_code,period,period_end,value,unit,source,source_url,series_id,method,priority,published_at,first_seen_at,collected_at,revision,raw_id`

func (p *Postgres) History(ctx context.Context, code string, from, to *time.Time, revisions bool) ([]macro.Observation, error) {
	query := `SELECT ` + observationColumns + ` FROM macro_observations WHERE indicator_code=$1 AND ($2::date IS NULL OR period_end >= $2) AND ($3::date IS NULL OR period_end <= $3) ORDER BY period_end,priority DESC,revision DESC`
	if !revisions {
		query = `SELECT DISTINCT ON (period_end) ` + observationColumns + ` FROM macro_observations WHERE indicator_code=$1 AND ($2::date IS NULL OR period_end >= $2) AND ($3::date IS NULL OR period_end <= $3) AND (NOT EXISTS (SELECT 1 FROM macro_source_policies p WHERE p.indicator_code=$1 AND p.active_batch_id IS NOT NULL)
          OR id IN (SELECT l.observation_id FROM macro_canonical_links l JOIN macro_source_policies p ON p.active_batch_id=l.batch_id WHERE p.indicator_code=$1)) ORDER BY period_end,priority DESC,revision DESC,collected_at DESC`
	}
	rows, err := p.Pool.Query(ctx, query, code, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []macro.Observation{}
	for rows.Next() {
		var o macro.Observation
		if err := rows.Scan(&o.IndicatorCode, &o.Period, &o.PeriodEnd, &o.Value, &o.Unit, &o.Source, &o.SourceURL, &o.SeriesID, &o.Method, &o.Priority, &o.PublishedAt, &o.FirstSeenAt, &o.CollectedAt, &o.Revision, &o.RawID); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, rows.Err()
}
func (p *Postgres) Attempts(ctx context.Context) (map[string]macro.Attempt, error) {
	rows, err := p.Pool.Query(ctx, `SELECT DISTINCT ON (indicator_code) indicator_code,checked_at,success,message FROM macro_attempts a WHERE (a.provider_id LIKE 'python:canonical:%') = EXISTS (SELECT 1 FROM macro_source_policies p WHERE p.indicator_code=a.indicator_code AND p.active_batch_id IS NOT NULL) ORDER BY indicator_code,checked_at DESC,success DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]macro.Attempt{}
	for rows.Next() {
		var a macro.Attempt
		if err := rows.Scan(&a.Code, &a.CheckedAt, &a.Success, &a.Message); err != nil {
			return nil, err
		}
		out[a.Code] = a
	}
	return out, rows.Err()
}
