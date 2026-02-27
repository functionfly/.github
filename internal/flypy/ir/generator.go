package ir

import (
	"encoding/json"
	"fmt"

	"github.com/functionfly/functionfly/internal/flypy/parser"
)

// IRType represents the type of an IR value
type IRType struct {
	Base    string  // "int", "float", "string", "bool", "list", "dict", "none", "unknown", "bytes", "object"
	Element *IRType // For lists: element type, for dicts: value type
	Key     *IRType // For dicts: key type (usually string)
	IsInput bool    // True if this comes from input parameters (should be serde_json::Value)
}

func (t IRType) String() string {
	if t.Element == nil {
		return t.Base
	}
	if t.Base == "list" {
		return fmt.Sprintf("list[%s]", t.Element.String())
	}
	if t.Base == "dict" {
		keyStr := "string"
		if t.Key != nil {
			keyStr = t.Key.String()
		}
		return fmt.Sprintf("dict[%s,%s]", keyStr, t.Element.String())
	}
	return t.Base
}

// Equals checks if two IRTypes are equal
func (t IRType) Equals(other IRType) bool {
	if t.Base != other.Base || t.IsInput != other.IsInput {
		return false
	}
	if t.Element == nil && other.Element == nil {
		return true
	}
	if t.Element == nil || other.Element == nil {
		return false
	}
	if !t.Element.Equals(*other.Element) {
		return false
	}
	if t.Key == nil && other.Key == nil {
		return true
	}
	if t.Key == nil || other.Key == nil {
		return false
	}
	return t.Key.Equals(*other.Key)
}

// Predefined types for convenience
var (
	IRTypeInt     = IRType{Base: "int"}
	IRTypeFloat   = IRType{Base: "float"}
	IRTypeString  = IRType{Base: "string"}
	IRTypeBool    = IRType{Base: "bool"}
	IRTypeList    = IRType{Base: "list", Element: &IRTypeString}                     // Default list of strings
	IRTypeDict    = IRType{Base: "dict", Key: &IRTypeString, Element: &IRTypeString} // Default dict
	IRTypeNone    = IRType{Base: "none"}
	IRTypeUnknown = IRType{Base: "unknown"}
	IRTypeBytes   = IRType{Base: "bytes"}
	IRTypeObject  = IRType{Base: "object"}
)

// Operation types for complex modules
const (
	// Standard operations
	OpAssign    = "assign"
	OpReturn    = "return"
	OpCall      = "call"
	OpBinOp     = "binop"
	OpUnaryOp   = "unaryop"
	OpCompare   = "compare"
	OpBoolOp    = "boolop"
	OpSubscript = "subscript"
	OpDict      = "dict"
	OpList      = "list"
	OpIf        = "if"
	OpFor       = "for"
	OpWhile     = "while"
	OpTry       = "try"
	OpRaise     = "raise"

	// CSV module operations (complex mode)
	OpCSVReader        = "csv_reader"
	OpCSVWriter        = "csv_writer"
	OpCSVReadRow       = "csv_read_row"
	OpCSVWriteRow      = "csv_write_row"
	OpCSVGetFieldnames = "csv_get_fieldnames"

	// IO module operations (complex mode)
	OpStringIO   = "string_io"
	OpBytesIO    = "bytes_io"
	OpIOWrite    = "io_write"
	OpIORead     = "io_read"
	OpIOGetValue = "io_getvalue"
	OpIOSeek     = "io_seek"
	OpIOTell     = "io_tell"

	// Regex module operations (complex mode)
	OpReMatch   = "re_match"
	OpReSearch  = "re_search"
	OpReSub     = "re_sub"
	OpReFindall = "re_findall"
	OpReSplit   = "re_split"
	OpReCompile = "re_compile"

	// Datetime module operations (complex mode)
	OpDatetimeParse  = "datetime_parse"
	OpDatetimeFormat = "datetime_format"
	OpDatetimeNow    = "datetime_now"
	OpTimedelta      = "timedelta"
	OpDateAdd        = "date_add"
	OpDateSub        = "date_sub"

	// JSON operations (all modes)
	OpJSONLoads = "json_loads"
	OpJSONDumps = "json_dumps"

	// Hash operations (complex mode)
	OpHashMD5    = "hash_md5"
	OpHashSHA256 = "hash_sha256"

	// Base64 operations (complex mode)
	OpBase64Encode = "base64_encode"
	OpBase64Decode = "base64_decode"
)

// Module represents an IR module (collection of functions)
type Module struct {
	Name      string
	Functions []*Function
	Imports   []string
	Metadata  map[string]interface{}
	Mode      string // "deterministic", "complex", or "compatible"
}

// Function represents an IR function
type Function struct {
	Name          string
	Parameters    []Parameter
	Body          []Operation
	ReturnType    IRType
	Pure          bool
	Deterministic bool
}

// Parameter represents a function parameter
type Parameter struct {
	Name         string
	Type         IRType
	DefaultValue interface{}
}

