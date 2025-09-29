package main

import (
	"net"
	"net/http"
	"time"
)

const apiHost = "api.example.com:443"

func main() {
	// HTTP GET request
	resp, _ := http.Get("https://www.google.com/search")
	defer resp.Body.Close()

	// HTTP POST request
	http.Post("http://localhost:8080/api", "application/json", nil)

	// TCP dial with constant
	conn, _ := net.Dial("tcp", apiHost)
	defer conn.Close()

	// TCP dial with timeout
	conn2, _ := net.DialTimeout("tcp", "database.internal:5432", 5*time.Second)
	defer conn2.Close()

	// UDP dial
	udpConn, _ := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP("8.8.8.8"),
		Port: 53,
	})
	defer udpConn.Close()
}
