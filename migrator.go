package immudb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/codenotary/immudb/pkg/client"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
)

type Migrator struct {
	migrator.Migrator
}

func (m Migrator) GetImmuclient() (client.ImmuClient, error) {
	return m.Dialector.(Dialector).GetImmuclient(m.DB)
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
				values = append(values, clause.Column{Name: stmt.Schema.PrimaryFields[0].Name})
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
		immucli, err := m.GetImmuclient()
		if err != nil {
			return err
		}
		_, err = immucli.DescribeTable(context.Background(), stmt.Table)
		if err != nil {
			return err
		}
		count = 1
		return nil
	})

	return count > 0
}

func (m Migrator) DropTable(values ...interface{}) error {
	return errors.New("not implemented")
}

func (m Migrator) HasColumn(value interface{}, name string) bool {
	var count int
	m.Migrator.RunWithValue(value, func(stmt *gorm.Statement) error {
		return nil
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
	return ErrConstraintsNotImplemented
}

func (m Migrator) DropConstraint(interface{}, string) error {
	return ErrConstraintsNotImplemented
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
		immucli, err := m.GetImmuclient()
		if err != nil {
			return err
		}
		resp, err := immucli.DescribeTable(context.Background(), stmt.Table)
		if err != nil {
			return err
		}
		for _, r := range resp.Rows {
			for _, c := range r.GetValues() {
				// @todo add missing properties in colunb description. Not sure where they are needed
				column := Column{
					name: c.GetS(),
				}
				columnTypes = append(columnTypes, column)
				break
			}
		}
		return nil
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
