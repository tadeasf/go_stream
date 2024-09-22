package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

const version = "1.0.0"

var (
	cyan    = color.New(color.FgCyan).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
)

type Config struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

var configDir = filepath.Join(os.Getenv("HOME"), ".config", "go_stream")
var configFile = filepath.Join(configDir, "config.yaml")

func main() {
	app := &cli.App{
		Name:    "go_stream",
		Usage:   "A simple video streaming server",
		Version: version,
		Commands: []*cli.Command{
			{
				Name:  "serve",
				Usage: "Start the streaming server",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "Search for videos recursively",
					},
					&cli.IntFlag{
						Name:    "port",
						Aliases: []string{"p"},
						Value:   8069,
						Usage:   "Port to serve on",
					},
					&cli.BoolFlag{
						Name:  "auth",
						Usage: "Enable basic authentication",
					},
				},
				Action: serveAction,
			},
			{
				Name:   "basic_auth",
				Usage:  "Set up basic authentication",
				Action: basicAuthAction,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func serveAction(c *cli.Context) error {
	recursive := c.Bool("recursive")
	port := c.Int("port")
	useAuth := c.Bool("auth")

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a channel to handle interrupts
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Create a channel for input
	inputCh := make(chan string, 1)

	// Start a goroutine to read user input with path completion
	go func() {
		p := prompt.New(
			func(in string) {
				inputCh <- in
			},
			pathCompleter,
			prompt.OptionPrefix("Enter the directory path: "),
			prompt.OptionPrefixTextColor(prompt.Yellow),
			prompt.OptionSuggestionBGColor(prompt.DarkGray),
			prompt.OptionSuggestionTextColor(prompt.White),
			prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
			prompt.OptionSelectedSuggestionTextColor(prompt.Black),
			prompt.OptionInputTextColor(prompt.Cyan),
		)

		go func() {
			select {
			case <-ctx.Done():
				// Force the prompt to exit by sending a newline
				fmt.Println()
			}
		}()

		inputCh <- p.Input()
	}()

	// Wait for either user input or interrupt
	var dir string
	select {
	case input := <-inputCh:
		dir = strings.TrimSpace(input)
		cancel() // Cancel the context to stop the prompt
	case <-interrupt:
		cancel() // Cancel the context to stop the prompt
		fmt.Println(yellow("\nInterrupted. Exiting..."))
		return nil
	}

	if dir == "" {
		return fmt.Errorf(yellow("directory path cannot be empty"))
	}

	// Rest of the function remains the same
	videos := findVideos(dir, recursive)
	ip := getOutboundIP()
	playlist := generatePlaylist(videos, ip, port)

	var handler http.Handler = http.DefaultServeMux

	if useAuth {
		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load authentication config: %w", err)
		}
		handler = basicAuth(handler, config.Username, config.Password)
	}

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

	// Wait for interrupt signal to gracefully shutdown the server
	<-interrupt
	fmt.Println(yellow("\nShutting down server..."))

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	fmt.Println(green("Server gracefully stopped"))
	return nil
}

func basicAuthAction(c *cli.Context) error {
	fmt.Println(cyan("Setting up basic authentication"))

	username := prompt.Input("Enter username: ", nil)
	password := prompt.Input("Enter password: ", nil)

	config := Config{
		Username: username,
		Password: password,
	}

	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(&config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = os.WriteFile(configFile, data, 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println(green("Basic authentication configured successfully"))
	return nil
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
	return prompt.FilterHasPrefix(suggestions, d.GetWordBeforeCursor(), true)
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func basicAuth(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
