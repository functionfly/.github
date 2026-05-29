package ghost

import "strings"

func mapGoTypeToSQL(goType string) string {
	switch goType {
	case "string":
		return "VARCHAR(255)"
	case "int":
		return "INTEGER"
	case "bool":
		return "BOOLEAN"
	case "timestamp":
		return "TIMESTAMPTZ"
	case "uuid":
		return "UUID"
	case "jsonb":
		return "JSONB"
	case "text":
		return "TEXT"
	default:
		return "VARCHAR(255)"
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteByte(byte(r + 32))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
