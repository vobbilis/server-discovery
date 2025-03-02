import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
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
} from '@mui/material';
import { format } from 'date-fns';

function ServerDetails() {
  const { id } = useParams();
  const [server, setServer] = useState(null);
  const [discoveries, setDiscoveries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [tabValue, setTabValue] = useState(0);

  useEffect(() => {
    // Fetch server details
    fetch(`/api/servers/${id}`)
      .then(response => {
        if (!response.ok) {
          throw new Error('Failed to fetch server details');
        }
        return response.json();
      })
      .then(data => {
        setServer(data);
        
        // Fetch server discoveries
        return fetch(`/api/servers/${id}/discoveries`);
      })
      .then(response => {
        if (!response.ok) {
          throw new Error('Failed to fetch server discoveries');
        }
        return response.json();
      })
      .then(data => {
        setDiscoveries(data);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  }, [id]);

  const handleTabChange = (event, newValue) => {
    setTabValue(newValue);
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
    <Box>
      <Typography variant="h4" gutterBottom>
        Server Details
      </Typography>
      
      <Paper sx={{ p: 3, mb: 3 }}>
        <Grid container spacing={2}>
          <Grid item xs={12} md={6}>
            <Typography variant="h5">{server.hostname}</Typography>
            <Typography variant="subtitle1" color="textSecondary">
              {server.region || 'Unknown region'} - Port {server.port}
            </Typography>
            <Box sx={{ mt: 2 }}>
              {server.tags.map(tag => (
                <Chip 
                  key={tag.key} 
                  label={`${tag.key}: ${tag.value}`} 
                  sx={{ mr: 1, mb: 1 }} 
                />
              ))}
            </Box>
          </Grid>
          <Grid item xs={12} md={6}>
            <Typography variant="body1">
              <strong>OS:</strong> {server.os_name} {server.os_version}
            </Typography>
            <Typography variant="body1">
              <strong>CPU:</strong> {server.cpu_model} ({server.cpu_count} cores)
            </Typography>
            <Typography variant="body1">
              <strong>Memory:</strong> {server.memory_total_gb} GB
            </Typography>
            <Typography variant="body1">
              <strong>Disk:</strong> {server.disk_free_gb} GB free of {server.disk_total_gb} GB
            </Typography>
            <Typography variant="body1">
              <strong>Last Boot:</strong> {server.last_boot_time ? 
                format(new Date(server.last_boot_time), 'yyyy-MM-dd HH:mm') : 
                'Unknown'
              }
            </Typography>
          </Grid>
        </Grid>
      </Paper>
      
      <Paper sx={{ mb: 3 }}>
        <Tabs value={tabValue} onChange={handleTabChange}>
          <Tab label="IP Addresses" />
          <Tab label="Installed Software" />
          <Tab label="Running Services" />
          <Tab label="Open Ports" />
          <Tab label="Discovery History" />
        </Tabs>
        <Divider />
        
        <Box sx={{ p: 2 }}>
          {tabValue === 0 && (
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Address</TableCell>
                    <TableCell>Subnet Mask</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {server.ip_addresses && server.ip_addresses.map((ip, index) => (
                    <TableRow key={index}>
                      <TableCell>{ip.address}</TableCell>
                      <TableCell>{ip.subnet_mask}</TableCell>
                    </TableRow>
                  ))}
                  {(!server.ip_addresses || server.ip_addresses.length === 0) && (
                    <TableRow>
                      <TableCell colSpan={2} align="center">No IP addresses found</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
          
          {tabValue === 1 && (
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Name</TableCell>
                    <TableCell>Version</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {server.installed_software && server.installed_software.map((software, index) => (
                    <TableRow key={index}>
                      <TableCell>{software.name}</TableCell>
                      <TableCell>{software.version}</TableCell>
                    </TableRow>
                  ))}
                  {(!server.installed_software || server.installed_software.length === 0) && (
                    <TableRow>
                      <TableCell colSpan={2} align="center">No software found</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
          
          {tabValue === 2 && (
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Name</TableCell>
                    <TableCell>Status</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {server.running_services && server.running_services.map((service, index) => (
                    <TableRow key={index}>
                      <TableCell>{service.name}</TableCell>
                      <TableCell>{service.status}</TableCell>
                    </TableRow>
                  ))}
                  {(!server.running_services || server.running_services.length === 0) && (
                    <TableRow>
                      <TableCell colSpan={2} align="center">No services found</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
          
          {tabValue === 3 && (
            <TableContainer component={Paper} sx={{ mt: 3 }}>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Local Address</TableCell>
                    <TableCell>Local Port</TableCell>
                    <TableCell>Remote Address</TableCell>
                    <TableCell>Remote Port</TableCell>
                    <TableCell>State</TableCell>
                    <TableCell>Service</TableCell>
                    <TableCell>Process</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {server.open_ports && server.open_ports.map((port, index) => (
                    <TableRow key={index}>
                      <TableCell>{port.localIP || '0.0.0.0'}</TableCell>
                      <TableCell>{port.local_port}</TableCell>
                      <TableCell>{port.remoteIP || '*'}</TableCell>
                      <TableCell>{port.remotePort || '*'}</TableCell>
                      <TableCell>{port.state}</TableCell>
                      <TableCell>{port.description || 'Unknown'}</TableCell>
                      <TableCell>{port.processName ? `${port.processName} (${port.processID})` : 'Unknown'}</TableCell>
                    </TableRow>
                  ))}
                  {(!server.open_ports || server.open_ports.length === 0) && (
                    <TableRow>
                      <TableCell colSpan={7} align="center">No open ports found</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
          
          {tabValue === 4 && (
            <TableContainer>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Date</TableCell>
                    <TableCell>Status</TableCell>
                    <TableCell>Message</TableCell>
                    <TableCell>Details</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {discoveries.map(discovery => (
                    <TableRow key={discovery.id}>
                      <TableCell>
                        {format(new Date(discovery.end_time), 'yyyy-MM-dd HH:mm')}
                      </TableCell>
                      <TableCell>
                        {discovery.success ? 
                          <Chip label="Success" color="success" size="small" /> : 
                          <Chip label="Failed" color="error" size="small" />
                        }
                      </TableCell>
                      <TableCell>{discovery.message}</TableCell>
                      <TableCell>
                        <Link to={`/discoveries/${discovery.id}`}>
                          View Details
                        </Link>
                      </TableCell>
                    </TableRow>
                  ))}
                  {discoveries.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={4} align="center">No discovery history found</TableCell>
                    </TableRow>
                  )}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </Box>
      </Paper>
    </Box>
  );
}

export default ServerDetails; 