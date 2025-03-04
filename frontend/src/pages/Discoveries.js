import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { API_BASE_URL } from '../config';
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
  Container,
  Chip,
  Alert
} from '@mui/material';
import { format, isValid } from 'date-fns';
import ServerDetailsPanel from '../components/ServerDetailsPanel';

function Discoveries() {
  const [discoveries, setDiscoveries] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedServer, setSelectedServer] = useState(null);
  const [serverLoading, setServerLoading] = useState(false);
  const [serverError, setServerError] = useState(null);

  const formatDate = (dateString) => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    return isValid(date) ? format(date, 'PPpp') : '-';
  };

  useEffect(() => {
    fetchDiscoveries();
  }, []);

  const fetchDiscoveries = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/discoveries`);
      setDiscoveries(response.data);
    } catch (err) {
      setError(err.message);
      console.error('Error fetching discoveries:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleServerClick = async (serverId) => {
    setServerLoading(true);
    setServerError(null);
    try {
      const response = await axios.get(`${API_BASE_URL}/api/servers/${serverId}`);
      setSelectedServer(response.data);
    } catch (err) {
      setServerError(err.message);
      console.error('Error fetching server details:', err);
    } finally {
      setServerLoading(false);
    }
  };

  const handleClosePanel = () => {
    setSelectedServer(null);
    setServerError(null);
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box p={3}>
        <Alert severity="error">Error loading discoveries: {error}</Alert>
      </Box>
    );
  }

  return (
    <Container maxWidth="xl">
      <Box sx={{ mt: 4, mb: 4 }}>
        <Typography variant="h4" gutterBottom>
          Discovery History
        </Typography>
        
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>Discovery #</TableCell>
                <TableCell>Target Server</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Started At</TableCell>
                <TableCell>Completed At</TableCell>
                <TableCell>Error</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {discoveries.map((discovery) => (
                <TableRow key={discovery.id}>
                  <TableCell>{discovery.id}</TableCell>
                  <TableCell>
                    <Chip
                      label={`${discovery.server_id}${discovery.hostname ? ` (${discovery.hostname})` : ''}`}
                      onClick={() => handleServerClick(discovery.server_id)}
                      color="primary"
                      variant="outlined"
                      sx={{ cursor: 'pointer' }}
                    />
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={discovery.status}
                      color={
                        discovery.status === 'completed' ? 'success' :
                        discovery.status === 'failed' ? 'error' :
                        'default'
                      }
                    />
                  </TableCell>
                  <TableCell>{formatDate(discovery.started_at)}</TableCell>
                  <TableCell>{formatDate(discovery.completed_at)}</TableCell>
                  <TableCell>{discovery.error || '-'}</TableCell>
                </TableRow>
              ))}
              {discoveries.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} align="center">
                    No discoveries found
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TableContainer>
      </Box>

      {selectedServer && (
        <ServerDetailsPanel
          server={selectedServer}
          loading={serverLoading}
          error={serverError}
          onClose={handleClosePanel}
        />
      )}
    </Container>
  );
}

export default Discoveries; 