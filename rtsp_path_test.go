package main

import "testing"

func TestParseCameraIDFromRTSPPath(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"/camera/key/59584", "59584"},
		{"/camera/key/1", "1"},
		{"/59584", "59584"},
		{"/key", "key"},
		{"", ""},
		{"/camera/key", ""},
		{"/camera/key/a/b", ""},
	}
	for _, tc := range tests {
		if got := parseCameraIDFromRTSPPath(tc.path); got != tc.want {
			t.Errorf("parseCameraIDFromRTSPPath(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

func TestPublicRTSPPath(t *testing.T) {
	if got := publicRTSPPath("123"); got != "/camera/key/123" {
		t.Fatalf("publicRTSPPath = %q", got)
	}
}
