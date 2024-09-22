package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	dir       string
	recursive bool
	port      int
)

func init() {
	flag.StringVar(&dir, "dir", "", "Directory to search for videos")
	flag.BoolVar(&recursive, "r", false, "Recursive search")
	flag.IntVar(&port, "p", 8069, "Port to serve on")
	flag.Parse()

	if dir == "" {
		log.Fatal("Please provide a directory using the -dir flag")
	}
}

func main() {
	videos := findVideos(dir, recursive)
	ip := getOutboundIP()
	playlist := generatePlaylist(videos, ip, port)

	http.HandleFunc("/playlist.m3u8", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Write([]byte(playlist))
	})

	http.Handle("/videos/", http.StripPrefix("/videos/", http.FileServer(http.Dir(dir))))

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("Serving playlist at http://%s:%d/playlist.m3u8", ip, port)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func findVideos(root string, recursive bool) []string {
	var videos []string
	videoExtensions := []string{".mp4", ".avi", ".mov", ".mkv", ".webm"}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && !recursive && path != root {
			return filepath.SkipDir
		}
		for _, ext := range videoExtensions {
			if strings.HasSuffix(strings.ToLower(path), ext) {
				relPath, _ := filepath.Rel(root, path)
				videos = append(videos, relPath)
				break
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	return videos
}

func generatePlaylist(videos []string, ip net.IP, port int) string {
	playlist := "#EXTM3U\n"
	for _, video := range videos {
		playlist += fmt.Sprintf("#EXTINF:-1,%s\n", video)
		playlist += fmt.Sprintf("http://%s:%d/videos/%s\n", ip, port, video)
	}
	return playlist
}

// Get preferred outbound IP of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

