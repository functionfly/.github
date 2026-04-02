package manifest

import "testing"

func TestValidate_RuntimePython312AndGo121(t *testing.T) {
	t.Parallel()
	for _, rt := range []string{"python3.12", "go1.21"} {
		m := &Manifest{
			Name:    "hello-fn",
			Version: "1.0.0",
			Runtime: rt,
		}
		if err := m.Validate(); err != nil {
			t.Fatalf("runtime %q: %v", rt, err)
		}
	}
}

func TestValidate_EntryExtensionGo(t *testing.T) {
	t.Parallel()
	m := &Manifest{
		Name:    "hello-fn",
		Version: "1.0.0",
		Runtime: "go1.21",
		Entry:   "handler.go",
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_InvalidRuntime(t *testing.T) {
	t.Parallel()
	m := &Manifest{Name: "x", Version: "1.0.0", Runtime: "python3.10"}
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for python3.10")
	}
}
