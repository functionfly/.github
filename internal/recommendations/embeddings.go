package recommendations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// flyEmbedTripleEmbeddingRequest is the request body for the AI service triple embed endpoint.
type flyEmbedTripleEmbeddingRequest struct {
	FunctionID   string                 `json:"function_id"`
	Name         string                 `json:"name"`
	Title        string                 `json:"title,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Category     string                 `json:"category,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Manifest     map[string]interface{} `json:"manifest,omitempty"`
	SourceCode   string                 `json:"source_code,omitempty"`
	Runtime      string                 `json:"runtime,omitempty"`
	Capabilities []string               `json:"capabilities,omitempty"`
}

// flyEmbedTripleEmbeddingResponse is the response from the AI service triple embed endpoint.
type flyEmbedTripleEmbeddingResponse struct {
	FunctionID        string    `json:"function_id"`
	ContractEmbedding []float64 `json:"contract_embedding"`
	SemanticEmbedding []float64 `json:"semantic_embedding"`
	CodeEmbedding     []float64 `json:"code_embedding"`
	ContractText      string    `json:"contract_text"`
	SemanticText      string    `json:"semantic_text"`
	CodeText          string    `json:"code_text"`
	EmbeddingModel    string    `json:"embedding_model"`
	LatencyMs         float64   `json:"latency_ms"`
}

// flyEmbedTripleQueryRequest is the request body for the AI service triple query endpoint.
type flyEmbedTripleQueryRequest struct {
	Query string `json:"query"`
}

// flyEmbedTripleQueryResponse is the response from the AI service triple query endpoint.
type flyEmbedTripleQueryResponse struct {
	Query          string    `json:"query"`
	ContractVector []float64 `json:"contract_vector"`
	SemanticVector []float64 `json:"semantic_vector"`
	CodeVector     []float64 `json:"code_vector"`
	Dimensions     int       `json:"dimensions"`
	LatencyMs      float64   `json:"latency_ms"`
}

// tripleQueryVectors holds the three query vectors from the AI service.
type tripleQueryVectors struct {
	contract []float32
	semantic []float32
	code     []float32
}

// UpsertFunctionEmbedding stores or updates the vector embedding for a function (pgvector vector(1536)).
func (s *Service) UpsertFunctionEmbedding(ctx context.Context, functionID uuid.UUID, embedding []float32, embeddedText *string, model string) error {
	if model == "" {
		model = "text-embedding-ada-002"
	}
	return s.repo.UpsertFunctionEmbedding(ctx, &FunctionEmbedding{
		FunctionID:     functionID,
		Embedding:      embedding,
		EmbeddedText:   embeddedText,
		EmbeddingModel: model,
	})
}

// SearchSimilarByEmbedding returns recommendations by vector similarity (cosine). Score = 1 - cosine_distance.
// excludeFunctionID is optional (e.g. the query function). Requires registry to resolve function details.
func (s *Service) SearchSimilarByEmbedding(ctx context.Context, queryEmbedding []float32, limit int, excludeFunctionID *uuid.UUID) ([]RecommendationResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Get tenant ID from context if available for tenant isolation
	tenantID, _ := ctx.Value("tenant_id").(string)

	var withDist []*FunctionEmbeddingWithDistance
	var err error

	if tenantID != "" {
		tID, parseErr := uuid.Parse(tenantID)
		if parseErr == nil {
			withDist, err = s.repo.SearchFunctionEmbeddingsByVectorForTenant(ctx, tID, queryEmbedding, limit, excludeFunctionID)
		} else {
			withDist, err = s.repo.SearchFunctionEmbeddingsByVector(ctx, queryEmbedding, limit, excludeFunctionID, nil)
		}
	} else {
		withDist, err = s.repo.SearchFunctionEmbeddingsByVector(ctx, queryEmbedding, limit, excludeFunctionID, nil)
	}

	if err != nil || len(withDist) == 0 {
		return nil, err
	}
	if s.registry == nil {
		return nil, nil
	}
	var results []RecommendationResult
	for _, e := range withDist {
		fn, err := s.registry.GetFunctionByID(e.FunctionID)
		if err != nil {
			continue
		}
		// Cosine distance is in [0, 2]; similarity = 1 - distance, clamped to [0, 1].
		score := 1 - e.Distance
		if score < 0 {
			score = 0
		}
		if score > 1 {
			score = 1
		}
		results = append(results, RecommendationResult{
			FunctionID:           fn.ID,
			Author:               fn.Author,
			Name:                 fn.Name,
			Title:                fn.Title.String,
			Description:          fn.Description.String,
			Category:             fn.Category.String,
			Tags:                 s.parseTags(fn.Tags),
			PopularityScore:      fn.PopularityScore,
			ReliabilityScore:     fn.ReliabilityScore,
			Score:                score,
			RecommendationType:   RecommendationTypeSimilar,
		})
	}
	return results, nil
}

