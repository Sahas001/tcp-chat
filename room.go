package main

import (
	"strings"
	"sync"
)

type RoomId int

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

func (r *Room) BroadcastMessage(msg string, fromUser string) *ErrorChan {
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
					"/" + addstr[len(addstr)-1] + ": " + msg + "\n",
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
