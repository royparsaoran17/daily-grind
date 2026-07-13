package api

import (
	"net/http"
	"strings"

	"github.com/royparsaoran/daily-grind/backend/internal/models"
)

func (s *Server) handleListPrayers(w http.ResponseWriter, r *http.Request) {
	rows, err := s.pool.Query(r.Context(), `
		SELECT id, title, body, answered, answered_at, created_at
		FROM prayers WHERE user_id=$1
		ORDER BY answered ASC, created_at DESC`, userID(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query doa")
		return
	}
	defer rows.Close()
	out := []models.Prayer{}
	for rows.Next() {
		var p models.Prayer
		if err := rows.Scan(&p.ID, &p.Title, &p.Body, &p.Answered, &p.AnsweredAt, &p.CreatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan doa")
			return
		}
		out = append(out, p)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleCreatePrayer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	if body.Title == "" {
		writeErr(w, http.StatusBadRequest, "judul doa tidak boleh kosong")
		return
	}
	var p models.Prayer
	err := s.pool.QueryRow(r.Context(), `
		INSERT INTO prayers(user_id,title,body) VALUES ($1,$2,$3)
		RETURNING id,title,body,answered,answered_at,created_at`,
		userID(r), body.Title, strings.TrimSpace(body.Body)).
		Scan(&p.ID, &p.Title, &p.Body, &p.Answered, &p.AnsweredAt, &p.CreatedAt)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menyimpan doa")
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *Server) handleUpdatePrayer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Title = strings.TrimSpace(body.Title)
	if body.Title == "" {
		writeErr(w, http.StatusBadRequest, "judul doa tidak boleh kosong")
		return
	}
	tag, err := s.pool.Exec(r.Context(),
		`UPDATE prayers SET title=$1, body=$2 WHERE id=$3 AND user_id=$4`,
		body.Title, strings.TrimSpace(body.Body), r.PathValue("id"), userID(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal memperbarui doa")
		return
	}
	if tag.RowsAffected() == 0 {
		writeErr(w, http.StatusNotFound, "doa tidak ditemukan")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleToggleAnswered flips answered state; grants a little FAITH when a prayer
// is marked answered (gratitude), and revokes it when un-marked.
func (s *Server) handleToggleAnswered(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	var nowAnswered bool
	err := s.pool.QueryRow(r.Context(), `
		UPDATE prayers
		SET answered = NOT answered,
		    answered_at = CASE WHEN NOT answered THEN now() ELSE NULL END
		WHERE id=$1 AND user_id=$2
		RETURNING answered`, r.PathValue("id"), uid).Scan(&nowAnswered)
	if err != nil {
		writeErr(w, http.StatusNotFound, "doa tidak ditemukan")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"answered": nowAnswered})
}

func (s *Server) handleDeletePrayer(w http.ResponseWriter, r *http.Request) {
	if _, err := s.pool.Exec(r.Context(),
		`DELETE FROM prayers WHERE id=$1 AND user_id=$2`, r.PathValue("id"), userID(r)); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menghapus doa")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