// Operation represents an IR operation
type Operation struct {
	Type     string
	Result   string
	Operands []Value
	Type_    IRType
	Kind     ValueKind
	Value    interface{}
	Module   string // Module name for module-specific operations (e.g., "csv", "io", "re")
	Method   string // Method name for method calls

	// Block-structured control flow support
	Body      []Operation // For if/for/while body
	ElseBody  []Operation // For if/else
	Target    string      // For loop variable (for x in ...)
	Iterator  Value       // For loop iterator
	Condition Value       // For while/if condition
	HasElse   bool        // Track if if-statement has else branch

	// Try/except support
	Handlers    []ExceptionHandler // For try/except
	FinallyBody []Operation        // For finally block
}

// ExceptionHandler represents an except handler
type ExceptionHandler struct {
	ExceptionType string      // Exception type name (e.g., "ValueError", "Exception")
	VarName       string      // Variable name for "as e" clause
	Body          []Operation // Handler body
}

// Value represents an IR value (literal or reference)
type Value struct {
	Type  IRType
	Kind  ValueKind
	Value interface{}
	// CanUnwrap indicates this value can be unwrapped by the backend
	// For example, io.StringIO(x) can be unwrapped to just x when passed to csv.DictReader
	CanUnwrap bool
	// UnwrapTo specifies what type to unwrap to: "string", "bytes", etc.
	UnwrapTo string
}

// ValueKind represents the kind of value
type ValueKind int

const (
	Literal ValueKind = iota
	Reference
	ParameterRef
	Call
	BinOp
	UnaryOp
	Compare
	BoolOp
	Subscript
	Dict
	List
	Index
	Slice
	ModuleCall // For module function calls like csv.reader()
	MethodCall // For method calls like output.getvalue()
	ListComp   // For list comprehensions
	DictComp   // For dict comprehensions
	FString    // For f-string formatted strings
)

// Generate generates an IR module from Python AST
func Generate(pythonAST *parser.PythonAST, name string) (*Module, error) {
	module := &Module{
		Name:      name,
		Functions: make([]*Function, 0),
		Imports:   make([]string, 0),
		Metadata:  make(map[string]interface{}),
	}

	// Extract functions from AST
	functions := parser.GetFunctions(pythonAST)

	for _, fn := range functions {
		irFunc, err := convertFunction(fn)
		if err != nil {
			return nil, fmt.Errorf("failed to convert function %s: %w", parser.GetFunctionName(fn), err)
		}
		module.Functions = append(module.Functions, irFunc)
	}

	return module, nil
}

func convertFunction(fn map[string]interface{}) (*Function, error) {
	funcName := parser.GetFunctionName(fn)
	args := parser.GetFunctionArgs(fn)
	argNames := parser.GetArgNames(args)
	body := parser.GetFunctionBody(fn)

	irFunc := &Function{
		Name:          funcName,
		Parameters:    make([]Parameter, 0),
		Body:          make([]Operation, 0),
		ReturnType:    IRTypeUnknown,
		Pure:          true,
		Deterministic: true,
	}

	// Add parameters
	for _, argName := range argNames {
		irFunc.Parameters = append(irFunc.Parameters, Parameter{
			Name: argName,
			Type: IRType{Base: "unknown", IsInput: true}, // Input parameters are serde_json::Value
		})
	}

	// Convert body statements
	for _, stmt := range body {
		if stmtMap, ok := stmt.(map[string]interface{}); ok {
			ops, err := convertStatement(stmtMap)
			if err != nil {
				return nil, err
			}
			irFunc.Body = append(irFunc.Body, ops...)
		}
	}

	return irFunc, nil
}

