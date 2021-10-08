package gorm

import (
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
)

func Test_Transactions(t *testing.T) {
	db, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	type Author struct {
		ID    uint `gorm:"primarykey"`
		Name  string
		Email string
	}

	err = db.AutoMigrate(&Author{})
	require.NoError(t, err)

	err = db.Transaction(func(tx *gorm.DB) error {
		// do some database operations in the transaction (use 'tx' from this point, not 'db')
		if err := tx.Create(&Author{Name: "Giraffe"}).Error; err != nil {
			// return any error will rollback
			return err
		}

		if err := tx.Create(&Author{Name: "Lion"}).Error; err != nil {
			return err
		}

		// return nil will commit the whole transaction
		return nil
	})

	require.NoError(t, err)
}
