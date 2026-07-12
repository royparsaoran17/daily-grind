package api

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/royparsaoran/daily-grind/backend/internal/models"
)

var dateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// resolveDate maps a path segment to a SQL date expression + arg. "today"
// becomes current_date; otherwise a strict YYYY-MM-DD literal is required.
func resolveDate(seg string) (expr string, arg any, ok bool) {
	if seg == "today" || seg == "" {
		return "current_date", nil, true
	}
	if dateRe.MatchString(seg) {
		return "$2::date", seg, true
	}
	return "", nil, false
}

func (s *Server) handleListJournal(w http.ResponseWriter, r *http.Request) {
	rows, err := s.pool.Query(r.Context(), `
		SELECT id, to_char(entry_date,'YYYY-MM-DD'), title, body,
		       COALESCE(mood,''), COALESCE(prompt,''), updated_at
		FROM journal_entries WHERE user_id=$1
		ORDER BY entry_date DESC`, userID(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query jurnal")
		return
	}
	defer rows.Close()

	out := []models.JournalEntry{}
	for rows.Next() {
		var e models.JournalEntry
		if err := rows.Scan(&e.ID, &e.Date, &e.Title, &e.Body, &e.Mood, &e.Prompt, &e.UpdatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan jurnal")
			return
		}
		out = append(out, e)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleGetJournal(w http.ResponseWriter, r *http.Request) {
	dateExpr, dateArg, ok := resolveDate(r.PathValue("date"))
	if !ok {
		writeErr(w, http.StatusBadRequest, "format tanggal harus YYYY-MM-DD")
		return
	}
	args := []any{userID(r)}
	if dateArg != nil {
		args = append(args, dateArg)
	}

	var e models.JournalEntry
	err := s.pool.QueryRow(r.Context(), `
		SELECT id, to_char(entry_date,'YYYY-MM-DD'), title, body,
		       COALESCE(mood,''), COALESCE(prompt,''), updated_at
		FROM journal_entries
		WHERE user_id=$1 AND entry_date=`+dateExpr, args...).
		Scan(&e.ID, &e.Date, &e.Title, &e.Body, &e.Mood, &e.Prompt, &e.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		// Empty (unsaved) entry for that date — return 200 with a null id so the
		// client can render a blank editor.
		writeJSON(w, http.StatusOK, nil)
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query jurnal")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

// handleUpsertJournal creates or updates the entry for a date.
func (s *Server) handleUpsertJournal(w http.ResponseWriter, r *http.Request) {
	dateExpr, dateArg, ok := resolveDate(r.PathValue("date"))
	if !ok {
		writeErr(w, http.StatusBadRequest, "format tanggal harus YYYY-MM-DD")
		return
	}
	var body struct {
		Title  string `json:"title"`
		Body   string `json:"body"`
		Mood   string `json:"mood"`
		Prompt string `json:"prompt"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if strings.TrimSpace(body.Title) == "" && strings.TrimSpace(body.Body) == "" {
		writeErr(w, http.StatusBadRequest, "judul atau isi jurnal tidak boleh kosong")
		return
	}

	// $1 user, [$2 date if literal], then title/body/mood/prompt.
	args := []any{userID(r)}
	if dateArg != nil {
		args = append(args, dateArg)
	}
	args = append(args, strings.TrimSpace(body.Title), body.Body,
		strings.TrimSpace(body.Mood), strings.TrimSpace(body.Prompt))
	// Placeholder numbers shift when a date literal is present.
	base := len(args) - 4 // index of title arg (0-based count before title)
	p := func(n int) string { return "$" + strconv.Itoa(base+n) }

	var e models.JournalEntry
	err := s.pool.QueryRow(r.Context(), `
		INSERT INTO journal_entries(user_id, entry_date, title, body, mood, prompt)
		VALUES ($1, `+dateExpr+`, `+p(1)+`, `+p(2)+`, NULLIF(`+p(3)+`,''), NULLIF(`+p(4)+`,''))
		ON CONFLICT (user_id, entry_date) DO UPDATE SET
			title=EXCLUDED.title, body=EXCLUDED.body, mood=EXCLUDED.mood,
			prompt=COALESCE(EXCLUDED.prompt, journal_entries.prompt), updated_at=now()
		RETURNING id, to_char(entry_date,'YYYY-MM-DD'), title, body,
		          COALESCE(mood,''), COALESCE(prompt,''), updated_at`, args...).
		Scan(&e.ID, &e.Date, &e.Title, &e.Body, &e.Mood, &e.Prompt, &e.UpdatedAt)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menyimpan jurnal")
		return
	}
	writeJSON(w, http.StatusOK, e)
}

func (s *Server) handleDeleteJournal(w http.ResponseWriter, r *http.Request) {
	dateExpr, dateArg, ok := resolveDate(r.PathValue("date"))
	if !ok {
		writeErr(w, http.StatusBadRequest, "format tanggal harus YYYY-MM-DD")
		return
	}
	args := []any{userID(r)}
	if dateArg != nil {
		args = append(args, dateArg)
	}
	if _, err := s.pool.Exec(r.Context(),
		`DELETE FROM journal_entries WHERE user_id=$1 AND entry_date=`+dateExpr, args...); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menghapus jurnal")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