func convertStatement(stmt map[string]interface{}) ([]Operation, error) {
	ops := make([]Operation, 0)

	if parser.IsReturn(stmt) {
		// Handle return statement
		retVal := parser.GetReturnValue(stmt)
		if retVal != nil {
			if retMap, ok := retVal.(map[string]interface{}); ok {
				val, _, err := convertExpression(retMap)
				if err != nil {
					return nil, err
				}
				ops = append(ops, Operation{
					Type:     "return",
					Operands: []Value{val},
				})
			}
		} else {
			// Return without value (returns None)
			ops = append(ops, Operation{
				Type:     "return",
				Operands: []Value{{Type: IRTypeNone, Kind: Literal, Value: nil}},
			})
		}
	} else if parser.IsAssign(stmt) {
		// Handle assignment
		target := parser.GetAssignTarget(stmt)
		value := parser.GetAssignValue(stmt)

		if targetMap, ok := target.(map[string]interface{}); ok {
			if parser.IsName(targetMap) {
				varName := parser.GetNameID(targetMap)

				if valueMap, ok := value.(map[string]interface{}); ok {
					val, valType, err := convertExpression(valueMap)
					if err != nil {
						return nil, err
					}
					ops = append(ops, Operation{
						Type:     "assign",
						Result:   varName,
						Operands: []Value{val},
						Type_:    valType,
					})
				}
			} else if parser.IsSubscript(targetMap) {
				// Handle subscript assignment: arr[i] = value, dict[key] = value
				valueExpr := parser.GetSubscriptValue(targetMap)
				sliceExpr := parser.GetSubscriptSlice(targetMap)

				targetVal, _, err := convertExpression(valueExpr.(map[string]interface{}))
				if err != nil {
					return nil, err
				}

				var indexVal Value
				if sliceMap, ok := sliceExpr.(map[string]interface{}); ok {
					indexVal, _, err = convertExpression(sliceMap)
					if err != nil {
						return nil, err
					}
				}

				assignVal, _, err := convertExpression(value.(map[string]interface{}))
				if err != nil {
					return nil, err
				}

				ops = append(ops, Operation{
					Type:     "assign_subscript",
					Operands: []Value{targetVal, indexVal, assignVal},
				})
			}
		}
	} else if parser.IsExpr(stmt) {
		// Handle expression statement - might be a call
		if value, ok := stmt["value"].(map[string]interface{}); ok {
			val, _, err := convertExpression(value)
			if err != nil {
				return nil, err
			}
			// Store expression as a statement (for side effects like method calls)
			ops = append(ops, Operation{
				Type:  "expr",
				Value: val,
			})
		}
	} else if parser.IsIf(stmt) {
		// Handle if statement with proper block structure
		test := parser.GetIfTest(stmt)
		ifBody := parser.GetIfBody(stmt)
		elseBody := parser.GetIfOrelse(stmt)

		var testVal Value
		var err error
		if testMap, ok := test.(map[string]interface{}); ok {
			testVal, _, err = convertExpression(testMap)
			if err != nil {
				return nil, err
			}
		}

		ifOp := Operation{
			Type:      "if",
			Condition: testVal,
			Body:      make([]Operation, 0),
			ElseBody:  make([]Operation, 0),
			HasElse:   len(elseBody) > 0,
		}

		// Convert if body
		for _, ifStmt := range ifBody {
			if ifStmtMap, ok := ifStmt.(map[string]interface{}); ok {
				ifOps, err := convertStatement(ifStmtMap)
				if err != nil {
					return nil, err
				}
				ifOp.Body = append(ifOp.Body, ifOps...)
			}
		}

		// Convert else body
		for _, elseStmt := range elseBody {
			if elseStmtMap, ok := elseStmt.(map[string]interface{}); ok {
				elseOps, err := convertStatement(elseStmtMap)
				if err != nil {
					return nil, err
				}
				ifOp.ElseBody = append(ifOp.ElseBody, elseOps...)
			}
		}

		ops = append(ops, ifOp)
	} else if parser.IsFor(stmt) {
		// Handle for loop with proper block structure
		target := parser.GetForTarget(stmt)
		iter := parser.GetForIter(stmt)
		body := parser.GetForBody(stmt)

		var targetName string
		var targetNames []string // For tuple unpacking like: for i, row in enumerate(data)
		if targetMap, ok := target.(map[string]interface{}); ok {
			// Check if target is a tuple (for tuple unpacking)
			if nodeType, ok := targetMap["__node_type__"].(string); ok && nodeType == "Tuple" {
				// Handle tuple unpacking: for i, row in enumerate(data)
				if elts, ok := targetMap["elts"].([]interface{}); ok {
					targetNames = make([]string, 0, len(elts))
					for _, elt := range elts {
						if eltMap, ok := elt.(map[string]interface{}); ok {
							targetNames = append(targetNames, parser.GetNameID(eltMap))
						}
					}
				}
			} else {
				targetName = parser.GetNameID(targetMap)
			}
		}

		var iterVal Value
		var err error
		if iterMap, ok := iter.(map[string]interface{}); ok {
			iterVal, _, err = convertExpression(iterMap)
			if err != nil {
				return nil, err
			}
		}

		forOp := Operation{
			Type:     "for",
			Target:   targetName,
			Iterator: iterVal,
			Body:     make([]Operation, 0),
		}

		// Store tuple target names if present
		if len(targetNames) > 0 {
			forOp.Value = targetNames
		}

		// Convert for body
		for _, bodyStmt := range body {
			if bodyStmtMap, ok := bodyStmt.(map[string]interface{}); ok {
				bodyOps, err := convertStatement(bodyStmtMap)
				if err != nil {
					return nil, err
				}
				forOp.Body = append(forOp.Body, bodyOps...)
			}
		}

		ops = append(ops, forOp)
	} else if parser.IsWhile(stmt) {
		// Handle while loop with proper block structure
		test := parser.GetWhileTest(stmt)
		body := parser.GetWhileBody(stmt)

		var testVal Value
		var err error
		if testMap, ok := test.(map[string]interface{}); ok {
			testVal, _, err = convertExpression(testMap)
			if err != nil {
				return nil, err
			}
		}

		whileOp := Operation{
			Type:      "while",
			Condition: testVal,
			Body:      make([]Operation, 0),
		}

		// Convert while body
		for _, bodyStmt := range body {
			if bodyStmtMap, ok := bodyStmt.(map[string]interface{}); ok {
				bodyOps, err := convertStatement(bodyStmtMap)
				if err != nil {
					return nil, err
				}
				whileOp.Body = append(whileOp.Body, bodyOps...)
			}
		}

		ops = append(ops, whileOp)
	} else if parser.IsBreak(stmt) {
		// Handle break statement
		ops = append(ops, Operation{
			Type: "break",
		})
	} else if parser.IsContinue(stmt) {
		// Handle continue statement
		ops = append(ops, Operation{
			Type: "continue",
		})
	} else if parser.IsAugAssign(stmt) {
		// Handle augmented assignment (+=, -=, *=, etc.)
		target := parser.GetAugAssignTarget(stmt)
		op := parser.GetAugAssignOp(stmt)
		value := parser.GetAugAssignValue(stmt)

		if targetMap, ok := target.(map[string]interface{}); ok {
			if parser.IsName(targetMap) {
				varName := parser.GetNameID(targetMap)

				if valueMap, ok := value.(map[string]interface{}); ok {
					val, valType, err := convertExpression(valueMap)
					if err != nil {
						return nil, err
					}
					ops = append(ops, Operation{
						Type:     "aug_assign",
						Result:   varName,
						Operands: []Value{val},
						Type_:    valType,
						Module:   op, // Reuse Module field for the operator
					})
				}
			}
		}
	} else if parser.IsTry(stmt) {
		// Handle try/except/finally
		tryBody := parser.GetTryBody(stmt)
		handlers := parser.GetTryHandlers(stmt)
		elseBody := parser.GetTryOrelse(stmt)
		finallyBody := parser.GetTryFinalbody(stmt)

		tryOp := Operation{
			Type:        "try",
			Body:        make([]Operation, 0),
			Handlers:    make([]ExceptionHandler, 0),
			ElseBody:    make([]Operation, 0),
			FinallyBody: make([]Operation, 0),
		}

		// Convert try body
		for _, tryStmt := range tryBody {
			if tryStmtMap, ok := tryStmt.(map[string]interface{}); ok {
				tryOps, err := convertStatement(tryStmtMap)
				if err != nil {
					return nil, err
				}
				tryOp.Body = append(tryOp.Body, tryOps...)
			}
		}

		// Convert exception handlers
		for _, handler := range handlers {
			if handlerMap, ok := handler.(map[string]interface{}); ok {
				excHandler := ExceptionHandler{
					Body: make([]Operation, 0),
				}

				// Get exception type
				if excType := parser.GetExceptHandlerType(handlerMap); excType != nil {
					if excTypeMap, ok := excType.(map[string]interface{}); ok {
						if parser.IsName(excTypeMap) {
							excHandler.ExceptionType = parser.GetNameID(excTypeMap)
						}
					}
				}

				// Get exception variable name
				excHandler.VarName = parser.GetExceptHandlerName(handlerMap)

				// Convert handler body
				handlerBody := parser.GetExceptHandlerBody(handlerMap)
				for _, hStmt := range handlerBody {
					if hStmtMap, ok := hStmt.(map[string]interface{}); ok {
						hOps, err := convertStatement(hStmtMap)
						if err != nil {
							return nil, err
						}
						excHandler.Body = append(excHandler.Body, hOps...)
					}
				}

				tryOp.Handlers = append(tryOp.Handlers, excHandler)
			}
		}

		// Convert else body (executed if no exception)
		for _, elseStmt := range elseBody {
			if elseStmtMap, ok := elseStmt.(map[string]interface{}); ok {
				elseOps, err := convertStatement(elseStmtMap)
				if err != nil {
					return nil, err
				}
				tryOp.ElseBody = append(tryOp.ElseBody, elseOps...)
			}
		}

		// Convert finally body
		for _, finallyStmt := range finallyBody {
			if finallyStmtMap, ok := finallyStmt.(map[string]interface{}); ok {
				finallyOps, err := convertStatement(finallyStmtMap)
				if err != nil {
					return nil, err
				}
				tryOp.FinallyBody = append(tryOp.FinallyBody, finallyOps...)
			}
		}

		ops = append(ops, tryOp)
	} else if parser.IsRaise(stmt) {
		// Handle raise statement
		exc := parser.GetRaiseExc(stmt)
		var excVal Value
		if exc != nil {
			if excMap, ok := exc.(map[string]interface{}); ok {
				excVal, _, _ = convertExpression(excMap)
			}
		}
		ops = append(ops, Operation{
			Type:     "raise",
			Operands: []Value{excVal},
		})
	}

	return ops, nil
}

