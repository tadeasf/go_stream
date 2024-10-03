import React, { useState } from 'react';
import { Button, Container, CircularProgress } from '@mui/material';
import Grid from '../components/Grid';
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
  const [sortModel, setSortModel] = useState<GridSortModel>([]);

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

  const getSortedVideos = () => {
    if (!videos || sortModel.length === 0) return videos;
    const { field, sort } = sortModel[0];
    return [...videos].sort((a, b) => {
      const aValue: string | number = a[field as keyof Video];
      const bValue: string | number = b[field as keyof Video];

      if (aValue < bValue) return sort === 'asc' ? -1 : 1;
      if (aValue > bValue) return sort === 'asc' ? 1 : -1;
      return 0;
    });
  };

  const handleSortModelChange = (newSortModel: GridSortModel) => {
    setSortModel(newSortModel);
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

  const sortedVideos = getSortedVideos();

  return (
    <Container>
      <Grid 
        videos={sortedVideos || []}
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