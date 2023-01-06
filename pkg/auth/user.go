package auth

import (
	"sync"
)

const GlobalUser = "global_user"

type User struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	UUID     string `yaml:"uuid"`
}

func (u *User) Copy() *User {
	if u == nil {
		return nil
	}
	return &User{
		Username: u.Username,
		Password: u.Password,
		UUID:     u.UUID,
	}
}

type rosters struct {
	nameIndex map[string]*User
	uuidIndex map[string]*User
	rw        sync.RWMutex
}

func (r *rosters) Load(users ...User) {
	r.rw.Lock()
	defer r.rw.Unlock()
	r.nameIndex = make(map[string]*User, len(users))
	r.uuidIndex = make(map[string]*User, len(users))
	for i := range users {
		user := users[i].Copy()
		if user.Username != "" {
			r.nameIndex[user.Username] = user
		}
		if user.UUID != "" {
			r.uuidIndex[user.UUID] = user
		}
	}
}

func (r *rosters) GetByUsername(username string) *User {
	r.rw.RLock()
	defer r.rw.RUnlock()
	return r.nameIndex[username].Copy()
}

func (r *rosters) GetByUUID(uuid string) *User {
	r.rw.RLock()
	defer r.rw.RUnlock()
	return r.uuidIndex[uuid].Copy()
}

func (r *rosters) Auth(username string, password string) bool {
	user := r.GetByUsername(username)
	if user == nil {
		return false
	}
	return user.Password == password
}

var defaultUsers = &rosters{}

func Load(users ...User) {
	defaultUsers.Load(users...)
}

func GetByUsername(username string) *User {
	return defaultUsers.GetByUsername(username)
}

func GetByUUID(uid string) *User {
	return defaultUsers.GetByUUID(uid)
}

func Auth(username string, password string) bool {
	return defaultUsers.Auth(username, password)
}
