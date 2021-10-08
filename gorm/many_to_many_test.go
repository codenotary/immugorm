package gorm

import (
	"github.com/stretchr/testify/require"
	"testing"
)

type Lang struct {
	ID    int
	Name  string
	Users []*Usr `gorm:"many2many:usrs_langs;"`
}
type Usr struct {
	ID        int
	Name      string
	Languages []Lang `gorm:"many2many:usrs_langs;"`
}

func TestManyToMany(t *testing.T) {
	db, close, err := OpenDB()
	require.NoError(t, err)
	defer close()
	err = db.AutoMigrate(&Usr{}, &Lang{})
	require.NoError(t, err)

	user := &Usr{
		Name:      "it-en",
		Languages: []Lang{{Name: "it"}, {Name: "en"}},
	}

	err = db.Create(user).Error
	require.NoError(t, err)

	var user2 Usr
	err = db.Find(&user2, "id = ?", user.ID).Error
	require.NoError(t, err)
	// missing JOIN alias for INNER JOIN
	err = db.Model(&user2).Association("Languages").Find(&user2.Languages)
	require.NoError(t, err)
	require.Len(t, user2.Languages, 2)
}
