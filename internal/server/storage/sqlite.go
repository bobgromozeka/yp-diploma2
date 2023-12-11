package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

	return err
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

	return err
}

func (s *SQLiteDataKeeperStorage) CreateCard(ctx context.Context, params CreateCardParams) error {
	_, err := s.db.ExecContext(ctx, "insert into cards (user_id, name, number, valid_through_month, valid_through_year, cvv, description) values ($1, $2, $3, $4, $5, $6, $7)",
		params.UserID, params.Name, params.Number, params.ValidThroughMonth, params.ValidThroughYear, params.Cvv, params.Description,
	)

	return err
}

func (s *SQLiteDataKeeperStorage) GetAllCards(ctx context.Context, userID int) ([]Card, error) {
	c, err := s.db.QueryContext(ctx, "select * from cards where user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	var cards []Card
	for c.Next() {
		var card Card
		if scanErr := c.Scan(&card.ID, &card.UserID, &card.Name, &card.Number, &card.ValidThroughMonth, &card.ValidThroughYear, &card.CVV, &card.Description); scanErr != nil {
			return nil, scanErr
		}

		cards = append(cards, card)
	}

	return cards, nil
}

func (s *SQLiteDataKeeperStorage) RemoveCard(ctx context.Context, userID int, ID int) error {
	_, err := s.db.ExecContext(ctx, "delete from cards where user_id = $1 and id = $2", userID, ID)

	return err
}

func (s *SQLiteDataKeeperStorage) CreateBin(ctx context.Context, userID int, name string, data []byte, description *string) error {
	fmt.Println(data)
	_, err := s.db.ExecContext(ctx, "insert into bins (user_id, name, data, description) values ($1, $2, $3, $4)",
		userID, name, data, description)

	return err
}

func (s *SQLiteDataKeeperStorage) GetAllBins(ctx context.Context, userID int) ([]Bin, error) {
	b, err := s.db.QueryContext(ctx, "select * from bins where user_id = $1", userID)
	if err != nil {
		return nil, err
	}

	var bins []Bin
	for b.Next() {
		var bin Bin
		if scanErr := b.Scan(&bin.ID, &bin.UserID, &bin.Name, &bin.Data, &bin.Description); scanErr != nil {
			return nil, scanErr
		}

		bins = append(bins, bin)
	}

	return bins, nil
}

func (s *SQLiteDataKeeperStorage) RemoveBin(ctx context.Context, userID, ID int) error {
	_, err := s.db.ExecContext(ctx, "delete from bins where user_id = $1 and ID = $2", userID, ID)

	return err
}
