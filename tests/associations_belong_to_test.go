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
	"gorm.io/gorm"
	"testing"
)

func Test_BelongsTo(t *testing.T) {
	db, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	type Company struct {
		gorm.Model
		Name string
	}

	type User struct {
		gorm.Model
		Name      string
		CompanyID int
		Company   Company
		ManagerID int
		Manager   *User
	}

	db.AutoMigrate(&Company{}, &User{})

	company1 := &Company{
		Name: "MyCompany",
	}

	user := &User{
		Name:    "user1",
		Company: *company1,
	}

	err = db.Create(user).Error
	require.NoError(t, err)

	var newUser User
	err = db.Where("id = ?", user.ID).First(&newUser).Error
	require.NoError(t, err)

	userWithManager := &User{
		Name:      "user2",
		CompanyID: user.CompanyID,
		Manager: &User{
			Name: "manager",
		},
	}

	err = db.Create(userWithManager).Error
	require.NoError(t, err)

	var user2 User
	db.Find(&user2, "id = ?", userWithManager.ID)
	pointerOfUser := &user2
	if err := db.Model(&pointerOfUser).Association("Company").Find(&user2.Company); err != nil {
		t.Errorf("failed to query users, got error %#v", err)
	}
	user2.Manager = &User{}
	db.Model(&user2).Association("Manager").Find(user2.Manager)

	require.NotNil(t, user2)

	// Append
	err = db.Create(&User{
		Name: "user-append",
	}).Error
	require.NoError(t, err)

	err = db.Create(&Company{
		Name: "company-append",
	}).Error
	require.NoError(t, err)

	var userAppend User
	err = db.Where("name = ?", "user-append").First(&userAppend).Error
	require.NoError(t, err)

	companyAppend := &Company{
		Name: "company-append",
	}

	err = db.Model(&userAppend).Association("Company").Append(&companyAppend)
	require.NoError(t, err)

	companyReplaced := &Company{
		Name: "company-replaced",
	}
	err = db.Model(&userAppend).Association("Company").Replace(&companyReplaced)
	require.NoError(t, err)

}
