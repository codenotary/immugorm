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
	embsql "github.com/codenotary/immudb/embedded/sql"
	immuschema "github.com/codenotary/immudb/pkg/api/schema"
	"github.com/codenotary/immudb/pkg/client"
	"github.com/codenotary/immudb/pkg/stdlib"
	"gorm.io/gorm"
	"strings"
)

func (dialector *Dialector) verify(db *gorm.DB) {
	rows, err := db.Rows()
	if err != nil {
		db.AddError(err)
		return
	}
	dbName := dialector.opts.Database
	tableName := db.Statement.Table
	pkeyName := db.Statement.Schema.PrioritizedPrimaryField.DBName

	if from, ok := db.Statement.Clauses["FROM"]; ok {
		if _, ok := from.AfterExpression.(TimeTravel); ok {
			db.AddError(ErrTimeTravelNotAvailable)
			return
		}
	}

	r, err := getImmuRowFromSQLRow(dbName, tableName, rows)
	if err != nil {
		db.AddError(err)
		return
	}

	pkey, err := getPrimaryKeyFromRow(quoteImmuCol(pkeyName, dbName, tableName), r)
	if err != nil {
		db.AddError(err)
		return
	}

	var ic client.ImmuClient
	sqlDB, err := db.DB()
	if err != nil {
		db.AddError(err)
		return
	}
	conn, err := sqlDB.Conn(context.Background())
	if err != nil {
		db.AddError(err)
		return
	}
	conn.Raw(func(driverConn interface{}) error {
		ic = driverConn.(*stdlib.Conn).GetImmuClient()
		err = ic.VerifyRow(context.Background(), r, tableName, pkey)
		if err != nil {
			if err.Error() == "data is corrupted" {
				db.AddError(ErrCorruptedData)
			} else {
				db.AddError(err)
			}
		}
		return nil
	})
	conn.Close()
	return
}

func getPrimaryKeyFromRow(pkeyName string, r *immuschema.Row) ([]*immuschema.SQLValue, error) {
	for i, v := range r.Values {
		if r.Columns[i] == pkeyName {
			return []*immuschema.SQLValue{v}, nil
		}
	}
	return nil, errors.New("primary key not found")
}
func getImmuRowFromSQLRow(dbName, tableName string, rows *sql.Rows) (*immuschema.Row, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	immucols := make([]string, len(cols))
	for i, c := range cols {
		immucols[i] = quoteImmuCol(c, dbName, tableName)
	}
	ctype, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	vals := make([]interface{}, len(cols))

	for i, v := range ctype {
		switch v.DatabaseTypeName() {
		case "VARCHAR":
			vals[i] = new(sql.NullString)
		case "INTEGER":
			vals[i] = new(sql.NullInt64)
		case "BOOLEAN":
			vals[i] = new(sql.NullBool)
		case "BLOB":
			vals[i] = new([]byte)
		case "ANY":
			vals[i] = new(sql.NullString)
		case "TIMESTAMP":
			vals[i] = new(sql.NullTime)
		default:
			vals[i] = new(sql.NullString)
		}
	}

	for rows.Next() {
		err = rows.Scan(vals...)
	}

	tvals := make([]*immuschema.SQLValue, 0)
	for _, v := range vals {
		var s *immuschema.SQLValue
		switch t := v.(type) {
		case *sql.NullString:
			if t.Valid {
				s = &immuschema.SQLValue{Value: &immuschema.SQLValue_S{S: t.String}}
			}
		case *sql.NullInt64:
			if t.Valid {
				s = &immuschema.SQLValue{Value: &immuschema.SQLValue_N{N: t.Int64}}
			}
		case *sql.NullBool:
			if t.Valid {
				s = &immuschema.SQLValue{Value: &immuschema.SQLValue_B{B: t.Bool}}
			}
		case *[]byte:
			if len(*t) > 0 {
				s = &immuschema.SQLValue{Value: &immuschema.SQLValue_Bs{Bs: *t}}
			}
		case *sql.NullTime:
			if t.Valid {
				s = &immuschema.SQLValue{Value: &immuschema.SQLValue_Ts{Ts: embsql.TimeToInt64(t.Time)}}
			}
		default:
			s = nil
		}
		if s == nil {
			s = &immuschema.SQLValue{Value: &immuschema.SQLValue_Null{}}
		}
		tvals = append(tvals, s)
	}
	r := &immuschema.Row{
		Columns: immucols,
		Values:  tvals,
	}

	return r, nil
}

func quoteImmuCol(col, dbname, tablename string) string {
	return strings.Join([]string{"(" + dbname, tablename, col + ")"}, ".")
}
