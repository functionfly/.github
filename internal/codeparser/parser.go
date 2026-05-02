package codeparser

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func Parse(code string, language string) (*ParseResult, error) {
	if code == "" {
		return nil, ErrEmptyCode
	}

	if len(code) > 102400 {
		return nil, ErrCodeTooLarge
	}

	detectedLang := language
	confidence := 100.0

	if language == "" || language == "auto" {
		detectedLang, confidence = DetectLanguage(code)
	}

	if detectedLang == "unknown" || detectedLang == "" {
		return &ParseResult{
			Language:       "unknown",
			Confidence:     0,
			Functions:      []ParsedFunction{},
			RawCode:        code,
			RawCodeLength:  len(code),
			DetectedAt:     time.Now(),
		}, nil
	}

	var functions []ParsedFunction

	switch detectedLang {
	case "python":
		functions = parsePython(code)
	case "javascript":
		functions = parseJavaScript(code)
	case "typescript":
		functions = parseTypeScript(code)
	case "go":
		functions = parseGo(code)
	case "rust":
		functions = parseRust(code)
	case "ruby":
		functions = parseRuby(code)
	default:
		functions = parseGeneric(code, detectedLang)
	}

	if len(functions) == 0 {
		singleFunc := createGenericFunction(code, detectedLang)
		functions = []ParsedFunction{singleFunc}
	}

	return &ParseResult{
		Language:       detectedLang,
		Confidence:     confidence,
		Functions:      functions,
		RawCode:        code,
		RawCodeLength:  len(code),
		DetectedAt:     time.Now(),
	}, nil
}

func parsePython(code string) []ParsedFunction {
	var functions []ParsedFunction

	funcRegex := regexp.MustCompile(`(?m)^(async\s+)?def\s+(\w+)\s*\(([^)]*)\)(?:\s*->\s*([^:]+))?\s*:`)
	matches := funcRegex.FindAllStringSubmatchIndex(code, -1)

	for _, match := range matches {
		if len(match) < 10 {
			continue
		}

		isAsync := match[2] != -1
		funcName := code[match[4]:match[5]]
		paramsStr := code[match[6]:match[7]]

		var returnType string
		if match[8] != -1 && match[9] != -1 {
			returnType = strings.TrimSpace(code[match[8]:match[9]])
		}

		startLine := countLines(code[:match[0]])
		startPos := match[0]
		endPos := findPythonFunctionEnd(code, startPos)

		funcCode := code[startPos:endPos]
		params := parsePythonParams(paramsStr)
		signature := buildPythonSignature(funcName, params, returnType, isAsync)
		docstring := extractPythonDocstring(funcCode)

		functions = append(functions, ParsedFunction{
			ID:         generateFunctionID(),
			Name:       funcName,
			Language:   "python",
			Signature:  signature,
			Parameters: params,
			ReturnType: returnType,
			Docstring:  docstring,
			Code:       strings.TrimSpace(funcCode),
			StartLine:  startLine + 1,
			EndLine:    startLine + countLines(funcCode),
		})
	}

	return functions
}

func parsePythonParams(paramsStr string) []Parameter {
	if strings.TrimSpace(paramsStr) == "" {
		return []Parameter{}
	}

	var params []Parameter
	parts := strings.Split(paramsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		param := Parameter{}

		if strings.Contains(part, "=") {
			eqIdx := strings.Index(part, "=")
			param.Name = strings.TrimSpace(part[:eqIdx])
			param.HasDefault = true
			defaultVal := strings.TrimSpace(part[eqIdx+1:])
			param.DefaultValue = defaultVal
		} else {
			parts2 := strings.Split(part, ":")
			if len(parts2) == 2 {
				param.Name = strings.TrimSpace(parts2[0])
				param.Type = strings.TrimSpace(parts2[1])
			} else {
				param.Name = part
			}
		}

		params = append(params, param)
	}

	return params
}

