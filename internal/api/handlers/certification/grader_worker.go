package certification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

// GradingWorker processes practical challenge grading tasks from the queue
// Uses SELECT ... FOR UPDATE SKIP LOCKED for safe concurrent dequeue
type GradingWorker struct {
	certRepo     *storage.CertificationRepository
	workerID     string
	pollInterval time.Duration
	logger       *logrus.Entry
	httpClient   *http.Client
}

// NewGradingWorker creates a new grading worker
func NewGradingWorker(certRepo *storage.CertificationRepository) *GradingWorker {
	hostname, _ := os.Hostname()
	workerID := fmt.Sprintf("grader-%s-%d", hostname, time.Now().UnixNano())

	return &GradingWorker{
		certRepo:     certRepo,
		workerID:     workerID,
		pollInterval: 10 * time.Second,
		logger:       logrus.WithField("component", "cert_grading_worker").WithField("worker_id", workerID),
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// Start begins the grading worker loop. Blocks until context is cancelled.
func (w *GradingWorker) Start(ctx context.Context) {
	w.logger.Info("Cert grading worker started")
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Cert grading worker stopped")
			return
		case <-ticker.C:
			w.processNext(ctx)
		}
	}
}

func (w *GradingWorker) processNext(ctx context.Context) {
	item, err := w.certRepo.DequeueGrading(ctx, w.workerID)
	if err != nil {
		w.logger.WithError(err).Error("Failed to dequeue grading task")
		return
	}
	if item == nil {
		return // nothing to process
	}

	w.logger.WithFields(logrus.Fields{
		"queue_id":    item.ID,
		"exam_id":     item.ExamID,
		"challenge_id": item.ChallengeID,
	}).Info("Processing grading task")

	// Get the challenge config for grading
	challenge, err := w.certRepo.GetChallengeByID(ctx, item.ChallengeID)
	if err != nil || challenge == nil {
		w.logger.WithError(err).Error("Failed to get challenge for grading")
		if failErr := w.certRepo.FailGrading(ctx, item.ID, "challenge not found"); failErr != nil {
			w.logger.WithError(failErr).Error("Failed to mark grading as failed")
		}
		return
	}

	// Get the exam to find the deployed function URL
	exam, err := w.certRepo.GetExamByID(ctx, item.ExamID)
	if err != nil || exam == nil {
		w.logger.WithError(err).Error("Failed to get exam for grading")
		if failErr := w.certRepo.FailGrading(ctx, item.ID, "exam not found"); failErr != nil {
			w.logger.WithError(failErr).Error("Failed to mark grading as failed")
		}
		return
	}

	// Execute the practical challenge grading
	result, err := w.gradeChallenge(ctx, challenge, exam)
	if err != nil {
		w.logger.WithError(err).Error("Grading failed")
		failMsg := err.Error()
		if item.Attempts >= item.MaxAttempts {
			if failErr := w.certRepo.FailGrading(ctx, item.ID, failMsg); failErr != nil {
				w.logger.WithError(failErr).Error("Failed to mark grading as permanently failed")
			}
		} else {
			// Re-queue by resetting status to pending
			if failErr := w.certRepo.FailGrading(ctx, item.ID, fmt.Sprintf("attempt %d failed: %s", item.Attempts, failMsg)); failErr != nil {
				w.logger.WithError(failErr).Error("Failed to mark grading as failed")
			}
		}
		return
	}

	if err := w.certRepo.CompleteGrading(ctx, item.ID, result); err != nil {
		w.logger.WithError(err).Error("Failed to mark grading complete")
		return
	}

	w.logger.WithFields(logrus.Fields{
		"queue_id": item.ID,
		"score":    result["score"],
	}).Info("Grading task completed")
}

// gradeChallenge executes the grading logic for a practical challenge
func (w *GradingWorker) gradeChallenge(ctx context.Context, challenge *storage.CertPracticalChallenge, exam *storage.CertExam) (storage.JSONMap, error) {
	gradingConfig := challenge.GradingConfig

	// Extract grading type from config
	gradeType, _ := gradingConfig["type"].(string)

	switch gradeType {
	case "deployment_check":
		return w.gradeDeploymentCheck(ctx, gradingConfig, exam)
	case "http_response":
		return w.gradeHTTPResponse(ctx, gradingConfig)
	case "state_check":
		return w.gradeStateCheck(ctx, gradingConfig)
	default:
		// For now, mark as scored with a default if type is unknown
		return storage.JSONMap{
			"score":    0,
			"feedback": fmt.Sprintf("Unknown grading type: %s", gradeType),
		}, nil
	}
}

