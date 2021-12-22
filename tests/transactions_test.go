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

	var author Author
	err = db.Find(&author, "Name = ?", "Giraffe").Error
	require.NoError(t, err)
}
