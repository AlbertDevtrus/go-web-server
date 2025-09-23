package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
)

type HTTPReq struct {
	method  string
	uri     []byte
	version string
	headers [][]byte
}

type HTTPRes struct {
	code    int
	headers [][]byte
	body    BodyReader
}

type BodyReader struct {
	length int
	read   func(p []byte) (n int, err error)
}

type HTTPError struct {
	Code    int
	Message string
}

var (
	ErrHeaderTooLarge   = errors.New("http: header too large")
	ErrIncompleteHeader = errors.New("http: incomplete header")
)

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:8080")

	if err != nil {
		print(err)
	}

	defer ln.Close()

	fmt.Println("Server Running on: ")
	fmt.Println(ln.Addr())

	for {
		conn, err := ln.Accept()

		if err != nil {
			print(err)
		}

		go serveClient(conn)
	}
}

func serveClient(conn net.Conn) {
	defer conn.Close()

	var message []byte
	tmp := make([]byte, 1024)

	for {
		n, err := conn.Read(tmp)

		if err == io.EOF && len(message) == 0 {
			return
		}

		if err != nil {
			fmt.Println("Reading error:", err)
			return
		}

		message = append(message, tmp[:n]...)

		for {
			msg, err := cutMessage(message)

			if err != nil {
				fmt.Println("Error while parsing the message: ", err)
			}

			if msg != nil {
				break
			}
		}

		reqBody, err := readerFromReq(conn, message, tmp)

		if err != nil {
			fmt.Println("Error reading from the request: ", err)
		}

		res := handleReq(message, reqBody)

		io.Copy(io.Discard, bodyReader)
	}

}

func cutMessage(buf []byte) (*HTTPReq, error) {
	kMaxHeaderLen := 1024 * 8

	separetion := []byte("\r\n\r\n")
	indx := bytes.Index(buf, separetion)

	if indx < 0 {
		if len(buf) >= kMaxHeaderLen {
			fmt.Println("Error header is to large")
			return nil, ErrHeaderTooLarge
		}

		return nil, ErrIncompleteHeader
	}

	msg := make([]byte, indx+len(separetion))
	copy(msg, buf[:indx+len(separetion)])

	parseMsg, err := parseHTTPReq(msg)

	if err != nil {
		fmt.Println("Error parsing the request:", err)
		return nil, err
	}

	return parseMsg, nil
}

func parseHTTPReq(data []byte) (*HTTPReq, error) {
	lines := splitLines(data)

	method, uri, version, err := parseReqLine(lines[0])

	if err != nil {
		fmt.Println("Error parsing the request line: ", err)
		return nil, err
	}

	var headers [][]byte
	var header []byte

	for i := 0; i < len(lines)-1; i++ {
		copy(header, lines[i])

		headers = append(headers, header)
	}

	return &HTTPReq{
		method:  method,
		uri:     uri,
		version: version,
		headers: headers,
	}, nil
}

func readerFromReq(conn net.TCPConn, buf []byte, req HTTPReq) (BodyReader, error) {
	bodyLen := -1
	contentLen := fieldGet(req.headers, []byte("Content-Length"))

	if len(contentLen) > 1 {
		var err error
		bodyLen, err = strconv.Atoi(string(contentLen))

		if err != nil {
			return BodyReader{}, fmt.Errorf("Error parsing content length: %w", err)
		}

		if bodyLen < 0 {
			return BodyReader{}, fmt.Errorf("Bad content length: %d", bodyLen)
		}
	}

	bodyAllowed := !(req.method == "GET" || req.method == "HEAD")
	chunked := fieldGet(req.headers, []byte("Transfer-Encoding"))
	isChunked := bytes.Equal(chunked, []byte("chunked"))

	if !bodyAllowed && (bodyLen > 0 || isChunked) {
		return BodyReader{}, fmt.Errorf("HTTP body not allowed")
	}

	if !bodyAllowed {
		bodyLen = 0
	}

	if bodyLen >= 0 {
		return readerFromConnLength(conn, buf, bodyLen), nil
	} else if isChunked {
		return BodyReader{}, fmt.Errorf("HTTP Not supported")
	} else {
		return BodyReader{}, fmt.Errorf("HTTP Not supported")
	}
}