func buildPythonSignature(name string, params []Parameter, returnType string, isAsync bool) string {
	prefix := ""
	if isAsync {
		prefix = "async "
	}

	paramStrs := make([]string, len(params))
	for i, p := range params {
		if p.Type != "" {
			paramStrs[i] = p.Name + ": " + p.Type
		} else {
			paramStrs[i] = p.Name
		}
		if p.HasDefault {
			paramStrs[i] += "=" + p.DefaultValue
		}
	}

	sig := prefix + "def " + name + "(" + strings.Join(paramStrs, ", ") + ")"

	if returnType != "" {
		sig += " -> " + returnType
	}

	return sig
}

func extractPythonDocstring(funcCode string) string {
	funcCode = strings.TrimSpace(funcCode)
	lines := strings.Split(funcCode, "\n")
	if len(lines) < 2 {
		return ""
	}

	firstLine := lines[1]
	firstLine = strings.TrimSpace(firstLine)

	if len(firstLine) >= 6 {
		if strings.HasPrefix(firstLine, "\"\"\"") && strings.HasSuffix(firstLine, "\"\"\"") {
			return strings.Trim(firstLine[3:], "\"")
		}
		if strings.HasPrefix(firstLine, "'''") && strings.HasSuffix(firstLine, "'''") {
			return strings.Trim(firstLine[3:], "'")
		}
	}

	restCode := strings.Join(lines[1:], "\n")

	if strings.Contains(restCode, "\"\"\"") {
		start := strings.Index(restCode, "\"\"\"")
		end := strings.Index(restCode[start+3:], "\"\"\"")
		if end != -1 {
			return strings.TrimSpace(restCode[start+3:start+3+end])
		}
	}

	return ""
}

func parseJavaScript(code string) []ParsedFunction {
	var functions []ParsedFunction

	funcRegex := regexp.MustCompile(`(?m)(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(([^)]*)\)\s*\{`)
	matches := funcRegex.FindAllStringSubmatchIndex(code, -1)

	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		funcName := code[match[2]:match[3]]
		paramsStr := code[match[4]:match[5]]

		startLine := countLines(code[:match[0]])
		startPos := match[0]
		endPos := findBlockEnd(code, startPos)

		funcCode := code[startPos:endPos]
		params := parseJSParams(paramsStr)

		isAsync := isAsyncFunc(code, startPos)
		signature := buildJSSignature(funcName, params, false, isAsync)

		functions = append(functions, ParsedFunction{
			ID:         generateFunctionID(),
			Name:       funcName,
			Language:   "javascript",
			Signature:  signature,
			Parameters: params,
			Code:       strings.TrimSpace(funcCode),
			StartLine:  startLine + 1,
			EndLine:    startLine + countLines(funcCode),
		})
	}

	constRegex := regexp.MustCompile(`(?m)(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\(([^)]*)\)\s*=>\s*\{`)
	constMatches := constRegex.FindAllStringSubmatchIndex(code, -1)

	for _, match := range constMatches {
		if len(match) < 6 {
			continue
		}

		funcName := code[match[2]:match[3]]
		paramsStr := code[match[4]:match[5]]

		startLine := countLines(code[:match[0]])
		startPos := match[0]
		endPos := findBlockEnd(code, startPos)

		funcCode := code[startPos:endPos]
		params := parseJSParams(paramsStr)

		isAsync := isAsyncFunc(code, startPos)
		signature := buildJSSignature(funcName, params, true, isAsync)

		functions = append(functions, ParsedFunction{
			ID:         generateFunctionID(),
			Name:       funcName,
			Language:   "javascript",
			Signature:  signature,
			Parameters: params,
			Code:       strings.TrimSpace(funcCode),
			StartLine:  startLine + 1,
			EndLine:    startLine + countLines(funcCode),
		})
	}

	return functions
}

func parseTypeScript(code string) []ParsedFunction {
	return parseJavaScript(code)
}

