package support

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for support data access
type Repository interface {
	// Conversations
	CreateConversation(ctx context.Context, c *SupportConversation) error
	GetConversation(ctx context.Context, id uuid.UUID) (*SupportConversation, error)
	ListConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*SupportConversation, error)
	ListActiveConversations(ctx context.Context, limit, offset int) ([]*SupportConversation, error)
	UpdateConversationStatus(ctx context.Context, id uuid.UUID, status SupportStatus) error
	UpdateConversationStaff(ctx context.Context, id uuid.UUID, staffID uuid.UUID) error
	ResolveConversation(ctx context.Context, id uuid.UUID, resolvedBy uuid.UUID, note string) error
	IncrementAIAttempts(ctx context.Context, id uuid.UUID) error

	// Messages
	CreateMessage(ctx context.Context, m *SupportMessage) error
	ListMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*SupportMessage, error)
	GetMessage(ctx context.Context, id uuid.UUID) (*SupportMessage, error)

	// Staff availability
	UpsertStaffAvailability(ctx context.Context, a *StaffAvailability) error
	GetStaffAvailability(ctx context.Context, staffID uuid.UUID) (*StaffAvailability, error)
	ListOnlineStaff(ctx context.Context) ([]*StaffAvailability, error)
	SetStaffOnline(ctx context.Context, staffID uuid.UUID, online bool) error
	IncrementActiveChats(ctx context.Context, staffID uuid.UUID) error
	DecrementActiveChats(ctx context.Context, staffID uuid.UUID) error

	// Emergency
	CreateEmergencyRequest(ctx context.Context, e *EmergencyFixRequest) error
	GetEmergencyRequest(ctx context.Context, id uuid.UUID) (*EmergencyFixRequest, error)
	UpdateEmergencyStatus(ctx context.Context, id uuid.UUID, staffID uuid.UUID, status string) error
	ListPendingEmergencies(ctx context.Context) ([]*EmergencyFixRequest, error)

	// Participants
	AddParticipant(ctx context.Context, p *SupportConversationParticipant) error
	RemoveParticipant(ctx context.Context, conversationID, userID uuid.UUID) error
	ListParticipants(ctx context.Context, conversationID uuid.UUID) ([]*SupportConversationParticipant, error)
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateConversation creates a new support conversation
func (r *PostgresRepository) CreateConversation(ctx context.Context, c *SupportConversation) error {
	query := `
		INSERT INTO support_conversations (
			user_id, type, status, priority, title, function_ref,
			deployment_id, deployment_logs, deployment_error,
			ai_handled, ai_attempts, is_emergency, metadata
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`
	functionRef := c.FunctionRefJSON
	if functionRef == nil {
		functionRef = json.RawMessage("{}")
	}
	metadata := c.Metadata
	if metadata == nil {
		metadata = json.RawMessage("{}")
	}
	return r.db.QueryRowContext(ctx, query,
		c.UserID, c.Type, c.Status, c.Priority, c.Title, functionRef,
		c.DeploymentID, c.DeploymentLogs, c.DeploymentError,
		c.AIHandled, c.AIAttempts, c.IsEmergency, metadata,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

// GetConversation retrieves a support conversation by ID
func (r *PostgresRepository) GetConversation(ctx context.Context, id uuid.UUID) (*SupportConversation, error) {
	query := `
		SELECT
			id,
			user_id,
			type::text,
			status::text,
			priority::text,
			title,
			function_ref::text,
			deployment_id::text,
			deployment_logs,
			deployment_error,
			ai_handled,
			ai_attempts,
			staff_id::text,
			staff_joined_at,
			resolved_at,
			resolved_by_id::text,
			resolution_note,
			is_emergency,
			emergency_code,
			metadata,
			created_at,
			updated_at
		FROM support_conversations
		WHERE id = $1
	`
	c := &SupportConversation{}
	var functionRefJSON sql.NullString
	var staffID, resolvedByID, deploymentID sql.NullString
	var staffJoinedAt, resolvedAt sql.NullTime
	var resolutionNote sql.NullString
	var emergencyCode sql.NullString
	var deploymentLogs, deploymentError sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.UserID, &c.Type, &c.Status, &c.Priority, &c.Title, &functionRefJSON,
		&deploymentID, &deploymentLogs, &deploymentError,
		&c.AIHandled, &c.AIAttempts, &staffID, &staffJoinedAt,
		&resolvedAt, &resolvedByID, &resolutionNote,
		&c.IsEmergency, &emergencyCode, &c.Metadata, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if functionRefJSON.Valid {
		c.FunctionRefJSON = json.RawMessage(functionRefJSON.String)
		var ref FunctionRef
		json.Unmarshal([]byte(functionRefJSON.String), &ref)
		c.FunctionRef = &ref
	}

	if staffID.Valid {
		id, _ := uuid.Parse(staffID.String)
		c.StaffID = &id
	}
	if resolvedByID.Valid {
		id, _ := uuid.Parse(resolvedByID.String)
		c.ResolvedByID = &id
	}
	if deploymentID.Valid {
		id, _ := uuid.Parse(deploymentID.String)
		c.DeploymentID = &id
	}
	if staffJoinedAt.Valid {
		c.StaffJoinedAt = &staffJoinedAt.Time
	}
	if resolvedAt.Valid {
		c.ResolvedAt = &resolvedAt.Time
	}
	if resolutionNote.Valid {
		c.ResolutionNote = resolutionNote.String
	}
	if emergencyCode.Valid {
		c.EmergencyCode = emergencyCode.String
	}
	if deploymentLogs.Valid {
		c.DeploymentLogs = deploymentLogs.String
	}
	if deploymentError.Valid {
		c.DeploymentError = deploymentError.String
	}

	return c, nil
}

// ListConversations lists support conversations for a user
func (r *PostgresRepository) ListConversations(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*SupportConversation, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT
			id,
			user_id,
			type::text,
			status::text,
			priority::text,
			title,
			function_ref::text,
			deployment_id::text,
			deployment_logs,
			deployment_error,
			ai_handled,
			ai_attempts,
			staff_id::text,
			staff_joined_at,
			resolved_at,
			resolved_by_id::text,
			resolution_note,
			is_emergency,
			emergency_code,
			metadata,
			created_at,
			updated_at
		FROM support_conversations
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanConversations(rows)
}

// ListActiveConversations lists all active support conversations
func (r *PostgresRepository) ListActiveConversations(ctx context.Context, limit, offset int) ([]*SupportConversation, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT
			id,
			user_id,
			type::text,
			status::text,
			priority::text,
			title,
			function_ref::text,
			deployment_id::text,
			deployment_logs,
			deployment_error,
			ai_handled,
			ai_attempts,
			staff_id::text,
			staff_joined_at,
			resolved_at,
			resolved_by_id::text,
			resolution_note,
			is_emergency,
			emergency_code,
			metadata,
			created_at,
			updated_at
		FROM support_conversations
		WHERE status IN ('active', 'pending', 'escalated')
		ORDER BY
			CASE WHEN is_emergency THEN 0 WHEN priority = 'critical' THEN 1 WHEN priority = 'high' THEN 2 ELSE 3 END,
			created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanConversations(rows)
}

func (r *PostgresRepository) scanConversations(rows *sql.Rows) ([]*SupportConversation, error) {
	var conversations []*SupportConversation
	for rows.Next() {
		c := &SupportConversation{}
		var functionRefJSON sql.NullString
		var staffID, resolvedByID, deploymentID sql.NullString
		var staffJoinedAt, resolvedAt sql.NullTime
		var resolutionNote sql.NullString
		var emergencyCode sql.NullString
		var deploymentLogs, deploymentError sql.NullString

		err := rows.Scan(
			&c.ID, &c.UserID, &c.Type, &c.Status, &c.Priority, &c.Title, &functionRefJSON,
			&deploymentID, &deploymentLogs, &deploymentError,
			&c.AIHandled, &c.AIAttempts, &staffID, &staffJoinedAt,
			&resolvedAt, &resolvedByID, &resolutionNote,
			&c.IsEmergency, &emergencyCode, &c.Metadata, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if functionRefJSON.Valid {
			c.FunctionRefJSON = json.RawMessage(functionRefJSON.String)
			var ref FunctionRef
			json.Unmarshal([]byte(functionRefJSON.String), &ref)
			c.FunctionRef = &ref
		}

		if staffID.Valid {
			id, _ := uuid.Parse(staffID.String)
			c.StaffID = &id
		}
		if resolvedByID.Valid {
			id, _ := uuid.Parse(resolvedByID.String)
			c.ResolvedByID = &id
		}
		if deploymentID.Valid {
			id, _ := uuid.Parse(deploymentID.String)
			c.DeploymentID = &id
		}
		if staffJoinedAt.Valid {
			c.StaffJoinedAt = &staffJoinedAt.Time
		}
		if resolvedAt.Valid {
			c.ResolvedAt = &resolvedAt.Time
		}
		if resolutionNote.Valid {
			c.ResolutionNote = resolutionNote.String
		}
		if emergencyCode.Valid {
			c.EmergencyCode = emergencyCode.String
		}
		if deploymentLogs.Valid {
			c.DeploymentLogs = deploymentLogs.String
		}
		if deploymentError.Valid {
			c.DeploymentError = deploymentError.String
		}

		conversations = append(conversations, c)
	}
	return conversations, rows.Err()
}

// UpdateConversationStatus updates the status of a support conversation
func (r *PostgresRepository) UpdateConversationStatus(ctx context.Context, id uuid.UUID, status SupportStatus) error {
	query := `UPDATE support_conversations SET status = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	return err
}

// UpdateConversationStaff assigns staff to a conversation
func (r *PostgresRepository) UpdateConversationStaff(ctx context.Context, id uuid.UUID, staffID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE support_conversations SET staff_id = $1, staff_joined_at = $2, status = 'active', updated_at = $3 WHERE id = $4`
	_, err := r.db.ExecContext(ctx, query, staffID, now, now, id)
	return err
}

// ResolveConversation marks a conversation as resolved
func (r *PostgresRepository) ResolveConversation(ctx context.Context, id uuid.UUID, resolvedBy uuid.UUID, note string) error {
	now := time.Now()
	query := `
		UPDATE support_conversations
		SET status = 'resolved', resolved_at = $1, resolved_by_id = $2, resolution_note = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.ExecContext(ctx, query, now, resolvedBy, note, now, id)
	return err
}

// IncrementAIAttempts increments the AI attempts counter
func (r *PostgresRepository) IncrementAIAttempts(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE support_conversations SET ai_attempts = ai_attempts + 1, updated_at = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// CreateMessage creates a new support message
func (r *PostgresRepository) CreateMessage(ctx context.Context, m *SupportMessage) error {
	query := `
		INSERT INTO support_messages (conversation_id, author_id, author_type, message_type, content, ai_confidence, ai_model, embeddings, attachments)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at
	`
	return r.db.QueryRowContext(ctx, query,
		m.ConversationID, m.AuthorID, m.AuthorType, m.MessageType, m.Content,
		m.AIConfidence, m.AIModel, m.Embeddings, m.Attachments,
	).Scan(&m.ID, &m.CreatedAt)
}

// ListMessages lists messages for a support conversation
func (r *PostgresRepository) ListMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*SupportMessage, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT id, conversation_id, author_id, author_type, message_type, content,
			ai_confidence, ai_model, embeddings, attachments, created_at
		FROM support_messages
		WHERE conversation_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*SupportMessage
	for rows.Next() {
		m := &SupportMessage{}
		var aiConfidence sql.NullFloat64
		var aiModel sql.NullString
		var embeddings, attachments []byte

		err := rows.Scan(
			&m.ID, &m.ConversationID, &m.AuthorID, &m.AuthorType, &m.MessageType, &m.Content,
			&aiConfidence, &aiModel, &embeddings, &attachments, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if aiConfidence.Valid {
			m.AIConfidence = &aiConfidence.Float64
		}
		if aiModel.Valid {
			m.AIModel = aiModel.String
		}
		if len(embeddings) > 0 {
			m.Embeddings = json.RawMessage(embeddings)
		}
		if len(attachments) > 0 {
			m.Attachments = json.RawMessage(attachments)
		}

		messages = append(messages, m)
	}
	return messages, rows.Err()
}

// GetMessage retrieves a support message by ID
func (r *PostgresRepository) GetMessage(ctx context.Context, id uuid.UUID) (*SupportMessage, error) {
	query := `
		SELECT id, conversation_id, author_id, author_type, message_type, content,
			ai_confidence, ai_model, embeddings, attachments, created_at
		FROM support_messages
		WHERE id = $1
	`
	m := &SupportMessage{}
	var aiConfidence sql.NullFloat64
	var aiModel sql.NullString
	var embeddings, attachments []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.ConversationID, &m.AuthorID, &m.AuthorType, &m.MessageType, &m.Content,
		&aiConfidence, &aiModel, &embeddings, &attachments, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if aiConfidence.Valid {
		m.AIConfidence = &aiConfidence.Float64
	}
	if aiModel.Valid {
		m.AIModel = aiModel.String
	}
	if len(embeddings) > 0 {
		m.Embeddings = json.RawMessage(embeddings)
	}
	if len(attachments) > 0 {
		m.Attachments = json.RawMessage(attachments)
	}

	return m, nil
}

// UpsertStaffAvailability inserts or updates staff availability
func (r *PostgresRepository) UpsertStaffAvailability(ctx context.Context, a *StaffAvailability) error {
	query := `
		INSERT INTO staff_availability (staff_id, is_online, last_seen, max_chats, active_chats, can_accept, specialties)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (staff_id) DO UPDATE SET
			is_online = EXCLUDED.is_online,
			last_seen = EXCLUDED.last_seen,
			max_chats = EXCLUDED.max_chats,
			active_chats = EXCLUDED.active_chats,
			can_accept = EXCLUDED.can_accept,
			specialties = EXCLUDED.specialties,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		a.StaffID, a.IsOnline, a.LastSeen, a.MaxChats, a.ActiveChats, a.CanAccept, a.Specialties,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

// GetStaffAvailability gets staff availability by staff ID
func (r *PostgresRepository) GetStaffAvailability(ctx context.Context, staffID uuid.UUID) (*StaffAvailability, error) {
	query := `
		SELECT id, staff_id, is_online, last_seen, max_chats, active_chats, can_accept, specialties, created_at, updated_at
		FROM staff_availability
		WHERE staff_id = $1
	`
	a := &StaffAvailability{}
	var specialties sql.NullString

	err := r.db.QueryRowContext(ctx, query, staffID).Scan(
		&a.ID, &a.StaffID, &a.IsOnline, &a.LastSeen, &a.MaxChats, &a.ActiveChats, &a.CanAccept, &specialties, &a.CreatedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if specialties.Valid {
		a.Specialties = json.RawMessage(specialties.String)
	}

	return a, nil
}

// ListOnlineStaff lists all online staff
func (r *PostgresRepository) ListOnlineStaff(ctx context.Context) ([]*StaffAvailability, error) {
	query := `
		SELECT id, staff_id, is_online, last_seen, max_chats, active_chats, can_accept, specialties, created_at, updated_at
		FROM staff_availability
		WHERE is_online = true AND can_accept = true AND active_chats < max_chats
		ORDER BY active_chats ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staff []*StaffAvailability
	for rows.Next() {
		a := &StaffAvailability{}
		var specialties sql.NullString

		err := rows.Scan(
			&a.ID, &a.StaffID, &a.IsOnline, &a.LastSeen, &a.MaxChats, &a.ActiveChats, &a.CanAccept, &specialties, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if specialties.Valid {
			a.Specialties = json.RawMessage(specialties.String)
		}

		staff = append(staff, a)
	}
	return staff, rows.Err()
}

// SetStaffOnline sets staff online/offline status
func (r *PostgresRepository) SetStaffOnline(ctx context.Context, staffID uuid.UUID, online bool) error {
	now := time.Now()
	query := `
		INSERT INTO staff_availability (staff_id, is_online, last_seen)
		VALUES ($1, $2, $3)
		ON CONFLICT (staff_id) DO UPDATE SET
			is_online = EXCLUDED.is_online,
			last_seen = EXCLUDED.last_seen,
			updated_at = CURRENT_TIMESTAMP
	`
	_, err := r.db.ExecContext(ctx, query, staffID, online, now)
	return err
}

// IncrementActiveChats increments active chats for a staff member
func (r *PostgresRepository) IncrementActiveChats(ctx context.Context, staffID uuid.UUID) error {
	query := `UPDATE staff_availability SET active_chats = active_chats + 1, updated_at = $1 WHERE staff_id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), staffID)
	return err
}

// DecrementActiveChats decrements active chats for a staff member
func (r *PostgresRepository) DecrementActiveChats(ctx context.Context, staffID uuid.UUID) error {
	query := `UPDATE staff_availability SET active_chats = GREATEST(active_chats - 1, 0), updated_at = $1 WHERE staff_id = $2`
	_, err := r.db.ExecContext(ctx, query, time.Now(), staffID)
	return err
}

// CreateEmergencyRequest creates an emergency fix request
func (r *PostgresRepository) CreateEmergencyRequest(ctx context.Context, e *EmergencyFixRequest) error {
	query := `
		INSERT INTO emergency_fix_requests (conversation_id, user_id, function_id, reason, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRowContext(ctx, query,
		e.ConversationID, e.UserID, e.FunctionID, e.Reason, e.Status,
	).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

// GetEmergencyRequest retrieves an emergency request by ID
func (r *PostgresRepository) GetEmergencyRequest(ctx context.Context, id uuid.UUID) (*EmergencyFixRequest, error) {
	query := `
		SELECT id, conversation_id, user_id, function_id, reason, status,
			staff_id, staff_accepted_at, resolved_at, fix_description, created_at, updated_at
		FROM emergency_fix_requests
		WHERE id = $1
	`
	e := &EmergencyFixRequest{}
	var staffID sql.NullString
	var staffAcceptedAt, resolvedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.ConversationID, &e.UserID, &e.FunctionID, &e.Reason, &e.Status,
		&staffID, &staffAcceptedAt, &resolvedAt, &e.FixDescription, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if staffID.Valid {
		id, _ := uuid.Parse(staffID.String)
		e.StaffID = &id
	}
	if staffAcceptedAt.Valid {
		e.StaffAcceptedAt = &staffAcceptedAt.Time
	}
	if resolvedAt.Valid {
		e.ResolvedAt = &resolvedAt.Time
	}

	return e, nil
}

// UpdateEmergencyStatus updates the status of an emergency request
func (r *PostgresRepository) UpdateEmergencyStatus(ctx context.Context, id uuid.UUID, staffID uuid.UUID, status string) error {
	now := time.Now()
	var query string
	var args []interface{}

	if status == "accepted" {
		query = `UPDATE emergency_fix_requests SET staff_id = $1, staff_accepted_at = $2, status = $3, updated_at = $4 WHERE id = $5`
		args = []interface{}{staffID, now, status, now, id}
	} else if status == "resolved" {
		query = `UPDATE emergency_fix_requests SET status = $1, resolved_at = $2, updated_at = $3 WHERE id = $4`
		args = []interface{}{status, now, now, id}
	} else {
		query = `UPDATE emergency_fix_requests SET status = $1, updated_at = $2 WHERE id = $3`
		args = []interface{}{status, now, id}
	}

	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

// ListPendingEmergencies lists all pending emergency requests
func (r *PostgresRepository) ListPendingEmergencies(ctx context.Context) ([]*EmergencyFixRequest, error) {
	query := `
		SELECT id, conversation_id, user_id, function_id, reason, status,
			staff_id, staff_accepted_at, resolved_at, fix_description, created_at, updated_at
		FROM emergency_fix_requests
		WHERE status IN ('pending', 'accepted')
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []*EmergencyFixRequest
	for rows.Next() {
		e := &EmergencyFixRequest{}
		var staffID sql.NullString
		var staffAcceptedAt, resolvedAt sql.NullTime

		err := rows.Scan(
			&e.ID, &e.ConversationID, &e.UserID, &e.FunctionID, &e.Reason, &e.Status,
			&staffID, &staffAcceptedAt, &resolvedAt, &e.FixDescription, &e.CreatedAt, &e.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if staffID.Valid {
			id, _ := uuid.Parse(staffID.String)
			e.StaffID = &id
		}
		if staffAcceptedAt.Valid {
			e.StaffAcceptedAt = &staffAcceptedAt.Time
		}
		if resolvedAt.Valid {
			e.ResolvedAt = &resolvedAt.Time
		}

		requests = append(requests, e)
	}
	return requests, rows.Err()
}

// AddParticipant adds a participant to a support conversation
func (r *PostgresRepository) AddParticipant(ctx context.Context, p *SupportConversationParticipant) error {
	query := `
		INSERT INTO support_conversation_participants (conversation_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT DO NOTHING
		RETURNING id
	`
	return r.db.QueryRowContext(ctx, query, p.ConversationID, p.UserID, p.Role, p.JoinedAt).Scan(&p.ID)
}

// RemoveParticipant removes a participant from a support conversation
func (r *PostgresRepository) RemoveParticipant(ctx context.Context, conversationID, userID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE support_conversation_participants SET left_at = $1 WHERE conversation_id = $2 AND user_id = $3 AND left_at IS NULL`
	_, err := r.db.ExecContext(ctx, query, now, conversationID, userID)
	return err
}

// ListParticipants lists participants in a support conversation
func (r *PostgresRepository) ListParticipants(ctx context.Context, conversationID uuid.UUID) ([]*SupportConversationParticipant, error) {
	query := `
		SELECT id, conversation_id, user_id, role, joined_at, left_at
		FROM support_conversation_participants
		WHERE conversation_id = $1 AND left_at IS NULL
		ORDER BY joined_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*SupportConversationParticipant
	for rows.Next() {
		p := &SupportConversationParticipant{}
		var leftAt sql.NullTime

		err := rows.Scan(&p.ID, &p.ConversationID, &p.UserID, &p.Role, &p.JoinedAt, &leftAt)
		if err != nil {
			return nil, err
		}

		if leftAt.Valid {
			p.LeftAt = &leftAt.Time
		}

		participants = append(participants, p)
	}
	return participants, rows.Err()
}

// Helper to convert error to not found
func notFound(err error) error {
	if err == sql.ErrNoRows {
		return fmt.Errorf("not found")
	}
	return err
}
