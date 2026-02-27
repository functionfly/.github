package devtools

import (
	"strings"
)

// parseIndexColumns extracts column names from index definition
func (dt *DatabaseDevTools) parseIndexColumns(definition string) []string {
	// Enhanced parsing to handle complex index definitions including expressions
	start := strings.Index(definition, "(")
	end := strings.LastIndex(definition, ")")

	if start == -1 || end == -1 || end <= start {
		dt.logger.WithField("definition", definition).Debug("No parentheses found in index definition")
		return []string{}
	}

	columnsStr := definition[start+1 : end]

	// Handle nested parentheses and complex expressions
	var columns []string
	var current strings.Builder
	parenDepth := 0

	for i, char := range columnsStr {
		switch char {
		case '(':
			parenDepth++
			current.WriteRune(char)
		case ')':
			parenDepth--
			current.WriteRune(char)
		case ',':
			if parenDepth == 0 {
				// End of column expression
				col := strings.TrimSpace(current.String())
				if col != "" {
					columns = append(columns, col)
				}
				current.Reset()
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}

		// Handle last column
		if i == len(columnsStr)-1 {
			col := strings.TrimSpace(current.String())
			if col != "" {
				columns = append(columns, col)
			}
		}
	}

	dt.logger.WithField("definition", definition).WithField("parsed_columns", columns).Debug("Parsed index columns")
	return columns
}

// parseIndexType determines index type from definition
func (dt *DatabaseDevTools) parseIndexType(definition string) string {
	defLower := strings.ToLower(definition)

	// Check for USING clause with specific index types
	indexTypes := []string{"btree", "hash", "gist", "gin", "spgist", "brin", "rtree", "bloom"}

	for _, idxType := range indexTypes {
		if strings.Contains(defLower, "using "+idxType) {
			return idxType
		}
	}

	// Check for special index patterns
	if strings.Contains(defLower, "unique") && strings.Contains(defLower, "using") {
		// Extract the type from unique index
		parts := strings.Split(defLower, "using ")
		if len(parts) > 1 {
			typePart := strings.Fields(parts[1])[0]
			return strings.TrimSpace(typePart)
		}
	}

	// Default to btree for most PostgreSQL indexes
	return "btree"
}

// parseConstraintDefinition extracts detailed information from constraint definitions
func (dt *DatabaseDevTools) parseConstraintDefinition(constraintType string, definition string) map[string]interface{} {
	result := map[string]interface{}{
		"type":       constraintType,
		"columns":    []string{},
		"references": map[string]interface{}{},
	}

	switch constraintType {
	case "p": // Primary key
		if columns := dt.extractColumnsFromConstraint(definition); len(columns) > 0 {
			result["columns"] = columns
		}
	case "f": // Foreign key
		if refs := dt.extractForeignKeyReferences(definition); len(refs) > 0 {
			result["references"] = refs
		}
		if columns := dt.extractColumnsFromConstraint(definition); len(columns) > 0 {
			result["columns"] = columns
		}
	case "u": // Unique
		if columns := dt.extractColumnsFromConstraint(definition); len(columns) > 0 {
			result["columns"] = columns
		}
	case "c": // Check
		result["expression"] = strings.TrimSpace(definition)
	case "x": // Exclusion
		result["expression"] = strings.TrimSpace(definition)
	}

	return result
}

// extractColumnsFromConstraint extracts column names from constraint definitions
func (dt *DatabaseDevTools) extractColumnsFromConstraint(definition string) []string {
	// Look for patterns like "column_name" or (column1, column2)
	var columns []string

	// Simple regex-like extraction for quoted identifiers
	words := strings.FieldsFunc(definition, func(r rune) bool {
		return r == '(' || r == ')' || r == ',' || r == ' '
	})

	for _, word := range words {
		word = strings.Trim(word, "\"'`")
		if word != "" && !isSQLKeyword(word) {
			columns = append(columns, word)
		}
	}

	return columns
}

// extractForeignKeyReferences extracts foreign key relationship information
func (dt *DatabaseDevTools) extractForeignKeyReferences(definition string) map[string]interface{} {
	result := map[string]interface{}{
		"table":   "",
		"columns": []string{},
	}

	// Look for REFERENCES table_name (columns)
	refsIndex := strings.Index(strings.ToLower(definition), "references")
	if refsIndex == -1 {
		return result
	}

	refsPart := definition[refsIndex+10:] // Skip "REFERENCES"

	// Extract table name
	parts := strings.Fields(strings.TrimSpace(refsPart))
	if len(parts) > 0 {
		result["table"] = strings.Trim(parts[0], "\"`")
	}

	// Extract referenced columns
	if parenStart := strings.Index(refsPart, "("); parenStart != -1 {
		if parenEnd := strings.Index(refsPart[parenStart:], ")"); parenEnd != -1 {
			colsStr := refsPart[parenStart+1 : parenStart+parenEnd]
			cols := strings.Split(colsStr, ",")
			for _, col := range cols {
				col = strings.TrimSpace(strings.Trim(col, "\"`"))
				if col != "" {
					result["columns"] = append(result["columns"].([]string), col)
				}
			}
		}
	}

	return result
}

// isSQLKeyword checks if a word is a common SQL keyword
func isSQLKeyword(word string) bool {
	keywords := []string{
		"primary", "key", "foreign", "references", "unique", "check", "constraint",
		"not", "null", "default", "using", "on", "with", "and", "or",
	}

	wordLower := strings.ToLower(word)
	for _, kw := range keywords {
		if wordLower == kw {
			return true
		}
	}
	return false
}