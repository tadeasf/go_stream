import { spawn } from 'child_process';
import axios from 'axios';
import inquirer from 'inquirer';

(async () => {
  // Prompt the user for the backend path
  const { path } = await inquirer.prompt([
    {
      type: 'input',
      name: 'path',
      message: 'Enter the path for the backend:',
      default: '/home/tadeas/xxx',
    }
  ]);

  // Start the backend with the provided path
  const backend = spawn('go', ['run', '/home/tadeas/go_stream/src/main.go', 'api', '--path', path, '-r'], {
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

      // Handle frontend process termination
      frontend.on('close', (code) => {
        console.log(`Frontend process exited with code ${code}`);
        backend.kill();
        process.exit();
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
})();
