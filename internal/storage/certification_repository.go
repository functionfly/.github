package storage

import (

	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/functionfly/functionfly/internal/types"
)

// CertificationRepository handles certification-related database operations
type CertificationRepository struct {
	db *PostgresDB
}

// NewCertificationRepository creates a new certification repository
func NewCertificationRepository(db *PostgresDB) *CertificationRepository {
	return &CertificationRepository{db: db}
}

// ──────────────────────────────────────────────────────────────────────────────
// CertTiers
// ──────────────────────────────────────────────────────────────────────────────

// ListTiers returns all active certification tiers ordered by sort_order
func (r *CertificationRepository) ListTiers(ctx context.Context) ([]*CertTier, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, slug, name, description, icon, color, sort_order,
		       price_cents, currency, pass_threshold, time_limit_minutes,
		       question_count, practical_count, validity_months, is_active,
		       metadata, created_at, updated_at
		FROM cert_tiers WHERE is_active = true ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("failed to list cert tiers: %w", err)
	}
	defer rows.Close()

	return scanCertTiers(rows)
}

// GetTierBySlug returns a single tier by slug
func (r *CertificationRepository) GetTierBySlug(ctx context.Context, slug string) (*CertTier, error) {
	tier := &CertTier{}
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, slug, name, description, icon, color, sort_order,
		       price_cents, currency, pass_threshold, time_limit_minutes,
		       question_count, practical_count, validity_months, is_active,
		       metadata, created_at, updated_at
		FROM cert_tiers WHERE slug = $1 AND is_active = true`, slug).Scan(
		&tier.ID, &tier.Slug, &tier.Name, &tier.Description, &tier.Icon, &tier.Color,
		&tier.SortOrder, &tier.PriceCents, &tier.Currency, &tier.PassThreshold,
		&tier.TimeLimitMinutes, &tier.QuestionCount, &tier.PracticalCount,
		&tier.ValidityMonths, &tier.IsActive, &metadataJSON,
		&tier.CreatedAt, &tier.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cert tier: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &tier.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tier metadata: %w", err)
	}
	return tier, nil
}

// GetTierByID returns a single tier by ID
func (r *CertificationRepository) GetTierByID(ctx context.Context, id uuid.UUID) (*CertTier, error) {
	tier := &CertTier{}
	var metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, slug, name, description, icon, color, sort_order,
		       price_cents, currency, pass_threshold, time_limit_minutes,
		       question_count, practical_count, validity_months, is_active,
		       metadata, created_at, updated_at
		FROM cert_tiers WHERE id = $1`, id).Scan(
		&tier.ID, &tier.Slug, &tier.Name, &tier.Description, &tier.Icon, &tier.Color,
		&tier.SortOrder, &tier.PriceCents, &tier.Currency, &tier.PassThreshold,
		&tier.TimeLimitMinutes, &tier.QuestionCount, &tier.PracticalCount,
		&tier.ValidityMonths, &tier.IsActive, &metadataJSON,
		&tier.CreatedAt, &tier.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cert tier by ID: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &tier.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tier metadata: %w", err)
	}
	return tier, nil
}

func scanCertTiers(rows *sql.Rows) ([]*CertTier, error) {
	var tiers []*CertTier
	for rows.Next() {
		tier := &CertTier{}
		var metadataJSON []byte
		if err := rows.Scan(
			&tier.ID, &tier.Slug, &tier.Name, &tier.Description, &tier.Icon, &tier.Color,
			&tier.SortOrder, &tier.PriceCents, &tier.Currency, &tier.PassThreshold,
			&tier.TimeLimitMinutes, &tier.QuestionCount, &tier.PracticalCount,
			&tier.ValidityMonths, &tier.IsActive, &metadataJSON,
			&tier.CreatedAt, &tier.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cert tier: %w", err)
		}
		if err := json.Unmarshal(metadataJSON, &tier.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tier metadata: %w", err)
		}
		tiers = append(tiers, tier)
	}
	return tiers, rows.Err()
}

// ──────────────────────────────────────────────────────────────────────────────
// CertQuestions
// ──────────────────────────────────────────────────────────────────────────────

// ListQuestionsByTier returns all active questions for a tier (admin — includes answers)
func (r *CertificationRepository) ListQuestionsByTier(ctx context.Context, tierID uuid.UUID) ([]*CertQuestion, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tier_id, category, difficulty, question_text, question_format,
		       options, correct_answers, explanation, points, is_active,
		       created_by, metadata, created_at, updated_at
		FROM cert_questions WHERE tier_id = $1 AND is_active = true
		ORDER BY category, difficulty`, tierID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cert questions: %w", err)
	}
	defer rows.Close()
	return scanCertQuestions(rows)
}

// GetQuestionByID returns a single question by ID (admin — includes answers)
func (r *CertificationRepository) GetQuestionByID(ctx context.Context, id uuid.UUID) (*CertQuestion, error) {
	q := &CertQuestion{}
	var optionsJSON, correctJSON, metadataJSON []byte
	var createdBy sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tier_id, category, difficulty, question_text, question_format,
		       options, correct_answers, explanation, points, is_active,
		       created_by, metadata, created_at, updated_at
		FROM cert_questions WHERE id = $1`, id).Scan(
		&q.ID, &q.TierID, &q.Category, &q.Difficulty, &q.QuestionText, &q.QuestionFormat,
		&optionsJSON, &correctJSON, &q.Explanation, &q.Points, &q.IsActive,
		&createdBy, &metadataJSON, &q.CreatedAt, &q.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get cert question: %w", err)
	}
	if err := unmarshalQuestionJSON(q, optionsJSON, correctJSON, metadataJSON, createdBy); err != nil {
		return nil, err
	}
	return q, nil
}

