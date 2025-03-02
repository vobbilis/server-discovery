import React from 'react';
import { Outlet } from 'react-router-dom';
import { Box } from '@mui/material';

function Layout() {
  return (
    <Box sx={{ width: '100%', p: 2 }}>
      <Outlet />
    </Box>
  );
}

export default Layout; 