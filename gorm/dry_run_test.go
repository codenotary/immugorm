package gorm

import (
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
)

func TestOpenDB(t *testing.T) {
	db, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	type Author struct {
		ID    string
		Name  string
		Email string
	}

	var author Author
	stmt := db.Session(&gorm.Session{DryRun: true}).First(&author, 1).Statement
	sql := stmt.SQL.String()
	require.Equal(t, "SELECT * FROM authors WHERE authors.id = ? ORDER BY authors.id LIMIT 1", sql)
}
