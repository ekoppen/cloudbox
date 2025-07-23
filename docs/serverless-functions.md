# Serverless Functions

CloudBox provides a powerful serverless function execution environment supporting multiple runtimes with secure Docker isolation.

## Overview

The function execution system supports:
- **Multi-Runtime Support** - Node.js, Python, and Go
- **Docker Isolation** - Secure containerized execution
- **Resource Management** - Memory limits and execution timeouts
- **Real-time Logs** - Comprehensive execution logging
- **Auto-scaling** - Automatic resource allocation
- **Native Fallback** - Direct execution when Docker unavailable

## Supported Runtimes

### Node.js (JavaScript)
- **Versions**: 18.x (default), 16.x, 14.x
- **Entry Point**: `index.handler` (default)
- **Package Manager**: npm/yarn support
- **Async/Await**: Full ES2022 support

### Python
- **Versions**: 3.9 (default), 3.8, 3.7
- **Entry Point**: `handler` function (default)
- **Package Manager**: pip support
- **Libraries**: Standard library + custom dependencies

### Go
- **Versions**: 1.19 (default), 1.18
- **Entry Point**: `main` function
- **Compilation**: Automatic compilation and execution
- **Performance**: Native binary execution

## Creating Functions

### Basic Function Structure

**JavaScript Example:**
```javascript
// Simple handler function
exports.handler = async (data, context) => {
  console.log('Received data:', data);
  
  return {
    message: 'Hello from CloudBox!',
    timestamp: new Date().toISOString(),
    input: data
  };
};

// With error handling
exports.handler = async (data, context) => {
  try {
    if (!data.name) {
      throw new Error('Name is required');
    }
    
    return {
      greeting: `Hello, ${data.name}!`,
      method: context.method,
      headers: context.headers
    };
  } catch (error) {
    console.error('Function error:', error.message);
    throw error;
  }
};
```

**Python Example:**
```python
import json
import datetime

def handler(data, context):
    """Main function handler"""
    print(f"Received data: {data}")
    
    return {
        'message': 'Hello from Python!',
        'timestamp': datetime.datetime.now().isoformat(),
        'input': data
    }

# With error handling
def handler(data, context):
    if 'name' not in data:
        raise ValueError('Name is required')
    
    return {
        'greeting': f"Hello, {data['name']}!",
        'method': context.get('method'),
        'headers': context.get('headers', {})
    }
```

**Go Example:**
```go
package main

import (
    "encoding/json"
    "fmt"
    "time"
)

type Response struct {
    Message   string      `json:"message"`
    Timestamp string      `json:"timestamp"`
    Input     interface{} `json:"input"`
}

func handler(data map[string]interface{}, context map[string]interface{}) (Response, error) {
    fmt.Printf("Received data: %+v\n", data)
    
    return Response{
        Message:   "Hello from Go!",
        Timestamp: time.Now().Format(time.RFC3339),
        Input:     data,
    }, nil
}
```

## Function Configuration

### Creating Functions via API

```bash
curl -X POST http://localhost:8080/api/projects/1/functions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "process-payment",
    "description": "Process payment transactions",
    "runtime": "nodejs18",
    "language": "javascript",
    "code": "exports.handler = async (data) => { return { status: \"success\" }; }",
    "entry_point": "handler",
    "timeout": 30,
    "memory": 128,
    "environment": {
      "NODE_ENV": "production",
      "API_KEY": "secret-key"
    },
    "dependencies": {
      "axios": "^1.0.0",
      "lodash": "^4.17.21"
    },
    "is_public": true
  }'
```

### JavaScript SDK

```javascript
const function = await cloudbox.functions.create({
  name: 'user-validator',
  description: 'Validate user data',
  runtime: 'nodejs18',
  language: 'javascript',
  code: `
    exports.handler = async (data) => {
      if (!data.email || !data.email.includes('@')) {
        throw new Error('Invalid email address');
      }
      return { valid: true, userId: data.id };
    };
  `,
  timeout: 10,
  memory: 64,
  isPublic: false
});
```

### Configuration Options

| Option | Description | Default | Range |
|--------|-------------|---------|--------|
| `timeout` | Execution timeout (seconds) | 30 | 1-300 |
| `memory` | Memory limit (MB) | 128 | 64-1024 |
| `runtime` | Runtime environment | nodejs18 | nodejs18, python3.9, go1.19 |
| `entry_point` | Function entry point | index.handler | Any valid function |
| `is_public` | Public API access | false | true/false |

## Executing Functions

### Direct Execution (Authenticated)

```bash
curl -X POST http://localhost:8080/api/projects/1/functions/123/execute \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "data": {
      "userId": "user123",
      "action": "process"
    },
    "headers": {
      "User-Agent": "MyApp/1.0"
    }
  }'
```

### Public Execution (No Auth Required)

```bash
# Execute by function name (if public)
curl -X POST http://localhost:8080/p/1/functions/process-payment \
  -H "Content-Type: application/json" \
  -d '{
    "amount": 99.99,
    "currency": "USD"
  }'

# Or with query parameters
curl -X GET "http://localhost:8080/p/1/functions/get-status?userId=123&format=json"
```

### JavaScript SDK

```javascript
// Authenticated execution
const result = await cloudbox.functions.execute('process-payment', {
  amount: 99.99,
  currency: 'USD',
  userId: 'user123'
});

console.log('Result:', result.response);
console.log('Execution time:', result.execution_time, 'ms');

// With timeout override
const result = await cloudbox.functions.execute('slow-function', data, {
  timeout: 60000 // 60 seconds
});
```

## Function Deployment

### Deployment Process