// gradeDeploymentCheck verifies a function was deployed and responds correctly
func (w *GradingWorker) gradeDeploymentCheck(ctx context.Context, config storage.JSONMap, exam *storage.CertExam) (storage.JSONMap, error) {
	practicalResults := exam.PracticalResults
	if practicalResults == nil || len(practicalResults) == 0 {
		return storage.JSONMap{"score": 0, "feedback": "No deployment URL submitted"}, nil
	}

	deployedURL, ok := practicalResults["deployed_url"].(string)
	if !ok || deployedURL == "" {
		urls, urlMapOK := practicalResults["urls"].(map[string]interface{})
		if !urlMapOK {
			return storage.JSONMap{"score": 0, "feedback": "Missing 'deployed_url' or 'urls' in practical results"}, nil
		}
		primaryURL, primaryOK := urls["primary"].(string)
		if !primaryOK || primaryURL == "" {
			return storage.JSONMap{"score": 0, "feedback": "Missing 'urls.primary' in practical results"}, nil
		}
		deployedURL = primaryURL
	}

	checkURL := strings.TrimRight(deployedURL, "/")

	expectedPath, _ := config["expected_path"].(string)
	if expectedPath != "" {
		checkURL = strings.TrimRight(checkURL, "/") + "/" + strings.TrimLeft(expectedPath, "/")
	}

	expectedStatus := 200
	if sv, ok := config["expected_status"].(float64); ok {
		expectedStatus = int(sv)
	}

	if expectedStatus == 0 {
		expectedStatus = 200
	}

	req, err := http.NewRequestWithContext(ctx, "GET", checkURL, nil)
	if err != nil {
		return storage.JSONMap{"score": 0, "feedback": fmt.Sprintf("Invalid URL %q: %v", checkURL, err)}, nil
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return storage.JSONMap{"score": 0, "feedback": fmt.Sprintf("Deployment at %q is not reachable: %v", checkURL, err)}, nil
	}
	defer resp.Body.Close()

	score := 0
	if resp.StatusCode == expectedStatus {
		score = 100
	} else if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		score = 50
	}

	if expectedBody, ok := config["expected_body"].(string); ok && expectedBody != "" {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return storage.JSONMap{"score": 0, "feedback": fmt.Sprintf("Failed to read response body from %s: %v", checkURL, err)}, nil
		}
		if !strings.Contains(string(bodyBytes), expectedBody) {
			return storage.JSONMap{
				"score":    max(0, score-30),
				"feedback": fmt.Sprintf("Deployment at %q does not contain expected body content", checkURL),
			}, nil
		}
	}

	feedback := fmt.Sprintf("Deployment verified at %s (HTTP %d)", checkURL, resp.StatusCode)
	if score == 100 {
		feedback = "Deployment fully verified"
	}

	return storage.JSONMap{"score": score, "feedback": feedback}, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// gradeHTTPResponse calls an endpoint and verifies the response
func (w *GradingWorker) gradeHTTPResponse(ctx context.Context, config storage.JSONMap) (storage.JSONMap, error) {
	if config == nil {
		return storage.JSONMap{"score": 0, "feedback": "HTTP response grading requires 'config'"}, nil
	}

	url, ok := config["url"].(string)
	if !ok || url == "" {
		return storage.JSONMap{"score": 0, "feedback": "HTTP response grading requires 'url' in config"}, nil
	}

	method := "GET"
	if m, ok := config["method"].(string); ok && m != "" {
		method = strings.ToUpper(m)
	}

	expectedStatus, ok := config["expected_status"].(float64)
	if !ok {
		expectedStatus = 200
	}

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return storage.JSONMap{"score": 0, "feedback": fmt.Sprintf("Failed to create request: %v", err)}, nil
	}

	if headers, ok := config["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if vs, ok := v.(string); ok {
				req.Header.Set(k, vs)
			}
		}
	}

	if body, ok := config["body"].(string); ok && body != "" {
		req.Header.Set("Content-Type", "application/json")
		req.Body = io.NopCloser(strings.NewReader(body))
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return storage.JSONMap{"score": 0, "feedback": fmt.Sprintf("Request failed: %v", err)}, nil
	}
	defer resp.Body.Close()

	if int(resp.StatusCode) != int(expectedStatus) {
		return storage.JSONMap{
			"score":    0,
			"feedback": fmt.Sprintf("Expected status %d, got %d", int(expectedStatus), resp.StatusCode),
		}, nil
	}

	if expectedBody, ok := config["expected_body"].(string); ok && expectedBody != "" {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return storage.JSONMap{"score": 0, "feedback": fmt.Sprintf("Failed to read response body: %v", err)}, nil
		}
		bodyStr := string(bodyBytes)
		if !strings.Contains(bodyStr, expectedBody) {
			return storage.JSONMap{
				"score":    0,
				"feedback": fmt.Sprintf("Response body does not contain expected text: %s", expectedBody),
			}, nil
		}
	}

	if expectedJSON, ok := config["expected_json"].(map[string]interface{}); ok {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return storage.JSONMap{"score": 0, "feedback": fmt.Sprintf("Failed to read response body: %v", err)}, nil
		}
		var responseJSON map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &responseJSON); err != nil {
			return storage.JSONMap{"score": 0, "feedback": fmt.Sprintf("Response is not valid JSON: %v", err)}, nil
		}
		for k, expected := range expectedJSON {
			actual, ok := responseJSON[k]
			if !ok {
				return storage.JSONMap{
					"score":    0,
					"feedback": fmt.Sprintf("Expected JSON key '%s' not found in response", k),
				}, nil
			}
			if actual != expected {
				return storage.JSONMap{
					"score":    0,
					"feedback": fmt.Sprintf("JSON key '%s': expected %v, got %v", k, expected, actual),
				}, nil
			}
		}
	}

	return storage.JSONMap{
		"score":    100,
		"feedback": fmt.Sprintf("HTTP response verified (status %d)", resp.StatusCode),
	}, nil
}

