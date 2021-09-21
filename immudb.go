package immudb

import (
	"database/sql/driver"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/stdlib"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

const DriverName = "immudb"

type Dialector struct {
	DriverName string
	opts       *client.Options
	cli        client.ImmuClient
}

func Open(opts *client.Options) gorm.Dialector {
	if opts == nil {
		opts = client.DefaultOptions()
	}
	return &Dialector{
		DriverName: DriverName,
		opts:       opts,
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

	return
}

func (dialector Dialector) ClauseBuilders() map[string]clause.ClauseBuilder {
	return map[string]clause.ClauseBuilder{
		"SELECT": func(c clause.Clause, builder clause.Builder) {
			b := &Stmt{
				Builder: builder,
			}
			c.Build(b)
		},
		"INSERT": func(c clause.Clause, builder clause.Builder) {
			b := &Stmt{
				Builder: builder,
			}
			c.Build(b)
		},
		"UPDATE": func(c clause.Clause, builder clause.Builder) {
			b := &Stmt{
				Builder: builder,
			}

			upsert := Upsert{
				Table: currentTable,
			}

			nc := clause.Clause{
				Name:                "UPSERT",
				BeforeExpression:    c.BeforeExpression,
				AfterNameExpression: c.AfterNameExpression,
				AfterExpression:     c.AfterExpression,
				Expression:          upsert,
				Builder:             c.Builder,
			}
			nc.Build(b)
		},
		"SET": func(c clause.Clause, builder clause.Builder) {
			b := &Stmt{
				Builder: builder,
			}

			st := builder.(*gorm.Statement)

			var eq clause.Eq
			for _, e1 := range st.Clauses["WHERE"].Expression.(clause.Where).Exprs {
				eq = e1.(clause.Eq)
			}
			// when UPSERT where is not supported
			delete(st.Clauses, "WHERE")

			pKeyVal := eq.Value
			pKeyCol := clause.Column{
				Name: eq.Column.(string),
			}
			cv := clause.Values{}
			for _, a := range c.Expression.(clause.Set) {
				cv.Columns = []clause.Column{pKeyCol, a.Column}
				cv.Values = [][]interface{}{{a.Value, pKeyVal}}
			}
			nc := clause.Clause{
				Name:                "",
				BeforeExpression:    c.BeforeExpression,
				AfterNameExpression: c.AfterNameExpression,
				AfterExpression:     c.AfterExpression,
				Expression:          cv,
				Builder:             c.Builder,
			}
			nc.Build(b)
		},
		"VALUES": func(c clause.Clause, builder clause.Builder) {
			b := &Stmt{
				Builder: builder,
			}
			c.Build(b)
		},
		"WHERE": func(c clause.Clause, builder clause.Builder) {
			b := &Stmt{
				Builder: builder,
			}

			var ne []clause.Expression
			for _, e := range c.Expression.(clause.Where).Exprs {
				switch ie := e.(type) {
				case clause.Eq:
					nc := &Eq{
						ie,
					}
					ne = append(ne, nc)
				default:
					ne = append(ne, e)
				}
			}
			nc := clause.Clause{
				Name:                c.Name,
				BeforeExpression:    c.BeforeExpression,
				AfterNameExpression: c.AfterNameExpression,
				AfterExpression:     c.AfterExpression,
				Expression: clause.Where{
					Exprs: ne,
				},
				Builder: c.Builder,
			}
			nc.Build(b)
		},
		"ORDER BY": func(c clause.Clause, builder clause.Builder) {
			b := &Stmt{
				Builder: builder,
			}
			c.Build(b)
		},

		"GROUP BY": func(c clause.Clause, builder clause.Builder) {
			b := &Stmt{
				Builder: builder,
			}
			c.Build(b)
		},
		"LIMIT": func(c clause.Clause, builder clause.Builder) {
			b := &Stmt{
				Builder: builder,
			}
			c.Build(b)
		},
		"FOR": func(c clause.Clause, builder clause.Builder) {
			b := &Stmt{
				Builder: builder,
			}
			c.Build(b)
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
		return "VARCHAR"
	case schema.Bytes:
		return "BLOB"
	case schema.Time:
		return "INTEGER"
	}

	return string(field.DataType)
}

func (dialectopr Dialector) SavePoint(tx *gorm.DB, name string) error {
	return ErrNotImplemented
}

func (dialectopr Dialector) RollbackTo(tx *gorm.DB, name string) error {
	return ErrNotImplemented
}

type columnConverter struct{}

func (cc columnConverter) ColumnConverter(idx int) driver.ValueConverter {
	return nil
}

type valueConverter struct{}

func (cc valueConverter) ConvertValue(v interface{}) (driver.Value, error) {
	return nil, nil
}

/*type ImmuNamer struct {}

func (i *ImmuNamer)TableName(table string) string{
	return strings.ToLower(table)
}
func (i *ImmuNamer)SchemaName(table string) string{
	return ""
}
func (i *ImmuNamer)ColumnName(table, column string) string{
	return strings.ToLower(column)
}
func (i *ImmuNamer)JoinTableName(table string) string{
	return table
}
func (i *ImmuNamer)RelationshipFKName(relationship schema.Relationship) string{
	return ""
}
func (i *ImmuNamer)CheckerName(table, column string) string{
	return ""
}
func (i *ImmuNamer)IndexName(table, column string) string{
	return ""
}*/

func (dialector Dialector) verify(db *gorm.DB) {
	//st := db.Statement.SQL.String()
	//rows, err := db.Rows()
	//db.AddError(errors.Errorf("tampered %s", st))
}

func upsert(db *gorm.DB) {
	println("CIPPA")
}
