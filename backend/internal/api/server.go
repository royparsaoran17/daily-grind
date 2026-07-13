// Package api wires HTTP handlers over the database.
package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/royparsaoran/daily-grind/backend/internal/auth"
	"github.com/royparsaoran/daily-grind/backend/internal/config"
)

// Server bundles shared dependencies for the handlers.
type Server struct {
	pool *pgxpool.Pool
	auth *auth.Service
	cfg  config.Config
}

func NewServer(pool *pgxpool.Pool, authSvc *auth.Service, cfg config.Config) *Server {
	return &Server{pool: pool, auth: authSvc, cfg: cfg}
}

// Routes returns the fully-wired handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Public auth
	mux.HandleFunc("POST /api/auth/register", s.handleRegister)
	mux.HandleFunc("POST /api/auth/login", s.handleLogin)

	// Authenticated
	mux.Handle("GET /api/me", s.authed(s.handleMe))
	mux.Handle("PUT /api/me", s.authed(s.handleUpdateMe))
	mux.Handle("DELETE /api/me", s.authed(s.handleDeleteAccount))
	mux.Handle("PUT /api/me/password", s.authed(s.handleChangePassword))
	mux.Handle("POST /api/me/onboard", s.authed(s.handleOnboard))
	mux.Handle("PUT /api/me/locale", s.authed(s.handleSetLocale))
	mux.Handle("PUT /api/me/timezone", s.authed(s.handleSetTimezone))
	mux.Handle("GET /api/me/avatar/signature", s.authed(s.handleAvatarSignature))
	mux.Handle("PUT /api/me/avatar", s.authed(s.handleSetAvatar))
	mux.Handle("DELETE /api/me/avatar", s.authed(s.handleRemoveAvatar))
	mux.Handle("GET /api/achievements", s.authed(s.handleAchievements))
	mux.Handle("GET /api/categories", s.authed(s.handleCategories))

	mux.Handle("GET /api/quests", s.authed(s.handleListQuests))
	mux.Handle("POST /api/quests", s.authed(s.handleCreateQuest))
	mux.Handle("PUT /api/quests/{id}", s.authed(s.handleUpdateQuest))
	mux.Handle("POST /api/quests/{id}/complete", s.authed(s.handleCompleteQuest))
	mux.Handle("DELETE /api/quests/{id}/complete", s.authed(s.handleUncompleteQuest))
	mux.Handle("DELETE /api/quests/{id}", s.authed(s.handleDeleteQuest))

	mux.Handle("GET /api/friends", s.authed(s.handleFriends))
	mux.Handle("GET /api/friends/requests", s.authed(s.handleListRequests))
	mux.Handle("POST /api/friends/requests/{id}/accept", s.authed(s.handleAcceptRequest))
	mux.Handle("POST /api/friends/requests/{id}/reject", s.authed(s.handleRejectRequest))
	mux.Handle("GET /api/users/search", s.authed(s.handleSearchUsers))
	mux.Handle("POST /api/friends/{id}", s.authed(s.handleAddFriend))
	mux.Handle("DELETE /api/friends/{id}", s.authed(s.handleRemoveFriend))

	mux.Handle("GET /api/uploads/signature", s.authed(s.handleUploadSignature))

	mux.Handle("GET /api/analytics", s.authed(s.handleAnalytics))
	mux.Handle("GET /api/analytics/heatmap", s.authed(s.handleHeatmap))

	mux.Handle("GET /api/reading-plans", s.authed(s.handleListPlans))
	mux.Handle("GET /api/reading-plans/{id}", s.authed(s.handlePlanDetail))
	mux.Handle("POST /api/reading-plans/{id}/enroll", s.authed(s.handleEnrollPlan))
	mux.Handle("DELETE /api/reading-plans/{id}/enroll", s.authed(s.handleLeavePlan))
	mux.Handle("POST /api/reading-plans/{id}/days/{day}/complete", s.authed(s.handleCompletePlanDay))
	mux.Handle("DELETE /api/reading-plans/{id}/days/{day}/complete", s.authed(s.handleUncompletePlanDay))

	mux.Handle("GET /api/prayers", s.authed(s.handleListPrayers))
	mux.Handle("POST /api/prayers", s.authed(s.handleCreatePrayer))
	mux.Handle("PUT /api/prayers/{id}", s.authed(s.handleUpdatePrayer))
	mux.Handle("POST /api/prayers/{id}/answer", s.authed(s.handleToggleAnswered))
	mux.Handle("DELETE /api/prayers/{id}", s.authed(s.handleDeletePrayer))

	mux.Handle("GET /api/feed", s.authed(s.handleFeed))
	mux.Handle("POST /api/feed", s.authed(s.handleCreatePost))
	mux.Handle("POST /api/feed/{id}/like", s.authed(s.handleToggleLike))
	mux.Handle("POST /api/feed/{id}/comments", s.authed(s.handleCreateComment))

	mux.Handle("POST /api/streak/freeze", s.authed(s.handleBuyFreeze))

	mux.Handle("GET /api/bible", s.authed(s.handleBible))
	mux.Handle("GET /api/bible/books", s.authed(s.handleBibleBooks))
	mux.Handle("GET /api/bible/marks", s.authed(s.handleListMarks))
	mux.Handle("POST /api/bible/marks", s.authed(s.handleToggleMark))
	mux.Handle("GET /api/bible/bookmarks", s.authed(s.handleListBookmarks))

	mux.Handle("GET /api/devotional/today", s.authed(s.handleTodayDevotional))
	mux.Handle("POST /api/devotional/{id}/complete", s.authed(s.handleCompleteDevotional))

	mux.Handle("GET /api/journal", s.authed(s.handleListJournal))
	mux.Handle("GET /api/journal/{date}", s.authed(s.handleGetJournal))
	mux.Handle("PUT /api/journal/{date}", s.authed(s.handleUpsertJournal))
	mux.Handle("DELETE /api/journal/{date}", s.authed(s.handleDeleteJournal))

	return s.cors(logging(mux))
}

// ----- middleware ----------------------------------------------------------

type ctxKey string

const userIDKey ctxKey = "userID"

// authed wraps a handler that requires a valid bearer token.
func (s *Server) authed(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		const prefix = "Bearer "
		if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
			writeErr(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		uid, err := s.auth.Verify(header[len(prefix):])
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func userID(r *http.Request) string {
	v, _ := r.Context().Value(userIDKey).(string)
	return v
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", s.cfg.CORSOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// ----- response helpers ----------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decode(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
