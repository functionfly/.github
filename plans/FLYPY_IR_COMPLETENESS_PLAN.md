# FlyPy IR Generator and Rust Emitter Completeness Plan

## Problem Statement

The FlyPy compiler successfully compiles Python code but the generated Rust/WASM code doesn't implement the full Python logic. The IR generator only handles basic operations, and the Rust emitter has significant gaps in code generation.

## Current Implementation Status

### IR Generator ([`internal/flypy/ir/generator.go`](internal/flypy/ir/generator.go))

#### ✅ Implemented

| Feature | Status | Notes |
|---------|--------|-------|
| Return statements | ✅ | Basic return with value |
| Assignments | ✅ | Simple variable assignment |
| If statements | ⚠️ Partial | Converts but no proper block structure |
| Binary operations | ✅ | Add, Sub, Mult, Div, Mod, etc. |
| Unary operations | ✅ | Not, Invert, USub, UAdd |
| Comparisons | ✅ | Eq, NotEq, Lt, LtE, Gt, GtE |
| Boolean operations | ✅ | And, Or |
| Constants | ✅ | int, float, string, bool, None |
| Name references | ✅ | Variable references |
| Subscripts | ✅ | Index access |
| Attribute access | ✅ | obj.attr |
| Dict literals | ✅ | {key: value} |
| List literals | ✅ | [elem1, elem2] |
| Function calls | ⚠️ Partial | Basic calls, no module/method tracking |

#### ❌ Missing

| Feature | Priority | Complexity | Notes |
|---------|----------|------------|-------|
| For loops | HIGH | Medium | Critical for data processing |
| While loops | HIGH | Medium | Needed for iterative algorithms |
| Augmented assignment | MEDIUM | Low | +=, -=, *=, etc. |
| Break/Continue | HIGH | Low | Loop control flow |
| Try/Except | LOW | High | Error handling |
| With statements | LOW | Medium | Context managers |
| Lambda expressions | MEDIUM | Medium | Anonymous functions |
| List comprehensions | HIGH | Medium | Pythonic data transformation |
| Dict comprehensions | MEDIUM | Medium | Dict creation pattern |
| Set literals | LOW | Low | {1, 2, 3} |
| Slice operations | HIGH | Medium | arr[1:3], arr[::-1] |
| Conditional expressions | MEDIUM | Low | x if cond else y |
| Named expressions | LOW | Low | Walrus operator := |
| Starred expressions | LOW | Medium | *args, **kwargs |
| Yield/Yield from | LOW | High | Generators |

### Rust Emitter ([`internal/flypy/backend/rust_emitter.go`](internal/flypy/backend/rust_emitter.go))

#### ✅ Implemented

| Feature | Status | Notes |
|---------|--------|-------|
| Template structure | ✅ | Deterministic and Complex mode templates |
| Input/Output structs | ✅ | Serde-based JSON handling |
| Assignment emission | ✅ | let x = value; |
| Return emission | ✅ | return Output { result }; |
| Binary operations | ✅ | +, -, *, /, % |
| Built-in functions | ⚠️ Partial | len, abs, str, int, float, bool |
| Regex helpers | ✅ | re_match, re_search, re_sub, re_findall, re_split |
| CSV helpers | ✅ | csv_reader, csv_writer |
| IO helpers | ✅ | StringIO, BytesIO structs |
| Hash helpers | ✅ | hash_sha256, hash_md5 |
| Base64 helpers | ✅ | base64_encode, base64_decode |

#### ❌ Missing

| Feature | Priority | Complexity | Notes |
|---------|----------|------------|-------|
| If statement emission | HIGH | Medium | if/else blocks |
| For loop emission | HIGH | Medium | for x in iter { } |
| While loop emission | HIGH | Medium | while cond { } |
| Comparison emission | HIGH | Low | ==, !=, <, <=, >, >= |
| Boolean op emission | HIGH | Low | &&, \|\| |
| Unary op emission | MEDIUM | Low | !, - |
| Subscript emission | HIGH | Low | arr[index], dict[key] |
| Dict literal emission | MEDIUM | Low | HashMap or json! macro |
| List literal emission | MEDIUM | Low | vec! macro |
| Slice emission | HIGH | Medium | arr[1..3] |
| Method call emission | HIGH | High | obj.method args |
| Module call emission | HIGH | High | csv.reader, io.StringIO |
| JSON operations | HIGH | Low | json.loads, json.dumps |
| Datetime operations | MEDIUM | Medium | chrono crate integration |
| List comprehension | HIGH | High | Rust iterator patterns |

