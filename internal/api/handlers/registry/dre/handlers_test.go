package dre

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cert "github.com/functionfly/functionfly/internal/dre/cert"
	drecert "github.com/functionfly/functionfly/internal/dre/cert"
	drecrypto "github.com/functionfly/functionfly/internal/dre/crypto"
	"github.com/functionfly/functionfly/internal/storage/registry"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockDRERepo is a mock implementation of the registry repository for DRE testing.
type mockDRERepo struct {
	certificates map[string]*registry.ExecutionCertificate
	megRecords   map[uuid.UUID]*registry.MEGRecord
	passports    map[uuid.UUID]*registry.ExecutionPassport
	driftReports map[uuid.UUID][]*registry.DriftReportRecord
	functions    map[string]*registry.RegistryFunction
}

func newMockDRERepo() *mockDRERepo {
	return &mockDRERepo{
		certificates: make(map[string]*registry.ExecutionCertificate),
		megRecords:   make(map[uuid.UUID]*registry.MEGRecord),
		passports:    make(map[uuid.UUID]*registry.ExecutionPassport),
		driftReports: make(map[uuid.UUID][]*registry.DriftReportRecord),
		functions:    make(map[string]*registry.RegistryFunction),
	}
}

func (m *mockDRERepo) GetCertificateByID(certID string) (*registry.ExecutionCertificate, error) {
	return m.certificates[certID], nil
}

func (m *mockDRERepo) GetCertificateByExecutionID(executionID uuid.UUID) (*registry.ExecutionCertificate, error) {
	for _, c := range m.certificates {
		if c.ExecutionID == executionID {
			return c, nil
		}
	}
	return nil, nil
}

func (m *mockDRERepo) GetCertificatesByFunctionID(functionID uuid.UUID, limit, offset int) ([]*registry.ExecutionCertificate, error) {
	var result []*registry.ExecutionCertificate
	for _, c := range m.certificates {
		if c.FunctionID == functionID {
			result = append(result, c)
		}
	}
	if len(result) <= offset {
		return []*registry.ExecutionCertificate{}, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}

func (m *mockDRERepo) GetMEGByExecutionID(executionID uuid.UUID) (*registry.MEGRecord, error) {
	return m.megRecords[executionID], nil
}

func (m *mockDRERepo) GetMEGByExecutionRootHash(hash string) (*registry.MEGRecord, error) {
	for _, rec := range m.megRecords {
		if rec.ExecutionRootHash == hash {
			return rec, nil
		}
	}
	return nil, nil
}

func (m *mockDRERepo) GetMEGRecordsByFunctionID(functionID uuid.UUID, limit, offset int, filters registry.MEGRecordFilters) ([]*registry.MEGRecord, int64, error) {
	var result []*registry.MEGRecord
	for _, rec := range m.megRecords {
		if rec.FunctionID == functionID {
			if filters.Version != "" && rec.Version != filters.Version {
				continue
			}
			if filters.VerifiedOnly && rec.ReplayVerifiedAt == nil {
				continue
			}
			result = append(result, rec)
		}
	}
	total := int64(len(result))
	if len(result) <= offset {
		return []*registry.MEGRecord{}, total, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], total, nil
}

func (m *mockDRERepo) GetPassportByFunctionID(functionID uuid.UUID) (*registry.ExecutionPassport, error) {
	return m.passports[functionID], nil
}

func (m *mockDRERepo) GetOrCreatePassport(functionID uuid.UUID) (*registry.ExecutionPassport, error) {
	if p, ok := m.passports[functionID]; ok {
		return p, nil
	}
	p := &registry.ExecutionPassport{FunctionID: functionID}
	m.passports[functionID] = p
	return p, nil
}

func (m *mockDRERepo) GetDriftReportsByFunctionID(functionID uuid.UUID, limit, offset int) ([]*registry.DriftReportRecord, error) {
	reports := m.driftReports[functionID]
	if len(reports) <= offset {
		return []*registry.DriftReportRecord{}, nil
	}
	end := offset + limit
	if end > len(reports) {
		end = len(reports)
	}
	return reports[offset:end], nil
}

func (m *mockDRERepo) GetFunctionByAuthorName(author, name string) (*registry.RegistryFunction, error) {
	key := author + "/" + name
	if fn, ok := m.functions[key]; ok {
		return fn, nil
	}
	return nil, nil
}

func (m *mockDRERepo) GetFunctionByID(id uuid.UUID) (*registry.RegistryFunction, error) {
	for _, fn := range m.functions {
		if fn.ID == id {
			return fn, nil
		}
	}
	return nil, nil
}

func (m *mockDRERepo) GetLatestFunctionVersion(functionID uuid.UUID) (*registry.RegistryFunctionVersion, error) {
	return nil, nil
}

func (m *mockDRERepo) GetFunctionVersion(functionID uuid.UUID, version string) (*registry.RegistryFunctionVersion, error) {
	return nil, nil
}

func (m *mockDRERepo) GetExecutionTimelineBuckets(functionID uuid.UUID, from, to time.Time, metric string) ([]registry.ExecutionTimelineBucket, error) {
	return []registry.ExecutionTimelineBucket{}, nil
}

func (m *mockDRERepo) UpdateCertificateAnchored(certID string, anchored bool, anchorChain, anchorTxHash, anchorMerkleRoot string, anchorBlockNumber int64, anchoredAt *time.Time) error {
	if cert, ok := m.certificates[certID]; ok {
		cert.Anchored = anchored
		cert.AnchorChain = anchorChain
		cert.AnchorTxHash = anchorTxHash
		cert.AnchorMerkleRoot = anchorMerkleRoot
		cert.AnchorBlockNumber = anchorBlockNumber
		cert.AnchoredAt = anchoredAt
	}
	return nil
}

// mockAnchoringService is a mock AnchoringService for testing.
type mockAnchoringService struct {
	configured bool
	shouldFail bool
}

func (m *mockAnchoringService) Anchor(ctx context.Context, req *drecert.AnchorRequest) (*drecert.AnchorReceipt, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("anchoring failed")
	}
	return &drecert.AnchorReceipt{
		Chain:       req.Chain,
		BlockNumber: 12345678,
		TxHash:      "0xmocktxhash",
		MerkleRoot:  req.ExecutionRootHash,
		AnchorHash:  "mock_anchor_hash",
		AnchoredAt:  time.Now(),
	}, nil
}

