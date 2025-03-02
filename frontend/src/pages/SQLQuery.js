import React, { useState } from 'react';
import {
  Box,
  Typography,
  Paper,
  TextField,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  CircularProgress,
  Alert,
  Divider,
} from '@mui/material';
import PlayArrowIcon from '@mui/icons-material/PlayArrow';

function SQLQuery() {
  const [query, setQuery] = useState('SELECT * FROM server_discovery.servers LIMIT 10');
  const [results, setResults] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const handleQueryChange = (event) => {
    setQuery(event.target.value);
  };

  const executeQuery = () => {
    setLoading(true);
    setError(null);
    setResults(null);

    fetch('/api/query', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ query }),
    })
      .then(response => {
        if (!response.ok) {
          return response.text().then(text => {
            throw new Error(text);
          });
        }
        return response.json();
      })
      .then(data => {
        setResults(data);
        setLoading(false);
      })
      .catch(err => {
        setError(err.message);
        setLoading(false);
      });
  };

  // Helper function to get all unique columns from results
  const getColumns = (results) => {
    if (!results || results.length === 0) return [];
    const columns = new Set();
    results.forEach(row => {
      Object.keys(row).forEach(key => columns.add(key));
    });
    return Array.from(columns);
  };

  const columns = results ? getColumns(results) : [];

  return (
    <Box>
      <Typography variant="h4" gutterBottom>
        SQL Query
      </Typography>
      
      <Paper sx={{ p: 2, mb: 3 }}>
        <Typography variant="subtitle1" gutterBottom>
          Enter your SQL query below:
        </Typography>
        <TextField
          fullWidth
          multiline
          rows={5}
          variant="outlined"
          value={query}
          onChange={handleQueryChange}
          sx={{ mb: 2, fontFamily: 'monospace' }}
        />
        <Button
          variant="contained"
          color="primary"
          onClick={executeQuery}
          disabled={loading}
          startIcon={loading ? <CircularProgress size={20} /> : <PlayArrowIcon />}
        >
          Execute Query
        </Button>
      </Paper>

      {error && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      )}

      {results && (
        <Paper>
          <Box sx={{ p: 2 }}>
            <Typography variant="h6">
              Results ({results.length} rows)
            </Typography>
          </Box>
          <Divider />
          <TableContainer sx={{ maxHeight: 500 }}>
            <Table stickyHeader>
              <TableHead>
                <TableRow>
                  {columns.map(column => (
                    <TableCell key={column}>{column}</TableCell>
                  ))}
                </TableRow>
              </TableHead>
              <TableBody>
                {results.map((row, rowIndex) => (
                  <TableRow key={rowIndex} hover>
                    {columns.map(column => (
                      <TableCell key={`${rowIndex}-${column}`}>
                        {row[column] !== null ? 
                          (typeof row[column] === 'object' ? 
                            JSON.stringify(row[column]) : 
                            String(row[column])
                          ) : 
                          'NULL'
                        }
                      </TableCell>
                    ))}
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Paper>
      )}
    </Box>
  );
}

export default SQLQuery; 