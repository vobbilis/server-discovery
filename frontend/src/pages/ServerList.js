import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import axios from 'axios';
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
  Container,
  IconButton,
  Tooltip,
  Snackbar,
  Alert,
  Grid,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  TablePagination,
  Button,
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import CheckCircleIcon from '@mui/icons-material/CheckCircle';
import ErrorIcon from '@mui/icons-material/Error';
import ClearIcon from '@mui/icons-material/Clear';
import { format } from 'date-fns';
import VisibilityIcon from '@mui/icons-material/Visibility';
import RefreshIcon from '@mui/icons-material/Refresh';
import SvgIcon from '@mui/material/SvgIcon';
import ServerDetailsPanel from '../components/ServerDetailsPanel';

const ServerTypeIcon = ({ osType }) => {
  if (osType === 'linux') {
    return <SvgIcon><path d="M21 10.12h-6.78l2.74-2.82c-2.73-2.7-7.15-2.8-9.88-.1-2.73 2.71-2.73 7.08 0 9.79s7.15 2.71 9.88 0C18.32 15.65 19 14.08 19 12.1h2c0 1.98-.88 4.55-2.64 6.29-3.51 3.48-9.21 3.48-12.72 0-3.5-3.47-3.53-9.11-.02-12.58s9.14-3.47 12.65 0L21 3v7.12zM12.5 8v4.25l3.5 2.08-.72 1.21L11 13V8h1.5z"/></SvgIcon>;
  }
  return <SvgIcon><path d="M3 5a2 2 0 012-2h14a2 2 0 012 2v14a2 2 0 01-2 2H5a2 2 0 01-2-2V5zm7 14h4v-4h4V5H5v10h4v4z"/></SvgIcon>;
};

