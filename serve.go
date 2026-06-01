package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"blog-cms-v2/src/definitions"
	"blog-cms-v2/src/libs"
)

func runHTTPServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /dev/posts", httpGetAllPosts)
	mux.HandleFunc("GET /dev/post/{slug}", httpGetPostBySlug)
	mux.HandleFunc("GET /dev/post/{slug}/comments", httpGetComments)
	mux.HandleFunc("POST /dev/post/{slug}/comments", httpPostComment)
	mux.HandleFunc("POST /dev/post/{slug}/comments/{commentId}/recommend", httpRecommendComment)

	addr := ":" + port
	log.Printf("Local API listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, cors(mux)); err != nil {
		log.Fatal(err)
	}
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var commentRepo libs.CommentRepository

func getCommentRepo(r *http.Request) (libs.CommentRepository, error) {
	if commentRepo == nil {
		var err error
		commentRepo, err = libs.NewCommentRepository(r.Context())
		if err != nil {
			return nil, err
		}
	}
	return commentRepo, nil
}

func httpGetAllPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := libs.NewPostService().GetAllPosts(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errMap(err))
		return
	}
	writeJSON(w, http.StatusOK, posts)
}

func httpGetPostBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		writeJSON(w, http.StatusBadRequest, errMap(fmt.Errorf("slug is not defined")))
		return
	}
	post, err := libs.NewPostService().GetPostBySlug(r.Context(), slug)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errMap(err))
		return
	}
	if post == nil {
		writeJSON(w, http.StatusNotFound, errMap(fmt.Errorf("post not found: %s", slug)))
		return
	}
	writeJSON(w, http.StatusOK, post)
}

func httpGetComments(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	repo, err := getCommentRepo(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errMap(err))
		return
	}
	comments, err := repo.GetComments(r.Context(), slug)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errMap(err))
		return
	}
	if comments == nil {
		comments = []definitions.Comment{}
	}
	writeJSON(w, http.StatusOK, comments)
}

func httpPostComment(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	repo, err := getCommentRepo(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errMap(err))
		return
	}

	var body struct {
		Author   string `json:"author"`
		Text     string `json:"text"`
		ParentID string `json:"parentId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errMap(fmt.Errorf("invalid body")))
		return
	}

	body.Author = strings.TrimSpace(body.Author)
	body.Text = strings.TrimSpace(body.Text)
	if body.Author == "" || body.Text == "" {
		writeJSON(w, http.StatusBadRequest, errMap(fmt.Errorf("author and text required")))
		return
	}

	c := definitions.Comment{
		Slug:      slug,
		CommentID: libs.NewCommentID(),
		Author:    body.Author,
		Text:      body.Text,
		CreatedAt: time.Now().UTC(),
		Active:    false, // local dev: auto-approve for testing
		ParentID:  body.ParentID,
	}
	// In local dev, auto-activate so you can see comments immediately
	c.Active = true

	if err := repo.PutComment(r.Context(), c); err != nil {
		writeJSON(w, http.StatusInternalServerError, errMap(err))
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "pending"})
}

func httpRecommendComment(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	commentID := r.PathValue("commentId")
	repo, err := getCommentRepo(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errMap(err))
		return
	}
	if err := repo.IncrementRecommend(r.Context(), slug, commentID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errMap(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func errMap(err error) map[string]string {
	return map[string]string{"error": err.Error()}
}