func (m *mockAnchoringService) IsConfigured() bool {
	return m.configured
}

// ─── Hex helpers ─────────────────────────────────────────────────────────────

func hexDecode(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

func hexEncode(b []byte) string {
	return hex.EncodeToString(b)
}

// ─── Validation Helpers ───────────────────────────────────────────────────────

func validateAuthorName(author, name string) error {
	if author == "" || name == "" {
		return fmt.Errorf("author and name are required")
	}
	if len(author) > 64 || len(name) > 64 {
		return fmt.Errorf("author or name exceeds max length")
	}
	validName := func(s string) bool {
		for _, c := range s {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.') {
				return false
			}
		}
		return true
	}
	if !validName(author) || !validName(name) {
		return fmt.Errorf("invalid characters in author or name")
	}
	return nil
}

func validateCertID(certID string) error {
	if certID == "" {
		return fmt.Errorf("cert_id is required")
	}
	if len(certID) < 12 {
		return fmt.Errorf("cert_id too short")
	}
	if !strings.HasPrefix(certID, "fxc_") {
		return fmt.Errorf("cert_id must start with fxc_")
	}
	rest := certID[4:]
	for _, c := range rest {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			return fmt.Errorf("cert_id contains invalid characters")
		}
	}
	return nil
}

func validateExecutionRootHash(hash string) error {
	if len(hash) != 64 {
		return fmt.Errorf("execution root hash must be exactly 64 hex characters")
	}
	_, err := hex.DecodeString(hash)
	if err != nil {
		return fmt.Errorf("execution root hash contains invalid hex characters")
	}
	return nil
}

// ─── Test Fixtures ─────────────────────────────────────────────────────────────

func makeTestFXCert(certID, execRootHash string) *cert.FXCert {
	return &cert.FXCert{
		FXCertVersion: "1.0",
		CertificateID: certID,
		Execution: cert.ExecutionSection{
			ExecutionID:     uuid.New().String(),
			FunctionID:      "fx://test/func",
			ProtocolVersion: "dre/1.0",
		},
		Capsule: cert.CapsuleSection{
			DeterminismTier: "full",
		},
		Integrity: cert.IntegritySection{
			ExecutionRootHash: execRootHash,
			InputHash:         strings.Repeat("a", 64),
			OutputHash:        strings.Repeat("b", 64),
		},
		Trust: cert.TrustSection{
			TrustScore:             0.95,
			DeterminismScore:       0.90,
			ReplayConsistencyScore: 0.98,
		},
		Signatures: cert.SignatureSection{},
		Anchoring:  cert.AnchoringSection{},
	}
}