func parseGo(code string) []ParsedFunction {
	var functions []ParsedFunction

	standaloneFuncRegex := regexp.MustCompile(`(?m)^func\s+(\w+)\s*\(([^)]*)\)(?:\s+([^\s{]+))?\s*\{`)
	matches := standaloneFuncRegex.FindAllStringSubmatchIndex(code, -1)

	for _, match := range matches {
		if len(match) < 8 {
			continue
		}

		funcName := code[match[2]:match[3]]
		paramsStr := code[match[4]:match[5]]

		var returnType string
		if match[6] != -1 && match[7] != -1 {
			returnType = strings.TrimSpace(code[match[6]:match[7]])
		}

		startLine := countLines(code[:match[0]])
		startPos := match[0]
		endPos := findBlockEnd(code, startPos)

		funcCode := code[startPos:endPos]
		params := parseGoParams(paramsStr)
		signature := buildGoSignature(funcName, params, returnType)

		functions = append(functions, ParsedFunction{
			ID:         generateFunctionID(),
			Name:       funcName,
			Language:   "go",
			Signature:  signature,
			Parameters: params,
			ReturnType: returnType,
			Code:       strings.TrimSpace(funcCode),
			StartLine:  startLine + 1,
			EndLine:    startLine + countLines(funcCode),
		})
	}

	methodRegex := regexp.MustCompile(`(?m)^func\s+(\w+)\s+\([^)]+\)\s+(\w+)\s*\(([^)]*)\)(?:\s+([^\s{]+))?\s*\{`)
	methodMatches := methodRegex.FindAllStringSubmatchIndex(code, -1)

	for _, match := range methodMatches {
		if len(match) < 10 {
			continue
		}

		receiver := code[match[2]:match[3]]
		funcName := code[match[4]:match[5]]
		paramsStr := code[match[6]:match[7]]

		var returnType string
		if match[8] != -1 && match[9] != -1 {
			returnType = strings.TrimSpace(code[match[8]:match[9]])
		}

		startLine := countLines(code[:match[0]])
		startPos := match[0]
		endPos := findBlockEnd(code, startPos)

		funcCode := code[startPos:endPos]
		params := parseGoParams(paramsStr)

		parts := strings.Fields(receiver)
		fullName := funcName
		if len(parts) >= 2 {
			fullName = parts[1] + "." + funcName
		}

		signature := buildGoSignature(fullName, params, returnType)

		functions = append(functions, ParsedFunction{
			ID:         generateFunctionID(),
			Name:       funcName,
			Language:   "go",
			Signature:  signature,
			Parameters: params,
			ReturnType: returnType,
			Code:       strings.TrimSpace(funcCode),
			StartLine:  startLine + 1,
			EndLine:    startLine + countLines(funcCode),
		})
	}

	return functions
}

func parseRust(code string) []ParsedFunction {
	var functions []ParsedFunction

	funcRegex := regexp.MustCompile(`(?m)^(?:pub\s+)?(?:async\s+)?fn\s+(\w+)\s*(?:<[^>]*>)?\(([^)]*)\)(?:\s*->\s*([^={]+))?`)
	matches := funcRegex.FindAllStringSubmatchIndex(code, -1)

	for _, match := range matches {
		if len(match) < 8 {
			continue
		}

		funcName := code[match[2]:match[3]]
		paramsStr := code[match[4]:match[5]]

		var returnType string
		if match[6] != -1 && match[7] != -1 {
			returnType = strings.TrimSpace(code[match[6]:match[7]])
		}

		startLine := countLines(code[:match[0]])
		startPos := match[0]

		bracePos := strings.Index(code[startPos:], "{")
		if bracePos == -1 {
			continue
		}

		endPos := findBlockEnd(code, startPos+bracePos)

		funcCode := code[startPos:endPos]
		params := parseRustParams(paramsStr)
		signature := buildRustSignature(funcName, params, returnType)

		functions = append(functions, ParsedFunction{
			ID:         generateFunctionID(),
			Name:       funcName,
			Language:   "rust",
			Signature:  signature,
			Parameters: params,
			ReturnType: returnType,
			Code:       strings.TrimSpace(funcCode),
			StartLine:  startLine + 1,
			EndLine:    startLine + countLines(funcCode),
		})
	}

	return functions
}

