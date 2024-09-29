import { useState, useEffect } from 'react';
import axios from 'axios';
import { DataGrid, GridColDef, GridPaginationModel, GridRenderCellParams } from '@mui/x-data-grid';
import ReactPlayer from 'react-player';
import { Container, Box, TextField, Checkbox, FormControlLabel, Button, Autocomplete, IconButton, Dialog, DialogTitle, DialogContent, DialogActions } from '@mui/material';
import DeleteIcon from '@mui/icons-material/Delete';
import { DEFAULT_USERNAME, DEFAULT_PASSWORD } from './config';

interface Video {
  id: string;
  path: string;
  size: number;
}

interface PathSuggestion {
  path: string;
  isDir: boolean;
}

const API_URL = 'http://185.187.169.230:8069';

const formatSize = (sizeInBytes: number): string => {
  if (sizeInBytes >= 1073741824) {
    return `${(sizeInBytes / 1073741824).toFixed(2)} GB`;
  } else if (sizeInBytes >= 1048576) {
    return `${(sizeInBytes / 1048576).toFixed(2)} MB`;
  } else {
    return `${sizeInBytes} bytes`;
  }
};

function App() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [selectedVideo, setSelectedVideo] = useState<string | null>(null);
  const [paginationModel, setPaginationModel] = useState<GridPaginationModel>({
    pageSize: 5,
    page: 0,
  });
  const [newPath, setNewPath] = useState('');
  const [isRecursive, setIsRecursive] = useState(false);
  const [pathSuggestions, setPathSuggestions] = useState<PathSuggestion[]>([]);
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [showLoginDialog, setShowLoginDialog] = useState(true);

  useEffect(() => {
    fetchVideos();
  }, []);

  const fetchVideos = () => {
    axios.get(`${API_URL}/api/v1/playlist/list`)
      .then(response => {
        setVideos(response.data);
        if (response.data.length > 0) {
          setSelectedVideo(`${API_URL}/videos/${response.data[0].path}`);
        }
      })
      .catch(error => console.error('Error fetching videos:', error));
  };

  const handleRowClick = (params: any) => {
    setSelectedVideo(`${API_URL}/videos/${params.row.path}`);
  };

  const handleChangePath = () => {
    axios.post(`${API_URL}/api/v1/playlist`, {
      path: newPath,
      args: isRecursive ? '-r' : ''
    })
    .then(() => {
      // Wait for the server to restart
      setTimeout(() => {
        fetchVideos();
        setNewPath('');
      }, 2000);
    })
    .catch(error => console.error('Error changing path:', error));
  };

  const handlePathChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const path = event.target.value;
    setNewPath(path);
    fetchPathSuggestions(path);
  };

  const fetchPathSuggestions = (path: string) => {
    axios.get(`${API_URL}/api/v1/path-suggestions`, { params: { path } })
      .then(response => {
        setPathSuggestions(response.data);
      })
      .catch(error => console.error('Error fetching path suggestions:', error));
  };

  const handleDeleteVideo = (id: string) => {
    axios.delete(`${API_URL}/api/v1/playlist/${id}`)
      .then(() => {
        fetchVideos();
      })
      .catch(error => console.error('Error deleting video:', error));
  };

  const handleLogin = () => {
    if (username === DEFAULT_USERNAME && password === DEFAULT_PASSWORD) {
      setIsLoggedIn(true);
      setShowLoginDialog(false);
    } else {
      alert('Invalid credentials');
    }
  };

  const columns: GridColDef[] = [
    { field: 'id', headerName: 'ID', width: 70 },
    { field: 'path', headerName: 'Name', width: 500 },
    { 
      field: 'size', 
      headerName: 'Size', 
      width: 150,
      renderCell: (params: GridRenderCellParams) => formatSize(params.row.size),
    },
    {
      field: 'actions',
      headerName: 'Actions',
      width: 100,
      renderCell: (params: GridRenderCellParams) => (
        <IconButton onClick={() => handleDeleteVideo(params.row.id)}>
          <DeleteIcon />
        </IconButton>
      ),
    },
  ];

  if (!isLoggedIn) {
    return (
      <Dialog open={showLoginDialog} onClose={() => {}}>
        <DialogTitle>Login</DialogTitle>
        <DialogContent>
          <TextField
            autoFocus
            margin="dense"
            label="Username"
            type="text"
            fullWidth
            value={username}
            onChange={(e) => setUsername(e.target.value)}
          />
          <TextField
            margin="dense"
            label="Password"
            type="password"
            fullWidth
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={handleLogin}>Login</Button>
        </DialogActions>
      </Dialog>
    );
  }

  return (
    <Container maxWidth={false}>
      <Box sx={{ width: '100%', aspectRatio: '16/9', maxHeight: 'calc(100vh - 300px)' }}>
        {selectedVideo && (
          <ReactPlayer 
            url={selectedVideo} 
            controls 
            width="100%" 
            height="100%"
            style={{ backgroundColor: '#000' }}
          />
        )}
      </Box>
      <Box sx={{ height: 400, width: '100%', mt: 2 }}>
        <DataGrid
          rows={videos}
          columns={columns}
          paginationModel={paginationModel}
          onPaginationModelChange={setPaginationModel}
          pageSizeOptions={[5, 10, 25]}
          onRowClick={handleRowClick}
        />
      </Box>
      <Box sx={{ mt: 2, display: 'flex', alignItems: 'center' }}>
        <Autocomplete
          freeSolo
          options={pathSuggestions}
          getOptionLabel={(option) => typeof option === 'string' ? option : option.path}
          renderOption={(props, option) => {
            return (
              <li {...props} key={option.path}>
                {option.path}
              </li>
            );
          }}
          renderInput={(params) => (
            <TextField
              {...params}
              label="New Path"
              variant="outlined"
              onChange={handlePathChange}
            />
          )}
          value={newPath}
          onChange={(_, newValue) => {
            let updatedPath = '';
            if (typeof newValue === 'string') {
              updatedPath = newValue;
            } else if (newValue && newValue.path) {
              updatedPath = newValue.path;
            }
            
            // Append a "/" if it's not already there
            if (updatedPath && !updatedPath.endsWith('/')) {
              updatedPath += '/';
            }
            
            setNewPath(updatedPath);
            fetchPathSuggestions(updatedPath);
          }}
          sx={{ flexGrow: 1, mr: 2 }}
        />
        <FormControlLabel
          control={
            <Checkbox
              checked={isRecursive}
              onChange={(e) => setIsRecursive(e.target.checked)}
            />
          }
          label="Recursive"
        />
        <Button variant="contained" onClick={handleChangePath} sx={{ ml: 2 }}>
          Change Path
        </Button>
      </Box>
    </Container>
  );
}

export default App;