## Implementation Plan

### Phase 1: Core Control Flow (Critical)

#### 1.1 For Loops in IR Generator

**File**: [`internal/flypy/ir/generator.go`](internal/flypy/ir/generator.go)

```go
// Add to convertStatement function
} else if parser.IsFor(stmt) {
    target := parser.GetForTarget(stmt)
    iter := parser.GetForIter(stmt)
    body := parser.GetForBody(stmt)
    
    // Create loop variable
    targetName := parser.GetNameID(target.(map[string]interface{}))
    
    // Convert iterator expression
    iterVal, _, err := convertExpression(iter.(map[string]interface{}))
    if err != nil {
        return nil, err
    }
    
    ops = append(ops, Operation{
        Type:     "for",
        Result:   targetName,
        Operands: []Value{iterVal},
    })
    
    // Convert body with proper nesting
    for _, bodyStmt := range body {
        if bodyStmtMap, ok := bodyStmt.(map[string]interface{}); ok {
            bodyOps, err := convertStatement(bodyStmtMap)
            if err != nil {
                return nil, err
            }
            ops = append(ops, bodyOps...)
        }
    }
    
    // Add end marker
    ops = append(ops, Operation{Type: "endfor"})
}
```

#### 1.2 For Loops in Rust Emitter

**File**: [`internal/flypy/backend/rust_emitter.go`](internal/flypy/backend/rust_emitter.go)

```go
func generateOperation(op ir.Operation) string {
    switch op.Type {
    // ... existing cases ...
    case "for":
        iterExpr := generateValue(op.Operands[0])
        return fmt.Sprintf("    for %s in %s {\n", op.Result, iterExpr)
    case "endfor":
        return "    }\n"
    }
}
```

#### 1.3 While Loops

Similar pattern to for loops with test condition.

#### 1.4 If/Else Blocks

Need proper block structure with endif markers:

```go
// IR
ops = append(ops, Operation{Type: "if", Operands: []Value{testVal}})
// ... body operations ...
ops = append(ops, Operation{Type: "endif"})
```

### Phase 2: Expression Completeness

#### 2.1 Comparison Emission

```go
func generateValue(val ir.Value) string {
    // ... existing cases ...
    case ir.Compare:
        if compVal, ok := val.Value.(map[string]interface{}); ok {
            left := generateValue(compVal["left"].(ir.Value))
            comps := compVal["comparators"].([]ir.Value)
            ops := compVal["ops"].([]string)
            
            // Handle chained comparisons: a < b < c
            var parts []string
            parts = append(parts, fmt.Sprintf("(%s %s %s)", 
                left, pyOpToRustOp(ops[0]), generateValue(comps[0])))
            
            for i := 1; i < len(comps); i++ {
                parts = append(parts, fmt.Sprintf("(%s %s %s)",
                    generateValue(comps[i-1]), pyOpToRustOp(ops[i]), generateValue(comps[i])))
            }
            return strings.Join(parts, " && ")
        }
    }
}
```

#### 2.2 Boolean Operations

```go
case ir.BoolOp:
    if boolVal, ok := val.Value.(map[string]interface{}); ok {
        op := boolVal["op"].(string)  // "And" or "Or"
        values := boolVal["values"].([]ir.Value)
        
        rustOp := "&&"
        if op == "Or" {
            rustOp = "||"
        }
        
        var parts []string
        for _, v := range values {
            parts = append(parts, generateValue(v))
        }
        return strings.Join(parts, fmt.Sprintf(" %s ", rustOp))
    }
```

#### 2.3 Subscript Emission

```go
case ir.Subscript:
    if subVal, ok := val.Value.(map[string]interface{}); ok {
        value := generateValue(subVal["value"].(ir.Value))
        index := generateValue(subVal["index"].(ir.Value))
        return fmt.Sprintf("%s[%s]", value, index)
    }
```

#### 2.4 Dict and List Literals

