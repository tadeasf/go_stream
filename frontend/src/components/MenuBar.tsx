import React from 'react';
import { AppBar, Toolbar, Button } from '@mui/material';
import { Link } from 'react-router-dom';

const MenuBar: React.FC = () => {
  return (
    <AppBar position="static">
      <Toolbar>
        <Button color="inherit" component={Link} to="/">
          Player
        </Button>
        <Button color="inherit" component={Link} to="/playlist-maker">
          Playlist Maker
        </Button>
      </Toolbar>
    </AppBar>
  );
};

export default MenuBar;