func convertExpression(expr map[string]interface{}) (Value, IRType, error) {
	if parser.IsConstant(expr) {
		val := parser.GetConstantValue(expr)
		irType := inferType(val)
		return Value{
			Type:  irType,
			Kind:  Literal,
			Value: val,
		}, irType, nil
	}

	if parser.IsName(expr) {
		name := parser.GetNameID(expr)
		return Value{
			Type:  IRTypeUnknown,
			Kind:  Reference,
			Value: name,
		}, IRTypeUnknown, nil
	}

	if parser.IsBinOp(expr) {
		op, left, right := parser.GetBinOpInfo(expr)
		leftVal, leftType, err := convertExpression(left.(map[string]interface{}))
		if err != nil {
			return Value{}, IRTypeUnknown, err
		}
		rightVal, rightType, err := convertExpression(right.(map[string]interface{}))
		if err != nil {
			return Value{}, IRTypeUnknown, err
		}

		resultType := IRTypeInt
		if leftType.Base == "float" || rightType.Base == "float" {
			resultType = IRTypeFloat
		}

		return Value{
			Type: resultType,
			Kind: BinOp,
			Value: map[string]interface{}{
				"op":    op,
				"left":  leftVal,
				"right": rightVal,
			},
		}, resultType, nil
	}

	if parser.IsCall(expr) {
		funcName := parser.GetCallFunc(expr)
		args := parser.GetCallArgs(expr)
		keywords := parser.GetCallKeywords(expr)

		operands := make([]Value, 0)
		for _, arg := range args {
			if argMap, ok := arg.(map[string]interface{}); ok {
				val, _, err := convertExpression(argMap)
				if err != nil {
					return Value{}, IRTypeUnknown, err
				}
				operands = append(operands, val)
			}
		}

		// Convert keyword arguments
		kwargs := make(map[string]Value)
		for _, kw := range keywords {
			if kw.Value != nil {
				if valMap, ok := kw.Value.(map[string]interface{}); ok {
					val, _, err := convertExpression(valMap)
					if err == nil {
						kwargs[kw.Arg] = val
					}
				}
			}
		}

		// Check if this is a method call (obj.method())
		// The funcName from GetCallFunc for Attribute calls returns just the attr name
		// We need to check the func node to see if it's an Attribute
		if funcNode, ok := expr["func"].(map[string]interface{}); ok {
			if parser.IsAttribute(funcNode) {
				// This is a method call like obj.method() or module.function()
				receiver := parser.GetAttributeValue(funcNode)
				methodName := parser.GetAttributeAttr(funcNode)

				// Check if the receiver is a module name (e.g., io, csv, json, re)
				if receiverMap, ok := receiver.(map[string]interface{}); ok && parser.IsName(receiverMap) {
					receiverName := parser.GetNameID(receiverMap)
					if isKnownModule(receiverName) {
						// This is a module call like io.StringIO(), csv.DictWriter()
						returnType := inferModuleCallReturnType(receiverName, methodName)
						value := Value{
							Type: returnType,
							Kind: ModuleCall,
							Value: map[string]interface{}{
								"module": receiverName,
								"func":   methodName,
								"args":   operands,
								"kwargs": kwargs,
							},
						}
						// Set CanUnwrap flag for io.StringIO - can be unwrapped to string
						// when passed to functions like csv.DictReader
						if receiverName == "io" && methodName == "StringIO" {
							value.CanUnwrap = true
							value.UnwrapTo = "string"
						}
						return value, returnType, nil
					}
				}

				receiverVal, _, err := convertExpression(receiver.(map[string]interface{}))
				if err != nil {
					return Value{}, IRTypeUnknown, err
				}

				return Value{
					Type: IRTypeUnknown,
					Kind: MethodCall,
					Value: map[string]interface{}{
						"receiver": receiverVal,
						"method":   methodName,
						"args":     operands,
					},
				}, IRTypeUnknown, nil
			}

			// Check if it's a module call (module.function())
			if parser.IsName(funcNode) {
				fullName := parser.GetNameID(funcNode)
				// Check for known module patterns
				module, funcPart := splitModuleFunction(fullName)
				if module != "" {
					return Value{
						Type: IRTypeUnknown,
						Kind: ModuleCall,
						Value: map[string]interface{}{
							"module": module,
							"func":   funcPart,
							"args":   operands,
						},
					}, IRTypeUnknown, nil
				}
			}
		}

		// Regular function call - try to infer the return type
		returnType := inferTypeFromExpr(expr)
		return Value{
			Type: returnType,
			Kind: Call,
			Value: map[string]interface{}{
				"func": funcName,
				"args": operands,
			},
		}, returnType, nil
	}

	if parser.IsSubscript(expr) {
		value := parser.GetSubscriptValue(expr)
		slice := parser.GetSubscriptSlice(expr)

		valueVal, _, err := convertExpression(value.(map[string]interface{}))
		if err != nil {
			return Value{}, IRTypeUnknown, err
		}

		var sliceVal Value
		if sliceMap, ok := slice.(map[string]interface{}); ok {
			// Check if this is a Slice node (for arr[1:3] syntax)
			if parser.IsSlice(sliceMap) {
				sliceVal, _, err = convertSliceExpression(sliceMap)
				if err != nil {
					return Value{}, IRTypeUnknown, err
				}
			} else {
				// Regular index expression
				sliceVal, _, err = convertExpression(sliceMap)
				if err != nil {
					return Value{}, IRTypeUnknown, err
				}
			}
		}

		return Value{
			Type: IRTypeUnknown,
			Kind: Subscript,
			Value: map[string]interface{}{
				"value": valueVal,
				"index": sliceVal,
			},
		}, IRTypeUnknown, nil
	}

	if parser.IsAttribute(expr) {
		attr := parser.GetAttributeAttr(expr)
		value := parser.GetAttributeValue(expr)

		valueVal, _, err := convertExpression(value.(map[string]interface{}))
		if err != nil {
			return Value{}, IRTypeUnknown, err
		}

		return Value{
			Type: IRTypeUnknown,
			Kind: Reference,
			Value: map[string]interface{}{
				"attr":  attr,
				"value": valueVal,
			},
		}, IRTypeUnknown, nil
	}

	if parser.IsCompare(expr) {
		ops := parser.GetCompareOps(expr)
		left := parser.GetCompareLeft(expr)
		comparators := parser.GetCompareComparators(expr)

		leftVal, _, err := convertExpression(left.(map[string]interface{}))
		if err != nil {
			return Value{}, IRTypeUnknown, err
		}

		compVals := make([]Value, 0)
		for _, comp := range comparators {
			if compMap, ok := comp.(map[string]interface{}); ok {
				compVal, _, err := convertExpression(compMap)
				if err != nil {
					return Value{}, IRTypeUnknown, err
				}
				compVals = append(compVals, compVal)
			}
		}

		return Value{
			Type: IRTypeBool,
			Kind: Compare,
			Value: map[string]interface{}{
				"ops":         ops,
				"left":        leftVal,
				"comparators": compVals,
			},
		}, IRTypeBool, nil
	}

	if parser.IsBoolOp(expr) {
		op := parser.GetBoolOpOp(expr)
		values := parser.GetBoolOpValues(expr)

		valList := make([]Value, 0)
		for _, v := range values {
			if vMap, ok := v.(map[string]interface{}); ok {
				val, _, err := convertExpression(vMap)
				if err != nil {
					return Value{}, IRTypeUnknown, err
				}
				valList = append(valList, val)
			}
		}

		return Value{
			Type: IRTypeBool,
			Kind: BoolOp,
			Value: map[string]interface{}{
				"op":     op,
				"values": valList,
			},
		}, IRTypeBool, nil
	}

	if parser.IsUnaryOp(expr) {
		op := parser.GetUnaryOpOp(expr)
		operand := parser.GetUnaryOpOperand(expr)

		operandVal, operandType, err := convertExpression(operand.(map[string]interface{}))
		if err != nil {
			return Value{}, IRTypeUnknown, err
		}

		return Value{
			Type: operandType,
			Kind: UnaryOp,
			Value: map[string]interface{}{
				"op":      op,
				"operand": operandVal,
			},
		}, operandType, nil
	}

	if parser.IsDict(expr) {
		keys := parser.GetDictKeys(expr)
		values := parser.GetDictValues(expr)

		keyList := make([]Value, 0)
		valList := make([]Value, 0)

		for _, k := range keys {
			if kMap, ok := k.(map[string]interface{}); ok {
				keyVal, _, err := convertExpression(kMap)
				if err != nil {
					return Value{}, IRTypeUnknown, err
				}
				keyList = append(keyList, keyVal)
			}
		}

		for _, v := range values {
			if vMap, ok := v.(map[string]interface{}); ok {
				valVal, _, err := convertExpression(vMap)
				if err != nil {
					return Value{}, IRTypeUnknown, err
				}
				valList = append(valList, valVal)
			}
		}

		// Infer value type from the values
		elementType := IRTypeString // Default
		if len(valList) > 0 {
			elementType = valList[0].Type
		}
		dictType := IRType{Base: "dict", Key: &IRTypeString, Element: &elementType}

		return Value{
			Type: dictType,
			Kind: Dict,
			Value: map[string]interface{}{
				"keys":   keyList,
				"values": valList,
			},
		}, dictType, nil
	}

	if parser.IsList(expr) {
		elts := parser.GetListElts(expr)

		eltsList := make([]Value, 0)
		for _, e := range elts {
			if eMap, ok := e.(map[string]interface{}); ok {
				eltVal, _, err := convertExpression(eMap)
				if err != nil {
					return Value{}, IRTypeUnknown, err
				}
				eltsList = append(eltsList, eltVal)
			}
		}

		// Infer element type from the elements
		elementType := IRTypeString // Default
		if len(eltsList) > 0 {
			elementType = eltsList[0].Type
		}
		listType := IRType{Base: "list", Element: &elementType}

		return Value{
			Type: listType,
			Kind: List,
			Value: map[string]interface{}{
				"elements": eltsList,
			},
		}, listType, nil
	}

	// Handle f-strings: f"col_{i}"
	if parser.IsJoinedStr(expr) {
		parts := parser.GetJoinedStrValues(expr)
		fstringParts := make([]Value, 0)
		for _, part := range parts {
			if partMap, ok := part.(map[string]interface{}); ok {
				if parser.IsFormattedValue(partMap) {
					innerExpr := parser.GetFormattedValueExpr(partMap)
					if innerMap, ok := innerExpr.(map[string]interface{}); ok {
						val, _, err := convertExpression(innerMap)
						if err != nil {
							return Value{}, IRTypeUnknown, err
						}
						fstringParts = append(fstringParts, val)
					}
				} else if parser.IsConstant(partMap) {
					val, _, err := convertExpression(partMap)
					if err != nil {
						return Value{}, IRTypeUnknown, err
					}
					fstringParts = append(fstringParts, val)
				}
			}
		}
		return Value{
			Type: IRTypeString,
			Kind: FString,
			Value: map[string]interface{}{
				"parts": fstringParts,
			},
		}, IRTypeString, nil
	}

	// Handle list comprehensions: [x*2 for x in items if x > 0]
	if parser.IsListComp(expr) {
		elt := parser.GetListCompElt(expr)
		generators := parser.GetListCompGenerators(expr)

		// Convert element expression
		var eltVal Value
		var err error
		if eltMap, ok := elt.(map[string]interface{}); ok {
			eltVal, _, err = convertExpression(eltMap)
			if err != nil {
				return Value{}, IRTypeUnknown, err
			}
		}

		// Convert generators (usually just one for simple comprehensions)
		genList := make([]map[string]interface{}, 0)
		for _, gen := range generators {
			genInfo := make(map[string]interface{})

			// Get target (loop variable)
			target := parser.GetCompGenTarget(gen)
			if targetMap, ok := target.(map[string]interface{}); ok {
				genInfo["target"] = parser.GetNameID(targetMap)
			}

			// Get iterator
			iter := parser.GetCompGenIter(gen)
			if iterMap, ok := iter.(map[string]interface{}); ok {
				iterVal, _, err := convertExpression(iterMap)
				if err != nil {
					return Value{}, IRTypeUnknown, err
				}
				genInfo["iterator"] = iterVal
			}

			// Get if conditions (filters)
			ifs := parser.GetCompGenIfs(gen)
			ifConditions := make([]Value, 0)
			for _, ifCond := range ifs {
				if ifMap, ok := ifCond.(map[string]interface{}); ok {
					ifVal, _, err := convertExpression(ifMap)
					if err != nil {
						return Value{}, IRTypeUnknown, err
					}
					ifConditions = append(ifConditions, ifVal)
				}
			}
			genInfo["conditions"] = ifConditions

			genList = append(genList, genInfo)
		}

		return Value{
			Type: IRTypeList,
			Kind: ListComp,
			Value: map[string]interface{}{
				"element":    eltVal,
				"generators": genList,
			},
		}, IRTypeList, nil
	}

	return Value{
		Type:  IRTypeUnknown,
		Kind:  Literal,
		Value: nil,
	}, IRTypeUnknown, nil
}

