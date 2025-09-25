package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

// Test para parseReqLine
func TestParseReqLine(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		wantMethod  string
		wantURI     string
		wantVersion string
		wantErr     bool
	}{
		{
			name:        "Valid GET request",
			input:       []byte("GET /hello HTTP/1.1\r\n"),
			wantMethod:  "GET",
			wantURI:     "/hello",
			wantVersion: "HTTP/1.1",
			wantErr:     false,
		},
		{
			name:        "Valid POST request",
			input:       []byte("POST /echo HTTP/1.1\r\n"),
			wantMethod:  "POST",
			wantURI:     "/echo",
			wantVersion: "HTTP/1.1",
			wantErr:     false,
		},
		{
			name:        "Invalid format - missing parts",
			input:       []byte("GET /hello\r\n"),
			wantMethod:  "",
			wantURI:     "",
			wantVersion: "",
			wantErr:     true,
		},
		{
			name:        "Invalid format - too many parts",
			input:       []byte("GET /hello HTTP/1.1 extra\r\n"),
			wantMethod:  "GET",
			wantURI:     "/hello",
			wantVersion: "HTTP/1.1 extra",
			wantErr:     false, // El código actual acepta esto
		},
		{
			name:        "Empty line",
			input:       []byte("\r\n"),
			wantMethod:  "",
			wantURI:     "",
			wantVersion: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, uri, version, err := parseReqLine(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseReqLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if method != tt.wantMethod {
				t.Errorf("parseReqLine() method = %v, want %v", method, tt.wantMethod)
			}

			if string(uri) != tt.wantURI {
				t.Errorf("parseReqLine() uri = %v, want %v", string(uri), tt.wantURI)
			}

			if version != tt.wantVersion {
				t.Errorf("parseReqLine() version = %v, want %v", version, tt.wantVersion)
			}
		})
	}
}

// Test para fieldGet
func TestFieldGet(t *testing.T) {
	headers := [][]byte{
		[]byte("Host: localhost:8080"),
		[]byte("Content-Type: application/json"),
		[]byte("Content-Length: 123"),
		[]byte("Authorization: Bearer token123"),
	}

	tests := []struct {
		name           string
		selectedHeader []byte
		want           string
	}{
		{
			name:           "Get Host header",
			selectedHeader: []byte("Host"),
			want:           "localhost:8080",
		},
		{
			name:           "Get Content-Type header",
			selectedHeader: []byte("Content-Type"),
			want:           "application/json",
		},
		{
			name:           "Get Content-Length header",
			selectedHeader: []byte("Content-Length"),
			want:           "123",
		},
		{
			name:           "Get non-existent header",
			selectedHeader: []byte("X-Custom"),
			want:           "",
		},
		{
			name:           "Case sensitive search",
			selectedHeader: []byte("host"), // lowercase
			want:           "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fieldGet(headers, tt.selectedHeader)
			got := string(result)
			if got != tt.want {
				t.Errorf("fieldGet() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test para splitLines
func TestSplitLines(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  int // número de líneas esperadas
	}{
		{
			name:  "Single line",
			input: []byte("GET /hello HTTP/1.1\r\n"),
			want:  1,
		},
		{
			name:  "Multiple lines",
			input: []byte("GET /hello HTTP/1.1\r\nHost: localhost\r\nContent-Type: text/plain\r\n"),
			want:  3,
		},
		{
			name:  "Empty input",
			input: []byte(""),
			want:  0,
		},
		{
			name:  "Line without CRLF",
			input: []byte("GET /hello HTTP/1.1"),
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := splitLines(tt.input)
			if len(lines) != tt.want {
				t.Errorf("splitLines() returned %d lines, want %d", len(lines), tt.want)
			}
		})
	}
}

// Test para parseHTTPReq
func TestParseHTTPReq(t *testing.T) {
	tests := []struct {
		name       string
		input      []byte
		wantMethod string
		wantURI    string
		wantErr    bool
	}{
		{
			name: "Valid GET request with headers",
			input: []byte("GET /hello HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"User-Agent: test-client\r\n" +
				"\r\n"),
			wantMethod: "GET",
			wantURI:    "/hello",
			wantErr:    false,
		},
		{
			name: "Valid POST request",
			input: []byte("POST /echo HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"Content-Length: 5\r\n" +
				"\r\n"),
			wantMethod: "POST",
			wantURI:    "/echo",
			wantErr:    false,
		},
		{
			name: "Invalid request line",
			input: []byte("INVALID\r\n" +
				"Host: localhost:8080\r\n" +
				"\r\n"),
			wantMethod: "",
			wantURI:    "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := parseHTTPReq(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseHTTPReq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if req.method != tt.wantMethod {
					t.Errorf("parseHTTPReq() method = %v, want %v", req.method, tt.wantMethod)
				}
				if string(req.uri) != tt.wantURI {
					t.Errorf("parseHTTPReq() uri = %v, want %v", string(req.uri), tt.wantURI)
				}
			}
		})
	}
}

