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

function DiscoveryDetails() {
  const { id } = useParams();
  const [discovery, setDiscovery] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [tabValue, setTabValue] = useState(0);

  useEffect(() => {
    fetch(`/api/discoveries/${id}`)
      .then(response => {
        if (!response.ok) {
          throw new Error('Failed to fetch discovery details');
        }
        return response.json();
      })
      .then(data => {
        setDiscovery(data);
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

  const duration = new Date(discovery.end_time) - new Date(discovery.start_time);
  const durationSeconds = Math.round(duration / 1000);

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Discovery Details
      </Typography>
      
      <Paper sx={{ p: 3, mb: 3 }}>
        <Grid container spacing={2}>
          <Grid item xs={12} md={6}>
            <Typography variant="h5">
              <Link to={`/servers/${discovery.server_id}`} style={{ textDecoration: 'none', color: '#90caf9' }}>
                {discovery.server_hostname}
              </Link>
            </Typography>
            <Typography variant="subtitle1" color="textSecondary">
              {discovery.server_region || 'Unknown region'} - Port {discovery.server_port}
            </Typography>
            <Box sx={{ mt: 2 }}>
              {discovery.success ? 
                <Chip label="Success" color="success" /> : 
                <Chip label="Failed" color="error" />
              }
            </Box>
          </Grid>
          <Grid item xs={12} md={6}>
            <Typography variant="body1">
              <strong>Start Time:</strong> {format(new Date(discovery.start_time), 'yyyy-MM-dd HH:mm:ss')}
            </Typography>
            <Typography variant="body1">
              <strong>End Time:</strong> {format(new Date(discovery.end_time), 'yyyy-MM-dd HH:mm:ss')}
            </Typography>
            <Typography variant="body1">
              <strong>Duration:</strong> {durationSeconds} seconds
            </Typography>
            <Typography variant="body1">
              <strong>Message:</strong> {discovery.message}
            </Typography>
            {discovery.error && (
              <Typography variant="body1" color="error">
                <strong>Error:</strong> {discovery.error}
              </Typography>
            )}
            {discovery.output_path && (
              <Typography variant="body1">
                <strong>Output Path:</strong> {discovery.output_path}
              </Typography>
            )}
          </Grid>
        </Grid>
      </Paper>
      
      {discovery.success && (
        <Paper sx={{ mb: 3 }}>
          <Tabs value={tabValue} onChange={handleTabChange}>
            <Tab label="System Info" />
            <Tab label="IP Addresses" />
            <Tab label="Installed Software" />
            <Tab label="Running Services" />
            <Tab label="Open Ports" />
          </Tabs>
          <Divider />
          
          <Box sx={{ p: 2 }}>
            {tabValue === 0 && (
              <Grid container spacing={2}>
                <Grid item xs={12} md={6}>
                  <Typography variant="body1">
                    <strong>OS:</strong> {discovery.os_name} {discovery.os_version}
                  </Typography>
                  <Typography variant="body1">
                    <strong>CPU:</strong> {discovery.cpu_model} ({discovery.cpu_count} cores)
                  </Typography>
                  <Typography variant="body1">
                    <strong>Memory:</strong> {discovery.memory_total_gb} GB
                  </Typography>
                </Grid>
                <Grid item xs={12} md={6}>
                  <Typography variant="body1">
                    <strong>Disk Total:</strong> {discovery.disk_total_gb} GB
                  </Typography>
                  <Typography variant="body1">
                    <strong>Disk Free:</strong> {discovery.disk_free_gb} GB
                  </Typography>
                  <Typography variant="body1">
                    <strong>Last Boot:</strong> {discovery.last_boot_time ? 
                      format(new Date(discovery.last_boot_time), 'yyyy-MM-dd HH:mm') : 
                      'Unknown'
                    }
                  </Typography>
                </Grid>
              </Grid>
            )}
            
            {tabValue === 1 && (
              <TableContainer>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Address</TableCell>
                      <TableCell>Subnet Mask</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {discovery.ip_addresses && discovery.ip_addresses.map((ip, index) => (
                      <TableRow key={index}>
                        <TableCell>{ip.address}</TableCell>
                        <TableCell>{ip.subnet_mask}</TableCell>
                      </TableRow>
                    ))}
                    {(!discovery.ip_addresses || discovery.ip_addresses.length === 0) && (
                      <TableRow>
                        <TableCell colSpan={2} align="center">No IP addresses found</TableCell>
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
                      <TableCell>Version</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {discovery.installed_software && discovery.installed_software.map((software, index) => (
                      <TableRow key={index}>
                        <TableCell>{software.name}</TableCell>
                        <TableCell>{software.version}</TableCell>
                      </TableRow>
                    ))}
                    {(!discovery.installed_software || discovery.installed_software.length === 0) && (
                      <TableRow>
                        <TableCell colSpan={2} align="center">No software found</TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
            
            {tabValue === 3 && (
              <TableContainer>
                <Table>
                  <TableHead>
                    <TableRow>
                      <TableCell>Name</TableCell>
                      <TableCell>Status</TableCell>
                    </TableRow>
                  </TableHead>
                  <TableBody>
                    {discovery.running_services && discovery.running_services.map((service, index) => (
                      <TableRow key={index}>
                        <TableCell>{service.name}</TableCell>
                        <TableCell>{service.status}</TableCell>
                      </TableRow>
                    ))}
                    {(!discovery.running_services || discovery.running_services.length === 0) && (
                      <TableRow>
                        <TableCell colSpan={2} align="center">No services found</TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
            
            {tabValue === 4 && (
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
                    {discovery.open_ports && discovery.open_ports.map((port, index) => (
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
                    {(!discovery.open_ports || discovery.open_ports.length === 0) && (
                      <TableRow>
                        <TableCell colSpan={7} align="center">No open ports found</TableCell>
                      </TableRow>
                    )}
                  </TableBody>
                </Table>
              </TableContainer>
            )}
          </Box>
        </Paper>
      )}
    </Box>
  );
}

export default DiscoveryDetails; 