// EmbedFunctionViaAIService calls the AI service to generate triple embeddings and stores them.
// This is called by the backfill CLI and the publish goroutine.
func (s *Service) EmbedFunctionViaAIService(ctx context.Context, functionID uuid.UUID, name, title, description, category string, tags []string, manifest map[string]interface{}, sourceCode, runtime string, capabilities []string) error {
	reqBody := flyEmbedTripleEmbeddingRequest{
		FunctionID:   functionID.String(),
		Name:         name,
		Title:        title,
		Description:  description,
		Category:     category,
		Tags:         tags,
		Manifest:     manifest,
		SourceCode:   sourceCode,
		Runtime:      runtime,
		Capabilities: capabilities,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/flyembed/embed", s.flyEmbedSvcURL)
	req, err := s.makeAuthenticatedRequest(ctx, "POST", url, body)
	if err != nil {
		return err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call ai-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("ai-service authentication failed (401): check API key configuration")
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("ai-service authorization failed (403): insufficient permissions")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ai-service returned status %d", resp.StatusCode)
	}

	var embedResp flyEmbedTripleEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert []float64 to []float32 for pgvector
	contractVec := make([]float32, len(embedResp.ContractEmbedding))
	semanticVec := make([]float32, len(embedResp.SemanticEmbedding))
	codeVec := make([]float32, len(embedResp.CodeEmbedding))
	for i, v := range embedResp.ContractEmbedding {
		contractVec[i] = float32(v)
	}
	for i, v := range embedResp.SemanticEmbedding {
		semanticVec[i] = float32(v)
	}
	for i, v := range embedResp.CodeEmbedding {
		codeVec[i] = float32(v)
	}

	embeddingModel := embedResp.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "flyembed-v1"
	}

	triple := &FunctionEmbeddingTriple{
		FunctionID:        functionID,
		ContractEmbedding: contractVec,
		SemanticEmbedding: semanticVec,
		CodeEmbedding:     codeVec,
		ContractText:      &embedResp.ContractText,
		SemanticText:      &embedResp.SemanticText,
		CodeText:          &embedResp.CodeText,
		EmbeddingModel:    embeddingModel,
		EmbeddingVersion:  1,
	}

	return s.repo.UpsertFunctionEmbeddingTriple(ctx, triple)
}

// UpsertTripleEmbedding stores a triple embedding for a function (called directly, not via AI service).
func (s *Service) UpsertTripleEmbedding(ctx context.Context, functionID uuid.UUID, contract, semantic, code []float32, contractText, semanticText, codeText *string, model string) error {
	if model == "" {
		model = "flyembed-v1"
	}
	return s.repo.UpsertFunctionEmbeddingTriple(ctx, &FunctionEmbeddingTriple{
		FunctionID:        functionID,
		ContractEmbedding: contract,
		SemanticEmbedding: semantic,
		CodeEmbedding:     code,
		ContractText:      contractText,
		SemanticText:      semanticText,
		CodeText:          codeText,
		EmbeddingModel:    model,
		EmbeddingVersion:  1,
	})
}

// SearchSimilarByTripleEmbedding uses triple vectors when available, falls back to single vector.
func (s *Service) SearchSimilarByTripleEmbedding(ctx context.Context, query string, limit int, excludeFunctionID *uuid.UUID) ([]RecommendationResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Check if triples exist
	tripleCount, err := s.repo.CountTripleEmbeddings(ctx)
	if err != nil || tripleCount == 0 {
		// Fall back to single-vector search via AI service query
		logrus.Debug("No triple embeddings found, falling back to single-vector search")
		return s.searchBySingleVector(ctx, query, limit, excludeFunctionID)
	}

	// Generate triple query vectors via AI service
	queryVec, err := s.generateTripleQueryVectors(ctx, query)
	if err != nil {
		logrus.WithError(err).Warn("Failed to generate triple query vectors, falling back to single-vector")
		return s.searchBySingleVector(ctx, query, limit, excludeFunctionID)
	}

	// Run triple-vector search
	weights := DefaultTripleSearchWeights()

	// Get tenant ID from context if available for tenant isolation
	tenantID, _ := ctx.Value("tenant_id").(string)
	var tenantUUID *uuid.UUID
	if tenantID != "" {
		if tID, err := uuid.Parse(tenantID); err == nil {
			tenantUUID = &tID
		}
	}

	tripleResults, err := s.repo.SearchByTripleVector(ctx, queryVec.contract, queryVec.semantic, queryVec.code, weights, limit, excludeFunctionID, tenantUUID)
	if err != nil {
		return nil, fmt.Errorf("triple vector search failed: %w", err)
	}

	// Resolve function details from registry
	return s.tripleResultsToRecommendations(ctx, tripleResults)
}

