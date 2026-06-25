package storage

import (
	"context"

	"github.com/google/uuid"
)

// PostgresDB methods: FWOS employee operations.

// Department operations
func (db *PostgresDB) CreateDepartment(ctx context.Context, dept *Department) (*Department, error) {
	return db.employeeRepository.CreateDepartment(ctx, dept)
}

func (db *PostgresDB) GetDepartmentByID(ctx context.Context, id int64) (*Department, error) {
	return db.employeeRepository.GetDepartmentByID(ctx, id)
}

func (db *PostgresDB) ListDepartments(ctx context.Context, tenantID uuid.UUID) ([]*Department, error) {
	return db.employeeRepository.ListDepartments(ctx, tenantID)
}

func (db *PostgresDB) UpdateDepartment(ctx context.Context, id int64, updates map[string]interface{}) error {
	return db.employeeRepository.UpdateDepartment(ctx, id, updates)
}

func (db *PostgresDB) DeleteDepartment(ctx context.Context, id int64) error {
	return db.employeeRepository.DeleteDepartment(ctx, id)
}

// Employee operations
func (db *PostgresDB) CreateEmployee(ctx context.Context, emp *Employee) (*Employee, error) {
	return db.employeeRepository.CreateEmployee(ctx, emp)
}

func (db *PostgresDB) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*Employee, error) {
	return db.employeeRepository.GetEmployeeByID(ctx, id)
}

func (db *PostgresDB) GetEmployeeByUserID(ctx context.Context, userID uuid.UUID) (*Employee, error) {
	return db.employeeRepository.GetEmployeeByUserID(ctx, userID)
}

func (db *PostgresDB) GetEmployeeByFFID(ctx context.Context, ffid string) (*Employee, error) {
	return db.employeeRepository.GetEmployeeByFFID(ctx, ffid)
}

