package wasm

import "sync"

// DelegatingHostHandler forwards host calls to a delegate that can be swapped per execution.
// WASM runtimes bind host functions once at init; this wrapper allows tenant/fabric-scoped handlers.
type DelegatingHostHandler struct {
	mu       sync.RWMutex
	delegate HostFunctionHandler
}

func NewDelegatingHostHandler(initial HostFunctionHandler) *DelegatingHostHandler {
	if initial == nil {
		initial = NewDefaultHostHandler(nil)
	}
	return &DelegatingHostHandler{delegate: initial}
}

func (d *DelegatingHostHandler) SetDelegate(handler HostFunctionHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if handler == nil {
		handler = NewDefaultHostHandler(nil)
	}
	d.delegate = handler
}

func (d *DelegatingHostHandler) current() HostFunctionHandler {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.delegate
}

func (d *DelegatingHostHandler) Log(message string) { d.current().Log(message) }

func (d *DelegatingHostHandler) Fetch(request string) (string, error) {
	return d.current().Fetch(request)
}

func (d *DelegatingHostHandler) KVGet(key string) (string, error) {
	return d.current().KVGet(key)
}

func (d *DelegatingHostHandler) KVSet(key, value string) error {
	return d.current().KVSet(key, value)
}

func (d *DelegatingHostHandler) GetEnv(name string) string {
	return d.current().GetEnv(name)
}

func (d *DelegatingHostHandler) AIInference(model string, input []byte, params string) (string, error) {
	return d.current().AIInference(model, input, params)
}

func (d *DelegatingHostHandler) StateGet(path string) (string, error) {
	return d.current().StateGet(path)
}

func (d *DelegatingHostHandler) StateSet(path string, value string) error {
	return d.current().StateSet(path, value)
}

func (d *DelegatingHostHandler) StateDelete(path string) error {
	return d.current().StateDelete(path)
}

func (d *DelegatingHostHandler) StateGetFabric(fabricID string) (string, error) {
	return d.current().StateGetFabric(fabricID)
}

func (d *DelegatingHostHandler) StateCreateSnapshot(path string, label string) (string, error) {
	return d.current().StateCreateSnapshot(path, label)
}

func (d *DelegatingHostHandler) GetAttestation(attestationID string) (string, error) {
	return d.current().GetAttestation(attestationID)
}

func (d *DelegatingHostHandler) Delegate(targetFunctionID string, input string, options string) (string, error) {
	return d.current().Delegate(targetFunctionID, input, options)
}
