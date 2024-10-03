import React, { useState } from 'react';
import { Button, Container, CircularProgress } from '@mui/material';
import Grid from './Grid';
import axios from 'axios';
import { useQuery } from 'react-query';
import { GridSortModel } from '@mui/x-data-grid';

// Add this function to get the API URL
const getApiUrl = () => {
  return `${window.location.protocol}//${window.location.hostname}:8069`;
};

interface Video {
  id: string;
  path: string;
  size: number;
}

const PlaylistMaker: React.FC = () => {
  const [selectedVideos, setSelectedVideos] = useState<string[]>([]);

  const { data: videos, isLoading, error } = useQuery<Video[]>('videos', async () => {
    const response = await axios.get(`${getApiUrl()}/api/v1/playlist/list`);
    return response.data;
  });

  const handleVideoSelect = (_videoPath: string, videoId: string, isSelected: boolean) => {
    if (isSelected) {
      setSelectedVideos([...selectedVideos, videoId]);
    } else {
      setSelectedVideos(selectedVideos.filter(id => id !== videoId));
    }
  };

  const handleSortModelChange = (newSortModel: GridSortModel) => {
    // If you want to implement sorting logic in the future, you can do it here
    console.log('Sort model changed:', newSortModel);
  };

  const generatePlaylist = async () => {
    try {
      const response = await axios.post(`${getApiUrl()}/api/v1/generate-playlist`, { videoIds: selectedVideos }, { responseType: 'blob' });
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', 'playlist.m3u8');
      document.body.appendChild(link);
      link.click();
      link.remove();
    } catch (error) {
      console.error('Error generating playlist:', error);
    }
  };

  if (isLoading) return <CircularProgress />;
  if (error) return <div>An error occurred: {(error as Error).message}</div>;

  return (
    <Container>
      <Grid 
        videos={videos || []}
        onVideoSelect={handleVideoSelect}
        currentVideoId={null}
        onSortModelChange={handleSortModelChange}
        showCheckbox={true}
      />
      <Button variant="contained" color="primary" onClick={generatePlaylist} style={{ marginTop: '20px' }}>
        Generate Playlist
      </Button>
    </Container>
  );
};

export default PlaylistMaker;