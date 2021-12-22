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
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHasMany(t *testing.T) {

	db, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	type Pet struct {
		ID     uint `gorm:"primarykey"`
		UserID *uint
		Name   string
	}

	type User struct {
		ID        uint `gorm:"primarykey"`
		Name      string
		CompanyID int
		ManagerID int
		Manager   *User
		Pets      []*Pet
	}

	db.AutoMigrate(&User{}, &Pet{})

	user := &User{
		Name: "userWithPets",
		Pets: []*Pet{{Name: "pet1"}, {Name: "pet2"}},
	}

	err = db.Create(user).Error
	require.NoError(t, err)

	var newUserWithPets User
	err = db.Where("name = ?", "userWithPets").First(&newUserWithPets).Error
	require.NoError(t, err)

	var user2 User
	db.Find(&user2, "id = ?", user.ID)
	err = db.Model(&user2).Association("Pets").Find(&user2.Pets)
	require.NoError(t, err)
	require.Len(t, user2.Pets, 2)
}
