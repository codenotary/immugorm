package immudb

import (
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

type TimeTravel struct {
	txId uint64
	mode string
}

func (tt TimeTravel) ModifyStatement(stmt *gorm.Statement) {
	clause := stmt.Clauses["FROM"]

	if clause.AfterExpression == nil {
		clause.AfterExpression = tt
	} else if old, ok := clause.AfterExpression.(TimeTravel); ok {
		clause.AfterExpression = old
	} else {
		clause.AfterExpression = Exprs{clause.AfterExpression, tt}
	}
	stmt.Clauses["FROM"] = clause
}

func BeforeTx(tx uint64) TimeTravel {
	return TimeTravel{
		txId: tx, mode: "before"}
}

func AfterTx(tx uint64) TimeTravel {
	return TimeTravel{
		txId: tx, mode: "after",
	}
}

func (tt TimeTravel) Build(builder clause.Builder) {
	if st, ok := builder.(*gorm.Statement); ok {
		old := st.SQL.String()
		new := strings.Replace(old, st.Table, fmt.Sprintf("(%s %s TX %d)", st.Table, strings.ToUpper(tt.mode), tt.txId), -1)
		st.SQL.Reset()
		st.SQL.WriteString(new)
	}
}

type Exprs []clause.Expression

func (exprs Exprs) Build(builder clause.Builder) {
	for idx, expr := range exprs {
		if idx > 0 {
			builder.WriteByte(' ')
		}
		expr.Build(builder)
	}
}
