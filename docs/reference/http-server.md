# HTTP Server Configuration

The Loom HTTP server provides REST API, Server-Sent Events (SSE) streaming, CORS support, and Swagger UI documentation.

## Quick Start

Start the server with HTTP endpoints enabled:

```bash
# Default HTTP port: 5006
looms serve --http-port 5006

# Disable HTTP (gRPC only)
looms serve --http-port 0
```

## Endpoints

### Core Endpoints

- **Health Check**: `GET /health`
- **Swagger UI**: `GET /swagger-ui`
- **OpenAPI Spec**: `GET /openapi.json`
- **SSE Streaming**: `POST /v1/weave:stream`

### API Endpoints

All gRPC endpoints are available via HTTP/REST:
- `POST /v1/weave` - Execute agent query
- `POST /v1/sessions` - Create session
- `GET /v1/sessions/{session_id}` - Get session
- `POST /v1/weave:stream` - Stream agent execution (SSE)

See `/swagger-ui` for complete API documentation.

## CORS Configuration

### Development (Default)

By default, CORS is permissive for developer experience:

```yaml
server:
  http_port: 5006
  cors:
    enabled: true
    allowed_origins: ["*"]  # ⚠️ INSECURE for production!
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"]
    allowed_headers: ["*"]
    exposed_headers: ["Content-Length", "Content-Type"]
    allow_credentials: false
    max_age: 86400  # 24 hours
```

**⚠️ Security Warning**: Wildcard origins allow any website to access your API. The server logs a warning on startup.

### Production Configuration

For production deployments, **always specify allowed origins**:

```yaml
server:
  http_port: 5006
  cors:
    enabled: true
    # ✅ Specify exact domains
    allowed_origins:
      - "https://yourdomain.com"
      - "https://app.yourdomain.com"
      - "https://www.yourdomain.com"

    # ✅ Restrict methods to what's needed
    allowed_methods:
      - "GET"
      - "POST"
      - "PUT"
      - "DELETE"
      - "OPTIONS"

    # ✅ Specify required headers only
    allowed_headers:
      - "Content-Type"
      - "Authorization"
      - "X-Request-ID"

    exposed_headers:
      - "Content-Length"
      - "Content-Type"
      - "X-Request-ID"

    # ✅ Enable credentials for authenticated APIs
    allow_credentials: true

    max_age: 86400
```

**Example config**: [`examples/config/looms-production-cors.yaml`](../../examples/config/looms-production-cors.yaml)

### CORS Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable/disable CORS |
| `allowed_origins` | []string | `["*"]` | Allowed origin domains (wildcard or specific) |
| `allowed_methods` | []string | `["GET", "POST", ...]` | Allowed HTTP methods |
| `allowed_headers` | []string | `["*"]` | Allowed request headers |
| `exposed_headers` | []string | `["Content-Length", ...]` | Headers exposed to browser |
| `allow_credentials` | bool | `false` | Allow credentials (cookies, auth headers) |
| `max_age` | int | `86400` | Preflight cache duration (seconds) |

### Security Validations

The server enforces security rules:

1. **Wildcard + Credentials Blocked**: Cannot use `allowed_origins: ["*"]` with `allow_credentials: true` (browsers reject this)
2. **Startup Warnings**: Wildcard origins trigger warnings in logs
3. **Production Detection**: Server checks for production indicators

### Environment Variables

Override CORS settings via environment:

```bash
# Set via config file
export LOOM_CONFIG=/path/to/config.yaml

# Or configure in looms.yaml
```

## Swagger UI

### Accessing Documentation

Open your browser to:
```
http://localhost:5006/swagger-ui
```

The Swagger UI provides:
- Interactive API documentation
- Request/response schemas
- "Try it out" functionality
- Complete endpoint listing

### OpenAPI Specification

The raw OpenAPI spec is available at:
```
http://localhost:5006/openapi.json
```

Use this URL for:
- API client generation
- Postman/Insomnia imports
- Contract testing
- Documentation tools

### CDN-Based UI

Swagger UI is loaded from CDN (no bundling required):
- Zero impact on binary size
- Always up-to-date UI
- Fast load times

