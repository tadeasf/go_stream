package commands

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var BasicAuthCmd = &cobra.Command{
	Use:   "basic_auth",
	Short: "Set up basic authentication",
	RunE:  basicAuthAction,
}

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
)

type Config struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

var (
	configDir  = filepath.Join(os.Getenv("HOME"), ".config", "go_stream")
	configFile = filepath.Join(configDir, "config.yaml")
)

func basicAuthAction(cmd *cobra.Command, args []string) error {
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
