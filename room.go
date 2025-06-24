package main

import (
	"fmt"
	"net"
	"slices"
	"strings"
	"sync"
)

type RoomId int

var (
	roomRegistry = make(map[RoomId]*Room)
	roomMu       sync.Mutex
	reset        = "\033[0m" // Reset color
	bold         = "\033[1m"
	mentionColor = "\033[96m"
	roomColors   = []string{
		"\033[31m", // Red
		"\033[32m", // Green
		"\033[33m", // Yellow
		"\033[0m",  // Reset color
	}
)

type Room struct {
	RoomID   RoomId
	Users    []*User
	RoomName string
	Secret   string
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

func (r *Room) GetUsers() []*User {
	return r.Users
}

func (r *Room) AddUser(user *User) bool {
	users := r.GetUsers()
	for _, roomUser := range users {
		if roomUser.Addr.String() == user.Addr.String() {
			return false
		}
	}
	r.Users = append(r.Users, user)
	r.AssignColors()
	return true
}

func GetCreateRoom(id RoomId, secret string) (*Room, error) {
	roomMu.Lock()
	defer roomMu.Unlock()

	if room, exists := roomRegistry[id]; exists {
		if room.Secret != secret {
			return nil, fmt.Errorf("wrong secret key for the room %d", id)
		}
		return room, nil
	}

	newRoom := &Room{
		RoomID:   id,
		Users:    []*User{},
		RoomName: "General",
		Secret:   secret,
	}

	roomRegistry[id] = newRoom
	return newRoom, nil
}

func (r *Room) BroadcastMessage(msg string, fromUser string, user *User) *ErrorChan {
	username := user.Username
	color := user.Color
	userConn := *user.conn
	errChan := ErrorChan{
		ErrMap: make(map[string]error),
	}

	isDM := strings.HasPrefix(msg, "/dm")
	dmTargets := ExtractMentions(msg)

	var wg sync.WaitGroup
	for _, usr := range r.GetUsers() {
		if usr.Addr.String() == fromUser {
			continue
		}

		if isDM && !InList(usr.Username, dmTargets) {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			conn := *usr.GetConnection()
			addstr := strings.Split(fromUser, ":")

			var msgLine string

			if isDM {
				msg = strings.TrimPrefix(msg, "/dm ")
				msgLine = fmt.Sprintf("%s/%s %s%s%s (DM): %s\n", addstr[len(addstr)-1], reset, color, user.Username, reset, msg)
				userConn.Write([]byte("✔️ Message sent to: " + usr.Username + "\n"))
				// userConn.Write([]byte("✔️ Message sent to: " + strings.Join(dmTargets, ", ") + "\n"))
			} else {

				highlightedMsg := msg

				for _, name := range dmTargets {
					if name == usr.Username {
						highlightedMsg = strings.ReplaceAll(
							highlightedMsg,
							"@"+name,
							bold+mentionColor+"@"+name+reset,
						)
						highlightedMsg = "\a" + highlightedMsg
					}
				}

				msgLine = fmt.Sprintf(
					"%s/%s %s%s%s: %s\n",
					addstr[len(addstr)-1], reset, color, username, reset, highlightedMsg,
				)
			}
			_, err := conn.Write([]byte(
				// "/" + addstr[len(addstr)-1] + "/" + username + ": " + msg + "\n",
				msgLine,
			))
			if err != nil {
				errChan.AddNewError(conn.RemoteAddr().String(), err)
			}
		}()

	}

	wg.Wait()
	return &errChan
}

func InList(name string, list []string) bool {
	return slices.Contains(list, name)
}

func RemoveUserFromRoom(user *User) {
	roomMu.Lock()
	defer roomMu.Unlock()

	room := roomRegistry[user.RoomID]
	newUsers := []*User{}

	for _, u := range room.Users {
		if u.Addr.String() != user.Addr.String() {
			newUsers = append(newUsers, u)
		}
	}
	room.Users = newUsers

	if len(room.Users) == 0 {
		delete(roomRegistry, user.RoomID)
	} else {
		room.AssignColors()
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

func (r *Room) AssignColors() {
	for i := range r.Users {
		if i < len(roomColors) {
			r.Users[i].Color = roomColors[i]
		} else {
			r.Users[i].Color = "" // Use last color if more users than colors
		}
	}
}

func (r *Room) ListUsers(conn net.Conn) {
	users := r.Users
	for i, user := range users {
		usernames := fmt.Sprintf("%d. %s\n", i+1, user.Username)
		conn.Write([]byte(usernames))
	}
}
