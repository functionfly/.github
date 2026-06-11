module github.com/functionfly/functionfly/cmd/flypy-go

go 1.25.0

require github.com/functionfly/functionfly v0.0.0

require (
	github.com/bytecodealliance/wasmtime-go/v19 v19.0.0 // indirect
	github.com/sirupsen/logrus v1.9.3 // indirect
	golang.org/x/sys v0.42.0 // indirect
)

replace github.com/functionfly/functionfly => ../../
