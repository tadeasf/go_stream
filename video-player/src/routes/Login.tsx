import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { TextField, Button, Box, Typography, Alert, CircularProgress } from '@mui/material';

// Hardcoded credentials
const VALID_USERNAME = 'admin';
const VALID_PASSWORD = 'admin';

function Login() {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [errorMessage, setErrorMessage] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();

  const handleLogin = () => {
    setErrorMessage('');
    if (username === VALID_USERNAME && password === VALID_PASSWORD) {
      setIsLoading(true);
      console.log('Login successful');
      setTimeout(() => {
        navigate('/player');
      }, 1000); // Simulate a delay before navigation
    } else {
      setErrorMessage('Invalid username or password');
    }
  };

  if (isLoading) {
    return (
      <Box
        sx={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          height: '100vh',
        }}
      >
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', alignItems: 'center', mt: 8 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        Login
      </Typography>
      {errorMessage && (
        <Alert severity="error" sx={{ mb: 2, width: '100%', maxWidth: 300 }}>
          {errorMessage}
        </Alert>
      )}
      <TextField
        label="Username"
        value={username}
        onChange={(e) => setUsername(e.target.value)}
        margin="normal"
      />
      <TextField
        label="Password"
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        margin="normal"
      />
      <Button
        variant="contained"
        onClick={handleLogin}
        sx={{ mt: 2 }}
      >
        Login
      </Button>
    </Box>
  );
}

export default Login;