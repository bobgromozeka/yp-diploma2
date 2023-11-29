package storage

import (
	"context"
	"errors"
)

var ErrLoginAlreadyExists = errors.New("user with this login already exists")

type User struct {
	ID       int
	Login    string
	Password string
}

type UserStorage interface {
	CreateUser(ctx context.Context, login string, password []byte) error
	GetUser(ctx context.Context, login string) (*User, error)
}
