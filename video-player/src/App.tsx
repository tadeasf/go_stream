import { useState, useEffect } from 'react';
import axios from 'axios';
import { DataGrid, GridColDef, GridPaginationModel, GridRenderCellParams } from '@mui/x-data-grid';
import ReactPlayer from 'react-player';
import { Container, Typography, Box } from '@mui/material';

interface Video {
  id: string;
  path: string;
  size: number;
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

const columns: GridColDef[] = [
  { field: 'id', headerName: 'ID', width: 70 },
  { field: 'path', headerName: 'Name', width: 500 },
  { 
    field: 'size', 
    headerName: 'Size', 
    width: 150,
    renderCell: (params: GridRenderCellParams) => formatSize(params.row.size),
  },
];

function App() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [selectedVideo, setSelectedVideo] = useState<string | null>(null);
  const [paginationModel, setPaginationModel] = useState<GridPaginationModel>({
    pageSize: 5,
    page: 0,
  });

  useEffect(() => {
    axios.get(`${API_URL}/api/v1/playlist/list`)
      .then(response => {
        setVideos(response.data);
        if (response.data.length > 0) {
          setSelectedVideo(`${API_URL}/videos/${response.data[0].path}`);
        }
      })
      .catch(error => console.error('Error fetching videos:', error));
  }, []);

  const handleRowClick = (params: any) => {
    setSelectedVideo(`${API_URL}/videos/${params.row.path}`);
  };

  return (
    <Container maxWidth={false}>
      <Typography variant="h4" component="h1" gutterBottom>
        Video Player
      </Typography>
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
    </Container>
  );
}

export default App;