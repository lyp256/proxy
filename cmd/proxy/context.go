package main

import (
	"context"

	"github.com/lyp256/proxy/pkg/auth"
)

type typeKey struct{}

var userKey typeKey

func withUser(ctx context.Context, user *auth.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func getUser(ctx context.Context) *auth.User {
	v := ctx.Value(userKey)
	u, _ := v.(*auth.User)
	return u
}
