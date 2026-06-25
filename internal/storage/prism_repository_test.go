package storage

import (
	"testing"
)

func TestPrismCellTableName(t *testing.T) {
	cell := PrismCell{}
	if cell.TableName() != "prism_cells" {
		t.Errorf("expected prism_cells, got %s", cell.TableName())
	}
}

func TestPrismHeartbeatTableName(t *testing.T) {
	hb := PrismHeartbeat{}
	if hb.TableName() != "prism_heartbeats" {
		t.Errorf("expected prism_heartbeats, got %s", hb.TableName())
	}
}

func TestPrismExecutionResultTableName(t *testing.T) {
	result := PrismExecutionResult{}
	if result.TableName() != "prism_execution_results" {
		t.Errorf("expected prism_execution_results, got %s", result.TableName())
	}
}

func TestPrismCapabilityTableName(t *testing.T) {
	cap := PrismCapability{}
	if cap.TableName() != "prism_capabilities" {
		t.Errorf("expected prism_capabilities, got %s", cap.TableName())
	}
}

func TestPrismRuntimeStatusTableName(t *testing.T) {
	status := PrismRuntimeStatus{}
	if status.TableName() != "prism_runtime_status" {
		t.Errorf("expected prism_runtime_status, got %s", status.TableName())
	}
}
