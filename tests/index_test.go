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

func TestIndexes(t *testing.T) {
	type Profile struct {
		ID      uint
		Refer   uint   `gorm:"uniqueIndex"`
		Name    string `gorm:"index;size:252"`
		Content []byte `gorm:"index;size:252"`
		Age     sql.NullInt64
		Email   string `gorm:"size:252;unique_index"`
		Role    string `gorm:"size:252"` // set field size to 252
		//MemberNumber *string `gorm:"unique;not null;size:252"` // set member number to unique and not null not yet supported
		//MemberNumber *string `gorm:"not null;size:252"` // set member number not null
		Num      int    `gorm:"AUTO_INCREMENT"`      // set num to auto incrementable
		Address  string `gorm:"index:addr;size:252"` // create index with name `addr` for address
		IgnoreMe int    `gorm:"-"`                   // ignore this field
	}

	DB, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	if err := DB.AutoMigrate(&Profile{}); err != nil {
		t.Fatalf("Failed to migrate, got error: %v", err)
	}

	//myNumber := "myNumber"

	err = DB.Create(&Profile{
		Refer:   2,
		Name:    "name",
		Content: []byte(`content`),
		Age:     sql.NullInt64{},
		Email:   "my@email.it",
		Role:    "my_role",
		//MemberNumber: &myNumber,
		Address:  "my_address",
		IgnoreMe: 55,
	}).Error
	require.NoError(t, err)
}

func TestCompositeIndex(t *testing.T) {
	type User struct {
		ID     uint
		Name   string `gorm:"index:idx_member;size:252"`
		Number string `gorm:"index:idx_member;size:252"`
	}

	DB, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	if err := DB.AutoMigrate(&User{}); err != nil {
		t.Fatalf("Failed to migrate, got error: %v", err)
	}

	err = DB.Create(&User{
		Name:   "name",
		Number: "number",
	}).Error

	require.NoError(t, err)
}
