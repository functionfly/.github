package flymachines

import (
	"context"
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type MachineCode struct {
	SourceCode string `json:"source"`
	Input     string `json:"input"`
	Manifest  string `json:"manifest,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type MachineResult struct {
	Result   string `json:"result"`
	ExitCode int    `json:"exit_code"`
	Error    string `json:"error,omitempty"`
}

type FlyMachinesHandler struct {
	redisClient  *redis.Client
	machineStore sync.Map
	machineSecret string
}

func NewFlyMachinesHandler(redisClient *redis.Client) *FlyMachinesHandler {
	secret := os.Getenv("FUNCTIONFLY_MACHINE_SECRET")
	if secret == "" {
		logrus.Warn("FUNCTIONFLY_MACHINE_SECRET not set - Machine auth will be insecure")
	}
	return &FlyMachinesHandler{
		redisClient:  redisClient,
		machineSecret: secret,
	}
}

func (h *FlyMachinesHandler) machineCodeKey(executionID string) string {
	return fmt.Sprintf("flymachines:code:%s", executionID)
}

func (h *FlyMachinesHandler) machineResultKey(executionID string) string {
	return fmt.Sprintf("flymachines:result:%s", executionID)
}

func (h *FlyMachinesHandler) StoreMachineCode(ctx context.Context, executionID string, code *MachineCode) error {
	code.CreatedAt = time.Now()

	if h.redisClient != nil {
		data, err := json.Marshal(code)
		if err != nil {
			return fmt.Errorf("failed to marshal code: %w", err)
		}
		return h.redisClient.Set(ctx, h.machineCodeKey(executionID), data, 10*time.Minute).Err()
	}

	h.machineStore.Store(executionID, code)
	return nil
}

func (h *FlyMachinesHandler) GetMachineCode(ctx context.Context, executionID string) (*MachineCode, error) {
	if h.redisClient != nil {
		data, err := h.redisClient.Get(ctx, h.machineCodeKey(executionID)).Bytes()
		if err == redis.Nil {
			return nil, fmt.Errorf("code not found for execution %s", executionID)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get code from redis: %w", err)
		}
		var code MachineCode
		if err := json.Unmarshal(data, &code); err != nil {
			return nil, fmt.Errorf("failed to unmarshal code: %w", err)
		}
		return &code, nil
	}

	val, ok := h.machineStore.Load(executionID)
	if !ok {
		return nil, fmt.Errorf("code not found for execution %s", executionID)
	}
	code, ok := val.(*MachineCode)
	if !ok {
		return nil, fmt.Errorf("invalid code stored for execution %s", executionID)
	}
	return code, nil
}

func (h *FlyMachinesHandler) StoreMachineResult(ctx context.Context, executionID string, result *MachineResult) error {
	if h.redisClient != nil {
		data, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		return h.redisClient.Set(ctx, h.machineResultKey(executionID), data, 10*time.Minute).Err()
	}

	h.machineStore.Store(executionID+"_result", result)
	return nil
}

func (h *FlyMachinesHandler) GetMachineResult(ctx context.Context, executionID string) (*MachineResult, error) {
	if h.redisClient != nil {
		data, err := h.redisClient.Get(ctx, h.machineResultKey(executionID)).Bytes()
		if err == redis.Nil {
			return nil, fmt.Errorf("result not found for execution %s", executionID)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to get result from redis: %w", err)
		}
		var result MachineResult
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
		return &result, nil
	}

	val, ok := h.machineStore.Load(executionID + "_result")
	if !ok {
		return nil, fmt.Errorf("result not found for execution %s", executionID)
	}
	result, ok := val.(*MachineResult)
	if !ok {
		return nil, fmt.Errorf("invalid result stored for execution %s", executionID)
	}
	return result, nil
}

func (h *FlyMachinesHandler) DeleteMachineCode(ctx context.Context, executionID string) error {
	if h.redisClient != nil {
		return h.redisClient.Del(ctx, h.machineCodeKey(executionID)).Err()
	}
	h.machineStore.Delete(executionID)
	return nil
}

func (h *FlyMachinesHandler) verifyMachineSecret(r *http.Request) bool {
	if h.machineSecret == "" {
		return true
	}
	sig := r.Header.Get("X-Machine-Secret")
	if sig == "" {
		sig = r.Header.Get("Fly-Machine-Secret")
	}
	if sig == "" {
		return false
	}
	expected := h.machineSecret
	return hmac.Equal([]byte(sig), []byte(expected))
}

func (h *FlyMachinesHandler) HandleServeCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.verifyMachineSecret(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	executionID := vars["execution_id"]
	if executionID == "" {
		http.Error(w, "execution_id is required", http.StatusBadRequest)
		return
	}

	code, err := h.GetMachineCode(r.Context(), executionID)
	if err != nil {
		logrus.WithError(err).WithField("execution_id", executionID).Warn("Failed to get machine code")
		http.Error(w, "Code not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(code); err != nil {
		logrus.WithError(err).Error("Failed to encode code response")
	}
}

func (h *FlyMachinesHandler) HandleReceiveResult(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !h.verifyMachineSecret(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	executionID := vars["execution_id"]
	if executionID == "" {
		http.Error(w, "execution_id is required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		logrus.WithError(err).Error("Failed to read result body")
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var result MachineResult
	if err := json.Unmarshal(body, &result); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal result")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.StoreMachineResult(r.Context(), executionID, &result); err != nil {
		logrus.WithError(err).Error("Failed to store machine result")
		http.Error(w, "Failed to store result", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

type MachineStopRequest struct {
	MachineID string `json:"machine_id"`
	Reason    string `json:"reason,omitempty"`
}

func (h *FlyMachinesHandler) HandleStopMachine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	executionID := vars["execution_id"]
	if executionID == "" {
		http.Error(w, "execution_id is required", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "stopping", "execution_id": executionID})
}

type MachineHealthRequest struct {
	MachineID string `json:"machine_id"`
	Status    string `json:"status"`
}

func (h *FlyMachinesHandler) HandleMachineHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	vars := mux.Vars(r)
	executionID := vars["execution_id"]
	if executionID == "" {
		http.Error(w, "execution_id is required", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var health MachineHealthRequest
	if err := json.Unmarshal(body, &health); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": executionID,
		"machine_id":   health.MachineID,
		"status":       health.Status,
	}).Debug("Machine health update")

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type RegisterExecutionRequest struct {
	ExecutionID uuid.UUID `json:"execution_id"`
	SourceCode  string    `json:"source_code"`
	Input      string    `json:"input"`
	Manifest   string    `json:"manifest,omitempty"`
}

func (h *FlyMachinesHandler) HandleRegisterExecution(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	var req RegisterExecutionRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	code := &MachineCode{
		SourceCode: req.SourceCode,
		Input:      req.Input,
		Manifest:   req.Manifest,
	}

	if err := h.StoreMachineCode(r.Context(), req.ExecutionID.String(), code); err != nil {
		logrus.WithError(err).Error("Failed to store machine code")
		http.Error(w, "Failed to store code", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}
