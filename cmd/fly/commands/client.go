package commands

import (
	"encoding/json"
	"fmt"
	"os"
)

// vaultAPIClient is a minimal stub HTTP client used by the orphan vault CLI
// commands in this package. The ff CLI itself was moved to a separate
// repository (functionfly/ff-cli); this stub keeps the package compilable so
// the shared helpers (isUUID, parseExpiryHours, etc.) and their tests can stay
// in-tree. Replace with a real client if/when the CLI is re-introduced.
type vaultAPIClient struct{}

func (c *vaultAPIClient) Get(path string, out interface{}) error {
	return fmt.Errorf("ff CLI removed; install functionfly/ff-cli to call %s", path)
}

func (c *vaultAPIClient) Post(path string, body interface{}, out interface{}) error {
	return fmt.Errorf("ff CLI removed; install functionfly/ff-cli to call %s", path)
}

func (c *vaultAPIClient) Delete(path string, out interface{}) error {
	return fmt.Errorf("ff CLI removed; install functionfly/ff-cli to call %s", path)
}

func (c *vaultAPIClient) Patch(path string, body interface{}, out interface{}) error {
	return fmt.Errorf("ff CLI removed; install functionfly/ff-cli to call %s", path)
}

// NewAPIClient returns a stub API client for the orphan vault CLI commands.
func NewAPIClient() (*vaultAPIClient, error) {
	return &vaultAPIClient{}, nil
}

func printJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
