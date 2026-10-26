package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/trungvdn/investment-ai/apps/macro-api/internal/macro"
	"time"
)

func (p *Postgres) ClaimSource(ctx context.Context, id string, now time.Time) (string, error) {
	if _, err := p.Pool.Exec(ctx, `INSERT INTO macro_source_state(provider,next_due) VALUES($1,$2) ON CONFLICT DO NOTHING`, id, now); err != nil {
		return "", err
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	var claimed string
	err := p.Pool.QueryRow(ctx, `UPDATE macro_source_state SET lease_until=$2::timestamptz+interval '10 minutes',lease_token=$3 WHERE provider=$1 AND next_due <= $2 AND (lease_until IS NULL OR lease_until <= $2) RETURNING lease_token`, id, now, token).Scan(&claimed)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return claimed, err
}
func (p *Postgres) FinishSource(ctx context.Context, id, token string, now, next time.Time, kind, message string) error {
	tag, err := p.Pool.Exec(ctx, `UPDATE macro_source_state SET next_due=$4,last_checked=$3,last_success=CASE WHEN $5='' THEN $3 ELSE last_success END,consecutive_failures=CASE WHEN $5='' THEN 0 ELSE consecutive_failures+1 END,error_kind=$5,message=$6,lease_until=NULL,lease_token=NULL WHERE provider=$1 AND lease_token=$2 AND lease_until > $3`, id, token, now, next, kind, message)
	if err == nil && tag.RowsAffected() != 1 {
		return errors.New("source lease expired or replaced")
	}
	return err
}
func (p *Postgres) SourceStates(ctx context.Context) ([]macro.SourceState, error) {
	rows, err := p.Pool.Query(ctx, `SELECT provider,next_due,lease_until,last_checked,last_success,consecutive_failures,error_kind,message FROM macro_source_state ORDER BY provider`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []macro.SourceState{}
	for rows.Next() {
		var s macro.SourceState
		if err = rows.Scan(&s.Provider, &s.NextDue, &s.LeaseUntil, &s.LastChecked, &s.LastSuccess, &s.Failures, &s.Kind, &s.Message); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// MakeSourceDue is an operator-only CLI action, not exposed over HTTP.
func (p *Postgres) MakeSourceDue(ctx context.Context, id string) error {
	tag, err := p.Pool.Exec(ctx, `INSERT INTO macro_source_state(provider,next_due) VALUES($1,now()) ON CONFLICT(provider) DO UPDATE SET next_due=now() WHERE macro_source_state.lease_until IS NULL OR macro_source_state.lease_until <= now()`, id)
	if err == nil && tag.RowsAffected() != 1 {
		return errors.New("source has an active lease; wait for the current owner")
	}
	return err
}
