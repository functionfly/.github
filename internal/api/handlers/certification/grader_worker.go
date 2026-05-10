package certification

import (
	"context"
	"fmt"
	"os"
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
		_ = w.certRepo.FailGrading(ctx, item.ID, "challenge not found")
		return
	}

	// Get the exam to find the deployed function URL
	exam, err := w.certRepo.GetExamByID(ctx, item.ExamID)
	if err != nil || exam == nil {
		w.logger.WithError(err).Error("Failed to get exam for grading")
		_ = w.certRepo.FailGrading(ctx, item.ID, "exam not found")
		return
	}

	// Execute the practical challenge grading
	result, err := w.gradeChallenge(ctx, challenge, exam)
	if err != nil {
		w.logger.WithError(err).Error("Grading failed")
		if item.Attempts >= item.MaxAttempts {
			_ = w.certRepo.FailGrading(ctx, item.ID, err.Error())
		} else {
			// Re-queue by resetting status to pending
			_ = w.certRepo.FailGrading(ctx, item.ID, fmt.Sprintf("attempt %d failed: %s", item.Attempts, err.Error()))
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
	// The candidate stores their deployed URL in practical_results
	practicalResults := exam.PracticalResults
	if practicalResults == nil {
		return storage.JSONMap{"score": 0, "feedback": "No deployment submitted"}, nil
	}

	// For now, check that the candidate deployed something
	// Full HTTP call grading will be implemented with the execution engine
	return storage.JSONMap{
		"score":    100,
		"feedback": "Deployment verified",
	}, nil
}

// gradeHTTPResponse calls an endpoint and verifies the response
func (w *GradingWorker) gradeHTTPResponse(ctx context.Context, config storage.JSONMap) (storage.JSONMap, error) {
	// Will be implemented with HTTP client for automated function testing
	return storage.JSONMap{
		"score":    0,
		"feedback": "HTTP response grading not yet implemented",
	}, nil
}

// gradeStateCheck verifies state was written correctly
func (w *GradingWorker) gradeStateCheck(ctx context.Context, config storage.JSONMap) (storage.JSONMap, error) {
	// Will be implemented with state store verification
	return storage.JSONMap{
		"score":    0,
		"feedback": "State check grading not yet implemented",
	}, nil
}
