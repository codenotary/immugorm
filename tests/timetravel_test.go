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
	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/server"
	"github.com/codenotary/immudb/pkg/server/servertest"
	immugorm "github.com/codenotary/immugorm"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"os"
	"testing"
)

func TestTimeTravelQuery(t *testing.T) {
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

	db, err := gorm.Open(immugorm.OpenWithOptions(opts, &immugorm.ImmuGormConfig{Verify: false}), &gorm.Config{})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&Entity{})
	require.NoError(t, err)

	err = db.Create(&Entity{
		B:    true,
		I32:  0,
		Ui32: 0,
		Bb:   []byte(`control`),
	}).Error
	var e Entity
	err = db.First(&e, 1).Error

	// Create
	for i := 1; i <= 10; i++ {
		fi := int32(i)
		si := uint32(i)
		err = db.Model(&e).Updates(map[string]interface{}{"B": true, "I32": fi, "Ui32": si, "Bb": []byte(`control`)}).Error
		require.NoError(t, err)
	}

	var entity Entity
	err = db.Last(&entity, 1).Error
	require.NoError(t, err)
	var entityTT Entity
	err = db.Clauses(immugorm.BeforeTx(9)).Last(&entityTT, 1).Error
	require.NoError(t, err)
	require.Equal(t, entityTT.Ui32, uint32(5))
	require.Equal(t, entityTT.I32, int32(5))
	err = db.Clauses(immugorm.BeforeTx(8)).Last(&entityTT, 1).Error
	require.NoError(t, err)
	require.Equal(t, entityTT.Ui32, uint32(4))
	require.Equal(t, entityTT.I32, int32(4))
	err = db.Clauses(immugorm.BeforeTx(5)).Last(&entityTT, 1).Error
	require.NoError(t, err)
	require.Equal(t, entityTT.Ui32, uint32(1))
	require.Equal(t, entityTT.I32, int32(1))
}