1. **Code Validation** - Syntax and entry point validation
2. **Dependency Installation** - Install packages and dependencies  
3. **Container Preparation** - Prepare Docker execution environment
4. **Function Registration** - Register function for execution
5. **Health Check** - Verify function is ready to receive requests

### Deployment via API

```bash
curl -X POST http://localhost:8080/api/projects/1/functions/123/deploy \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Monitoring Deployment

```javascript
// Deploy function
await cloudbox.functions.deploy(functionId);

// Monitor deployment status
const checkStatus = async () => {
  const func = await cloudbox.functions.get(functionId);
  
  switch (func.status) {
    case 'draft':
      console.log('Function not deployed');
      break;
    case 'building':
      console.log('Building function...');
      setTimeout(checkStatus, 2000);
      break;
    case 'deployed':
      console.log('Function deployed successfully!');
      break;
    case 'failed':
      console.error('Deployment failed:', func.build_logs);
      break;
  }
};

checkStatus();
```

## Execution Environment

### Container Specifications

**Security Features:**
- **Isolated Execution** - Each function runs in its own container
- **No Network Access** - Functions cannot access external networks
- **Limited File System** - Read-only file system with temp directory
- **Resource Limits** - CPU and memory limits enforced
- **Non-root User** - Functions run as unprivileged user

**Available Resources:**
- **File System** - `/workspace` working directory
- **Environment Variables** - Custom environment configuration
- **Standard I/O** - Console logging and error output
- **Temporary Storage** - Limited temp file access

### Context Object

Functions receive a context object with request information:

```javascript
{
  data: {}, // Request data/payload
  headers: {}, // HTTP headers
  method: "POST", // HTTP method
  path: "/p/1/functions/my-function", // Request path
  // Additional metadata
}
```

## Monitoring & Logs

### Execution Logs

```bash
# Get function execution logs
curl -X GET http://localhost:8080/api/projects/1/functions/123/logs?limit=100 \
  -H "Authorization: Bearer YOUR_TOKEN"
```

```javascript
// Get recent executions
const logs = await cloudbox.functions.getLogs(functionId, {
  limit: 50,
  since: '2024-01-15T00:00:00Z'
});

logs.forEach(execution => {
  console.log(`${execution.created_at}: ${execution.status}`);
  console.log(`Duration: ${execution.execution_time}ms`);
  console.log(`Memory: ${execution.memory_usage} bytes`);
  if (execution.logs) {
    console.log('Output:', execution.logs);
  }
});
```

### Performance Metrics

Each execution records:
- **Execution Time** - Function runtime in milliseconds
- **Memory Usage** - Peak memory consumption
- **Status** - Success/error status
- **Response Size** - Output data size
- **Error Details** - Stack traces and error messages

## Error Handling

### Common Error Types

**Timeout Errors:**
```json
{
  "error": "Function execution timed out after 30 seconds",
  "code": "TIMEOUT",
  "execution_time": 30000
}
```

**Runtime Errors:**
```json
{
  "error": "ReferenceError: undefined variable 'foo'",
  "code": "RUNTIME_ERROR",
  "logs": "Stack trace...",
  "execution_time": 156
}
```

**Deployment Errors:**
```json
{
  "error": "Failed to install dependencies",
  "code": "DEPLOYMENT_ERROR",
  "build_logs": "npm install failed..."
}
```

### Error Recovery

```javascript
try {
  const result = await cloudbox.functions.execute('my-function', data);
  return result.response;
} catch (error) {
  switch (error.code) {
    case 'TIMEOUT':
      console.log('Function timed out, try with smaller payload');
      break;
    case 'RUNTIME_ERROR':
      console.error('Function error:', error.message);
      break;
    case 'NOT_DEPLOYED':
      console.log('Function needs to be deployed first');
      break;
    default:
      console.error('Unexpected error:', error);
  }
}
```

## Best Practices

### Performance Optimization

1. **Minimize Cold Starts** - Keep functions lightweight
2. **Efficient Dependencies** - Only include necessary packages
3. **Connection Pooling** - Reuse database connections when possible
4. **Caching** - Cache frequently accessed data
5. **Memory Management** - Monitor and optimize memory usage

### Security Guidelines

1. **Input Validation** - Always validate input data
2. **Error Handling** - Don't expose sensitive information in errors
3. **Environment Variables** - Use env vars for secrets
4. **Principle of Least Privilege** - Request minimal permissions
5. **Regular Updates** - Keep dependencies updated

### Code Organization

```javascript
// Good: Modular and testable
const validateInput = (data) => {
  if (!data.email) throw new Error('Email required');
  if (!data.name) throw new Error('Name required');
};

const processUser = async (userData) => {
  // Business logic here
  return { id: 'user123', status: 'created' };
};

exports.handler = async (data, context) => {
  try {
    validateInput(data);
    const result = await processUser(data);
    return { success: true, data: result };
  } catch (error) {
    console.error('Handler error:', error.message);
    return { success: false, error: error.message };
  }
};
```

## Advanced Features

### Custom Docker Images (Future)

```yaml
# cloudbox-function.yml
runtime: custom
image: my-org/my-runtime:latest
build:
  dockerfile: ./Dockerfile
  context: ./
```

### Function Scaling (Future)

```javascript
await cloudbox.functions.updateScaling(functionId, {
  minInstances: 1,
  maxInstances: 10,
  targetConcurrency: 100
});
```

### Scheduled Functions (Future)

```javascript
await cloudbox.functions.addSchedule(functionId, {
  cron: '0 0 * * *', // Daily at midnight
  timezone: 'UTC',
  payload: { type: 'daily_cleanup' }
});
```