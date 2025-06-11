package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

type Server struct {
	Addr     string
	Listener net.Listener
}

func NewServer(addr string) *Server {
	return &Server{
		Addr: addr,
	}
}

func (s *Server) Listen() error {
	ln, err := net.Listen("tcp", s.Addr)
	s.Listener = ln
	return err
}

func main() {
	server := NewServer(":2000")
	err := server.Listen()
	if err != nil {
		panic(err)
	}
}

//TODO: Refactor these function and complete main function.

func HandleIncomingConnection(conn net.Conn) {
	defer conn.Close()
	_, err := conn.Write([]byte("Connected to Server Successfully!\n"))
	if err != nil {
		log.Println("Failed writing")
	}
	for {
		buf := make([]byte, 64)

		n, err := conn.Read(buf)
		if err != nil {
			log.Println("Connection Abrupted")
			break
		}
		fmt.Printf("Client: %s\n", string(buf[:n]))
	}
}

func HandleOutgoingConnection(conn net.Conn) {
	defer conn.Close()
	for {
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error while reading the line")
		}
		_, err = conn.Write(fmt.Appendf(nil, "Server: %s", line))
		if err != nil {
			log.Printf("Connection Abrupted: %s\n", err.Error())
		}
	}
}