```go
case ir.Dict:
    if dictVal, ok := val.Value.(map[string]interface{}); ok {
        keys := dictVal["keys"].([]ir.Value)
        values := dictVal["values"].([]ir.Value)
        
        var entries []string
        for i, k := range keys {
            entries = append(entries, fmt.Sprintf("(%s, %s)", 
                generateValue(k), generateValue(values[i])))
        }
        return fmt.Sprintf("json!({%s})", strings.Join(entries, ", "))
    }

case ir.List:
    if listVal, ok := val.Value.(map[string]interface{}); ok {
        elements := listVal["elements"].([]ir.Value)
        
        var elts []string
        for _, e := range elements {
            elts = append(elts, generateValue(e))
        }
        return fmt.Sprintf("vec![%s]", strings.Join(elts, ", "))
    }
```

### Phase 3: Module and Method Calls

#### 3.1 Module Call Detection in IR

```go
// In convertExpression, handle module calls
if parser.IsCall(expr) {
    funcName := parser.GetCallFunc(expr)
    
    // Check if it's a module call like csv.reader or io.StringIO
    if strings.Contains(funcName, ".") {
        parts := strings.SplitN(funcName, ".", 2)
        return Value{
            Type: IRTypeUnknown,
            Kind: ModuleCall,
            Value: map[string]interface{}{
                "module": parts[0],
                "func":   parts[1],
                "args":   operands,
            },
        }, IRTypeUnknown, nil
    }
    // ... existing code ...
}
```

#### 3.2 Method Call Detection

```go
// Handle method calls like output.getvalue
if parser.IsAttribute(expr) {
    // Check if this is being called
    attr := parser.GetAttributeAttr(expr)
    value := parser.GetAttributeValue(expr)
    
    // This might be a method reference
    return Value{
        Type: IRTypeUnknown,
        Kind: MethodCall,
        Value: map[string]interface{}{
            "receiver": valueVal,
            "method":   attr,
        },
    }, IRTypeUnknown, nil
}
```

#### 3.3 Module Call Emission in Rust

```go
func generateModuleCall(module, funcName string, args []string) string {
    switch module {
    case "csv":
        switch funcName {
        case "reader":
            return fmt.Sprintf("csv_reader(%s)", args[0])
        case "writer":
            return fmt.Sprintf("csv_writer(%s)", args[0])
        }
    case "io":
        switch funcName {
        case "StringIO":
            if len(args) > 0 {
                return fmt.Sprintf("StringIO::from_string(%s)", args[0])
            }
            return "StringIO::new()"
        case "BytesIO":
            return "BytesIO::new()"
        }
    case "json":
        switch funcName {
        case "loads":
            return fmt.Sprintf("serde_json::from_str(%s).unwrap()", args[0])
        case "dumps":
            return fmt.Sprintf("serde_json::to_string(%s).unwrap()", args[0])
        }
    case "re":
        switch funcName {
        case "match":
            return fmt.Sprintf("re_match(%s, %s)", args[0], args[1])
        case "search":
            return fmt.Sprintf("re_search(%s, %s)", args[0], args[1])
        case "sub":
            return fmt.Sprintf("re_sub(%s, %s, %s)", args[0], args[1], args[2])
        }
    }
    return fmt.Sprintf("/* unknown module call: %s.%s */", module, funcName)
}
```

### Phase 4: List Comprehensions

#### 4.1 IR for List Comprehensions

```go
// Add new operation type
OpListComp = "list_comp"

// In convertExpression
if parser.IsListComp(expr) {
    elt := parser.GetListCompElt(expr)
    generators := parser.GetListCompGenerators(expr)
    
    // First generator is the main one
    target := generators[0].Target
    iter := generators[0].Iter
    ifs := generators[0].Ifs
    
    eltVal, _, _ := convertExpression(elt)
    iterVal, _, _ := convertExpression(iter)
    
    return Value{
        Type: IRTypeList,
        Kind: ListComp,
        Value: map[string]interface{}{
            "element":   eltVal,
            "target":    target,
            "iterator":  iterVal,
            "conditions": ifs,
        },
    }, IRTypeList, nil
}
```

#### 4.2 Rust Emission for List Comprehensions