// gradeStateCheck verifies state was written correctly
func (w *GradingWorker) gradeStateCheck(ctx context.Context, config storage.JSONMap) (storage.JSONMap, error) {
	stateChecksRaw, ok := config["state_checks"].([]interface{})
	if !ok || len(stateChecksRaw) == 0 {
		return storage.JSONMap{"score": 0, "feedback": "State check grading requires 'state_checks' in config"}, nil
	}

	stateAPIURL, _ := config["state_api_url"].(string)
	if stateAPIURL == "" {
		stateAPIURL = os.Getenv("STATE_API_URL")
		if stateAPIURL == "" {
			return storage.JSONMap{"score": 0, "feedback": "State check grading requires 'state_api_url' in config or STATE_API_URL env var"}, nil
		}
	}

	for _, checkRaw := range stateChecksRaw {
		check, ok := checkRaw.(map[string]interface{})
		if !ok {
			continue
		}

		store, _ := check["store"].(string)
		key, _ := check["key"].(string)
		operator, _ := check["operator"].(string)
		expectedValue := check["value"]

		if store == "" || key == "" {
			return storage.JSONMap{"score": 0, "feedback": "State check requires 'store' and 'key'"}, nil
		}

		stateURL := fmt.Sprintf("%s/%s/%s", strings.TrimSuffix(stateAPIURL, "/"), store, key)
		req, err := http.NewRequestWithContext(ctx, "GET", stateURL, nil)
		if err != nil {
			return storage.JSONMap{"score": 0, "feedback": fmt.Sprintf("Failed to create state request: %v", err)}, nil
		}

		resp, err := w.httpClient.Do(req)
		if err != nil {
			return storage.JSONMap{"score": 0, "feedback": fmt.Sprintf("State request failed: %v", err)}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return storage.JSONMap{
				"score":    0,
				"feedback": fmt.Sprintf("State key '%s' returned status %d", key, resp.StatusCode),
			}, nil
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return storage.JSONMap{"score": 0, "feedback": fmt.Sprintf("Failed to read state response: %v", err)}, nil
		}

		switch operator {
		case "eq":
			if !jsonEqual(string(bodyBytes), expectedValue) {
				return storage.JSONMap{
					"score":    0,
					"feedback": fmt.Sprintf("State key '%s': value mismatch (expected %v)", key, expectedValue),
				}, nil
			}
		case "exists":
		case "contains":
			if !strings.Contains(string(bodyBytes), fmt.Sprintf("%v", expectedValue)) {
				return storage.JSONMap{
					"score":    0,
					"feedback": fmt.Sprintf("State key '%s' does not contain '%v'", key, expectedValue),
				}, nil
			}
		default:
		}
	}

	return storage.JSONMap{
		"score":    100,
		"feedback": fmt.Sprintf("State check verified (%d rules)", len(stateChecksRaw)),
	}, nil
}

func jsonEqual(a string, b interface{}) bool {
	var aVal, bVal interface{}
	if err := json.Unmarshal([]byte(a), &aVal); err != nil {
		aVal = a
	}
	if mb, ok := b.(string); ok {
		if err := json.Unmarshal([]byte(mb), &bVal); err != nil {
			bVal = mb
		}
	} else {
		bVal = b
	}
	return fmt.Sprintf("%v", aVal) == fmt.Sprintf("%v", bVal)
}
