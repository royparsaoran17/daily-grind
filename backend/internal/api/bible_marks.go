package api

import (
	"net/http"
	"strconv"

	"github.com/royparsaoran/daily-grind/backend/internal/models"
)

// handleListMarks returns the marks for one chapter (?book_id=&chapter=), so the
// reader can render highlights and bookmark indicators.
func (s *Server) handleListMarks(w http.ResponseWriter, r *http.Request) {
	bookID := r.URL.Query().Get("book_id")
	chapter, _ := strconv.Atoi(r.URL.Query().Get("chapter"))
	rows, err := s.pool.Query(r.Context(), `
		SELECT book_id, chapter, verse, kind FROM bible_marks
		WHERE user_id=$1 AND book_id=$2 AND chapter=$3`, userID(r), bookID, chapter)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query marks")
		return
	}
	defer rows.Close()
	out := []models.BibleMark{}
	for rows.Next() {
		var m models.BibleMark
		if err := rows.Scan(&m.BookID, &m.Chapter, &m.Verse, &m.Kind); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan mark")
			return
		}
		out = append(out, m)
	}
	writeJSON(w, http.StatusOK, out)
}

// handleToggleMark adds or removes a highlight/bookmark on a verse.
func (s *Server) handleToggleMark(w http.ResponseWriter, r *http.Request) {
	var body struct {
		BookID  string `json:"book_id"`
		Chapter int    `json:"chapter"`
		Verse   int    `json:"verse"`
		Kind    string `json:"kind"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Kind != "highlight" && body.Kind != "bookmark" {
		writeErr(w, http.StatusBadRequest, "kind tidak valid")
		return
	}
	if body.BookID == "" || body.Chapter <= 0 || body.Verse <= 0 {
		writeErr(w, http.StatusBadRequest, "referensi ayat tidak valid")
		return
	}

	uid := userID(r)
	tag, err := s.pool.Exec(r.Context(),
		`DELETE FROM bible_marks WHERE user_id=$1 AND book_id=$2 AND chapter=$3 AND verse=$4 AND kind=$5`,
		uid, body.BookID, body.Chapter, body.Verse, body.Kind)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal memproses tanda")
		return
	}
	marked := false
	if tag.RowsAffected() == 0 {
		if _, err := s.pool.Exec(r.Context(),
			`INSERT INTO bible_marks(user_id,book_id,chapter,verse,kind) VALUES ($1,$2,$3,$4,$5)
			 ON CONFLICT DO NOTHING`,
			uid, body.BookID, body.Chapter, body.Verse, body.Kind); err != nil {
			writeErr(w, http.StatusInternalServerError, "gagal menyimpan tanda")
			return
		}
		marked = true
	}
	writeJSON(w, http.StatusOK, map[string]bool{"marked": marked})
}

// handleListBookmarks returns all bookmarked verses with their text.
func (s *Server) handleListBookmarks(w http.ResponseWriter, r *http.Request) {
	rows, err := s.pool.Query(r.Context(), `
		SELECT m.book_id, COALESCE(bb.name, m.book_id), m.chapter, m.verse, COALESCE(v.text_id,'')
		FROM bible_marks m
		LEFT JOIN bible_books bb ON bb.id = m.book_id
		LEFT JOIN bible_verses v ON v.book_id=m.book_id AND v.chapter=m.chapter AND v.verse=m.verse
		WHERE m.user_id=$1 AND m.kind='bookmark'
		ORDER BY m.created_at DESC`, userID(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query bookmarks")
		return
	}
	defer rows.Close()
	out := []models.Bookmark{}
	for rows.Next() {
		var b models.Bookmark
		if err := rows.Scan(&b.BookID, &b.Book, &b.Chapter, &b.Verse, &b.Text); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan bookmark")
			return
		}
		out = append(out, b)
	}
	writeJSON(w, http.StatusOK, out)
}
