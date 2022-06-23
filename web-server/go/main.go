package main

import (
	"fmt"
	"io"
	"net"
)

type WebServer struct {
	addr     *net.TCPAddr
	listener net.Listener
}

func NewWebServer(serverAddress string) (ws *WebServer, err error) {
	addr, err := net.ResolveTCPAddr("tcp4", serverAddress)
	if err != nil {
		return
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return
	}

	ws = &WebServer{
		addr:     addr,
		listener: listener,
	}
	return
}

func (ws *WebServer) Start() {
	fmt.Printf("Start listening on %s\n", ws.addr.String())
	for {
		conn, err := ws.listener.Accept()
		if err != nil {
			continue
		}

		go ws.handleOne(conn)
	}
}

func (ws *WebServer) handleOne(conn net.Conn) {
	defer conn.Close()

	io.WriteString(conn, "HTTP/1.1 200 OK\r\n"+
		"Content-Type: text/html; charset=utf-8\r\n"+
		"Content-Length: 20\r\n"+
		"\r\n"+
		"<h1>hello world</h1>")
}

func main() {
	serverAddress := ":8888"

	ws, err := NewWebServer(serverAddress)
	if err != nil {
		panic(err)
	}

	ws.Start()
}
