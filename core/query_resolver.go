package core

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"query-service/db"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

// ExtractRequestedFields inspects p.Info to extract the list of requested field names.
func ExtractRequestedFields(info graphql.ResolveInfo) []string {
	var fields []string
	for _, f := range info.FieldASTs {
		if f.SelectionSet != nil {
			for _, sel := range f.SelectionSet.Selections {
				if field, ok := sel.(*ast.Field); ok {
					fields = append(fields, field.Name.Value)
				}
			}
		}
	}
	return fields
}

// deriveTableName returns the table name based on the type name.
// If the lowercased typeName already ends in "s", we assume it's plural and use it as is.
// Otherwise, we append an "s".
func deriveTableName(typeName string) string {
	return strings.ToLower(typeName)

}

// ResolveSingle builds a dynamic SQL query for a single record.
func ResolveSingle(typeName string, p graphql.ResolveParams) (interface{}, error) {
	requested := ExtractRequestedFields(p.Info)
	if len(requested) == 0 {
		requested = []string{"id"}
	}
	tableName := deriveTableName(typeName)
	query := fmt.Sprintf(`SELECT %s FROM "%s" WHERE id = '%s'`,
		strings.Join(requested, ","), tableName, p.Args["id"].(string))
	log.Println("SQL Query:", query)
	row := db.DB.QueryRow(query)
	values := make([]interface{}, len(requested))
	for i := range requested {
		var v sql.NullString
		values[i] = &v
	}
	err := row.Scan(values...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	result := map[string]interface{}{}
	for i, field := range requested {
		val := values[i].(*sql.NullString)
		if val.Valid {
			result[field] = val.String
		} else {
			result[field] = nil
		}
	}
	return result, nil
}

// ResolveMultiple builds a dynamic SQL query for multiple records.
func ResolveMultiple(typeName string, p graphql.ResolveParams) (interface{}, error) {
	requested := ExtractRequestedFields(p.Info)
	if len(requested) == 0 {
		requested = []string{"id"}
	}
	tableName := deriveTableName(typeName)
	query := fmt.Sprintf(`SELECT %s FROM "%s"`,
		strings.Join(requested, ","), tableName)
	log.Println("SQL Query:", query)
	rows, err := db.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(requested))
		for i := range requested {
			var v sql.NullString
			values[i] = &v
		}
		err := rows.Scan(values...)
		if err != nil {
			return nil, err
		}
		record := map[string]interface{}{}
		for i, field := range requested {
			val := values[i].(*sql.NullString)
			if val.Valid {
				record[field] = val.String
			} else {
				record[field] = nil
			}
		}
		results = append(results, record)
	}
	return results, nil
}
