package main

import (
	"strings"
)

type auth map[string]string

func (u auth) Auth(user, password string) bool {
	p, ok := u[user]
	if !ok {
		return false
	}
	if p == "" {
		return true
	}

	return p == password
}

func parseAuthUsers(rawUsers []string) auth {
	users := make(map[string]string)
	for _, user := range rawUsers {
		parts := strings.SplitN(user, ":", 2)
		if len(parts) == 2 {
			users[parts[0]] = parts[1]
		} else {
			users[parts[0]] = ""
		}
	}
	return users
}