func parseRuby(code string) []ParsedFunction {
	var functions []ParsedFunction

	funcRegex := regexp.MustCompile(`(?m)^(private|protected|public)?\s*def\s+(\w+)(?:\s*\(([^)]*)\))?`)
	matches := funcRegex.FindAllStringSubmatchIndex(code, -1)

	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		funcName := code[match[4]:match[5]]

		var paramsStr string
		if match[6] != -1 && match[7] != -1 {
			paramsStr = code[match[6]:match[7]]
		}

		startLine := countLines(code[:match[0]])
		startPos := match[0]

		endPos := findRubyMethodEnd(code, startPos)

		funcCode := code[startPos:endPos]
		params := parseRubyParams(paramsStr)
		signature := buildRubySignature(funcName, params)

		functions = append(functions, ParsedFunction{
			ID:         generateFunctionID(),
			Name:       funcName,
			Language:   "ruby",
			Signature:  signature,
			Parameters: params,
			Code:       strings.TrimSpace(funcCode),
			StartLine:  startLine + 1,
			EndLine:    startLine + countLines(funcCode),
		})
	}

	return functions
}

func parseGeneric(code string, language string) []ParsedFunction {
	return []ParsedFunction{createGenericFunction(code, language)}
}

func createGenericFunction(code string, language string) ParsedFunction {
	lines := strings.Split(code, "\n")
	firstLine := strings.TrimSpace(lines[0])

	name := "main"
	words := strings.Fields(firstLine)
	for i := len(words) - 1; i >= 0; i-- {
		word := words[i]
		word = strings.TrimPrefix(word, "*")
		word = strings.TrimSuffix(word, "(")
		if len(word) > 0 && unicode.IsLetter(rune(word[0])) {
			name = word
			break
		}
	}

	return ParsedFunction{
		ID:        generateFunctionID(),
		Name:      name,
		Language:  language,
		Signature: language + " " + name + "(...)",
		Code:      strings.TrimSpace(code),
		StartLine: 1,
		EndLine:   len(lines),
	}
}

func parseJSParams(paramsStr string) []Parameter {
	return parseGenericParams(paramsStr)
}

func parseGoParams(paramsStr string) []Parameter {
	if strings.TrimSpace(paramsStr) == "" {
		return []Parameter{}
	}

	var params []Parameter
	parts := strings.Split(paramsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		param := Parameter{}

		if strings.Contains(part, " ") {
			spaceIdx := strings.LastIndex(part, " ")
			param.Name = strings.TrimSpace(part[spaceIdx:])
			param.Type = strings.TrimSpace(part[:spaceIdx])
		} else {
			param.Name = part
		}

		params = append(params, param)
	}

	return params
}

func parseRustParams(paramsStr string) []Parameter {
	return parseGenericParams(paramsStr)
}

func parseRubyParams(paramsStr string) []Parameter {
	return parseGenericParams(paramsStr)
}

func parseGenericParams(paramsStr string) []Parameter {
	if strings.TrimSpace(paramsStr) == "" {
		return []Parameter{}
	}

	var params []Parameter
	parts := strings.Split(paramsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		param := Parameter{Name: part}

		if strings.Contains(part, ":") {
			colonIdx := strings.Index(part, ":")
			param.Name = strings.TrimSpace(part[:colonIdx])
			param.Type = strings.TrimSpace(part[colonIdx+1:])
		}

		if strings.Contains(part, "=") {
			param.HasDefault = true
		}

		params = append(params, param)
	}

	return params
}

func buildJSSignature(name string, params []Parameter, isArrow bool, isAsync bool) string {
	prefix := ""
	if isAsync {
		prefix = "async "
	}

	paramStrs := make([]string, len(params))
	for i, p := range params {
		if p.Type != "" {
			paramStrs[i] = p.Name + ": " + p.Type
		} else {
			paramStrs[i] = p.Name
		}
	}

	if isArrow {
		return prefix + "const " + name + " = (" + strings.Join(paramStrs, ", ") + ") =>"
	}

	return prefix + "function " + name + "(" + strings.Join(paramStrs, ", ") + ")"
}

