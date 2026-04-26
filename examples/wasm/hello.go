// FunctionFly example: Hello from Go
// Build: GOOS=wasip1 GOARCH=wasm go build -o hello.wasm .
// Manifest: {"name":"hello-go","version":"1.0.0","runtime":"go","entry":"hello.go"}
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Response struct {
	Message string `json:"message"`
	InputLen int   `json:"input_length"`
	OK       bool   `json:"ok"`
}

func main() {
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
		}
		if err != nil {
			break
		}
	}

	resp := Response{
		Message:  "Hello from Go!",
		InputLen: len(buf),
		OK:       true,
	}

	data, _ := json.Marshal(resp)
	fmt.Print(string(data))
}
