package gorm

import (
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"testing"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func TestTimestamp(t *testing.T) {
	db, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	err = db.AutoMigrate(&Product{})
	require.NoError(t, err)

	err = db.Debug().Create(&Product{Code: "D42", Price: 100}).Error
	require.NoError(t, err)

	var product Product
	err = db.First(&product, 1).Error
	require.NoError(t, err)

	err = db.Delete(&product, 1).Error
	require.NoError(t, err)

	var prods []Product
	// related to https://gorm.io/docs/delete.html
	db.Unscoped().Where("Code = 'D42'").Find(&prods)
	require.Equal(t, 1, len(prods))
	db.Where("Code = 'D42'").Find(&prods)
	require.Equal(t, 0, len(prods))
}
