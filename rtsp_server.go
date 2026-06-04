package main

import (
	"log"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
)

type rtspHandler struct {
	store *streamStore
}

func startRTSPServer(listen string, store *streamStore) error {
	h := &rtspHandler{store: store}
	srv := &gortsplib.Server{
		Handler:     h,
		RTSPAddress: listen,
	}
	log.Printf("rtsp redirect listening on %s", listen)
	return srv.StartAndWait()
}

func (h *rtspHandler) OnDescribe(ctx *gortsplib.ServerHandlerOnDescribeCtx) (*base.Response, *gortsplib.ServerStream, error) {
	cameraID := parseCameraIDFromRTSPPath(ctx.Path)
	if cameraID == "" || isReservedPath(cameraID) {
		return &base.Response{StatusCode: base.StatusNotFound}, nil, nil
	}

	target, ok := h.store.get(cameraID)
	if !ok {
		return &base.Response{StatusCode: base.StatusNotFound}, nil, nil
	}

	return rtspRedirectResponse(target), nil, nil
}

func (h *rtspHandler) OnSetup(ctx *gortsplib.ServerHandlerOnSetupCtx) (*base.Response, *gortsplib.ServerStream, error) {
	cameraID := parseCameraIDFromRTSPPath(ctx.Path)
	if cameraID == "" {
		return &base.Response{StatusCode: base.StatusNotFound}, nil, nil
	}

	target, ok := h.store.get(cameraID)
	if !ok {
		return &base.Response{StatusCode: base.StatusNotFound}, nil, nil
	}

	return rtspRedirectResponse(target), nil, nil
}

func rtspRedirectResponse(targetRTSP string) *base.Response {
	return &base.Response{
		StatusCode: base.StatusFound,
		Header: base.Header{
			"Location": base.HeaderValue{targetRTSP},
		},
	}
}
