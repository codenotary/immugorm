package immudb

import (
	"context"
	"database/sql"
	"errors"
	immuschema "github.com/codenotary/immudb/pkg/api/schema"
	"gorm.io/gorm"
	"strings"
)

func (dialector Dialector) verify(db *gorm.DB) {
	rows, err := db.Rows()
	if err != nil {
		db.AddError(err)
		return
	}
	dbName := dialector.opts.Database
	tableName := db.Statement.Table
	pkeyName := db.Statement.Schema.PrioritizedPrimaryField.DBName

	r, err := getImmuRowFromSQLRow(dbName, tableName, rows)
	if err != nil {
		db.AddError(err)
		return
	}

	immucli, err := dialector.GetImmuclient(db)
	if err != nil {
		db.AddError(err)
		return
	}

	pkey, err := getPrimaryKeyFromRow(quoteImmuCol(pkeyName, dbName, tableName), r)
	if err != nil {
		db.AddError(err)
		return
	}

	err = immucli.VerifyRow(context.Background(), r, tableName, pkey)
	if err != nil {
		if err.Error() == "data is corrupted" {
			db.AddError(ErrCorruptedData)
		} else {
			db.AddError(err)
		}
	}
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
