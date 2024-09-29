import { spawn } from 'child_process';
import axios from 'axios';
import inquirer from 'inquirer';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

(async () => {
  const { videoPath } = await inquirer.prompt([
    {
      type: 'input',
      name: 'videoPath',
      message: 'Enter the path for the video directory:',
      default: '/home/tadeas/xxx',
    }
  ]);

  const goStreamPath = path.join(__dirname, 'go_stream');
  const backend = spawn(goStreamPath, ['api', '--path', videoPath, '-r'], {
    cwd: path.join(__dirname, '..'),
    stdio: 'inherit'
  });

  const checkBackend = async () => {
    try {
      await axios.get('http://localhost:8069/api/v1/playlist/list');
      console.log('Backend is ready. Starting frontend...');
      
      const frontend = spawn('npm', ['run', 'start:frontend'], {
        cwd: path.join(__dirname, 'frontend'),
        stdio: 'inherit'
      });

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

  setTimeout(checkBackend, 2000);

  process.on('SIGINT', () => {
    backend.kill();
    process.exit();
  });
})();