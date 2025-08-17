#!/usr/bin/env node

// Test script to simulate frontend API calls

async function testPluginsAPI() {
  try {
    console.log('Testing plugins API endpoint...');
    console.log('URL: http://localhost:8080/api/v1/plugins/active');
    
    const response = await fetch('http://localhost:8080/api/v1/plugins/active', {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Origin': 'http://localhost:3001'
      },
      credentials: 'include'
    });
    
    console.log('Status:', response.status, response.statusText);
    console.log('CORS headers:', {
      'Access-Control-Allow-Origin': response.headers.get('access-control-allow-origin'),
      'Access-Control-Allow-Credentials': response.headers.get('access-control-allow-credentials'),
      'Access-Control-Allow-Methods': response.headers.get('access-control-allow-methods')
    });
    
    if (response.ok) {
      const data = await response.json();
      console.log('Success:', data.success);
      console.log('Plugins count:', data.plugins?.length || 0);
      if (data.plugins?.length > 0) {
        console.log('First plugin details:', {
          name: data.plugins[0].name,
          version: data.plugins[0].version,
          status: data.plugins[0].status,
          type: data.plugins[0].type
        });
      }
    } else {
      const text = await response.text();
      console.log('Error response:', text);
    }
  } catch (error) {
    console.error('Network error:', error.message);
  }
}

testPluginsAPI();