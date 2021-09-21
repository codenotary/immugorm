package immudb

import "gorm.io/gorm/clause"

var currentTable = clause.Table{Name: clause.CurrentTable}

type Upsert struct {
	Table    clause.Table
	Modifier string
}

// Name insert clause name
func (upsert Upsert) Name() string {
	return "UPSERT"
}

// Build build upsert clause
func (upsert Upsert) Build(builder clause.Builder) {
	if upsert.Modifier != "" {
		builder.WriteString(upsert.Modifier)
		builder.WriteByte(' ')
	}

	builder.WriteString("INTO ")
	if upsert.Table.Name == "" {
		builder.WriteQuoted(currentTable)
	} else {
		builder.WriteQuoted(upsert.Table)
	}
}

// MergeClause merge insert clause
func (upsert Upsert) MergeClause(clause *clause.Clause) {
	if v, ok := clause.Expression.(Upsert); ok {
		if upsert.Modifier == "" {
			upsert.Modifier = v.Modifier
		}
		if upsert.Table.Name == "" {
			upsert.Table = v.Table
		}
	}
	clause.Expression = upsert
}
