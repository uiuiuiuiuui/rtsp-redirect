package main

import "strings"

const rtspPathPrefix = "camera/key/"

func publicRTSPPath(cameraID string) string {
	return "/" + rtspPathPrefix + cameraID
}

func parseCameraIDFromRTSPPath(path string) string {
	path = strings.Trim(path, "/")
	if path == "" {
		return ""
	}

	if strings.HasPrefix(path, rtspPathPrefix) {
		id := strings.TrimPrefix(path, rtspPathPrefix)
		if id != "" && !strings.Contains(id, "/") {
			return id
		}
		return ""
	}

	if !strings.Contains(path, "/") {
		return path
	}
	return ""
}