func (db *PostgresDB) ListEmployees(ctx context.Context, tenantID uuid.UUID, opts ListEmployeesOpts) ([]*Employee, int, error) {
	return db.employeeRepository.ListEmployees(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateEmployee(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.employeeRepository.UpdateEmployee(ctx, id, updates)
}

func (db *PostgresDB) DeleteEmployee(ctx context.Context, id uuid.UUID) error {
	return db.employeeRepository.DeleteEmployee(ctx, id)
}

func (db *PostgresDB) GetDirectReports(ctx context.Context, managerID uuid.UUID) ([]*Employee, error) {
	return db.employeeRepository.GetDirectReports(ctx, managerID)
}

func (db *PostgresDB) GetOrgChart(ctx context.Context, tenantID uuid.UUID) ([]*Employee, error) {
	return db.employeeRepository.GetOrgChart(ctx, tenantID)
}

// Employee Department operations
func (db *PostgresDB) AddEmployeeDepartment(ctx context.Context, ed *EmployeeDepartment) error {
	return db.employeeRepository.AddEmployeeDepartment(ctx, ed)
}

// Employee Skill operations
func (db *PostgresDB) AddEmployeeSkill(ctx context.Context, skill *EmployeeSkill) (*EmployeeSkill, error) {
	return db.employeeRepository.AddEmployeeSkill(ctx, skill)
}

func (db *PostgresDB) GetEmployeeSkills(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeSkill, error) {
	return db.employeeRepository.GetEmployeeSkills(ctx, employeeID)
}

func (db *PostgresDB) UpdateEmployeeSkill(ctx context.Context, id int64, updates map[string]interface{}) error {
	return db.employeeRepository.UpdateEmployeeSkill(ctx, id, updates)
}

func (db *PostgresDB) RemoveEmployeeSkill(ctx context.Context, id int64) error {
	return db.employeeRepository.RemoveEmployeeSkill(ctx, id)
}

// Employee Certification operations
func (db *PostgresDB) AddEmployeeCertification(ctx context.Context, cert *EmployeeCertification) (*EmployeeCertification, error) {
	return db.employeeRepository.AddEmployeeCertification(ctx, cert)
}

func (db *PostgresDB) GetEmployeeCertifications(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeCertification, error) {
	return db.employeeRepository.GetEmployeeCertifications(ctx, employeeID)
}

// Employee Achievement operations
func (db *PostgresDB) AwardEmployeeAchievement(ctx context.Context, ach *EmployeeAchievement) (*EmployeeAchievement, error) {
	return db.employeeRepository.AwardEmployeeAchievement(ctx, ach)
}

func (db *PostgresDB) GetEmployeeAchievements(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeAchievement, error) {
	return db.employeeRepository.GetEmployeeAchievements(ctx, employeeID)
}

// Project operations
func (db *PostgresDB) CreateProject(ctx context.Context, proj *Project) (*Project, error) {
	return db.employeeRepository.CreateProject(ctx, proj)
}

func (db *PostgresDB) GetProjectByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	return db.employeeRepository.GetProjectByID(ctx, id)
}

func (db *PostgresDB) ListProjects(ctx context.Context, tenantID uuid.UUID, opts ListProjectsOpts) ([]*Project, int, error) {
	return db.employeeRepository.ListProjects(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateProject(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.employeeRepository.UpdateProject(ctx, id, updates)
}

func (db *PostgresDB) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return db.employeeRepository.DeleteProject(ctx, id)
}

// Task operations
func (db *PostgresDB) CreateTask(ctx context.Context, task *Task) (*Task, error) {
	return db.employeeRepository.CreateTask(ctx, task)
}

func (db *PostgresDB) GetTaskByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	return db.employeeRepository.GetTaskByID(ctx, id)
}

func (db *PostgresDB) ListTasks(ctx context.Context, tenantID uuid.UUID, opts ListTasksOpts) ([]*Task, int, error) {
	return db.employeeRepository.ListTasks(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateTask(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.employeeRepository.UpdateTask(ctx, id, updates)
}

func (db *PostgresDB) UpdateTaskStatus(ctx context.Context, id uuid.UUID, status string) error {
	return db.employeeRepository.UpdateTaskStatus(ctx, id, status)
}

func (db *PostgresDB) DeleteTask(ctx context.Context, id uuid.UUID) error {
	return db.employeeRepository.DeleteTask(ctx, id)
}

// Task Comment operations
func (db *PostgresDB) CreateTaskComment(ctx context.Context, comment *TaskComment) (*TaskComment, error) {
	return db.employeeRepository.CreateTaskComment(ctx, comment)
}

func (db *PostgresDB) GetTaskComments(ctx context.Context, taskID uuid.UUID) ([]*TaskComment, error) {
	return db.employeeRepository.GetTaskComments(ctx, taskID)
}

// Learning Course operations
func (db *PostgresDB) CreateLearningCourse(ctx context.Context, course *LearningCourse) (*LearningCourse, error) {
	return db.employeeRepository.CreateLearningCourse(ctx, course)
}

func (db *PostgresDB) GetLearningCourseByID(ctx context.Context, id uuid.UUID) (*LearningCourse, error) {
	return db.employeeRepository.GetLearningCourseByID(ctx, id)
}

func (db *PostgresDB) ListLearningCourses(ctx context.Context, tenantID uuid.UUID) ([]*LearningCourse, error) {
	return db.employeeRepository.ListLearningCourses(ctx, tenantID)
}

// Employee Learning operations
func (db *PostgresDB) EnrollCourse(ctx context.Context, el *EmployeeLearning) (*EmployeeLearning, error) {
	return db.employeeRepository.EnrollCourse(ctx, el)
}

func (db *PostgresDB) GetEmployeeLearning(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeLearning, error) {
	return db.employeeRepository.GetEmployeeLearning(ctx, employeeID)
}

func (db *PostgresDB) UpdateLearningProgress(ctx context.Context, id int64, updates map[string]interface{}) error {
	return db.employeeRepository.UpdateLearningProgress(ctx, id, updates)
}

// Knowledge Article operations
func (db *PostgresDB) CreateKnowledgeArticle(ctx context.Context, article *KnowledgeArticle) (*KnowledgeArticle, error) {
	return db.employeeRepository.CreateKnowledgeArticle(ctx, article)
}

func (db *PostgresDB) GetKnowledgeArticleByID(ctx context.Context, id uuid.UUID) (*KnowledgeArticle, error) {
	return db.employeeRepository.GetKnowledgeArticleByID(ctx, id)
}

func (db *PostgresDB) GetKnowledgeArticleBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*KnowledgeArticle, error) {
	return db.employeeRepository.GetKnowledgeArticleBySlug(ctx, tenantID, slug)
}

func (db *PostgresDB) ListKnowledgeArticles(ctx context.Context, tenantID uuid.UUID, opts ListKnowledgeOpts) ([]*KnowledgeArticle, int, error) {
	return db.employeeRepository.ListKnowledgeArticles(ctx, tenantID, opts)
}

func (db *PostgresDB) SearchKnowledgeArticles(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]*KnowledgeArticle, error) {
	return db.employeeRepository.SearchKnowledgeArticles(ctx, tenantID, query, limit)
}

func (db *PostgresDB) UpdateKnowledgeArticle(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.employeeRepository.UpdateKnowledgeArticle(ctx, id, updates)
}

func (db *PostgresDB) DeleteKnowledgeArticle(ctx context.Context, id uuid.UUID) error {
	return db.employeeRepository.DeleteKnowledgeArticle(ctx, id)
}

// Compensation operations
func (db *PostgresDB) CreateCompensationRecord(ctx context.Context, rec *CompensationRecord) (*CompensationRecord, error) {
	return db.employeeRepository.CreateCompensationRecord(ctx, rec)
}

func (db *PostgresDB) GetActiveCompensation(ctx context.Context, employeeID uuid.UUID) (*CompensationRecord, error) {
	return db.employeeRepository.GetActiveCompensation(ctx, employeeID)
}

func (db *PostgresDB) UpdateCompensationRecord(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.employeeRepository.UpdateCompensationRecord(ctx, id, updates)
}

// Equity Grant operations
func (db *PostgresDB) CreateEquityGrant(ctx context.Context, grant *EquityGrant) (*EquityGrant, error) {
	return db.employeeRepository.CreateEquityGrant(ctx, grant)
}

func (db *PostgresDB) ListEquityGrants(ctx context.Context, employeeID uuid.UUID) ([]*EquityGrant, error) {
	return db.employeeRepository.ListEquityGrants(ctx, employeeID)
}

// Compensation Access Log operations
func (db *PostgresDB) LogCompensationAccess(ctx context.Context, log *CompensationAccessLog) error {
	return db.employeeRepository.LogCompensationAccess(ctx, log)
}

func (db *PostgresDB) GetCompensationAccessLog(ctx context.Context, employeeID uuid.UUID) ([]*CompensationAccessLog, error) {
	return db.employeeRepository.GetCompensationAccessLog(ctx, employeeID)
}

// FWOS Notification operations
func (db *PostgresDB) CreateNotification(ctx context.Context, notif *FWOSNotification) (*FWOSNotification, error) {
	return db.employeeRepository.CreateNotification(ctx, notif)
}

func (db *PostgresDB) ListNotifications(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit, offset int) ([]*FWOSNotification, int, error) {
	return db.employeeRepository.ListNotifications(ctx, userID, unreadOnly, limit, offset)
}

func (db *PostgresDB) CountUnreadNotifications(ctx context.Context, userID uuid.UUID) (int, error) {
	return db.employeeRepository.CountUnreadNotifications(ctx, userID)
}

func (db *PostgresDB) MarkNotificationRead(ctx context.Context, id uuid.UUID) error {
	return db.employeeRepository.MarkNotificationRead(ctx, id)
}

func (db *PostgresDB) MarkAllNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	return db.employeeRepository.MarkAllNotificationsRead(ctx, userID)
}

// Phase 2: AI Chat operations

func (db *PostgresDB) CreateChatSession(ctx context.Context, sess *AIChatSession) (*AIChatSession, error) {
	return db.phase2Repository.CreateChatSession(ctx, sess)
}

func (db *PostgresDB) GetChatSessionByID(ctx context.Context, id uuid.UUID) (*AIChatSession, error) {
	return db.phase2Repository.GetChatSessionByID(ctx, id)
}

func (db *PostgresDB) ListChatSessions(ctx context.Context, userID uuid.UUID, opts ListAIChatSessionsOpts) ([]*AIChatSession, int, error) {
	return db.phase2Repository.ListChatSessions(ctx, userID, opts)
}

func (db *PostgresDB) UpdateChatSessionTitle(ctx context.Context, id uuid.UUID, title string) error {
	return db.phase2Repository.UpdateChatSessionTitle(ctx, id, title)
}

func (db *PostgresDB) DeleteChatSession(ctx context.Context, id uuid.UUID) error {
	return db.phase2Repository.DeleteChatSession(ctx, id)
}

func (db *PostgresDB) CreateChatMessage(ctx context.Context, msg *AIChatMessage) (*AIChatMessage, error) {
	return db.phase2Repository.CreateChatMessage(ctx, msg)
}

