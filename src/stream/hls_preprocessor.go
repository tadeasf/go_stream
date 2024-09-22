package stream

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func preprocessAction(c *cli.Context) error {
	inputDir := c.String("dir")
	outputDir := c.String("output")

	videos := findVideos(inputDir, true)

	for _, video := range videos {
		inputPath := filepath.Join(inputDir, video)
		outputPath := filepath.Join(outputDir, video)
		outputDir := filepath.Dir(outputPath)

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}

		if err := preprocessVideo(inputPath, outputPath); err != nil {
			return fmt.Errorf("failed to preprocess video %s: %v", video, err)
		}
	}

	return nil
}

func preprocessVideo(inputPath, outputPath string) error {
	outputDir := filepath.Dir(outputPath)
	outputName := filepath.Base(outputPath)

	cmd := exec.Command(
		"ffmpeg",
		"-i", inputPath,
		"-profile:v", "main",
		"-crf", "23",
		"-g", "60",
		"-sc_threshold", "0",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", filepath.Join(outputDir, outputName+"_%03d.ts"),
		filepath.Join(outputDir, outputName+".m3u8"),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
