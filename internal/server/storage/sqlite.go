package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/mattn/go-sqlite3"
)

type SQLiteUserStorage struct {
	db *sql.DB
}

func NewSQLiteUserStorage(db *sql.DB) *SQLiteUserStorage {
	return &SQLiteUserStorage{
		db: db,
	}
}

func (s *SQLiteUserStorage) CreateUser(ctx context.Context, login string, password []byte) error {
	_, err := s.db.ExecContext(
		ctx, "insert into users (login, password) values ($1, $2)", login, password,
	)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return ErrLoginAlreadyExists
			}
		}
		return err
	}

	return nil
}

func (s *SQLiteUserStorage) GetUser(ctx context.Context, login string) (*User, error) {
	u := s.db.QueryRowContext(ctx, "select id, login, password from users where login = $1", login)

	var uPass string
	var uID int
	var uLogin string
	if scanErr := u.Scan(&uID, &uLogin, &uPass); scanErr != nil {
		if errors.Is(scanErr, sql.ErrNoRows) {
			return nil, nil
		}
		log.Default().Println("[GetUser] query user error: ", scanErr)
		return nil, scanErr
	}

	return &User{
		ID:       uID,
		Login:    uLogin,
		Password: uPass,
	}, nil
}
