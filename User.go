package main

import "net"

type User struct {
	Addr   net.Addr
	RoomID RoomId
	conn   *net.Conn
	Room   *Room
}

func (u *User) GetConnection() *net.Conn {
	return u.conn
}

func (u *User) GetRoom() *Room {
	return u.Room
}
