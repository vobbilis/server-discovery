import React from 'react';
import { Box, Typography, Paper } from '@mui/material';

function TestContent() {
  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        Test Content
      </Typography>
      <Paper sx={{ p: 2, mb: 2 }}>
        <Typography variant="body1">
          This is a test component to verify that content is rendering correctly.
        </Typography>
      </Paper>
      {Array.from({ length: 20 }).map((_, index) => (
        <Paper key={index} sx={{ p: 2, mb: 2 }}>
          <Typography variant="h6">Item {index + 1}</Typography>
          <Typography variant="body1">
            Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam in dui mauris.
          </Typography>
        </Paper>
      ))}
    </Box>
  );
}

export default TestContent; 