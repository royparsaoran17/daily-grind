package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/royparsaoran/daily-grind/backend/internal/models"
)

const planFaithReward = 15

// planDef is a static reading-plan definition (content lives in code).
type planDef struct {
	id, title, desc, icon string
	days                  []models.ReadingPlanDay
}

// oneChapterPerDay builds an N-day plan reading one chapter per day of a book.
func oneChapterPerDay(id, title, desc, icon, bookID, bookName string, chapters int) planDef {
	days := make([]models.ReadingPlanDay, 0, chapters)
	for c := 1; c <= chapters; c++ {
		label := bookName + " " + strconv.Itoa(c)
		days = append(days, models.ReadingPlanDay{
			Day:      c,
			Label:    label,
			Readings: []models.Reading{{BookID: bookID, Chapter: c, Label: label}},
		})
	}
	return planDef{id, title, desc, icon, days}
}

// readingPlans is the catalog. Add entries here and they appear automatically.
var readingPlans = []planDef{
	oneChapterPerDay("yohanes-21", "Injil Yohanes", "Kenali Yesus lewat Injil Yohanes, satu pasal sehari.", "ph-cross", "JHN", "Yohanes", 21),
	oneChapterPerDay("amsal-31", "Amsal 31 Hari", "Satu pasal Amsal setiap hari — hikmat untuk sebulan.", "ph-lightbulb", "PRO", "Amsal", 31),
	oneChapterPerDay("kisah-28", "Kisah Para Rasul", "Perjalanan gereja mula-mula, satu pasal sehari.", "ph-fire-simple", "ACT", "Kisah Para Rasul", 28),
	oneChapterPerDay("mazmur-30", "Mazmur Pilihan", "30 hari bersama Kitab Mazmur.", "ph-music-notes", "PSA", "Mazmur", 30),
}

func findPlan(id string) *planDef {
	for i := range readingPlans {
		if readingPlans[i].id == id {
			return &readingPlans[i]
		}
	}
	return nil
}

