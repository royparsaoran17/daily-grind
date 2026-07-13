package api

import (
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/royparsaoran/daily-grind/backend/internal/models"
)

var shortDays = []string{"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"}

// handleHeatmap returns per-day completion counts for the last ~year, for a
// GitHub-style calendar heatmap. Dates are in the user's timezone.
func (s *Server) handleHeatmap(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	const days = 371 // 53 weeks
	out := []models.HeatmapDay{}
	err := s.txTZ(r.Context(), uid, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `
			WITH d AS (
				SELECT generate_series(current_date - $2::int + 1, current_date, interval '1 day')::date AS day
			)
			SELECT to_char(d.day,'YYYY-MM-DD'), COALESCE(count(qc.id),0)
			FROM d
			LEFT JOIN quest_completions qc ON qc.completed_on=d.day AND qc.user_id=$1
			GROUP BY d.day ORDER BY d.day`, uid, days)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var h models.HeatmapDay
			if err := rows.Scan(&h.Date, &h.Count); err != nil {
				return err
			}
			out = append(out, h)
		}
		return rows.Err()
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query heatmap")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// handleAnalytics returns progress + history data for the charts:
//   - daily completions & EXP for the last 14 days
//   - completions grouped by category
//   - current attribute distribution + a few headline totals
func (s *Server) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	ctx := r.Context()
	const days = 14

	var a models.Analytics
	a.Daily = []models.AnalyticsDay{}
	a.ByCategory = []models.CategoryCount{}

	// Everything below runs in the user's timezone so "today", the 14-day window,
	// and the week boundary all align with their local calendar.
	err := s.txTZ(ctx, uid, func(tx pgx.Tx) error {
		// Daily series (last 14 days), zero-filled via generate_series.
		rows, err := tx.Query(ctx, `
			WITH d AS (
				SELECT generate_series(current_date - $2::int + 1, current_date, interval '1 day')::date AS day
			)
			SELECT to_char(d.day,'YYYY-MM-DD'),
			       extract(dow from d.day)::int,
			       COALESCE(count(qc.id), 0),
			       COALESCE(sum(qc.exp_awarded), 0)
			FROM d
			LEFT JOIN quest_completions qc ON qc.completed_on = d.day AND qc.user_id = $1
			GROUP BY d.day
			ORDER BY d.day`, uid, days)
		if err != nil {
			return err
		}
		for rows.Next() {
			var day models.AnalyticsDay
			var dow int
			if err := rows.Scan(&day.Date, &dow, &day.Completions, &day.EXP); err != nil {
				rows.Close()
				return err
			}
			day.Label = shortDays[dow]
			a.Daily = append(a.Daily, day)
		}
		rows.Close()

		// Completions grouped by category.
		crows, err := tx.Query(ctx, `
			SELECT c.label, c.attribute, count(qc.id)
			FROM quest_completions qc
			JOIN quests q ON q.id = qc.quest_id
			JOIN categories c ON c.id = q.category_id
			WHERE qc.user_id = $1
			GROUP BY c.label, c.attribute
			ORDER BY count(qc.id) DESC`, uid)
		if err != nil {
			return err
		}
		for crows.Next() {
			var cc models.CategoryCount
			if err := crows.Scan(&cc.Category, &cc.Attribute, &cc.Count); err != nil {
				crows.Close()
				return err
			}
			a.ByCategory = append(a.ByCategory, cc)
		}
		crows.Close()

		// Headline totals.
		return tx.QueryRow(ctx, `
			SELECT
				(SELECT count(*) FROM quest_completions WHERE user_id=$1),
				(SELECT count(DISTINCT completed_on) FROM quest_completions WHERE user_id=$1),
				(SELECT COALESCE(sum(exp_awarded),0) FROM quest_completions
				 WHERE user_id=$1 AND completed_on >= date_trunc('week', current_date))`, uid).
			Scan(&a.TotalCompletions, &a.ActiveDays, &a.ThisWeekEXP)
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query analytics")
		return
	}

	// Attributes + streak (loadUser is already timezone-aware).
	u, err := s.loadUser(ctx, s.pool, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "load user")
		return
	}
	a.Attributes = u.Attr
	a.CurrentStreak = u.Streak

	writeJSON(w, http.StatusOK, a)
}
