import { useState, useEffect } from 'react';
import { useQuery } from 'react-query';
import { Container, Box, CircularProgress } from '@mui/material';
import axios from 'axios';
import VideoPlayer from '../components/VideoPlayer';
import Grid from '../components/Grid';

const API_URL = 'http://185.187.169.230:8069';

interface Video {
  id: string;
  path: string;
  size: number;
}

function Player() {
  const [selectedVideo, setSelectedVideo] = useState<string | null>(null);
  const [currentVideoId, setCurrentVideoId] = useState<string | null>(null);
  const [isVideoLoading, setIsVideoLoading] = useState(true);

  const { data: videos, isLoading, error } = useQuery<Video[]>('videos', async () => {
    const response = await axios.get(`${API_URL}/api/v1/playlist/list`);
    console.log('API Response:', response.data);
    return response.data;
  });

  useEffect(() => {
    if (videos && videos.length > 0) {
      const firstVideo = videos[0];
      setSelectedVideo(`${API_URL}/videos/${firstVideo.path}`);
      setCurrentVideoId(firstVideo.id);
    }
  }, [videos]);

  const handleVideoReady = () => {
    setIsVideoLoading(false);
  };

  const handleVideoSelect = (videoPath: string, videoId: string) => {
    setSelectedVideo(videoPath);
    setCurrentVideoId(videoId);
    setIsVideoLoading(true);
  };

  const handlePrevNext = (direction: 'prev' | 'next') => {
    if (!videos || videos.length === 0) return;
    const currentIndex = videos.findIndex(video => video.id === currentVideoId);
    if (currentIndex === -1) return;

    const newIndex = direction === 'prev' ? 
      (currentIndex - 1 + videos.length) % videos.length : 
      (currentIndex + 1) % videos.length;

    const newVideo = videos[newIndex];
    handleVideoSelect(`${API_URL}/videos/${newVideo.path}`, newVideo.id);
  };

  if (isLoading) return <CircularProgress />;
  if (error) return <div>An error occurred: {(error as Error).message}</div>;

  return (
    <Container maxWidth={false}>
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
          <Grid videos={videos} onVideoSelect={handleVideoSelect} currentVideoId={currentVideoId} />
        )}
      </Box>
    </Container>
  );
}

export default Player;