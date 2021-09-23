package immudb

import (
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
