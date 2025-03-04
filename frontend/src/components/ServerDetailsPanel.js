import React from 'react';
import {
  Box,
  Typography,
  Paper,
  Grid,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListSubheader,
  Chip,
  IconButton,
  CircularProgress,
  Alert,
  Backdrop,
  Badge,
  LinearProgress,
} from '@mui/material';
import {
  Computer as ComputerIcon,
  Memory as MemoryIcon,
  Storage as StorageIcon,
  NetworkCheck as NetworkIcon,
  Apps as AppsIcon,
  Security as SecurityIcon,
  History as HistoryIcon,
  Close as CloseIcon,
  RadioButtonChecked as EstablishedIcon,
  RadioButtonUnchecked as ListenIcon,
  Speed as SpeedIcon,
} from '@mui/icons-material';
import { format, isValid } from 'date-fns';

function ServerDetailsPanel({ server, loading, error, onClose }) {
  console.log('ServerDetailsPanel rendered with:', { server, loading, error });
  
  if (!server) {
    console.log('No server data provided');
    return null;
  }

  const formatDate = (dateString) => {
    if (!dateString) return 'Unknown';
    const date = new Date(dateString);
    return isValid(date) ? format(date, 'yyyy-MM-dd HH:mm:ss') : 'Invalid date';
  };

  return (
    <>
      <Backdrop
        sx={{
          color: '#fff',
          zIndex: (theme) => theme.zIndex.drawer - 1,
        }}
        open={true}
        onClick={onClose}
      />
      <Box
        sx={{
          position: 'fixed',
          top: 0,
          right: 0,
          width: '40%',
          height: '100vh',
          bgcolor: 'background.paper',
          boxShadow: -3,
          overflow: 'auto',
          zIndex: 1200,
          transition: 'transform 0.3s ease-in-out',
          transform: 'translateX(0)',
        }}
      >
        <Box sx={{ p: 2, display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <Typography variant="h6">Server Details</Typography>
          <IconButton onClick={onClose} size="small">
            <CloseIcon />
          </IconButton>
        </Box>
        <Divider />

        {loading ? (
          <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
            <CircularProgress />
          </Box>
        ) : error ? (
          <Box p={3}>
            <Alert severity="error">Error loading server details: {error}</Alert>
          </Box>
        ) : (
          <Box sx={{ p: 2 }}>
            {/* Basic Information */}
            <Box sx={{ mb: 3 }}>
              <Typography variant="subtitle1" sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                <ComputerIcon sx={{ mr: 1 }} /> Basic Information
              </Typography>
              <Grid container spacing={2}>
                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">Hostname</Typography>
                  <Typography>{server.hostname || 'Unknown'}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">IP</Typography>
                  <Typography>
                    {server.ip || (server.ip_addresses && server.ip_addresses.length > 0 ? server.ip_addresses[0].ip_address : 'Unknown')}
                    {server.ip_addresses && server.ip_addresses.length > 1 && (
                      <Chip
                        label={`+${server.ip_addresses.length - 1} more`}
                        size="small"
                        color="info"
                        variant="outlined"
                        sx={{ ml: 1 }}
                      />
                    )}
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">OS Type</Typography>
                  <Typography>{server.os_type || 'Unknown'}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">Region</Typography>
                  <Typography>{server.region || 'Unknown'}</Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography variant="body2" color="text.secondary">Status</Typography>
                  <Chip
                    label={server.status || 'Unknown'}
                    color={server.status === 'active' ? 'success' : 'default'}
                    size="small"
                  />
                </Grid>
              </Grid>
            </Box>

            {/* Tags */}
            <Box sx={{ mb: 3 }}>
              <Typography variant="subtitle1" sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                <SecurityIcon sx={{ mr: 1 }} /> Tags
              </Typography>
              <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                {server.tags && server.tags.length > 0 ? (
                  server.tags.map((tag, index) => (
                    <Chip
                      key={`${tag.tag_name}-${index}`}
                      label={`${tag.tag_name}: ${tag.tag_value}`}
                      size="small"
                      color="primary"
                      variant="outlined"
                    />
                  ))
                ) : (
                  <Typography variant="body2" color="text.secondary">No tags available</Typography>
                )}
              </Box>
            </Box>

            {/* Hardware Information */}
            <Box sx={{ mb: 3 }}>
              <Typography variant="subtitle1" sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                <MemoryIcon sx={{ mr: 1 }} /> Hardware Information
              </Typography>
              <Grid container spacing={2}>
                {/* CPU Section */}
                <Grid item xs={12}>
                  <Paper variant="outlined" sx={{ p: 2 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
                      <Typography variant="body2" color="text.secondary">CPU</Typography>
                      {server.metrics?.cpu_usage && (
                        <Chip
                          label={`${server.metrics.cpu_usage.toFixed(1)}%`}
                          size="small"
                          color={server.metrics.cpu_usage > 80 ? 'error' : server.metrics.cpu_usage > 60 ? 'warning' : 'success'}
                          variant="outlined"
                        />
                      )}
                    </Box>
                    <Typography variant="body1" sx={{ mb: 1 }}>{server.cpu_model || 'Unknown'}</Typography>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Chip
                        label={`${server.cpu_count || 0} cores`}
                        size="small"
                        color="primary"
                        variant="outlined"
                      />
                      {server.metrics?.cpu_usage && (
                        <Box sx={{ flexGrow: 1 }}>
                          <LinearProgress 
                            variant="determinate" 
                            value={server.metrics.cpu_usage} 
                            color={server.metrics.cpu_usage > 80 ? 'error' : server.metrics.cpu_usage > 60 ? 'warning' : 'success'}
                            sx={{ height: 6, borderRadius: 3 }}
                          />
                        </Box>
                      )}
                    </Box>
                  </Paper>
                </Grid>

                {/* Memory Section */}
                <Grid item xs={12}>
                  <Paper variant="outlined" sx={{ p: 2 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
                      <Typography variant="body2" color="text.secondary">Memory</Typography>
                      {server.metrics?.memory_used && server.memory_total_gb && (
                        <Chip
                          label={`${((server.metrics.memory_used / (server.memory_total_gb * 1024 * 1024 * 1024)) * 100).toFixed(1)}%`}
                          size="small"
                          color={server.metrics.memory_used / (server.memory_total_gb * 1024 * 1024 * 1024) > 0.8 ? 'error' : server.metrics.memory_used / (server.memory_total_gb * 1024 * 1024 * 1024) > 0.6 ? 'warning' : 'success'}
                          variant="outlined"
                        />
                      )}
                    </Box>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                      <Typography variant="body1">
                        {server.memory_total_gb ? `${server.memory_total_gb.toFixed(1)} GB` : 'Unknown'}
                      </Typography>
                      {server.metrics?.memory_used && (
                        <Typography color="text.secondary">
                          ({((server.metrics.memory_used / (1024 * 1024 * 1024)).toFixed(1))} GB used)
                        </Typography>
                      )}
                    </Box>
                    {server.metrics?.memory_used && server.memory_total_gb && (
                      <LinearProgress 
                        variant="determinate" 
                        value={(server.metrics.memory_used / (server.memory_total_gb * 1024 * 1024 * 1024)) * 100} 
                        color={server.metrics.memory_used / (server.memory_total_gb * 1024 * 1024 * 1024) > 0.8 ? 'error' : server.metrics.memory_used / (server.memory_total_gb * 1024 * 1024 * 1024) > 0.6 ? 'warning' : 'success'}
                        sx={{ height: 6, borderRadius: 3 }}
                      />
                    )}
                  </Paper>
                </Grid>

                {/* Disk Section */}
                <Grid item xs={12}>
                  <Paper variant="outlined" sx={{ p: 2 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', mb: 1 }}>
                      <Typography variant="body2" color="text.secondary">Disk Space</Typography>
                      {server.disk_free_gb && server.disk_total_gb && (
                        <Chip
                          label={`${((server.disk_total_gb - server.disk_free_gb) / server.disk_total_gb * 100).toFixed(1)}%`}
                          size="small"
                          color={(server.disk_total_gb - server.disk_free_gb) / server.disk_total_gb > 0.8 ? 'error' : (server.disk_total_gb - server.disk_free_gb) / server.disk_total_gb > 0.6 ? 'warning' : 'success'}
                          variant="outlined"
                        />
                      )}
                    </Box>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 1 }}>
                      <Typography variant="body1">
                        {server.disk_total_gb ? `${server.disk_total_gb.toFixed(1)} GB` : 'Unknown'}
                      </Typography>
                      {server.disk_free_gb && (
                        <Typography color="text.secondary">
                          ({((server.disk_total_gb - server.disk_free_gb).toFixed(1))} GB used)
                        </Typography>
                      )}
                    </Box>
                    {server.disk_free_gb && server.disk_total_gb && (
                      <LinearProgress 
                        variant="determinate" 
                        value={((server.disk_total_gb - server.disk_free_gb) / server.disk_total_gb) * 100} 
                        color={(server.disk_total_gb - server.disk_free_gb) / server.disk_total_gb > 0.8 ? 'error' : (server.disk_total_gb - server.disk_free_gb) / server.disk_total_gb > 0.6 ? 'warning' : 'success'}
                        sx={{ height: 6, borderRadius: 3 }}
                      />
                    )}
                  </Paper>
                </Grid>

                {/* System Section */}
                <Grid item xs={12}>
                  <Paper variant="outlined" sx={{ p: 2 }}>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>System</Typography>
                    <Grid container spacing={2}>
                      <Grid item xs={12} sm={6}>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <HistoryIcon fontSize="small" color="action" />
                          <Typography variant="body2">
                            Last Boot: {formatDate(server.last_boot_time)}
                          </Typography>
                        </Box>
                      </Grid>
                      <Grid item xs={12} sm={6}>
                        {server.metrics?.load_average && (
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <SpeedIcon fontSize="small" color="action" />
                            <Typography variant="body2">
                              Load: {server.metrics.load_average.toFixed(2)}
                            </Typography>
                            <Chip
                              label={server.metrics.load_average > 4 ? 'High' : server.metrics.load_average > 2 ? 'Medium' : 'Low'}
                              size="small"
                              color={server.metrics.load_average > 4 ? 'error' : server.metrics.load_average > 2 ? 'warning' : 'success'}
                              variant="outlined"
                            />
                          </Box>
                        )}
                      </Grid>
                      <Grid item xs={12}>
                        {server.metrics?.process_count && (
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <AppsIcon fontSize="small" color="action" />
                            <Typography variant="body2">
                              {server.metrics.process_count} processes
                            </Typography>
                            <Chip
                              label={server.metrics.process_count > 400 ? 'High' : server.metrics.process_count > 200 ? 'Medium' : 'Low'}
                              size="small"
                              color={server.metrics.process_count > 400 ? 'error' : server.metrics.process_count > 200 ? 'warning' : 'success'}
                              variant="outlined"
                            />
                          </Box>
                        )}
                      </Grid>
                    </Grid>
                  </Paper>
                </Grid>
              </Grid>
            </Box>

            {/* Network Information */}
            <Box sx={{ mb: 3 }}>
              <Typography variant="subtitle1" sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                <NetworkIcon sx={{ mr: 1 }} /> Network Information
              </Typography>
              <List dense sx={{ bgcolor: 'background.default', borderRadius: 1 }}>
                <ListSubheader sx={{ 
                  bgcolor: 'background.paper', 
                  borderBottom: '1px solid',
                  borderColor: 'divider',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 1
                }}>
                  <NetworkIcon fontSize="small" />
                  <Typography variant="subtitle2">IP Addresses</Typography>
                  <Chip
                    label={`Total: ${[...new Set(server.ip_addresses?.map(ip => ip.ip_address))]?.length || 0}`}
                    size="small"
                    color="default"
                    sx={{ ml: 'auto' }}
                  />
                </ListSubheader>
                {server.ip_addresses && server.ip_addresses.length > 0 ? (
                  Array.from(new Set(server.ip_addresses.map(ip => JSON.stringify({ ip_address: ip.ip_address, interface_name: ip.interface_name }))))
                    .map(strIp => JSON.parse(strIp))
                    .map((ip, index) => (
                    <ListItem 
                      key={`${ip.ip_address}-${index}`}
                      sx={{ 
                        py: 1,
                        pl: 2,
                        borderLeft: '4px solid',
                        borderColor: 'info.main',
                        '&:hover': {
                          bgcolor: 'action.hover'
                        }
                      }}
                    >
                      <ListItemText
                        primary={
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Typography 
                              component="span" 
                              sx={{ 
                                fontFamily: 'monospace',
                                fontSize: '0.9rem',
                                fontWeight: 500
                              }}
                            >
                              {ip.ip_address}
                            </Typography>
                            <Chip
                              label={ip.interface_name || 'Unknown'}
                              size="small"
                              color="info"
                              variant="outlined"
                              sx={{ 
                                height: 20,
                                '& .MuiChip-label': {
                                  px: 1,
                                  fontSize: '0.7rem'
                                }
                              }}
                            />
                          </Box>
                        }
                      />
                    </ListItem>
                  ))
                ) : (
                  <ListItem>
                    <ListItemText primary="No IP addresses found" />
                  </ListItem>
                )}
              </List>
            </Box>

            {/* Services */}
            <Box sx={{ mb: 3 }}>
              <Typography variant="subtitle1" sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                <AppsIcon sx={{ mr: 1 }} /> Services
              </Typography>
              <List dense sx={{ bgcolor: 'background.default', borderRadius: 1 }}>
                <ListSubheader sx={{ 
                  bgcolor: 'background.paper', 
                  borderBottom: '1px solid',
                  borderColor: 'divider',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 1
                }}>
                  <AppsIcon fontSize="small" />
                  <Typography variant="subtitle2">Running Services</Typography>
                  <Chip
                    label={`Total: ${[...new Set(server.services?.map(service => service.name))]?.length || 0}`}
                    size="small"
                    color="default"
                    sx={{ ml: 'auto' }}
                  />
                </ListSubheader>
                {server.services && server.services.length > 0 ? (
                  Array.from(new Set(server.services.map(service => JSON.stringify({ name: service.name, status: service.status }))))
                    .map(strService => JSON.parse(strService))
                    .map((service, index) => (
                    <ListItem 
                      key={`${service.name}-${index}`}
                      sx={{ 
                        py: 1,
                        pl: 2,
                        borderLeft: '4px solid',
                        borderColor: service.status === 'running' ? 'success.main' : 'error.main',
                        '&:hover': {
                          bgcolor: 'action.hover'
                        }
                      }}
                    >
                      <ListItemText
                        primary={
                          <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                            <Typography 
                              component="span" 
                              sx={{ 
                                fontSize: '0.9rem',
                                fontWeight: 500
                              }}
                            >
                              {service.name}
                            </Typography>
                            <Chip
                              label={service.status}
                              size="small"
                              color={service.status === 'running' ? 'success' : 'error'}
                              sx={{ 
                                height: 20,
                                '& .MuiChip-label': {
                                  px: 1,
                                  fontSize: '0.7rem',
                                  textTransform: 'capitalize'
                                }
                              }}
                            />
                          </Box>
                        }
                      />
                    </ListItem>
                  ))
                ) : (
                  <ListItem>
                    <ListItemText primary="No services found" />
                  </ListItem>
                )}
              </List>
            </Box>

            {/* Security */}
            <Box sx={{ mb: 3 }}>
              <Typography variant="subtitle1" sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                <SecurityIcon sx={{ mr: 1 }} /> Security
              </Typography>
              <List dense sx={{ bgcolor: 'background.default', borderRadius: 1 }}>
                <ListSubheader sx={{ 
                  bgcolor: 'background.paper', 
                  borderBottom: '1px solid',
                  borderColor: 'divider',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 1
                }}>
                  <SecurityIcon fontSize="small" />
                  <Typography variant="subtitle2">Open Ports</Typography>
                  <Chip
                    label={`Total: ${server.open_ports?.length || 0}`}
                    size="small"
                    color="default"
                    sx={{ ml: 'auto' }}
                  />
                </ListSubheader>
                {server.open_ports && server.open_ports.length > 0 ? (
                  <>
                    {/* Listening Ports */}
                    <ListSubheader 
                      sx={{ 
                        bgcolor: 'warning.lighter', 
                        display: 'flex', 
                        alignItems: 'center', 
                        gap: 1,
                        pl: 2
                      }}
                    >
                      <ListenIcon fontSize="small" color="warning" />
                      <Typography variant="subtitle2">Listening Ports</Typography>
                      <Chip
                        label={server.open_ports.filter(port => port.state === 'LISTEN').length}
                        size="small"
                        color="warning"
                        sx={{ ml: 'auto', bgcolor: 'warning.main', color: 'warning.contrastText' }}
                      />
                    </ListSubheader>
                    {server.open_ports
                      .filter(port => port.state === 'LISTEN')
                      .map((port, index) => (
                        <ListItem 
                          key={`listen-${index}`} 
                          sx={{ 
                            py: 1,
                            pl: 2,
                            borderLeft: '4px solid',
                            borderColor: 'warning.main',
                            '&:hover': {
                              bgcolor: 'action.hover'
                            }
                          }}
                        >
                          <ListItemText
                            primary={
                              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <Badge
                                  badgeContent={port.local_port}
                                  color="warning"
                                  sx={{ 
                                    mr: 1,
                                    '& .MuiBadge-badge': {
                                      fontWeight: 'bold'
                                    }
                                  }}
                                >
                                  <Box sx={{ width: 24 }} />
                                </Badge>
                                <Typography 
                                  component="span" 
                                  sx={{ 
                                    fontFamily: 'monospace',
                                    fontSize: '0.9rem',
                                    fontWeight: 500
                                  }}
                                >
                                  {port.local_ip}
                                </Typography>
                              </Box>
                            }
                            secondary={
                              <Box sx={{ mt: 0.5, display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                                <Typography 
                                  variant="body2" 
                                  color="text.secondary"
                                  sx={{ 
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 0.5,
                                    fontSize: '0.8rem'
                                  }}
                                >
                                  {port.description || 'Unknown service'}
                                </Typography>
                                {port.process_name && (
                                  <Typography 
                                    variant="body2" 
                                    color="text.secondary"
                                    sx={{ 
                                      display: 'flex',
                                      alignItems: 'center',
                                      gap: 0.5,
                                      fontSize: '0.8rem'
                                    }}
                                  >
                                    <span style={{ opacity: 0.7 }}>Process:</span>
                                    {port.process_name}
                                    {port.process_id && (
                                      <Chip 
                                        label={`PID: ${port.process_id}`}
                                        size="small"
                                        variant="outlined"
                                        sx={{ 
                                          height: 20,
                                          '& .MuiChip-label': {
                                            px: 1,
                                            fontSize: '0.7rem'
                                          }
                                        }}
                                      />
                                    )}
                                  </Typography>
                                )}
                              </Box>
                            }
                          />
                        </ListItem>
                      ))}

                    {/* Established Connections */}
                    <ListSubheader 
                      sx={{ 
                        bgcolor: 'success.lighter',
                        display: 'flex', 
                        alignItems: 'center', 
                        gap: 1,
                        mt: 2,
                        pl: 2
                      }}
                    >
                      <EstablishedIcon fontSize="small" color="success" />
                      <Typography variant="subtitle2">Established Connections</Typography>
                      <Chip
                        label={[...new Set(server.open_ports
                          .filter(port => port.state === 'ESTABLISHED')
                          .map(port => JSON.stringify({
                            local_port: port.local_port,
                            local_ip: port.local_ip,
                            remote_port: port.remote_port,
                            remote_ip: port.remote_ip
                          })))].length}
                        size="small"
                        color="success"
                        sx={{ ml: 'auto', bgcolor: 'success.main', color: 'success.contrastText' }}
                      />
                    </ListSubheader>
                    {Array.from(new Set(server.open_ports
                      .filter(port => port.state === 'ESTABLISHED')
                      .map(port => JSON.stringify({
                        local_port: port.local_port,
                        local_ip: port.local_ip,
                        remote_port: port.remote_port,
                        remote_ip: port.remote_ip,
                        description: port.description,
                        process_name: port.process_name,
                        process_id: port.process_id,
                        state: port.state
                      }))))
                      .map(strPort => JSON.parse(strPort))
                      .map((port, index) => (
                        <ListItem 
                          key={`established-${port.local_port}-${port.remote_port}-${index}`} 
                          sx={{ 
                            py: 1,
                            pl: 2,
                            borderLeft: '4px solid',
                            borderColor: 'success.main',
                            '&:hover': {
                              bgcolor: 'action.hover'
                            }
                          }}
                        >
                          <ListItemText
                            primary={
                              <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                                <Badge
                                  badgeContent={port.local_port}
                                  color="success"
                                  sx={{ 
                                    mr: 1,
                                    '& .MuiBadge-badge': {
                                      fontWeight: 'bold'
                                    }
                                  }}
                                >
                                  <Box sx={{ width: 24 }} />
                                </Badge>
                                <Typography 
                                  component="span" 
                                  sx={{ 
                                    fontFamily: 'monospace',
                                    fontSize: '0.9rem',
                                    fontWeight: 500
                                  }}
                                >
                                  {port.local_ip}
                                </Typography>
                                <Typography 
                                  component="span" 
                                  color="text.secondary" 
                                  sx={{ 
                                    fontFamily: 'monospace',
                                    fontSize: '0.9rem'
                                  }}
                                >
                                  → {port.remote_ip}:{port.remote_port}
                                </Typography>
                              </Box>
                            }
                            secondary={
                              <Box sx={{ mt: 0.5, display: 'flex', flexDirection: 'column', gap: 0.5 }}>
                                <Typography 
                                  variant="body2" 
                                  color="text.secondary"
                                  sx={{ 
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 0.5,
                                    fontSize: '0.8rem'
                                  }}
                                >
                                  {port.description || 'Unknown service'}
                                </Typography>
                                {port.process_name && (
                                  <Typography 
                                    variant="body2" 
                                    color="text.secondary"
                                    sx={{ 
                                      display: 'flex',
                                      alignItems: 'center',
                                      gap: 0.5,
                                      fontSize: '0.8rem'
                                    }}
                                  >
                                    <span style={{ opacity: 0.7 }}>Process:</span>
                                    {port.process_name}
                                    {port.process_id && (
                                      <Chip 
                                        label={`PID: ${port.process_id}`}
                                        size="small"
                                        variant="outlined"
                                        sx={{ 
                                          height: 20,
                                          '& .MuiChip-label': {
                                            px: 1,
                                            fontSize: '0.7rem'
                                          }
                                        }}
                                      />
                                    )}
                                  </Typography>
                                )}
                              </Box>
                            }
                          />
                        </ListItem>
                      ))}
                  </>
                ) : (
                  <ListItem>
                    <ListItemText primary="No open ports found" />
                  </ListItem>
                )}
              </List>
            </Box>

            {/* Installed Software */}
            <Box sx={{ mb: 3 }}>
              <Typography variant="subtitle1" sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                <AppsIcon sx={{ mr: 1 }} /> Installed Software
              </Typography>
              <List dense>
                {server.installed_software && server.installed_software.length > 0 ? (
                  server.installed_software.map((software, index) => (
                    <ListItem key={index}>
                      <ListItemText
                        primary={software.name}
                        secondary={`Version: ${software.version}`}
                      />
                    </ListItem>
                  ))
                ) : (
                  <ListItem>
                    <ListItemText primary="No installed software found" />
                  </ListItem>
                )}
              </List>
            </Box>
          </Box>
        )}
      </Box>
    </>
  );
}

export default ServerDetailsPanel; 