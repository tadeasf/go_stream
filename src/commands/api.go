package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
)

var ApiCmd = &cobra.Command{
	Use:   "api",
	Short: "Start the streaming server with REST API",
	RunE:  apiAction,
}

var directoryPath string

func init() {
	ApiCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Search for videos recursively")
	ApiCmd.Flags().IntVarP(&port, "port", "p", 8069, "Port to serve on")
	ApiCmd.Flags().BoolVar(&useAuth, "auth", false, "Enable basic authentication")
	ApiCmd.Flags().StringVar(&sortBy, "sort", "", "Sort videos by: name or size")
	ApiCmd.Flags().StringVarP(&directoryPath, "path", "d", "", "Directory path to search for videos")
	ApiCmd.MarkFlagRequired("path")
}

type VideoInfo struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type PathRequest struct {
	Path string `json:"path"`
	Args string `json:"args"`
}

type PathSuggestion struct {
	Path  string `json:"path"`
	IsDir bool   `json:"isDir"`
}

func apiAction(cmd *cobra.Command, args []string) error {
	recursive, _ := cmd.Flags().GetBool("recursive")
	port, _ := cmd.Flags().GetInt("port")
	useAuth, _ := cmd.Flags().GetBool("auth")
	sortBy, _ := cmd.Flags().GetString("sort")

	if directoryPath == "" {
		return fmt.Errorf("%s", yellow("directory path cannot be empty"))
	}

	videos := FindVideosConcurrent(directoryPath, recursive)

	// Sort videos based on the sortBy flag
	sortVideos(videos, sortBy)

	ip := GetOutboundIP()
	router := mux.NewRouter()

	if useAuth {
		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load authentication config: %s", err)
		}
		router.Use(func(next http.Handler) http.Handler {
			return BasicAuth(next, config.Username, config.Password)
		})
	}

	videoInfos := make([]VideoInfo, len(videos))
	for i, v := range videos {
		videoInfos[i] = VideoInfo{
			ID:   strconv.Itoa(i + 1),
			Path: v.Path,
			Size: v.Size,
		}
	}

	router.HandleFunc("/api/v1/playlist", func(w http.ResponseWriter, r *http.Request) {
		playlist := generatePlaylist(videos, ip, port, useAuth, "", "")
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Write([]byte(playlist))
	}).Methods("GET")

	router.HandleFunc("/api/v1/playlist", func(w http.ResponseWriter, r *http.Request) {
		var req PathRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Restart the server with the new path
		executable, _ := os.Executable()
		cmd := exec.Command(executable, "api", "--path", req.Path)
		if strings.Contains(req.Args, "-r") {
			cmd.Args = append(cmd.Args, "-r")
		}
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Start()
		if err != nil {
			http.Error(w, "Failed to restart server", http.StatusInternalServerError)
			return
		}

		// Exit the current process
		go func() {
			time.Sleep(1 * time.Second)
			os.Exit(0)
		}()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Server restarting with new path"})
	}).Methods("POST")

	router.HandleFunc("/api/v1/playlist/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(videoInfos)
	}).Methods("GET")

	router.HandleFunc("/api/v1/playlist/{video_id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			vars := mux.Vars(r)
			id, err := strconv.Atoi(vars["video_id"])
			if err != nil || id < 1 || id > len(videos) {
				http.Error(w, "Invalid video ID", http.StatusBadRequest)
				return
			}

			// Remove the video from the slice
			videos = append(videos[:id-1], videos[id:]...)

			// Update videoInfos
			videoInfos = make([]VideoInfo, len(videos))
			for i, v := range videos {
				videoInfos[i] = VideoInfo{
					ID:   strconv.Itoa(i + 1),
					Path: v.Path,
					Size: v.Size,
				}
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"message": "Video deleted successfully"})
			return
		}

		vars := mux.Vars(r)
		id, err := strconv.Atoi(vars["video_id"])
		if err != nil || id < 1 || id > len(videos) {
			http.Error(w, "Invalid video ID", http.StatusBadRequest)
			return
		}
		video := videos[id-1]
		streamURL := fmt.Sprintf("http://%s:%d/videos/%s", ip, port, video.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"url": streamURL})
	}).Methods("GET", "DELETE")

	router.PathPrefix("/videos/").Handler(http.StripPrefix("/videos/", http.FileServer(http.Dir(directoryPath))))

	// Create a new CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS", "DELETE"},
		AllowedHeaders: []string{"*"},
	})

	// Wrap the router with the CORS handler
	handler := c.Handler(router)

	// Add path suggestions endpoint
	router.HandleFunc("/api/v1/path-suggestions", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Query().Get("path")
		suggestions, err := getPathSuggestions(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(suggestions)
	}).Methods("GET")

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Printf(green("Starting API server at http://%s:%d\n"), ip, port)
	fmt.Println(yellow("Available endpoints:"))
	fmt.Printf("  GET /api/v1/playlist\n")
	fmt.Printf("  POST /api/v1/playlist\n")
	fmt.Printf("  GET /api/v1/playlist/list\n")
	fmt.Printf("  GET /api/v1/playlist/{video_id}\n")
	fmt.Printf("  GET /api/v1/path-suggestions\n")

	return http.ListenAndServe(addr, handler)
}

func getPathSuggestions(basePath string) ([]PathSuggestion, error) {
	dir := filepath.Dir(basePath)
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var suggestions []PathSuggestion
	for _, file := range files {
		// Skip hidden directories (those starting with a dot)
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		// Only include directories
		if !file.IsDir() {
			continue
		}

		fullPath := filepath.Join(dir, file.Name())
		if strings.HasPrefix(fullPath, basePath) {
			suggestions = append(suggestions, PathSuggestion{
				Path:  fullPath,
				IsDir: true, // This will always be true now
			})
		}
	}
	return suggestions, nil
}

func sortVideos(videos []Video, sortBy string) {
	switch sortBy {
	case "name":
		sort.Slice(videos, func(i, j int) bool {
			return videos[i].Path > videos[j].Path // Sort descending
		})
	case "size":
		sort.Slice(videos, func(i, j int) bool {
			return videos[i].Size > videos[j].Size // Sort descending
		})
	}
}
