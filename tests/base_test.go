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
	"bytes"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/server"
	"github.com/codenotary/immudb/pkg/server/servertest"
	immugorm "github.com/codenotary/immugorm"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"os"
	"testing"
	"time"
)

func TestBase(t *testing.T) {
	type Product struct {
		ID     uint `gorm:"primarykey"`
		Code   string
		Price  uint
		Amount uint
	}

	options := server.DefaultOptions()
	bs := servertest.NewBufconnServer(options)

	bs.Start()
	defer bs.Stop()

	defer os.RemoveAll(options.Dir)
	defer os.Remove(".state-")

	opts := client.DefaultOptions().WithDialOptions(
		[]grpc.DialOption{grpc.WithContextDialer(bs.Dialer), grpc.WithInsecure()},
	)

	opts.Username = "immudb"
	opts.Password = "immudb"
	opts.Database = "defaultdb"

	db, err := gorm.Open(immugorm.Open(opts, &immugorm.ImmuGormConfig{Verify: true}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&Product{})
	require.NoError(t, err)

	// Create
	err = db.Create(&Product{Code: "D43", Price: 100, Amount: 500}).Error
	require.NoError(t, err)

	// Read
	var product Product
	// find product with integer primary key
	err = db.First(&product, 1).Error
	require.NoError(t, err)

	// find product with code D42
	err = db.First(&product, "code = ?", "D43").Error
	require.NoError(t, err)

	// Update - update product's price to 200
	err = db.Model(&product).Update("Price", 888).Error
	require.NoError(t, err)

	// Update - update multiple fields
	err = db.Model(&product).Updates(Product{Price: 200, Code: "F42"}).Error
	require.NoError(t, err)

	err = db.Model(&product).Updates(map[string]interface{}{"Price": 200, "Code": "F42"}).Error
	require.NoError(t, err)

	// Delete - delete product
	err = db.Delete(&product, 1).Error
	require.NoError(t, err)
}

func TestMigrationWithPreviousData(t *testing.T) {
	type Product struct {
		ID     uint `gorm:"primarykey"`
		Code   string
		Price  uint
		Amount uint
	}

	options := server.DefaultOptions()
	bs := servertest.NewBufconnServer(options)

	bs.Start()
	defer bs.Stop()

	defer os.RemoveAll(options.Dir)
	defer os.Remove(".state-")

	opts := client.DefaultOptions().WithDialOptions(
		[]grpc.DialOption{grpc.WithContextDialer(bs.Dialer), grpc.WithInsecure()},
	)

	opts.Username = "immudb"
	opts.Password = "immudb"
	opts.Database = "defaultdb"

	db, err := gorm.Open(immugorm.Open(opts, &immugorm.ImmuGormConfig{Verify: false}), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&Product{})
	require.NoError(t, err)

	err = db.Create(&Product{Code: "D43", Price: 100, Amount: 500}).Error
	require.NoError(t, err)

	err = db.AutoMigrate(&Product{})
	require.NoError(t, err)
}

type Entity struct {
	ID   uint `gorm:"primarykey"`
	B    bool
	I32  int32
	Ui32 uint32
	Bb   []byte
	Time time.Time
}

func TestTypes(t *testing.T) {

	options := server.DefaultOptions()
	bs := servertest.NewBufconnServer(options)

	bs.Start()
	defer bs.Stop()

	defer os.RemoveAll(options.Dir)
	defer os.Remove(".state-")

	opts := client.DefaultOptions().WithDialOptions(
		[]grpc.DialOption{grpc.WithContextDialer(bs.Dialer), grpc.WithInsecure()},
	)

	opts.Username = "immudb"
	opts.Password = "immudb"
	opts.Database = "defaultdb"

	db, err := gorm.Open(immugorm.Open(opts, &immugorm.ImmuGormConfig{Verify: false}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&Entity{})
	require.NoError(t, err)

	now := time.Now()
	// Create
	err = db.Create(&Entity{
		B:    true,
		I32:  int32(3),
		Ui32: uint32(4),
		Bb:   []byte(`content`),
		Time: now,
	}).Error
	require.NoError(t, err)

	var e Entity
	err = db.First(&e, 1).Error
	require.NoError(t, err)
	require.Equal(t, true, e.B)
	require.Equal(t, int32(3), e.I32)
	require.Equal(t, uint32(4), e.Ui32)
	require.WithinDuration(t, now, e.Time, 1*time.Millisecond)
	require.True(t, bytes.Equal([]byte(`content`), e.Bb))
}
