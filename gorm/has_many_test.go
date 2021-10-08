package gorm

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
