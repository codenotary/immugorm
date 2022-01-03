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

type Lang struct {
	ID    uint
	Name  string
	Users []*Usr `gorm:"many2many:usrs_langs;"`
}
type Usr struct {
	ID        uint
	Name      string
	Languages []*Lang `gorm:"many2many:usrs_langs;"`
}

func TestManyToMany(t *testing.T) {
	db, close, err := OpenDB()
	require.NoError(t, err)
	defer close()
	err = db.AutoMigrate(&Usr{}, &Lang{})
	require.NoError(t, err)

	var languages = []Lang{
		{Name: "language-many2many-append-1-1"},
		{Name: "language-many2many-append-2-1"},
	}
	db.Create(&languages)

	user := &Usr{
		Name: "it-en-fr",
	}

	err = db.Create(user).Error
	require.NoError(t, err)

	var user2 Usr
	err = db.Find(&user2, "id = ?", user.ID).Error
	require.NoError(t, err)

	err = db.Model(&user).Association("Languages").Append(&languages)
	require.NoError(t, err)

	db.Model(&user).Association("Languages").Append(&Lang{
		Name: "it",
	})
	db.Model(&user).Association("Languages").Append(&Lang{
		Name: "en",
	})
	db.Model(&user).Association("Languages").Append(&Lang{
		Name: "fr",
	})

	var user3 Usr
	err = db.Find(&user3, "id = ?", user.ID).Error
	require.NoError(t, err)

	err = db.Model(&user2).Association("Languages").Find(&user2.Languages)
	require.NoError(t, err)
	require.Len(t, user2.Languages, 2)
}
