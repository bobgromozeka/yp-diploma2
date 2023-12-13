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

type Card struct {
	ID                int
	UserID            int
	Name              string
	Number            string
	ValidThroughMonth int
	ValidThroughYear  int
	CVV               int
	Description       *string
}

type Bin struct {
	ID          int
	UserID      int
	Name        string
	Data        []byte
	Description *string
}

type CreateCardParams struct {
	UserID            int
	Name              string
	Number            string
	ValidThroughMonth int
	ValidThroughYear  int
	Cvv               int
	Description       *string
}

//go:generate mockgen -source ./storage.go -destination storage_mock.go -package storage

type UserStorage interface {
	CreateUser(ctx context.Context, login string, password []byte) error
	GetUser(ctx context.Context, login string) (*User, error)
}

type DataKeeperStorage interface {
	CreatePasswordPair(ctx context.Context, userID int, login, password string, description *string) error
	GetAllPasswordPairs(ctx context.Context, userID int) ([]PasswordPair, error)
	RemovePasswordPair(ctx context.Context, userID, ID int) error

	CreateText(ctx context.Context, userID int, name string, text string, description *string) error
	GetAllTexts(ctx context.Context, userID int) ([]Text, error)
	RemoveText(ctx context.Context, userID, ID int) error

	CreateCard(ctx context.Context, params CreateCardParams) error
	GetAllCards(ctx context.Context, userID int) ([]Card, error)
	RemoveCard(ctx context.Context, userID, ID int) error

	CreateBin(ctx context.Context, userID int, name string, data []byte, description *string) error
	GetAllBins(ctx context.Context, userID int) ([]Bin, error)
	RemoveBin(ctx context.Context, userID, ID int) error
}
