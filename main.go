package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const tokenTTL = time.Hour

type server struct {
	store *streamStore
}

type registerRequest struct {
	URL string `json:"url"`
}

type registerResponse struct {
	RedirectURL string `json:"redirect_url"`
	Token       string `json:"token"`
	ExpiresAt   string `json:"expires_at"`
}

func main() {
	srv := &server{store: newStreamStore(tokenTTL)}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.handleHealth)
	mux.HandleFunc("/api/streams", srv.handleRegisterStream)
	mux.HandleFunc("/r/", srv.handleRedirectByToken)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	log.Printf("rtsp-redirect listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func (s *server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *server) handleRegisterStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	rtspURL, err := validateRTSPURL(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, expiresAt, err := s.store.create(rtspURL)
	if err != nil {
		http.Error(w, "failed to create token", http.StatusInternalServerError)
		return
	}

	redirectURL := publicURL(r, "/r/"+token)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(registerResponse{
		RedirectURL: redirectURL,
		Token:       token,
		ExpiresAt:   expiresAt.UTC().Format(time.RFC3339),
	})
}

func (s *server) handleRedirectByToken(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, "/r/")
	token = strings.Trim(token, "/")
	if token == "" || strings.Contains(token, "/") {
		http.NotFound(w, r)
		return
	}

	rtspURL, ok := s.store.get(token)
	if !ok {
		http.Error(w, "link not found or expired", http.StatusNotFound)
		return
	}

	w.Header().Set("Location", rtspURL)
	w.WriteHeader(http.StatusMovedPermanently)
}

func publicURL(r *http.Request, path string) string {
	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}
	return requestScheme(r) + "://" + host + path
}

func requestScheme(r *http.Request) string {
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func validateRTSPURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errBadRequest("url is required")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", errBadRequest("invalid url")
	}
	if !isRTSPScheme(u.Scheme) {
		return "", errBadRequest("only rtsp and rtsps schemes are allowed")
	}
	if u.Host == "" {
		return "", errBadRequest("url must contain host")
	}
	return u.String(), nil
}

func isRTSPScheme(scheme string) bool {
	switch strings.ToLower(scheme) {
	case "rtsp", "rtsps":
		return true
	default:
		return false
	}
}

type badRequestError string

func (e badRequestError) Error() string { return string(e) }

func errBadRequest(msg string) error { return badRequestError(msg) }
