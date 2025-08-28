package main

import (
	"fmt"
	"net"
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

//  Add backpressure

func handleConection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("Remote address")
	fmt.Println(conn.RemoteAddr())

	data := make(chan []byte, 10)

	defer close(data)

	go func() {
		for msg := range data {
			if _, err := conn.Write(msg); err != nil {
				fmt.Println("Writing error:", err)
				return
			}
		}
	}()

	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)

		if err != nil {
			print(err)
			return
		}

		response := buffer[:n]

		if string(response) == "Hola\n" {
			select {
			case data <- []byte("Adios\n"):
			default:
				fmt.Println("Clossing conection")
				return
			}
		}

	}

}

// Basic handle conection
// func handleConection(conn net.Conn) {
// 	fmt.Println("Remote Address")
// 	fmt.Println(conn.RemoteAddr())

// 	defer conn.Close()

// 	buffer := make([]byte, 1024)

// 	for {
// 		n, err := conn.Read(buffer)

// 		if err != nil {
// 			print(err)
// 			return
// 		}

// 		data := string(buffer[:n])

// 		fmt.Println("Data")
// 		fmt.Println(data)

// 		conn.Write(buffer)

// 		if strings.Contains(data, "q") {
// 			fmt.Println("Closing conection")
// 			conn.Close()
// 			return
// 		}
// 	}

// }
