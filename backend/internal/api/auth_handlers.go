package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/royparsaoran/daily-grind/backend/internal/auth"
)

type authResponse struct {
	Token string `json:"token"`
	User  any    `json:"user"`
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))
	if body.Name == "" || body.Email == "" || len(body.Password) < 6 {
		writeErr(w, http.StatusBadRequest, "name, email and a 6+ char password are required")
		return
	}

	hash, err := auth.HashPassword(body.Password)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not hash password")
		return
	}

	var id string
	err = s.pool.QueryRow(r.Context(), `
		INSERT INTO users(name,email,password_hash) VALUES ($1,$2,$3) RETURNING id`,
		body.Name, body.Email, hash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeErr(w, http.StatusConflict, "email already registered")
			return
		}
		writeErr(w, http.StatusInternalServerError, "could not create account")
		return
	}

	s.issueFor(w, r, id)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Email = strings.ToLower(strings.TrimSpace(body.Email))

	var id, hash string
	err := s.pool.QueryRow(r.Context(),
		`SELECT id,password_hash FROM users WHERE email=$1`, body.Email).Scan(&id, &hash)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !auth.CheckPassword(hash, body.Password)) {
		writeErr(w, http.StatusUnauthorized, "email atau kata sandi salah")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "login failed")
		return
	}

	s.issueFor(w, r, id)
}

// issueFor mints a token and returns it alongside the user profile.
func (s *Server) issueFor(w http.ResponseWriter, r *http.Request, id string) {
	token, err := s.auth.Issue(id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	u, err := s.loadUser(r.Context(), s.pool, id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "could not load user")
		return
	}
	writeJSON(w, http.StatusOK, authResponse{Token: token, User: u})
}
