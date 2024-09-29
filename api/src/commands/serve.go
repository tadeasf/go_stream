package commands

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"bufio"

	"github.com/c-bata/go-prompt"
	"github.com/eiannone/keyboard"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

var (
	recursive bool
	port      int
	useAuth   bool
	sortBy    string
)

var ServeCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the streaming server",
	RunE:  serveAction,
}

func init() {
	ServeCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Search for videos recursively")
	ServeCmd.Flags().IntVarP(&port, "port", "p", 8069, "Port to serve on")
	ServeCmd.Flags().BoolVar(&useAuth, "auth", false, "Enable basic authentication")
	ServeCmd.Flags().StringVar(&sortBy, "sort", "", "Sort videos by: name or size")
}

type Video struct {
	Path string
	Size int64
}

func serveAction(cmd *cobra.Command, args []string) error {
	recursive, _ := cmd.Flags().GetBool("recursive")
	port, _ := cmd.Flags().GetInt("port")
	useAuth, _ := cmd.Flags().GetBool("auth")
	sortBy, _ := cmd.Flags().GetString("sort")

	// Channel to signal server stop
	stop := make(chan struct{})

	var dir string
	if isatty.IsTerminal(os.Stdin.Fd()) {
		// If we're in a terminal, use prompt
		p := prompt.New(
			func(in string) {},
			pathCompleter,
			prompt.OptionPrefix("Enter the directory path: "),
			prompt.OptionPrefixTextColor(prompt.Yellow),
			prompt.OptionSuggestionBGColor(prompt.DarkGray),
			prompt.OptionSuggestionTextColor(prompt.White),
			prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
			prompt.OptionSelectedSuggestionTextColor(prompt.Black),
			prompt.OptionInputTextColor(prompt.Cyan),
		)
		dir = p.Input()
	} else {
		// If not in a terminal, read from stdin
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the directory path: ")
		dir, _ = reader.ReadString('\n')
	}

	dir = strings.TrimSpace(dir)
	if dir == "" {
		return fmt.Errorf("%s", yellow("directory path cannot be empty"))
	}

	// Find videos concurrently
	videos := FindVideosConcurrent(dir, recursive)

	// Sort videos based on the sortBy flag
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

	// Debug print
	fmt.Println("Sorted videos:")
	for _, v := range videos {
		fmt.Printf("Path: %s, Size: %d\n", v.Path, v.Size)
	}

	ip := GetOutboundIP()
	var handler http.Handler = http.DefaultServeMux

	var username, password string
	if useAuth {
		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load authentication config: %s", err)
		}
		username = config.Username
		password = config.Password
		handler = BasicAuth(handler, username, password)
	}

	playlist := generatePlaylist(videos, ip, port, useAuth, username, password)

	http.HandleFunc("/playlist.m3u8", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Write([]byte(playlist))
	})

	http.Handle("/videos/", http.StripPrefix("/videos/", http.FileServer(http.Dir(dir))))

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		fmt.Printf(green("Serving playlist at http://%s:%d/playlist.m3u8\n"), ip, port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Printf("ListenAndServe(): %v", err)
		}
	}()

	fmt.Println("Press 'q' to quit")

	// Handle user input using keyboard package
	go func() {
		if err := keyboard.Open(); err != nil {
			log.Printf("Error opening keyboard: %v", err)
			return
		}
		defer keyboard.Close()

		for {
			char, key, err := keyboard.GetKey()
			if err != nil {
				log.Printf("Error getting key: %v", err)
				continue
			}
			if char == 'q' || char == 'Q' || key == keyboard.KeyCtrlC {
				close(stop)
				return
			}
		}
	}()

	// Handle interrupt
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	select {
	case <-c:
		fmt.Println(yellow("\nInterrupt received, shutting down..."))
	case <-stop:
		fmt.Println(yellow("\nQuit command received, shutting down..."))
	}

	fmt.Println(yellow("\nShutting down server..."))

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to gracefully shutdown the server
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %s", err)
	}

	fmt.Println(green("Server gracefully stopped"))
	return nil
}

func FindVideosConcurrent(root string, recursive bool) []Video {
	var wg sync.WaitGroup
	videoExtensions := []string{".mp4", ".avi", ".mov", ".mkv", ".webm"}
	videosCh := make(chan Video)

	wg.Add(1)
	go func() {
		defer wg.Done()
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && !recursive && path != root {
				return filepath.SkipDir
			}
			for _, ext := range videoExtensions {
				if strings.HasSuffix(strings.ToLower(path), ext) {
					relPath, _ := filepath.Rel(root, path)
					video := Video{
						Path: relPath,
						Size: info.Size(),
					}
					videosCh <- video
					break
				}
			}
			return nil
		})
	}()

	go func() {
		wg.Wait()
		close(videosCh)
	}()

	var videos []Video
	for video := range videosCh {
		videos = append(videos, video)
	}

	return videos
}

func generatePlaylist(videos []Video, ip net.IP, port int, useAuth bool, username, password string) string {
	var sb strings.Builder
	sb.WriteString("#EXTM3U\n")
	for _, video := range videos {
		sb.WriteString(fmt.Sprintf("#EXTINF:-1,%s\n", video.Path))
		if useAuth {
			sb.WriteString(fmt.Sprintf("http://%s:%s@%s:%d/videos/%s\n", username, password, ip, port, video.Path))
		} else {
			sb.WriteString(fmt.Sprintf("http://%s:%d/videos/%s\n", ip, port, video.Path))
		}
	}
	return sb.String()
}

func GetOutboundIP() net.IP {
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
	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}
