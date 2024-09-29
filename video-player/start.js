const { spawn } = require('child_process');
const axios = require('axios');

// Start the backend
const backend = spawn('go', ['run', '/home/tadeas/go_stream/src/main.go', 'api', '--path', '/home/tadeas/xxx', '-r'], {
  cwd: '..',
  stdio: 'inherit'
});

// Function to check if the backend is ready
const checkBackend = async () => {
  try {
    await axios.get('http://localhost:8069/api/v1/playlist/list');
    console.log('Backend is ready. Starting frontend...');
    // Start the frontend
    const frontend = spawn('npm', ['run', 'start:frontend'], {
      stdio: 'inherit'
    });
  } catch (error) {
    console.log('Backend not ready yet. Retrying in 1 second...');
    setTimeout(checkBackend, 1000);
  }
};

// Start checking the backend after a short delay
setTimeout(checkBackend, 2000);

// Handle process termination
process.on('SIGINT', () => {
  backend.kill();
  process.exit();
});