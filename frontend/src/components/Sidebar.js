import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import {
  Box,
  List,
  ListItem,
  ListItemIcon,
  ListItemText,
  Divider,
  IconButton,
  Tooltip
} from '@mui/material';
import {
  Dashboard as DashboardIcon,
  Storage as ServerIcon,
  FindInPage as DiscoveryIcon,
  Code as QueryIcon,
  ChevronLeft as ChevronLeftIcon,
  ChevronRight as ChevronRightIcon
} from '@mui/icons-material';

function Sidebar({ open, toggleSidebar }) {
  const location = useLocation();
  
  const menuItems = [
    { text: 'Dashboard', path: '/', icon: <DashboardIcon /> },
    { text: 'Servers', path: '/servers', icon: <ServerIcon /> },
    { text: 'Discoveries', path: '/discoveries', icon: <DiscoveryIcon /> },
    { text: 'SQL Query', path: '/query', icon: <QueryIcon /> },
  ];

  console.log("Sidebar rendering, open:", open);

  return (
    <Box
      sx={{
        position: 'fixed',
        top: 0,
        left: 0,
        width: open ? 240 : 60,
        height: '100vh',
        backgroundColor: 'background.paper',
        borderRight: '1px solid rgba(255, 255, 255, 0.12)',
        transition: 'width 0.3s ease',
        zIndex: (theme) => theme.zIndex.drawer,
        pt: 8, // Space for AppBar
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden'
      }}
    >
      <Box
        sx={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: open ? 'flex-end' : 'center',
          padding: '8px',
        }}
      >
        <IconButton onClick={toggleSidebar}>
          {open ? <ChevronLeftIcon /> : <ChevronRightIcon />}
        </IconButton>
      </Box>
      <Divider />
      <List>
        {menuItems.map((item) => {
          const isActive = location.pathname === item.path;
          
          return (
            <ListItem
              button
              component={Link}
              to={item.path}
              key={item.text}
              sx={{
                backgroundColor: isActive ? 'rgba(144, 202, 249, 0.2)' : 'transparent',
                '&:hover': {
                  backgroundColor: isActive ? 'rgba(144, 202, 249, 0.3)' : 'rgba(255, 255, 255, 0.08)',
                },
                borderLeft: isActive ? '4px solid #90caf9' : '4px solid transparent',
                paddingLeft: open ? 2 : 1.5,
              }}
            >
              {open ? (
                <>
                  <ListItemIcon sx={{ minWidth: 40, color: isActive ? 'primary.main' : 'inherit' }}>
                    {item.icon}
                  </ListItemIcon>
                  <ListItemText primary={item.text} />
                </>
              ) : (
                <Tooltip title={item.text} placement="right">
                  <ListItemIcon sx={{ minWidth: 0, justifyContent: 'center', color: isActive ? 'primary.main' : 'inherit' }}>
                    {item.icon}
                  </ListItemIcon>
                </Tooltip>
              )}
            </ListItem>
          );
        })}
      </List>
    </Box>
  );
}

export default Sidebar; 