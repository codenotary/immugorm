package gorm

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
