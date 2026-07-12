package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/royparsaoran/daily-grind/backend/internal/models"
)

// handleBible returns verses for a book/chapter. Accepts ?bookId=PSA (USFM) or
// ?book=Mazmur (name), plus ?chapter=. Defaults to Mazmur 23.
func (s *Server) handleBible(w http.ResponseWriter, r *http.Request) {
	bookID := r.URL.Query().Get("bookId")
	if bookID == "" {
		// Resolve by (case-insensitive) name if provided, else default to PSA.
		if name := r.URL.Query().Get("book"); name != "" {
			_ = s.pool.QueryRow(r.Context(),
				`SELECT id FROM bible_books WHERE upper(name)=upper($1)`, name).Scan(&bookID)
		}
		if bookID == "" {
			bookID = "PSA"
		}
	}
	chapter, err := strconv.Atoi(r.URL.Query().Get("chapter"))
	if err != nil || chapter <= 0 {
		chapter = 23
	}

	var bookName string
	_ = s.pool.QueryRow(r.Context(), `SELECT name FROM bible_books WHERE id=$1`, bookID).Scan(&bookName)

	rows, err := s.pool.Query(r.Context(), `
		SELECT verse, text_id, COALESCE(text_en,''), COALESCE(meaning,'')
		FROM bible_verses WHERE book_id=$1 AND chapter=$2 ORDER BY verse`, bookID, chapter)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query bible")
		return
	}
	defer rows.Close()

	out := []models.Verse{}
	for rows.Next() {
		var v models.Verse
		if err := rows.Scan(&v.Verse, &v.TextID, &v.TextEN, &v.Meaning); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan verse")
			return
		}
		out = append(out, v)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"book_id": bookID, "book": bookName, "chapter": chapter, "verses": out,
	})
}

// handleBibleBooks lists canonical books with the chapter numbers that actually
// have verses, so the picker reflects whatever content is loaded (curated seed
// or full import).
func (s *Server) handleBibleBooks(w http.ResponseWriter, r *http.Request) {
	rows, err := s.pool.Query(r.Context(), `
		SELECT bb.id, bb.name, bb.ordinal, bb.testament,
		       array_agg(DISTINCT v.chapter ORDER BY v.chapter) AS chapters
		FROM bible_books bb
		JOIN bible_verses v ON v.book_id = bb.id
		GROUP BY bb.id, bb.name, bb.ordinal, bb.testament
		ORDER BY bb.ordinal`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query books")
		return
	}
	defer rows.Close()

	out := []models.BibleBook{}
	for rows.Next() {
		var b models.BibleBook
		if err := rows.Scan(&b.ID, &b.Name, &b.Ordinal, &b.Testament, &b.Chapters); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan book")
			return
		}
		out = append(out, b)
	}
	writeJSON(w, http.StatusOK, out)
}

// materializeToday ensures a devotional row exists for today by copying one
// pool entry, chosen deterministically from the date so every user sees the
// same devotional each day and it is stable across requests. This is the
// offline generator; a production upgrade could instead have a scheduled job
// write an LLM-authored devotional for tomorrow into this same table.
func (s *Server) materializeToday(r *http.Request) error {
	_, err := s.pool.Exec(r.Context(), `
		INSERT INTO devotionals(for_date,title,passage,verse_text,reflection,prayer,faith_reward)
		SELECT current_date, title, passage, verse_text, reflection, prayer, faith_reward
		FROM devotional_pool
		ORDER BY id
		OFFSET (SELECT CASE WHEN count(*)=0 THEN 0
		               ELSE mod((current_date - DATE '1970-01-01')::int, count(*)::int) END
		        FROM devotional_pool)
		LIMIT 1
		ON CONFLICT (for_date) DO NOTHING`)
	return err
}

func (s *Server) handleTodayDevotional(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	if err := s.materializeToday(r); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menyiapkan renungan")
		return
	}
	var d models.Devotional
	var dateVal string
	err := s.pool.QueryRow(r.Context(), `
		SELECT d.id, to_char(d.for_date,'YYYY-MM-DD'), d.title, d.passage, d.verse_text,
		       d.reflection, d.prayer, d.faith_reward,
		       EXISTS (SELECT 1 FROM devotional_completions dc
		               WHERE dc.devotional_id=d.id AND dc.user_id=$1) AS completed
		FROM devotionals d
		WHERE d.for_date=current_date
		ORDER BY d.for_date DESC LIMIT 1`, uid).
		Scan(&d.ID, &dateVal, &d.Title, &d.Passage, &d.VerseText, &d.Reflection, &d.Prayer, &d.FaithReward, &d.Completed)
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "belum ada renungan hari ini")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query devotional")
		return
	}
	d.Date = dateVal
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) handleCompleteDevotional(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	devID := r.PathValue("id")

	err := pgx.BeginFunc(r.Context(), s.pool, func(tx pgx.Tx) error {
		var reward int
		if err := tx.QueryRow(r.Context(),
			`SELECT faith_reward FROM devotionals WHERE id=$1`, devID).Scan(&reward); err != nil {
			return err
		}
		tag, err := tx.Exec(r.Context(), `
			INSERT INTO devotional_completions(devotional_id,user_id)
			VALUES ($1,$2) ON CONFLICT DO NOTHING`, devID, uid)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return nil // already completed
		}
		// Devotionals grow FAITH and grant EXP equal to the faith reward.
		if err := awardRewards(r.Context(), tx, uid, "faith", reward, 0); err != nil {
			return err
		}
		return touchDailyStreak(r.Context(), tx, uid)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		writeErr(w, http.StatusNotFound, "renungan tidak ditemukan")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menyelesaikan renungan")
		return
	}

	u, _ := s.loadUser(r.Context(), s.pool, uid)
	writeJSON(w, http.StatusOK, u)
}
