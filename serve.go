package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"blog-cms-v2/src/libs"
)

func runHTTPServer() {
	port := envOrDefault("PORT", "3000")
	mux := http.NewServeMux()

	mux.HandleFunc("GET /dev/posts", httpGetAllPosts)
	mux.HandleFunc("GET /dev/post/{slug}", httpGetPostBySlug)

	addr := ":" + port
	log.Printf("Local API listening on http://localhost%s", addr)
	log.Printf("  GET  http://localhost%s/dev/posts", addr)
	log.Printf("  GET  http://localhost%s/dev/post/{slug}", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
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

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func errMap(err error) map[string]string {
	return map[string]string{"error": err.Error()}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
