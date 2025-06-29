package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"
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

	for {

		conn, err := server.Listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		reader := bufio.NewReader(conn)

		conn.Write([]byte("Enter room number:... "))
		roomStr, _ := reader.ReadString('\n')
		roomStr = strings.TrimSpace(roomStr)

		conn.Write([]byte("Enter room secret:..."))
		secret, _ := reader.ReadString('\n')
		secret = strings.TrimSpace(secret)

		roomIdInt, err := strconv.Atoi(roomStr)
		if err != nil {
			conn.Write([]byte("Error: Invalid room number\n"))
			conn.Close()
			continue
		}
		room, err := GetCreateRoom(RoomId(roomIdInt), secret)
		if err != nil {
			conn.Write([]byte("Error: " + err.Error() + "\n"))
			conn.Close()
			continue
		}

		user := &User{
			Addr:   conn.RemoteAddr(),
			conn:   &conn,
			RoomID: RoomId(roomIdInt),
			Room:   room,
		}
		user.SetUsername()

		room.AddUser(user)

		go HandleIncomingConnection(user)
	}
}

func HandleIncomingConnection(user *User) {
	conn := *user.GetConnection()
	defer conn.Close()
	user.GetRoom()
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			log.Println("User disconnected:", user.Username)
			RemoveUserFromRoom(user)
			break
		}
		msg := string(buf[:n])
		room := user.GetRoom()

		trimmedMsg := strings.TrimSpace(msg)

		if trimmedMsg == "/room" {
			ListRooms(conn)
			continue
		} else if trimmedMsg == "/users" {
			user.Room.ListUsers(conn)
		}

		errC := room.BroadcastMessage(msg, conn.RemoteAddr().String(), user)
		if len(errC.ErrMap) != 0 {
			log.Println("Some users could not receive the message")
		}

	}
}