func buildGoSignature(name string, params []Parameter, returnType string) string {
	paramStrs := make([]string, len(params))
	for i, p := range params {
		paramStrs[i] = p.Name + " " + p.Type
	}

	sig := "func " + name + "(" + strings.Join(paramStrs, ", ") + ")"

	if returnType != "" {
		sig += " " + returnType
	}

	return sig
}

func buildRustSignature(name string, params []Parameter, returnType string) string {
	paramStrs := make([]string, len(params))
	for i, p := range params {
		if p.Type != "" {
			paramStrs[i] = p.Name + ": " + p.Type
		} else {
			paramStrs[i] = p.Name
		}
	}

	sig := "fn " + name + "(" + strings.Join(paramStrs, ", ") + ")"

	if returnType != "" {
		sig += " -> " + returnType
	}

	return sig
}

func buildRubySignature(name string, params []Parameter) string {
	paramStrs := make([]string, len(params))
	for i, p := range params {
		paramStrs[i] = p.Name
	}

	return "def " + name + "(" + strings.Join(paramStrs, ", ") + ")"
}

func findPythonFunctionEnd(code string, startPos int) int {
	lines := strings.Split(code[startPos:], "\n")

	if len(lines) < 2 {
		return len(code)
	}

	firstIndent := -1
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		firstIndent = countIndent(line)
		break
	}

	if firstIndent == -1 {
		return len(code)
	}

	depth := 0
	for i, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		currentIndent := countIndent(line)

		if currentIndent <= firstIndent && trimmed != "" && !strings.HasSuffix(trimmed, ":") {
			if depth == 0 {
				return startPos + len(strings.Join(lines[:i+1], "\n"))
			}
		}

		if strings.HasSuffix(trimmed, ":") && !strings.Contains(trimmed, "\"\"\"") && !strings.Contains(trimmed, "'''") {
			depth++
		}

		if trimmed == "end" {
			depth--
			if depth == 0 {
				return startPos + len(strings.Join(lines[:i+2], "\n"))
			}
		}

		if depth > 0 && currentIndent <= firstIndent {
			depth--
		}
	}

	return len(code)
}

func findRubyMethodEnd(code string, startPos int) int {
	lines := strings.Split(code[startPos:], "\n")

	if len(lines) < 2 {
		return len(code)
	}

	firstIndent := -1
	for _, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		firstIndent = countIndent(line)
		break
	}

	if firstIndent == -1 {
		return len(code)
	}

	for i, line := range lines[1:] {
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		currentIndent := countIndent(line)

		if currentIndent <= firstIndent && trimmed != "" {
			return startPos + len(strings.Join(lines[:i+1], "\n"))
		}

		if trimmed == "end" {
			return startPos + len(strings.Join(lines[:i+2], "\n"))
		}
	}

	return len(code)
}

func findBlockEnd(code string, startPos int) int {
	braceCount := 0
	for i := startPos; i < len(code); i++ {
		if code[i] == '{' {
			braceCount++
		} else if code[i] == '}' {
			braceCount--
			if braceCount == 0 {
				return i + 1
			}
		}
	}
	return len(code)
}

func countLines(s string) int {
	return strings.Count(s, "\n")
}

func countIndent(s string) int {
	count := 0
	for _, c := range s {
		if c == ' ' {
			count++
		} else if c == '\t' {
			count += 4
		} else {
			break
		}
	}
	return count
}

func generateFunctionID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36) + "-" + randomString(8)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

func isAsyncFunc(code string, pos int) bool {
	before := code[:pos]
	lines := strings.Split(before, "\n")
	if len(lines) == 0 {
		return false
	}
	lastLine := strings.TrimSpace(lines[len(lines)-1])
	return strings.Contains(lastLine, "async")
}