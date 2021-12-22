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

	err = db.Delete(&product).Error
	require.NoError(t, err)

	var prods []Product
	// related to https://gorm.io/docs/delete.html
	db.Unscoped().Where("Code = 'D42'").Find(&prods)
	require.Equal(t, 1, len(prods))
	db.Where("Code = 'D42'").Find(&prods)
	require.Equal(t, 0, len(prods))
}