```go
case ir.ListComp:
    if compVal, ok := val.Value.(map[string]interface{}); ok {
        element := generateValue(compVal["element"].(ir.Value))
        target := compVal["target"].(string)
        iterator := generateValue(compVal["iterator"].(ir.Value))
        
        // Python: [x*2 for x in items if x > 0]
        // Rust: items.iter().filter(|x| **x > 0).map(|x| x * 2).collect()
        
        var chain strings.Builder
        chain.WriteString(fmt.Sprintf("%s.iter()", iterator))
        
        // Add filter conditions
        if conditions, ok := compVal["conditions"].([]ir.Value); ok {
            for _, cond := range conditions {
                condStr := generateValue(cond)
                chain.WriteString(fmt.Sprintf(".filter(|%s| %s)", target, condStr))
            }
        }
        
        chain.WriteString(fmt.Sprintf(".map(|%s| %s).collect::<Vec<_>>()", target, element))
        return chain.String()
    }
```

### Phase 5: Slice Operations

```go
// Slice in IR
type Slice struct {
    Start *Value
    Stop  *Value
    Step  *Value
}

// Rust emission for slices
func generateSlice(value string, slice Slice) string {
    start := "0"
    stop := fmt.Sprintf("%s.len()", value)
    step := "1"
    
    if slice.Start != nil {
        start = generateValue(*slice.Start)
    }
    if slice.Stop != nil {
        stop = generateValue(*slice.Stop)
    }
    
    if slice.Step != nil && generateValue(*slice.Step) != "1" {
        // Complex slice with step
        step = generateValue(*slice.Step)
        return fmt.Sprintf("%s[%s..%s].iter().step_by(%s).cloned().collect::<Vec<_>>()", 
            value, start, stop, step)
    }
    
    return fmt.Sprintf("%s[%s..%s].to_vec()", value, start, stop)
}
```

## Architecture Improvements

### 1. Block-Structured IR

Current IR is flat. Need hierarchical structure:

```go
type Block struct {
    Statements []Statement
}

type Statement struct {
    Type     string
    Value    interface{}
    Body     *Block      // For if/for/while
    ElseBody *Block      // For if/else
}

type IfStatement struct {
    Condition Value
    ThenBlock *Block
    ElseBlock *Block
}

type ForStatement struct {
    Target   string
    Iterator Value
    Body     *Block
}
```

### 2. Type Tracking

Add proper type inference:

```go
type TypeContext struct {
    Variables map[string]IRType
    Returns   IRType
}

func (tc *TypeContext) InferType(expr Value) IRType {
    // Type inference logic
}
```

### 3. Scope Management

Track variable scopes for proper Rust emission:

```go
type Scope struct {
    Parent   *Scope
    Variables map[string]VariableInfo
}

type VariableInfo struct {
    Type     IRType
    Mutable  bool
    Borrowed bool
}
```

## Testing Strategy

### Unit Tests

1. Test each IR conversion function independently
2. Test each Rust emission function independently
3. Compare generated Rust code against expected output

### Integration Tests

1. Compile Python functions and execute the resulting WASM
2. Compare output against Python interpreter output
3. Test edge cases and error conditions

### Test Cases

```python
# Test for loop
def test_for(input):
    result = []
    for i in input["items"]:
        result.append(i * 2)
    return {"result": result}

# Test list comprehension
def test_listcomp(input):
    return {"doubled": [x * 2 for x in input["items"] if x > 0]}

# Test nested loops
def test_nested(input):
    result = []
    for row in input["matrix"]:
        for val in row:
            result.append(val)
    return {"flattened": result}

# Test if/else
def test_if(input):
    if input["value"] > 10:
        return {"status": "large"}
    else:
        return {"status": "small"}

# Test dict operations
def test_dict(input):
    data = input["data"]
    return {
        "keys": list(data.keys()),
        "values": list(data.values()),
        "has_key": "name" in data
    }

# Test CSV operations
def test_csv(input):
    import csv
    import io
    
    output = io.StringIO()
    writer = csv.writer(output)
    for row in input["rows"]:
        writer.writerow(row)
    
    return {"csv": output.getvalue()}
```

## Implementation Order

1. **Week 1**: Core control flow (for, while, if/else blocks)
2. **Week 2**: Expression completeness (comparisons, booleans, subscripts)
3. **Week 3**: Module and method calls
4. **Week 4**: List comprehensions and slices
5. **Week 5**: Testing and bug fixes

## Success Criteria

1. All test cases pass with correct output
2. Generated Rust code compiles without errors
3. WASM execution matches Python interpreter output
4. No runtime panics in generated code
5. Proper error messages for unsupported features
