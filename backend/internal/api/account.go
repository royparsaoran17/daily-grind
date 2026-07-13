package api

import (
	"net/http"

	"github.com/royparsaoran/daily-grind/backend/internal/auth"
)

// handleChangePassword verifies the current password and sets a new one.
func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(body.NewPassword) < 6 {
		writeErr(w, http.StatusBadRequest, "kata sandi baru minimal 6 karakter")
		return
	}

	uid := userID(r)
	var hash string
	if err := s.pool.QueryRow(r.Context(),
		`SELECT password_hash FROM users WHERE id=$1`, uid).Scan(&hash); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal memuat akun")
		return
	}
	if !auth.CheckPassword(hash, body.CurrentPassword) {
		writeErr(w, http.StatusBadRequest, "kata sandi saat ini salah")
		return
	}
	newHash, err := auth.HashPassword(body.NewPassword)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal mengubah kata sandi")
		return
	}
	if _, err := s.pool.Exec(r.Context(),
		`UPDATE users SET password_hash=$1 WHERE id=$2`, newHash, uid); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menyimpan kata sandi")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteAccount permanently removes the account (cascades to all data).
func (s *Server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	uid := userID(r)
	var hash string
	if err := s.pool.QueryRow(r.Context(),
		`SELECT password_hash FROM users WHERE id=$1`, uid).Scan(&hash); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal memuat akun")
		return
	}
	if !auth.CheckPassword(hash, body.Password) {
		writeErr(w, http.StatusBadRequest, "kata sandi salah")
		return
	}
	if _, err := s.pool.Exec(r.Context(), `DELETE FROM users WHERE id=$1`, uid); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menghapus akun")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleSetLocale updates the user's preferred UI language.
func (s *Server) handleSetLocale(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Locale string `json:"locale"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Locale != "id" && body.Locale != "en" {
		writeErr(w, http.StatusBadRequest, "locale tidak didukung")
		return
	}
	if _, err := s.pool.Exec(r.Context(),
		`UPDATE users SET locale=$1 WHERE id=$2`, body.Locale, userID(r)); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menyimpan bahasa")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleOnboard marks onboarding complete.
func (s *Server) handleOnboard(w http.ResponseWriter, r *http.Request) {
	if _, err := s.pool.Exec(r.Context(),
		`UPDATE users SET onboarded_at=now() WHERE id=$1 AND onboarded_at IS NULL`, userID(r)); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menyimpan onboarding")
		return
	}
	u, err := s.loadUser(r.Context(), s.pool, userID(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal memuat profil")
		return
	}
	writeJSON(w, http.StatusOK, u)
}
