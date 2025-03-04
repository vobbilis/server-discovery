import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
import { API_BASE_URL } from '../config';
import {
  Grid,
  Card,
  CardContent,
  Typography,
  CircularProgress,
  Box,
  Paper,
  Divider,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  Button,
  Drawer,
  IconButton,
  List,
  ListItem,
  ListItemText,
  ListSubheader,
  Container
} from '@mui/material';
import { PieChart, Pie, Cell, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip, CartesianGrid } from 'recharts';
import CloseIcon from '@mui/icons-material/Close';

function Dashboard() {
  const [runningDiscoveries, setRunningDiscoveries] = useState([]);
  const [selectedServer, setSelectedServer] = useState(null);
  const [serverDetails, setServerDetails] = useState(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [detailsLoading, setDetailsLoading] = useState(false);
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await axios.get(`${API_BASE_URL}/api/stats`);
        setStats(response.data);
        setLoading(false);
      } catch (err) {
        setError(err.message);
        setLoading(false);
      }
    };
    fetchStats();
  }, []);

  // Handle running a discovery on a server
  const handleRunDiscovery = async (serverId) => {
    try {
      // Add the server ID to the running discoveries list
      setRunningDiscoveries([...runningDiscoveries, serverId]);
      
      // Make API call to start discovery
      const response = await fetch(`/api/servers/${serverId}/discover`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        }
      });
      
      if (!response.ok) {
        throw new Error(`API returned ${response.status}: ${response.statusText}`);
      }
      
      // Refresh stats after a short delay to allow the discovery to start
      setTimeout(() => {
        fetch('/api/stats');
        // Remove the server from running discoveries
        setRunningDiscoveries(runningDiscoveries.filter(id => id !== serverId));
      }, 2000);
    } catch (err) {
      console.error(`Error starting discovery for server ${serverId}:`, err);
      setError(`Failed to start discovery for server ID ${serverId}`);
      // Remove the server from running discoveries
      setRunningDiscoveries(runningDiscoveries.filter(id => id !== serverId));
    }
  };

  // Handle opening the server details drawer
  const handleServerClick = async (serverId) => {
    try {
      setDetailsLoading(true);
      setDrawerOpen(true);
      
      // Find the server in the current stats
      const server = stats.servers.find(s => s.id === serverId);
      setSelectedServer(server);
      
      // Fetch detailed server information
      const response = await fetch(`/api/servers/${serverId}`);
      if (!response.ok) {
        throw new Error(`API returned ${response.status}: ${response.statusText}`);
      }
      
      const data = await response.json();
      setServerDetails(data);
    } catch (err) {
      console.error(`Error fetching server details for server ${serverId}:`, err);
      setError(`Failed to load server details for ID ${serverId}`);
    } finally {
      setDetailsLoading(false);
    }
  };

  // Handle closing the drawer
  const handleCloseDrawer = () => {
    setDrawerOpen(false);
  };

  // Debug render
  console.log('Dashboard rendering with stats:', stats, 'loading:', loading, 'error:', error);

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

  if (!stats) {
    return (
      <Box sx={{ mt: 4 }}>
        <Typography color="error" variant="h6">
          No statistics available
        </Typography>
      </Box>
    );
  }

  const successRate = stats.success_rate || 0;
  const serverCount = stats.server_count || 0;
  const discoveryCount = stats.discovery_count || 0;

  // Convert region distribution to chart data
  const regionData = Object.entries(stats.region_distribution || {}).map(([name, value]) => ({
    name,
    value,
  }));

  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

  return (
    <Container maxWidth="xl">
      <Box sx={{ mt: 4, mb: 4 }}>
        <Typography variant="h4" gutterBottom>
          Dashboard
        </Typography>

        <Grid container spacing={3}>
          {/* Summary Cards */}
          <Grid item xs={12} md={4}>
            <Paper 
              component={Link} 
              to="/servers"
              sx={{ 
                p: 2, 
                display: 'flex', 
                flexDirection: 'column', 
                height: 140,
                textDecoration: 'none',
                cursor: 'pointer',
                '&:hover': {
                  bgcolor: 'action.hover'
                }
              }}
            >
              <Typography color="text.secondary" gutterBottom>
                Total Servers
              </Typography>
              <Typography component="p" variant="h4">
                {serverCount}
              </Typography>
            </Paper>
          </Grid>
          <Grid item xs={12} md={4}>
            <Paper 
              component={Link} 
              to="/discoveries"
              sx={{ 
                p: 2, 
                display: 'flex', 
                flexDirection: 'column', 
                height: 140,
                textDecoration: 'none',
                cursor: 'pointer',
                '&:hover': {
                  bgcolor: 'action.hover'
                }
              }}
            >
              <Typography color="text.secondary" gutterBottom>
                Total Discoveries
              </Typography>
              <Typography component="p" variant="h4">
                {discoveryCount}
              </Typography>
            </Paper>
          </Grid>
          <Grid item xs={12} md={4}>
            <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column', height: 140 }}>
              <Typography color="text.secondary" gutterBottom>
                Success Rate
              </Typography>
              <Typography component="p" variant="h4">
                {successRate.toFixed(1)}%
              </Typography>
            </Paper>
          </Grid>

          {/* Region Distribution Chart */}
          <Grid item xs={12}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Server Distribution by Region
              </Typography>
              <Box sx={{ height: 400 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart
                    data={regionData}
                    margin={{
                      top: 20,
                      right: 30,
                      left: 20,
                      bottom: 5,
                    }}
                  >
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis />
                    <Tooltip />
                    <Bar dataKey="value" fill="#8884d8" />
                  </BarChart>
                </ResponsiveContainer>
              </Box>
            </Paper>
          </Grid>

          {/* Region Distribution Pie Chart */}
          <Grid item xs={12} md={6}>
            <Paper sx={{ p: 2 }}>
              <Typography variant="h6" gutterBottom>
                Region Distribution (Pie Chart)
              </Typography>
              <Box sx={{ height: 400 }}>
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={regionData}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      label={({ name, percent }) => `${name} (${(percent * 100).toFixed(0)}%)`}
                      outerRadius={150}
                      fill="#8884d8"
                      dataKey="value"
                    >
                      {regionData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip />
                  </PieChart>
                </ResponsiveContainer>
              </Box>
            </Paper>
          </Grid>
        </Grid>
      </Box>
    </Container>
  );
}

export default Dashboard; 