package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

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
	ApiCmd.Flags().StringVar(&sortBy, "sort", "", "Sort videos by: name, size, or duration")
	ApiCmd.Flags().StringVarP(&directoryPath, "path", "d", "", "Directory path to search for videos")
	ApiCmd.MarkFlagRequired("path")
}

type VideoInfo struct {
	ID       string  `json:"id"`
	Path     string  `json:"path"`
	Size     int64   `json:"size"`
	Duration float64 `json:"duration"`
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
			ID:       strconv.Itoa(i + 1),
			Path:     v.Path,
			Size:     v.Size,
			Duration: v.Duration.Seconds(),
		}
	}

	router.HandleFunc("/api/v1/playlist", func(w http.ResponseWriter, r *http.Request) {
		playlist := generatePlaylist(videos, ip, port, useAuth, "", "")
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Write([]byte(playlist))
	}).Methods("GET")

	router.HandleFunc("/api/v1/playlist/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(videoInfos)
	}).Methods("GET")

	router.HandleFunc("/api/v1/playlist/{video_id}", func(w http.ResponseWriter, r *http.Request) {
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
	}).Methods("GET")

	router.PathPrefix("/videos/").Handler(http.StripPrefix("/videos/", http.FileServer(http.Dir(directoryPath))))

	// Create a new CORS handler
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"*"},
	})

	// Wrap the router with the CORS handler
	handler := c.Handler(router)

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Printf(green("Starting API server at http://%s:%d\n"), ip, port)
	fmt.Println(yellow("Available endpoints:"))
	fmt.Printf("  GET /api/v1/playlist\n")
	fmt.Printf("  GET /api/v1/playlist/list\n")
	fmt.Printf("  GET /api/v1/playlist/{video_id}\n")

	return http.ListenAndServe(addr, handler)
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
	case "duration":
		sort.Slice(videos, func(i, j int) bool {
			return videos[i].Duration > videos[j].Duration // Sort descending
		})
	}
}
