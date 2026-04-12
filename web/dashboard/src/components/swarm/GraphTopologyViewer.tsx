import { useEffect, useRef, useState, useCallback } from 'react';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Slider } from '@/components/ui/slider';
import { 
  ZoomIn, 
  ZoomOut, 
  RotateCcw, 
  Network,
  Circle,
  Triangle,
  Square,
  Loader2
} from 'lucide-react';
import { agentApi } from '@/api/agent';

interface AgentNode {
  id: string;
  name: string;
  swarmRole: 'worker' | 'manager' | 'infrastructure' | 'independent';
  status: 'active' | 'suspended' | 'pending';
  parentId?: string;
  children: string[];
  x?: number;
  y?: number;
  level: number;
}

interface AgentEdge {
  from: string;
  to: string;
  type: 'parent' | 'message';
}

interface GraphTopologyViewerProps {
  rootAgentId: string;
  height?: number;
}

const roleIcons = {
  worker: Circle,
  manager: Triangle,
  infrastructure: Square,
  independent: Circle,
};

const roleColors = {
  worker: 'bg-blue-500',
  manager: 'bg-purple-500',
  infrastructure: 'bg-orange-500',
  independent: 'bg-gray-500',
};

const statusColors = {
  active: 'border-green-500',
  suspended: 'border-red-500',
  pending: 'border-yellow-500',
};