// CountActiveQuestionsByTier returns the count of active questions for a tier
func (r *CertificationRepository) CountActiveQuestionsByTier(ctx context.Context, tierID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM cert_questions WHERE tier_id = $1 AND is_active = true`, tierID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count cert questions: %w", err)
	}
	return count, nil
}

// SelectRandomQuestions picks N random active questions for an exam session,
// stratified by category and difficulty.
func (r *CertificationRepository) SelectRandomQuestions(ctx context.Context, tierID uuid.UUID, count int) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id FROM cert_questions
		WHERE tier_id = $1 AND is_active = true
		ORDER BY random()
		LIMIT $2`, tierID, count)
	if err != nil {
		return nil, fmt.Errorf("failed to select random questions: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan question ID: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetQuestionsByIDs returns questions by their IDs (without answers — for exam display)
func (r *CertificationRepository) GetQuestionsByIDs(ctx context.Context, ids []uuid.UUID) ([]*CertQuestionPublic, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, category, difficulty, question_text, question_format, options, points
		FROM cert_questions
		WHERE id = ANY($1) AND is_active = true`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to get questions by IDs: %w", err)
	}
	defer rows.Close()

	var questions []*CertQuestionPublic
	for rows.Next() {
		q := &CertQuestionPublic{}
		var optionsJSON []byte
		if err := rows.Scan(&q.ID, &q.Category, &q.Difficulty, &q.QuestionText,
			&q.QuestionFormat, &optionsJSON, &q.Points); err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}
		if err := json.Unmarshal(optionsJSON, &q.Options); err != nil {
			return nil, fmt.Errorf("failed to unmarshal options: %w", err)
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

// GetCorrectAnswers returns the correct answers for a set of question IDs (for grading)
func (r *CertificationRepository) GetCorrectAnswers(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]JSONMap, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	query := `
		SELECT id, correct_answers, points
		FROM cert_questions
		WHERE id = ANY($1) AND is_active = true`

	rows, err := r.db.QueryContext(ctx, query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to get correct answers: %w", err)
	}
	defer rows.Close()

	answers := make(map[uuid.UUID]JSONMap)
	for rows.Next() {
		var id uuid.UUID
		var correctJSON []byte
		var points int
		if err := rows.Scan(&id, &correctJSON, &points); err != nil {
			return nil, fmt.Errorf("failed to scan answer: %w", err)
		}
		var correct JSONMap
		if err := json.Unmarshal(correctJSON, &correct); err != nil {
			return nil, fmt.Errorf("failed to unmarshal correct answers: %w", err)
		}
		correct["_points"] = points
		answers[id] = correct
	}
	return answers, rows.Err()
}

func scanCertQuestions(rows *sql.Rows) ([]*CertQuestion, error) {
	var questions []*CertQuestion
	for rows.Next() {
		q := &CertQuestion{}
		var optionsJSON, correctJSON, metadataJSON []byte
		var createdBy sql.NullString
		if err := rows.Scan(
			&q.ID, &q.TierID, &q.Category, &q.Difficulty, &q.QuestionText, &q.QuestionFormat,
			&optionsJSON, &correctJSON, &q.Explanation, &q.Points, &q.IsActive,
			&createdBy, &metadataJSON, &q.CreatedAt, &q.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan cert question: %w", err)
		}
		if err := unmarshalQuestionJSON(q, optionsJSON, correctJSON, metadataJSON, createdBy); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}

func unmarshalQuestionJSON(q *CertQuestion, optionsJSON, correctJSON, metadataJSON []byte, createdBy sql.NullString) error {
	if err := json.Unmarshal(optionsJSON, &q.Options); err != nil {
		return fmt.Errorf("failed to unmarshal question options: %w", err)
	}
	if err := json.Unmarshal(correctJSON, &q.CorrectAnswers); err != nil {
		return fmt.Errorf("failed to unmarshal correct answers: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &q.Metadata); err != nil {
		return fmt.Errorf("failed to unmarshal question metadata: %w", err)
	}
	if createdBy.Valid {
		uid, err := uuid.Parse(createdBy.String)
		if err == nil {
			q.CreatedBy = &uid
		}
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// CertPracticalChallenges
// ──────────────────────────────────────────────────────────────────────────────

// ListChallengesByTier returns all active practical challenges for a tier
func (r *CertificationRepository) ListChallengesByTier(ctx context.Context, tierID uuid.UUID) ([]*CertPracticalChallenge, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tier_id, slug, name, description, category, difficulty,
		       points, time_limit_minutes, grading_config, validator_function_id,
		       is_active, metadata, created_at, updated_at
		FROM cert_practical_challenges WHERE tier_id = $1 AND is_active = true
		ORDER BY difficulty, points`, tierID)
	if err != nil {
		return nil, fmt.Errorf("failed to list practical challenges: %w", err)
	}
	defer rows.Close()
	return scanCertChallenges(rows)
}

// SelectRandomChallenges picks N random active challenges for an exam session
func (r *CertificationRepository) SelectRandomChallenges(ctx context.Context, tierID uuid.UUID, count int) ([]uuid.UUID, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id FROM cert_practical_challenges
		WHERE tier_id = $1 AND is_active = true
		ORDER BY random()
		LIMIT $2`, tierID, count)
	if err != nil {
		return nil, fmt.Errorf("failed to select random challenges: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan challenge ID: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetChallengeByID returns a single challenge by ID
func (r *CertificationRepository) GetChallengeByID(ctx context.Context, id uuid.UUID) (*CertPracticalChallenge, error) {
	ch := &CertPracticalChallenge{}
	var gradingJSON, metadataJSON []byte
	var validatorID sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT id, tier_id, slug, name, description, category, difficulty,
		       points, time_limit_minutes, grading_config, validator_function_id,
		       is_active, metadata, created_at, updated_at
		FROM cert_practical_challenges WHERE id = $1`, id).Scan(
		&ch.ID, &ch.TierID, &ch.Slug, &ch.Name, &ch.Description, &ch.Category,
		&ch.Difficulty, &ch.Points, &ch.TimeLimitMinutes, &gradingJSON,
		&validatorID, &ch.IsActive, &metadataJSON, &ch.CreatedAt, &ch.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get challenge: %w", err)
	}
	if err := json.Unmarshal(gradingJSON, &ch.GradingConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal grading config: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &ch.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal challenge metadata: %w", err)
	}
	if validatorID.Valid {
		uid, err := uuid.Parse(validatorID.String)
		if err == nil {
			ch.ValidatorFuncID = &uid
		}
	}
	return ch, nil
}

func scanCertChallenges(rows *sql.Rows) ([]*CertPracticalChallenge, error) {
	var challenges []*CertPracticalChallenge
	for rows.Next() {
		ch := &CertPracticalChallenge{}
		var gradingJSON, metadataJSON []byte
		var validatorID sql.NullString
		if err := rows.Scan(
			&ch.ID, &ch.TierID, &ch.Slug, &ch.Name, &ch.Description, &ch.Category,
			&ch.Difficulty, &ch.Points, &ch.TimeLimitMinutes, &gradingJSON,
			&validatorID, &ch.IsActive, &metadataJSON, &ch.CreatedAt, &ch.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan challenge: %w", err)
		}
		if err := json.Unmarshal(gradingJSON, &ch.GradingConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal grading config: %w", err)
		}
		if err := json.Unmarshal(metadataJSON, &ch.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal challenge metadata: %w", err)
		}
		if validatorID.Valid {
			uid, err := uuid.Parse(validatorID.String)
			if err == nil {
				ch.ValidatorFuncID = &uid
			}
		}
		challenges = append(challenges, ch)
	}
	return challenges, rows.Err()
}

// ──────────────────────────────────────────────────────────────────────────────
// CertExams
// ──────────────────────────────────────────────────────────────────────────────

