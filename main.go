package main

import (
	"fmt"
	"net"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:8080")

	if err != nil {
		print(err)
	}

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

func handleConection(conn net.Conn) {
	fmt.Println("Remote Address")
	fmt.Println(conn.RemoteAddr())

	tcpConn := conn.(*net.TCPConn)

	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)

		if err != nil {
			print(err)
			return
		}

		data := string(buffer[:n])

		fmt.Println("Data")
		fmt.Println(data)

		if strings.Contains(data, "q") {
			fmt.Println("Closing conection")
			conn.Close()
			return
		}
	}

}
