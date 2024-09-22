package stream

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func serveHLSSegment(w http.ResponseWriter, r *http.Request, videoPath string) {
	segmentPath := strings.TrimSuffix(videoPath, filepath.Ext(videoPath))
	segmentNumber, err := strconv.Atoi(filepath.Ext(segmentPath)[1:])
	if err != nil {
		http.Error(w, "Invalid segment number", http.StatusBadRequest)
		return
	}

	tsPath := fmt.Sprintf("%s%d.ts", segmentPath, segmentNumber)
	file, err := os.Open(tsPath)
	if err != nil {
		http.Error(w, "Segment not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "video/MP2T")
	io.Copy(w, file)
}
