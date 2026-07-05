//go:build !cgo

package functions

import (
	"context"
	"fmt"

	"github.com/functionfly/functionfly/internal/api/types"
	"github.com/functionfly/functionfly/internal/storage"
)

func executeTestFunctionImpl(ctx context.Context, h *Handler, req *types.TestFunctionRequest, user *storage.User) (*types.TestFunctionResponse, error) {
	return &types.TestFunctionResponse{
		Success:         false,
		Output:          nil,
		ExecutionTimeMs: 0,
		Logs: []*storage.FunctionLog{{
			Message: fmt.Sprintf("test-function execution requires CGO (rebuild with CGO_ENABLED=1): functionId=%v", req.FunctionId),
		}},
	}, nil
}