package gorm

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_BelongsTo(t *testing.T) {
	db, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	type Company struct {
		ID   int
		Name string
	}

	type User struct {
		ID        uint `gorm:"primarykey"`
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
	/// NOT WORKING MISSING UPSERT
	companyReplaced := &Company{
		Name: "company-replaced",
	}
	err = db.Model(&userAppend).Association("Company").Replace(&companyReplaced)
	require.NoError(t, err)

	var userWithCompanyReplaced User
	err = db.Where("name = ?", "user-append").First(&userWithCompanyReplaced).Error
	require.NoError(t, err)
	require.Equal(t, userWithCompanyReplaced.Company.Name, "company-replaced")
}
