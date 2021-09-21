package immudb

import (
	"gorm.io/gorm/clause"
	"reflect"
)

type Eq struct {
	clause.Eq
}

func (ieq Eq) Build(builder clause.Builder) {
	eq := ieq.Eq
	builder.WriteQuoted(eq.Column)
	switch eq.Value.(type) {
	case []string, []int, []int32, []int64, []uint, []uint32, []uint64, []interface{}:
		builder.WriteString(" IN (")
		rv := reflect.ValueOf(eq.Value)
		for i := 0; i < rv.Len(); i++ {
			if i > 0 {
				builder.WriteByte(',')
			}
			builder.AddVar(builder, rv.Index(i).Interface())
		}
		builder.WriteByte(')')
	default:
		builder.WriteString(" = ")
		builder.AddVar(builder, eq.Value)
	}
}
