package testutils

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertDbHas(t *testing.T, db *sql.DB, stmt string, params ...any) {
	var rowExists bool

	res := db.QueryRow(
		fmt.Sprintf("SELECT exists(%s)", strings.TrimRight(stmt, ";")), params...,
	)
	err := res.Scan(&rowExists)

	assert.NoError(t, err)
	assert.Truef(t, rowExists, "record not found")
}

func AssertDbMissing(t *testing.T, db *sql.DB, stmt string, params ...any) {
	t.Helper()

	var rowExists bool

	res := db.QueryRow(fmt.Sprintf("SELECT exists(%s)", stmt), params...)
	err := res.Scan(&rowExists)

	assert.NoError(t, err)
	assert.Falsef(t, rowExists, "record found")
}
