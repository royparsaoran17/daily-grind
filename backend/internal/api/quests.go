package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/royparsaoran/daily-grind/backend/internal/models"
)

var dayNames = []string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}

// streakWindow is the max gap (in days) between completions that still counts
// as "consecutive" for a quest's streak, based on its frequency.
func streakWindow(frequency string) int {
	switch frequency {
	case "weekly":
		return 7
	case "monthly":
		return 31
	default:
		return 1
	}
}

// periodPredicate returns a SQL boolean matching `col` (a DATE column) to the
// current period for the given frequency: same day / week / month.
func periodPredicate(frequency, col string) string {
	switch frequency {
	case "weekly":
		return "date_trunc('week'," + col + ")=date_trunc('week',current_date)"
	case "monthly":
		return "date_trunc('month'," + col + ")=date_trunc('month',current_date)"
	default:
		return col + "=current_date"
	}
}

// scheduleLabel renders a human-readable schedule for the UI.
func scheduleLabel(frequency string, weekday, dom *int) string {
	switch frequency {
	case "weekly":
		if weekday != nil && *weekday >= 0 && *weekday <= 6 {
			return "Setiap " + dayNames[*weekday]
		}
		return "Setiap minggu"
	case "monthly":
		if dom != nil && *dom >= 1 && *dom <= 31 {
			return "Tanggal " + strconv.Itoa(*dom)
		}
		return "Setiap bulan"
	default:
		return "Setiap hari"
	}
}

// rewardFor returns EXP and coins for a difficulty level.
func rewardFor(difficulty string) (exp, coins int) {
	switch difficulty {
	case "easy":
		return 20, 8
	case "hard":
		return 60, 25
	default:
		return 40, 15
	}
}

