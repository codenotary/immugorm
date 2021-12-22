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

func Test_HasOne(t *testing.T) {
	type CreditCard struct {
		ID     int
		Name   string
		UserID uint
	}

	type User struct {
		ID         int
		Name       string
		CreditCard *CreditCard
	}

	db, close, err := OpenDB()
	require.NoError(t, err)
	defer close()
	db.AutoMigrate(&CreditCard{}, &User{})

	cc := &CreditCard{
		Name: "MyCC",
	}

	user := &User{
		Name:       "myUser",
		CreditCard: cc,
	}

	err = db.Create(user).Error
	require.NoError(t, err)

	var user2 User
	err = db.Find(&user2, "id = ?", user.ID).Error
	require.NoError(t, err)
	err = db.Model(&user2).Association("CreditCard").Find(&user2.CreditCard)
	require.NoError(t, err)
	require.Equal(t, user2.CreditCard.Name, "MyCC")
}