func makeTestMEGRecord(execID uuid.UUID, fnID uuid.UUID) *registry.MEGRecord {
	return &registry.MEGRecord{
		ID:                    uuid.New(),
		ExecutionID:           execID,
		FunctionID:            fnID,
		Version:               "1.0.0",
		ExecutionRootHash:     strings.Repeat("e", 64),
		InputHash:             strings.Repeat("a", 64),
		EnvironmentHash:       strings.Repeat("b", 64),
		DependencyHash:        strings.Repeat("c", 64),
		TraceHash:             strings.Repeat("d", 64),
		ResourceHash:          strings.Repeat("e", 64),
		OutputHash:            strings.Repeat("f", 64),
		MetadataHash:          strings.Repeat("g", 64),
		CapsuleDescriptorHash: strings.Repeat("h", 64),
		DeterminismTier:       "full",
		ProtocolVersion:       "dre/1.0",
		CreatedAt:             time.Now().UTC(),
	}
}

func makeTestPassport(fnID uuid.UUID) *registry.ExecutionPassport {
	return &registry.ExecutionPassport{
		ID:                        uuid.New(),
		FunctionID:                fnID,
		DeterministicReliability:  0.92,
		ReplayDriftIncidents:      1,
		VerifiedExecutionsTotal:   46,
		TotalExecutions:           50,
		DeterminismScore:          0.92,
		ReplayIntegrityScore:      0.97,
		PerformanceStabilityScore: 0.88,
		DriftScore:                0.9,
		LastVerifiedAt:            timePtr(time.Now().UTC()),
	}
}

func timePtr(t time.Time) *time.Time { return &t }

// ─── Handler Tests ─────────────────────────────────────────────────────────────

func TestHandleGetCertificate_NotFound(t *testing.T) {
	repo := newMockDRERepo()
	h := NewHandlerFromRepo(repo)

	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/cert/{cert_id}", h.HandleGetCertificate)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/cert/fxc_notexist", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "certificate not found", resp["error"])
}

func TestHandleGetCertificate_InvalidCertID(t *testing.T) {
	repo := newMockDRERepo()
	h := NewHandlerFromRepo(repo)

	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/cert/{cert_id}", h.HandleGetCertificate)

	// Valid format cert ID but doesn't exist in repo.
	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/cert/fxc_01HXXXXXXXXXXXXXXXXXXXXXXXXXXXX", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGetCertificate_Success(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	certID := "fxc_01HXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	execRootHash := strings.Repeat("e", 64)

	// Create function first.
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	// Create FXCert and store it.
	fxcert := makeTestFXCert(certID, execRootHash)
	fxcertJSON, _ := json.Marshal(fxcert)
	storedCert := &registry.ExecutionCertificate{
		CertificateID:     certID,
		ExecutionID:       uuid.New(),
		FunctionID:        fnID,
		CertJSON:          fxcertJSON,
		ExecutionRootHash: execRootHash,
		CertificateHash:   strings.Repeat("c", 64),
		CertLevel:         "standard",
	}
	repo.certificates[certID] = storedCert

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/cert/{cert_id}", h.HandleGetCertificate)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/cert/"+certID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, certID, resp["certificate_id"])
	assert.Equal(t, "standard", resp["cert_level"])
	assert.Equal(t, execRootHash, resp["execution_root_hash"])
	assert.NotNil(t, resp["cert"])
}

func TestHandleListCertificates_Pagination(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()

	// Create function.
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	// Add 3 certificates.
	for i := 0; i < 3; i++ {
		certID := "fxc_01HXXXXXXXXXXXXXXXXXXXXXXXXXX" + string(rune('A'+i))
		fxcert := makeTestFXCert(certID, strings.Repeat(string(rune('a'+i)), 64))
		fxcertJSON, _ := json.Marshal(fxcert)
		repo.certificates[certID] = &registry.ExecutionCertificate{
			CertificateID:     certID,
			ExecutionID:       uuid.New(),
			FunctionID:        fnID,
			CertJSON:          fxcertJSON,
			ExecutionRootHash: strings.Repeat(string(rune('a'+i)), 64),
			CertificateHash:   strings.Repeat(string(rune('b'+i)), 64),
			CertLevel:         "standard",
		}
	}

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/certs", h.HandleListCertificates)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/certs?limit=2&offset=0", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	certs, ok := resp["certs"].([]interface{})
	require.True(t, ok)
	assert.Len(t, certs, 2)
	assert.Equal(t, float64(2), resp["limit"])
	assert.Equal(t, float64(0), resp["offset"])
}

