package gorm

import (
	"database/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRawQuery(t *testing.T) {

	type Result struct {
		ID   int
		Name string
		Age  int
	}

	db, close, err := OpenDB()
	require.NoError(t, err)
	defer close()
	var result Result
	err = db.Exec("CREATE TABLE users (id INTEGER AUTO_INCREMENT,name VARCHAR,age INTEGER,PRIMARY KEY ID);").Error
	require.NoError(t, err)
	err = db.Exec("INSERT INTO users (name ,age) VALUES ('michele', 40)").Error
	require.NoError(t, err)
	err = db.Raw("SELECT id, name, age FROM users WHERE name = ?", "michele").Scan(&result).Error
	require.NoError(t, err)
}

func TestNamedArguments(t *testing.T) {

	type User struct {
		ID   int
		Name string
		Age  int
	}

	db, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	err = db.Exec("CREATE TABLE users (id INTEGER AUTO_INCREMENT,name VARCHAR,age INTEGER,PRIMARY KEY ID);").Error
	require.NoError(t, err)
	err = db.Exec("INSERT INTO users (name ,age) VALUES ('michele', 40)").Error
	require.NoError(t, err)

	var user User
	db.Where("name = @name1 OR name = @name2", sql.Named("name1", "michele"), sql.Named("name2", "jhon")).Find(&user)
	require.NoError(t, err)
}