func inferType(val interface{}) IRType {
	switch v := val.(type) {
	case int, int64:
		return IRTypeInt
	case float64:
		return IRTypeFloat
	case string:
		return IRTypeString
	case bool:
		return IRTypeBool
	case nil:
		return IRTypeNone
	case []interface{}:
		// Try to infer element type from the first element
		elementType := IRTypeString // Default to string
		if len(v) > 0 {
			elementType = inferType(v[0])
		}
		return IRType{Base: "list", Element: &elementType}
	case map[string]interface{}:
		// Try to infer value type from the first value
		elementType := IRTypeString // Default to string
		for _, val := range v {
			elementType = inferType(val)
			break
		}
		return IRType{Base: "dict", Key: &IRTypeString, Element: &elementType}
	default:
		// Try to handle JSON number types from parsing
		if _, ok := v.(json.Number); ok {
			// Could be int or float, default to float for safety
			return IRTypeFloat
		}
		return IRTypeUnknown
	}
}

// inferModuleCallReturnType infers the return type of a module function call
func inferModuleCallReturnType(module, funcName string) IRType {
	switch module {
	case "csv":
		switch funcName {
		case "reader":
			// csv.reader returns an iterator of lists of strings
			elementType := IRType{Base: "list", Element: &IRTypeString}
			return IRType{Base: "list", Element: &elementType}
		case "DictReader":
			// DictReader returns an iterator of dicts
			elementType := IRType{Base: "dict", Key: &IRTypeString, Element: &IRTypeString}
			return IRType{Base: "list", Element: &elementType}
		}
	case "io":
		switch funcName {
		case "StringIO":
			return IRTypeString // StringIO acts like a string
		case "BytesIO":
			return IRType{Base: "bytes"}
		}
	case "json":
		switch funcName {
		case "loads":
			return IRTypeUnknown // JSON parsing returns unknown type
		case "dumps":
			return IRTypeString
		}
	case "re":
		switch funcName {
		case "match", "search", "findall", "split":
			return IRType{Base: "list", Element: &IRTypeString}
		case "compile":
			return IRTypeUnknown // Compiled regex
		}
	}
	return IRTypeUnknown
}