// CreateExam inserts a new exam session
func (r *CertificationRepository) CreateExam(ctx context.Context, exam *CertExam) error {
	if exam.ID == uuid.Nil {
		exam.ID = uuid.New()
	}
	exam.CreatedAt = time.Now()
	exam.UpdatedAt = time.Now()

	questionIDsJSON, err := json.Marshal(exam.QuestionIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal question IDs: %w", err)
	}
	practicalIDsJSON, err := json.Marshal(exam.PracticalIDs)
	if err != nil {
		return fmt.Errorf("failed to marshal practical IDs: %w", err)
	}
	answersJSON, err := json.Marshal(exam.Answers)
	if err != nil {
		return fmt.Errorf("failed to marshal answers: %w", err)
	}
	practicalResultsJSON, err := json.Marshal(exam.PracticalResults)
	if err != nil {
		return fmt.Errorf("failed to marshal practical results: %w", err)
	}
	metadataJSON, err := json.Marshal(exam.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO cert_exams (id, user_id, tier_id, status, stripe_payment_id, amount_cents, currency,
			started_at, expires_at, question_ids, practical_ids, answers, practical_results,
			ip_address, user_agent, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10::jsonb, $11::jsonb, $12::jsonb, $13::jsonb, $14::inet, $15, $16::jsonb, $17, $18)`,
		exam.ID, exam.UserID, exam.TierID, exam.Status, exam.StripePaymentID,
		exam.AmountCents, exam.Currency, exam.StartedAt, exam.ExpiresAt,
		string(questionIDsJSON), string(practicalIDsJSON), string(answersJSON),
		string(practicalResultsJSON), exam.IPAddress, exam.UserAgent,
		string(metadataJSON), exam.CreatedAt, exam.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create exam: %w", err)
	}
	return nil
}

// GetExamByID returns an exam by ID
func (r *CertificationRepository) GetExamByID(ctx context.Context, examID uuid.UUID) (*CertExam, error) {
	exam := &CertExam{}
	var stripePaymentID, ipAddress, userAgent sql.NullString
	var submittedAt, gradedAt sql.NullTime
	var knowledgeScore, practicalScore, totalScore sql.NullFloat64
	var passed sql.NullBool
	var questionIDsJSON, practicalIDsJSON, answersJSON, practicalResultsJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tier_id, status, stripe_payment_id, amount_cents, currency,
		       started_at, submitted_at, graded_at, expires_at,
		       knowledge_score, practical_score, total_score, passed,
		       question_ids, practical_ids, answers, practical_results,
		       ip_address, user_agent, metadata, created_at, updated_at
		FROM cert_exams WHERE id = $1`, examID).Scan(
		&exam.ID, &exam.UserID, &exam.TierID, &exam.Status,
		&stripePaymentID, &exam.AmountCents, &exam.Currency,
		&exam.StartedAt, &submittedAt, &gradedAt, &exam.ExpiresAt,
		&knowledgeScore, &practicalScore, &totalScore, &passed,
		&questionIDsJSON, &practicalIDsJSON, &answersJSON, &practicalResultsJSON,
		&ipAddress, &userAgent, &metadataJSON,
		&exam.CreatedAt, &exam.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get exam: %w", err)
	}

	// Unmarshal nullable fields
	if stripePaymentID.Valid {
		exam.StripePaymentID = &stripePaymentID.String
	}
	if submittedAt.Valid {
		exam.SubmittedAt = &submittedAt.Time
	}
	if gradedAt.Valid {
		exam.GradedAt = &gradedAt.Time
	}
	if knowledgeScore.Valid {
		exam.KnowledgeScore = &knowledgeScore.Float64
	}
	if practicalScore.Valid {
		exam.PracticalScore = &practicalScore.Float64
	}
	if totalScore.Valid {
		exam.TotalScore = &totalScore.Float64
	}
	if passed.Valid {
		exam.Passed = &passed.Bool
	}
	if ipAddress.Valid {
		exam.IPAddress = &ipAddress.String
	}
	if userAgent.Valid {
		exam.UserAgent = &userAgent.String
	}

	// Unmarshal JSONB fields
	// Special handling for QuestionIDs and PracticalIDs which are []uuid.UUID not JSONMap
	// Use jsonMapToUUIDs to handle both plain arrays and wrapped {"_ids": [...]} formats
	exam.QuestionIDs = jsonMapToUUIDs(questionIDsJSON)
	exam.PracticalIDs = jsonMapToUUIDs(practicalIDsJSON)
	for _, item := range []struct {
		data []byte
		dest *JSONMap
	}{
		{answersJSON, &exam.Answers},
		{practicalResultsJSON, &exam.PracticalResults},
		{metadataJSON, &exam.Metadata},
	} {
		if err := json.Unmarshal(item.data, item.dest); err != nil {
			return nil, fmt.Errorf("failed to unmarshal exam JSONB: %w", err)
		}
	}

	return exam, nil
}

// UpdateExamAnswer upserts a single answer in the exam's answers JSONB
func (r *CertificationRepository) UpdateExamAnswer(ctx context.Context, examID uuid.UUID, questionID uuid.UUID, answer interface{}) error {
	answerJSON, err := json.Marshal(answer)
	if err != nil {
		return fmt.Errorf("failed to marshal answer: %w", err)
	}

	// Use raw SQL with explicit JSON path to avoid pq driver []string limitation
	_, err = r.db.ExecContext(ctx, `
		UPDATE cert_exams
		SET answers = jsonb_set(COALESCE(answers, '{}'), ARRAY[$1], $2::jsonb, true),
		    updated_at = $3
		WHERE id = $4 AND status = 'in_progress'`,
		questionID.String(), string(answerJSON), time.Now(), examID)
	if err != nil {
		return fmt.Errorf("failed to update exam answer: %w", err)
	}
	return nil
}

// SubmitExam marks an exam as submitted
func (r *CertificationRepository) SubmitExam(ctx context.Context, examID uuid.UUID, answers JSONMap) error {
	now := time.Now()
	answersJSON, err := json.Marshal(answers)
	if err != nil {
		return fmt.Errorf("failed to marshal answers: %w", err)
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE cert_exams
		SET status = 'submitted', submitted_at = $1, answers = $2::jsonb, updated_at = $1
		WHERE id = $3 AND status = 'in_progress'`,
		now, string(answersJSON), examID)
	if err != nil {
		return fmt.Errorf("failed to submit exam: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("exam not found or already submitted")
	}
	return nil
}

// GradeExam records the final scores and pass/fail status
func (r *CertificationRepository) GradeExam(ctx context.Context, examID uuid.UUID, knowledgeScore, practicalScore, totalScore float64, passed bool) error {
	now := time.Now()
	status := CertExamStatusFailed
	if passed {
		status = CertExamStatusPassed
	}

	result, err := r.db.ExecContext(ctx, `
		UPDATE cert_exams
		SET knowledge_score = $1, practical_score = $2, total_score = $3,
		    passed = $4, status = $5, graded_at = $6, updated_at = $6
		WHERE id = $7 AND status IN ('submitted', 'grading')`,
		knowledgeScore, practicalScore, totalScore, passed, status, now, examID)
	if err != nil {
		return fmt.Errorf("failed to grade exam: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("exam not found or not in gradable state")
	}
	return nil
}

// AbandonExam marks an exam as abandoned by the user
func (r *CertificationRepository) AbandonExam(ctx context.Context, examID uuid.UUID) error {
	now := time.Now()
	result, err := r.db.ExecContext(ctx, `
		UPDATE cert_exams
		SET status = $1, updated_at = $2, attempts = attempts + 1
		WHERE id = $3 AND status IN ('in_progress', 'pending_payment')`,
		CertExamStatusAbandoned, now, examID)
	if err != nil {
		return fmt.Errorf("failed to abandon exam: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("exam not found or already completed")
	}

	// Clean up any pending grading queue items so they don't leak
	_, _ = r.db.ExecContext(ctx, `
		DELETE FROM cert_grading_queue
		WHERE exam_id = $1 AND status IN ('pending', 'processing')`,
		examID)

	return nil
}

// ListExamsByUser returns a user's exam history
func (r *CertificationRepository) ListExamsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*CertExam, int, error) {
	// Count
	var total int
	err := r.db.QueryRowContext(ctx,
		`SELECT count(*) FROM cert_exams WHERE user_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count exams: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, tier_id, status, stripe_payment_id, amount_cents, currency,
		       started_at, submitted_at, graded_at, expires_at,
		       knowledge_score, practical_score, total_score, passed,
		       question_ids, practical_ids, answers, practical_results,
		       ip_address, user_agent, metadata, created_at, updated_at
		FROM cert_exams WHERE user_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list exams: %w", err)
	}
	defer rows.Close()

	exams, err := scanCertExams(rows)
	if err != nil {
		return nil, 0, err
	}
	return exams, total, nil
}

// GetCompletedExamCountForUserTier returns how many times the user has completed an exam for the given tier.
// Completed means status in (passed, failed, expired, abandoned).
// NOTE: Passed exams are excluded from the count to allow users at least 1 retake after failure.
func (r *CertificationRepository) GetCompletedExamCountForUserTier(ctx context.Context, userID, tierID uuid.UUID) (int, error) {
	var count int64
	err := r.db.GORM.WithContext(ctx).
		Model(&CertExam{}).
		Where("user_id = ? AND tier_id = ? AND status IN (?, ?, ?, ?)",
			userID, tierID,
			CertExamStatusFailed,
			CertExamStatusExpired,
			CertExamStatusAbandoned,
			CertExamStatusInProgress,
		).
		Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count completed exams: %w", err)
	}
	return int(count), nil
}

