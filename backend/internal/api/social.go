package api

import (
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/royparsaoran/daily-grind/backend/internal/models"
)

func (s *Server) handleFriends(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	// Weekly EXP is summed from quest completions in the last 7 days. Includes
	// the current user so the client can render the leaderboard with "kamu".
	rows, err := s.pool.Query(r.Context(), `
		WITH circle AS (
			SELECT friend_id AS id FROM friendships WHERE user_id=$1
			UNION SELECT $1
		)
		SELECT u.id, u.name, COALESCE(u.avatar_url,''), u.level, u.streak,
		       COALESCE((SELECT SUM(exp_awarded) FROM quest_completions qc
		                 WHERE qc.user_id=u.id AND qc.completed_on >= current_date - 6), 0) AS weekly,
		       (u.id=$1) AS is_me
		FROM users u JOIN circle ON circle.id=u.id
		ORDER BY weekly DESC, u.level DESC`, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query friends")
		return
	}
	defer rows.Close()

	out := []models.Friend{}
	for rows.Next() {
		var f models.Friend
		if err := rows.Scan(&f.ID, &f.Name, &f.Avatar, &f.Level, &f.Streak, &f.WeeklyEXP, &f.IsMe); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan friend")
			return
		}
		out = append(out, f)
	}
	writeJSON(w, http.StatusOK, out)
}

// handleSearchUsers finds users by name to add as friends (excludes self).
func (s *Server) handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusOK, []models.UserSearchResult{})
		return
	}
	rows, err := s.pool.Query(r.Context(), `
		SELECT u.id, u.name, COALESCE(u.avatar_url,''), u.level, u.title,
		       CASE
		           WHEN EXISTS (SELECT 1 FROM friendships f WHERE f.user_id=$1 AND f.friend_id=u.id) THEN 'friend'
		           WHEN EXISTS (SELECT 1 FROM friend_requests r WHERE r.from_id=$1 AND r.to_id=u.id) THEN 'outgoing'
		           WHEN EXISTS (SELECT 1 FROM friend_requests r WHERE r.from_id=u.id AND r.to_id=$1) THEN 'incoming'
		           ELSE 'none' END AS status
		FROM users u
		WHERE u.id <> $1 AND u.name ILIKE '%' || $2 || '%'
		ORDER BY u.name
		LIMIT 20`, uid, q)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query users")
		return
	}
	defer rows.Close()

	out := []models.UserSearchResult{}
	for rows.Next() {
		var u models.UserSearchResult
		if err := rows.Scan(&u.ID, &u.Name, &u.Avatar, &u.Level, &u.Title, &u.Status); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan user")
			return
		}
		out = append(out, u)
	}
	writeJSON(w, http.StatusOK, out)
}

// handleListRequests returns the incoming friend requests for the current user.
func (s *Server) handleListRequests(w http.ResponseWriter, r *http.Request) {
	rows, err := s.pool.Query(r.Context(), `
		SELECT u.id, u.name, COALESCE(u.avatar_url,''), u.level, u.title
		FROM friend_requests fr JOIN users u ON u.id=fr.from_id
		WHERE fr.to_id=$1
		ORDER BY fr.created_at DESC`, userID(r))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query requests")
		return
	}
	defer rows.Close()
	out := []models.FriendRequest{}
	for rows.Next() {
		var fr models.FriendRequest
		if err := rows.Scan(&fr.ID, &fr.Name, &fr.Avatar, &fr.Level, &fr.Title); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan request")
			return
		}
		out = append(out, fr)
	}
	writeJSON(w, http.StatusOK, out)
}

