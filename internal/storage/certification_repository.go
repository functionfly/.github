package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
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

	// Build placeholders $1, $2, ...
	args := make([]interface{}, len(ids))
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT id, category, difficulty, question_text, question_format, options, points
		FROM cert_questions
		WHERE id IN (%s) AND is_active = true`, joinStrings(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
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

	args := make([]interface{}, len(ids))
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	query := fmt.Sprintf(`
		SELECT id, correct_answers, points
		FROM cert_questions
		WHERE id IN (%s) AND is_active = true`, joinStrings(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
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
	for _, item := range []struct {
		data []byte
		dest *JSONMap
	}{
		{questionIDsJSON, &exam.QuestionIDs},
		{practicalIDsJSON, &exam.PracticalIDs},
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
	now := time.Now()
	_, err := r.db.ExecContext(ctx, `
		UPDATE cert_exams
		SET answers = jsonb_set(COALESCE(answers, '{}'), $1, $2::jsonb, true),
		    updated_at = $3
		WHERE id = $4 AND status = 'in_progress'`,
		[]string{questionID.String()}, mustJSON(answer), now, examID)
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
		SET status = $1, updated_at = $2
		WHERE id = $3 AND status = 'in_progress'`,
		CertExamStatusAbandoned, now, examID)
	if err != nil {
		return fmt.Errorf("failed to abandon exam: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("exam not found or already submitted")
	}
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
	for _, item := range []struct {
		data []byte
		dest *JSONMap
	}{
		{questionIDsJSON, &exam.QuestionIDs},
		{practicalIDsJSON, &exam.PracticalIDs},
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
		for _, item := range []struct {
			data []byte
			dest *JSONMap
		}{
			{questionIDsJSON, &exam.QuestionIDs},
			{practicalIDsJSON, &exam.PracticalIDs},
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

// NextCredentialNumber generates the next sequential credential number for a tier
func (r *CertificationRepository) NextCredentialNumber(ctx context.Context, tierSlug string) (string, error) {
	var seq int
	err := r.db.QueryRowContext(ctx,
		fmt.Sprintf(`SELECT nextval('cert_credential_seq_%s')`, tierSlug)).Scan(&seq)
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
	TotalExams      int            `json:"total_exams"`
	TotalPassed     int            `json:"total_passed"`
	TotalFailed     int            `json:"total_failed"`
	TotalCredentials int           `json:"total_credentials"`
	ActiveCredentials int          `json:"active_credentials"`
	ByTier          map[string]int `json:"by_tier"`
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
