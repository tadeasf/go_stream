import React, { useState } from 'react';
import { Button, Container } from '@mui/material';
import Grid from '../components/Grid';
import axios from 'axios';

const PlaylistMaker: React.FC = () => {
  const [selectedVideos, setSelectedVideos] = useState<string[]>([]);

  const handleVideoSelect = (_videoPath: string, videoId: string, isSelected: boolean) => {
    if (isSelected) {
      setSelectedVideos([...selectedVideos, videoId]);
    } else {
      setSelectedVideos(selectedVideos.filter(id => id !== videoId));
    }
  };

  const generatePlaylist = async () => {
    try {
      const response = await axios.post('/api/v1/generate-playlist', { videoIds: selectedVideos }, { responseType: 'blob' });
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

  return (
    <Container>
      <Grid 
        videos={[]} // You need to provide the videos prop
        onVideoSelect={handleVideoSelect}
        currentVideoId={null}
        onSortModelChange={() => {}} // Add an empty function or implement sorting logic
        showCheckbox={true}
      />
      <Button variant="contained" color="primary" onClick={generatePlaylist} style={{ marginTop: '20px' }}>
        Generate Playlist
      </Button>
    </Container>
  );
};

export default PlaylistMaker;