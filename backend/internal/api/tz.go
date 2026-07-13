package api

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

// setSessionTZ sets the PostgreSQL session timezone (local to the transaction)
// to the user's saved timezone. After this, current_date / now() / date_trunc()
// and the completed_on DEFAULT all evaluate in the user's local time.
func setSessionTZ(ctx context.Context, tx pgx.Tx, userID string) error {
	_, err := tx.Exec(ctx,
		`SELECT set_config('timezone', COALESCE((SELECT timezone FROM users WHERE id=$1),'UTC'), true)`,
		userID)
	return err
}

// txTZ runs read-only work inside a transaction whose timezone is the user's,
// so date-based queries use the correct local "today".
func (s *Server) txTZ(ctx context.Context, userID string, fn func(tx pgx.Tx) error) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		if err := setSessionTZ(ctx, tx, userID); err != nil {
			return err
		}
		return fn(tx)
	})
}

// handleSetTimezone stores the user's IANA timezone (validated).
func (s *Server) handleSetTimezone(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Timezone string `json:"timezone"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if _, err := time.LoadLocation(body.Timezone); err != nil {
		writeErr(w, http.StatusBadRequest, "timezone tidak valid")
		return
	}
	if _, err := s.pool.Exec(r.Context(),
		`UPDATE users SET timezone=$1 WHERE id=$2`, body.Timezone, userID(r)); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menyimpan timezone")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