## Testing

### Health Check

```bash
curl http://localhost:5006/health
```

Response:
```json
{"status":"healthy"}
```

### CORS Preflight

```bash
curl -X OPTIONS http://localhost:5006/v1/weave \
  -H "Origin: https://example.com" \
  -H "Access-Control-Request-Method: POST" \
  -v
```

Expected headers:
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS, PATCH
Access-Control-Allow-Headers: *
Access-Control-Max-Age: 86400
```

### API Call with CORS

```bash
curl -X POST http://localhost:5006/v1/sessions \
  -H "Origin: https://myapp.com" \
  -H "Content-Type: application/json" \
  -d '{}' \
  -v
```

CORS headers will be present in response.

### SSE Streaming

```bash
curl -N -X POST http://localhost:5006/v1/weave:stream \
  -H "Content-Type: application/json" \
  -H "Origin: https://myapp.com" \
  -d '{"query":"test","session_id":""}' \
  -v
```

## Common Issues

### CORS Not Working

**Symptom**: Browser shows CORS errors despite configuration

**Solutions**:
1. Check server logs for CORS warnings
2. Verify `enabled: true` in config
3. Ensure origin matches exactly (including protocol and port)
4. Check browser console for specific error

### Preflight Failures

**Symptom**: OPTIONS requests fail with 404 or 405

**Solutions**:
1. Ensure CORS middleware is enabled
2. Check that `allowed_methods` includes `OPTIONS`
3. Verify server is listening on correct port

### Wildcard + Credentials Error

**Symptom**: Server refuses to start or logs fatal error

**Cause**: Invalid combination of `allowed_origins: ["*"]` with `allow_credentials: true`

**Solution**: Either:
- Use specific origins with credentials: `allowed_origins: ["https://app.com"]` + `allow_credentials: true`
- Or use wildcard without credentials: `allowed_origins: ["*"]` + `allow_credentials: false`

### Swagger UI Not Loading

**Symptom**: `/swagger-ui` returns 404

**Solutions**:
1. Check `http_port` is configured (not 0)
2. Verify server started successfully
3. Check logs for HTTP server startup message
4. Ensure CDN is accessible (requires internet for assets)

## Integration Examples

### Frontend (React/Vue/Angular)

```javascript
// Configure API client with CORS
const api = axios.create({
  baseURL: 'http://localhost:5006',
  withCredentials: true,  // If using credentials
  headers: {
    'Content-Type': 'application/json'
  }
});

// Call Loom API
const response = await api.post('/v1/weave', {
  query: 'Analyze sales data',
  session_id: sessionId
});
```

### SSE Streaming Client

```javascript
const eventSource = new EventSource(
  'http://localhost:5006/v1/weave:stream?query=test',
  {
    withCredentials: true  // If using credentials
  }
);

eventSource.onmessage = (event) => {
  const progress = JSON.parse(event.data);
  console.log('Progress:', progress);
};
```

## Performance

### Caching

- Preflight responses cached for 24 hours (configurable via `max_age`)
- Reduces OPTIONS requests from browsers
- Improves latency for cross-origin requests

### Concurrency

- HTTP server handles concurrent requests efficiently
- SSE streaming supports multiple simultaneous connections
- No connection limits (system defaults apply)

## Security Best Practices

1. **Production Origins**: Always use specific domains in production
2. **HTTPS Only**: Use HTTPS in production (configure TLS)
3. **Credentials**: Only enable if authentication required
4. **Header Restrictions**: Limit `allowed_headers` to necessary headers
5. **Method Restrictions**: Only allow methods your API uses
6. **Regular Updates**: Keep Loom updated for security patches

## Related Documentation

- [TLS Configuration](./tls.md) - HTTPS setup
- [CLI Reference](./cli.md) - Command-line options
- [Production Configuration](../../examples/config/looms-production-cors.yaml) - Example config

## Troubleshooting

Enable debug logging for CORS issues:

```yaml
logging:
  level: "debug"
  format: "json"
```

Check logs for CORS middleware activity.
