package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
)

const version = "1.0.0"

var (
	cyan    = color.New(color.FgCyan).SprintFunc()
	yellow  = color.New(color.FgYellow).SprintFunc()
	green   = color.New(color.FgGreen).SprintFunc()
	magenta = color.New(color.FgMagenta).SprintFunc()
)

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
				},
				Action: serveAction,
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

	fmt.Print(cyan("Enter the directory path: "))
	dir := prompt.Input("", pathCompleter,
		prompt.OptionPrefix(""),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
		prompt.OptionSuggestionTextColor(prompt.White),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSelectedSuggestionTextColor(prompt.Black),
	)
	dir = strings.TrimSpace(dir)

	if dir == "" {
		return fmt.Errorf(yellow("directory path cannot be empty"))
	}

	videos := findVideos(dir, recursive)
	ip := getOutboundIP()
	playlist := generatePlaylist(videos, ip, port)

	http.HandleFunc("/playlist.m3u8", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Write([]byte(playlist))
	})

	http.Handle("/videos/", http.StripPrefix("/videos/", http.FileServer(http.Dir(dir))))

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	fmt.Printf(green("Serving playlist at http://%s:%d/playlist.m3u8\n"), ip, port)
	return http.ListenAndServe(addr, nil)
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
	return prompt.FilterHasPrefix(suggestions, d.Text, true)
}
