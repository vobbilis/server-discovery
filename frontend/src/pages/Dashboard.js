import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
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
  ListSubheader
} from '@mui/material';
import { PieChart, Pie, Cell, ResponsiveContainer, BarChart, Bar, XAxis, YAxis, Tooltip } from 'recharts';
import CloseIcon from '@mui/icons-material/Close';

function Dashboard() {
  const [runningDiscoveries, setRunningDiscoveries] = useState([]);
  const [selectedServer, setSelectedServer] = useState(null);
  const [serverDetails, setServerDetails] = useState(null);
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [detailsLoading, setDetailsLoading] = useState(false);
  const [stats, setStats] = useState({
    serverCount: 0,
    discoveryCount: 0,
    successRate: 0,
    regions: {},
    lastDiscovery: null,
    recentDiscoveries: [],
    servers: []
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchStats();
  }, []);

  const fetchStats = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/stats');
      if (!response.ok) {
        throw new Error(`API returned ${response.status}: ${response.statusText}`);
      }
      const data = await response.json();
      setStats(data);
      setError(null);
    } catch (err) {
      console.error('Error fetching stats:', err);
      setError('Failed to load dashboard data. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

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
        fetchStats();
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
          No data available
        </Typography>
      </Box>
    );
  }

  // Prepare data for charts
  const successRateData = [
    { name: 'Success', value: stats.successRate },
    { name: 'Failure', value: 100 - stats.successRate },
  ];

  const regionData = Object.entries(stats.regions || {}).map(([region, count]) => ({
    name: region || 'Unknown',
    count,
  }));

  const COLORS = ['#00C49F', '#FF8042', '#0088FE', '#FFBB28', '#8884d8'];

  return (
    <>
      <Box sx={{ 
        backgroundColor: 'background.paper', 
        padding: 2, 
        borderRadius: 1,
        boxShadow: 1
      }}>
        <Typography variant="h4" gutterBottom>
          Server Discovery Dashboard
        </Typography>
        
        {/* Stats Cards */}
        <Grid container spacing={3} sx={{ mb: 3 }}>
          <Grid item xs={12} md={4}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Servers
                </Typography>
                <Typography variant="h3" color="primary">
                  {stats.serverCount}
                </Typography>
                <Typography variant="body2" color="textSecondary">
                  <Link to="/servers">View all servers</Link>
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Discoveries
                </Typography>
                <Typography variant="h3" color="primary">
                  <Link 
                    to="/discoveries" 
                    style={{ 
                      color: 'inherit', 
                      textDecoration: 'none',
                      '&:hover': {
                        textDecoration: 'underline',
                        color: 'primary.main'
                      }
                    }}
                  >
                    {stats.discoveryCount}
                  </Link>
                </Typography>
                <Typography variant="body2" color="textSecondary">
                  Total discovery operations
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Success Rate
                </Typography>
                <Typography variant="h3" color="primary">
                  {stats.successRate.toFixed(1)}%
                </Typography>
                <Typography variant="body2" color="textSecondary">
                  Discovery success rate
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
        
        {/* Charts and Tables */}
        <Grid container spacing={3}>
          <Grid item xs={12} md={6}>
            <Paper sx={{ p: 2, height: 300 }}>
              <Typography variant="h6" gutterBottom>
                Success Rate
              </Typography>
              <ResponsiveContainer width="100%" height="90%">
                <PieChart>
                  <Pie
                    data={successRateData}
                    cx="50%"
                    cy="50%"
                    labelLine={false}
                    outerRadius={80}
                    fill="#8884d8"
                    dataKey="value"
                    label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(1)}%`}
                  >
                    {successRateData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={index === 0 ? '#00C49F' : '#FF8042'} />
                    ))}
                  </Pie>
                  <Tooltip />
                </PieChart>
              </ResponsiveContainer>
            </Paper>
          </Grid>
          <Grid item xs={12} md={6}>
            <Paper sx={{ p: 2, height: 300 }}>
              <Typography variant="h6" gutterBottom>
                Servers by Region
              </Typography>
              <ResponsiveContainer width="100%" height="90%">
                <BarChart data={regionData}>
                  <XAxis dataKey="name" />
                  <YAxis />
                  <Tooltip />
                  <Bar dataKey="count" fill="#8884d8">
                    {regionData.map((entry, index) => (
                      <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </Paper>
          </Grid>
          <Grid item xs={12}>
            <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
              <Typography component="h2" variant="h6" color="primary" gutterBottom>
                Recent Discoveries
              </Typography>
              <TableContainer component={Paper}>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Hostname</TableCell>
                      <TableCell>Port</TableCell>
                      <TableCell>Region</TableCell>
                      <TableCell>OS Type</TableCell>
                      <TableCell>Tags</TableCell>
                      <TableCell>Last Discovery</TableCell>
                      <TableCell width="150">Actions</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {stats.servers && stats.servers.length > 0 ? (
                      stats.servers.map((server) => (
                        <TableRow key={server.id}>
                          <TableCell>
                            <Link 
                              href="#" 
                              onClick={(e) => {
                                e.preventDefault();
                                handleServerClick(server.id);
                              }}
                              style={{ cursor: 'pointer', textDecoration: 'none', color: '#1976d2' }}
                            >
                              {server.hostname}
                            </Link>
                          </TableCell>
                          <TableCell>{server.port}</TableCell>
                          <TableCell>{server.region}</TableCell>
                          <TableCell>
                            {server.port === 22 ? (
                              <Chip label="Linux" color="success" size="small" />
                            ) : (
                              <Chip label="Windows" color="primary" size="small" />
                            )}
                          </TableCell>
                          <TableCell>
                            {server.tags && server.tags.map((tag) => (
                              <Chip 
                                key={tag.key} 
                                label={`${tag.key}: ${tag.value}`} 
                                size="small" 
                                sx={{ mr: 0.5, mb: 0.5 }}
                              />
                            ))}
                          </TableCell>
                          <TableCell>
                            {server.last_discovery ? new Date(server.last_discovery).toLocaleString() : 'Never'}
                          </TableCell>
                          <TableCell>
                            <Button 
                              variant="contained" 
                              size="small" 
                              onClick={() => handleRunDiscovery(server.id)}
                              disabled={runningDiscoveries.includes(server.id)}
                            >
                              {runningDiscoveries.includes(server.id) ? 'Running...' : 'Run Discovery'}
                            </Button>
                          </TableCell>
                        </TableRow>
                      ))
                    ) : (
                      <TableRow>
                        <TableCell colSpan={7} align="center">
                          No servers available
                        </TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </TableContainer>
            </Paper>
          </Grid>
        </Grid>
      </Box>
      
      {/* Server Details Drawer */}
      <Drawer
        anchor="right"
        open={drawerOpen}
        onClose={handleCloseDrawer}
        sx={{
          width: 400,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: 400,
            boxSizing: 'border-box',
            padding: 2
          },
        }}
      >
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
          <Typography variant="h6">
            {selectedServer ? selectedServer.hostname : 'Server Details'}
          </Typography>
          <IconButton onClick={handleCloseDrawer}>
            <CloseIcon />
          </IconButton>
        </Box>
        
        <Divider sx={{ mb: 2 }} />
        
        {detailsLoading ? (
          <Box sx={{ display: 'flex', justifyContent: 'center', mt: 4 }}>
            <CircularProgress />
          </Box>
        ) : serverDetails ? (
          <>
            <List>
              <ListSubheader>Basic Information</ListSubheader>
              <ListItem>
                <ListItemText primary="Hostname" secondary={serverDetails.hostname} />
              </ListItem>
              <ListItem>
                <ListItemText primary="Region" secondary={serverDetails.region} />
              </ListItem>
              <ListItem>
                <ListItemText primary="OS" secondary={`${serverDetails.os_name} (${serverDetails.os_version})`} />
              </ListItem>
              <ListItem>
                <ListItemText 
                  primary="Last Boot Time" 
                  secondary={new Date(serverDetails.last_boot_time).toLocaleString()} 
                />
              </ListItem>
              
              <ListSubheader>Hardware</ListSubheader>
              <ListItem>
                <ListItemText primary="CPU" secondary={`${serverDetails.cpu_model} (${serverDetails.cpu_count} cores)`} />
              </ListItem>
              <ListItem>
                <ListItemText primary="Memory" secondary={`${serverDetails.memory_total_gb} GB`} />
              </ListItem>
              <ListItem>
                <ListItemText 
                  primary="Disk" 
                  secondary={`${serverDetails.disk_free_gb} GB free of ${serverDetails.disk_total_gb} GB`} 
                />
              </ListItem>
              
              <ListSubheader>Network</ListSubheader>
              {serverDetails.ip_addresses && serverDetails.ip_addresses.map((ip, index) => (
                <ListItem key={index}>
                  <ListItemText 
                    primary={ip.interface_name} 
                    secondary={ip.ip_address} 
                  />
                </ListItem>
              ))}
              
              <ListSubheader>Tags</ListSubheader>
              <ListItem>
                <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                  {serverDetails.tags && serverDetails.tags.map((tag) => (
                    <Chip 
                      key={tag.key} 
                      label={`${tag.key}: ${tag.value}`} 
                      size="small" 
                    />
                  ))}
                </Box>
              </ListItem>
            </List>
            
            <Box sx={{ mt: 2, display: 'flex', justifyContent: 'center' }}>
              <Button 
                variant="contained" 
                onClick={() => {
                  handleRunDiscovery(serverDetails.id);
                  handleCloseDrawer();
                }}
                disabled={runningDiscoveries.includes(serverDetails.id)}
              >
                {runningDiscoveries.includes(serverDetails.id) ? 'Discovery Running...' : 'Run Discovery'}
              </Button>
            </Box>
          </>
        ) : (
          <Typography>No server details available</Typography>
        )}
      </Drawer>
    </>
  );
}

export default Dashboard; 