// inferTypeFromExpr attempts to infer the type from an expression
func inferTypeFromExpr(expr map[string]interface{}) IRType {
	// Check for constant/literal
	if parser.IsConstant(expr) {
		val := parser.GetConstantValue(expr)
		return inferType(val)
	}

	// Check for list literal
	if parser.IsList(expr) {
		return IRTypeList
	}

	// Check for dict literal
	if parser.IsDict(expr) {
		return IRTypeDict
	}

	// Check for list comprehension
	if parser.IsListComp(expr) {
		return IRTypeList
	}

	// Check for dict comprehension
	if parser.IsDictComp(expr) {
		return IRTypeDict
	}

	// Check for comparison (always returns bool)
	if parser.IsCompare(expr) {
		return IRTypeBool
	}

	// Check for boolean operation (always returns bool)
	if parser.IsBoolOp(expr) {
		return IRTypeBool
	}

	// Check for binary operation - depends on operands
	if parser.IsBinOp(expr) {
		_, left, right := parser.GetBinOpInfo(expr)
		leftType := IRTypeUnknown
		rightType := IRTypeUnknown
		if leftMap, ok := left.(map[string]interface{}); ok {
			leftType = inferTypeFromExpr(leftMap)
		}
		if rightMap, ok := right.(map[string]interface{}); ok {
			rightType = inferTypeFromExpr(rightMap)
		}
		// If either operand is float, result is float
		if leftType.Base == "float" || rightType.Base == "float" {
			return IRTypeFloat
		}
		// If both are int, result is int
		if leftType.Base == "int" && rightType.Base == "int" {
			return IRTypeInt
		}
		// String concatenation
		if leftType.Base == "string" || rightType.Base == "string" {
			return IRTypeString
		}
	}

	// Check for unary operation
	if parser.IsUnaryOp(expr) {
		operand := parser.GetUnaryOpOperand(expr)
		if operandMap, ok := operand.(map[string]interface{}); ok {
			return inferTypeFromExpr(operandMap)
		}
	}

	// Check for call - try to infer from known functions
	if parser.IsCall(expr) {
		funcName := parser.GetCallFunc(expr)
		switch funcName {
		case "len":
			return IRTypeInt
		case "str":
			return IRTypeString
		case "int":
			return IRTypeInt
		case "float":
			return IRTypeFloat
		case "bool":
			return IRTypeBool
		case "list":
			return IRTypeList
		case "dict":
			return IRTypeDict
		case "range":
			return IRTypeList // range returns an iterable
		}
	}

	return IRTypeUnknown
}

