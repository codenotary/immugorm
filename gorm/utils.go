package gorm

import (
	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/client/tokenservice"
	"github.com/codenotary/immudb/pkg/server"
	"github.com/codenotary/immudb/pkg/server/servertest"
	immudb "github.com/codenotary/immugorm"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"os"
)

func OpenDB() (*gorm.DB, func(), error) {
	options := server.DefaultOptions().WithAuth(true)
	bs := servertest.NewBufconnServer(options)
	bs.Start()

	opts := client.DefaultOptions().WithDialOptions(
		[]grpc.DialOption{grpc.WithContextDialer(bs.Dialer), grpc.WithInsecure()},
	).WithTokenService(tokenservice.NewInmemoryTokenService())

	opts.Username = "immudb"
	opts.Password = "immudb"
	opts.Database = "defaultdb"

	db, err := gorm.Open(immudb.Open(opts, immudb.ImmuGormConfig{Verify: false}), &gorm.Config{})
	if err != nil {

	}

	close := func() {
		bs.Stop()
		os.RemoveAll(options.Dir)
		os.Remove(".state-")
	}

	return db, close, err
}
