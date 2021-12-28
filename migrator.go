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
	"errors"
	"fmt"
	"github.com/codenotary/immudb/pkg/client"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"strings"
)

type Migrator struct {
	migrator.Migrator
}

func (m Migrator) CreateTable(values ...interface{}) error {
	for _, value := range m.ReorderModels(values, false) {
		tx := m.DB.Session(&gorm.Session{})
		if err := m.RunWithValue(value, func(stmt *gorm.Statement) (errr error) {
			var (
				createTableSQL          = "CREATE TABLE ? ("
				values                  = []interface{}{m.CurrentTable(stmt)}
				hasPrimaryKeyInDataType bool
			)

			for _, dbName := range stmt.Schema.DBNames {
				field := stmt.Schema.FieldsByDBName[dbName]
				if !field.IgnoreMigration {
					createTableSQL += "? ?"
					hasPrimaryKeyInDataType = hasPrimaryKeyInDataType || strings.Contains(strings.ToUpper(string(field.DataType)), "PRIMARY KEY")
					values = append(values, clause.Column{Name: dbName}, m.DB.Migrator().FullDataTypeOf(field))
					createTableSQL += ","
				}
			}

			if !hasPrimaryKeyInDataType && len(stmt.Schema.PrimaryFields) > 0 {
				createTableSQL += "PRIMARY KEY ?,"
				values = append(values, clause.Column{Name: stmt.Schema.PrimaryFields[0].DBName})
			}

			for _, idx := range stmt.Schema.ParseIndexes() {
				if m.CreateIndexAfterCreateTable {
					defer func(value interface{}, name string) {
						if errr == nil {
							errr = tx.Migrator().CreateIndex(value, name)
						}
					}(value, idx.Name)
				} else {
					if idx.Class != "" {
						createTableSQL += idx.Class + " "
					}
					createTableSQL += "INDEX ? ?"

					if idx.Option != "" {
						createTableSQL += " " + idx.Option
					}

					createTableSQL += ","
					values = append(values, clause.Expr{SQL: idx.Name}, tx.Migrator().(migrator.BuildIndexOptionsInterface).BuildIndexOptions(idx.Fields, stmt))
				}
			}

			createTableSQL = strings.TrimSuffix(createTableSQL, ",")

			createTableSQL += ")"

			if tableOption, ok := m.DB.Get("gorm:table_options"); ok {
				createTableSQL += fmt.Sprint(tableOption)
			}

			errr = tx.Exec(createTableSQL, values...).Error
			return errr
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m Migrator) HasTable(value interface{}) bool {
	var count int
	m.Migrator.RunWithValue(value, func(stmt *gorm.Statement) error {
		return executeOnImmuClient(m.DB, func(ic client.ImmuClient) error {
			_, er := ic.DescribeTable(context.Background(), stmt.Table)
			if er != nil {
				st, ok := status.FromError(er)
				if ok && st.Message() == "table does not exist" {
					count = 0
					return nil
				}
				m.DB.AddError(er)
				return nil
			}
			count = 1
			return nil
		})
	})
	return count > 0
}

func (m Migrator) DropTable(values ...interface{}) error {
	return errors.New("not implemented")
}

func (m Migrator) HasColumn(value interface{}, name string) bool {
	var count int
	m.Migrator.RunWithValue(value, func(stmt *gorm.Statement) error {
		return executeOnImmuClient(m.DB, func(ic client.ImmuClient) error {
			resp, er := ic.DescribeTable(context.Background(), stmt.Table)
			if er != nil {
				return er
			}
			for _, v := range resp.Columns {
				if v.Name == name {
					count = 1
				}
			}
			return nil
		})
	})
	return count > 0
}

func (m Migrator) AlterColumn(value interface{}, name string) error {
	return errors.New("not implemented")
}

func (m Migrator) DropColumn(value interface{}, name string) error {
	return errors.New("not implemented")
}

func (m Migrator) CreateConstraint(interface{}, string) error {
	return errors.New("not implemented")
}

func (m Migrator) DropConstraint(interface{}, string) error {
	return errors.New("not implemented")
}

func (m Migrator) HasConstraint(value interface{}, name string) bool {
	return false
}

func (m Migrator) CurrentDatabase() (name string) {
	return "defaultdb"
}

func (m Migrator) CreateIndex(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if idx := stmt.Schema.LookIndex(name); idx != nil {
			values := []interface{}{clause.Table{Name: stmt.Table}, clause.Column{Name: idx.Fields[0].DBName}}
			createIndexSQL := "CREATE "
			if idx.Class != "" {
				createIndexSQL += idx.Class + " "
			}

			createIndexSQL += "INDEX ON ? (?)"

			return m.DB.Exec(createIndexSQL, values...).Error
		}

		return fmt.Errorf("failed to create index with name %v", name)
	})
}

func (m Migrator) HasIndex(value interface{}, name string) bool {
	var count int
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return nil
	})
	return count > 0
}

func (m Migrator) RenameIndex(value interface{}, oldName, newName string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return ErrNotImplemented
	})
}

func (m Migrator) DropIndex(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return ErrNotImplemented
	})
}

func (m Migrator) ColumnTypes(value interface{}) ([]gorm.ColumnType, error) {
	columnTypes := make([]gorm.ColumnType, 0)
	execErr := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return executeOnImmuClient(m.DB, func(ic client.ImmuClient) error {
			resp, err := ic.DescribeTable(context.Background(), stmt.Table)
			if err != nil {
				return err
			}
			for _, r := range resp.Rows {
				for _, c := range r.GetValues() {
					// @todo add missing properties in column description. Not sure where they are needed
					column := Column{
						name: c.GetS(),
					}
					columnTypes = append(columnTypes, column)
					break
				}
			}
			return nil
		})
	})
	return columnTypes, execErr
}

type Column struct {
	name              string
	nullable          sql.NullString
	datatype          string
	maxlen            sql.NullInt64
	precision         sql.NullInt64
	radix             sql.NullInt64
	scale             sql.NullInt64
	datetimeprecision sql.NullInt64
}

func (c Column) Name() string {
	return c.name
}

func (c Column) DatabaseTypeName() string {
	return c.datatype
}

func (c Column) Length() (length int64, ok bool) {
	ok = c.maxlen.Valid
	if ok {
		length = c.maxlen.Int64
	} else {
		length = 0
	}
	return
}

func (c Column) Nullable() (nullable bool, ok bool) {
	if c.nullable.Valid {
		nullable, ok = c.nullable.String == "YES", true
	} else {
		nullable, ok = false, false
	}
	return
}

func (c Column) DecimalSize() (precision int64, scale int64, ok bool) {
	if ok = c.precision.Valid && c.scale.Valid && c.radix.Valid && c.radix.Int64 == 10; ok {
		precision, scale = c.precision.Int64, c.scale.Int64
	} else if ok = c.datetimeprecision.Valid; ok {
		precision, scale = c.datetimeprecision.Int64, 0
	} else {
		precision, scale, ok = 0, 0, false
	}
	return
}

func (m *Migrator) RunWithoutForeignKey(fc func() error) error {
	return ErrNotImplemented
}
