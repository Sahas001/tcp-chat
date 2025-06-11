package main

type RoomId int

type Room struct {
	RoomID   RoomId
	Users    []User
	RoomName string
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
