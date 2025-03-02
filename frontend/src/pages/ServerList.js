import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
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
  Chip,
  CircularProgress,
  TextField,
  InputAdornment,
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import { format } from 'date-fns';

function ServerList() {
  const [servers, setServers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');

  useEffect(() => {
    fetch('/api/servers')
      .then(response => {
        if (!response.ok) {
          throw new Error('Failed to fetch servers');
        }
        return response.json();
      })
      .then(data => {
        setServers(data);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  }, []);

  const handleSearchChange = (event) => {
    setSearchTerm(event.target.value);
  };

  const filteredServers = servers.filter(server => {
    const searchLower = searchTerm.toLowerCase();
    return (
      server.hostname.toLowerCase().includes(searchLower) ||
      server.region.toLowerCase().includes(searchLower) ||
      server.tags.some(tag => 
        tag.key.toLowerCase().includes(searchLower) || 
        tag.value.toLowerCase().includes(searchLower)
      )
    );
  });

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
        Servers
      </Typography>
      
      <Box sx={{ mb: 3 }}>
        <TextField
          fullWidth
          variant="outlined"
          placeholder="Search servers by hostname, region, or tags..."
          value={searchTerm}
          onChange={handleSearchChange}
          InputProps={{
            startAdornment: (
              <InputAdornment position="start">
                <SearchIcon />
              </InputAdornment>
            ),
          }}
        />
      </Box>
      
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Hostname</TableCell>
              <TableCell>Port</TableCell>
              <TableCell>Region</TableCell>
              <TableCell>Tags</TableCell>
              <TableCell>Discoveries</TableCell>
              <TableCell>Last Discovery</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredServers.map(server => (
              <TableRow key={server.id} hover>
                <TableCell>
                  <Link to={`/servers/${server.id}`} style={{ textDecoration: 'none', color: '#90caf9' }}>
                    {server.hostname}
                  </Link>
                </TableCell>
                <TableCell>{server.port}</TableCell>
                <TableCell>{server.region || 'Unknown'}</TableCell>
                <TableCell>
                  {server.tags.map(tag => (
                    <Chip 
                      key={tag.key} 
                      label={`${tag.key}: ${tag.value}`} 
                      size="small" 
                      sx={{ mr: 0.5, mb: 0.5 }} 
                    />
                  ))}
                </TableCell>
                <TableCell>{server.discovery_count}</TableCell>
                <TableCell>
                  {server.last_discovery ? 
                    format(new Date(server.last_discovery), 'yyyy-MM-dd HH:mm') : 
                    'Never'
                  }
                </TableCell>
              </TableRow>
            ))}
            {filteredServers.length === 0 && (
              <TableRow>
                <TableCell colSpan={6} align="center">
                  No servers found matching your search criteria
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}

export default ServerList; 