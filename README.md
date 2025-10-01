# Go HTTP Web Server

A lightweight HTTP/1.1 web server implementation written in Go from scratch, without using the standard `net/http` package.

## Overview

This project implements a basic HTTP server that handles TCP connections directly, parses HTTP requests manually, and constructs HTTP responses according to the HTTP/1.1 specification. It includes unit tests, integration tests, and benchmarks.

## Features

- **Raw TCP Connection Handling**: Direct socket programming using `net.Conn`
- **HTTP/1.1 Protocol Support**: Manual parsing of HTTP requests and response generation
- **Concurrent Request Handling**: Each client connection handled in a separate goroutine
- **Echo Endpoint**: POST requests to `/echo` return the request body
- **Request Body Support**: Handles Content-Length based body reading
- **Error Handling**: Proper HTTP error responses (400, 413, etc.)
- **Comprehensive Testing**: Unit tests, integration tests, and benchmarks included

## Project Structure

```
.
├── main.go           # Server implementation
├── main_test.go      # Tests and benchmarks
└── go.mod           # Go module definition
```

## Getting Started

### Prerequisites

- Go 1.25.0 or later

### Installation

```bash
git clone https://github.com/AlbertDevtrus/go-web-server
cd go-web-server
```

### Running the Server

```bash
go run main.go
```

The server will start on `127.0.0.1:8080` and output:
```
Server Running on:
127.0.0.1:8080
```

### Testing

Run all tests:
```bash
go test -v
```

Run integration tests:
```bash
go test -v -run TestServerIntegration
```

Run benchmarks:
```bash
go test -bench=.
```

## Usage Examples

### Default Route (GET)

```bash
curl http://localhost:8080/
```

Response:
```
Hello world
```

### Echo Endpoint (POST)

```bash
curl -X POST http://localhost:8080/echo -d "Hello Server"
```

Response:
```
Hello Server
```

## Architecture

### Core Components

#### Data Structures

- **HTTPReq**: Represents an HTTP request with method, URI, version, and headers
- **HTTPRes**: Represents an HTTP response with status code, headers, and body reader
- **BodyReader**: Custom reader interface for streaming request/response bodies

#### Key Functions

**Server Functions:**
- `main()`: Initializes TCP listener and accepts connections (main.go:41)
- `serveClient(conn net.Conn)`: Handles individual client connections (main.go:64)

**Request Parsing:**
- `cutMessage(buf []byte)`: Extracts complete HTTP messages from buffer (main.go:146)
- `parseHTTPReq(data []byte)`: Parses HTTP request into HTTPReq struct (main.go:174)
- `parseReqLine(line []byte)`: Parses the HTTP request line (main.go:415)
- `splitLines(data []byte)`: Splits data by CRLF delimiters (main.go:391)
- `fieldGet(headers [][]byte, selectedHeader []byte)`: Extracts specific header value (main.go:431)

**Request Body Handling:**
- `readerFromReq(conn, buf, req)`: Creates body reader from request (main.go:202)
- `readerFromConnLength(conn, buf, remain)`: Reads Content-Length based bodies (main.go:252)
- `readerFromMemory(data []byte)`: Creates reader from in-memory data (main.go:317)

**Response Generation:**
- `handleReq(req, body)`: Routes requests and generates responses (main.go:296)
- `writeResponse(conn, resp)`: Writes HTTP response to connection (main.go:334)
- `encodeHTTPResp(resp)`: Encodes HTTP response status line (main.go:368)
- `getReasonPhrase(code)`: Maps status codes to reason phrases (main.go:376)

## HTTP Protocol Implementation Details

### Request Parsing

The server reads from TCP connections in chunks, accumulating data until a complete HTTP header is received (identified by `\r\n\r\n` sequence). The header size is limited to 8KB to prevent memory exhaustion attacks.

### Body Reading

- Supports Content-Length based body reading
- Implements streaming body reader to avoid loading entire body into memory
- Validates that GET/HEAD requests don't include bodies
- Returns appropriate errors for unsupported features (chunked transfer encoding)

### Response Generation

Responses include:
- HTTP/1.1 status line with appropriate status codes
- Custom `Server` header
- Content-Length header
- Streaming body write for efficient memory usage

### Error Handling

The server returns standard HTTP error responses:
- **400 Bad Request**: Malformed requests
- **413 Request Entity Too Large**: Headers exceeding 8KB limit
- **200 OK**: Successful requests

## Testing

The test suite includes:

### Unit Tests
- `TestParseReqLine`: Request line parsing
- `TestFieldGet`: Header field extraction
- `TestSplitLines`: Line splitting logic
- `TestParseHTTPReq`: Full request parsing
- `TestCutMessage`: Message extraction from buffer
- `TestReaderFromMemory`: Memory-based body reader
- `TestGetReasonPhrase`: Status code mapping
- `TestEncodeHTTPResp`: Response encoding

### Integration Tests
- `TestServerIntegration`: End-to-end server functionality testing

### Benchmarks
- `BenchmarkParseHTTPReq`: Request parsing performance
- `BenchmarkFieldGet`: Header lookup performance

## Limitations

Current implementation has the following limitations:

- No HTTPS/TLS support
- No chunked transfer encoding support
- No persistent connections (Connection: keep-alive)
- No request timeout handling
- Case-sensitive header matching
- Limited to Content-Length based body reading
- No compression support (gzip, deflate)
- No multipart form data parsing

## Implementation Notes

### Concurrency

Each client connection is handled in a separate goroutine, allowing the server to handle multiple concurrent connections. A `sync.WaitGroup` ensures response writing completes before closing connections.

### Memory Management

The implementation uses byte slices extensively and includes careful buffer management to avoid excessive allocations. Body readers use a streaming approach rather than loading entire bodies into memory.

### Error Recovery

The server logs errors but continues accepting new connections, ensuring one malformed request doesn't crash the entire server.

## License

This project is provided as-is for educational purposes.

## Contributing

This is a learning project. Feel free to explore and experiment with the code.

## Acknowledgments

Built from scratch to understand HTTP protocol implementation and low-level network programming in Go.
