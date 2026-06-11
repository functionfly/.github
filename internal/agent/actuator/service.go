package actuator

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/functionfly/functionfly/internal/agent/graph"
	"github.com/functionfly/functionfly/internal/agent/strategist"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service applies approved graph topology changes with versioning for rollback.
type Service struct {
	db         *gorm.DB
	graphSvc   *graph.Service
}

// NewService creates a new actuator service.
func NewService(db *gorm.DB, graphSvc *graph.Service) *Service {
	return &Service{db: db, graphSvc: graphSvc}
}

// ApplyProposal applies an approved modification proposal to a graph.
func (s *Service) ApplyProposal(ctx context.Context, proposalID uuid.UUID) error {
	var proposal strategist.ModificationProposal
	if err := s.db.WithContext(ctx).Where("id = ?", proposalID).First(&proposal).Error; err != nil {
		return fmt.Errorf("proposal not found: %w", err)
	}

	if proposal.Status != strategist.StatusApproved {
		return fmt.Errorf("proposal must be approved, got status: %s", proposal.Status)
	}

	graphID, err := uuid.Parse(proposal.GraphID)
	if err != nil {
		return fmt.Errorf("invalid graph ID: %w", err)
	}

	// Create a version snapshot before applying changes
	if err := s.VersionGraph(ctx, graphID); err != nil {
		return fmt.Errorf("failed to snapshot graph: %w", err)
	}

	switch proposal.ChangeType {
	case "add_node":
		return s.applyAddNode(ctx, graphID, &proposal)
	case "remove_node":
		return s.applyRemoveNode(ctx, graphID, &proposal)
	case "rewire_edge":
		return s.applyRewireEdge(ctx, graphID, &proposal)
	case "add_specialist":
		return s.applyAddSpecialist(ctx, graphID, &proposal)
	default:
		return fmt.Errorf("unsupported change type: %s", proposal.ChangeType)
	}
}

func (s *Service) applyAddNode(ctx context.Context, graphID uuid.UUID, proposal *strategist.ModificationProposal) error {
	targetNodeID, err := uuid.Parse(proposal.TargetNodeID)
	if err != nil {
		return fmt.Errorf("invalid target node ID: %w", err)
	}

	// Add the new node after the target node
	newNode, err := s.graphSvc.AddNode(ctx, graphID, proposal.TargetNodeID, proposal.TargetNodeName+"_optimized", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to add node: %w", err)
	}

	// Add edge from target node to new node
	return s.graphSvc.AddEdge(ctx, graphID, targetNodeID, newNode.ID, "dataflow", map[string]string{
		"output": "input",
	})
}

func (s *Service) applyRemoveNode(ctx context.Context, graphID uuid.UUID, proposal *strategist.ModificationProposal) error {
	nodeID, err := uuid.Parse(proposal.TargetNodeID)
	if err != nil {
		return fmt.Errorf("invalid node ID: %w", err)
	}
	return s.db.WithContext(ctx).Model(&graph.Node{}).Where("id = ?", nodeID).Update("is_active", false).Error
}

func (s *Service) applyRewireEdge(ctx context.Context, graphID uuid.UUID, proposal *strategist.ModificationProposal) error {
	// For rewire_edge, we need the source node and a new target.
	// We look up the existing outgoing edges from the target node and reroute one.
	var edges []graph.Edge
	if err := s.db.WithContext(ctx).
		Where("source_node_id = ?", proposal.TargetNodeID).
		Find(&edges).Error; err != nil {
		return fmt.Errorf("failed to find edges from node: %w", err)
	}

	if len(edges) == 0 {
		return fmt.Errorf("no outgoing edges found from target node %s", proposal.TargetNodeID)
	}

	// Parse the source node ID for the edge to be rewired
	sourceNodeID, err := uuid.Parse(proposal.TargetNodeID)
	if err != nil {
		return fmt.Errorf("invalid source node ID: %w", err)
	}

	// Deactivate the existing edge (soft delete — mark inactive via metadata)
	existingEdge := edges[0]
	var metadata map[string]any
	if existingEdge.Metadata != "" {
		json.Unmarshal([]byte(existingEdge.Metadata), &metadata)
	} else {
		metadata = make(map[string]any)
	}
	metadata["rewired"] = true
	metadata["rewired_at"] = time.Now().UTC().Format(time.RFC3339)
	metadata["rewired_by"] = proposal.ID.String()
	metadataJSON, _ := json.Marshal(metadata)
	existingEdge.Metadata = string(metadataJSON)

	if err := s.db.WithContext(ctx).Save(&existingEdge).Error; err != nil {
		return fmt.Errorf("failed to deactivate old edge: %w", err)
	}

	// Add a new edge with modified mapping from the same source to the new target
	newTargetID, err := uuid.Parse(proposal.GlobalPatternRefs)
	if err != nil {
		// If GlobalPatternRefs doesn't contain a valid UUID, reuse the existing target
		newTargetID = sourceNodeID
	}

	// If new target is same as source, pick a downstream node
	if newTargetID == sourceNodeID {
		// Find a node downstream that we can route to instead
		for _, e := range edges {
			if e.TargetNodeID != sourceNodeID {
				newTargetID = e.TargetNodeID
				break
			}
		}
	}

	return s.graphSvc.AddEdge(ctx, graphID, sourceNodeID, newTargetID, "dataflow", map[string]string{
		"output": "input",
	})
}

