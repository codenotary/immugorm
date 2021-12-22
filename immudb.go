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

package immudb

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/stdlib"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
	"regexp"
)

const DriverName = "immudb"

type ImmuGormConfig struct {
	Verify bool
}

type Dialector struct {
	DriverName string
	opts       *client.Options
	cfg        *ImmuGormConfig
	Conn       gorm.ConnPool
	DSN        string
}

func Open(opts *client.Options, cfg *ImmuGormConfig) gorm.Dialector {
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

	callbacks.RegisterDefaultCallbacks(db, &callbacks.Config{
		CreateClauses: []string{"INSERT", "VALUES", "ON CONFLICT"},
		UpdateClauses: []string{"UPDATE", "SET", "WHERE"},
		DeleteClauses: []string{"DELETE", "FROM", "WHERE"},
		QueryClauses:  []string{"SELECT", "FROM", "WHERE", "GROUP BY", "ORDER BY", "LIMIT" /*, "FOR"*/},
	})

	if dialector.Conn != nil {
		db.ConnPool = dialector.Conn
	} else if dialector.opts != nil {
		db.ConnPool = stdlib.OpenDB(dialector.opts)
	} else if dialector.DSN != "" {
		db.ConnPool, err = sql.Open(dialector.DriverName, dialector.DSN)
	}

	if db.ConnPool == nil {
		return fmt.Errorf("failed to open connection")
	}

	for k, v := range dialector.ClauseBuilders() {
		db.ClauseBuilders[k] = v
	}

	db.Config.SkipDefaultTransaction = true

	if dialector.cfg.Verify {
		db.Callback().Query().After("gorm:query").Register("immudb:after_query", dialector.verify)
	}
	return
}

func (dialector Dialector) ClauseBuilders() map[string]clause.ClauseBuilder {
	return map[string]clause.ClauseBuilder{
		"ON CONFLICT": func(c clause.Clause, builder clause.Builder) {
			_, ok := c.Expression.(clause.OnConflict)
			if !ok {
				c.Build(builder)
				return
			}
			builder.WriteString("ON CONFLICT DO NOTHING")
			return
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

var numericPlaceholder = regexp.MustCompile("\\$(\\d+)")

func (dialector Dialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, numericPlaceholder, `'`, vars...)
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
