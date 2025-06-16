package main

import (
	"bufio"
	"net"
)

type User struct {
	Addr     net.Addr
	Username string
	RoomID   RoomId
	conn     *net.Conn
	Room     *Room
	Color    string
}

func (u *User) GetConnection() *net.Conn {
	return u.conn
}

func (u *User) SetUsername() {
	conn := *u.GetConnection()

	conn.Write([]byte("Enter your username: "))
	reader := bufio.NewReader(conn)
	username, err := reader.ReadString('\n')
	if err != nil {
		conn.Write([]byte("Error reading username: " + err.Error() + "\n"))
		return
	}
	u.Username = username[:len(username)-1] // Remove the newline character
}

func (u *User) GetRoom() *Room {
	return u.Room
}