func (s *Server) handleListQuests(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	out := []models.Quest{}
	err := s.txTZ(r.Context(), uid, func(tx pgx.Tx) error {
		rows, err := tx.Query(r.Context(), `
			SELECT q.id, q.name, q.category_id, c.label, c.icon, c.attribute,
			       q.frequency, q.difficulty, q.exp_reward, q.coin_reward,
			       COALESCE(q.reminder,''), q.weekday, q.day_of_month,
			       -- effective streak: zero once the frequency window has lapsed
			       CASE
			           WHEN q.last_completed_on IS NULL THEN 0
			           WHEN (current_date - q.last_completed_on) <=
			                CASE q.frequency WHEN 'weekly' THEN 7 WHEN 'monthly' THEN 31 ELSE 1 END
			           THEN q.streak
			           ELSE 0 END AS streak,
			       -- done for the CURRENT period (day/week/month)
			       EXISTS (SELECT 1 FROM quest_completions qc
			               WHERE qc.quest_id=q.id AND (
			                 CASE q.frequency
			                   WHEN 'weekly'  THEN date_trunc('week', qc.completed_on)  = date_trunc('week', current_date)
			                   WHEN 'monthly' THEN date_trunc('month', qc.completed_on) = date_trunc('month', current_date)
			                   ELSE qc.completed_on = current_date END)) AS done,
			       -- scheduled for today?
			       CASE q.frequency
			           WHEN 'weekly'  THEN (q.weekday IS NULL OR extract(dow from current_date)::int = q.weekday)
			           WHEN 'monthly' THEN (q.day_of_month IS NULL OR extract(day from current_date)::int = q.day_of_month)
			           ELSE true END AS due_today
			FROM quests q
			JOIN categories c ON c.id=q.category_id
			WHERE q.user_id=$1 AND q.archived=false
			ORDER BY done ASC, q.created_at ASC`, uid)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var q models.Quest
			if err := rows.Scan(&q.ID, &q.Name, &q.CategoryID, &q.Category, &q.Icon, &q.Attribute,
				&q.Frequency, &q.Difficulty, &q.EXPReward, &q.CoinReward, &q.Reminder,
				&q.Weekday, &q.DayOfMonth, &q.Streak, &q.Done, &q.DueToday); err != nil {
				return err
			}
			q.Schedule = scheduleLabel(q.Frequency, q.Weekday, q.DayOfMonth)
			out = append(out, q)
		}
		return rows.Err()
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query quests")
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreateQuest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       string `json:"name"`
		CategoryID string `json:"category_id"`
		Frequency  string `json:"frequency"`
		Difficulty string `json:"difficulty"`
		Reminder   string `json:"reminder"`
		Weekday    *int   `json:"weekday"`
		DayOfMonth *int   `json:"day_of_month"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" || body.CategoryID == "" {
		writeErr(w, http.StatusBadRequest, "nama dan kategori wajib diisi")
		return
	}
	if body.Frequency == "" {
		body.Frequency = "daily"
	}
	if body.Difficulty == "" {
		body.Difficulty = "medium"
	}

	// Only keep the scheduling field relevant to the frequency, and validate it.
	var weekday, dom *int
	switch body.Frequency {
	case "weekly":
		if body.Weekday != nil {
			if *body.Weekday < 0 || *body.Weekday > 6 {
				writeErr(w, http.StatusBadRequest, "hari tidak valid")
				return
			}
			weekday = body.Weekday
		}
	case "monthly":
		if body.DayOfMonth != nil {
			if *body.DayOfMonth < 1 || *body.DayOfMonth > 31 {
				writeErr(w, http.StatusBadRequest, "tanggal tidak valid")
				return
			}
			dom = body.DayOfMonth
		}
	}
	exp, coins := rewardFor(body.Difficulty)

	var id string
	err := s.pool.QueryRow(r.Context(), `
		INSERT INTO quests(user_id,name,category_id,frequency,difficulty,exp_reward,coin_reward,reminder,weekday,day_of_month)
		VALUES ($1,$2,$3,$4,$5,$6,$7,NULLIF($8,''),$9,$10)
		RETURNING id`,
		userID(r), body.Name, body.CategoryID, body.Frequency, body.Difficulty, exp, coins,
		strings.TrimSpace(body.Reminder), weekday, dom).Scan(&id)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "kategori tidak valid atau quest gagal dibuat")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) handleUpdateQuest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name       string `json:"name"`
		CategoryID string `json:"category_id"`
		Frequency  string `json:"frequency"`
		Difficulty string `json:"difficulty"`
		Reminder   string `json:"reminder"`
		Weekday    *int   `json:"weekday"`
		DayOfMonth *int   `json:"day_of_month"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	if body.Name == "" || body.CategoryID == "" {
		writeErr(w, http.StatusBadRequest, "nama dan kategori wajib diisi")
		return
	}
	if body.Frequency == "" {
		body.Frequency = "daily"
	}
	if body.Difficulty == "" {
		body.Difficulty = "medium"
	}

	var weekday, dom *int
	switch body.Frequency {
	case "weekly":
		if body.Weekday != nil {
			if *body.Weekday < 0 || *body.Weekday > 6 {
				writeErr(w, http.StatusBadRequest, "hari tidak valid")
				return
			}
			weekday = body.Weekday
		}
	case "monthly":
		if body.DayOfMonth != nil {
			if *body.DayOfMonth < 1 || *body.DayOfMonth > 31 {
				writeErr(w, http.StatusBadRequest, "tanggal tidak valid")
				return
			}
			dom = body.DayOfMonth
		}
	}
	exp, coins := rewardFor(body.Difficulty)

	tag, err := s.pool.Exec(r.Context(), `
		UPDATE quests SET name=$1, category_id=$2, frequency=$3, difficulty=$4,
			exp_reward=$5, coin_reward=$6, reminder=NULLIF($7,''), weekday=$8, day_of_month=$9
		WHERE id=$10 AND user_id=$11 AND archived=false`,
		body.Name, body.CategoryID, body.Frequency, body.Difficulty, exp, coins,
		strings.TrimSpace(body.Reminder), weekday, dom, r.PathValue("id"), userID(r))
	if err != nil {
		writeErr(w, http.StatusBadRequest, "kategori tidak valid atau quest gagal diperbarui")
		return
	}
	if tag.RowsAffected() == 0 {
		writeErr(w, http.StatusNotFound, "quest tidak ditemukan")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleCompleteQuest(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	questID := r.PathValue("id")

	var (
		expReward, coinReward int
		attribute, frequency  string
	)
	err := pgx.BeginFunc(r.Context(), s.pool, func(tx pgx.Tx) error {
		if err := setSessionTZ(r.Context(), tx, uid); err != nil {
			return err
		}
		err := tx.QueryRow(r.Context(), `
			SELECT q.exp_reward, q.coin_reward, q.frequency, c.attribute
			FROM quests q JOIN categories c ON c.id=q.category_id
			WHERE q.id=$1 AND q.user_id=$2`, questID, uid).
			Scan(&expReward, &coinReward, &frequency, &attribute)
		if err != nil {
			return err
		}

		// Already completed this period? Then it's a no-op (no double reward).
		var already bool
		if err := tx.QueryRow(r.Context(),
			`SELECT EXISTS(SELECT 1 FROM quest_completions
			              WHERE quest_id=$1 AND `+periodPredicate(frequency, "completed_on")+`)`,
			questID).Scan(&already); err != nil {
			return err
		}
		if already {
			return nil
		}

		if _, err := tx.Exec(r.Context(), `
			INSERT INTO quest_completions(quest_id,user_id,exp_awarded,coin_awarded)
			VALUES ($1,$2,$3,$4)
			ON CONFLICT (quest_id, completed_on) DO NOTHING`,
			questID, uid, expReward, coinReward); err != nil {
			return err
		}

		// Advance per-quest streak, resetting if the previous completion fell
		// outside the frequency window.
		if _, err := tx.Exec(r.Context(), `
			UPDATE quests SET
				streak = CASE
					WHEN last_completed_on IS NULL THEN 1
					WHEN (current_date - last_completed_on) <= $2 THEN streak + 1
					ELSE 1 END,
				last_completed_on = current_date
			WHERE id=$1`, questID, streakWindow(frequency)); err != nil {
			return err
		}
		if err := awardRewards(r.Context(), tx, uid, attribute, expReward, coinReward); err != nil {
			return err
		}
		return touchDailyStreak(r.Context(), tx, uid)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "quest tidak ditemukan")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menyelesaikan quest")
		return
	}

	u, _ := s.loadUser(r.Context(), s.pool, uid)
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleUncompleteQuest(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	questID := r.PathValue("id")

	err := pgx.BeginFunc(r.Context(), s.pool, func(tx pgx.Tx) error {
		if err := setSessionTZ(r.Context(), tx, uid); err != nil {
			return err
		}
		var frequency, attribute string
		if err := tx.QueryRow(r.Context(), `
			SELECT q.frequency, c.attribute FROM quests q
			JOIN categories c ON c.id=q.category_id
			WHERE q.id=$1 AND q.user_id=$2`, questID, uid).Scan(&frequency, &attribute); err != nil {
			return err
		}

		// Remove this period's completion (the most recent one in-period).
		var exp, coins int
		err := tx.QueryRow(r.Context(), `
			DELETE FROM quest_completions
			WHERE id = (SELECT id FROM quest_completions
			            WHERE quest_id=$1 AND `+periodPredicate(frequency, "completed_on")+`
			            ORDER BY completed_on DESC LIMIT 1)
			RETURNING exp_awarded, coin_awarded`,
			questID).Scan(&exp, &coins)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // nothing to undo
		}
		if err != nil {
			return err
		}

		if _, err := tx.Exec(r.Context(), `
			UPDATE quests SET
				streak = GREATEST(streak-1,0),
				last_completed_on = (SELECT MAX(completed_on) FROM quest_completions WHERE quest_id=$1)
			WHERE id=$1`, questID); err != nil {
			return err
		}
		return revokeRewards(r.Context(), tx, uid, attribute, exp, coins)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "quest tidak ditemukan")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal membatalkan quest")
		return
	}

	u, _ := s.loadUser(r.Context(), s.pool, uid)
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleDeleteQuest(w http.ResponseWriter, r *http.Request) {
	tag, err := s.pool.Exec(r.Context(),
		`UPDATE quests SET archived=true WHERE id=$1 AND user_id=$2`,
		r.PathValue("id"), userID(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menghapus quest")
		return
	}
	if tag.RowsAffected() == 0 {
		writeErr(w, http.StatusNotFound, "quest tidak ditemukan")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
