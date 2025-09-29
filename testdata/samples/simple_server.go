package main

import (
	"net"
	"net/http"
)

const serverPort = ":8080"

func main() {
	// HTTP server on port 3000
	http.ListenAndServe(":3000", nil)

	// TCP listener with constant
	listener, _ := net.Listen("tcp", serverPort)
	defer listener.Close()

	// HTTPS server
	http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil)

	// UDP listener
	conn, _ := net.ListenUDP("udp", &net.UDPAddr{Port: 9090})
	defer conn.Close()
}