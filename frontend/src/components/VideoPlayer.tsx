import { useState, useRef, useEffect } from 'react';
import ReactPlayer from 'react-player';
import { Box, Button, Typography } from '@mui/material';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import PauseIcon from '@mui/icons-material/Pause';
import LoopIcon from '@mui/icons-material/Loop';
import SkipPreviousIcon from '@mui/icons-material/SkipPrevious';
import SkipNextIcon from '@mui/icons-material/SkipNext';

interface VideoPlayerProps {
  selectedVideo: string | null;
  onReady?: () => void;
  onPrevNext: (direction: 'prev' | 'next') => void;
  currentVideoId: string | null;
}

function VideoPlayer({ selectedVideo, onReady, onPrevNext, currentVideoId }: VideoPlayerProps) {
  const [isPlaying, setIsPlaying] = useState(false);
  const [isLooping, setIsLooping] = useState(false);
  const [volume, setVolume] = useState(0.15);
  const playerRef = useRef<ReactPlayer>(null);

  useEffect(() => {
    setIsPlaying(false);
  }, [selectedVideo]);

  const handlePlayPause = () => {
    setIsPlaying(!isPlaying);
  };

  const handleLoopToggle = () => {
    setIsLooping(!isLooping);
  };

  const handleVolumeChange = (newVolume: number) => {
    setVolume(newVolume);
  };

  if (!selectedVideo) {
    return <Typography>No video selected</Typography>;
  }

  return (
    <Box>
      <ReactPlayer
        ref={playerRef}
        url={selectedVideo}
        controls
        width="100%"
        height="auto"
        style={{ backgroundColor: '#000' }}
        playing={isPlaying}
        loop={isLooping}
        volume={volume}
        onPlay={() => setIsPlaying(true)}
        onPause={() => setIsPlaying(false)}
        onVolumeChange={(e: any) => handleVolumeChange(parseFloat(e.target.volume))}
        onReady={onReady}
      />
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', mt: 2, gap: 2 }}>
        <Button variant="contained" onClick={() => onPrevNext('prev')} startIcon={<SkipPreviousIcon />}>
          Previous
        </Button>
        <Button
          variant="contained"
          startIcon={isPlaying ? <PauseIcon /> : <PlayArrowIcon />}
          onClick={handlePlayPause}
        >
          {isPlaying ? 'Pause' : 'Play'}
        </Button>
        <Button
          variant="contained"
          startIcon={<LoopIcon />}
          onClick={handleLoopToggle}
          color={isLooping ? "secondary" : "primary"}
        >
          Loop: {isLooping ? "On" : "Off"}
        </Button>
        <Button variant="contained" onClick={() => onPrevNext('next')} startIcon={<SkipNextIcon />}>
          Next
        </Button>
      </Box>
      <Typography variant="subtitle1" align="center" sx={{ mt: 2 }}>
        Currently playing: Video ID {currentVideoId}
      </Typography>
    </Box>
  );
}

export default VideoPlayer;