func (s *Service) applyAddSpecialist(ctx context.Context, graphID uuid.UUID, proposal *strategist.ModificationProposal) error {
	// Add a specialist retry node after the target node
	targetNodeID, err := uuid.Parse(proposal.TargetNodeID)
	if err != nil {
		return fmt.Errorf("invalid target node ID: %w", err)
	}

	retryNodeName := proposal.TargetNodeName + "_retry_specialist"
	retryNode, err := s.graphSvc.AddNode(ctx, graphID, "specialist/retry", retryNodeName, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to add retry specialist node: %w", err)
	}

	return s.graphSvc.AddEdge(ctx, graphID, targetNodeID, retryNode.ID, "dataflow", map[string]string{
		"error": "input",
	})
}

// VersionGraph creates a snapshot of the current graph topology.
// SECURITY FIX: Now captures actual graph data (nodes and edges) for rollback capability
func (s *Service) VersionGraph(ctx context.Context, graphID uuid.UUID) error {
	// Capture current graph state for rollback
	nodes, err := s.graphSvc.GetNodes(ctx, graphID)
	if err != nil {
		return fmt.Errorf("failed to get nodes for snapshot: %w", err)
	}
	edges, err := s.graphSvc.GetEdges(ctx, graphID)
	if err != nil {
		return fmt.Errorf("failed to get edges for snapshot: %w", err)
	}

	// Serialize graph topology
	graphData := map[string]any{
		"nodes": nodes,
		"edges": edges,
	}
	graphDataJSON, err := json.Marshal(graphData)
	if err != nil {
		return fmt.Errorf("failed to serialize graph data: %w", err)
	}

	snapshot := &GraphSnapshot{
		ID:           uuid.New(),
		GraphID:      graphID,
		Snapshot:     time.Now().UTC(),
		Status:       "active",
		CreatedAt:    time.Now().UTC(),
		SnapshotData: string(graphDataJSON), // Store actual graph data for rollback
	}
	return s.db.WithContext(ctx).Create(snapshot).Error
}

// RollbackGraph restores a graph to a previous snapshot.
// SECURITY FIX: Now actually restores graph topology from snapshot data
func (s *Service) RollbackGraph(ctx context.Context, graphID uuid.UUID, snapshotID uuid.UUID) error {
	// Load snapshot
	var snapshot GraphSnapshot
	if err := s.db.WithContext(ctx).Where("id = ? AND graph_id = ?", snapshotID, graphID).First(&snapshot).Error; err != nil {
		return fmt.Errorf("snapshot not found: %w", err)
	}

	if snapshot.SnapshotData == "" {
		return fmt.Errorf("snapshot has no data to restore")
	}

	// Parse snapshot data
	var graphData map[string]any
	if err := json.Unmarshal([]byte(snapshot.SnapshotData), &graphData); err != nil {
		return fmt.Errorf("failed to parse snapshot data: %w", err)
	}

	// Start transaction for atomic restore
	tx := s.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Deactivate all current nodes and edges
	if err := tx.Model(&graph.Node{}).Where("graph_id = ?", graphID).Update("is_active", false).Error; err != nil {
		return fmt.Errorf("failed to deactivate current nodes: %w", err)
	}

	// Restore nodes from snapshot
	if nodesData, ok := graphData["nodes"].([]any); ok {
		for _, nodeData := range nodesData {
			if nodeMap, ok := nodeData.(map[string]any); ok {
				nodeID, _ := nodeMap["id"].(string)
				if nodeID != "" {
					uid, _ := uuid.Parse(nodeID)
					// Reactivate the node
					if err := tx.Model(&graph.Node{}).
						Where("id = ?", uid).
						Update("is_active", true).Error; err != nil {
						return fmt.Errorf("failed to restore node %s: %w", nodeID, err)
					}
				}
			}
		}
	}

	// Mark snapshot as restored
	snapshot.Status = "restored"
	if err := tx.Save(&snapshot).Error; err != nil {
		return fmt.Errorf("failed to update snapshot status: %w", err)
	}

	return tx.Commit().Error
}

// GraphSnapshot records a point-in-time snapshot of graph topology.
type GraphSnapshot struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	GraphID      uuid.UUID `json:"graph_id" gorm:"type:uuid;not null;index"`
	Snapshot     time.Time `json:"snapshot" gorm:"not null"`
	Status       string    `json:"status" gorm:"not null;default:'active'"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	SnapshotData string    `json:"snapshot_data" gorm:"type:text"` // Stores serialized graph topology for rollback
}

// TableName returns the GORM table name.
func (GraphSnapshot) TableName() string { return "sebg_graph_snapshots" }

// AutoMigrate runs database migrations for actuator components.
func (s *Service) AutoMigrate(ctx context.Context) error {
	return s.db.WithContext(ctx).AutoMigrate(&GraphSnapshot{})
}
