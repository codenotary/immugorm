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
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"

	"gorm.io/gorm"
)

func TestEmbeddedStruct(t *testing.T) {
	type ReadOnly struct {
		ReadOnly *bool
	}

	type BasePost struct {
		Id    int64
		Title string
		URL   string
		ReadOnly
	}

	type Author struct {
		ID    string
		Name  string
		Email string
	}

	type HNPost struct {
		BasePost
		Author  // Embedded struct
		Upvotes int32
	}

	type EngadgetPost struct {
		BasePost BasePost `gorm:"Embedded"`
		Author   Author   `gorm:"Embedded"` // Embedded struct
		ImageUrl string
	}

	DB, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	if err := DB.Migrator().AutoMigrate(&HNPost{}, &EngadgetPost{}); err != nil {
		t.Fatalf("failed to auto migrate, got error: %v", err)
	}

	stmt := gorm.Statement{DB: DB}
	if err := stmt.Parse(&EngadgetPost{}); err != nil {
		t.Fatalf("failed to parse embedded struct")
	} else if len(stmt.Schema.PrimaryFields) != 1 {
		t.Errorf("should have only one primary field with embedded struct, but got %v", len(stmt.Schema.PrimaryFields))
	}

	// save embedded struct
	err = DB.Save(&HNPost{BasePost: BasePost{Title: "news"}}).Error
	require.NoError(t, err)
	/*
		DB.Save(&HNPost{BasePost: BasePost{Title: "hn_news"}})
		var news HNPost
		if err := DB.First(&news, "title = ?", "hn_news").Error; err != nil {
			t.Errorf("no error should happen when query with embedded struct, but got %v", err)
		} else if news.Title != "hn_news" {
			t.Errorf("embedded struct's value should be scanned correctly")
		}

		DB.Save(&EngadgetPost{BasePost: BasePost{Title: "engadget_news"}})
		var egNews EngadgetPost
		if err := DB.First(&egNews, "title = ?", "engadget_news").Error; err != nil {
			t.Errorf("no error should happen when query with embedded struct, but got %v", err)
		} else if egNews.BasePost.Title != "engadget_news" {
			t.Errorf("embedded struct's value should be scanned correctly")
		}*/
}

func TestEmbeddedPointerTypeStruct(t *testing.T) {
	type BasePost struct {
		Id    int64
		Title string
		URL   string
	}

	type HNPost struct {
		*BasePost
		Upvotes int32
	}

	DB, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	if err := DB.Migrator().AutoMigrate(&HNPost{}); err != nil {
		t.Fatalf("failed to auto migrate, got error: %v", err)
	}

	DB.Create(&HNPost{BasePost: &BasePost{Title: "embedded_pointer_type"}})

	var hnPost HNPost
	if err := DB.First(&hnPost, "title = ?", "embedded_pointer_type").Error; err != nil {
		t.Errorf("No error should happen when find embedded pointer type, but got %v", err)
	}

	if hnPost.Title != "embedded_pointer_type" {
		t.Errorf("Should find correct value for embedded pointer type")
	}
}

type Content struct {
	Content interface{} `gorm:"type:String"`
}

func (c Content) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (c *Content) Scan(src interface{}) error {
	b, ok := src.([]byte)
	if !ok {
		return errors.New("Embedded.Scan byte assertion failed")
	}

	var value Content
	if err := json.Unmarshal(b, &value); err != nil {
		return err
	}

	*c = value

	return nil
}

// @todo uncomment when timestamp will be available
func _TestEmbeddedScanValuer(t *testing.T) {
	type HNPost struct {
		gorm.Model
		Content
	}

	DB, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	DB.Migrator().DropTable(&HNPost{})
	if err := DB.Migrator().AutoMigrate(&HNPost{}); err != nil {
		t.Fatalf("failed to auto migrate, got error: %v", err)
	}

	hnPost := HNPost{Content: Content{Content: "hello world"}}

	if err := DB.Create(&hnPost).Error; err != nil {
		t.Errorf("Failed to create got error %v", err)
	}
}

// @todo uncomment when timestamp and foreign keys will be available
func _TestEmbeddedRelations(t *testing.T) {
	type AdvancedUser struct {
		Usr      `gorm:"embedded"`
		Advanced bool
	}

	DB, close, err := OpenDB()
	require.NoError(t, err)
	defer close()

	if err := DB.AutoMigrate(&AdvancedUser{}); err != nil {
		if DB.Dialector.Name() != "sqlite" {
			t.Errorf("Failed to auto migrate advanced user, got error %v", err)
		}
	}
}
