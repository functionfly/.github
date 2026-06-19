package studio

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ProjectsHandler handles Studio project and file HTTP requests.
type ProjectsHandler struct {
	repo *ProjectRepository
}

// NewProjectsHandler creates a new projects handler.
func NewProjectsHandler(repo *ProjectRepository) *ProjectsHandler {
	return &ProjectsHandler{repo: repo}
}

func (h *ProjectsHandler) scope(r *http.Request) (workspaceScope, bool) {
	tenantID := getTenantID(r)
	userID := getUserID(r)
	if tenantID == "" || userID == "" {
		return workspaceScope{}, false
	}
	return workspaceScope{
		TenantID:    tenantID,
		UserID:      userID,
		Environment: getEnvironment(r),
	}, true
}

// HandleGetWorkspace GET /v1/studio/workspace
func (h *ProjectsHandler) HandleGetWorkspace(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.scope(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	state, err := h.repo.GetWorkspace(r.Context(), scope)
	if err != nil {
		logrus.WithError(err).Error("studio projects: failed to get workspace")
		writeJSONError(w, http.StatusInternalServerError, "Failed to load workspace")
		return
	}

	writeJSON(w, http.StatusOK, state)
}

// HandleSaveWorkspaceSession PUT /v1/studio/workspace/session
func (h *ProjectsHandler) HandleSaveWorkspaceSession(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.scope(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		ActiveProjectID *string `json:"active_project_id"`
		ActiveFileID    *string `json:"active_file_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.repo.SaveWorkspaceSession(r.Context(), scope, req.ActiveProjectID, req.ActiveFileID); err != nil {
		logrus.WithError(err).Warn("studio projects: failed to save workspace session")
		writeJSONError(w, http.StatusBadRequest, "Failed to save workspace session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"active_project_id": req.ActiveProjectID,
		"active_file_id":    req.ActiveFileID,
	})
}

// HandleCreateProject POST /v1/studio/projects
func (h *ProjectsHandler) HandleCreateProject(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.scope(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req struct {
		Name           string `json:"name"`
		StarterContent string `json:"starter_content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	project, err := h.repo.CreateProject(r.Context(), scope, req.Name, req.StarterContent)
	if err != nil {
		logrus.WithError(err).Error("studio projects: failed to create project")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create project")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"project": project})
}

// HandleGetProject GET /v1/studio/projects/{id}
func (h *ProjectsHandler) HandleGetProject(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.scope(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := mux.Vars(r)["id"]
	project, err := h.repo.GetProject(r.Context(), scope, projectID)
	if err != nil {
		logrus.WithError(err).Error("studio projects: failed to get project")
		writeJSONError(w, http.StatusInternalServerError, "Failed to get project")
		return
	}
	if project == nil {
		writeJSONError(w, http.StatusNotFound, "Project not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"project": project})
}

// HandleUpdateProject PATCH /v1/studio/projects/{id}
func (h *ProjectsHandler) HandleUpdateProject(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.scope(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := mux.Vars(r)["id"]
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	project, err := h.repo.UpdateProject(r.Context(), scope, projectID, req.Name)
	if err != nil {
		logrus.WithError(err).Error("studio projects: failed to update project")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update project")
		return
	}
	if project == nil {
		writeJSONError(w, http.StatusNotFound, "Project not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"project": project})
}

// HandleDeleteProject DELETE /v1/studio/projects/{id}
func (h *ProjectsHandler) HandleDeleteProject(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.scope(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := mux.Vars(r)["id"]
	if err := h.repo.DeleteProject(r.Context(), scope, projectID); err != nil {
		logrus.WithError(err).Warn("studio projects: failed to delete project")
		writeJSONError(w, http.StatusNotFound, "Project not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Project deleted"})
}

// HandleDuplicateProject POST /v1/studio/projects/{id}/duplicate
func (h *ProjectsHandler) HandleDuplicateProject(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.scope(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := mux.Vars(r)["id"]
	project, err := h.repo.DuplicateProject(r.Context(), scope, projectID)
	if err != nil {
		logrus.WithError(err).Error("studio projects: failed to duplicate project")
		writeJSONError(w, http.StatusInternalServerError, "Failed to duplicate project")
		return
	}
	if project == nil {
		writeJSONError(w, http.StatusNotFound, "Project not found")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"project": project})
}

// HandleCreateFile POST /v1/studio/projects/{id}/files
func (h *ProjectsHandler) HandleCreateFile(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.scope(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	projectID := mux.Vars(r)["id"]
	var req struct {
		Name    string `json:"name"`
		Dir     string `json:"dir"`
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	file, err := h.repo.CreateFile(r.Context(), scope, projectID, req.Name, req.Dir, req.Content)
	if err != nil {
		logrus.WithError(err).Error("studio projects: failed to create file")
		writeJSONError(w, http.StatusInternalServerError, "Failed to create file")
		return
	}
	if file == nil {
		writeJSONError(w, http.StatusNotFound, "Project not found")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"file": file})
}

// HandleUpdateFile PATCH /v1/studio/projects/{projectId}/files/{fileId}
func (h *ProjectsHandler) HandleUpdateFile(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.scope(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	projectID := vars["id"]
	fileID := vars["fileId"]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	file, err := h.repo.UpdateFile(r.Context(), scope, projectID, fileID, updates)
	if err != nil {
		logrus.WithError(err).Error("studio projects: failed to update file")
		writeJSONError(w, http.StatusInternalServerError, "Failed to update file")
		return
	}
	if file == nil {
		writeJSONError(w, http.StatusNotFound, "File not found")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"file": file})
}

// HandleDeleteFile DELETE /v1/studio/projects/{projectId}/files/{fileId}
func (h *ProjectsHandler) HandleDeleteFile(w http.ResponseWriter, r *http.Request) {
	scope, ok := h.scope(r)
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	projectID := vars["id"]
	fileID := vars["fileId"]

	if err := h.repo.DeleteFile(r.Context(), scope, projectID, fileID); err != nil {
		logrus.WithError(err).Warn("studio projects: failed to delete file")
		writeJSONError(w, http.StatusBadRequest, "Failed to delete file")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "File deleted"})
}
