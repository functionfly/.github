# Plan: Full Python Feature Support for FlyPy

## Executive Summary

All Python features (for loops, while loops, list comprehensions, try/except) are already implemented end-to-end in the FlyPy compiler. The only issue is **mode restrictions** in [`internal/flypy/restrictions/enforcer.go`](internal/flypy/restrictions/enforcer.go) that limit these features to "complex" mode only.

This plan removes those restrictions to provide full support in all modes.

## Current Implementation Status

| Feature | IR Generator | Backend | Restrictions Enforcer |
|---------|--------------|---------|----------------------|
| For loops | ✅ [`ir/generator.go:370`](internal/flypy/ir/generator.go:370) | ✅ [`backend/operations.go:418`](internal/flypy/backend/operations.go:418) | ⚠️ Only in complex mode |
| While loops | ✅ [`ir/generator.go:428`](internal/flypy/ir/generator.go:428) | ✅ [`backend/operations.go:452`](internal/flypy/backend/operations.go:452) | ⚠️ Only in complex mode |
| List comprehensions | ✅ [`ir/generator.go:938`](internal/flypy/ir/generator.go:938) | ✅ [`backend/values.go:307`](internal/flypy/backend/values.go:307) | ⚠️ Not explicitly handled |
| Try/except | ✅ [`ir/generator.go:495`](internal/flypy/ir/generator.go:495) | ✅ [`backend/operations.go:474`](internal/flypy/backend/operations.go:474) | ⚠️ Not explicitly handled |

## Required Changes

### 1. Update [`internal/flypy/restrictions/enforcer.go`](internal/flypy/restrictions/enforcer.go)

#### 1.1 For Loops (lines 327-337)
**Current code:**
```go
// Check for for loops - allowed in complex mode
if parser.IsFor(stmt) {
    if mode == ModeDeterministic {
        *errors = append(*errors, CompileError{
            Type:    ForbiddenFeature,
            Message: "for loops are not supported in deterministic mode; use --mode complex",
            Line:    0,
        })
    }
    // In complex mode, for loops are allowed
}
```

**Change:** Remove the restriction so for loops are allowed in all modes.

#### 1.2 While Loops (lines 339-349)
**Current code:**
```go
// Check for while loops - allowed in complex mode with limits
if parser.IsWhile(stmt) {
    if mode == ModeDeterministic {
        *errors = append(*errors, CompileError{
            Type:    ForbiddenFeature,
            Message: "while loops are not supported in deterministic mode; use --mode complex",
            Line:    0,
        })
    }
    // In complex mode, while loops are allowed (with runtime limits)
}
```

**Change:** Remove the restriction so while loops are allowed in all modes.

#### 1.3 Add List Comprehension Handling
**Add to `checkStatementWithMode` function:**
```go
// Check for list comprehensions in expressions
if parser.IsListComp(stmt) {
    // List comprehensions are allowed in all modes
}
```

Actually, list comprehensions appear in expressions, not statements. The check should be added to `checkExpressionWithMode` function instead.

#### 1.4 Add Try/Except Handling
**Add to `checkStatementWithMode` function:**
```go
// Check for try/except blocks - allowed in all modes
if parser.IsTry(stmt) {
    // Try/except is allowed in all modes
    // Check the body and handlers recursively
    for _, tryStmt := range parser.GetTryBody(stmt) {
        if tryStmtMap, ok := tryStmt.(map[string]interface{}); ok {
            checkStatementWithMode(tryStmtMap, errors, mode)
        }
    }
    // Check exception handlers
    for _, handler := range parser.GetTryHandlers(stmt) {
        if handlerMap, ok := handler.(map[string]interface{}); ok {
            for _, handlerStmt := range parser.GetHandlerBody(handlerMap) {
                if handlerStmtMap, ok := handlerStmt.(map[string]interface{}); ok {
                    checkStatementWithMode(handlerStmtMap, errors, mode)
                }
            }
        }
    }
    // Check finally block
    for _, finallyStmt := range parser.GetTryFinalbody(stmt) {
        if finallyStmtMap, ok := finallyStmt.(map[string]interface{}); ok {
            checkStatementWithMode(finallyStmtMap, errors, mode)
        }
    }
}
```

### 2. Update Tests

#### 2.1 Add tests in [`internal/flypy/restrictions/complex_mode_test.go`](internal/flypy/restrictions/complex_mode_test.go)

- Add test for for loops in deterministic mode
- Add test for while loops in deterministic mode  
- Add test for list comprehensions in deterministic mode
- Add test for try/except in deterministic mode

### 3. Update Documentation

#### 3.1 Update [`plans/MVP_GAP_ANALYSIS.md`](plans/MVP_GAP_ANALYSIS.md)

Update the Python Feature Coverage table to reflect full support:

| Feature | Status | Notes |
|---------|--------|-------|
| For loops | ✅ Full Support | All modes |
| While loops | ✅ Full Support | All modes |
| List comprehensions | ✅ Full Support | All modes |
| Try/except blocks | ✅ Full Support | All modes |

## Implementation Steps (Code Mode)

1. **Modify `checkStatementWithMode`** - Remove for/while loop restrictions
2. **Add list comprehension check** - Add to `checkExpressionWithMode`  
3. **Add try/except check** - Add to `checkStatementWithMode`
4. **Add tests** - Add unit tests in `complex_mode_test.go`
5. **Verify** - Run `go test ./internal/flypy/...`
6. **Update docs** - Update `MVP_GAP_ANALYSIS.md`

## Risk Assessment

- **Low Risk**: These are removal of restrictions, not new feature implementations
- **Existing tests should pass**: The backend already handles these features
- **Mode detection**: No changes needed to mode detection logic
