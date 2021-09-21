package immudb

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Stmt struct {
	clause.Builder
}

func (istmt *Stmt) WriteQuoted(value interface{}) {
	stmt := istmt.Builder.(*gorm.Statement)
	istmt.QuoteTo(&stmt.SQL, value)
}

func (istmt *Stmt) QuoteTo(writer clause.Writer, field interface{}) {
	stmt := istmt.Builder.(*gorm.Statement)
	switch v := field.(type) {
	case clause.Table:
		if v.Name == clause.CurrentTable {
			if stmt.TableExpr != nil {
				stmt.TableExpr.Build(stmt)
			} else {
				stmt.DB.Dialector.QuoteTo(writer, stmt.Table)
			}
		} else if v.Raw {
			writer.WriteString(v.Name)
		} else {
			stmt.DB.Dialector.QuoteTo(writer, v.Name)
		}

		if v.Alias != "" {
			writer.WriteByte(' ')
			stmt.DB.Dialector.QuoteTo(writer, v.Alias)
		}
	case clause.Column:
		writer.WriteByte(' ')
		if v.Name == clause.PrimaryKey {
			if stmt.Schema == nil {
				stmt.DB.AddError(gorm.ErrModelValueRequired)
			} else if stmt.Schema.PrioritizedPrimaryField != nil {
				stmt.DB.Dialector.QuoteTo(writer, stmt.Schema.PrioritizedPrimaryField.DBName)
			} else if len(stmt.Schema.DBNames) > 0 {
				stmt.DB.Dialector.QuoteTo(writer, stmt.Schema.DBNames[0])
			}
		} else if v.Raw {
			writer.WriteString(v.Name)
		} else {
			stmt.DB.Dialector.QuoteTo(writer, v.Name)
		}

		if v.Alias != "" {
			writer.WriteString(" AS ")
			stmt.DB.Dialector.QuoteTo(writer, v.Alias)
		}
	case []clause.Column:
		writer.WriteByte('(')
		for idx, d := range v {
			if idx > 0 {
				writer.WriteString(",")
			}
			stmt.QuoteTo(writer, d)
		}
		writer.WriteByte(')')
	case clause.Expr:
		v.Build(stmt)
	case string:
		stmt.DB.Dialector.QuoteTo(writer, v)
	case []string:
		writer.WriteByte('(')
		for idx, d := range v {
			if idx > 0 {
				writer.WriteString(",")
			}
			stmt.DB.Dialector.QuoteTo(writer, d)
		}
		writer.WriteByte(')')
	default:
		stmt.DB.Dialector.QuoteTo(writer, fmt.Sprint(field))
	}
}

func (istmt *Stmt) AddVar(writer clause.Writer, vars ...interface{}) {
	stmt := istmt.Builder.(*gorm.Statement)
	stmt.AddVar(writer, vars...)
}
