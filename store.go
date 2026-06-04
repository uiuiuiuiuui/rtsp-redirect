package main

import "sync"

type streamStore struct {
	mu    sync.RWMutex
	items map[string]string // без TTL: запись живёт до upsert через POST /api/streams
}

func newStreamStore() *streamStore {
	return &streamStore{items: make(map[string]string)}
}

func (s *streamStore) upsert(cameraID, rtspURL string) {
	s.mu.Lock()
	s.items[cameraID] = rtspURL
	s.mu.Unlock()
}

func (s *streamStore) get(cameraID string) (string, bool) {
	s.mu.RLock()
	rtspURL, ok := s.items[cameraID]
	s.mu.RUnlock()
	return rtspURL, ok
}
