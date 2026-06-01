package main

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type streamEntry struct {
	URL       string
	ExpiresAt time.Time
}

type streamStore struct {
	mu    sync.RWMutex
	items map[string]streamEntry
	ttl   time.Duration
}

func newStreamStore(ttl time.Duration) *streamStore {
	s := &streamStore{
		items: make(map[string]streamEntry),
		ttl:   ttl,
	}
	go s.cleanupLoop()
	return s
}

func (s *streamStore) create(rtspURL string) (token string, expiresAt time.Time, err error) {
	token, err = newToken()
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt = time.Now().Add(s.ttl)
	s.mu.Lock()
	s.items[token] = streamEntry{URL: rtspURL, ExpiresAt: expiresAt}
	s.mu.Unlock()
	return token, expiresAt, nil
}

func (s *streamStore) get(token string) (string, bool) {
	s.mu.RLock()
	entry, ok := s.items[token]
	s.mu.RUnlock()
	if !ok || time.Now().After(entry.ExpiresAt) {
		return "", false
	}
	return entry.URL, true
}

func (s *streamStore) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for token, entry := range s.items {
			if now.After(entry.ExpiresAt) {
				delete(s.items, token)
			}
		}
		s.mu.Unlock()
	}
}

func newToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
