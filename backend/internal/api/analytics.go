package api

import (
	"net/http"

	"github.com/royparsaoran/daily-grind/backend/internal/models"
)

var shortDays = []string{"Min", "Sen", "Sel", "Rab", "Kam", "Jum", "Sab"}

// handleAnalytics returns progress + history data for the charts:
//   - daily completions & EXP for the last 14 days
//   - completions grouped by category
//   - current attribute distribution + a few headline totals
func (s *Server) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	ctx := r.Context()
	const days = 14

	var a models.Analytics

	// Daily series (last 14 days), zero-filled via generate_series.
	rows, err := s.pool.Query(ctx, `
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
		writeErr(w, http.StatusInternalServerError, "query analytics")
		return
	}
	defer rows.Close()

	a.Daily = []models.AnalyticsDay{}
	for rows.Next() {
		var day models.AnalyticsDay
		var dow int
		if err := rows.Scan(&day.Date, &dow, &day.Completions, &day.EXP); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan analytics")
			return
		}
		day.Label = shortDays[dow]
		a.Daily = append(a.Daily, day)
	}
	rows.Close()

	// Completions grouped by category.
	crows, err := s.pool.Query(ctx, `
		SELECT c.label, c.attribute, count(qc.id)
		FROM quest_completions qc
		JOIN quests q ON q.id = qc.quest_id
		JOIN categories c ON c.id = q.category_id
		WHERE qc.user_id = $1
		GROUP BY c.label, c.attribute
		ORDER BY count(qc.id) DESC`, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query category stats")
		return
	}
	defer crows.Close()
	a.ByCategory = []models.CategoryCount{}
	for crows.Next() {
		var cc models.CategoryCount
		if err := crows.Scan(&cc.Category, &cc.Attribute, &cc.Count); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan category stats")
			return
		}
		a.ByCategory = append(a.ByCategory, cc)
	}
	crows.Close()

	// Headline totals + attributes.
	u, err := s.loadUser(ctx, s.pool, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "load user")
		return
	}
	a.Attributes = u.Attr
	a.CurrentStreak = u.Streak

	if err := s.pool.QueryRow(ctx, `
		SELECT
			(SELECT count(*) FROM quest_completions WHERE user_id=$1),
			(SELECT count(DISTINCT completed_on) FROM quest_completions WHERE user_id=$1),
			(SELECT COALESCE(sum(exp_awarded),0) FROM quest_completions
			 WHERE user_id=$1 AND completed_on >= date_trunc('week', current_date))`, uid).
		Scan(&a.TotalCompletions, &a.ActiveDays, &a.ThisWeekEXP); err != nil {
		writeErr(w, http.StatusInternalServerError, "query totals")
		return
	}

	// (All dates are computed in SQL to stay in the DB timezone.)
	writeJSON(w, http.StatusOK, a)
}