// handleListPlans returns all plans with the user's progress summary.
func (s *Server) handleListPlans(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	// completed-day counts per plan
	counts := map[string]int{}
	enrolled := map[string]bool{}
	rows, err := s.pool.Query(r.Context(),
		`SELECT plan_id, count(*) FROM reading_plan_progress WHERE user_id=$1 GROUP BY plan_id`, uid)
	if err == nil {
		for rows.Next() {
			var pid string
			var n int
			if rows.Scan(&pid, &n) == nil {
				counts[pid] = n
			}
		}
		rows.Close()
	}
	erows, err := s.pool.Query(r.Context(),
		`SELECT plan_id FROM reading_plan_enrollments WHERE user_id=$1`, uid)
	if err == nil {
		for erows.Next() {
			var pid string
			if erows.Scan(&pid) == nil {
				enrolled[pid] = true
			}
		}
		erows.Close()
	}

	out := make([]models.ReadingPlan, 0, len(readingPlans))
	for _, p := range readingPlans {
		out = append(out, models.ReadingPlan{
			ID: p.id, Title: p.title, Description: p.desc, Icon: p.icon,
			TotalDays: len(p.days), Enrolled: enrolled[p.id], Completed: counts[p.id],
			FaithReward: planFaithReward,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// handlePlanDetail returns one plan with day-by-day readings + completion state.
func (s *Server) handlePlanDetail(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	p := findPlan(r.PathValue("id"))
	if p == nil {
		writeErr(w, http.StatusNotFound, "rencana tidak ditemukan")
		return
	}

	done := map[int]bool{}
	rows, err := s.pool.Query(r.Context(),
		`SELECT day_no FROM reading_plan_progress WHERE user_id=$1 AND plan_id=$2`, uid, p.id)
	if err == nil {
		for rows.Next() {
			var d int
			if rows.Scan(&d) == nil {
				done[d] = true
			}
		}
		rows.Close()
	}
	var enrolled bool
	_ = s.pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM reading_plan_enrollments WHERE user_id=$1 AND plan_id=$2)`, uid, p.id).Scan(&enrolled)

	days := make([]models.ReadingPlanDay, len(p.days))
	completed := 0
	for i, d := range p.days {
		d.Completed = done[d.Day]
		if d.Completed {
			completed++
		}
		days[i] = d
	}
	writeJSON(w, http.StatusOK, models.ReadingPlan{
		ID: p.id, Title: p.title, Description: p.desc, Icon: p.icon,
		TotalDays: len(p.days), Enrolled: enrolled, Completed: completed,
		FaithReward: planFaithReward, Days: days,
	})
}

func (s *Server) handleEnrollPlan(w http.ResponseWriter, r *http.Request) {
	if findPlan(r.PathValue("id")) == nil {
		writeErr(w, http.StatusNotFound, "rencana tidak ditemukan")
		return
	}
	if _, err := s.pool.Exec(r.Context(),
		`INSERT INTO reading_plan_enrollments(user_id,plan_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		userID(r), r.PathValue("id")); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal bergabung rencana")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLeavePlan(w http.ResponseWriter, r *http.Request) {
	// Leaving removes enrollment but keeps progress (so re-joining resumes).
	if _, err := s.pool.Exec(r.Context(),
		`DELETE FROM reading_plan_enrollments WHERE user_id=$1 AND plan_id=$2`,
		userID(r), r.PathValue("id")); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal keluar rencana")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleCompletePlanDay marks a day read and grants FAITH (once per day).
func (s *Server) handleCompletePlanDay(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	p := findPlan(r.PathValue("id"))
	if p == nil {
		writeErr(w, http.StatusNotFound, "rencana tidak ditemukan")
		return
	}
	dayNo, err := strconv.Atoi(r.PathValue("day"))
	if err != nil || dayNo < 1 || dayNo > len(p.days) {
		writeErr(w, http.StatusBadRequest, "hari tidak valid")
		return
	}

	err = pgx.BeginFunc(r.Context(), s.pool, func(tx pgx.Tx) error {
		if err := setSessionTZ(r.Context(), tx, uid); err != nil {
			return err
		}
		// Ensure enrolled.
		if _, err := tx.Exec(r.Context(),
			`INSERT INTO reading_plan_enrollments(user_id,plan_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, uid, p.id); err != nil {
			return err
		}
		tag, err := tx.Exec(r.Context(),
			`INSERT INTO reading_plan_progress(user_id,plan_id,day_no) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`,
			uid, p.id, dayNo)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return nil // already done, no double reward
		}
		if err := awardRewards(r.Context(), tx, uid, "faith", planFaithReward, 0); err != nil {
			return err
		}
		return touchDailyStreak(r.Context(), tx, uid)
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menyimpan progres")
		return
	}
	u, _ := s.loadUser(r.Context(), s.pool, uid)
	writeJSON(w, http.StatusOK, u)
}

// handleUncompletePlanDay unmarks a day (revokes the FAITH reward).
func (s *Server) handleUncompletePlanDay(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	dayNo, _ := strconv.Atoi(r.PathValue("day"))
	err := pgx.BeginFunc(r.Context(), s.pool, func(tx pgx.Tx) error {
		tag, err := tx.Exec(r.Context(),
			`DELETE FROM reading_plan_progress WHERE user_id=$1 AND plan_id=$2 AND day_no=$3`,
			uid, r.PathValue("id"), dayNo)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return nil
		}
		return revokeRewards(r.Context(), tx, uid, "faith", planFaithReward, 0)
	})
	if errors.Is(err, pgx.ErrNoRows) || err == nil {
		u, _ := s.loadUser(r.Context(), s.pool, uid)
		writeJSON(w, http.StatusOK, u)
		return
	}
	writeErr(w, http.StatusInternalServerError, "gagal membatalkan progres")
}
