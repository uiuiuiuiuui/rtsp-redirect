package main

import (
	"os"
	"strings"
)

type appConfig struct {
	httpPort        string
	rtspListen      string
	rtspPublicHost  string
	rtspPublicPort  string
}

func loadConfig() appConfig {
	httpPort := strings.TrimSpace(os.Getenv("PORT"))
	if httpPort == "" {
		httpPort = "8080"
	}

	rtspPort := strings.TrimSpace(os.Getenv("RTSP_PORT"))
	if rtspPort == "" {
		rtspPort = "8554"
	}

	rtspListen := strings.TrimSpace(os.Getenv("RTSP_LISTEN"))
	if rtspListen == "" {
		rtspListen = ":" + rtspPort
	}

	publicHost := strings.TrimSpace(os.Getenv("RTSP_PUBLIC_HOST"))
	if publicHost == "" {
		publicHost = "127.0.0.1"
	}

	publicPort := strings.TrimSpace(os.Getenv("RTSP_PUBLIC_PORT"))
	if publicPort == "" {
		publicPort = rtspPort
	}

	return appConfig{
		httpPort:       httpPort,
		rtspListen:     rtspListen,
		rtspPublicHost: publicHost,
		rtspPublicPort: publicPort,
	}
}

func (c appConfig) publicRTSPURL(cameraID string) string {
	path := publicRTSPPath(cameraID)
	return "rtsp://" + c.rtspPublicHost + ":" + c.rtspPublicPort + path
}
