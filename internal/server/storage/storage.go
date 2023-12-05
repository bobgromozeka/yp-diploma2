package storage

import (
	"context"
	"errors"
)

var ErrLoginAlreadyExists = errors.New("user with this login already exists")

type StoragesFactory interface {
	CreateUserStorage() UserStorage
	CreateDataKeeperStorage() DataKeeperStorage
}

type User struct {
	ID       int
	Login    string
	Password string
}

type PasswordPair struct {
	ID          int
	UserID      int
	Login       string
	Password    string
	Description *string
}

type Text struct {
	ID          int
	UserID      int
	Name        string
	T           string
	Description *string
}

type UserStorage interface {
	CreateUser(ctx context.Context, login string, password []byte) error
	GetUser(ctx context.Context, login string) (*User, error)
}

type DataKeeperStorage interface {
	CreatePasswordPair(ctx context.Context, userID int, login, password string, description *string) error
	GetAllPasswordPairs(ctx context.Context, userID int) ([]PasswordPair, error)
	RemovePasswordPair(ctx context.Context, userID int, ID int) error

	CreateText(ctx context.Context, userID int, name string, text string, description *string) error
	GetAllTexts(ctx context.Context, userID int) ([]Text, error)
	RemoveText(ctx context.Context, userID int, ID int) error
}
