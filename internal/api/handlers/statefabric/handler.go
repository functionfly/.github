package statefabric

import (
	"encoding/json"
	"net/http"

	repo "github.com/functionfly/functionfly/internal/storage/statefabric"
)

type Handler struct {
	repo *repo.Repository
}

func NewHandler(r *repo.Repository) *Handler {
	return &Handler{repo: r}
}

func notImplemented(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]string{"error": "state fabric endpoint not yet implemented"})
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request)            { notImplemented(w, r) }
func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request)          { notImplemented(w, r) }
func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request)             { notImplemented(w, r) }
func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request)          { notImplemented(w, r) }
func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request)          { notImplemented(w, r) }
func (h *Handler) HandleGetMetrics(w http.ResponseWriter, r *http.Request)      { notImplemented(w, r) }
func (h *Handler) HandleListStores(w http.ResponseWriter, r *http.Request)      { notImplemented(w, r) }
func (h *Handler) HandleCreateStore(w http.ResponseWriter, r *http.Request)     { notImplemented(w, r) }
func (h *Handler) HandleDeleteStore(w http.ResponseWriter, r *http.Request)     { notImplemented(w, r) }
func (h *Handler) HandleListPipelines(w http.ResponseWriter, r *http.Request)   { notImplemented(w, r) }
func (h *Handler) HandleCreatePipeline(w http.ResponseWriter, r *http.Request)  { notImplemented(w, r) }
func (h *Handler) HandleUpdatePipeline(w http.ResponseWriter, r *http.Request)  { notImplemented(w, r) }
func (h *Handler) HandleDeletePipeline(w http.ResponseWriter, r *http.Request)  { notImplemented(w, r) }
func (h *Handler) HandleExecutePipeline(w http.ResponseWriter, r *http.Request) { notImplemented(w, r) }
func (h *Handler) HandleListEvents(w http.ResponseWriter, r *http.Request)      { notImplemented(w, r) }
func (h *Handler) HandleListSnapshots(w http.ResponseWriter, r *http.Request)   { notImplemented(w, r) }
func (h *Handler) HandleCreateSnapshot(w http.ResponseWriter, r *http.Request)  { notImplemented(w, r) }
func (h *Handler) HandleDeleteSnapshot(w http.ResponseWriter, r *http.Request)  { notImplemented(w, r) }
func (h *Handler) HandleListReplays(w http.ResponseWriter, r *http.Request)     { notImplemented(w, r) }
func (h *Handler) HandleCreateReplay(w http.ResponseWriter, r *http.Request)    { notImplemented(w, r) }
func (h *Handler) HandleGetReplay(w http.ResponseWriter, r *http.Request)       { notImplemented(w, r) }
func (h *Handler) HandleGetStats(w http.ResponseWriter, r *http.Request)        { notImplemented(w, r) }
func (h *Handler) HandleListAll(w http.ResponseWriter, r *http.Request)         { notImplemented(w, r) }
func (h *Handler) HandleSuspendFabric(w http.ResponseWriter, r *http.Request)   { notImplemented(w, r) }
func (h *Handler) HandleResumeFabric(w http.ResponseWriter, r *http.Request)    { notImplemented(w, r) }
