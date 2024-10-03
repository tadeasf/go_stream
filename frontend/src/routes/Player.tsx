import { useState, useEffect } from 'react';
import { useQuery } from 'react-query';
import { Container, Box, CircularProgress, Button } from '@mui/material';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/UseAuth';
import VideoPlayer from '../components/VideoPlayer';
import Grid from '../components/Grid';
import { GridSortModel } from '@mui/x-data-grid';

// Replace the hardcoded API_URL with a function to get the base URL
const getApiUrl = () => {
  return `${window.location.protocol}//${window.location.hostname}:8069`;
};

interface Video {
  id: string;
  path: string;
  size: number;
}

function Player() {
  const [selectedVideo, setSelectedVideo] = useState<string | null>(null);
  const [currentVideoId, setCurrentVideoId] = useState<string | null>(null);
  const [isVideoLoading, setIsVideoLoading] = useState(true);
  const [sortModel, setSortModel] = useState<GridSortModel>([]);

  const { data: videos, isLoading, error } = useQuery<Video[]>('videos', async () => {
    const response = await axios.get(`${getApiUrl()}/api/v1/playlist/list`);
    console.log('API Response:', response.data);
    return response.data;
  });

  const navigate = useNavigate();
  const { logout } = useAuth();

  useEffect(() => {
    if (videos && videos.length > 0) {
      const firstVideo = videos[0];
      setSelectedVideo(`${getApiUrl()}/videos/${firstVideo.path}`);
      setCurrentVideoId(firstVideo.id);
    }
  }, [videos]);

  const handleVideoReady = () => {
    setIsVideoLoading(false);
  };

  const handleVideoSelect = (videoPath: string, videoId: string) => {
    setSelectedVideo(`${getApiUrl()}/videos/${videoPath}`);
    setCurrentVideoId(videoId);
    setIsVideoLoading(true);
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

  const handlePrevNext = (direction: 'prev' | 'next') => {
    const sortedVideos = getSortedVideos();
    if (!sortedVideos || sortedVideos.length === 0) return;
    const currentIndex = sortedVideos.findIndex(video => video.id === currentVideoId);
    if (currentIndex === -1) return;

    const newIndex = direction === 'prev' ? 
      (currentIndex - 1 + sortedVideos.length) % sortedVideos.length : 
      (currentIndex + 1) % sortedVideos.length;

    const newVideo = sortedVideos[newIndex];
    handleVideoSelect(newVideo.path, newVideo.id);
  };

  const handleSortModelChange = (newSortModel: GridSortModel) => {
    setSortModel(newSortModel);
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  if (isLoading) return <CircularProgress />;
  if (error) return <div>An error occurred: {(error as Error).message}</div>;

  return (
    <Container maxWidth={false}>
      <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
        <Button variant="contained" onClick={handleLogout}>Logout</Button>
      </Box>
      <Box sx={{ mt: 2, position: 'relative', minHeight: '400px' }}>
        {isVideoLoading && (
          <Box sx={{ position: 'absolute', top: 0, left: 0, right: 0, bottom: 0, display: 'flex', justifyContent: 'center', alignItems: 'center', backgroundColor: 'rgba(0, 0, 0, 0.1)' }}>
            <CircularProgress />
          </Box>
        )}
        <VideoPlayer 
          selectedVideo={selectedVideo} 
          onReady={handleVideoReady} 
          onPrevNext={handlePrevNext}
          currentVideoId={currentVideoId}
        />
      </Box>
      <Box sx={{ mt: 2 }}>
        {videos && (
          <Grid 
            videos={videos} 
            onVideoSelect={handleVideoSelect} 
            currentVideoId={currentVideoId}
            onSortModelChange={handleSortModelChange}
          />
        )}
      </Box>
    </Container>
  );
}

export default Player;