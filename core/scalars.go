package core

import (
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

var DateType = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Date",
	Description: "The custom Date scalar type",
	Serialize: func(value interface{}) interface{} {
		if t, ok := value.(time.Time); ok {
			return t.Format(time.RFC3339)
		}
		return nil
	},
	ParseValue: func(value interface{}) interface{} {
		if s, ok := value.(string); ok {
			t, err := time.Parse(time.RFC3339, s)
			if err == nil {
				return t
			}
		}
		return nil
	},
	ParseLiteral: func(valueAST ast.Value) interface{} {
		if v, ok := valueAST.(*ast.StringValue); ok {
			t, err := time.Parse(time.RFC3339, v.Value)
			if err == nil {
				return t
			}
		}
		return nil
	},
})
