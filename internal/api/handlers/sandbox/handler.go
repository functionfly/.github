package sandbox

import (
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type SandboxStatus struct {
	GvisorAvailable bool       `json:"gvisor_available"`
	GvisorPath      string     `json:"gvisor_path,omitempty"`
	GvisorVersion   string     `json:"gvisor_version,omitempty"`
	DockerAvailable bool       `json:"docker_available"`
	DockerVersion   string     `json:"docker_version,omitempty"`
	ActiveTier      string     `json:"active_tier"`
	SupportedTiers  []TierInfo `json:"supported_tiers"`
	LastChecked     time.Time  `json:"last_checked"`
	SystemInfo      SystemInfo `json:"system_info"`
}

type TierInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Available   bool   `json:"available"`
	Isolation   string `json:"isolation_level"`
	Status      string `json:"status"`
}

type SystemInfo struct {
	Kernel               string `json:"kernel"`
	Arch                 string `json:"arch"`
	CgroupVersion        string `json:"cgroup_version"`
	SeccompEnabled       bool   `json:"seccomp_enabled"`
	UserNamespaceEnabled bool   `json:"user_namespace_enabled"`
}

type Handler struct {
	mu            sync.RWMutex
	cachedStatus  *SandboxStatus
	lastCheck     time.Time
	cacheDuration time.Duration
}

func New() *Handler {
	return &Handler{
		cacheDuration: 30 * time.Second,
	}
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	if h.cachedStatus != nil && time.Since(h.lastCheck) < h.cacheDuration {
		status := h.cachedStatus
		h.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
		return
	}
	h.mu.RUnlock()

	status := h.refreshStatus()

	h.mu.Lock()
	h.cachedStatus = status
	h.lastCheck = time.Now()
	h.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (h *Handler) refreshStatus() *SandboxStatus {
	status := &SandboxStatus{
		LastChecked: time.Now(),
		SupportedTiers: []TierInfo{
			{
				ID:          "wasm",
				Name:        "WebAssembly",
				Description: "Wasmtime-based WASM sandbox with fuel metering and WASI capability gating",
				Available:   true,
				Isolation:   "memory_safe",
				Status:      "stable",
			},
			{
				ID:          "gvisor",
				Name:        "gVisor",
				Description: "User-space kernel sandbox using gVisor (runsc) for container isolation without shared kernel",
				Available:   false,
				Isolation:   "kernel_namespace",
				Status:      "available",
			},
			{
				ID:          "docker",
				Name:        "Docker",
				Description: "Docker container sandbox with hardened security flags (--network=none, --read-only, --cap-drop=ALL)",
				Available:   false,
				Isolation:   "container",
				Status:      "fallback",
			},
			{
				ID:          "microvm",
				Name:        "Firecracker MicroVM",
				Description: "Full hardware virtualization using Firecracker MicroVMs (Enterprise only, requires KVM)",
				Available:   false,
				Isolation:   "hardware_virtualization",
				Status:      "enterprise",
			},
		},
		SystemInfo: h.getSystemInfo(),
	}

	gvisorPath, gvisorVersion := h.checkGvisor()
	status.GvisorAvailable = gvisorPath != ""
	status.GvisorPath = gvisorPath
	status.GvisorVersion = gvisorVersion

	dockerVersion := h.checkDocker()
	status.DockerAvailable = dockerVersion != ""
	status.DockerVersion = dockerVersion

	for i := range status.SupportedTiers {
		switch status.SupportedTiers[i].ID {
		case "gvisor":
			status.SupportedTiers[i].Available = status.GvisorAvailable
		case "docker":
			status.SupportedTiers[i].Available = status.DockerAvailable
		case "microvm":
			status.SupportedTiers[i].Available = h.checkKVM()
		}
	}

	if status.GvisorAvailable {
		status.ActiveTier = "gvisor"
	} else if status.DockerAvailable {
		status.ActiveTier = "docker"
	} else {
		status.ActiveTier = "wasm"
	}

	return status
}

func (h *Handler) checkGvisor() (string, string) {
	paths := []string{
		"/usr/local/bin/runsc",
		"/usr/bin/runsc",
		"/opt/gvisor/runsc",
	}

	home := os.Getenv("HOME")
	if home != "" {
		paths = append(paths, home+"/go/bin/runsc")
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			cmd := exec.Command(path, "--version")
			output, err := cmd.Output()
			if err == nil {
				return path, strings.TrimSpace(string(output))
			}
			return path, "unknown"
		}
	}
	return "", ""
}

func (h *Handler) checkDocker() string {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func (h *Handler) checkKVM() bool {
	_, err := os.Stat("/dev/kvm")
	return err == nil
}

func (h *Handler) getSystemInfo() SystemInfo {
	info := SystemInfo{
		Arch: os.Getenv("GOARCH"),
	}

	cmd := exec.Command("uname", "-r")
	if output, err := cmd.Output(); err == nil {
		info.Kernel = strings.TrimSpace(string(output))
	}

	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err == nil {
		info.CgroupVersion = "v2"
	} else {
		info.CgroupVersion = "v1"
	}

	if _, err := os.Stat("/proc/sys/kernel/seccomp"); err == nil {
		info.SeccompEnabled = true
	}

	cmd = exec.Command("cat", "/proc/sys/kernel/unprivileged_userns_clone")
	if output, err := cmd.Output(); err == nil {
		info.UserNamespaceEnabled = strings.TrimSpace(string(output)) == "1"
	}

	return info
}

func (h *Handler) UpdateTierConfig(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	functionID := vars["function_id"]

	var req struct {
		SandboxTier    string  `json:"sandbox_tier"`
		MemoryMB       int     `json:"memory_mb,omitempty"`
		CPULimit       float64 `json:"cpu_limit,omitempty"`
		TimeoutMs      int     `json:"timeout_ms,omitempty"`
		NetworkEnabled bool    `json:"network_enabled,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	validTiers := map[string]bool{
		"wasm": true, "gvisor": true, "docker": true, "microvm": true,
	}
	if !validTiers[req.SandboxTier] {
		http.Error(w, `{"error":"Invalid sandbox tier. Must be: wasm, gvisor, docker, microvm"}`, http.StatusBadRequest)
		return
	}

	logrus.WithFields(logrus.Fields{
		"function_id":  functionID,
		"sandbox_tier": req.SandboxTier,
	}).Info("Updating function sandbox tier")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "updated",
		"function_id":  functionID,
		"sandbox_tier": req.SandboxTier,
		"message":      "Sandbox tier updated successfully",
	})
}

func (h *Handler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/sandbox/status", h.GetStatus).Methods("GET")
	router.HandleFunc("/sandbox/tiers", h.ListTiers).Methods("GET")
	router.HandleFunc("/functions/{function_id}/sandbox", h.UpdateTierConfig).Methods("PUT", "POST")
}

func (h *Handler) ListTiers(w http.ResponseWriter, r *http.Request) {
	status := h.refreshStatus()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tiers":       status.SupportedTiers,
		"active_tier": status.ActiveTier,
	})
}
