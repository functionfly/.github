// Example FunctionFly function in Go.
// Build: GOOS=wasip1 GOARCH=wasm go build -o hello.wasm .
package main

import (
	"encoding/json"
	"fmt"

	functionfly "github.com/functionfly/functionfly/sdk/go"
)

type HelloResponse struct {
	Message string `json:"message"`
	OK      bool   `json:"ok"`
}

type HelloFunction struct{}

func (f *HelloFunction) Handle(input string) (string, error) {
	functionfly.Log("Hello from Go function!")

	resp := HelloResponse{
		Message: "Hello from Go!",
		OK:      true,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return functionfly.Error("INTERNAL_ERROR", err.Error()), nil
	}

	return string(data), nil
}

func main() {
	functionfly.Run(&HelloFunction{})
}
