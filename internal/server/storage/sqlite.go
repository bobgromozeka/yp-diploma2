package storage

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/mattn/go-sqlite3"
)

type SQLiteStoragesFactory struct {
	db *sql.DB
}

func NewSQLiteStoragesFactory(db *sql.DB) *SQLiteStoragesFactory {
	return &SQLiteStoragesFactory{
		db: db,
	}
}

func (f *SQLiteStoragesFactory) CreateUserStorage() UserStorage {
	return newSQLiteUserStorage(f.db)
}

func (f *SQLiteStoragesFactory) CreateDataKeeperStorage() DataKeeperStorage {
	return newSQLiteDataKeeperStorage(f.db)
}

type SQLiteUserStorage struct {
	db *sql.DB
}

func newSQLiteUserStorage(db *sql.DB) *SQLiteUserStorage {
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

type SQLiteDataKeeperStorage struct {
	db *sql.DB
}

func newSQLiteDataKeeperStorage(db *sql.DB) *SQLiteDataKeeperStorage {
	return &SQLiteDataKeeperStorage{
		db: db,
	}
}

func (s *SQLiteDataKeeperStorage) CreatePasswordPair(ctx context.Context, userID int, login, password string, description *string) error {
	_, err := s.db.ExecContext(
		ctx, "insert into password_pairs (user_id, login, password, description) values ($1, $2, $3, $4)", userID, login, password, description,
	)

	return err
}

func (s *SQLiteDataKeeperStorage) GetAllPasswordPairs(ctx context.Context, userID int) ([]PasswordPair, error) {
	p, err := s.db.QueryContext(ctx, "select * from password_pairs where user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	var passwordPairs []PasswordPair
	for p.Next() {
		var pp PasswordPair
		if scanErr := p.Scan(&pp.ID, &pp.UserID, &pp.Login, &pp.Password, &pp.Description); scanErr != nil {
			return nil, scanErr
		}

		passwordPairs = append(passwordPairs, pp)
	}

	return passwordPairs, nil
}

func (s *SQLiteDataKeeperStorage) RemovePasswordPair(ctx context.Context, userID int, ID int) error {
	_, err := s.db.ExecContext(ctx, "delete from password_pairs where user_id = $1 and id = $2", userID, ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *SQLiteDataKeeperStorage) CreateText(ctx context.Context, userID int, name, text string, description *string) error {
	_, err := s.db.ExecContext(
		ctx, "insert into texts (user_id, name, text, description) values ($1, $2, $3, $4)", userID, name, text, description,
	)

	return err
}

func (s *SQLiteDataKeeperStorage) GetAllTexts(ctx context.Context, userID int) ([]Text, error) {
	t, err := s.db.QueryContext(ctx, "select * from texts where user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	var texts []Text
	for t.Next() {
		var text Text
		if scanErr := t.Scan(&text.ID, &text.UserID, &text.Name, &text.T, &text.Description); scanErr != nil {
			return nil, scanErr
		}

		texts = append(texts, text)
	}

	return texts, nil
}

func (s *SQLiteDataKeeperStorage) RemoveText(ctx context.Context, userID int, ID int) error {
	_, err := s.db.ExecContext(ctx, "delete from texts where user_id = $1 and id = $2", userID, ID)
	if err != nil {
		return err
	}

	return nil
}