func (db *PostgresDB) ListChatMessages(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]*AIChatMessage, error) {
	return db.phase2Repository.ListChatMessages(ctx, sessionID, limit, offset)
}

// Phase 2: Performance Goal operations

func (db *PostgresDB) CreatePerformanceGoal(ctx context.Context, goal *PerformanceGoal) (*PerformanceGoal, error) {
	return db.phase2Repository.CreatePerformanceGoal(ctx, goal)
}

func (db *PostgresDB) GetPerformanceGoalByID(ctx context.Context, id uuid.UUID) (*PerformanceGoal, error) {
	return db.phase2Repository.GetPerformanceGoalByID(ctx, id)
}

func (db *PostgresDB) ListPerformanceGoals(ctx context.Context, employeeID uuid.UUID, opts ListPerformanceGoalsOpts) ([]*PerformanceGoal, int, error) {
	return db.phase2Repository.ListPerformanceGoals(ctx, employeeID, opts)
}

func (db *PostgresDB) UpdatePerformanceGoal(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase2Repository.UpdatePerformanceGoal(ctx, id, updates)
}

// Phase 2: Performance Review operations

func (db *PostgresDB) CreatePerformanceReview(ctx context.Context, rev *PerformanceReview) (*PerformanceReview, error) {
	return db.phase2Repository.CreatePerformanceReview(ctx, rev)
}

func (db *PostgresDB) GetPerformanceReviewByID(ctx context.Context, id uuid.UUID) (*PerformanceReview, error) {
	return db.phase2Repository.GetPerformanceReviewByID(ctx, id)
}

func (db *PostgresDB) ListPerformanceReviews(ctx context.Context, employeeID uuid.UUID, opts ListPerformanceReviewsOpts) ([]*PerformanceReview, int, error) {
	return db.phase2Repository.ListPerformanceReviews(ctx, employeeID, opts)
}

func (db *PostgresDB) UpdatePerformanceReview(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase2Repository.UpdatePerformanceReview(ctx, id, updates)
}

// Phase 2: Peer Feedback operations

func (db *PostgresDB) CreatePeerFeedback(ctx context.Context, fb *PeerFeedback) (*PeerFeedback, error) {
	return db.phase2Repository.CreatePeerFeedback(ctx, fb)
}

func (db *PostgresDB) ListPeerFeedback(ctx context.Context, toEmployeeID uuid.UUID, limit, offset int) ([]*PeerFeedback, error) {
	return db.phase2Repository.ListPeerFeedback(ctx, toEmployeeID, limit, offset)
}

// Phase 2: Time Entry operations

func (db *PostgresDB) CreateTimeEntry(ctx context.Context, entry *TimeEntry) (*TimeEntry, error) {
	return db.phase2Repository.CreateTimeEntry(ctx, entry)
}

func (db *PostgresDB) GetTimeEntryByID(ctx context.Context, id uuid.UUID) (*TimeEntry, error) {
	return db.phase2Repository.GetTimeEntryByID(ctx, id)
}

func (db *PostgresDB) ListTimeEntries(ctx context.Context, employeeID uuid.UUID, opts ListTimeEntriesOpts) ([]*TimeEntry, int, error) {
	return db.phase2Repository.ListTimeEntries(ctx, employeeID, opts)
}

func (db *PostgresDB) UpdateTimeEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase2Repository.UpdateTimeEntry(ctx, id, updates)
}

func (db *PostgresDB) DeleteTimeEntry(ctx context.Context, id uuid.UUID) error {
	return db.phase2Repository.DeleteTimeEntry(ctx, id)
}

// Phase 2: PTO Request operations

func (db *PostgresDB) CreatePTORequest(ctx context.Context, req *PTORequest) (*PTORequest, error) {
	return db.phase2Repository.CreatePTORequest(ctx, req)
}

func (db *PostgresDB) GetPTORequestByID(ctx context.Context, id uuid.UUID) (*PTORequest, error) {
	return db.phase2Repository.GetPTORequestByID(ctx, id)
}

func (db *PostgresDB) ListPTORequests(ctx context.Context, employeeID uuid.UUID, opts ListPTORequestsOpts) ([]*PTORequest, int, error) {
	return db.phase2Repository.ListPTORequests(ctx, employeeID, opts)
}

func (db *PostgresDB) UpdatePTORequestStatus(ctx context.Context, id uuid.UUID, status string, approvedBy *uuid.UUID, notes *string) error {
	return db.phase2Repository.UpdatePTORequestStatus(ctx, id, status, approvedBy, notes)
}

// Phase 3: Innovation Grant operations

func (db *PostgresDB) CreateInnovationGrant(ctx context.Context, grant *InnovationGrant) (*InnovationGrant, error) {
	return db.phase3Repository.CreateInnovationGrant(ctx, grant)
}

func (db *PostgresDB) GetInnovationGrantByID(ctx context.Context, id uuid.UUID) (*InnovationGrant, error) {
	return db.phase3Repository.GetInnovationGrantByID(ctx, id)
}

