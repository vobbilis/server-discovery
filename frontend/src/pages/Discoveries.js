import React, { useState, useEffect } from 'react';
import { 
  Box, 
  Typography, 
  Paper, 
  Table, 
  TableBody, 
  TableCell, 
  TableContainer, 
  TableHead, 
  TableRow,
  CircularProgress,
  Chip
} from '@mui/material';
import { Link } from 'react-router-dom';

function Discoveries() {
  const [discoveries, setDiscoveries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    // In a real app, we would fetch all discoveries
    // For now, we'll just use the mock data from the stats endpoint
    fetch('/api/stats')
      .then(response => response.json())
      .then(data => {
        // Use the recentDiscoveries as our data source for now
        setDiscoveries(data.recentDiscoveries || []);
        setLoading(false);
      })
      .catch(error => {
        console.error('Error fetching discoveries:', error);
        setError('Failed to load discoveries');
        setLoading(false);
      });
  }, []);

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
        Discovery History
      </Typography>
      
      <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
        <TableContainer>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell>Server</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Time</TableCell>
                <TableCell>Details</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {discoveries.map((discovery) => (
                <TableRow key={discovery.id}>
                  <TableCell>{discovery.id}</TableCell>
                  <TableCell>{discovery.serverHostname}</TableCell>
                  <TableCell>
                    <Chip 
                      label={discovery.success ? 'Success' : 'Failed'} 
                      color={discovery.success ? 'success' : 'error'} 
                      size="small" 
                    />
                  </TableCell>
                  <TableCell>{new Date(discovery.endTime).toLocaleString()}</TableCell>
                  <TableCell>
                    <Link to={`/discoveries/${discovery.id}`}>
                      View Details
                    </Link>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </Paper>
    </Box>
  );
}

export default Discoveries; 