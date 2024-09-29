# Video Streaming Application

This project consists of a Go backend (API) and a React frontend for streaming videos.

## Prerequisites

- Go 1.23.1 or later
- Node.js 14.0.0 or later
- Bun 1.0.0 or later

## Setup Instructions

1. Build the Go backend:
   ```
   cd api
   go build -o go_stream src/main.go
   ```

2. Move the `go_stream` binary to the project root:
   ```
   mv go_stream ../
   cd ..
   ```

3. Install dependencies in the root directory:
   ```
   bun install
   ```

4. Install frontend dependencies:
   ```
   cd frontend
   bun install
   cd ..
   ```

5. Start the application:
   ```
   bun run start.js
   ```

This will start both the backend and frontend servers. Follow the prompts in the terminal to set up the video directory path.

## Usage

After starting the application, open your web browser and navigate to `http://localhost:5173` to access the video player interface.

For more detailed information about the API and frontend, refer to their respective README files in the `api` and `frontend` directories.