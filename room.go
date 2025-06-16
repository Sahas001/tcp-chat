package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type RoomId int

var (
	roomRegistry = make(map[RoomId]*Room)
	roomMu       sync.Mutex
)

type Room struct {
	RoomID   RoomId
	Users    []User
	RoomName string
}

type ErrorChan struct {
	mu     sync.Mutex
	ErrMap map[string]error
}

func (e *ErrorChan) AddNewError(routineAddr string, err error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.ErrMap[routineAddr] = err
}

func (r *Room) GetUsers() []User {
	return r.Users
}

func (r *Room) AddUser(user User) bool {
	users := r.GetUsers()
	for _, roomUser := range users {
		if roomUser.Addr.String() == user.Addr.String() {
			return false
		}
	}
	r.Users = append(r.Users, user)
	return true
}

func GetCreateRoom(id RoomId) *Room {
	roomMu.Lock()
	defer roomMu.Unlock()

	if room, exists := roomRegistry[id]; exists {
		return room
	}

	newRoom := &Room{
		RoomID:   id,
		Users:    []User{},
		RoomName: "General",
	}

	roomRegistry[id] = newRoom
	return newRoom
}

func (r *Room) BroadcastMessage(msg string, fromUser string, username string) *ErrorChan {
	errChan := ErrorChan{
		ErrMap: make(map[string]error),
	}

	var wg sync.WaitGroup
	for _, usr := range r.GetUsers() {
		if usr.Addr.String() != fromUser {
			wg.Add(1)
			go func() {
				conn := *usr.GetConnection()
				addstr := strings.Split(fromUser, ":")

				_, err := conn.Write([]byte(
					"/" + addstr[len(addstr)-1] + "/" + username + ": " + msg + "\n",
				))
				if err != nil {
					errChan.AddNewError(conn.RemoteAddr().String(), err)
				}
				wg.Done()
			}()

		}
	}

	wg.Wait()
	return &errChan
}

func RemoveUserFromRoom(user User) {
	roomMu.Lock()
	defer roomMu.Unlock()

	room := roomRegistry[user.RoomID]
	newUsers := []User{}

	for _, u := range room.Users {
		if u.Addr.String() != user.Addr.String() {
			newUsers = append(newUsers, u)
		}
	}
	room.Users = newUsers

	if len(room.Users) == 0 {
		delete(roomRegistry, user.RoomID)
	}
}

func ListRooms(conn net.Conn) {
	roomMu.Lock()
	defer roomMu.Unlock()

	if len(roomRegistry) == 0 {
		conn.Write([]byte("There is no active room present."))
		return
	}

	conn.Write([]byte("\033[1;34mActive Rooms:\033[0m\n")) // Blue bold title
	for _, room := range roomRegistry {
		line := fmt.Sprintf(
			"- %s (ID: %d, Users: %d)\n",
			room.RoomName,
			room.RoomID,
			len(room.Users),
		)
		conn.Write([]byte(line))
	}
}

func (r *Room) ListUsers(conn net.Conn) {
	users := r.Users
	for i, user := range users {
		usernames := fmt.Sprintf("%d. %s\n", i+1, user.Username)
		conn.Write([]byte(usernames))
	}
}
