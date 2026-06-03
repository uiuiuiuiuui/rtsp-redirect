package main

import (
	"fmt"
	"strings"
)

func parseCameraIDFromPath(path string) string {
	id := strings.Trim(path, "/")
	if id == "" || strings.Contains(id, "/") {
		return ""
	}
	id = strings.TrimSuffix(id, ".m3u")
	return id
}

func buildM3UPlaylist(cameraID, rtspURL string) string {
	title := "Camera " + cameraID
	return fmt.Sprintf("#EXTM3U\n#EXTVLCOPT:network-caching=1000\n#EXTINF:-1,%s\n%s\n", title, rtspURL)
}