function ServerList() {
  const [servers, setServers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'info' });
  const [filters, setFilters] = useState({
    status: 'all',
    osType: 'all',
    region: 'all'
  });
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [selectedServer, setSelectedServer] = useState(null);
  const [detailsOpen, setDetailsOpen] = useState(false);
  const [serverLoading, setServerLoading] = useState(false);
  const [serverError, setServerError] = useState(null);

  console.log('ServerList: Component rendering', { loading, error, serversCount: servers.length });

  useEffect(() => {
    console.log('ServerList: Initializing component');
    fetchServers();
  }, []);

  const fetchServers = async () => {
    console.log('ServerList: Fetching servers...');
    setLoading(true);
    try {
      const response = await axios.get('/api/servers', {
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        }
      });
      console.log('ServerList: Server response received', { 
        status: response.status,
        statusText: response.statusText,
        headers: response.headers,
        data: response.data
      });
      
      setServers(response.data || []);
      setLoading(false);
    } catch (err) {
      console.error('ServerList: Error fetching servers:', {
        message: err.message,
        response: err.response,
        request: err.request,
        config: err.config
      });
      setError(err.response?.data?.message || err.message);
      setLoading(false);
    }
  };

  const handleRunDiscovery = async (serverId) => {
    try {
      const response = await fetch(`/api/servers/${serverId}/discoveries`, {
        method: 'POST',
      });

      if (!response.ok) {
        throw new Error('Failed to start discovery');
      }

      const data = await response.json();
      setSnackbar({
        open: true,
        message: 'Discovery started successfully',
        severity: 'success',
      });

      // Refresh the server list after a short delay
      setTimeout(fetchServers, 2000);
    } catch (err) {
      setSnackbar({
        open: true,
        message: `Error starting discovery: ${err.message}`,
        severity: 'error',
      });
    }
  };

  const handleSearchChange = (event) => {
    setSearchTerm(event.target.value);
    setPage(0); // Reset to first page when searching
  };

  const handleCloseSnackbar = () => {
    setSnackbar(prev => ({ ...prev, open: false }));
  };

  const handleFilterChange = (filterType, value) => {
    setFilters(prev => ({
      ...prev,
      [filterType]: value
    }));
    setPage(0); // Reset to first page when filtering
  };

  const handleChangePage = (event, newPage) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
  };

  const handleClearFilters = () => {
    setSearchTerm('');
    setFilters({
      status: 'all',
      osType: 'all',
      region: 'all'
    });
    setPage(0);
  };

  const filteredServers = servers.filter(server => {
    const searchLower = searchTerm.toLowerCase();
    console.log('ServerList: Filtering servers', { 
      searchTerm, 
      filters,
      totalServers: servers.length 
    });
    
    const matchesSearch = 
      server.hostname?.toLowerCase().includes(searchLower) ||
      server.ip?.toLowerCase().includes(searchLower) ||
      server.region?.toLowerCase().includes(searchLower) ||
      server.os_type?.toLowerCase().includes(searchLower) ||
      server.os_name?.toLowerCase().includes(searchLower) ||
      server.os_version?.toLowerCase().includes(searchLower) ||
      (server.tags || []).some(tag => 
        tag.key?.toLowerCase().includes(searchLower) || 
        tag.value?.toLowerCase().includes(searchLower)
      );

    const matchesStatus = filters.status === 'all' || server.status === filters.status;
    const matchesOsType = filters.osType === 'all' || server.os_type === filters.osType;
    const matchesRegion = filters.region === 'all' || server.region === filters.region;

    return matchesSearch && matchesStatus && matchesOsType && matchesRegion;
  });

  // Get current page of servers
  const paginatedServers = filteredServers.slice(
    page * rowsPerPage,
    page * rowsPerPage + rowsPerPage
  );

  console.log('ServerList: Pagination results', {
    totalServers: servers.length,
    filteredCount: filteredServers.length,
    currentPageCount: paginatedServers.length,
    page,
    rowsPerPage
  });

  // Get unique values for filters
  const uniqueStatuses = [...new Set(servers.map(server => server.status))];
  const uniqueOsTypes = [...new Set(servers.map(server => server.os_type))];
  const uniqueRegions = [...new Set(servers.map(server => server.region))];

  const handleHostnameClick = async (server) => {
    setServerLoading(true);
    setServerError(null);
    setDetailsOpen(true);
    try {
      const response = await axios.get(`/api/servers/${server.id}`);
      setSelectedServer(response.data);
    } catch (err) {
      console.error('Error fetching server details:', err);
      setServerError(err.message);
      setSelectedServer(server); // Fallback to basic server info
    } finally {
      setServerLoading(false);
    }
  };

  const handleCloseDetails = () => {
    setDetailsOpen(false);
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
    <Container maxWidth="xl">
      <Box sx={{ mt: 4, mb: 4 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Typography variant="h4">
            Servers
          </Typography>
          <Button
            variant="outlined"
            startIcon={<ClearIcon />}
            onClick={handleClearFilters}
            disabled={!searchTerm && Object.values(filters).every(v => v === 'all')}
          >
            Clear Filters
          </Button>
        </Box>
        
        <Paper sx={{ p: 2, mb: 3 }}>
          <Grid container spacing={2}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                variant="outlined"
                placeholder="Search servers by hostname, IP, OS, region, or tags..."
                value={searchTerm}
                onChange={handleSearchChange}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <SearchIcon />
                    </InputAdornment>
                  ),
                  endAdornment: searchTerm && (
                    <InputAdornment position="end">
                      <IconButton
                        size="small"
                        onClick={() => setSearchTerm('')}
                        title="Clear search"
                      >
                        <ClearIcon />
                      </IconButton>
                    </InputAdornment>
                  ),
                }}
              />
            </Grid>
            <Grid item xs={12} md={4}>
              <FormControl fullWidth>
                <InputLabel>Status</InputLabel>
                <Select
                  value={filters.status}
                  onChange={(e) => handleFilterChange('status', e.target.value)}
                  label="Status"
                >
                  <MenuItem value="all">All Statuses</MenuItem>
                  {uniqueStatuses.map(status => (
                    <MenuItem key={status} value={status}>
                      {status}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} md={4}>
              <FormControl fullWidth>
                <InputLabel>OS Type</InputLabel>
                <Select
                  value={filters.osType}
                  onChange={(e) => handleFilterChange('osType', e.target.value)}
                  label="OS Type"
                >
                  <MenuItem value="all">All OS Types</MenuItem>
                  {uniqueOsTypes.map(osType => (
                    <MenuItem key={osType} value={osType}>
                      {osType}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} md={4}>
              <FormControl fullWidth>
                <InputLabel>Region</InputLabel>
                <Select
                  value={filters.region}
                  onChange={(e) => handleFilterChange('region', e.target.value)}
                  label="Region"
                >
                  <MenuItem value="all">All Regions</MenuItem>
                  {uniqueRegions.map(region => (
                    <MenuItem key={region} value={region}>
                      {region}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12}>
              <Typography variant="body2" color="textSecondary">
                Showing {paginatedServers.length} of {filteredServers.length} filtered servers (Total: {servers.length})
              </Typography>
            </Grid>
          </Grid>
        </Paper>
        
        <Paper>
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Type</TableCell>
                  <TableCell>Hostname</TableCell>
                  <TableCell>Region</TableCell>
                  <TableCell>Port</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell>Last Discovery</TableCell>
                  <TableCell>Actions</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {paginatedServers.map(server => (
                  <TableRow key={server.id}>
                    <TableCell>
                      <Tooltip title={server.osType === 'linux' ? 'Linux Server' : 'Windows Server'}>
                        <Box sx={{ display: 'flex', alignItems: 'center' }}>
                          <ServerTypeIcon osType={server.osType} />
                        </Box>
                      </Tooltip>
                    </TableCell>
                    <TableCell 
                      onClick={() => handleHostnameClick(server)}
                      sx={{ 
                        cursor: 'pointer',
                        '&:hover': {
                          color: 'primary.main',
                          textDecoration: 'underline'
                        }
                      }}
                    >
                      {server.hostname}
                    </TableCell>
                    <TableCell>{server.region || 'Unknown'}</TableCell>
                    <TableCell>{server.port}</TableCell>
                    <TableCell>
                      <Chip
                        label={server.status}
                        color={server.status === 'active' ? 'success' : 'error'}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      {server.lastChecked ? 
                        format(new Date(server.lastChecked), 'yyyy-MM-dd HH:mm') :
                        'Never'
                      }
                    </TableCell>
                    <TableCell>
                      <IconButton
                        component={Link}
                        to={`/servers/${server.id}`}
                        size="small"
                        title="View Details"
                      >
                        <VisibilityIcon />
                      </IconButton>
                      <IconButton
                        onClick={() => handleRunDiscovery(server.id)}
                        size="small"
                        title="Run Discovery"
                      >
                        <RefreshIcon />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                ))}
                {paginatedServers.length === 0 && (
                  <TableRow>
                    <TableCell colSpan={7} align="center">
                      No servers match the current filters
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
          <TablePagination
            rowsPerPageOptions={[5, 10, 25, 50, 100]}
            component="div"
            count={filteredServers.length}
            rowsPerPage={rowsPerPage}
            page={page}
            onPageChange={handleChangePage}
            onRowsPerPageChange={handleChangeRowsPerPage}
          />
        </Paper>
      </Box>

      <Snackbar
        open={snackbar.open}
        autoHideDuration={6000}
        onClose={handleCloseSnackbar}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
      >
        <Alert onClose={handleCloseSnackbar} severity={snackbar.severity}>
          {snackbar.message}
        </Alert>
      </Snackbar>

      {detailsOpen && selectedServer && (
        <ServerDetailsPanel
          server={selectedServer}
          loading={serverLoading}
          error={serverError}
          onClose={handleCloseDetails}
        />
      )}
    </Container>
  );
}

export default ServerList; 