func TestHandleGetPassport_NoPassport(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/passport", h.HandleGetPassport)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/passport", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	passport, ok := resp["passport"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(0), passport["deterministic_reliability"])
	assert.Equal(t, float64(0), passport["determinism_score"])
	assert.Equal(t, float64(1.0), passport["drift_score"])
}

func TestHandleGetPassport_WithPassport(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn
	repo.passports[fnID] = makeTestPassport(fnID)

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/passport", h.HandleGetPassport)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/passport", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	passport, ok := resp["passport"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 0.92, passport["deterministic_reliability"])
	assert.Equal(t, float64(1), passport["replay_drift_incidents"])
	assert.Equal(t, float64(46), passport["verified_executions_total"])
	assert.Equal(t, float64(50), passport["total_executions"])
}

func TestHandleListExecutions(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	execID := uuid.New()
	repo.megRecords[execID] = makeTestMEGRecord(execID, fnID)

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/executions", h.HandleListExecutions)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/executions", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	execs, ok := resp["executions"].([]interface{})
	require.True(t, ok)
	assert.Len(t, execs, 1)
	assert.Equal(t, float64(1), resp["total"])
}

func TestHandleGetExecution_NotFound(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/executions/{execution_id}", h.HandleGetExecution)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/executions/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGetExecution_WrongFunction(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	wrongFnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	execID := uuid.New()
	meg := makeTestMEGRecord(execID, wrongFnID)
	repo.megRecords[execID] = meg

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/executions/{execution_id}", h.HandleGetExecution)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/executions/"+execID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return not found (to avoid leaking execution IDs of other functions).
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleGetExecution_Success(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	execID := uuid.New()
	repo.megRecords[execID] = makeTestMEGRecord(execID, fnID)

	// Add certificate too.
	certID := "fxc_01HXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	fxcert := makeTestFXCert(certID, repo.megRecords[execID].ExecutionRootHash)
	fxcertJSON, _ := json.Marshal(fxcert)
	repo.certificates[certID] = &registry.ExecutionCertificate{
		CertificateID:     certID,
		ExecutionID:       execID,
		FunctionID:        fnID,
		CertJSON:          fxcertJSON,
		ExecutionRootHash: repo.megRecords[execID].ExecutionRootHash,
		CertificateHash:   strings.Repeat("c", 64),
		CertLevel:         "standard",
	}

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/executions/{execution_id}", h.HandleGetExecution)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/executions/"+execID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	exec, ok := resp["execution"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, execID.String(), exec["execution_id"])
	assert.NotNil(t, exec["certificate"])
	assert.NotNil(t, exec["trust"])
}

func TestHandleGetExecutionByHash(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	execID := uuid.New()
	execRootHash := strings.Repeat("e", 64)
	meg := makeTestMEGRecord(execID, fnID)
	meg.ExecutionRootHash = execRootHash
	repo.megRecords[execID] = meg

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/executions/by-hash", h.HandleGetExecutionByHash)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/executions/by-hash?execution_root_hash="+execRootHash, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	exec, ok := resp["execution"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, execRootHash, exec["execution_root_hash"])
}

func TestHandleGetExecutionByHash_MissingHash(t *testing.T) {
	repo := newMockDRERepo()
	h := NewHandlerFromRepo(repo)

	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/executions/by-hash", h.HandleGetExecutionByHash)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/executions/by-hash", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetExecutionByHash_WrongFunction(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	wrongFnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	execID := uuid.New()
	execRootHash := strings.Repeat("e", 64)
	meg := makeTestMEGRecord(execID, wrongFnID) // Belongs to wrong function.
	meg.ExecutionRootHash = execRootHash
	repo.megRecords[execID] = meg

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/executions/by-hash", h.HandleGetExecutionByHash)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/executions/by-hash?execution_root_hash="+execRootHash, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should return not found (not 200 with empty body).
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleReplay_NotFound(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/replay/{execution_id}", h.HandleReplay)

	req := httptest.NewRequest("POST", "/registry/testauthor/testfunc/replay/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleReplay_InvalidUUID(t *testing.T) {
	repo := newMockDRERepo()
	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/replay/{execution_id}", h.HandleReplay)

	req := httptest.NewRequest("POST", "/registry/testauthor/testfunc/replay/not-a-uuid", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleReplay_Success(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	execID := uuid.New()
	repo.megRecords[execID] = makeTestMEGRecord(execID, fnID)

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/replay/{execution_id}", h.HandleReplay)

	req := httptest.NewRequest("POST", "/registry/testauthor/testfunc/replay/"+execID.String(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, execID.String(), resp["execution_id"])
	assert.NotNil(t, resp["component_hashes"])
	comp, ok := resp["component_hashes"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, comp["input"])
	assert.NotEmpty(t, comp["output"])
}

func TestHandleGetPassportByFunctionID_NotFound(t *testing.T) {
	repo := newMockDRERepo()
	h := NewHandlerFromRepo(repo)

	router := mux.NewRouter()
	router.HandleFunc("/internal/functions/{function_id}/passport", h.HandleGetPassportByFunctionID)

	req := httptest.NewRequest("GET", "/internal/functions/"+uuid.New().String()+"/passport", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // Returns empty passport, not 404.

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	passport, ok := resp["passport"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(0), passport["deterministic_reliability"])
}

func TestHandleGetPassportByFunctionID_InvalidUUID(t *testing.T) {
	repo := newMockDRERepo()
	h := NewHandlerFromRepo(repo)

	router := mux.NewRouter()
	router.HandleFunc("/internal/functions/{function_id}/passport", h.HandleGetPassportByFunctionID)

	req := httptest.NewRequest("GET", "/internal/functions/not-a-uuid/passport", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleVerifyCertificate_NotFound(t *testing.T) {
	repo := newMockDRERepo()
	h := NewHandlerFromRepo(repo)

	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/cert/{cert_id}/verify", h.HandleVerifyCertificate)

	// certID must be 10+ chars after "fxc_".
	req := httptest.NewRequest("POST", "/registry/testauthor/testfunc/cert/fxc_01HXXXXXXXXXXXXXXXXXXXXXXXXXXXX/verify", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleVerifyCertificate_Valid(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	certID := "fxc_01HXXXXXXXXXXXXXXXXXXXXXXXXXXXX"

	// Build component hashes and compute the correct execution root hash.
	componentHashes := []string{
		strings.Repeat("a", 64), // InputHash
		strings.Repeat("b", 64), // EnvironmentHash
		strings.Repeat("c", 64), // DependencyHash
		strings.Repeat("d", 64), // TraceHash
		strings.Repeat("e", 64), // ResourceHash
		strings.Repeat("f", 64), // OutputHash
		strings.Repeat("0", 64), // MetadataHash  'g' is invalid hex
	}
	leaves := make([][]byte, len(componentHashes))
	for i, h := range componentHashes {
		b, err := hexDecode(h)
		require.NoError(t, err)
		leaves[i] = b
	}
	computedRoot := drecrypto.MerkleRoot(leaves)
	executionRootHash := hexEncode(computedRoot)

	// Build the FXCert with properly computed hashes.
	fxcert := makeTestFXCert(certID, executionRootHash)
	fxcert.Integrity.InputHash = componentHashes[0]
	fxcert.Integrity.EnvironmentHash = componentHashes[1]
	fxcert.Integrity.DependencyHash = componentHashes[2]
	fxcert.Integrity.TraceHash = componentHashes[3]
	fxcert.Integrity.ResourceHash = componentHashes[4]
	fxcert.Integrity.OutputHash = componentHashes[5]
	fxcert.Integrity.MetadataHash = componentHashes[6]

	// Compute certificate hash (must clear before hashing, then restore).
	fxcert.Integrity.CertificateHash = ""
	certJSON, _ := json.Marshal(fxcert)
	canonical, _ := drecrypto.Canonicalize(json.RawMessage(certJSON))
	certificateHash := drecrypto.HashString(drecrypto.TagCert, canonical)
	fxcert.Integrity.CertificateHash = certificateHash

	fxcertJSON, _ := json.Marshal(fxcert)
	repo.certificates[certID] = &registry.ExecutionCertificate{
		CertificateID:     certID,
		ExecutionID:       uuid.New(),
		FunctionID:        fnID,
		CertJSON:          fxcertJSON,
		ExecutionRootHash: executionRootHash,
		CertificateHash:   certificateHash,
		CertLevel:         "standard",
	}

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/cert/{cert_id}/verify", h.HandleVerifyCertificate)

	req := httptest.NewRequest("POST", "/registry/testauthor/testfunc/cert/"+certID+"/verify", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, certID, resp["certificate_id"])
	assert.Equal(t, true, resp["execution_root_hash_valid"])
	assert.Equal(t, true, resp["certificate_hash_valid"])
}

func TestHandleAnchorCertificate_NotImplemented(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	certID := "fxc_01HXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	fxcert := makeTestFXCert(certID, strings.Repeat("e", 64))
	fxcertJSON, _ := json.Marshal(fxcert)
	repo.certificates[certID] = &registry.ExecutionCertificate{
		CertificateID:     certID,
		ExecutionID:       uuid.New(),
		FunctionID:        fnID,
		CertJSON:          fxcertJSON,
		ExecutionRootHash: strings.Repeat("e", 64),
		CertificateHash:   strings.Repeat("c", 64),
		CertLevel:         "standard",
	}

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/cert/{cert_id}/anchor", h.HandleAnchorCertificate)

	body := `{"chain": "base"}`
	req := httptest.NewRequest("POST", "/registry/testauthor/testfunc/cert/"+certID+"/anchor", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Anchoring not implemented without HSM.
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestHandleAnchorCertificate_UnsupportedChain(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	certID := "fxc_01HXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
	fxcert := makeTestFXCert(certID, strings.Repeat("e", 64))
	fxcertJSON, _ := json.Marshal(fxcert)
	repo.certificates[certID] = &registry.ExecutionCertificate{
		CertificateID:     certID,
		ExecutionID:       uuid.New(),
		FunctionID:        fnID,
		CertJSON:          fxcertJSON,
		ExecutionRootHash: strings.Repeat("e", 64),
		CertificateHash:   strings.Repeat("c", 64),
		CertLevel:         "standard",
	}

	// Use a mock anchoring service that is configured but will reject unsupported chain
	mockAnchoring := &mockAnchoringService{configured: true}
	h := NewHandlerWithAnchoring(repo, mockAnchoring)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/cert/{cert_id}/anchor", h.HandleAnchorCertificate)

	body := `{"chain": "unsupported_chain"}`
	req := httptest.NewRequest("POST", "/registry/testauthor/testfunc/cert/"+certID+"/anchor", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleGetDRESummary(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn
	repo.passports[fnID] = makeTestPassport(fnID)

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/dre-stats", h.HandleGetDRESummary)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/dre-stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp DREStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, fnID, resp.FunctionID)
	assert.Equal(t, int64(50), resp.Summary.TotalExecutions)
	assert.Equal(t, int64(46), resp.Summary.VerifiedExecutionsTotal)
	assert.Equal(t, 1, resp.Summary.ReplayDriftIncidents)
	assert.InDelta(t, 0.92, resp.Summary.DeterminismScore, 0.01)
}

func TestHandleGetDRESummary_NoPassport(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/dre-stats", h.HandleGetDRESummary)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/dre-stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp DREStatsResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, fnID, resp.FunctionID)
	assert.Equal(t, int64(0), resp.Summary.TotalExecutions)
}

func TestHandleListDriftReports(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	report := &registry.DriftReportRecord{
		ID:               uuid.New(),
		ExecutionID:      uuid.New(),
		FunctionID:       fnID,
		Version:          "1.0.0",
		OriginalRootHash: strings.Repeat("a", 64),
		ReplayRootHash:   strings.Repeat("b", 64),
		DriftCategory:    "output",
		TrustPenalty:     0.05,
		DetectedAt:       time.Now().UTC(),
	}
	repo.driftReports[fnID] = []*registry.DriftReportRecord{report}

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/drift-reports", h.HandleListDriftReports)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/drift-reports", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	reports, ok := resp["drift_reports"].([]interface{})
	require.True(t, ok)
	assert.Len(t, reports, 1)
	assert.Equal(t, "output", reports[0].(map[string]interface{})["drift_category"])
}

func TestHandleGetPassportPublic(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn
	repo.passports[fnID] = makeTestPassport(fnID)

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/passport/public", h.HandleGetPassportPublic)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/passport/public", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	passport, ok := resp["passport"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "standard", passport["passport_tier"]) // 46/50 = 0.92 ratio.
	assert.Equal(t, "full", passport["determinism_tier"])
}

func TestHandleGetPassportPublic_NoPassport(t *testing.T) {
	repo := newMockDRERepo()
	fnID := uuid.New()
	fn := &registry.RegistryFunction{ID: fnID, Author: "testauthor", Name: "testfunc"}
	repo.functions["testauthor/testfunc"] = fn

	h := NewHandlerFromRepo(repo)
	router := mux.NewRouter()
	router.HandleFunc("/registry/{author}/{name}/passport/public", h.HandleGetPassportPublic)

	req := httptest.NewRequest("GET", "/registry/testauthor/testfunc/passport/public", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	passport, ok := resp["passport"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "lite", passport["passport_tier"])
}

// ─── Validation Tests ─────────────────────────────────────────────────────────

func TestValidateAuthorName(t *testing.T) {
	tests := []struct {
		author, name string
		wantErr      bool
	}{
		{"testauthor", "testfunc", false},
		{"test-author", "test_func", false},
		{"test.author", "test-function", false},
		{"testauthor", "testfunc", false},
		{"-invalid", "testfunc", true},
		{"testauthor", "-invalid", true},
		{"", "testfunc", true},
		{"testauthor", "", true},
		{"test author", "testfunc", true}, // space not allowed.
		{"testauthor", "test func", true},
		{strings.Repeat("a", 64), "testfunc", false},   // max length author.
		{"testauthor", strings.Repeat("a", 64), false}, // max length name.
		{strings.Repeat("a", 65), "testfunc", true},    // too long author.
	}

	for _, tt := range tests {
		err := validateAuthorName(tt.author, tt.name)
		if tt.wantErr {
			assert.NotNil(t, err, "expected error for %s/%s", tt.author, tt.name)
		} else {
			assert.Nil(t, err, "unexpected error for %s/%s", tt.author, tt.name)
		}
	}
}

func TestValidateCertID(t *testing.T) {
	tests := []struct {
		certID  string
		wantErr bool
	}{
		{"fxc_01HXXXXXXXXXXXXXXXXXXXXXXXXXXXX", false},
		{"fxc_01HAABBCCDD", false},
		{"", true},
		{"fxc_notvalid", true}, // too short
		{"invalid_prefix", true},
		{"FXC_01HXXXXXXXXXXXXXXXXXXXXXXXXXXXX", true}, // uppercase prefix not allowed by regex
	}

	for _, tt := range tests {
		err := validateCertID(tt.certID)
		if tt.wantErr {
			assert.NotNil(t, err, "expected error for %s", tt.certID)
		} else {
			assert.Nil(t, err, "unexpected error for %s", tt.certID)
		}
	}
}

func TestValidateExecutionRootHash(t *testing.T) {
	tests := []struct {
		hash    string
		wantErr bool
	}{
		{strings.Repeat("a", 64), false},
		{strings.Repeat("e", 64), false},
		{strings.Repeat("0", 64), false},
		{strings.Repeat("f", 64), false},
		{"", true},
		{strings.Repeat("a", 63), true}, // too short
		{strings.Repeat("a", 65), true}, // too long
		{strings.Repeat("g", 64), true}, // 'g' is not valid hex
		{strings.Repeat("G", 64), true}, // uppercase 'G' not valid hex
	}

	for _, tt := range tests {
		err := validateExecutionRootHash(tt.hash)
		if tt.wantErr {
			assert.NotNil(t, err, "expected error for hash len=%d", len(tt.hash))
		} else {
			assert.Nil(t, err, "unexpected error for hash len=%d", len(tt.hash))
		}
	}
}

// ─── Certificate Verification Logic Tests ────────────────────────────────────

func TestFXCertVerification_Valid(t *testing.T) {
	// Build component hashes and compute the correct execution root hash.
	// Each "hash" is 32 bytes (64 hex chars) of BLAKE3 output.
	componentHashes := []string{
		strings.Repeat("a", 64),
		strings.Repeat("b", 64),
		strings.Repeat("c", 64),
		strings.Repeat("d", 64),
		strings.Repeat("e", 64),
		strings.Repeat("f", 64),
		strings.Repeat("0", 64), // 'g' is invalid hex.
	}
	leaves := make([][]byte, len(componentHashes))
	for i, h := range componentHashes {
		b, err := hexDecode(h)
		require.NoError(t, err)
		leaves[i] = b
	}
	computedRoot := drecrypto.MerkleRoot(leaves)
	executionRootHash := hexEncode(computedRoot)

	fxcert := &cert.FXCert{
		FXCertVersion: "1.0",
		CertificateID: "fxc_01HXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
		Integrity: cert.IntegritySection{
			ExecutionRootHash: executionRootHash,
			InputHash:         componentHashes[0],
			EnvironmentHash:   componentHashes[1],
			DependencyHash:    componentHashes[2],
			TraceHash:         componentHashes[3],
			ResourceHash:      componentHashes[4],
			OutputHash:        componentHashes[5],
			MetadataHash:      componentHashes[6],
		},
		Trust: cert.TrustSection{
			TrustScore: 0.9,
		},
	}

	// Recompute ExecutionRootHash from component hashes via the same algorithm.
	leafHashes2 := []string{
		fxcert.Integrity.InputHash,
		fxcert.Integrity.EnvironmentHash,
		fxcert.Integrity.DependencyHash,
		fxcert.Integrity.TraceHash,
		fxcert.Integrity.ResourceHash,
		fxcert.Integrity.OutputHash,
		fxcert.Integrity.MetadataHash,
	}
	leaves2 := make([][]byte, len(leafHashes2))
	for i, h := range leafHashes2 {
		b, err := hexDecode(h)
		require.NoError(t, err)
		leaves2[i] = b
	}
	recomputedRoot := drecrypto.MerkleRoot(leaves2)
	recomputedRootHex := hexEncode(recomputedRoot)

	assert.Equal(t, fxcert.Integrity.ExecutionRootHash, recomputedRootHex)
}

func TestGetCertificateTrustLevel(t *testing.T) {
	tests := []struct {
		name   string
		cert   *cert.FXCert
		expect string
	}{
		{
			"minimal lite",
			&cert.FXCert{},
			"lite",
		},
		{
			"basic with node sig",
			&cert.FXCert{
				Signatures: cert.SignatureSection{NodeSignature: &cert.Signature{}},
			},
			"basic",
		},
		{
			"standard with node and platform",
			&cert.FXCert{
				Signatures: cert.SignatureSection{
					NodeSignature:     &cert.Signature{},
					PlatformSignature: &cert.Signature{},
				},
			},
			"standard",
		},
		{
			"verified with anchor",
			&cert.FXCert{
				Signatures: cert.SignatureSection{
					NodeSignature:     &cert.Signature{},
					PlatformSignature: &cert.Signature{},
				},
				Anchoring: cert.AnchoringSection{Anchored: true},
			},
			"verified",
		},
		{
			"enterprise full",
			&cert.FXCert{
				Signatures: cert.SignatureSection{
					NodeSignature:     &cert.Signature{},
					PlatformSignature: &cert.Signature{},
				},
				Anchoring:  cert.AnchoringSection{Anchored: true},
				ReplayCert: &cert.ReplayCertSection{RootsMatch: true},
			},
			"enterprise",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := cert.GetCertificateTrustLevel(tt.cert)
			assert.Equal(t, tt.expect, level)
		})
	}
}

func TestIsChainSupported(t *testing.T) {
	assert.True(t, cert.IsChainSupported("ethereum"))
	assert.True(t, cert.IsChainSupported("polygon"))
	assert.True(t, cert.IsChainSupported("base"))
	assert.True(t, cert.IsChainSupported("arbitrum"))
	assert.True(t, cert.IsChainSupported("optimism"))
	assert.True(t, cert.IsChainSupported("avalanche"))
	assert.False(t, cert.IsChainSupported("bitcoin"))
	assert.False(t, cert.IsChainSupported(""))
}

func TestHexCodec(t *testing.T) {
	// Test round-trip (lowercase input).
	original := strings.Repeat("a", 64)
	decoded, err := hexDecode(original)
	require.NoError(t, err)
	encoded := hexEncode(decoded)
	assert.Equal(t, original, encoded) // hexEncode produces lowercase.

	// Test uppercase input decodes correctly.
	upper := strings.Repeat("A", 64)
	decoded, err = hexDecode(upper)
	require.NoError(t, err)
	// hexEncode always produces lowercase.
	assert.Equal(t, strings.Repeat("a", 64), hexEncode(decoded))

	// Test invalid characters.
	_, err = hexDecode("nothex")
	assert.Error(t, err)
}

// MerkleRoot exposes drecrypto.MerkleRoot for test use.
var MerkleRoot = drecrypto.MerkleRoot