func readerFromConnLength(conn net.TCPConn, buf []byte, remain int) BodyReader {
	availableData := make([]byte, len(buf))
	copy(availableData, buf)

	return BodyReader{
		length: remain,
		read: func(p []byte) (int, error) {
			if remain == 0 {
				return 0, io.EOF
			}

			if len(availableData) == 0 {
				tempBuf := make([]byte, 4096)
				bytesRead, err := conn.Read(tempBuf)

				if err != nil {
					return 0, err
				}
				if bytesRead == 0 {
					return 0, fmt.Errorf("unexpected EOF from HTTP body")
				}

				availableData = tempBuf[:bytesRead]
			}

			toCopy := len(p)
			if len(availableData) < toCopy {
				toCopy = len(availableData)
			}

			if remain < toCopy {
				toCopy = remain
			}

			copy(p[:toCopy], availableData[:toCopy])

			availableData = availableData[toCopy:]
			remain -= toCopy

			return toCopy, nil
		},
	}
}

func handleReq(req HTTPReq, body BodyReader) HTTPRes {
	var resp BodyReader

	switch string(req.uri) {
	case "/echo":
		resp = body
	default:
		resp = readerFromMemory([]byte("Hello world\n"))
	}

	headers := [][]byte{
		[]byte("Server: from_scratch_http_server"),
	}

	return HTTPRes{
		code:    200,
		headers: headers,
		body:    resp,
	}
}

func readerFromMemory(data []byte) BodyReader {
	done := false

	return BodyReader{
		length: len(data),
		read: func(p []byte) (int, error) {
			if done {
				return 0, io.EOF
			} else {

				done = true
				return len(data), nil
			}
		},
	}
}

func writeResponse(conn net.Conn, resp HTTPRes) error {

	response := string(encodeHTTPResp(resp))
	response += fmt.Sprintf("Content-Length: %d\r\n\r\n", resp.body.length)

	_, err := conn.Write([]byte(response))

	if err != nil {
		return fmt.Errorf("error writing headers: %w", err)
	}

	buffer := make([]byte, 4096)
	for {
		n, err := resp.body.read(buffer)

		if n > 0 {
			_, writeErr := conn.Write(buffer[:n])
			if writeErr != nil {
				return fmt.Errorf("error writing body: %w", writeErr)
			}
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("error reading body: %w", err)
		}
	}

	return nil
}

func encodeHTTPResp(resp HTTPRes) []byte {
	reasonPhrase := getReasonPhrase(resp.code)

	encodedResp := fmt.Sprintf("HTTP/1.1 %d %s\r\n", resp.code, reasonPhrase)

	return []byte(encodedResp)
}

func getReasonPhrase(code int) string {
	reasons := map[int]string{
		200: "OK",
		201: "Created",
		400: "Bad Request",
		404: "Not Found",
		500: "Internal Server Error",
	}

	if phrase, exists := reasons[code]; exists {
		return phrase
	}
	return "Unknown"
}

func splitLines(data []byte) [][]byte {
	endLine := []byte("\r\n")

	var indx int
	var lines [][]byte

	for {
		indx = bytes.Index(data, endLine)

		if indx < 0 {
			break
		}

		line := make([]byte, indx+len(endLine))

		copy(line, data[:indx+len(endLine)])
		data = data[indx+len(endLine):]

		lines = append(lines, line)
	}

	return lines
}

func parseReqLine(line []byte) (method string, uri []byte, version string, err error) {
	line = bytes.TrimSpace(line)

	divisions := bytes.SplitN(line, []byte(" "), 3)

	if len(divisions) != 3 {
		err = fmt.Errorf("Request line with wrong format %q", string(line))
		return
	}

	method = string(divisions[0])
	uri = divisions[1]
	version = string(divisions[2])
	return
}

func fieldGet(headers [][]byte, selectedHeader []byte) []byte {
	for i := 0; i < len(headers); i++ {
		if bytes.HasPrefix(headers[i], selectedHeader) {
			parts := bytes.SplitN(headers[i], []byte(":"), 2)

			if len(parts) == 2 {
				return bytes.TrimSpace(parts[1])
			}
		}
	}

	return nil
}
