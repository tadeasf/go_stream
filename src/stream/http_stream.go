package stream

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/grafov/m3u8"
	"github.com/urfave/cli/v2"
)

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
)

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
		log.Printf("Error finding videos: %v", err)
	}

	return videos
}

func generateMasterPlaylist(videos []string, ip net.IP, port int) *m3u8.MasterPlaylist {
	masterPl := m3u8.NewMasterPlaylist()
	for _, video := range videos {
		videoURL := fmt.Sprintf("http://%s:%d/videos/%s", ip, port, url.PathEscape(video))
		masterPl.Append(videoURL, nil, m3u8.VariantParams{})
	}
	return masterPl
}

func serveVideoFile(w http.ResponseWriter, r *http.Request, videoPath string) {
	log.Printf("Serving video file: %s", videoPath)
	http.ServeFile(w, r, videoPath)
}

func serveHLSPlaylist(w http.ResponseWriter, r *http.Request, videoPath string) {
	log.Printf("Attempting to serve HLS playlist for: %s", videoPath)

	// Remove the .m3u8 extension if it exists
	videoPath = strings.TrimSuffix(videoPath, ".m3u8")

	// Check if the actual video file exists
	if _, err := os.Stat(videoPath); os.IsNotExist(err) {
		log.Printf("Video file not found: %s", videoPath)
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	// Generate a simple HLS playlist
	playlist := "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:10\n#EXT-X-MEDIA-SEQUENCE:0\n"
	playlist += fmt.Sprintf("#EXTINF:10.0,\n%s\n", filepath.Base(videoPath))
	playlist += "#EXT-X-ENDLIST\n"

	log.Printf("Generated HLS playlist:\n%s", playlist)

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Write([]byte(playlist))
}

func serveAction(c *cli.Context) error {
	recursive := c.Bool("recursive")
	port := c.Int("port")
	dir := c.String("dir")

	if dir == "" {
		fmt.Print(cyan("Enter the directory path: "))
		dir = prompt.Input("", pathCompleter,
			prompt.OptionPrefix(""),
			prompt.OptionPrefixTextColor(prompt.Yellow),
			prompt.OptionSuggestionBGColor(prompt.DarkGray),
			prompt.OptionSuggestionTextColor(prompt.White),
			prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
			prompt.OptionSelectedSuggestionTextColor(prompt.Black),
		)
		dir = strings.TrimSpace(dir)
	}

	if dir == "" {
		return fmt.Errorf(yellow("directory path cannot be empty"))
	}

	videos := findVideos(dir, recursive)
	ip := getOutboundIP()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)
	})

	http.HandleFunc("/playlist.m3u8", func(w http.ResponseWriter, r *http.Request) {
		masterPlaylist := generateMasterPlaylist(videos, ip, port)
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		playlistContent := masterPlaylist.String()
		log.Printf("Serving master playlist:\n%s", playlistContent)
		_, err := w.Write([]byte(playlistContent))
		if err != nil {
			log.Printf("Error writing master playlist: %v", err)
		}
	})

	http.HandleFunc("/videos/", func(w http.ResponseWriter, r *http.Request) {
		videoPath := filepath.Join(dir, strings.TrimPrefix(r.URL.Path, "/videos/"))
		decodedPath, _ := url.PathUnescape(videoPath)
		log.Printf("Requested video path: %s", decodedPath)

		if strings.HasSuffix(r.URL.Path, ".m3u8") {
			serveHLSPlaylist(w, r, decodedPath)
		} else {
			serveVideoFile(w, r, decodedPath)
		}
	})

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Printf(green("Serving playlist at http://%s:%d/playlist.m3u8\n"), ip, port)
	return http.ListenAndServe(addr, nil)
}

func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func pathCompleter(d prompt.Document) []prompt.Suggest {
	path := d.Text
	dir := filepath.Dir(path)
	if dir == "." {
		dir = ""
	}

	files, _ := os.ReadDir(dir)
	var suggestions []prompt.Suggest
	for _, file := range files {
		if file.IsDir() {
			suggestions = append(suggestions, prompt.Suggest{Text: filepath.Join(dir, file.Name()) + "/"})
		}
	}
	return prompt.FilterHasPrefix(suggestions, d.Text, true)
}