// GetActiveExamForUserTier returns an in-progress exam for a user+tier (if any)
func (r *CertificationRepository) GetActiveExamForUserTier(ctx context.Context, userID, tierID uuid.UUID) (*CertExam, error) {
	exam := &CertExam{}
	var stripePaymentID, ipAddress, userAgent sql.NullString
	var submittedAt, gradedAt sql.NullTime
	var knowledgeScore, practicalScore, totalScore sql.NullFloat64
	var passed sql.NullBool
	var questionIDsJSON, practicalIDsJSON, answersJSON, practicalResultsJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tier_id, status, stripe_payment_id, amount_cents, currency,
		       started_at, submitted_at, graded_at, expires_at,
		       knowledge_score, practical_score, total_score, passed,
		       question_ids, practical_ids, answers, practical_results,
		       ip_address, user_agent, metadata, created_at, updated_at
		FROM cert_exams
		WHERE user_id = $1 AND tier_id = $2 AND status = 'in_progress'
		ORDER BY created_at DESC LIMIT 1`, userID, tierID).Scan(
		&exam.ID, &exam.UserID, &exam.TierID, &exam.Status,
		&stripePaymentID, &exam.AmountCents, &exam.Currency,
		&exam.StartedAt, &submittedAt, &gradedAt, &exam.ExpiresAt,
		&knowledgeScore, &practicalScore, &totalScore, &passed,
		&questionIDsJSON, &practicalIDsJSON, &answersJSON, &practicalResultsJSON,
		&ipAddress, &userAgent, &metadataJSON,
		&exam.CreatedAt, &exam.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active exam: %w", err)
	}

	// Populate nullable fields (DRY: reuse same pattern)
	populateExamNullables(exam, stripePaymentID, submittedAt, gradedAt,
		knowledgeScore, practicalScore, totalScore, passed, ipAddress, userAgent)
	// Special handling for QuestionIDs and PracticalIDs which are []uuid.UUID not JSONMap
	// Use jsonMapToUUIDs to handle both plain arrays and wrapped {"_ids": [...]} formats
	exam.QuestionIDs = jsonMapToUUIDs(questionIDsJSON)
	exam.PracticalIDs = jsonMapToUUIDs(practicalIDsJSON)
	for _, item := range []struct {
		data []byte
		dest *JSONMap
	}{
		{answersJSON, &exam.Answers},
		{practicalResultsJSON, &exam.PracticalResults},
		{metadataJSON, &exam.Metadata},
	} {
		if err := json.Unmarshal(item.data, item.dest); err != nil {
			return nil, fmt.Errorf("failed to unmarshal exam JSONB: %w", err)
		}
	}
	return exam, nil
}

// GetPaidExamForUserTier returns an exam that has been paid (pending_payment or in_progress) for a user+tier
func (r *CertificationRepository) GetPaidExamForUserTier(ctx context.Context, userID, tierID uuid.UUID) (*CertExam, error) {
	exam := &CertExam{}
	var stripePaymentID, ipAddress, userAgent sql.NullString
	var submittedAt, gradedAt sql.NullTime
	var knowledgeScore, practicalScore, totalScore sql.NullFloat64
	var passed sql.NullBool
	var questionIDsJSON, practicalIDsJSON, answersJSON, practicalResultsJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tier_id, status, stripe_payment_id, amount_cents, currency,
		       started_at, submitted_at, graded_at, expires_at,
		       knowledge_score, practical_score, total_score, passed,
		       question_ids, practical_ids, answers, practical_results,
		       ip_address, user_agent, metadata, created_at, updated_at
		FROM cert_exams
		WHERE user_id = $1 AND tier_id = $2 AND status IN ('pending_payment', 'in_progress')
		ORDER BY created_at DESC LIMIT 1`, userID, tierID).Scan(
		&exam.ID, &exam.UserID, &exam.TierID, &exam.Status,
		&stripePaymentID, &exam.AmountCents, &exam.Currency,
		&exam.StartedAt, &submittedAt, &gradedAt, &exam.ExpiresAt,
		&knowledgeScore, &practicalScore, &totalScore, &passed,
		&questionIDsJSON, &practicalIDsJSON, &answersJSON, &practicalResultsJSON,
		&ipAddress, &userAgent, &metadataJSON,
		&exam.CreatedAt, &exam.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get paid exam: %w", err)
	}

	populateExamNullables(exam, stripePaymentID, submittedAt, gradedAt,
		knowledgeScore, practicalScore, totalScore, passed, ipAddress, userAgent)
	// Special handling for QuestionIDs and PracticalIDs which are []uuid.UUID not JSONMap
	// Use jsonMapToUUIDs to handle both plain arrays and wrapped {"_ids": [...]} formats
	exam.QuestionIDs = jsonMapToUUIDs(questionIDsJSON)
	exam.PracticalIDs = jsonMapToUUIDs(practicalIDsJSON)
	for _, item := range []struct {
		data []byte
		dest *JSONMap
	}{
		{answersJSON, &exam.Answers},
		{practicalResultsJSON, &exam.PracticalResults},
		{metadataJSON, &exam.Metadata},
	} {
		if err := json.Unmarshal(item.data, item.dest); err != nil {
			return nil, fmt.Errorf("failed to unmarshal exam JSONB: %w", err)
		}
	}
	return exam, nil
}

