// gen_dre_node_key generates an Ed25519 private key for DRE node signing and prints base64.
// Usage: go run scripts/gen_dre_node_key.go
// Then set DRE_NODE_PRIVATE_KEY=<output> or write output to a file and set DRE_NODE_PRIVATE_KEY_PATH.
//go:build ignore

package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func generateKey() {
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	fmt.Println(base64.StdEncoding.EncodeToString(priv))
}

func runKeyGen() {
	generateKey()
}

func main() {
	runKeyGen()
}
