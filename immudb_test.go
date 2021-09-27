package immudb

import (
	"bytes"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/client/tokenservice"
	"github.com/codenotary/immudb/pkg/server"
	"github.com/codenotary/immudb/pkg/server/servertest"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"os"
	"testing"
)

func TestSimpleQuery(t *testing.T) {
	options := server.DefaultOptions().WithAuth(true)
	bs := servertest.NewBufconnServer(options)

	bs.Start()
	defer bs.Stop()

	defer os.RemoveAll(options.Dir)
	defer os.Remove(".state-")

	opts := client.DefaultOptions().WithDialOptions(
		&[]grpc.DialOption{grpc.WithContextDialer(bs.Dialer), grpc.WithInsecure()},
	).WithTokenService(tokenservice.NewInmemoryTokenService())

	opts.Username = "immudb"
	opts.Password = "immudb"
	opts.Database = "defaultdb"

	db, err := gorm.Open(Open(opts, ImmuGormConfig{Verify: true}), &gorm.Config{})
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
	require.Equal(t, err, ErrDeleteNotImplemented)
}

type Product struct {
	ID     uint `gorm:"primarykey"`
	Code   string
	Price  uint
	Amount uint
}

func TestTypes(t *testing.T) {
	options := server.DefaultOptions().WithAuth(true)
	bs := servertest.NewBufconnServer(options)

	bs.Start()
	defer bs.Stop()

	defer os.RemoveAll(options.Dir)
	defer os.Remove(".state-")

	opts := client.DefaultOptions().WithDialOptions(
		&[]grpc.DialOption{grpc.WithContextDialer(bs.Dialer), grpc.WithInsecure()},
	).WithTokenService(tokenservice.NewInmemoryTokenService())

	opts.Username = "immudb"
	opts.Password = "immudb"
	opts.Database = "defaultdb"

	db, err := gorm.Open(Open(opts, ImmuGormConfig{Verify: true}), &gorm.Config{})
	require.NoError(t, err)

	// Migrate the schema
	err = db.AutoMigrate(&Entity{})
	require.NoError(t, err)

	// Create
	err = db.Create(&Entity{
		B:    true,
		I32:  int32(3),
		Ui32: uint32(4),
		Bb:   []byte(`content`),
	}).Error
	require.NoError(t, err)

	// Read
	var e Entity
	// find product with integer primary key
	err = db.First(&e, 1).Error
	require.NoError(t, err)
	require.Equal(t, true, e.B)
	require.Equal(t, int8(3), e.I32)
	require.Equal(t, uint8(4), e.Ui32)
	require.True(t, bytes.Equal([]byte(`content`), e.Bb))
}

type Entity struct {
	ID   uint `gorm:"primarykey"`
	B    bool
	I32  int32
	Ui32 uint32
	Bb   []byte
}

func TestTimeTravelQuery(t *testing.T) {
	options := server.DefaultOptions().WithAuth(true)
	bs := servertest.NewBufconnServer(options)

	bs.Start()
	defer bs.Stop()

	defer os.RemoveAll(options.Dir)
	defer os.Remove(".state-")

	opts := client.DefaultOptions().WithDialOptions(
		&[]grpc.DialOption{grpc.WithContextDialer(bs.Dialer), grpc.WithInsecure()},
	).WithTokenService(tokenservice.NewInmemoryTokenService())

	opts.Username = "immudb"
	opts.Password = "immudb"
	opts.Database = "defaultdb"

	db, err := gorm.Open(Open(opts, ImmuGormConfig{Verify: false}), &gorm.Config{})
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
	var entityTT Entity
	err = db.Clauses(BeforeTx(9)).Last(&entityTT, 1).Error
	require.NoError(t, err)
	require.Equal(t, entityTT.Ui32, uint32(5))
	require.Equal(t, entityTT.I32, int32(5))
	err = db.Clauses(BeforeTx(8)).Last(&entityTT, 1).Error
	require.NoError(t, err)
	require.Equal(t, entityTT.Ui32, uint32(4))
	require.Equal(t, entityTT.I32, int32(4))
	err = db.Clauses(BeforeTx(5)).Last(&entityTT, 1).Error
	require.NoError(t, err)
	require.Equal(t, entityTT.Ui32, uint32(1))
	require.Equal(t, entityTT.I32, int32(1))
}