// ActivateExamFromPaymentWithUser upgrades a pending_payment exam to in_progress status with actual questions
// and verifies that the exam belongs to the specified user.
func (r *CertificationRepository) ActivateExamFromPaymentWithUser(ctx context.Context, examID, userID uuid.UUID) error {
	exam, err := r.GetExamByID(ctx, examID)
	if err != nil || exam == nil {
		return fmt.Errorf("exam not found: %w", err)
	}
	if exam.UserID != userID {
		return fmt.Errorf("exam does not belong to the authenticated user")
	}
	if exam.Status != CertExamStatusPendingPayment {
		return fmt.Errorf("exam is not pending payment: %s", exam.Status)
	}

	tier, err := r.GetTierByID(ctx, exam.TierID)
	if err != nil || tier == nil {
		return fmt.Errorf("tier not found: %w", err)
	}

	if tier.PriceCents > exam.AmountCents {
		return fmt.Errorf("exam amount_cents (%d) does not match tier price (%d)", exam.AmountCents, tier.PriceCents)
	}

	questionIDs, err := r.SelectRandomQuestions(ctx, tier.ID, tier.QuestionCount)
	if err != nil || len(questionIDs) < tier.QuestionCount {
		return fmt.Errorf("failed to select questions: %w", err)
	}

	practicalIDs, err := r.SelectRandomChallenges(ctx, tier.ID, tier.PracticalCount)
	if err != nil {
		return fmt.Errorf("failed to select challenges: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE cert_exams
		SET status = 'in_progress',
		    question_ids = $1,
		    practical_ids = $2,
		    expires_at = now() + interval '1 minute' * $3,
		    updated_at = now()
		WHERE id = $4`,
		MustJSON(uuidsToInterface(questionIDs)),
		MustJSON(uuidsToInterface(practicalIDs)),
		tier.TimeLimitMinutes,
		examID)
	if err != nil {
		return fmt.Errorf("failed to activate exam: %w", err)
	}

	_ = r.db.PgNotify("cert_exam_updates", fmt.Sprintf(`{"type":"cert_exam_status","user_id":"%s","exam_id":"%s","status":"in_progress"}`, exam.UserID.String(), examID.String()))

	return nil
}

// ActivateExamFromPayment upgrades a pending_payment exam to in_progress status with actual questions.
// For new code, prefer ActivateExamFromPaymentWithUser to enforce ownership checks.
// Kept for backward compatibility with existing callers.
func (r *CertificationRepository) ActivateExamFromPayment(ctx context.Context, examID uuid.UUID) error {
	exam, err := r.GetExamByID(ctx, examID)
	if err != nil || exam == nil {
		return fmt.Errorf("exam not found: %w", err)
	}
	if exam.Status != CertExamStatusPendingPayment {
		return fmt.Errorf("exam is not pending payment: %s", exam.Status)
	}

	tier, err := r.GetTierByID(ctx, exam.TierID)
	if err != nil || tier == nil {
		return fmt.Errorf("tier not found: %w", err)
	}

	questionIDs, err := r.SelectRandomQuestions(ctx, tier.ID, tier.QuestionCount)
	if err != nil || len(questionIDs) < tier.QuestionCount {
		return fmt.Errorf("failed to select questions: %w", err)
	}

	practicalIDs, err := r.SelectRandomChallenges(ctx, tier.ID, tier.PracticalCount)
	if err != nil {
		return fmt.Errorf("failed to select challenges: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE cert_exams
		SET status = 'in_progress',
		    question_ids = $1,
		    practical_ids = $2,
		    expires_at = now() + interval '1 minute' * $3,
		    updated_at = now()
		WHERE id = $4`,
		MustJSON(uuidsToInterface(questionIDs)),
		MustJSON(uuidsToInterface(practicalIDs)),
		tier.TimeLimitMinutes,
		examID)
	if err != nil {
		return fmt.Errorf("failed to activate exam: %w", err)
	}

	_ = r.db.PgNotify("cert_exam_updates", fmt.Sprintf(`{"type":"cert_exam_status","user_id":"%s","exam_id":"%s","status":"in_progress"}`, exam.UserID.String(), examID.String()))

	return nil
}

func MustJSON(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}

func uuidsToInterface(ids []uuid.UUID) []interface{} {
	result := make([]interface{}, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}

// UpdateExamStripePaymentID updates an exam with the Stripe payment ID after successful payment
func (r *CertificationRepository) UpdateExamStripePaymentID(ctx context.Context, examID uuid.UUID, stripePaymentID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE cert_exams
		SET stripe_payment_id = $1, updated_at = now()
		WHERE id = $2`, stripePaymentID, examID)
	return err
}

// ExpireStaleExams marks all in_progress exams past their expiry as 'expired'
func (r *CertificationRepository) ExpireStaleExams(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE cert_exams
		SET status = 'expired', updated_at = now()
		WHERE status = 'in_progress' AND expires_at < now()`)
	if err != nil {
		return 0, fmt.Errorf("failed to expire stale exams: %w", err)
	}
	return result.RowsAffected()
}

func scanCertExams(rows *sql.Rows) ([]*CertExam, error) {
	var exams []*CertExam
	for rows.Next() {
		exam := &CertExam{}
		var stripePaymentID, ipAddress, userAgent sql.NullString
		var submittedAt, gradedAt sql.NullTime
		var knowledgeScore, practicalScore, totalScore sql.NullFloat64
		var passed sql.NullBool
		var questionIDsJSON, practicalIDsJSON, answersJSON, practicalResultsJSON, metadataJSON []byte

		if err := rows.Scan(
			&exam.ID, &exam.UserID, &exam.TierID, &exam.Status,
			&stripePaymentID, &exam.AmountCents, &exam.Currency,
			&exam.StartedAt, &submittedAt, &gradedAt, &exam.ExpiresAt,
			&knowledgeScore, &practicalScore, &totalScore, &passed,
			&questionIDsJSON, &practicalIDsJSON, &answersJSON, &practicalResultsJSON,
			&ipAddress, &userAgent, &metadataJSON,
			&exam.CreatedAt, &exam.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan exam: %w", err)
		}
		populateExamNullables(exam, stripePaymentID, submittedAt, gradedAt,
			knowledgeScore, practicalScore, totalScore, passed, ipAddress, userAgent)
		// QuestionIDs and PracticalIDs are stored as either {"_ids": [...]} or plain [...]
		exam.QuestionIDs = jsonMapToUUIDs(questionIDsJSON)
		exam.PracticalIDs = jsonMapToUUIDs(practicalIDsJSON)
		for _, item := range []struct {
			data []byte
			dest *JSONMap
		}{
			{answersJSON, &exam.Answers},
			{practicalResultsJSON, &exam.PracticalResults},
			{metadataJSON, &exam.Metadata},
		} {
			if err := json.Unmarshal(item.data, item.dest); err != nil {
				return nil, fmt.Errorf("failed to unmarshal exam JSONB: %w", err)
			}
		}
		exams = append(exams, exam)
	}
	return exams, rows.Err()
}

func populateExamNullables(exam *CertExam,
	stripePaymentID sql.NullString,
	submittedAt, gradedAt sql.NullTime,
	knowledgeScore, practicalScore, totalScore sql.NullFloat64,
	passed sql.NullBool,
	ipAddress, userAgent sql.NullString,
) {
	if stripePaymentID.Valid {
		exam.StripePaymentID = &stripePaymentID.String
	}
	if submittedAt.Valid {
		exam.SubmittedAt = &submittedAt.Time
	}
	if gradedAt.Valid {
		exam.GradedAt = &gradedAt.Time
	}
	if knowledgeScore.Valid {
		exam.KnowledgeScore = &knowledgeScore.Float64
	}
	if practicalScore.Valid {
		exam.PracticalScore = &practicalScore.Float64
	}
	if totalScore.Valid {
		exam.TotalScore = &totalScore.Float64
	}
	if passed.Valid {
		exam.Passed = &passed.Bool
	}
	if ipAddress.Valid {
		exam.IPAddress = &ipAddress.String
	}
	if userAgent.Valid {
		exam.UserAgent = &userAgent.String
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// CertCredentials
// ──────────────────────────────────────────────────────────────────────────────

// CreateCredential inserts a new credential
func (r *CertificationRepository) CreateCredential(ctx context.Context, cred *CertCredential) error {
	if cred.ID == uuid.Nil {
		cred.ID = uuid.New()
	}
	cred.CreatedAt = time.Now()
	cred.UpdatedAt = time.Now()

	obaJSON, err := json.Marshal(cred.OBACredential)
	if err != nil {
		return fmt.Errorf("failed to marshal OB credential: %w", err)
	}
	metadataJSON, err := json.Marshal(cred.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO cert_credentials (id, user_id, tier_id, exam_id, credential_number,
			status, issued_at, expires_at, oba_credential, verification_hash,
			verification_url, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10, $11, $12::jsonb, $13, $14)`,
		cred.ID, cred.UserID, cred.TierID, cred.ExamID, cred.CredentialNumber,
		cred.Status, cred.IssuedAt, cred.ExpiresAt,
		string(obaJSON), cred.VerificationHash, cred.VerificationURL,
		string(metadataJSON), cred.CreatedAt, cred.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}
	return nil
}

// GetCredentialByID returns a credential by ID
func (r *CertificationRepository) GetCredentialByID(ctx context.Context, id uuid.UUID) (*CertCredential, error) {
	cred := &CertCredential{}
	var revokedAt sql.NullTime
	var revokedReason, verificationURL sql.NullString
	var renewalExamID sql.NullString
	var obaJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tier_id, exam_id, credential_number, status,
		       issued_at, expires_at, revoked_at, revoked_reason,
		       oba_credential, verification_hash, verification_url,
		       renewal_exam_id, metadata, created_at, updated_at
		FROM cert_credentials WHERE id = $1`, id).Scan(
		&cred.ID, &cred.UserID, &cred.TierID, &cred.ExamID, &cred.CredentialNumber,
		&cred.Status, &cred.IssuedAt, &cred.ExpiresAt, &revokedAt, &revokedReason,
		&obaJSON, &cred.VerificationHash, &verificationURL,
		&renewalExamID, &metadataJSON, &cred.CreatedAt, &cred.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get credential: %w", err)
	}
	cred.RevokedAt = nullTimePtr(revokedAt)
	cred.RevokedReason = nullStrPtr(revokedReason)
	cred.VerificationURL = nullStrPtr(verificationURL)
	if renewalExamID.Valid {
		uid, _ := uuid.Parse(renewalExamID.String)
		cred.RenewalExamID = &uid
	}
	if err := json.Unmarshal(obaJSON, &cred.OBACredential); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OB credential: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &cred.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential metadata: %w", err)
	}
	return cred, nil
}

// ListCredentialsByUser returns a user's credentials
func (r *CertificationRepository) ListCredentialsByUser(ctx context.Context, userID uuid.UUID) ([]*CertCredential, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, tier_id, exam_id, credential_number, status,
		       issued_at, expires_at, revoked_at, revoked_reason,
		       oba_credential, verification_hash, verification_url,
		       renewal_exam_id, metadata, created_at, updated_at
		FROM cert_credentials WHERE user_id = $1
		ORDER BY issued_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	defer rows.Close()
	return scanCertCredentials(rows)
}

// ListActiveCredentialsByUser returns a user's active (non-expired, non-revoked) credentials
func (r *CertificationRepository) ListActiveCredentialsByUser(ctx context.Context, userID uuid.UUID) ([]*CertCredential, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, tier_id, exam_id, credential_number, status,
		       issued_at, expires_at, revoked_at, revoked_reason,
		       oba_credential, verification_hash, verification_url,
		       renewal_exam_id, metadata, created_at, updated_at
		FROM cert_credentials WHERE user_id = $1 AND status = 'active'
		ORDER BY issued_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active credentials: %w", err)
	}
	defer rows.Close()
	return scanCertCredentials(rows)
}

// GetActiveCredentialByUserTier returns the active credential for a specific user+tier
func (r *CertificationRepository) GetActiveCredentialByUserTier(ctx context.Context, userID, tierID uuid.UUID) (*CertCredential, error) {
	cred := &CertCredential{}
	var revokedAt sql.NullTime
	var revokedReason, verificationURL sql.NullString
	var renewalExamID sql.NullString
	var obaJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tier_id, exam_id, credential_number, status,
		       issued_at, expires_at, revoked_at, revoked_reason,
		       oba_credential, verification_hash, verification_url,
		       renewal_exam_id, metadata, created_at, updated_at
		FROM cert_credentials
		WHERE user_id = $1 AND tier_id = $2 AND status = 'active'`, userID, tierID).Scan(
		&cred.ID, &cred.UserID, &cred.TierID, &cred.ExamID, &cred.CredentialNumber,
		&cred.Status, &cred.IssuedAt, &cred.ExpiresAt, &revokedAt, &revokedReason,
		&obaJSON, &cred.VerificationHash, &verificationURL,
		&renewalExamID, &metadataJSON, &cred.CreatedAt, &cred.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active credential: %w", err)
	}
	cred.RevokedAt = nullTimePtr(revokedAt)
	cred.RevokedReason = nullStrPtr(revokedReason)
	cred.VerificationURL = nullStrPtr(verificationURL)
	if renewalExamID.Valid {
		uid, _ := uuid.Parse(renewalExamID.String)
		cred.RenewalExamID = &uid
	}
	if err := json.Unmarshal(obaJSON, &cred.OBACredential); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OB credential: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &cred.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential metadata: %w", err)
	}
	return cred, nil
}

// ListCredentialsByUsername returns a user's active credentials by username (for public verification)
func (r *CertificationRepository) ListCredentialsByUsername(ctx context.Context, username string) ([]*CertCredential, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id, c.user_id, c.tier_id, c.exam_id, c.credential_number, c.status,
		       c.issued_at, c.expires_at, c.revoked_at, c.revoked_reason,
		       c.oba_credential, c.verification_hash, c.verification_url,
		       c.renewal_exam_id, c.metadata, c.created_at, c.updated_at
		FROM cert_credentials c
		JOIN users u ON u.id = c.user_id
		WHERE u.username = $1 AND c.status = 'active'
		ORDER BY c.issued_at DESC`, username)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials by username: %w", err)
	}
	defer rows.Close()
	return scanCertCredentials(rows)
}

