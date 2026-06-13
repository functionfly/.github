package vault

import "context"

// testContext is a test helper that returns a fresh background context.
// The do() helper requires a non-nil context; tests don't need
// cancellation semantics so context.Background() is fine.
func testContext(_ interface{}) context.Context { return context.Background() }