// Test para cutMessage
func TestCutMessage(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr error
	}{
		{
			name: "Complete message",
			input: []byte("GET /hello HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n" +
				"\r\n" +
				"body content"),
			wantErr: nil,
		},
		{
			name: "Incomplete message",
			input: []byte("GET /hello HTTP/1.1\r\n" +
				"Host: localhost:8080\r\n"),
			wantErr: ErrIncompleteHeader,
		},
		{
			name:    "Header too large",
			input:   bytes.Repeat([]byte("X-Large-Header: "), 600), // > 8KB
			wantErr: ErrHeaderTooLarge,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := cutMessage(tt.input)
			if err != tt.wantErr {
				t.Errorf("cutMessage() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

// Test para readerFromMemory
func TestReaderFromMemory(t *testing.T) {
	data := []byte("Hello, World!")
	reader := readerFromMemory(data)

	if reader.length != len(data) {
		t.Errorf("readerFromMemory() length = %v, want %v", reader.length, len(data))
	}

	// Test reading the data
	buffer := make([]byte, len(data))
	n, err := reader.read(buffer)

	if err != nil {
		t.Errorf("readerFromMemory() read error = %v", err)
	}

	if n != len(data) {
		t.Errorf("readerFromMemory() read bytes = %v, want %v", n, len(data))
	}

	if !bytes.Equal(buffer[:n], data) {
		t.Errorf("readerFromMemory() read data = %v, want %v", buffer[:n], data)
	}

	// Test reading again should return EOF
	_, err = reader.read(buffer)
	if err != io.EOF {
		t.Errorf("readerFromMemory() second read should return EOF, got %v", err)
	}
}

// Test para getReasonPhrase
func TestGetReasonPhrase(t *testing.T) {
	tests := []struct {
		code int
		want string
	}{
		{200, "OK"},
		{201, "Created"},
		{400, "Bad Request"},
		{404, "Not Found"},
		{500, "Internal Server Error"},
		{999, "Unknown"}, // código no mapeado
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Code_%d", tt.code), func(t *testing.T) {
			got := getReasonPhrase(tt.code)
			if got != tt.want {
				t.Errorf("getReasonPhrase(%d) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

// Test para encodeHTTPResp
func TestEncodeHTTPResp(t *testing.T) {
	tests := []struct {
		name string
		resp HTTPRes
		want string
	}{
		{
			name: "200 OK response",
			resp: HTTPRes{code: 200},
			want: "HTTP/1.1 200 OK\r\n",
		},
		{
			name: "404 Not Found response",
			resp: HTTPRes{code: 404},
			want: "HTTP/1.1 404 Not Found\r\n",
		},
		{
			name: "Unknown status code",
			resp: HTTPRes{code: 999},
			want: "HTTP/1.1 999 Unknown\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodeHTTPResp(tt.resp)
			if string(got) != tt.want {
				t.Errorf("encodeHTTPResp() = %v, want %v", string(got), tt.want)
			}
		})
	}
}

// Test de integración - servidor completo
func TestServerIntegration(t *testing.T) {
	// Iniciar servidor en puerto de prueba
	ln, err := net.Listen("tcp", "127.0.0.1:0") // puerto automático
	if err != nil {
		t.Fatalf("Failed to start test server: %v", err)
	}
	defer ln.Close()

	// Ejecutar servidor en goroutine
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return // servidor cerrado
			}
			go serveClient(conn)
		}
	}()

	// Dar tiempo al servidor para iniciarse
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name           string
		request        string
		expectedStatus string
		expectedBody   string
	}{
		{
			name: "GET root path",
			request: "GET / HTTP/1.1\r\n" +
				"Host: localhost\r\n" +
				"\r\n",
			expectedStatus: "HTTP/1.1 200 OK",
			expectedBody:   "Hello world\n",
		},
		{
			name: "GET /echo path",
			request: "GET /echo HTTP/1.1\r\n" +
				"Host: localhost\r\n" +
				"\r\n",
			expectedStatus: "HTTP/1.1 200 OK",
			expectedBody:   "",
		},
		{
			name: "POST /echo with body",
			request: "POST /echo HTTP/1.1\r\n" +
				"Host: localhost\r\n" +
				"Content-Length: 12\r\n" +
				"\r\n" +
				"Hello Server",
			expectedStatus: "HTTP/1.1 200 OK",
			expectedBody:   "Hello Server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Conectar al servidor de prueba
			conn, err := net.Dial("tcp", ln.Addr().String())
			if err != nil {
				t.Fatalf("Failed to connect to test server: %v", err)
			}
			defer conn.Close()

			// Enviar request
			_, err = conn.Write([]byte(tt.request))
			if err != nil {
				t.Fatalf("Failed to send request: %v", err)
			}

			// Leer response
			response := make([]byte, 4096)
			n, err := conn.Read(response)
			if err != nil {
				t.Fatalf("Failed to read response: %v", err)
			}

			responseStr := string(response[:n])

			// Verificar status line
			if !strings.Contains(responseStr, tt.expectedStatus) {
				t.Errorf("Expected status %v not found in response: %v", tt.expectedStatus, responseStr)
			}

			// Verificar body si se espera contenido
			if tt.expectedBody != "" && !strings.Contains(responseStr, tt.expectedBody) {
				t.Errorf("Expected body %v not found in response: %v", tt.expectedBody, responseStr)
			}
		})
	}
}

// Benchmark para parseHTTPReq
func BenchmarkParseHTTPReq(b *testing.B) {
	data := []byte("GET /hello HTTP/1.1\r\n" +
		"Host: localhost:8080\r\n" +
		"User-Agent: benchmark-client\r\n" +
		"Content-Type: application/json\r\n" +
		"\r\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parseHTTPReq(data)
		if err != nil {
			b.Fatalf("parseHTTPReq failed: %v", err)
		}
	}
}

// Benchmark para fieldGet
func BenchmarkFieldGet(b *testing.B) {
	headers := [][]byte{
		[]byte("Host: localhost:8080"),
		[]byte("Content-Type: application/json"),
		[]byte("Content-Length: 123"),
		[]byte("Authorization: Bearer token123"),
		[]byte("User-Agent: benchmark-client"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fieldGet(headers, []byte("Content-Length"))
	}
}
