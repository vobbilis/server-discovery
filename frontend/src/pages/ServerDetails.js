import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../config';
import {
  Box,
  Typography,
  Paper,
  Grid,
  Chip,
  CircularProgress,
  Divider,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Tabs,
  Tab,
  List,
  ListItem,
  ListItemText,
  Container,
} from '@mui/material';
import { format } from 'date-fns';
import ServerDetailsPanel from '../components/ServerDetailsPanel';

/**
 * ServerDetails component displays detailed information about a server,
 * including its discovered open ports and configured services.
 * 
 * Note on Port Display:
 * - The "Open Ports" table shows actual network ports discovered during scanning
 * - These ports are different from service configuration ports
 * - A discovered open port may or may not correspond to a configured service
 */
function ServerDetails() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [server, setServer] = useState(null);
  const [discoveries, setDiscoveries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [tabValue, setTabValue] = useState(0);

  useEffect(() => {
    const fetchData = async () => {
      try {
        // Fetch server details
        const serverResponse = await axios.get(`${API_BASE_URL}/api/servers/${id}`);
        setServer(serverResponse.data);
        
        // Fetch server discoveries
        const discoveriesResponse = await axios.get(`${API_BASE_URL}/api/servers/${id}/discoveries`);
        setDiscoveries(discoveriesResponse.data);
        setLoading(false);
      } catch (err) {
        setError(err.message);
        setLoading(false);
      }
    };
    
    fetchData();
  }, [id]);

  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
  };

  const handleClose = () => {
    navigate('/discoveries');
  };

  if (loading) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box sx={{ mt: 4 }}>
        <Typography color="error" variant="h6">
          Error: {error}
        </Typography>
      </Box>
    );
  }

  return (
    <ServerDetailsPanel
      server={server}
      discoveries={discoveries}
      onClose={handleClose}
    />
  );
}

export default ServerDetails; 