package api

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// uploadFolders maps an upload "kind" to its Cloudinary folder (allow-list).
var uploadFolders = map[string]string{
	"avatar": "dailygrind/avatars",
	"post":   "dailygrind/posts",
}

// signUpload returns signed, direct-to-Cloudinary upload parameters for a folder.
// The API secret never leaves the server.
func (s *Server) signUpload(w http.ResponseWriter, folder string) {
	if s.cfg.CloudinaryCloudName == "" || s.cfg.CloudinaryAPISecret == "" {
		writeErr(w, http.StatusServiceUnavailable, "upload belum dikonfigurasi")
		return
	}
	ts := time.Now().Unix()
	// Signature = SHA1( "<params sorted, joined by &>" + api_secret ).
	toSign := fmt.Sprintf("folder=%s&timestamp=%d", folder, ts)
	sum := sha1.Sum([]byte(toSign + s.cfg.CloudinaryAPISecret))
	writeJSON(w, http.StatusOK, map[string]any{
		"cloud_name": s.cfg.CloudinaryCloudName,
		"api_key":    s.cfg.CloudinaryAPIKey,
		"timestamp":  ts,
		"folder":     folder,
		"signature":  hex.EncodeToString(sum[:]),
	})
}

// handleAvatarSignature signs an avatar upload.
func (s *Server) handleAvatarSignature(w http.ResponseWriter, r *http.Request) {
	s.signUpload(w, uploadFolders["avatar"])
}

// handleUploadSignature signs a generic upload for a given ?kind= (post, avatar).
func (s *Server) handleUploadSignature(w http.ResponseWriter, r *http.Request) {
	folder, ok := uploadFolders[r.URL.Query().Get("kind")]
	if !ok {
		writeErr(w, http.StatusBadRequest, "jenis upload tidak valid")
		return
	}
	s.signUpload(w, folder)
}

// handleSetAvatar stores the uploaded avatar's secure URL on the user.
func (s *Server) handleSetAvatar(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := decode(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid request body")
		return
	}
	url := strings.TrimSpace(body.URL)
	// Accept only https Cloudinary URLs to avoid storing arbitrary links.
	if !strings.HasPrefix(url, "https://res.cloudinary.com/") {
		writeErr(w, http.StatusBadRequest, "url tidak valid")
		return
	}
	if _, err := s.pool.Exec(r.Context(),
		`UPDATE users SET avatar_url=$1 WHERE id=$2`, url, userID(r)); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menyimpan avatar")
		return
	}
	u, _ := s.loadUser(r.Context(), s.pool, userID(r))
	writeJSON(w, http.StatusOK, u)
}

// handleRemoveAvatar clears the avatar (revert to initials).
func (s *Server) handleRemoveAvatar(w http.ResponseWriter, r *http.Request) {
	if _, err := s.pool.Exec(r.Context(),
		`UPDATE users SET avatar_url=NULL WHERE id=$1`, userID(r)); err != nil {
		writeErr(w, http.StatusInternalServerError, "gagal menghapus avatar")
		return
	}
	u, _ := s.loadUser(r.Context(), s.pool, userID(r))
	writeJSON(w, http.StatusOK, u)
}
