//! Execution graph for WASM Fusion

use std::collections::{HashMap, HashSet};
use serde::{Deserialize, Serialize};

use super::{FusionNode, FusionEdge};

/// A node ID in an execution graph
#[derive(Debug, Clone, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct NodeId(pub String);

impl NodeId {
    pub fn new(id: impl Into<String>) -> Self {
        Self(id.into())
    }
}

impl std::fmt::Display for NodeId {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

/// Result of executing a node
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct NodeResult {
    pub node_id: String,
    pub output: Option<Vec<u8>>,
    pub error: Option<String>,
    pub duration_ms: u64,
}

impl NodeResult {
    pub fn success(node_id: &str, output: Vec<u8>, duration_ms: u64) -> Self {
        Self {
            node_id: node_id.to_string(),
            output: Some(output),
            error: None,
            duration_ms,
        }
    }

    pub fn failure(node_id: &str, error: String, duration_ms: u64) -> Self {
        Self {
            node_id: node_id.to_string(),
            output: None,
            error: Some(error),
            duration_ms,
        }
    }
}

/// A compiled execution graph ready for execution
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ExecutionGraph {
    pub graph_id: String,
    nodes: HashMap<NodeId, FusionNode>,
    edges: Vec<FusionEdge>,
    adjacency: HashMap<NodeId, Vec<NodeId>>,
    execution_order: Vec<NodeId>,
}

impl ExecutionGraph {
    pub fn new(graph_id: &str) -> Self {
        Self {
            graph_id: graph_id.to_string(),
            nodes: HashMap::new(),
            edges: Vec::new(),
            adjacency: HashMap::new(),
            execution_order: Vec::new(),
        }
    }

    /// Add a node to the graph
    pub fn add_node(&mut self, node: FusionNode) {
        let node_id = NodeId::new(&node.node_id);
        self.nodes.insert(node_id.clone(), node);
    }

    /// Add an edge to the graph
    pub fn add_edge(&mut self, edge: FusionEdge) {
        // Build adjacency list
        let source = NodeId::new(&edge.source);
        let target = NodeId::new(&edge.target);

        self.adjacency.entry(source).or_default().push(target.clone());
        self.edges.push(edge);

        // Topological sort for execution order
        self.execution_order = self.topological_sort();
    }

    /// Get a node by ID
    pub fn get_node(&self, node_id: &NodeId) -> Option<&FusionNode> {
        self.nodes.get(node_id)
    }

    /// Get all nodes
    pub fn nodes(&self) -> &HashMap<NodeId, FusionNode> {
        &self.nodes
    }

    /// Get execution order (topologically sorted)
    pub fn execution_order(&self) -> &[NodeId] {
        &self.execution_order
    }

    /// Get all edges
    pub fn edges(&self) -> &[FusionEdge] {
        &self.edges
    }

    /// Perform topological sort to determine execution order
    fn topological_sort(&self) -> Vec<NodeId> {
        let mut in_degree: HashMap<NodeId, usize> = self.nodes.keys()
            .map(|id| (id.clone(), 0))
            .collect();

        // Calculate in-degrees
        for edge in &self.edges {
            let target = NodeId::new(&edge.target);
            if let Some(deg) = in_degree.get_mut(&target) {
                *deg += 1;
            }
        }

        // Start with nodes that have no dependencies
        let mut queue: Vec<NodeId> = in_degree.iter()
            .filter(|(_, &deg)| deg == 0)
            .map(|(id, _)| id.clone())
            .collect();

        let mut result = Vec::new();

        while let Some(node_id) = queue.pop() {
            result.push(node_id.clone());

            if let Some(neighbors) = self.adjacency.get(&node_id) {
                for neighbor in neighbors {
                    if let Some(deg) = in_degree.get_mut(neighbor) {
                        *deg -= 1;
                        if *deg == 0 {
                            queue.push(neighbor.clone());
                        }
                    }
                }
            }
        }

        result
    }

    /// Get nodes that have no dependencies (entry points)
    pub fn entry_nodes(&self) -> Vec<&FusionNode> {
        self.nodes.values()
            .filter(|node| {
                !self.edges.iter().any(|e| e.target == node.node_id)
            })
            .collect()
    }

    /// Get nodes that nothing depends on (exit points)
    pub fn exit_nodes(&self) -> Vec<&FusionNode> {
        let has_dependents: HashSet<String> = self.edges.iter()
            .map(|e| e.source.clone())
            .collect();

        self.nodes.values()
            .filter(|node| !has_dependents.contains(&node.node_id))
            .collect()
    }
}