export function GraphTopologyViewer({ rootAgentId, height = 600 }: GraphTopologyViewerProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [nodes, setNodes] = useState<Record<string, AgentNode>>({});
  const [edges, setEdges] = useState<AgentEdge[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [scale, setScale] = useState(1);
  const [offset, setOffset] = useState({ x: 0, y: 0 });
  const [selectedNode, setSelectedNode] = useState<string | null>(null);
  const [dragging, setDragging] = useState(false);
  const dragStart = useRef({ x: 0, y: 0 });

  // Load swarm topology
  const loadTopology = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const nodeMap: Record<string, AgentNode> = {};
      const edgeList: AgentEdge[] = [];

      // Build topology recursively
      const buildTopology = async (agentId: string, level: number, parentId?: string) => {
        if (nodeMap[agentId]) return; // Already processed

        try {
          const { agent } = await agentApi.getAgent(agentId);
          
          nodeMap[agentId] = {
            id: agentId,
            name: agent.name,
            swarmRole: (agent.swarmRole as AgentNode['swarmRole']) || 'independent',
            status: (agent.status as AgentNode['status']) || 'active',
            parentId,
            children: [],
            level,
          };

          if (parentId) {
            nodeMap[parentId].children.push(agentId);
            edgeList.push({ from: parentId, to: agentId, type: 'parent' });
          }

          // Get children
          const { children } = await agentApi.getChildren(agentId);
          for (const child of children) {
            await buildTopology(child.agentId, level + 1, agentId);
          }
        } catch (err) {
          console.warn(`Failed to load agent ${agentId}:`, err);
        }
      };

      await buildTopology(rootAgentId, 0);

      // Calculate positions using tree layout
      const calculatePositions = () => {
        const levelNodes: Record<number, string[]> = {};
        
        Object.values(nodeMap).forEach(node => {
          if (!levelNodes[node.level]) levelNodes[node.level] = [];
          levelNodes[node.level].push(node.id);
        });

        const maxLevel = Math.max(...Object.keys(levelNodes).map(Number));
        const canvasWidth = canvasRef.current?.width || 800;
        const canvasHeight = canvasRef.current?.height || 600;

        Object.entries(levelNodes).forEach(([level, ids]) => {
          const y = (canvasHeight / (maxLevel + 2)) * (Number(level) + 1);
          const xStep = canvasWidth / (ids.length + 1);
          
          ids.forEach((id, index) => {
            nodeMap[id].x = xStep * (index + 1);
            nodeMap[id].y = y;
          });
        });
      };

      calculatePositions();

      setNodes(nodeMap);
      setEdges(edgeList);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load topology');
    } finally {
      setLoading(false);
    }
  }, [rootAgentId]);

  useEffect(() => {
    loadTopology();
  }, [loadTopology]);

  // Draw canvas
  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas || Object.keys(nodes).length === 0) return;

    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    // Clear canvas
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    // Apply transformations
    ctx.save();
    ctx.translate(offset.x, offset.y);
    ctx.scale(scale, scale);

    // Draw edges
    edges.forEach(edge => {
      const from = nodes[edge.from];
      const to = nodes[edge.to];
      if (!from?.x || !from?.y || !to?.x || !to?.y) return;

      ctx.beginPath();
      ctx.moveTo(from.x, from.y);
      
      // Curved lines for parent relationships
      if (edge.type === 'parent') {
        const midY = (from.y + to.y) / 2;
        ctx.bezierCurveTo(from.x, midY, to.x, midY, to.x, to.y);
        ctx.strokeStyle = selectedNode === edge.from || selectedNode === edge.to ? '#6366f1' : '#94a3b8';
        ctx.lineWidth = selectedNode === edge.from || selectedNode === edge.to ? 3 : 1.5;
      } else {
        ctx.lineTo(to.x, to.y);
        ctx.strokeStyle = '#cbd5e1';
        ctx.setLineDash([5, 5]);
      }
      
      ctx.stroke();
      ctx.setLineDash([]);
    });

    // Draw nodes
    Object.values(nodes).forEach(node => {
      if (!node.x || !node.y) return;

      const isSelected = selectedNode === node.id;
      const radius = isSelected ? 35 : 30;

      // Node background
      ctx.beginPath();
      ctx.arc(node.x, node.y, radius, 0, Math.PI * 2);
      ctx.fillStyle = isSelected ? '#e0e7ff' : '#ffffff';
      ctx.fill();

      // Node border based on status
      const borderColor = {
        active: '#22c55e',
        suspended: '#ef4444',
        pending: '#eab308',
      }[node.status];
      
      ctx.beginPath();
      ctx.arc(node.x, node.y, radius, 0, Math.PI * 2);
      ctx.strokeStyle = borderColor;
      ctx.lineWidth = isSelected ? 4 : 2;
      ctx.stroke();

      // Role indicator
      const roleColor = {
        worker: '#3b82f6',
        manager: '#a855f7',
        infrastructure: '#f97316',
        independent: '#6b7280',
      }[node.swarmRole];

      ctx.beginPath();
      ctx.arc(node.x - radius * 0.6, node.y - radius * 0.6, 8, 0, Math.PI * 2);
      ctx.fillStyle = roleColor;
      ctx.fill();

      // Node name
      ctx.font = isSelected ? 'bold 12px sans-serif' : '11px sans-serif';
      ctx.fillStyle = '#1e293b';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      
      // Truncate name if too long
      let displayName = node.name;
      if (displayName.length > 12) {
        displayName = displayName.slice(0, 10) + '...';
      }
      
      ctx.fillText(displayName, node.x, node.y + 5);

      // Agent ID (small)
      ctx.font = '9px sans-serif';
      ctx.fillStyle = '#64748b';
      ctx.fillText(node.id.slice(0, 8), node.x, node.y + 18);
    });

    ctx.restore();
  }, [nodes, edges, scale, offset, selectedNode]);

  // Mouse handlers
  const handleMouseDown = (e: React.MouseEvent) => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const rect = canvas.getBoundingClientRect();
    const x = (e.clientX - rect.left - offset.x) / scale;
    const y = (e.clientY - rect.top - offset.y) / scale;

    // Check if clicked on a node
    let clickedNode: string | null = null;
    Object.values(nodes).forEach(node => {
      if (!node.x || !node.y) return;
      const dx = x - node.x;
      const dy = y - node.y;
      if (Math.sqrt(dx * dx + dy * dy) < 30) {
        clickedNode = node.id;
      }
    });

    if (clickedNode) {
      setSelectedNode(clickedNode);
    } else {
      setDragging(true);
      dragStart.current = { x: e.clientX - offset.x, y: e.clientY - offset.y };
    }
  };

  const handleMouseMove = (e: React.MouseEvent) => {
    if (!dragging) return;
    
    setOffset({
      x: e.clientX - dragStart.current.x,
      y: e.clientY - dragStart.current.y,
    });
  };

  const handleMouseUp = () => {
    setDragging(false);
  };

  const handleWheel = (e: React.WheelEvent) => {
    e.preventDefault();
    const delta = e.deltaY > 0 ? 0.9 : 1.1;
    setScale(s => Math.max(0.5, Math.min(2, s * delta)));
  };

  const selectedNodeData = selectedNode ? nodes[selectedNode] : null;

  if (loading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="py-8 text-center">
          <Network className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <p className="text-muted-foreground">{error}</p>
          <Button onClick={loadTopology} variant="outline" className="mt-4">
            Retry
          </Button>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="overflow-hidden">
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <Network className="h-5 w-5" />
              Swarm Topology
            </CardTitle>
            <CardDescription>
              {Object.keys(nodes).length} agents · {edges.length} connections
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="icon" onClick={() => setScale(s => Math.max(0.5, s - 0.1))}>
              <ZoomOut className="h-4 w-4" />
            </Button>
            <span className="text-xs text-muted-foreground w-12 text-center">
              {Math.round(scale * 100)}%
            </span>
            <Button variant="outline" size="icon" onClick={() => setScale(s => Math.min(2, s + 0.1))}>
              <ZoomIn className="h-4 w-4" />
            </Button>
            <Button variant="outline" size="icon" onClick={() => { setScale(1); setOffset({ x: 0, y: 0 }); }}>
              <RotateCcw className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </CardHeader>
      
      <div className="relative">
        <canvas
          ref={canvasRef}
          width={800}
          height={height}
          className="w-full cursor-grab active:cursor-grabbing"
          onMouseDown={handleMouseDown}
          onMouseMove={handleMouseMove}
          onMouseUp={handleMouseUp}
          onMouseLeave={handleMouseUp}
          onWheel={handleWheel}
          style={{ height }}
        />

        {/* Legend */}
        <div className="absolute bottom-4 left-4 bg-background/90 backdrop-blur rounded-lg p-3 border shadow-sm">
          <p className="text-xs font-medium mb-2">Roles</p>
          <div className="space-y-1">
            {Object.entries(roleColors).map(([role, color]) => {
              const Icon = roleIcons[role as keyof typeof roleIcons];
              return (
                <div key={role} className="flex items-center gap-2 text-xs">
                  <div className={`w-2 h-2 rounded-full ${color}`} />
                  <span className="capitalize">{role}</span>
                </div>
              );
            })}
          </div>
        </div>

        {/* Selected node details */}
        {selectedNodeData && (
          <div className="absolute top-4 right-4 w-64 bg-background/95 backdrop-blur rounded-lg p-4 border shadow-lg">
            <div className="flex items-center justify-between mb-2">
              <h4 className="font-semibold text-sm">{selectedNodeData.name}</h4>
              <Badge variant={selectedNodeData.status === 'active' ? 'default' : 'destructive'}>
                {selectedNodeData.status}
              </Badge>
            </div>
            <div className="space-y-2 text-xs text-muted-foreground">
              <p><span className="font-medium">ID:</span> {selectedNodeData.id}</p>
              <p><span className="font-medium">Role:</span> <span className="capitalize">{selectedNodeData.swarmRole}</span></p>
              <p><span className="font-medium">Level:</span> {selectedNodeData.level}</p>
              <p><span className="font-medium">Children:</span> {selectedNodeData.children.length}</p>
              {selectedNodeData.parentId && (
                <p><span className="font-medium">Parent:</span> {selectedNodeData.parentId.slice(0, 8)}...</p>
              )}
            </div>
            <div className="mt-3 flex gap-2">
              <Button 
                size="sm" 
                variant="outline" 
                className="flex-1 text-xs"
                onClick={() => window.location.href = `/agents/${selectedNodeData.id}`}
              >
                View Details
              </Button>
            </div>
          </div>
        )}
      </div>
    </Card>
  );
}

export default GraphTopologyViewer;
