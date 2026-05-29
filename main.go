package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func main() {
	addr := ":" + envOr("PORT", "8080")
	baseURL := strings.TrimRight(os.Getenv("BASE_URL"), "/")

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		redirect(w, r, baseURL, false)
	})
	mux.HandleFunc("/link", func(w http.ResponseWriter, r *http.Request) {
		redirect(w, r, baseURL, true)
	})

	log.Printf("rtsp-redirect listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func redirect(w http.ResponseWriter, r *http.Request, baseURL string, jsonOnly bool) {
	rtspURL, err := rtspURLFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	redirectURL := buildRedirectURL(baseURL, r, rtspURL)

	if jsonOnly || r.Header.Get("Accept") == "application/json" {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"redirect_url": redirectURL,
			"location":     rtspURL,
		})
		return
	}

	w.Header().Set("Location", rtspURL)
	w.WriteHeader(http.StatusMovedPermanently)
}

func rtspURLFromRequest(r *http.Request) (string, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("url"))
	if raw == "" && r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			return "", err
		}
		raw = strings.TrimSpace(r.FormValue("url"))
	}
	if raw == "" {
		return "", errBadRequest("query parameter url is required")
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

func buildRedirectURL(baseURL string, r *http.Request, rtspURL string) string {
	q := url.Values{"url": {rtspURL}}
	path := "/redirect?" + q.Encode()

	if baseURL != "" {
		return baseURL + path
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = "localhost:" + envOr("PORT", "8080")
	}
	return scheme + "://" + host + path
}

func isRTSPScheme(scheme string) bool {
	switch strings.ToLower(scheme) {
	case "rtsp", "rtsps":
		return true
	default:
		return false
	}
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

type badRequestError string

func (e badRequestError) Error() string { return string(e) }

func errBadRequest(msg string) error { return badRequestError(msg) }
