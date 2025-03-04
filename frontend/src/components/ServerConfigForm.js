import React, { useState } from 'react';
import {
  Box,
  Button,
  FormControl,
  FormControlLabel,
  FormLabel,
  Grid,
  MenuItem,
  Radio,
  RadioGroup,
  TextField,
  Typography,
  Paper,
} from '@mui/material';

const ServerConfigForm = ({ onSubmit, initialValues = {} }) => {
  const [formData, setFormData] = useState({
    hostname: '',
    port: '',
    username: '',
    password: '',
    osType: 'windows',
    region: '',
    useHttps: false,
    sshKeyPath: '',
    ...initialValues,
  });

  const handleChange = (event) => {
    const { name, value } = event.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleSubmit = (event) => {
    event.preventDefault();
    onSubmit(formData);
  };

  return (
    <Paper sx={{ p: 3 }}>
      <form onSubmit={handleSubmit}>
        <Grid container spacing={3}>
          <Grid item xs={12}>
            <Typography variant="h6" gutterBottom>
              Server Configuration
            </Typography>
          </Grid>

          <Grid item xs={12}>
            <FormControl component="fieldset">
              <FormLabel component="legend">Server Type</FormLabel>
              <RadioGroup
                row
                name="osType"
                value={formData.osType}
                onChange={handleChange}
              >
                <FormControlLabel
                  value="windows"
                  control={<Radio />}
                  label="Windows Server"
                />
                <FormControlLabel
                  value="linux"
                  control={<Radio />}
                  label="Linux Server"
                />
              </RadioGroup>
            </FormControl>
          </Grid>

          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Hostname"
              name="hostname"
              value={formData.hostname}
              onChange={handleChange}
              required
              helperText="Enter the server hostname or IP address"
            />
          </Grid>

          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Port"
              name="port"
              value={formData.port}
              onChange={handleChange}
              required
              type="number"
              helperText={formData.osType === 'linux' ? 'Default: 22 (SSH)' : 'Default: 5985 (WinRM)'}
            />
          </Grid>

          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Username"
              name="username"
              value={formData.username}
              onChange={handleChange}
              required
              helperText={formData.osType === 'linux' ? 'SSH username' : 'Windows username'}
            />
          </Grid>

          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Password"
              name="password"
              type="password"
              value={formData.password}
              onChange={handleChange}
              required={formData.osType === 'windows' || !formData.sshKeyPath}
              helperText={formData.osType === 'linux' ? 'Optional if using SSH key' : 'Windows password'}
            />
          </Grid>

          {formData.osType === 'linux' && (
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="SSH Key Path"
                name="sshKeyPath"
                value={formData.sshKeyPath}
                onChange={handleChange}
                helperText="Path to private SSH key file (optional)"
              />
            </Grid>
          )}

          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Region"
              name="region"
              value={formData.region}
              onChange={handleChange}
              select
            >
              <MenuItem value="us-east">US East</MenuItem>
              <MenuItem value="us-west">US West</MenuItem>
              <MenuItem value="eu-west">EU West</MenuItem>
              <MenuItem value="eu-central">EU Central</MenuItem>
              <MenuItem value="ap-southeast">AP Southeast</MenuItem>
            </TextField>
          </Grid>

          {formData.osType === 'windows' && (
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Radio
                    checked={formData.useHttps}
                    onChange={(e) => handleChange({
                      target: { name: 'useHttps', value: e.target.checked }
                    })}
                  />
                }
                label="Use HTTPS for WinRM connection"
              />
            </Grid>
          )}

          <Grid item xs={12}>
            <Box sx={{ display: 'flex', justifyContent: 'flex-end', mt: 2 }}>
              <Button
                type="submit"
                variant="contained"
                color="primary"
                size="large"
              >
                Save Server Configuration
              </Button>
            </Box>
          </Grid>
        </Grid>
      </form>
    </Paper>
  );
};

export default ServerConfigForm; 