package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// freezeCost is the coin price of one streak freeze ("grace day").
const freezeCost = 100

var errInsufficientCoins = errors.New("insufficient coins")

// touchDailyStreak advances the user's daily-activity streak because they did
// something streak-worthy today (completed a quest or a devotional). Missed
// days are bridged with streak freezes when available; otherwise the streak
// resets. Runs inside the caller's transaction.
//
// Rules, where gap = today - last_active_on (in days):
//   - never active  -> streak = 1
//   - gap == 0       -> already counted today, no change
//   - gap == 1       -> streak += 1
//   - gap >= 2       -> missed = gap-1; if enough freezes, spend them and
//                       streak += 1, else streak resets to 1
func touchDailyStreak(ctx context.Context, tx pgx.Tx, userID string) error {
	var streak, freezes int
	var gap *int // NULL when last_active_on is null
	err := tx.QueryRow(ctx, `
		SELECT streak, streak_freezes,
		       CASE WHEN last_active_on IS NULL THEN NULL
		            ELSE (current_date - last_active_on) END
		FROM users WHERE id=$1 FOR UPDATE`, userID).Scan(&streak, &freezes, &gap)
	if err != nil {
		return err
	}

	switch {
	case gap == nil:
		streak = 1
	case *gap == 0:
		// already active today; nothing changes
		return nil
	case *gap == 1:
		streak++
	default: // gap >= 2
		missed := *gap - 1
		if freezes >= missed {
			freezes -= missed
			streak++
		} else {
			streak = 1
		}
	}

	_, err = tx.Exec(ctx,
		`UPDATE users SET streak=$1, streak_freezes=$2, last_active_on=current_date WHERE id=$3`,
		streak, freezes, userID)
	return err
}

// handleBuyFreeze lets a user spend coins to bank a streak freeze.
func (s *Server) handleBuyFreeze(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	err := pgx.BeginFunc(r.Context(), s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(r.Context(),
			`UPDATE users SET coins = coins - $1, streak_freezes = streak_freezes + 1
			 WHERE id=$2 AND coins >= $1`, freezeCost, uid)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return errInsufficientCoins
		}
		return nil
	})
	if err == errInsufficientCoins {
		writeErr(w, http.StatusBadRequest, "koin tidak cukup untuk membeli pelindung streak")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal membeli pelindung streak")
		return
	}
	u, _ := s.loadUser(r.Context(), s.pool, uid)
	writeJSON(w, http.StatusOK, u)
}