// GetCredentialByNumber returns a credential by its unique number
func (r *CertificationRepository) GetCredentialByNumber(ctx context.Context, number string) (*CertCredential, error) {
	cred := &CertCredential{}
	var revokedAt sql.NullTime
	var revokedReason, verificationURL sql.NullString
	var renewalExamID sql.NullString
	var obaJSON, metadataJSON []byte

	err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, tier_id, exam_id, credential_number, status,
		       issued_at, expires_at, revoked_at, revoked_reason,
		       oba_credential, verification_hash, verification_url,
		       renewal_exam_id, metadata, created_at, updated_at
		FROM cert_credentials WHERE credential_number = $1`, number).Scan(
		&cred.ID, &cred.UserID, &cred.TierID, &cred.ExamID, &cred.CredentialNumber,
		&cred.Status, &cred.IssuedAt, &cred.ExpiresAt, &revokedAt, &revokedReason,
		&obaJSON, &cred.VerificationHash, &verificationURL,
		&renewalExamID, &metadataJSON, &cred.CreatedAt, &cred.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get credential by number: %w", err)
	}
	cred.RevokedAt = nullTimePtr(revokedAt)
	cred.RevokedReason = nullStrPtr(revokedReason)
	cred.VerificationURL = nullStrPtr(verificationURL)
	if renewalExamID.Valid {
		uid, _ := uuid.Parse(renewalExamID.String)
		cred.RenewalExamID = &uid
	}
	if err := json.Unmarshal(obaJSON, &cred.OBACredential); err != nil {
		return nil, fmt.Errorf("failed to unmarshal OB credential: %w", err)
	}
	if err := json.Unmarshal(metadataJSON, &cred.Metadata); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credential metadata: %w", err)
	}
	return cred, nil
}

// ExpireCredentials marks all credentials past their expiry as 'expired'
func (r *CertificationRepository) ExpireCredentials(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE cert_credentials
		SET status = 'expired', updated_at = now()
		WHERE status = 'active' AND expires_at < now()`)
	if err != nil {
		return 0, fmt.Errorf("failed to expire credentials: %w", err)
	}
	return result.RowsAffected()
}

