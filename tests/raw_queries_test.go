/*
Copyright 2021 CodeNotary, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tests

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
	require.Equal(t, "michele", result.Name)
	require.Equal(t, 40, result.Age)
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
	err = db.Where("name = @name1 OR name = @name2", sql.Named("name1", "michele"), sql.Named("name2", "jhon")).Find(&user).Error
	require.NoError(t, err)
}
