import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { Box } from '@mui/material';

import Navbar from './components/Navbar';
import Sidebar from './components/Sidebar';
import Dashboard from './pages/Dashboard';
import ServerList from './pages/ServerList';
import ServerDetails from './pages/ServerDetails';
import DiscoveryDetails from './pages/DiscoveryDetails';
import Discoveries from './pages/Discoveries';
import SQLQuery from './pages/SQLQuery';

const theme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: '#90caf9',
    },
    secondary: {
      main: '#f48fb1',
    },
    background: {
      default: '#121212',
      paper: '#1e1e1e',
    },
  },
});

function App() {
  const [sidebarOpen, setSidebarOpen] = useState(true);

  const toggleSidebar = () => {
    console.log("Toggling sidebar, current state:", sidebarOpen, "changing to:", !sidebarOpen);
    setSidebarOpen(!sidebarOpen);
  };

  console.log("App rendering, sidebarOpen:", sidebarOpen);

  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Router>
        <Box sx={{ 
          display: 'grid',
          gridTemplateAreas: `
            "header header"
            "sidebar main"
          `,
          gridTemplateRows: '64px 1fr',
          gridTemplateColumns: `${sidebarOpen ? '240px' : '60px'} 1fr`,
          minHeight: '100vh',
          transition: 'grid-template-columns 0.3s ease'
        }}>
          <Box sx={{ gridArea: 'header' }}>
            <Navbar toggleSidebar={toggleSidebar} />
          </Box>
          <Box sx={{ gridArea: 'sidebar' }}>
            <Sidebar open={sidebarOpen} toggleSidebar={toggleSidebar} />
          </Box>
          <Box sx={{ 
            gridArea: 'main', 
            p: 3,
            overflow: 'auto'
          }}>
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/servers" element={<ServerList />} />
              <Route path="/servers/:id" element={<ServerDetails />} />
              <Route path="/discoveries" element={<Discoveries />} />
              <Route path="/discoveries/:id" element={<DiscoveryDetails />} />
              <Route path="/query" element={<SQLQuery />} />
            </Routes>
          </Box>
        </Box>
      </Router>
    </ThemeProvider>
  );
}

export default App; 