# go_stream

go_stream is a simple video streaming server CLI application written in Go.

## Features

- Serve video files over HTTP
- Generate M3U8 playlists
- Basic authentication support
- Recursive directory scanning
- Sorting videos by name, size, or duration

## Installation

### Prerequisites

- Go 1.23.1 or later

### Building from source

1. Clone the repository:
   ```
   git clone https://github.com/tadeasf/go_stream.git
   cd go_stream
   ```

2. Build the application:
   ```
   go build -o go_stream src/main.go
   ```

3. (Optional) Move the binary to a directory in your PATH:
   ```
   sudo mv go_stream /usr/local/bin/
   ```

## Usage

### Serving videos

To start the streaming server:

```sh
go_stream serve [flags]
```


Flags:
- `-r, --recursive`: Search for videos recursively
- `-p, --port int`: Port to serve on (default 8069)
- `--auth`: Enable basic authentication
- `--sort string`: Sort videos by: name, size, or duration

Example:

```sh
go_stream serve -r -p 8080 --auth --sort size
```

```sh
go_stream serve -r -p 8080 --auth --sort size
```



### Setting up basic authentication

To configure basic authentication:

```sh
go_stream basic_auth
```

This command will prompt you to enter a username and password.

## Development

### Project Structure

- `src/main.go`: Entry point of the application
- `src/commands/serve.go`: Implementation of the serve command
- `src/commands/basicAuth.go`: Implementation of the basic_auth command and authentication utilities

### Adding new commands

1. Create a new file in the `src/commands` directory for your command.
2. Implement your command logic.
3. Add the command to `src/main.go` in the `init()` function:

   ```go
   rootCmd.AddCommand(commands.YourNewCommand)
   ```

### Modifying existing commands

To modify the behavior of existing commands, edit the corresponding files in the `src/commands` directory.

### Building for different platforms

To build for a specific platform, use the `GOOS` and `GOARCH` environment variables:

```sh
GOOS=linux GOARCH=amd64 go build -o go_stream_linux_amd64 src/main.go
GOOS=darwin GOARCH=amd64 go build -o go_stream_macos_amd64 src/main.go
GOOS=windows GOARCH=amd64 go build -o go_stream_windows_amd64.exe src/main.go
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License.