func (db *PostgresDB) ListInnovationGrants(ctx context.Context, tenantID uuid.UUID, opts ListInnovationGrantsOpts) ([]*InnovationGrant, int, error) {
	return db.phase3Repository.ListInnovationGrants(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateInnovationGrant(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase3Repository.UpdateInnovationGrant(ctx, id, updates)
}

func (db *PostgresDB) CreateInnovationGrantVote(ctx context.Context, vote *InnovationGrantVote) (*InnovationGrantVote, error) {
	return db.phase3Repository.CreateInnovationGrantVote(ctx, vote)
}

func (db *PostgresDB) GetInnovationGrantVoteByVoter(ctx context.Context, grantID, voterID uuid.UUID) (*InnovationGrantVote, error) {
	return db.phase3Repository.GetInnovationGrantVoteByVoter(ctx, grantID, voterID)
}

// Phase 3: Marketplace Opportunity operations

func (db *PostgresDB) CreateMarketplaceOpportunity(ctx context.Context, opp *MarketplaceOpportunity) (*MarketplaceOpportunity, error) {
	return db.phase3Repository.CreateMarketplaceOpportunity(ctx, opp)
}

func (db *PostgresDB) GetMarketplaceOpportunityByID(ctx context.Context, id uuid.UUID) (*MarketplaceOpportunity, error) {
	return db.phase3Repository.GetMarketplaceOpportunityByID(ctx, id)
}

func (db *PostgresDB) ListMarketplaceOpportunities(ctx context.Context, tenantID uuid.UUID, opts ListMarketplaceOpportunitiesOpts) ([]*MarketplaceOpportunity, int, error) {
	return db.phase3Repository.ListMarketplaceOpportunities(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateMarketplaceOpportunity(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase3Repository.UpdateMarketplaceOpportunity(ctx, id, updates)
}

func (db *PostgresDB) CreateMarketplaceApplication(ctx context.Context, app *MarketplaceApplication) (*MarketplaceApplication, error) {
	return db.phase3Repository.CreateMarketplaceApplication(ctx, app)
}

func (db *PostgresDB) GetMarketplaceApplicationByID(ctx context.Context, id uuid.UUID) (*MarketplaceApplication, error) {
	return db.phase3Repository.GetMarketplaceApplicationByID(ctx, id)
}

func (db *PostgresDB) ListMarketplaceApplications(ctx context.Context, opportunityID uuid.UUID, limit, offset int) ([]*MarketplaceApplication, error) {
	return db.phase3Repository.ListMarketplaceApplications(ctx, opportunityID, limit, offset)
}

func (db *PostgresDB) UpdateMarketplaceApplicationStatus(ctx context.Context, id uuid.UUID, status string) error {
	return db.phase3Repository.UpdateMarketplaceApplicationStatus(ctx, id, status)
}

// Phase 3: Career Path operations

func (db *PostgresDB) CreateCareerPath(ctx context.Context, path *CareerPath) (*CareerPath, error) {
	return db.phase3Repository.CreateCareerPath(ctx, path)
}

func (db *PostgresDB) GetCareerPathByID(ctx context.Context, id uuid.UUID) (*CareerPath, error) {
	return db.phase3Repository.GetCareerPathByID(ctx, id)
}

func (db *PostgresDB) ListCareerPaths(ctx context.Context, tenantID uuid.UUID, opts ListCareerPathsOpts) ([]*CareerPath, int, error) {
	return db.phase3Repository.ListCareerPaths(ctx, tenantID, opts)
}

func (db *PostgresDB) CreateEmployeeCareerProgress(ctx context.Context, prog *EmployeeCareerProgress) (*EmployeeCareerProgress, error) {
	return db.phase3Repository.CreateEmployeeCareerProgress(ctx, prog)
}

func (db *PostgresDB) GetEmployeeCareerProgressByEmployee(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeCareerProgress, error) {
	return db.phase3Repository.GetEmployeeCareerProgressByEmployee(ctx, employeeID)
}

// Phase 3: Mentorship Match operations

func (db *PostgresDB) CreateMentorshipMatch(ctx context.Context, match *MentorshipMatch) (*MentorshipMatch, error) {
	return db.phase3Repository.CreateMentorshipMatch(ctx, match)
}

func (db *PostgresDB) GetMentorshipMatchByID(ctx context.Context, id uuid.UUID) (*MentorshipMatch, error) {
	return db.phase3Repository.GetMentorshipMatchByID(ctx, id)
}

func (db *PostgresDB) ListMentorshipMatches(ctx context.Context, employeeID uuid.UUID, opts ListMentorshipMatchesOpts) ([]*MentorshipMatch, int, error) {
	return db.phase3Repository.ListMentorshipMatches(ctx, employeeID, opts)
}

func (db *PostgresDB) UpdateMentorshipMatch(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase3Repository.UpdateMentorshipMatch(ctx, id, updates)
}

// Phase 3: Document operations

func (db *PostgresDB) CreateDocument(ctx context.Context, doc *Document) (*Document, error) {
	return db.phase3Repository.CreateDocument(ctx, doc)
}

func (db *PostgresDB) GetDocumentByID(ctx context.Context, id uuid.UUID) (*Document, error) {
	return db.phase3Repository.GetDocumentByID(ctx, id)
}

func (db *PostgresDB) ListDocuments(ctx context.Context, tenantID uuid.UUID, opts ListDocumentsOpts) ([]*Document, int, error) {
	return db.phase3Repository.ListDocuments(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateDocument(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase3Repository.UpdateDocument(ctx, id, updates)
}

func (db *PostgresDB) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	return db.phase3Repository.DeleteDocument(ctx, id)
}

func (db *PostgresDB) IncrementDocumentViewCount(ctx context.Context, id uuid.UUID) error {
	return db.phase3Repository.IncrementDocumentViewCount(ctx, id)
}

func (db *PostgresDB) CreateDocumentShare(ctx context.Context, share *DocumentShare) (*DocumentShare, error) {
	return db.phase3Repository.CreateDocumentShare(ctx, share)
}

func (db *PostgresDB) ListDocumentShares(ctx context.Context, documentID uuid.UUID) ([]*DocumentShare, error) {
	return db.phase3Repository.ListDocumentShares(ctx, documentID)
}

// Phase 4: Team Health operations

func (db *PostgresDB) CreateTeamHealthMetric(ctx context.Context, m *TeamHealthMetric) (*TeamHealthMetric, error) {
	return db.phase4Repository.CreateTeamHealthMetric(ctx, m)
}

func (db *PostgresDB) GetTeamHealthMetrics(ctx context.Context, tenantID uuid.UUID, opts ListTeamHealthOpts) ([]*TeamHealthMetric, int, error) {
	return db.phase4Repository.GetTeamHealthMetrics(ctx, tenantID, opts)
}

func (db *PostgresDB) GetLatestTeamHealth(ctx context.Context, tenantID uuid.UUID, departmentID int64) (*TeamHealthMetric, error) {
	return db.phase4Repository.GetLatestTeamHealth(ctx, tenantID, departmentID)
}

func (db *PostgresDB) CalculateTeamHealth(ctx context.Context, tenantID uuid.UUID, departmentID int64) (*TeamHealthMetric, error) {
	return db.phase4Repository.CalculateTeamHealth(ctx, tenantID, departmentID)
}

// Phase 4: Skills Graph operations

func (db *PostgresDB) CreateSkillsGraphEntry(ctx context.Context, s *SkillsGraph) (*SkillsGraph, error) {
	return db.phase4Repository.CreateSkillsGraphEntry(ctx, s)
}

func (db *PostgresDB) GetSkillsGraph(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*SkillsGraph, int, error) {
	return db.phase4Repository.GetSkillsGraph(ctx, tenantID, limit, offset)
}

func (db *PostgresDB) GetSkillGaps(ctx context.Context, tenantID uuid.UUID, limit int) ([]*SkillsGraph, error) {
	return db.phase4Repository.GetSkillGaps(ctx, tenantID, limit)
}

func (db *PostgresDB) CalculateSkillsGraph(ctx context.Context, tenantID uuid.UUID) ([]*SkillsGraph, error) {
	return db.phase4Repository.CalculateSkillsGraph(ctx, tenantID)
}

// Phase 4: Reputation Score operations

func (db *PostgresDB) CreateReputationScore(ctx context.Context, s *ReputationScore) (*ReputationScore, error) {
	return db.phase4Repository.CreateReputationScore(ctx, s)
}

func (db *PostgresDB) GetReputationScores(ctx context.Context, employeeID uuid.UUID) ([]*ReputationScore, error) {
	return db.phase4Repository.GetReputationScores(ctx, employeeID)
}

func (db *PostgresDB) GetReputationLeaderboard(ctx context.Context, tenantID uuid.UUID, category string, limit, offset int) ([]*ReputationScore, int, error) {
	return db.phase4Repository.GetReputationLeaderboard(ctx, tenantID, category, limit, offset)
}

func (db *PostgresDB) UpdateReputationScore(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase4Repository.UpdateReputationScore(ctx, id, updates)
}

func (db *PostgresDB) CalculateReputation(ctx context.Context, employeeID uuid.UUID, tenantID uuid.UUID) ([]*ReputationScore, error) {
	return db.phase4Repository.CalculateReputation(ctx, employeeID, tenantID)
}

// Phase 4: Digital Badge operations

func (db *PostgresDB) CreateDigitalBadge(ctx context.Context, b *DigitalBadge) (*DigitalBadge, error) {
	return db.phase4Repository.CreateDigitalBadge(ctx, b)
}

func (db *PostgresDB) ListDigitalBadges(ctx context.Context, tenantID uuid.UUID, opts ListBadgesOpts) ([]*DigitalBadge, int, error) {
	return db.phase4Repository.ListDigitalBadges(ctx, tenantID, opts)
}

func (db *PostgresDB) GetDigitalBadgeBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (*DigitalBadge, error) {
	return db.phase4Repository.GetDigitalBadgeBySlug(ctx, tenantID, slug)
}

func (db *PostgresDB) UpdateDigitalBadge(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase4Repository.UpdateDigitalBadge(ctx, id, updates)
}

func (db *PostgresDB) DeleteDigitalBadge(ctx context.Context, id uuid.UUID) error {
	return db.phase4Repository.DeleteDigitalBadge(ctx, id)
}

// Phase 4: Employee Badge operations

func (db *PostgresDB) AwardEmployeeBadge(ctx context.Context, eb *EmployeeBadge) (*EmployeeBadge, error) {
	return db.phase4Repository.AwardEmployeeBadge(ctx, eb)
}

func (db *PostgresDB) GetEmployeeBadges(ctx context.Context, employeeID uuid.UUID) ([]*EmployeeBadge, error) {
	return db.phase4Repository.GetEmployeeBadges(ctx, employeeID)
}

func (db *PostgresDB) RevokeEmployeeBadge(ctx context.Context, employeeID, badgeID uuid.UUID) error {
	return db.phase4Repository.RevokeEmployeeBadge(ctx, employeeID, badgeID)
}

// Phase 4: Living Memory operations

func (db *PostgresDB) CreateLivingMemoryEntry(ctx context.Context, e *LivingMemoryEntry) (*LivingMemoryEntry, error) {
	return db.phase4Repository.CreateLivingMemoryEntry(ctx, e)
}

func (db *PostgresDB) GetLivingMemoryEntry(ctx context.Context, id uuid.UUID) (*LivingMemoryEntry, error) {
	return db.phase4Repository.GetLivingMemoryEntry(ctx, id)
}

func (db *PostgresDB) ListLivingMemoryEntries(ctx context.Context, tenantID uuid.UUID, opts ListLivingMemoryOpts) ([]*LivingMemoryEntry, int, error) {
	return db.phase4Repository.ListLivingMemoryEntries(ctx, tenantID, opts)
}

func (db *PostgresDB) SearchLivingMemory(ctx context.Context, tenantID uuid.UUID, opts SearchLivingMemoryOpts) ([]*LivingMemoryEntry, error) {
	return db.phase4Repository.SearchLivingMemory(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateLivingMemoryEntry(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase4Repository.UpdateLivingMemoryEntry(ctx, id, updates)
}

func (db *PostgresDB) IncrementLivingMemoryViewCount(ctx context.Context, id uuid.UUID) error {
	return db.phase4Repository.IncrementLivingMemoryViewCount(ctx, id)
}

func (db *PostgresDB) DeleteLivingMemoryEntry(ctx context.Context, id uuid.UUID) error {
	return db.phase4Repository.DeleteLivingMemoryEntry(ctx, id)
}

// Phase 4: Mission Control operations

func (db *PostgresDB) CreateMissionControlSnapshot(ctx context.Context, s *MissionControlSnapshot) (*MissionControlSnapshot, error) {
	return db.phase4Repository.CreateMissionControlSnapshot(ctx, s)
}

func (db *PostgresDB) GetLatestMissionControlSnapshot(ctx context.Context, tenantID uuid.UUID) (*MissionControlSnapshot, error) {
	return db.phase4Repository.GetLatestMissionControlSnapshot(ctx, tenantID)
}

func (db *PostgresDB) ListMissionControlSnapshots(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*MissionControlSnapshot, int, error) {
	return db.phase4Repository.ListMissionControlSnapshots(ctx, tenantID, limit, offset)
}

func (db *PostgresDB) GenerateMissionControlSnapshot(ctx context.Context, tenantID uuid.UUID) (*MissionControlSnapshot, error) {
	return db.phase4Repository.GenerateMissionControlSnapshot(ctx, tenantID)
}

// Phase 5: Incident operations

func (db *PostgresDB) CreateFWOSIncident(ctx context.Context, inc *FWOSIncident) (*FWOSIncident, error) {
	return db.phase5Repository.CreateFWOSIncident(ctx, inc)
}

func (db *PostgresDB) GetFWOSIncidentByID(ctx context.Context, id uuid.UUID) (*FWOSIncident, error) {
	return db.phase5Repository.GetFWOSIncidentByID(ctx, id)
}

func (db *PostgresDB) ListFWOSIncidents(ctx context.Context, tenantID uuid.UUID, opts ListIncidentsOpts) ([]*FWOSIncident, int, error) {
	return db.phase5Repository.ListFWOSIncidents(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateFWOSIncident(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase5Repository.UpdateFWOSIncident(ctx, id, updates)
}

func (db *PostgresDB) CreateIncidentEvent(ctx context.Context, ev *IncidentEvent) (*IncidentEvent, error) {
	return db.phase5Repository.CreateIncidentEvent(ctx, ev)
}

func (db *PostgresDB) ListIncidentEvents(ctx context.Context, incidentID uuid.UUID, limit, offset int) ([]*IncidentEvent, error) {
	return db.phase5Repository.ListIncidentEvents(ctx, incidentID, limit, offset)
}

func (db *PostgresDB) AddIncidentResponder(ctx context.Context, resp *IncidentResponder) (*IncidentResponder, error) {
	return db.phase5Repository.AddIncidentResponder(ctx, resp)
}

func (db *PostgresDB) ListIncidentResponders(ctx context.Context, incidentID uuid.UUID) ([]*IncidentResponder, error) {
	return db.phase5Repository.ListIncidentResponders(ctx, incidentID)
}

// Phase 5: Postmortem operations

func (db *PostgresDB) CreatePostmortem(ctx context.Context, pm *Postmortem) (*Postmortem, error) {
	return db.phase5Repository.CreatePostmortem(ctx, pm)
}

func (db *PostgresDB) GetPostmortemByIncident(ctx context.Context, incidentID uuid.UUID) (*Postmortem, error) {
	return db.phase5Repository.GetPostmortemByIncident(ctx, incidentID)
}

func (db *PostgresDB) UpdatePostmortem(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase5Repository.UpdatePostmortem(ctx, id, updates)
}

// Phase 5: Lifecycle operations

func (db *PostgresDB) CreateLifecycleEvent(ctx context.Context, ev *LifecycleEvent) (*LifecycleEvent, error) {
	return db.phase5Repository.CreateLifecycleEvent(ctx, ev)
}

func (db *PostgresDB) ListLifecycleEvents(ctx context.Context, employeeID uuid.UUID, opts ListLifecycleEventsOpts) ([]*LifecycleEvent, int, error) {
	return db.phase5Repository.ListLifecycleEvents(ctx, employeeID, opts)
}

func (db *PostgresDB) CreateLifecycleWorkflow(ctx context.Context, wf *LifecycleWorkflow) (*LifecycleWorkflow, error) {
	return db.phase5Repository.CreateLifecycleWorkflow(ctx, wf)
}

func (db *PostgresDB) ListLifecycleWorkflows(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*LifecycleWorkflow, int, error) {
	return db.phase5Repository.ListLifecycleWorkflows(ctx, tenantID, limit, offset)
}

func (db *PostgresDB) GetLifecycleWorkflowByID(ctx context.Context, id uuid.UUID) (*LifecycleWorkflow, error) {
	return db.phase5Repository.GetLifecycleWorkflowByID(ctx, id)
}

func (db *PostgresDB) CreateLifecycleWorkflowInstance(ctx context.Context, inst *LifecycleWorkflowInstance) (*LifecycleWorkflowInstance, error) {
	return db.phase5Repository.CreateLifecycleWorkflowInstance(ctx, inst)
}

func (db *PostgresDB) GetLifecycleWorkflowInstance(ctx context.Context, id uuid.UUID) (*LifecycleWorkflowInstance, error) {
	return db.phase5Repository.GetLifecycleWorkflowInstance(ctx, id)
}

func (db *PostgresDB) UpdateLifecycleWorkflowInstance(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase5Repository.UpdateLifecycleWorkflowInstance(ctx, id, updates)
}

// Phase 5: Feature Flag operations

func (db *PostgresDB) CreateFeatureFlag(ctx context.Context, ff *FeatureFlag) (*FeatureFlag, error) {
	return db.phase5Repository.CreateFeatureFlag(ctx, ff)
}

func (db *PostgresDB) GetFeatureFlagByKey(ctx context.Context, tenantID uuid.UUID, key string) (*FeatureFlag, error) {
	return db.phase5Repository.GetFeatureFlagByKey(ctx, tenantID, key)
}

func (db *PostgresDB) ListFeatureFlags(ctx context.Context, tenantID uuid.UUID, opts ListFeatureFlagsOpts) ([]*FeatureFlag, int, error) {
	return db.phase5Repository.ListFeatureFlags(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateFeatureFlag(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase5Repository.UpdateFeatureFlag(ctx, id, updates)
}

func (db *PostgresDB) DeleteFeatureFlag(ctx context.Context, id uuid.UUID) error {
	return db.phase5Repository.DeleteFeatureFlag(ctx, id)
}

// Phase 5: Data Classification operations

func (db *PostgresDB) CreateDataClassification(ctx context.Context, dc *DataClassification) (*DataClassification, error) {
	return db.phase5Repository.CreateDataClassification(ctx, dc)
}

func (db *PostgresDB) GetDataClassification(ctx context.Context, tenantID uuid.UUID, resourceType string, resourceID uuid.UUID) (*DataClassification, error) {
	return db.phase5Repository.GetDataClassification(ctx, tenantID, resourceType, resourceID)
}

func (db *PostgresDB) ListDataClassifications(ctx context.Context, tenantID uuid.UUID, opts ListDataClassificationsOpts) ([]*DataClassification, int, error) {
	return db.phase5Repository.ListDataClassifications(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateDataClassification(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase5Repository.UpdateDataClassification(ctx, id, updates)
}

func (db *PostgresDB) DeleteDataClassification(ctx context.Context, id uuid.UUID) error {
	return db.phase5Repository.DeleteDataClassification(ctx, id)
}

// Phase 5: Employee Certificate operations

func (db *PostgresDB) IssueCertificate(ctx context.Context, cert *EmployeeCertificate) (*EmployeeCertificate, error) {
	return db.phase5Repository.IssueCertificate(ctx, cert)
}

func (db *PostgresDB) GetCertificateBySerial(ctx context.Context, serial string) (*EmployeeCertificate, error) {
	return db.phase5Repository.GetCertificateBySerial(ctx, serial)
}

func (db *PostgresDB) ListCertificates(ctx context.Context, employeeID uuid.UUID, opts ListCertificatesOpts) ([]*EmployeeCertificate, int, error) {
	return db.phase5Repository.ListCertificates(ctx, employeeID, opts)
}

func (db *PostgresDB) RevokeCertificate(ctx context.Context, id uuid.UUID, reason string) error {
	return db.phase5Repository.RevokeCertificate(ctx, id, reason)
}

// Phase 5: FWOS Event operations

func (db *PostgresDB) CreateFWOSEvent(ctx context.Context, ev *FWOSEvent) (*FWOSEvent, error) {
	return db.phase5Repository.CreateFWOSEvent(ctx, ev)
}

func (db *PostgresDB) ListFWOSEvents(ctx context.Context, tenantID uuid.UUID, opts ListFWOSEventsOpts) ([]*FWOSEvent, int, error) {
	return db.phase5Repository.ListFWOSEvents(ctx, tenantID, opts)
}

// Phase 6: Email Account operations

func (db *PostgresDB) CreateEmailAccount(ctx context.Context, ea *EmailAccount) (*EmailAccount, error) {
	return db.phase6Repository.CreateEmailAccount(ctx, ea)
}

func (db *PostgresDB) GetEmailAccountByID(ctx context.Context, id uuid.UUID) (*EmailAccount, error) {
	return db.phase6Repository.GetEmailAccountByID(ctx, id)
}

func (db *PostgresDB) ListEmailAccounts(ctx context.Context, tenantID uuid.UUID, opts ListEmailAccountsOpts) ([]*EmailAccount, int, error) {
	return db.phase6Repository.ListEmailAccounts(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateEmailAccount(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase6Repository.UpdateEmailAccount(ctx, id, updates)
}

// Phase 6: Device operations

func (db *PostgresDB) CreateDevice(ctx context.Context, d *Device) (*Device, error) {
	return db.phase6Repository.CreateDevice(ctx, d)
}

func (db *PostgresDB) GetDeviceByID(ctx context.Context, id uuid.UUID) (*Device, error) {
	return db.phase6Repository.GetDeviceByID(ctx, id)
}

func (db *PostgresDB) ListDevices(ctx context.Context, tenantID uuid.UUID, opts ListDevicesOpts) ([]*Device, int, error) {
	return db.phase6Repository.ListDevices(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateDevice(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase6Repository.UpdateDevice(ctx, id, updates)
}

// Phase 6: SSO Provisioning Config operations

func (db *PostgresDB) CreateSSOProvisioningConfig(ctx context.Context, cfg *SSOProvisioningConfig) (*SSOProvisioningConfig, error) {
	return db.phase6Repository.CreateSSOProvisioningConfig(ctx, cfg)
}

func (db *PostgresDB) GetSSOProvisioningConfigByID(ctx context.Context, id uuid.UUID) (*SSOProvisioningConfig, error) {
	return db.phase6Repository.GetSSOProvisioningConfigByID(ctx, id)
}

func (db *PostgresDB) ListSSOProvisioningConfigs(ctx context.Context, tenantID uuid.UUID, opts ListSSOProvisioningConfigsOpts) ([]*SSOProvisioningConfig, int, error) {
	return db.phase6Repository.ListSSOProvisioningConfigs(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateSSOProvisioningConfig(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase6Repository.UpdateSSOProvisioningConfig(ctx, id, updates)
}

// Phase 6: SSO Provisioning Log operations

func (db *PostgresDB) CreateSSOProvisioningLog(ctx context.Context, log *SSOProvisioningLog) (*SSOProvisioningLog, error) {
	return db.phase6Repository.CreateSSOProvisioningLog(ctx, log)
}

func (db *PostgresDB) ListSSOProvisioningLogs(ctx context.Context, configID uuid.UUID, opts ListSSOProvisioningLogsOpts) ([]*SSOProvisioningLog, int, error) {
	return db.phase6Repository.ListSSOProvisioningLogs(ctx, configID, opts)
}

// Phase 6: Wallet Pass operations

func (db *PostgresDB) CreateWalletPass(ctx context.Context, wp *WalletPass) (*WalletPass, error) {
	return db.phase6Repository.CreateWalletPass(ctx, wp)
}

func (db *PostgresDB) GetWalletPassByID(ctx context.Context, id uuid.UUID) (*WalletPass, error) {
	return db.phase6Repository.GetWalletPassByID(ctx, id)
}

func (db *PostgresDB) GetWalletPassByPassID(ctx context.Context, passID string) (*WalletPass, error) {
	return db.phase6Repository.GetWalletPassByPassID(ctx, passID)
}

func (db *PostgresDB) GetWalletPassByQRToken(ctx context.Context, qrToken string) (*WalletPass, error) {
	return db.phase6Repository.GetWalletPassByQRToken(ctx, qrToken)
}

func (db *PostgresDB) ListWalletPasses(ctx context.Context, employeeID uuid.UUID, opts ListWalletPassesOpts) ([]*WalletPass, int, error) {
	return db.phase6Repository.ListWalletPasses(ctx, employeeID, opts)
}

func (db *PostgresDB) UpdateWalletPass(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.phase6Repository.UpdateWalletPass(ctx, id, updates)
}

// Phase 6: Push Subscription operations

func (db *PostgresDB) CreatePushSubscription(ctx context.Context, ps *PushSubscription) (*PushSubscription, error) {
	return db.phase6Repository.CreatePushSubscription(ctx, ps)
}

func (db *PostgresDB) ListPushSubscriptions(ctx context.Context, userID uuid.UUID) ([]*PushSubscription, error) {
	return db.phase6Repository.ListPushSubscriptions(ctx, userID)
}

func (db *PostgresDB) DeletePushSubscription(ctx context.Context, id uuid.UUID) error {
	return db.phase6Repository.DeletePushSubscription(ctx, id)
}

// Phase 6: Notification Preference operations

func (db *PostgresDB) UpsertNotificationPreference(ctx context.Context, pref *NotificationPreference) (*NotificationPreference, error) {
	return db.phase6Repository.UpsertNotificationPreference(ctx, pref)
}

func (db *PostgresDB) ListNotificationPreferences(ctx context.Context, userID uuid.UUID, opts ListNotificationPreferencesOpts) ([]*NotificationPreference, int, error) {
	return db.phase6Repository.ListNotificationPreferences(ctx, userID, opts)
}

func (db *PostgresDB) DeleteNotificationPreference(ctx context.Context, id uuid.UUID) error {
	return db.phase6Repository.DeleteNotificationPreference(ctx, id)
}

// Remaining: Feedback Round operations

func (db *PostgresDB) CreateFeedbackRound(ctx context.Context, fr *FeedbackRound) (*FeedbackRound, error) {
	return db.remainingRepository.CreateFeedbackRound(ctx, fr)
}

func (db *PostgresDB) GetFeedbackRoundByID(ctx context.Context, id uuid.UUID) (*FeedbackRound, error) {
	return db.remainingRepository.GetFeedbackRoundByID(ctx, id)
}

func (db *PostgresDB) ListFeedbackRounds(ctx context.Context, tenantID uuid.UUID, opts ListFeedbackRoundsOpts) ([]*FeedbackRound, int, error) {
	return db.remainingRepository.ListFeedbackRounds(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdateFeedbackRound(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.remainingRepository.UpdateFeedbackRound(ctx, id, updates)
}

func (db *PostgresDB) CreateFeedbackRoundAssignment(ctx context.Context, a *FeedbackRoundAssignment) (*FeedbackRoundAssignment, error) {
	return db.remainingRepository.CreateFeedbackRoundAssignment(ctx, a)
}

func (db *PostgresDB) ListFeedbackRoundAssignments(ctx context.Context, roundID uuid.UUID) ([]*FeedbackRoundAssignment, error) {
	return db.remainingRepository.ListFeedbackRoundAssignments(ctx, roundID)
}

func (db *PostgresDB) UpdateFeedbackRoundAssignment(ctx context.Context, id int64, updates map[string]interface{}) error {
	return db.remainingRepository.UpdateFeedbackRoundAssignment(ctx, id, updates)
}

func (db *PostgresDB) CreateFeedbackRoundResponse(ctx context.Context, r *FeedbackRoundResponse) (*FeedbackRoundResponse, error) {
	return db.remainingRepository.CreateFeedbackRoundResponse(ctx, r)
}

func (db *PostgresDB) ListFeedbackRoundResponses(ctx context.Context, assignmentID int64) ([]*FeedbackRoundResponse, error) {
	return db.remainingRepository.ListFeedbackRoundResponses(ctx, assignmentID)
}

func (db *PostgresDB) GetFeedbackRoundResults(ctx context.Context, roundID uuid.UUID) ([]map[string]interface{}, error) {
	return db.remainingRepository.GetFeedbackRoundResults(ctx, roundID)
}

// Remaining: Goal Cascade operations

func (db *PostgresDB) GetGoalTree(ctx context.Context, tenantID uuid.UUID) ([]*PerformanceGoal, error) {
	return db.remainingRepository.GetGoalTree(ctx, tenantID)
}

func (db *PostgresDB) UpdateGoalCascade(ctx context.Context, id uuid.UUID, parentGoalID *uuid.UUID, goalLevel, cascadeVisibility string) error {
	return db.remainingRepository.UpdateGoalCascade(ctx, id, parentGoalID, goalLevel, cascadeVisibility)
}

// Remaining: Document Signature operations

func (db *PostgresDB) CreateDocumentSignature(ctx context.Context, ds *DocumentSignature) (*DocumentSignature, error) {
	return db.remainingRepository.CreateDocumentSignature(ctx, ds)
}

func (db *PostgresDB) GetDocumentSignatureByID(ctx context.Context, id uuid.UUID) (*DocumentSignature, error) {
	return db.remainingRepository.GetDocumentSignatureByID(ctx, id)
}

func (db *PostgresDB) ListDocumentSignatures(ctx context.Context, documentID uuid.UUID, opts ListDocumentSignaturesOpts) ([]*DocumentSignature, int, error) {
	return db.remainingRepository.ListDocumentSignatures(ctx, documentID, opts)
}

func (db *PostgresDB) UpdateDocumentSignature(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.remainingRepository.UpdateDocumentSignature(ctx, id, updates)
}

// Remaining: Certificate Key operations

func (db *PostgresDB) CreateCertificateKey(ctx context.Context, ck *CertificateKey) (*CertificateKey, error) {
	return db.remainingRepository.CreateCertificateKey(ctx, ck)
}

func (db *PostgresDB) GetCertificateKeysByCertID(ctx context.Context, certificateID uuid.UUID) ([]*CertificateKey, error) {
	return db.remainingRepository.GetCertificateKeysByCertID(ctx, certificateID)
}

// Remaining: Wallet Pass Template operations

func (db *PostgresDB) CreateWalletPassTemplate(ctx context.Context, t *WalletPassTemplate) (*WalletPassTemplate, error) {
	return db.remainingRepository.CreateWalletPassTemplate(ctx, t)
}

func (db *PostgresDB) ListWalletPassTemplates(ctx context.Context, tenantID uuid.UUID, opts ListWalletPassTemplatesOpts) ([]*WalletPassTemplate, int, error) {
	return db.remainingRepository.ListWalletPassTemplates(ctx, tenantID, opts)
}

// Remaining: Org Chart Import operations

func (db *PostgresDB) CreateOrgChartImport(ctx context.Context, imp *OrgChartImport) (*OrgChartImport, error) {
	return db.remainingRepository.CreateOrgChartImport(ctx, imp)
}

func (db *PostgresDB) GetOrgChartImportByID(ctx context.Context, id uuid.UUID) (*OrgChartImport, error) {
	return db.remainingRepository.GetOrgChartImportByID(ctx, id)
}

func (db *PostgresDB) UpdateOrgChartImport(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.remainingRepository.UpdateOrgChartImport(ctx, id, updates)
}

// Remaining: Package Registry operations

func (db *PostgresDB) CreatePackage(ctx context.Context, pkg *PackageRegistry) (*PackageRegistry, error) {
	return db.remainingRepository.CreatePackage(ctx, pkg)
}

func (db *PostgresDB) GetPackageByID(ctx context.Context, id uuid.UUID) (*PackageRegistry, error) {
	return db.remainingRepository.GetPackageByID(ctx, id)
}

func (db *PostgresDB) GetPackageByName(ctx context.Context, tenantID uuid.UUID, name, registryType string) (*PackageRegistry, error) {
	return db.remainingRepository.GetPackageByName(ctx, tenantID, name, registryType)
}

func (db *PostgresDB) ListPackages(ctx context.Context, tenantID uuid.UUID, opts ListPackageRegistryOpts) ([]*PackageRegistry, int, error) {
	return db.remainingRepository.ListPackages(ctx, tenantID, opts)
}

func (db *PostgresDB) UpdatePackage(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	return db.remainingRepository.UpdatePackage(ctx, id, updates)
}

func (db *PostgresDB) CreatePackageVersion(ctx context.Context, v *PackageVersion) (*PackageVersion, error) {
	return db.remainingRepository.CreatePackageVersion(ctx, v)
}

func (db *PostgresDB) ListPackageVersions(ctx context.Context, packageID uuid.UUID, opts ListPackageVersionsOpts) ([]*PackageVersion, int, error) {
	return db.remainingRepository.ListPackageVersions(ctx, packageID, opts)
}
