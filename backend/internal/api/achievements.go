package api

import (
	"net/http"

	"github.com/royparsaoran/daily-grind/backend/internal/models"
)

// achievement definitions. `unlock` decides whether the user has earned it from
// their live stats; `target`/`stat` drive the progress bar the client renders.
type achievementDef struct {
	id, name, icon, color, hint string
	stat                        string // which stat the progress compares against
	target                      int
}

var achievementDefs = []achievementDef{
	{"quests-10", "Pemula", "ph-fill ph-seal-check", "#2c9c68", "Selesaikan 10 quest", "quests", 10},
	{"streak-7", "Streak 7", "ph-fill ph-fire", "#e09a24", "Jaga streak 7 hari", "streak", 7},
	{"level-10", "Level 10", "ph-fill ph-medal", "#8c2f3a", "Capai level 10", "level", 10},
	{"quests-100", "100 Quest", "ph-fill ph-trophy", "#2c9c68", "Selesaikan 100 quest", "quests", 100},
	{"streak-30", "Streak 30", "ph-fill ph-flame", "#e0574f", "Jaga streak 30 hari", "streak", 30},
	{"devo-7", "Renungan 7", "ph-fill ph-book-open", "#c88a1c", "Selesaikan 7 renungan", "devotionals", 7},
	{"coins-1000", "Sultan", "ph-fill ph-coins", "#b57c17", "Kumpulkan 1.000 koin", "coins", 1000},
	{"level-25", "Legenda", "ph-fill ph-crown", "#8c2f3a", "Capai level 25", "level", 25},
}

func (s *Server) handleAchievements(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)

	var level, streak, coins, questsDone, devosDone int
	err := s.pool.QueryRow(r.Context(), `
		SELECT
			u.level,
			-- effective (break-aware) streak, same rule as loadUser
			CASE WHEN u.last_active_on IS NULL THEN 0
			     WHEN (current_date - u.last_active_on) <= 1 THEN u.streak
			     WHEN u.streak_freezes >= (current_date - u.last_active_on - 1) THEN u.streak
			     ELSE 0 END,
			u.coins,
			(SELECT count(*) FROM quest_completions qc WHERE qc.user_id=u.id),
			(SELECT count(*) FROM devotional_completions dc WHERE dc.user_id=u.id)
		FROM users u WHERE u.id=$1`, uid).
		Scan(&level, &streak, &coins, &questsDone, &devosDone)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal memuat pencapaian")
		return
	}

	stats := map[string]int{
		"level":       level,
		"streak":      streak,
		"coins":       coins,
		"quests":      questsDone,
		"devotionals": devosDone,
	}

	out := make([]models.Achievement, 0, len(achievementDefs))
	for _, d := range achievementDefs {
		cur := stats[d.stat]
		out = append(out, models.Achievement{
			ID:       d.id,
			Name:     d.name,
			Icon:     d.icon,
			Color:    d.color,
			Hint:     d.hint,
			Unlocked: cur >= d.target,
			Progress: cur,
			Target:   d.target,
		})
	}
	writeJSON(w, http.StatusOK, out)
}