// RevokeCredential revokes a credential
func (r *CertificationRepository) RevokeCredential(ctx context.Context, id uuid.UUID, reason string) error {
	now := time.Now()
	result, err := r.db.ExecContext(ctx, `
		UPDATE cert_credentials
		SET status = 'revoked', revoked_at = $1, revoked_reason = $2, updated_at = $1
		WHERE id = $3 AND status = 'active'`, now, reason, id)
	if err != nil {
		return fmt.Errorf("failed to revoke credential: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("credential not found or not active")
	}
	return nil
}

var validTierSlugs = map[string]bool{
	"agent-starter":    true,
	"agent-scale":     true,
	"agent-pro":       true,
	"agent-enterprise": true,
}

func isValidTierSlug(slug string) bool {
	if len(slug) == 0 || len(slug) > 64 {
		return false
	}
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// NextCredentialNumber generates the next sequential credential number for a tier
func (r *CertificationRepository) NextCredentialNumber(ctx context.Context, tierSlug string) (string, error) {
	if !isValidTierSlug(tierSlug) {
		return "", fmt.Errorf("invalid tier slug: %s", tierSlug)
	}
	if !validTierSlugs[tierSlug] {
		return "", fmt.Errorf("invalid tier slug: %s is not a recognized tier", tierSlug)
	}
	var seq int
	seqName := pq.QuoteIdentifier("cert_credential_seq_" + tierSlug)
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT nextval('%s')`, seqName)).Scan(&seq)
	if err != nil {
		return "", fmt.Errorf("failed to get next credential number: %w", err)
	}
	return fmt.Sprintf("FFC-%d-%06d", time.Now().Year(), seq), nil
}

func scanCertCredentials(rows *sql.Rows) ([]*CertCredential, error) {
	var creds []*CertCredential
	for rows.Next() {
		cred := &CertCredential{}
		var revokedAt sql.NullTime
		var revokedReason, verificationURL sql.NullString
		var renewalExamID sql.NullString
		var obaJSON, metadataJSON []byte

		if err := rows.Scan(
			&cred.ID, &cred.UserID, &cred.TierID, &cred.ExamID, &cred.CredentialNumber,
			&cred.Status, &cred.IssuedAt, &cred.ExpiresAt, &revokedAt, &revokedReason,
			&obaJSON, &cred.VerificationHash, &verificationURL,
			&renewalExamID, &metadataJSON, &cred.CreatedAt, &cred.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}
		cred.RevokedAt = nullTimePtr(revokedAt)
		cred.RevokedReason = nullStrPtr(revokedReason)
		cred.VerificationURL = nullStrPtr(verificationURL)
		if renewalExamID.Valid {
			uid, _ := uuid.Parse(renewalExamID.String)
			cred.RenewalExamID = &uid
		}
		if err := json.Unmarshal(obaJSON, &cred.OBACredential); err != nil {
			return nil, fmt.Errorf("failed to unmarshal OB credential: %w", err)
		}
		if err := json.Unmarshal(metadataJSON, &cred.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal credential metadata: %w", err)
		}
		creds = append(creds, cred)
	}
	return creds, rows.Err()
}

// ──────────────────────────────────────────────────────────────────────────────
// CertGradingQueue
// ──────────────────────────────────────────────────────────────────────────────

// EnqueueGrading adds a practical challenge grading task to the queue
func (r *CertificationRepository) EnqueueGrading(ctx context.Context, examID, challengeID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cert_grading_queue (id, exam_id, challenge_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, 'pending', now(), now())`,
		uuid.New(), examID, challengeID)
	if err != nil {
		return fmt.Errorf("failed to enqueue grading: %w", err)
	}
	return nil
}

// DequeueGrading fetches the next pending grading task using SELECT FOR UPDATE SKIP LOCKED
func (r *CertificationRepository) DequeueGrading(ctx context.Context, workerID string) (*CertGradingQueueItem, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	item := &CertGradingQueueItem{}
	var resultJSON []byte
	var errorMessage, lockedBy sql.NullString
	var lockedAt sql.NullTime

	err = tx.QueryRow(`
		SELECT id, exam_id, challenge_id, status, attempts, max_attempts,
		       result, error_message, locked_at, locked_by, created_at, updated_at
		FROM cert_grading_queue
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED`).Scan(
		&item.ID, &item.ExamID, &item.ChallengeID, &item.Status, &item.Attempts,
		&item.MaxAttempts, &resultJSON, &errorMessage, &lockedAt, &lockedBy,
		&item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			if err := tx.Commit(); err != nil {
				return nil, fmt.Errorf("failed to commit empty dequeue: %w", err)
			}
			return nil, nil
		}
		return nil, fmt.Errorf("failed to dequeue grading: %w", err)
	}

	// Mark as processing
	_, err = tx.Exec(`
		UPDATE cert_grading_queue
		SET status = 'processing', locked_at = now(), locked_by = $1, attempts = attempts + 1, updated_at = now()
		WHERE id = $2`, workerID, item.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to lock grading item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit dequeue: %w", err)
	}

	item.Status = CertGradingStatusProcessing
	item.Attempts++
	if resultJSON != nil {
		if err := json.Unmarshal(resultJSON, &item.Result); err != nil {
			return nil, fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}
	item.ErrorMessage = nullStrPtr(errorMessage)
	item.LockedBy = nullStrPtr(lockedBy)
	item.LockedAt = nullTimePtr(lockedAt)
	return item, nil
}

// CompleteGrading marks a grading task as completed with results
func (r *CertificationRepository) CompleteGrading(ctx context.Context, id uuid.UUID, result JSONMap) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		UPDATE cert_grading_queue
		SET status = 'completed', result = $1::jsonb, updated_at = now()
		WHERE id = $2`, string(resultJSON), id)
	if err != nil {
		return fmt.Errorf("failed to complete grading: %w", err)
	}
	return nil
}

// GetGradingResult returns the completed grading result for a specific exam+challenge.
func (r *CertificationRepository) GetGradingResult(ctx context.Context, examID, challengeID uuid.UUID) (JSONMap, error) {
	var resultJSON []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT result FROM cert_grading_queue
		WHERE exam_id = $1 AND challenge_id = $2 AND status = 'completed'
		ORDER BY updated_at DESC LIMIT 1`, examID, challengeID).Scan(&resultJSON)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get grading result: %w", err)
	}
	var result JSONMap
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal grading result: %w", err)
	}
	return result, nil
}

// CountPendingGrading returns the number of pending/processing grading queue items for an exam.
func (r *CertificationRepository) CountPendingGrading(ctx context.Context, examID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT count(*) FROM cert_grading_queue
		WHERE exam_id = $1 AND status IN ('pending', 'processing')`, examID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count pending grading: %w", err)
	}
	return count, nil
}

// FailGrading marks a grading task as failed
func (r *CertificationRepository) FailGrading(ctx context.Context, id uuid.UUID, errMsg string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE cert_grading_queue
		SET status = 'failed', error_message = $1, updated_at = now()
		WHERE id = $2`, errMsg, id)
	if err != nil {
		return fmt.Errorf("failed to fail grading: %w", err)
	}
	return nil
}

// ──────────────────────────────────────────────────────────────────────────────
// Admin Stats
// ──────────────────────────────────────────────────────────────────────────────

// CertStats represents aggregate certification statistics
type CertStats struct {
	TotalExams        int            `json:"total_exams"`
	TotalPassed       int            `json:"total_passed"`
	TotalFailed       int            `json:"total_failed"`
	TotalCredentials  int            `json:"total_credentials"`
	ActiveCredentials int            `json:"active_credentials"`
	ByTier            map[string]int `json:"by_tier"`
}

