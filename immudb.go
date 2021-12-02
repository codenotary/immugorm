package immudb

import (
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/stdlib"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

const DriverName = "immudb"

type ImmuGormConfig struct {
	Verify bool
}

type Dialector struct {
	DriverName string
	opts       *client.Options
	cfg        ImmuGormConfig
}

func Open(opts *client.Options, cfg ImmuGormConfig) gorm.Dialector {
	if opts == nil {
		opts = client.DefaultOptions()
	}
	return &Dialector{
		DriverName: DriverName,
		opts:       opts,
		cfg:        cfg,
	}
}

func (dialector Dialector) Name() string {
	return "immudb"
}

func (dialector Dialector) Initialize(db *gorm.DB) (err error) {
	if dialector.DriverName == "" {
		dialector.DriverName = DriverName
	}

	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{})

	db.ConnPool = stdlib.OpenDB(dialector.opts)
	for k, v := range dialector.ClauseBuilders() {
		db.ClauseBuilders[k] = v
	}

	db.Config.SkipDefaultTransaction = true
	db.Config.DisableAutomaticPing = true
	db.Config.AllowGlobalUpdate = true

	if dialector.cfg.Verify {
		db.Callback().Query().After("gorm:query").Register("immudb:after_query", dialector.verify)
	}
	return
}

func (dialector Dialector) ClauseBuilders() map[string]clause.ClauseBuilder {
	return map[string]clause.ClauseBuilder{
		"ON CONFLICT": func(c clause.Clause, b clause.Builder) {
			println("do nothing")
		},
	}
}

func (dialector Dialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return nil
}

func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{
		migrator.Migrator{
			Config: migrator.Config{
				DB:                          db,
				Dialector:                   dialector,
				CreateIndexAfterCreateTable: true,
			}},
	}
}

func (dialector Dialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	writer.WriteByte('?')
}

func (dialector Dialector) QuoteTo(writer clause.Writer, str string) {
	writer.WriteString(str)
	return
}

func (dialector Dialector) Explain(sql string, vars ...interface{}) string {
	return ""
}

func (dialector Dialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool:
		return "BOOLEAN"
	case schema.Int, schema.Uint:
		if field.AutoIncrement && field.PrimaryKey {
			return "INTEGER AUTO_INCREMENT"
		} else {
			return "INTEGER"
		}
	case schema.String:
		if field.Size > 0 {
			return fmt.Sprintf("VARCHAR[%d]", field.Size)
		}
		return "VARCHAR"
	case schema.Bytes:
		if field.Size > 0 {
			return fmt.Sprintf("BLOB[%d]", field.Size)
		}
		return "BLOB"
	case schema.Time:
		return "TIMESTAMP"
	}

	return string(field.DataType)
}

func (dialector Dialector) SavePoint(tx *gorm.DB, name string) error {
	return ErrNotImplemented
}

func (dialector Dialector) RollbackTo(tx *gorm.DB, name string) error {
	return ErrNotImplemented
}

func (dialector Dialector) GetImmuclient(db *gorm.DB) (client.ImmuClient, error) {
	sqlDb, err := db.DB()
	if err != nil {
		db.AddError(err)
	}

	dri := sqlDb.Driver()
	conn, err := dri.(*stdlib.Driver).GetNewConnByOptions(context.TODO(), dialector.opts)
	if err != nil {
		return nil, err
	}

	return conn.GetImmuClient(), nil
}

type columnConverter struct{}

func (cc columnConverter) ColumnConverter(idx int) driver.ValueConverter {
	return nil
}

type valueConverter struct{}

func (cc valueConverter) ConvertValue(v interface{}) (driver.Value, error) {
	return nil, nil
}