// KnownModules is a set of Python module names that we support in complex mode
var KnownModules = map[string]bool{
	"csv":       true,
	"io":        true,
	"re":        true,
	"json":      true,
	"datetime":  true,
	"hashlib":   true,
	"base64":    true,
	"math":      true,
	"itertools": true,
	"functools": true,
	"operator":  true,
	"string":    true,
	"textwrap":  true,
	"uuid":      true,
}

// splitModuleFunction checks if a name is a known module and splits it
// Returns (module, function) if it's a module function, ("", "") otherwise
func splitModuleFunction(name string) (string, string) {
	// Check for dotted names like "csv.reader"
	for i := 0; i < len(name); i++ {
		if name[i] == '.' {
			module := name[:i]
			if KnownModules[module] {
				return module, name[i+1:]
			}
		}
	}
	return "", ""
}

// isKnownModule checks if a name is a known Python module
func isKnownModule(name string) bool {
	return KnownModules[name]
}

// convertSliceExpression converts a Python slice expression to IR
// Handles: arr[1:3], arr[:5], arr[2:], arr[::2], arr[1:5:2]
func convertSliceExpression(sliceMap map[string]interface{}) (Value, IRType, error) {
	lower := parser.GetSliceLower(sliceMap)
	upper := parser.GetSliceUpper(sliceMap)
	step := parser.GetSliceStep(sliceMap)

	sliceInfo := make(map[string]interface{})

	// Convert lower bound
	if lower != nil {
		if lowerMap, ok := lower.(map[string]interface{}); ok {
			lowerVal, _, err := convertExpression(lowerMap)
			if err != nil {
				return Value{}, IRTypeUnknown, err
			}
			sliceInfo["lower"] = lowerVal
		}
	}

	// Convert upper bound
	if upper != nil {
		if upperMap, ok := upper.(map[string]interface{}); ok {
			upperVal, _, err := convertExpression(upperMap)
			if err != nil {
				return Value{}, IRTypeUnknown, err
			}
			sliceInfo["upper"] = upperVal
		}
	}

	// Convert step
	if step != nil {
		if stepMap, ok := step.(map[string]interface{}); ok {
			stepVal, _, err := convertExpression(stepMap)
			if err != nil {
				return Value{}, IRTypeUnknown, err
			}
			sliceInfo["step"] = stepVal
		}
	}

	return Value{
		Type:  IRTypeList, // Slices return lists
		Kind:  Slice,
		Value: sliceInfo,
	}, IRTypeList, nil
}