// handleAddFriend sends a friend request. If the target already requested the
// current user, the request is auto-accepted (mutual). If already friends, no-op.
func (s *Server) handleAddFriend(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	friendID := r.PathValue("id")
	if friendID == uid {
		writeErr(w, http.StatusBadRequest, "tidak bisa menambahkan diri sendiri")
		return
	}
	var exists bool
	if err := s.pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, friendID).Scan(&exists); err != nil || !exists {
		writeErr(w, http.StatusNotFound, "pengguna tidak ditemukan")
		return
	}

	// Already friends?
	var already bool
	_ = s.pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM friendships WHERE user_id=$1 AND friend_id=$2)`, uid, friendID).Scan(&already)
	if already {
		writeJSON(w, http.StatusOK, map[string]string{"status": "friend"})
		return
	}

	// Reverse request pending? -> auto-accept.
	var reverse bool
	_ = s.pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM friend_requests WHERE from_id=$1 AND to_id=$2)`, friendID, uid).Scan(&reverse)
	if reverse {
		if err := s.acceptRequest(r, friendID); err != nil {
			writeErr(w, http.StatusInternalServerError, "gagal menerima permintaan")
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "friend"})
		return
	}

	if _, err := s.pool.Exec(r.Context(),
		`INSERT INTO friend_requests(from_id,to_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, uid, friendID); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal mengirim permintaan")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "outgoing"})
}

// acceptRequest turns a pending request (fromID -> current user) into a friendship.
func (s *Server) acceptRequest(r *http.Request, fromID string) error {
	uid := userID(r)
	return pgx.BeginFunc(r.Context(), s.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(r.Context(),
			`INSERT INTO friendships(user_id,friend_id) VALUES ($1,$2),($2,$1) ON CONFLICT DO NOTHING`,
			uid, fromID); err != nil {
			return err
		}
		_, err := tx.Exec(r.Context(),
			`DELETE FROM friend_requests WHERE (from_id=$1 AND to_id=$2) OR (from_id=$2 AND to_id=$1)`,
			fromID, uid)
		return err
	})
}

// handleAcceptRequest accepts an incoming request (path id = requester's id).
func (s *Server) handleAcceptRequest(w http.ResponseWriter, r *http.Request) {
	fromID := r.PathValue("id")
	var ok bool
	if err := s.pool.QueryRow(r.Context(),
		`SELECT EXISTS(SELECT 1 FROM friend_requests WHERE from_id=$1 AND to_id=$2)`, fromID, userID(r)).Scan(&ok); err != nil || !ok {
		writeErr(w, http.StatusNotFound, "permintaan tidak ditemukan")
		return
	}
	if err := s.acceptRequest(r, fromID); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menerima permintaan")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleRejectRequest declines an incoming request.
func (s *Server) handleRejectRequest(w http.ResponseWriter, r *http.Request) {
	if _, err := s.pool.Exec(r.Context(),
		`DELETE FROM friend_requests WHERE from_id=$1 AND to_id=$2`, r.PathValue("id"), userID(r)); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menolak permintaan")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleRemoveFriend removes a friendship in both directions.
func (s *Server) handleRemoveFriend(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	friendID := r.PathValue("id")
	if _, err := s.pool.Exec(r.Context(),
		`DELETE FROM friendships WHERE (user_id=$1 AND friend_id=$2) OR (user_id=$2 AND friend_id=$1)`,
		uid, friendID); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menghapus teman")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleFeed(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	rows, err := s.pool.Query(r.Context(), `
		SELECT p.id, p.user_id, u.name, COALESCE(u.avatar_url,''), u.level, p.body,
		       COALESCE(p.photo_url,''), COALESCE(p.badge,''), p.created_at,
		       (SELECT count(*) FROM post_likes pl WHERE pl.post_id=p.id) AS likes,
		       EXISTS (SELECT 1 FROM post_likes pl WHERE pl.post_id=p.id AND pl.user_id=$1) AS liked
		FROM posts p JOIN users u ON u.id=p.user_id
		ORDER BY p.created_at DESC
		LIMIT 50`, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "query feed")
		return
	}
	defer rows.Close()

	posts := []models.Post{}
	index := map[string]int{}
	for rows.Next() {
		var p models.Post
		if err := rows.Scan(&p.ID, &p.UserID, &p.Author, &p.AuthorAvatar, &p.AuthorLevel, &p.Body,
			&p.PhotoURL, &p.Badge, &p.CreatedAt, &p.Likes, &p.LikedByMe); err != nil {
			writeErr(w, http.StatusInternalServerError, "scan post")
			return
		}
		p.Comments = []models.Comment{}
		index[p.ID] = len(posts)
		posts = append(posts, p)
	}
	rows.Close()

	if len(posts) > 0 {
		crows, err := s.pool.Query(r.Context(), `
			SELECT c.id, c.post_id, c.user_id, u.name, COALESCE(u.avatar_url,''), c.body, c.created_at
			FROM comments c JOIN users u ON u.id=c.user_id
			WHERE c.post_id = ANY($1)
			ORDER BY c.created_at ASC`, keys(index))
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "query comments")
			return
		}
		defer crows.Close()
		for crows.Next() {
			var postID string
			var c models.Comment
			if err := crows.Scan(&c.ID, &postID, &c.UserID, &c.Author, &c.AuthorAvatar, &c.Body, &c.CreatedAt); err != nil {
				writeErr(w, http.StatusInternalServerError, "scan comment")
				return
			}
			if i, ok := index[postID]; ok {
				posts[i].Comments = append(posts[i].Comments, c)
			}
		}
	}

	writeJSON(w, http.StatusOK, posts)
}

func (s *Server) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Body     string `json:"body"`
		PhotoURL string `json:"photo_url"`
		Badge    string `json:"badge"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Body = strings.TrimSpace(body.Body)
	if body.Body == "" {
		writeErr(w, http.StatusBadRequest, "isi postingan tidak boleh kosong")
		return
	}
	var id string
	err := s.pool.QueryRow(r.Context(), `
		INSERT INTO posts(user_id,body,photo_url,badge)
		VALUES ($1,$2,NULLIF($3,''),NULLIF($4,'')) RETURNING id`,
		userID(r), body.Body, strings.TrimSpace(body.PhotoURL), strings.TrimSpace(body.Badge)).Scan(&id)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal membuat postingan")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) handleToggleLike(w http.ResponseWriter, r *http.Request) {
	uid := userID(r)
	postID := r.PathValue("id")

	tag, err := s.pool.Exec(r.Context(),
		`DELETE FROM post_likes WHERE post_id=$1 AND user_id=$2`, postID, uid)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal memproses like")
		return
	}
	liked := false
	if tag.RowsAffected() == 0 {
		if _, err := s.pool.Exec(r.Context(),
			`INSERT INTO post_likes(post_id,user_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, postID, uid); err != nil {
			writeErr(w, http.StatusInternalServerError, "gagal menambah like")
			return
		}
		liked = true
	}

	var count int
	_ = s.pool.QueryRow(r.Context(),
		`SELECT count(*) FROM post_likes WHERE post_id=$1`, postID).Scan(&count)
	writeJSON(w, http.StatusOK, map[string]any{"liked": liked, "likes": count})
}

func (s *Server) handleCreateComment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Body string `json:"body"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	body.Body = strings.TrimSpace(body.Body)
	if body.Body == "" {
		writeErr(w, http.StatusBadRequest, "balasan tidak boleh kosong")
		return
	}
	var c models.Comment
	err := s.pool.QueryRow(r.Context(), `
		WITH ins AS (
			INSERT INTO comments(post_id,user_id,body) VALUES ($1,$2,$3)
			RETURNING id,user_id,body,created_at
		)
		SELECT ins.id, ins.user_id, u.name, COALESCE(u.avatar_url,''), ins.body, ins.created_at
		FROM ins JOIN users u ON u.id=ins.user_id`,
		r.PathValue("id"), userID(r), body.Body).
		Scan(&c.ID, &c.UserID, &c.Author, &c.AuthorAvatar, &c.Body, &c.CreatedAt)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menambah balasan")
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

func keys(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
