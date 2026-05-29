package ghost

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logrus.WithError(err).Error("failed to encode JSON response")
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]interface{}{
		"ok":    false,
		"error": map[string]string{"code": code, "message": message},
	})
}

func toFrontendBuild(build *BuildState) map[string]interface{} {
	name := build.Goal
	if name == "" {
		name = build.Description
	}
	if name == "" {
		name = "Ghost Build"
	}
	return map[string]interface{}{
		"id":        build.ID,
		"name":      name,
		"status":    phaseToFrontendStatus(build.Phase),
		"taskId":    build.CurrentTaskID,
		"createdAt": build.StartedAt.UTC().Format(time.RFC3339),
		"updatedAt": build.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func phaseToFrontendStatus(phase GhostPhase) string {
	switch phase {
	case PhaseComplete:
		return "ready"
	case PhaseError:
		return "failed"
	case PhaseBuilding, PhaseDeploying, PhaseMonitoring, PhaseProvisioning:
		return "building"
	default:
		return "creating"
	}
}

func toFrontendTasks(build *BuildState) []map[string]interface{} {
	tasks := make([]map[string]interface{}, 0, len(build.Tasks))
	for _, task := range build.Tasks {
		status := string(task.Status)
		if build.HumanApprovalRequired && build.CurrentTaskID == task.ID {
			status = "awaiting_approval"
		}
		item := map[string]interface{}{
			"id":        task.ID,
			"status":    status,
			"title":     task.Title,
			"createdAt": build.StartedAt.UTC().Format(time.RFC3339),
			"updatedAt": build.UpdatedAt.UTC().Format(time.RFC3339),
		}
		if task.Description != "" {
			item["description"] = task.Description
		}
		if task.CompletedAt != nil {
			item["completedAt"] = task.CompletedAt.UTC().Format(time.RFC3339)
		}
		if len(task.Logs) > 0 {
			logs := make([]map[string]interface{}, 0, len(task.Logs))
			for _, log := range task.Logs {
				logs = append(logs, map[string]interface{}{
					"timestamp": log.Timestamp.UTC().Format(time.RFC3339),
					"level":     log.Level,
					"message":   log.Message,
				})
			}
			item["logs"] = logs
		}
		tasks = append(tasks, item)
	}
	return tasks
}
