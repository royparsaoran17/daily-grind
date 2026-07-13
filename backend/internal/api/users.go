package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/royparsaoran/daily-grind/backend/internal/models"
)

// expForLevel is the EXP required to advance from `level` to the next.
// Tuned so level 12 -> 13 costs 500, matching the design mockups.
func expForLevel(level int) int {
	return 200 + level*25
}

type querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// loadUser fetches a full user/character record.
func (s *Server) loadUser(ctx context.Context, q querier, id string) (*models.User, error) {
	var u models.User
	// today = the user's LOCAL date (respects their timezone), so streaks break
	// on the right calendar day regardless of server timezone.
	err := q.QueryRow(ctx, `
		SELECT id,name,email,title,level,exp,coins,
		       CASE
		           WHEN last_active_on IS NULL THEN 0
		           WHEN ((now() AT TIME ZONE timezone)::date - last_active_on) <= 1 THEN streak
		           WHEN streak_freezes >= ((now() AT TIME ZONE timezone)::date - last_active_on - 1) THEN streak
		           ELSE 0 END AS streak,
		       streak_freezes, (onboarded_at IS NOT NULL) AS onboarded, locale, COALESCE(avatar_url,''), timezone,
		       str,vit,int_,wis,faith,created_at
		FROM users WHERE id=$1`, id).Scan(
		&u.ID, &u.Name, &u.Email, &u.Title, &u.Level, &u.EXP, &u.Coins, &u.Streak, &u.StreakFreezes, &u.Onboarded, &u.Locale, &u.AvatarURL, &u.Timezone,
		&u.Attr.STR, &u.Attr.VIT, &u.Attr.INT, &u.Attr.WIS, &u.Attr.FAITH, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	u.NextEXP = expForLevel(u.Level)
	return &u, nil
}

// awardRewards adds EXP/coins, applies level-ups, and bumps the mapped
// attribute. Runs inside the caller's transaction.
func awardRewards(ctx context.Context, tx pgx.Tx, userID, attribute string, exp, coins int) error {
	var level, curExp, coinsNow int
	if err := tx.QueryRow(ctx,
		`SELECT level,exp,coins FROM users WHERE id=$1 FOR UPDATE`, userID).
		Scan(&level, &curExp, &coinsNow); err != nil {
		return err
	}

	curExp += exp
	for curExp >= expForLevel(level) {
		curExp -= expForLevel(level)
		level++
	}
	coinsNow += coins

	// Column name is validated against a fixed allow-list by the caller.
	col := map[string]string{
		"str": "str", "vit": "vit", "int": "int_", "wis": "wis", "faith": "faith",
	}[attribute]
	if col == "" {
		col = "str"
	}

	_, err := tx.Exec(ctx,
		`UPDATE users SET level=$1, exp=$2, coins=$3, `+col+`=`+col+`+1 WHERE id=$4`,
		level, curExp, coinsNow, userID)
	return err
}

// revokeRewards reverses awardRewards when a completion is undone.
func revokeRewards(ctx context.Context, tx pgx.Tx, userID, attribute string, exp, coins int) error {
	var level, curExp, coinsNow int
	if err := tx.QueryRow(ctx,
		`SELECT level,exp,coins FROM users WHERE id=$1 FOR UPDATE`, userID).
		Scan(&level, &curExp, &coinsNow); err != nil {
		return err
	}

	curExp -= exp
	for curExp < 0 && level > 1 {
		level--
		curExp += expForLevel(level)
	}
	if curExp < 0 {
		curExp = 0
	}
	coinsNow -= coins
	if coinsNow < 0 {
		coinsNow = 0
	}

	col := map[string]string{
		"str": "str", "vit": "vit", "int": "int_", "wis": "wis", "faith": "faith",
	}[attribute]
	if col == "" {
		col = "str"
	}

	_, err := tx.Exec(ctx,
		`UPDATE users SET level=$1, exp=$2, coins=$3, `+col+`=GREATEST(`+col+`-1,0) WHERE id=$4`,
		level, curExp, coinsNow, userID)
	return err
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	u, err := s.loadUser(r.Context(), s.pool, userID(r))
	if err != nil {
		writeErr(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// handleUpdateMe updates editable profile fields (name and title).
func (s *Server) handleUpdateMe(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name  string `json:"name"`
		Title string `json:"title"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeErr(w, http.StatusBadRequest, "nama tidak boleh kosong")
		return
	}
	// Title is optional; blank keeps the current value.
	if _, err := s.pool.Exec(r.Context(),
		`UPDATE users SET name=$1, title=COALESCE(NULLIF($2,''), title) WHERE id=$3`,
		name, strings.TrimSpace(body.Title), userID(r)); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal memperbarui profil")
		return
	}
	u, err := s.loadUser(r.Context(), s.pool, userID(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal memuat profil")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := s.pool.Query(r.Context(),
		`SELECT id,label,icon,attribute FROM categories ORDER BY label`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query categories")
		return
	}
	defer rows.Close()

	out := []models.Category{}
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Label, &c.Icon, &c.Attribute); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan category")
			return
		}
		out = append(out, c)
	}
	writeJSON(w, http.StatusOK, out)
}
