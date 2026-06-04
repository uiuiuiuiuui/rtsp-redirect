package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type server struct {
	store  *streamStore
	config appConfig
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
	cfg := loadConfig()
	srv := &server{
		store:  newStreamStore(),
		config: cfg,
	}

	go func() {
		if err := startRTSPServer(cfg.rtspListen, srv.store); err != nil {
			log.Fatalf("rtsp server: %v", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", srv.handleHealth)
	mux.HandleFunc("/api/streams", srv.handleRegisterStream)

	addr := ":" + cfg.httpPort
	log.Printf("http api listening on %s", addr)
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
		RedirectURL: s.config.publicRTSPURL(cameraID),
		URL:         rtspURL,
	})
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
