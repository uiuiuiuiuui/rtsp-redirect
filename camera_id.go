package main

import (
	"net/url"
	"strings"
)

func resolveCameraID(explicitID, rtspURL string) (string, error) {
	id := strings.TrimSpace(explicitID)
	if id != "" {
		if err := validateCameraID(id); err != nil {
			return "", err
		}
		return id, nil
	}

	id, err := extractCameraIDFromRTSP(rtspURL)
	if err != nil {
		return "", err
	}
	if err := validateCameraID(id); err != nil {
		return "", err
	}
	return id, nil
}

func extractCameraIDFromRTSP(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", errBadRequest("invalid url")
	}

	segment := lastPathSegment(u.Path)
	if segment == "" {
		return "", errBadRequest("url path must contain stream name")
	}

	if idx := strings.LastIndex(segment, "_"); idx >= 0 && idx < len(segment)-1 {
		return segment[idx+1:], nil
	}
	return segment, nil
}

func lastPathSegment(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

func validateCameraID(id string) error {
	if id == "" {
		return errBadRequest("camera id is empty")
	}
	if isReservedPath(id) {
		return errBadRequest("camera id is reserved")
	}
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return errBadRequest("camera id contains invalid characters")
	}
	return nil
}

func isReservedPath(segment string) bool {
	switch strings.ToLower(segment) {
	case "health", "api":
		return true
	default:
		return false
	}
}
