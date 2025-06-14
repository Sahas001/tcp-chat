package main

import (
	"log"
	"net"
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
	var roomId RoomId = 1

	room := Room{
		RoomID:   roomId,
		RoomName: "General",
	}

	for {
		conn, err := server.Listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		user := User{
			Addr:   conn.RemoteAddr(),
			conn:   &conn,
			RoomID: roomId,
			Room:   &room,
		}
		user.SetUsername()

		room.AddUser(user)

		go HandleIncomingConnection(user)
	}
}

func HandleIncomingConnection(user User) {
	conn := *user.GetConnection()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("Error reading from connection:", err)
		}
		msg := string(buf[:n])
		room := user.GetRoom()

		errC := room.BroadcastMessage(msg, conn.RemoteAddr().String(), user.Username)
		if len(errC.ErrMap) != 0 {
			log.Println("Some users could not receive the message")
		}

	}
}
