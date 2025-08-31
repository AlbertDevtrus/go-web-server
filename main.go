package main

import (
	"bytes"
	"fmt"
	"net"
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

		go handleConection(conn)
	}
}

//  Add backpressure

func handleConection(conn net.Conn) {
	fmt.Println("Remote address")
	fmt.Println(conn.RemoteAddr())

	data := make(chan []byte, 10)

	go func() {
		for msg := range data {
			if _, err := conn.Write(msg); err != nil {
				fmt.Println("Writing error:", err)
				return
			}
		}

	}()

	var message []byte
	tmp := make([]byte, 1024)

	for {
		defer conn.Close()
		n, err := conn.Read(tmp)

		if err != nil {
			fmt.Println("Reading error:", err)
			return
		}

		message = append(message, tmp[:n]...)

		for {
			var msg []byte
			msg, message = cutMessage(message)

			if msg == nil {
				break
			}

			var response []byte

			if bytes.Contains(msg, []byte("quit\n")) {
				data <- []byte("bye\n")
				return
			} else {
				response = []byte("Echo: " + string(msg))
			}

			data <- []byte(response)
		}
	}

}

func cutMessage(buf []byte) ([]byte, []byte) {
	indx := bytes.IndexByte(buf, '\n')

	if indx < 0 {
		return nil, buf
	}

	msg := make([]byte, indx+1)
	copy(msg, buf[:indx+1])

	buf = buf[indx+1:]

	return msg, buf
}
