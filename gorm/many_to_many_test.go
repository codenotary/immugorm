package gorm

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
		//Languages: []Lang{{Name: "it"}, {Name: "en"}, {Name: "fr"}},
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

	/*langsRes := db.Find(&Lang{})
	require.NotNil(t, langsRes)*/

	/*var langs []*Lang
	langsRes.Scan(langs)*/
	var user3 Usr
	err = db.Find(&user3, "id = ?", user.ID).Error
	require.NoError(t, err)
	// missing JOIN alias for INNER JOIN
	err = db.Model(&user2).Association("Languages").Find(&user2.Languages)
	require.NoError(t, err)
	require.Len(t, user2.Languages, 2)
}