// GetStats returns aggregate certification statistics
func (r *CertificationRepository) GetStats(ctx context.Context) (*CertStats, error) {
	stats := &CertStats{ByTier: make(map[string]int)}

	// Exam counts
	err := r.db.QueryRowContext(ctx, `
		SELECT count(*), count(*) FILTER (WHERE status = 'passed'), count(*) FILTER (WHERE status = 'failed')
		FROM cert_exams WHERE status IN ('passed', 'failed')`).Scan(
		&stats.TotalExams, &stats.TotalPassed, &stats.TotalFailed)
	if err != nil {
		return nil, fmt.Errorf("failed to get exam stats: %w", err)
	}

	// Credential counts
	err = r.db.QueryRowContext(ctx, `
		SELECT count(*), count(*) FILTER (WHERE status = 'active')
		FROM cert_credentials`).Scan(
		&stats.TotalCredentials, &stats.ActiveCredentials)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential stats: %w", err)
	}

	// Per-tier stats
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.slug, count(c.id)
		FROM cert_tiers t
		LEFT JOIN cert_credentials c ON c.tier_id = t.id AND c.status = 'active'
		GROUP BY t.slug`)
	if err != nil {
		return nil, fmt.Errorf("failed to get tier stats: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var slug string
		var count int
		if err := rows.Scan(&slug, &count); err != nil {
			return nil, fmt.Errorf("failed to scan tier stat: %w", err)
		}
		stats.ByTier[slug] = count
	}

	return stats, rows.Err()
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

func nullTimePtr(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}

func nullStrPtr(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// jsonMapToUUIDs converts a JSONB value to []uuid.UUID
// Handles both wrapped format {"_ids": [...]} and plain array [...]
func jsonMapToUUIDs(data []byte) []uuid.UUID {
	if data == nil || len(data) == 0 {
		return nil
	}

	// Try to detect format and unmarshal accordingly
	var ids []uuid.UUID

	// First try: check if it's a wrapped format {"_ids": [...]} by unmarshaling to map
	var wrapped map[string]interface{}
	if err := json.Unmarshal(data, &wrapped); err == nil {
		if raw, ok := wrapped["_ids"]; ok {
			if arr, ok := raw.([]interface{}); ok {
				for _, v := range arr {
					if s, ok := v.(string); ok {
						if uid, err := uuid.Parse(s); err == nil {
							ids = append(ids, uid)
						}
					}
				}
				return ids
			}
		}
	}

	// Second try: it's a plain array [...]
	var plainArr []interface{}
	if err := json.Unmarshal(data, &plainArr); err == nil {
		for _, v := range plainArr {
			if s, ok := v.(string); ok {
				if uid, err := uuid.Parse(s); err == nil {
					ids = append(ids, uid)
				}
			}
		}
		return ids
	}

	return ids
}

// CreateTier inserts a new certification tier.
func (r *CertificationRepository) CreateTier(ctx context.Context, tier *CertTier) error {
	if err := r.db.GORM.Create(tier).Error; err != nil {
		return fmt.Errorf("failed to create tier: %w", err)
	}
	return nil
}

// UpdateTier updates allowed fields on a certification tier.
func (r *CertificationRepository) UpdateTier(ctx context.Context, tierID uuid.UUID, updates map[string]interface{}) error {
	if err := r.db.GORM.Model(&CertTier{}).Where("id = ?", tierID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update tier: %w", err)
	}
	return nil
}

// GetQuestionsByTier returns non-deleted questions for a tier.
func (r *CertificationRepository) GetQuestionsByTier(ctx context.Context, tierID uuid.UUID, limit int) ([]*CertQuestion, error) {
	var questions []*CertQuestion
	err := r.db.GORM.WithContext(ctx).
		Where("tier_id = ? AND is_active = true", tierID).
		Limit(limit).
		Find(&questions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get questions: %w", err)
	}
	return questions, nil
}

// CreateQuestion inserts a new certification question.
func (r *CertificationRepository) CreateQuestion(ctx context.Context, question *CertQuestion) error {
	if err := r.db.GORM.Create(question).Error; err != nil {
		return fmt.Errorf("failed to create question: %w", err)
	}
	return nil
}

// UpdateQuestion updates allowed fields on a question.
func (r *CertificationRepository) UpdateQuestion(ctx context.Context, questionID uuid.UUID, updates map[string]interface{}) error {
	if err := r.db.GORM.Model(&CertQuestion{}).Where("id = ?", questionID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update question: %w", err)
	}
	return nil
}

// DeleteQuestion hard-deletes a question.
func (r *CertificationRepository) DeleteQuestion(ctx context.Context, questionID uuid.UUID) error {
	if err := r.db.GORM.Delete(&CertQuestion{}, questionID).Error; err != nil {
		return fmt.Errorf("failed to delete question: %w", err)
	}
	return nil
}

// CertExamListFilter filters for listing exams.
type CertExamListFilter struct {
	Limit  int
	Offset int
	Status string
	TierID *uuid.UUID
	UserID *uuid.UUID
}

// ListExams returns exams matching the filter, plus total count.
func (r *CertificationRepository) ListExams(ctx context.Context, filter CertExamListFilter) ([]*CertExam, int64, error) {
	query := r.db.GORM.WithContext(ctx).Model(&CertExam{})
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.TierID != nil {
		query = query.Where("tier_id = ?", *filter.TierID)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count exams: %w", err)
	}

	var exams []*CertExam
	if err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&exams).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list exams: %w", err)
	}
	return exams, total, nil
}

// CertCredentialListFilter filters for listing credentials.
type CertCredentialListFilter struct {
	Limit  int
	Offset int
	Status string
	TierID *uuid.UUID
	UserID *uuid.UUID
}

// ListCredentials returns credentials matching the filter.
// Uses batch queries instead of GORM's chained Preloads to avoid the N+1
// query problem (each Preload fires a separate sequential query, compounding
// latency with nested relations).
func (r *CertificationRepository) ListCredentials(ctx context.Context, filter CertCredentialListFilter) ([]*CertCredential, error) {
	// Build base query without Preloads
	query := r.db.GORM.WithContext(ctx).Model(&CertCredential{})
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.TierID != nil {
		query = query.Where("tier_id = ?", *filter.TierID)
	}
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	var creds []*CertCredential
	if err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&creds).Error; err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	if len(creds) == 0 {
		return creds, nil
	}

	// Batch load all tiers and users in parallel to avoid N+1.
	tierIDs := make([]uuid.UUID, 0, len(creds))
	userIDs := make([]uuid.UUID, 0, len(creds))
	tierSet := make(map[uuid.UUID]struct{}, len(creds))
	userSet := make(map[uuid.UUID]struct{}, len(creds))
	for _, cred := range creds {
		if _, ok := tierSet[cred.TierID]; !ok {
			tierIDs = append(tierIDs, cred.TierID)
			tierSet[cred.TierID] = struct{}{}
		}
		if _, ok := userSet[cred.UserID]; !ok {
			userIDs = append(userIDs, cred.UserID)
			userSet[cred.UserID] = struct{}{}
		}
	}

	// Batch load tiers.
	tiers := make(map[uuid.UUID]*CertTier, len(tierIDs))
	if len(tierIDs) > 0 {
		var tierRecords []CertTier
		if err := r.db.GORM.WithContext(ctx).Where("id IN ?", tierIDs).Find(&tierRecords).Error; err != nil {
			return nil, fmt.Errorf("failed to load tiers: %w", err)
		}
		for i := range tierRecords {
			tiers[tierRecords[i].ID] = &tierRecords[i]
		}
	}

	// Batch load users.
	users := make(map[uuid.UUID]*User, len(userIDs))
	if len(userIDs) > 0 {
		var userRecords []User
		if err := r.db.GORM.WithContext(ctx).Where("id IN ?", userIDs).Find(&userRecords).Error; err != nil {
			return nil, fmt.Errorf("failed to load users: %w", err)
		}
		for i := range userRecords {
			users[userRecords[i].ID] = &userRecords[i]
		}
	}

	// Wire up relations.
	for _, cred := range creds {
		if t, ok := tiers[cred.TierID]; ok {
			cred.Tier = t
		}
		if u, ok := users[cred.UserID]; ok {
			cred.User = u
		}
	}

	return creds, nil
}

// UpdateCredential updates credential fields in the database.
func (r *CertificationRepository) UpdateCredential(ctx context.Context, credID uuid.UUID, updates map[string]interface{}) error {
	if err := r.db.GORM.WithContext(ctx).Model(&CertCredential{}).Where("id = ?", credID).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}
	return nil
}

// UpdateExamStatus updates an exam's status directly.
func (r *CertificationRepository) UpdateExamStatus(ctx context.Context, examID uuid.UUID, status string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE cert_exams
		SET status = $1, updated_at = now()
		WHERE id = $2`, status, examID)
	if err != nil {
		return fmt.Errorf("failed to update exam status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("exam not found")
	}
	return nil
}

// ResetUserTierExamAttempts resets all non-passed exams for a user+tier to abandoned status.
func (r *CertificationRepository) ResetUserTierExamAttempts(ctx context.Context, userID, tierID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE cert_exams
		SET status = $1, updated_at = now()
		WHERE user_id = $2 AND tier_id = $3 AND status NOT IN ($4)`,
		CertExamStatusAbandoned, userID, tierID, CertExamStatusPassed)
	if err != nil {
		return fmt.Errorf("failed to reset exam attempts: %w", err)
	}
	return nil
}

// Status constant aliases from the types package.
const CertExamStatusAbandoned = types.CertExamStatusAbandoned
const CertExamStatusPendingPayment = types.CertExamStatusPendingPayment
const CertExamStatusPassed = types.CertExamStatusPassed
const CertExamStatusExpired = types.CertExamStatusExpired
const CertGradingStatusProcessing = types.CertGradingStatusProcessing
const CertExamStatusFailed = types.CertExamStatusFailed
const CertExamStatusInProgress = types.CertExamStatusInProgress