// SearchByTripleEmbedding generates 3 query embeddings via AI service, runs triple MaxSim.
func (s *Service) SearchByTripleEmbedding(ctx context.Context, query string, weights *TripleSearchWeights, limit int) ([]RecommendationResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if weights == nil {
		w := DefaultTripleSearchWeights()
		weights = &w
	}

	queryVecs, err := s.generateTripleQueryVectors(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query vectors: %w", err)
	}

	// Get tenant ID from context if available for tenant isolation
	tenantID, _ := ctx.Value("tenant_id").(string)
	var tenantUUID *uuid.UUID
	if tenantID != "" {
		if tID, err := uuid.Parse(tenantID); err == nil {
			tenantUUID = &tID
		}
	}

	results, err := s.repo.SearchByTripleVector(ctx, queryVecs.contract, queryVecs.semantic, queryVecs.code, *weights, limit, nil, tenantUUID)
	if err != nil {
		return nil, err
	}

	return s.tripleResultsToRecommendations(ctx, results)
}

// FindComposableFunctions finds functions whose contract inputs match the target function's outputs.
func (s *Service) FindComposableFunctions(ctx context.Context, functionID uuid.UUID, limit int) ([]RecommendationResult, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Get the function's embedding triple
	triple, err := s.repo.GetFunctionEmbeddingTriple(ctx, functionID)
	if err != nil || triple == nil {
		return nil, fmt.Errorf("function has no triple embedding: %w", err)
	}

	// Search by contract vector only (matches output schema to input schema)
	results, err := s.repo.SearchByContractVector(ctx, triple.SemanticEmbedding, limit)
	if err != nil {
		return nil, err
	}

	// Filter out the source function
	filtered := make([]TripleSearchResult, 0, len(results))
	for _, r := range results {
		if r.FunctionID != functionID {
			filtered = append(filtered, r)
		}
		if len(filtered) >= limit {
			break
		}
	}

	return s.tripleResultsToRecommendations(ctx, filtered)
}

// generateTripleQueryVectors generates triple query vectors from the AI service.
func (s *Service) generateTripleQueryVectors(ctx context.Context, query string) (*tripleQueryVectors, error) {
	reqBody := flyEmbedTripleQueryRequest{Query: query}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/flyembed/query", s.flyEmbedSvcURL)
	req, err := s.makeAuthenticatedRequest(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("ai-service authentication failed (401): check API key configuration")
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("ai-service authorization failed (403): insufficient permissions")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ai-service returned status %d", resp.StatusCode)
	}

	var queryResp flyEmbedTripleQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, err
	}

	contract := make([]float32, len(queryResp.ContractVector))
	semantic := make([]float32, len(queryResp.SemanticVector))
	code := make([]float32, len(queryResp.CodeVector))
	for i, v := range queryResp.ContractVector {
		contract[i] = float32(v)
	}
	for i, v := range queryResp.SemanticVector {
		semantic[i] = float32(v)
	}
	for i, v := range queryResp.CodeVector {
		code[i] = float32(v)
	}

	return &tripleQueryVectors{contract: contract, semantic: semantic, code: code}, nil
}

// tripleResultsToRecommendations converts triple search results to recommendations.
func (s *Service) tripleResultsToRecommendations(ctx context.Context, results []TripleSearchResult) ([]RecommendationResult, error) {
	if s.registry == nil {
		return nil, nil
	}
	var recs []RecommendationResult
	for _, r := range results {
		fn, err := s.registry.GetFunctionByID(r.FunctionID)
		if err != nil {
			continue
		}
		recs = append(recs, RecommendationResult{
			FunctionID:           fn.ID,
			Author:               fn.Author,
			Name:                 fn.Name,
			Title:                fn.Title.String,
			Description:          fn.Description.String,
			Category:             fn.Category.String,
			Tags:                 s.parseTags(fn.Tags),
			PopularityScore:      fn.PopularityScore,
			ReliabilityScore:     fn.ReliabilityScore,
			Score:                r.CombinedScore,
			RecommendationType:   RecommendationTypeSimilar,
		})
	}
	return recs, nil
}

// searchBySingleVector falls back to single-vector search via the AI service embed endpoint.
func (s *Service) searchBySingleVector(ctx context.Context, query string, limit int, excludeFunctionID *uuid.UUID) ([]RecommendationResult, error) {
	// Generate a single embedding for the query
	embedReq := map[string]interface{}{"text": query}
	body, err := json.Marshal(embedReq)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/embed", s.flyEmbedSvcURL)
	req, err := s.makeAuthenticatedRequest(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("ai-service authentication failed (401): check API key configuration")
	}
	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("ai-service authorization failed (403): insufficient permissions")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embed endpoint returned status %d", resp.StatusCode)
	}

	var embedResp struct {
		Embedding []float64 `json:"embedding"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, err
	}

	vec := make([]float32, len(embedResp.Embedding))
	for i, v := range embedResp.Embedding {
		vec[i] = float32(v)
	}

	return s.SearchSimilarByEmbedding(ctx, vec, limit, excludeFunctionID)
}
