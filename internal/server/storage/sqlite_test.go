package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/bobgromozeka/yp-diploma2/pkg/testutils"
)

type testConn struct {
	conn *sql.DB
	name string
}

var connection testConn

func getConnection(t *testing.T) *sql.DB {
	if connection.conn == nil {
		connection.name = uuid.New().String()
		conn, connErr := sql.Open("sqlite3", connection.name)
		if connErr != nil {
			t.Fatal(connErr)
			return nil
		}

		connection.conn = conn
		Bootstrap(connection.conn)
	}

	return connection.conn
}

func cleanupConnection() {
	if connection.conn != nil {
		connection.conn.Close()
		os.Remove(connection.name)
	}
}

func truncateTable(t *testing.T, table string, conn *sql.DB) {
	_, err := conn.Exec(
		fmt.Sprintf("DELETE FROM %s;", table),
	) // It`s not allowed to use params and tables variables like $1 in truncate query
	if err != nil {
		t.Fatalf("Could not connect to database: %s", err)
	}
}

func TestSQLiteStorage(t *testing.T) {
	t.Cleanup(func() {
		cleanupConnection()
	})

	t.Run("SQLiteUserStorage_CreateUser", func(t *testing.T) {
		conn := getConnection(t)
		s := NewSQLiteStoragesFactory(conn).CreateUserStorage()

		err := s.CreateUser(context.Background(), "login", []byte("pass"))
		assert.Nil(t, err)

		testutils.AssertDbHas(t, conn, "select * from users where login = $1", "login")
	})

	t.Run("SQLiteUsersStorage_GetUser", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateUserStorage()

		err := s.CreateUser(context.Background(), "login", []byte("pass"))
		assert.Nil(t, err)

		u, uErr := s.GetUser(context.Background(), "login")

		assert.Nil(t, uErr)
		assert.Equal(t, "login", u.Login)
	})

	t.Run("SQLiteDataKeeperStorage_CreatePasswordPair", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "password_pairs", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()

		err := s.CreatePasswordPair(context.Background(), 1, "login", "pass", nil)
		assert.Nil(t, err)

		testutils.AssertDbHas(t, conn, "select * from password_pairs where user_id = $1 and login = $2 and password = $3", 1, "login", "pass")
	})

	t.Run("SQLiteDataKeeperStorage_GetAllPasswordPairs", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "password_pairs", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()
		_ = s.CreatePasswordPair(context.Background(), 1, "login", "pass", nil)
		_ = s.CreatePasswordPair(context.Background(), 1, "login1", "pass1", nil)

		pps, err := s.GetAllPasswordPairs(context.Background(), 1)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(pps))

		assert.Equal(t, 1, pps[0].UserID)
		assert.Equal(t, "login", pps[0].Login)
		assert.Equal(t, "pass", pps[0].Password)

		assert.Equal(t, 1, pps[1].UserID)
		assert.Equal(t, "login1", pps[1].Login)
		assert.Equal(t, "pass1", pps[1].Password)
	})

	t.Run("SQLiteDataKeeperStorage_RemovePasswordPair", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "password_pairs", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()
		_ = s.CreatePasswordPair(context.Background(), 1, "login", "pass", nil)

		pp, _ := s.GetAllPasswordPairs(context.Background(), 1)

		err := s.RemovePasswordPair(context.Background(), 1, pp[0].ID)

		assert.Nil(t, err)
		testutils.AssertDbMissing(t, conn, "select * from password_pairs where id = $1", pp[0].ID)
	})

	t.Run("SQLiteDataKeeperStorage_CreateText", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "texts", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()

		err := s.CreateText(context.Background(), 1, "name", "text", nil)
		assert.Nil(t, err)

		testutils.AssertDbHas(t, conn, "select * from texts where user_id = $1 and name = $2 and text = $3", 1, "name", "text")
	})

	t.Run("SQLiteDataKeeperStorage_GetAllTexts", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "texts", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()
		_ = s.CreateText(context.Background(), 1, "name", "text", nil)
		_ = s.CreateText(context.Background(), 1, "name1", "text1", nil)

		texts, err := s.GetAllTexts(context.Background(), 1)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(texts))

		assert.Equal(t, 1, texts[0].UserID)
		assert.Equal(t, "name", texts[0].Name)
		assert.Equal(t, "text", texts[0].T)

		assert.Equal(t, 1, texts[1].UserID)
		assert.Equal(t, "name1", texts[1].Name)
		assert.Equal(t, "text1", texts[1].T)
	})

	t.Run("SQLiteDataKeeperStorage_RemoveText", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "texts", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()
		_ = s.CreateText(context.Background(), 1, "name", "text", nil)

		texts, _ := s.GetAllTexts(context.Background(), 1)

		err := s.RemoveText(context.Background(), 1, texts[0].ID)

		assert.Nil(t, err)
		testutils.AssertDbMissing(t, conn, "select * from texts where id = $1", texts[0].ID)
	})

	t.Run("SQLiteDataKeeperStorage_CreateCard", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "cards", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()

		err := s.CreateCard(context.Background(), CreateCardParams{
			UserID:            1,
			Name:              "name",
			Number:            "1234123412341234",
			ValidThroughMonth: 9,
			ValidThroughYear:  25,
			Cvv:               234,
			Description:       nil,
		})
		assert.Nil(t, err)

		testutils.AssertDbHas(t, conn, "select * from cards where user_id = $1 and name = $2 and number = $3", 1, "name", "1234123412341234")
	})

	t.Run("SQLiteDataKeeperStorage_GetAllCards", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "cards", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()
		_ = s.CreateCard(context.Background(), CreateCardParams{
			UserID:            1,
			Name:              "name",
			Number:            "1234123412341234",
			ValidThroughMonth: 9,
			ValidThroughYear:  25,
			Cvv:               234,
			Description:       nil,
		})
		_ = s.CreateCard(context.Background(), CreateCardParams{
			UserID:            1,
			Name:              "name1",
			Number:            "1234123412341234",
			ValidThroughMonth: 9,
			ValidThroughYear:  25,
			Cvv:               234,
			Description:       nil,
		})

		cards, err := s.GetAllCards(context.Background(), 1)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(cards))

		assert.Equal(t, 1, cards[0].UserID)
		assert.Equal(t, "name", cards[0].Name)
		assert.Equal(t, "1234123412341234", cards[0].Number)

		assert.Equal(t, 1, cards[1].UserID)
		assert.Equal(t, "name1", cards[1].Name)
		assert.Equal(t, "1234123412341234", cards[1].Number)
	})

	t.Run("SQLiteDataKeeperStorage_RemoveText", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "cards", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()
		_ = s.CreateCard(context.Background(), CreateCardParams{
			UserID:            1,
			Name:              "name",
			Number:            "1234123412341234",
			ValidThroughMonth: 9,
			ValidThroughYear:  25,
			Cvv:               234,
			Description:       nil,
		})

		cards, _ := s.GetAllCards(context.Background(), 1)

		err := s.RemoveCard(context.Background(), 1, cards[0].ID)

		assert.Nil(t, err)
		testutils.AssertDbMissing(t, conn, "select * from cards where id = $1", cards[0].ID)
	})

	t.Run("SQLiteDataKeeperStorage_CreateBin", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "bins", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()

		err := s.CreateBin(context.Background(), 1, "name", []byte("data"), nil)
		assert.Nil(t, err)

		testutils.AssertDbHas(t, conn, "select * from bins where user_id = $1 and name = $2 and data = $3", 1, "name", []byte("data"))
	})

	t.Run("SQLiteDataKeeperStorage_GetAllBins", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "bins", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()
		_ = s.CreateBin(context.Background(), 1, "name", []byte("data"), nil)
		_ = s.CreateBin(context.Background(), 1, "name1", []byte("data"), nil)

		bins, err := s.GetAllBins(context.Background(), 1)

		assert.Nil(t, err)
		assert.Equal(t, 2, len(bins))

		assert.Equal(t, 1, bins[0].UserID)
		assert.Equal(t, "name", bins[0].Name)
		assert.Equal(t, []byte("data"), bins[0].Data)

		assert.Equal(t, 1, bins[1].UserID)
		assert.Equal(t, "name1", bins[1].Name)
		assert.Equal(t, []byte("data"), bins[1].Data)
	})

	t.Run("SQLiteDataKeeperStorage_RemoveBin", func(t *testing.T) {
		conn := getConnection(t)
		truncateTable(t, "bins", conn)
		truncateTable(t, "users", conn)

		s := NewSQLiteStoragesFactory(conn).CreateDataKeeperStorage()
		_ = s.CreateBin(context.Background(), 1, "name", []byte("data"), nil)

		bins, _ := s.GetAllBins(context.Background(), 1)

		err := s.RemoveBin(context.Background(), 1, bins[0].ID)

		assert.Nil(t, err)
		testutils.AssertDbMissing(t, conn, "select * from bins where id = $1", bins[0].ID)
	})
}
