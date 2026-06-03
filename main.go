package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type server struct {
	store *streamStore
}

type registerRequest struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type registerResponse struct {
	ID          string `json:"id"`
	RedirectURL string `json:"redirect_url"`
	URL         string `json:"url"`
}

func main() {
	srv := &server{store: newStreamStore()}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.handleHealth)
	mux.HandleFunc("/api/streams", srv.handleRegisterStream)
	mux.HandleFunc("/", srv.handleRedirectByCameraID)

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

	cameraID, err := resolveCameraID(req.ID, rtspURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.store.upsert(cameraID, rtspURL)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(registerResponse{
		ID:          cameraID,
		RedirectURL: publicURL(r, "/"+cameraID),
		URL:         rtspURL,
	})
}

func (s *server) handleRedirectByCameraID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cameraID := parseCameraIDFromPath(r.URL.Path)
	if cameraID == "" {
		http.NotFound(w, r)
		return
	}
	if isReservedPath(cameraID) {
		http.NotFound(w, r)
		return
	}

	rtspURL, ok := s.store.get(cameraID)
	if !ok {
		http.Error(w, "camera not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "audio/x-mpegurl; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Disposition", "inline; filename=\""+cameraID+".m3u\"")
	_, _ = w.Write([]byte(buildM3UPlaylist(cameraID, rtspURL)))
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
