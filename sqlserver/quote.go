package sqlserver

import "strings"

// QuoteIdentifier quotes an identifier (table name, column name, etc.) for use in SQL Server.
// SQL Server uses square brackets for quoting: [identifier]
// Any closing brackets within the identifier are escaped by doubling them: ] -> ]]
func QuoteIdentifier(name string) string {
	return "[" + strings.ReplaceAll(name, "]", "]]") + "]"
}

// QuoteLiteral quotes a string literal for use in SQL Server.
// Single quotes are escaped by doubling them: ' -> ''
// The N prefix ensures proper Unicode handling.
func QuoteLiteral(s string) string {
	return "N'" + strings.ReplaceAll(s, "'", "''") + "'"
}
