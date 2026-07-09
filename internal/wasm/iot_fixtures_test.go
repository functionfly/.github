package wasm

import (
	"sync"
	"testing"
)

var tinyWasmStub = []byte{
	0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00,
}

var iotTestModuleOnce struct {
	sync.Once
	wasm []byte
	err  error
}

func iotTestModule(tb testing.TB) []byte {
	tb.Helper()
	iotTestModuleOnce.Once.Do(func() {
		iotTestModuleOnce.wasm = tinyWasmStub
	})
	return iotTestModuleOnce.wasm
}

func loadIotTestModule(tb testing.TB, runtime *WASM3IoTRuntime) {
	tb.Helper()
	wasm := iotTestModule(tb)
	if err := runtime.LoadModule(wasm); err != nil {
		tb.Skipf("LoadModule(tinyStub) failed: %v", err)
	}
}

func TestIoTFixtures_GetAllFixtures(t *testing.T) {
	fixtures := GetAllFixtures()

	if len(fixtures) != 5 {
		t.Errorf("expected 5 fixtures, got %d", len(fixtures))
	}

	expectedNames := []string{"sensor", "actuator", "state_machine", "memory_stress", "error"}
	for i, expected := range expectedNames {
		if fixtures[i].Name != expected {
			t.Errorf("expected fixture[%d].Name=%s, got %s", i, expected, fixtures[i].Name)
		}
	}
}

func TestIoTFixtures_GetFixtureByType(t *testing.T) {
	testCases := []struct {
		fixtureType IoTFixtureType
		expectedName string
	}{
		{FixtureSensor, "sensor"},
		{FixtureActuator, "actuator"},
		{FixtureStateMachine, "state_machine"},
		{FixtureMemoryStress, "memory_stress"},
		{FixtureError, "error"},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedName, func(t *testing.T) {
			fixture := GetFixtureByType(tc.fixtureType)
			if fixture == nil {
				t.Fatalf("expected fixture for type %s, got nil", tc.fixtureType)
			}
			if fixture.Name != tc.expectedName {
				t.Errorf("expected name %s, got %s", tc.expectedName, fixture.Name)
			}
		})
	}

	invalidType := IoTFixtureType("invalid")
	if fixture := GetFixtureByType(invalidType); fixture != nil {
		t.Error("expected nil for invalid fixture type")
	}
}

func TestIoTFixtures_GetFixtureByName(t *testing.T) {
	testCases := []struct {
		name        string
		expectedType IoTFixtureType
	}{
		{"sensor", FixtureSensor},
		{"actuator", FixtureActuator},
		{"state_machine", FixtureStateMachine},
		{"memory_stress", FixtureMemoryStress},
		{"error", FixtureError},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fixture := GetFixtureByName(tc.name)
			if fixture == nil {
				t.Fatalf("expected fixture for name %s, got nil", tc.name)
			}
			if fixture.Type != tc.expectedType {
				t.Errorf("expected type %s, got %s", tc.expectedType, fixture.Type)
			}
		})
	}

	if fixture := GetFixtureByName("nonexistent"); fixture != nil {
		t.Error("expected nil for nonexistent fixture name")
	}
}

func TestIoTFixtures_ValidateFixture(t *testing.T) {
	if err := ValidateFixture(nil); err == nil {
		t.Error("expected error for nil fixture")
	}

	validFixture := &SensorWasm
	if err := ValidateFixture(validFixture); err != nil {
		t.Errorf("expected no error for valid fixture: %v", err)
	}

	invalidFixture := &IoTFixture{
		Name: "invalid",
		Bytecode: []byte{0x00, 0x00},
	}
	if err := ValidateFixture(invalidFixture); err == nil {
		t.Error("expected error for invalid bytecode")
	}

	invalidMagic := &IoTFixture{
		Name: "invalid_magic",
		Bytecode: []byte{
			0x00, 0x00, 0x00, 0x00,
			0x01, 0x00, 0x00, 0x00,
		},
	}
	if err := ValidateFixture(invalidMagic); err == nil {
		t.Error("expected error for invalid magic number")
	}
}

func TestIoTFixtures_FixtureToHex(t *testing.T) {
	fixture := &SensorWasm
	hexStr := FixtureToHex(fixture)

	if len(hexStr) != len(fixture.Bytecode)*2 {
		t.Errorf("hex string length mismatch: expected %d, got %d", len(fixture.Bytecode)*2, len(hexStr))
	}

	decoded, err := HexToFixture(hexStr)
	if err != nil {
		t.Fatalf("failed to decode hex: %v", err)
	}

	if len(decoded.Bytecode) != len(fixture.Bytecode) {
		t.Errorf("decoded bytecode length mismatch")
	}
}

func TestIoTFixtures_HexToFixture(t *testing.T) {
	invalidHex := "not-valid-hex!"
	if _, err := HexToFixture(invalidHex); err == nil {
		t.Error("expected error for invalid hex string")
	}

	validHex := "0061736d01000000"
	fixture, err := HexToFixture(validHex)
	if err != nil {
		t.Fatalf("failed to decode valid hex: %v", err)
	}

	expected := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}
	if len(fixture.Bytecode) != len(expected) {
		t.Errorf("bytecode length mismatch")
	}
}

func TestIoTFixtureSuite_NewAndValidate(t *testing.T) {
	suite := NewIoTFixtureSuite("iot-tests")
	if suite.Name != "iot-tests" {
		t.Errorf("expected suite name iot-tests, got %s", suite.Name)
	}

	if len(suite.Fixtures) != 5 {
		t.Errorf("expected 5 fixtures in suite, got %d", len(suite.Fixtures))
	}

	if err := suite.Validate(); err != nil {
		t.Errorf("suite validation failed: %v", err)
	}
}

func TestIoTFixtures_AllValid(t *testing.T) {
	fixtures := GetAllFixtures()

	for _, fixture := range fixtures {
		if err := ValidateFixture(&fixture); err != nil {
			t.Errorf("fixture %s validation failed: %v", fixture.Name, err)
		}

		if len(fixture.Bytecode) < 8 {
			t.Errorf("fixture %s bytecode too short: %d bytes", fixture.Name, len(fixture.Bytecode))
		}

		if fixture.Name == "" {
			t.Error("fixture has empty name")
		}

		if fixture.Description == "" {
			t.Errorf("fixture %s has empty description", fixture.Name)
		}
	}
}

func TestIoTFixtures_BytecodeIntegrity(t *testing.T) {
	fixtures := GetAllFixtures()

	for _, fixture := range fixtures {
		wasm := fixture.Bytecode

		if wasm[0] != 0x00 || wasm[1] != 0x61 || wasm[2] != 0x73 || wasm[3] != 0x6d {
			t.Errorf("fixture %s has invalid WASM magic number", fixture.Name)
		}

		version := uint32(wasm[4]) | uint32(wasm[5])<<8 | uint32(wasm[6])<<16 | uint32(wasm[7])<<24
		if version != 1 && version != 13 {
			t.Errorf("fixture %s has unsupported WASM version: %d", fixture.Name, version)
		}
	}
}
