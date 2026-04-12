import { useMemo, useState, useCallback } from 'react';
import { Graph } from '@visx/network';
import type { Link, DefaultNode } from '@visx/network/lib/types';
import { Group } from '@visx/group';
import { Text } from '@visx/text';
import { scaleOrdinal } from '@visx/scale';
import { Zoom } from '@visx/zoom';
import { localPoint } from '@visx/event';
import type { AgentMemory, AgentMemoryType } from '@/types';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

interface MemoryNode extends Omit<DefaultNode, 'x' | 'y'> {
  id: string;
  type: AgentMemoryType;
  importance: number;
  accessCount: number;
  label: string;
  x: number;
  y: number;
}

interface MemoryLink extends Link<MemoryNode> {
  source: MemoryNode;
  target: MemoryNode;
  strength: number;
}

interface AgentMemoryGraphProps {
  memories: AgentMemory[];
  agentId: string;
}

const memoryTypeLabels: Record<AgentMemoryType, string> = {
  working: 'Working',
  longterm: 'Long-term',
  context: 'Context',
  episodic: 'Episodic',
};

function generateNodesAndLinks(memories: AgentMemory[]): { nodes: MemoryNode[]; links: MemoryLink[] } {
  const nodes: MemoryNode[] = memories.map((memory, index) => {
    const angle = (2 * Math.PI * index) / memories.length;
    const radius = 200;
    return {
      id: memory.id,
      type: memory.memory_type,
      importance: memory.importance_score,
      accessCount: memory.access_count,
      label: memory.content.slice(0, 50) + (memory.content.length > 50 ? '...' : ''),
      x: 300 + radius * Math.cos(angle),
      y: 250 + radius * Math.sin(angle),
    };
  });

  const links: MemoryLink[] = [];

  for (let i = 0; i < nodes.length; i++) {
    for (let j = i + 1; j < nodes.length; j++) {
      const sourceMem = memories[i];
      const targetMem = memories[j];

      if (sourceMem.memory_type === targetMem.memory_type) {
        links.push({
          source: nodes[i],
          target: nodes[j],
          strength: 0.5,
        });
      }

      if (sourceMem.importance_score > 0.8 || targetMem.importance_score > 0.8) {
        const existingLink = links.find(
          (l) => l.source.id === nodes[i].id && l.target.id === nodes[j].id
        );
        if (!existingLink) {
          links.push({
            source: nodes[i],
            target: nodes[j],
            strength: Math.max(sourceMem.importance_score, targetMem.importance_score),
          });
        }
      }
    }
  }

  return { nodes, links };
}

interface NodeComponentProps {
  node: MemoryNode;
  colorScale: (type: AgentMemoryType) => string;
  onClick: (node: MemoryNode, event: React.MouseEvent) => void;
  isSelected: boolean;
}

function NodeComponent({ node, colorScale, onClick, isSelected }: NodeComponentProps) {
  const size = Math.max(20, node.importance * 40 + node.accessCount * 2);
  const color = colorScale(node.type);

  const handleClick = useCallback((event: React.MouseEvent) => {
    onClick(node, event);
  }, [node, onClick]);

  return (
    <Group onClick={handleClick} style={{ cursor: 'pointer' }}>
      <circle
        cx={node.x ?? 0}
        cy={node.y ?? 0}
        r={size}
        fill={color}
        opacity={isSelected ? 1 : 0.8}
        stroke={isSelected ? '#fff' : 'none'}
        strokeWidth={isSelected ? 3 : 0}
      />
      <Text
        x={(node.x ?? 0) + size + 5}
        y={(node.y ?? 0)}
        fontSize={12}
        fill="currentColor"
        className="text-text-primary"
        verticalAnchor="middle"
      >
        {memoryTypeLabels[node.type]}
      </Text>
    </Group>
  );
}

function LinkComponent({ link }: { link: MemoryLink }) {
  const strokeWidth = Math.max(1, link.strength * 3);
  const opacity = Math.max(0.2, link.strength);

  return (
    <line
      x1={link.source.x ?? 0}
      y1={link.source.y ?? 0}
      x2={link.target.x ?? 0}
      y2={link.target.y ?? 0}
      stroke="#64748b"
      strokeWidth={strokeWidth}
      opacity={opacity}
    />
  );
}

export function AgentMemoryGraph({ memories, agentId }: AgentMemoryGraphProps) {
  const [selectedNode, setSelectedNode] = useState<MemoryNode | null>(null);

  const { nodes, links } = useMemo(() => generateNodesAndLinks(memories), [memories]);

  const colorScale = useMemo(
    () =>
      scaleOrdinal<AgentMemoryType, string>({
        domain: ['working', 'longterm', 'context', 'episodic'],
        range: ['#3b82f6', '#22c55e', '#a855f7', '#f97316'],
      }),
    []
  );

  const selectedMemory = useMemo(() => {
    if (!selectedNode) return null;
    return memories.find((m) => m.id === selectedNode.id) ?? null;
  }, [selectedNode, memories]);

  const stats = useMemo(() => {
    const byType = memories.reduce((acc, mem) => {
      acc[mem.memory_type] = (acc[mem.memory_type] ?? 0) + 1;
      return acc;
    }, {} as Record<AgentMemoryType, number>);

    const avgImportance =
      memories.length > 0
        ? memories.reduce((sum, m) => sum + m.importance_score, 0) / memories.length
        : 0;

    const totalAccess = memories.reduce((sum, m) => sum + m.access_count, 0);

    return { byType, avgImportance, totalAccess };
  }, [memories]);

  const handleNodeClick = useCallback((node: MemoryNode, event: React.MouseEvent) => {
    event.stopPropagation();
    setSelectedNode(node);
  }, []);

  const width = 800;
  const height = 500;

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap gap-4 items-center justify-between">
        <div className="flex flex-wrap gap-3">
          {(Object.keys(memoryTypeLabels) as AgentMemoryType[]).map((type) => (
            <div key={type} className="flex items-center gap-2">
              <div
                className="w-3 h-3 rounded-full"
                style={{ backgroundColor: colorScale(type) }}
              />
              <span className="text-sm text-muted-foreground">
                {memoryTypeLabels[type]} ({stats.byType[type] ?? 0})
              </span>
            </div>
          ))}
        </div>
        <div className="flex gap-4 text-sm text-muted-foreground">
          <span>Avg Importance: {stats.avgImportance.toFixed(2)}</span>
          <span>Total Access: {stats.totalAccess}</span>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-[1fr,350px]">
        <Card className="overflow-hidden">
          <Zoom
            width={width}
            height={height}
            scaleXMin={0.5}
            scaleXMax={3}
            scaleYMin={0.5}
            scaleYMax={3}
            initialTransformMatrix={{
              scaleX: 1,
              scaleY: 1,
              translateX: 0,
              translateY: 0,
              skewX: 0,
              skewY: 0,
            }}
          >
            {(zoom) => (
              <svg
                width="100%"
                height={height}
                viewBox={`0 0 ${width} ${height}`}
                className="bg-bg-secondary cursor-grab active:cursor-grabbing"
                onMouseDown={zoom.dragStart}
                onMouseMove={zoom.dragMove}
                onMouseUp={zoom.dragEnd}
                onMouseLeave={() => {
                  if (zoom.isDragging) zoom.dragEnd();
                }}
                onClick={(event) => {
                  const point = localPoint(event);
                  if (point && !selectedNode) {
                    zoom.scale({ scaleX: 1.2, scaleY: 1.2, point });
                  }
                }}
                onDoubleClick={(event) => {
                  const point = localPoint(event);
                  if (point) {
                    zoom.scale({ scaleX: 1.5, scaleY: 1.5, point });
                  }
                }}
              >
                <rect
                  width={width}
                  height={height}
                  fill="transparent"
                  onClick={() => setSelectedNode(null)}
                />
                <g transform={zoom.toString()}>
                  {links.map((link, i) => (
                    <LinkComponent key={i} link={link} />
                  ))}

                  {nodes.map((node) => (
                    <NodeComponent
                      key={node.id}
                      node={node}
                      colorScale={colorScale}
                      onClick={handleNodeClick}
                      isSelected={selectedNode?.id === node.id}
                    />
                  ))}
                </g>

                <g transform={`translate(${width - 120}, ${height - 80})`}>
                  <foreignObject width={110} height={70}>
                    <div className="flex flex-col gap-1 bg-bg-primary/90 p-2 rounded shadow-sm">
                      <button
                        onClick={() => zoom.scale({ scaleX: 1.2, scaleY: 1.2 })}
                        className="text-xs px-2 py-1 bg-bg-secondary rounded hover:bg-bg-tertiary"
                      >
                        Zoom In
                      </button>
                      <button
                        onClick={() => zoom.scale({ scaleX: 0.8, scaleY: 0.8 })}
                        className="text-xs px-2 py-1 bg-bg-secondary rounded hover:bg-bg-tertiary"
                      >
                        Zoom Out
                      </button>
                      <button
                        onClick={zoom.reset}
                        className="text-xs px-2 py-1 bg-bg-secondary rounded hover:bg-bg-tertiary"
                      >
                        Reset
                      </button>
                    </div>
                  </foreignObject>
                </g>
              </svg>
            )}
          </Zoom>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-base">
              {selectedMemory ? 'Memory Details' : 'Memory Network'}
            </CardTitle>
          </CardHeader>
          <CardContent>
            {selectedMemory ? (
              <div className="space-y-4">
                <div className="flex items-center gap-2">
                  <Badge
                    style={{
                      backgroundColor: colorScale(selectedMemory.memory_type) + '20',
                      borderColor: colorScale(selectedMemory.memory_type),
                      color: colorScale(selectedMemory.memory_type),
                    }}
                    variant="outline"
                  >
                    {memoryTypeLabels[selectedMemory.memory_type]}
                  </Badge>
                </div>

                <div className="space-y-2">
                  <p className="text-xs font-medium text-muted-foreground uppercase">Content</p>
                  <p className="text-sm">{selectedMemory.content}</p>
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-1">
                    <p className="text-xs text-muted-foreground">Importance</p>
                    <p className="font-semibold">{selectedMemory.importance_score.toFixed(2)}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs text-muted-foreground">Access Count</p>
                    <p className="font-semibold">{selectedMemory.access_count}</p>
                  </div>
                  <div className="space-y-1">
                    <p className="text-xs text-muted-foreground">Created</p>
                    <p className="font-semibold text-xs">
                      {new Date(selectedMemory.created_at).toLocaleDateString()}
                    </p>
                  </div>
                  {selectedMemory.expires_at && (
                    <div className="space-y-1">
                      <p className="text-xs text-muted-foreground">Expires</p>
                      <p className="font-semibold text-xs">
                        {new Date(selectedMemory.expires_at).toLocaleDateString()}
                      </p>
                    </div>
                  )}
                </div>

                <button
                  onClick={() => setSelectedNode(null)}
                  className="text-sm text-brand-500 hover:underline"
                >
                  Clear selection
                </button>
              </div>
            ) : (
              <div className="text-center py-8 text-muted-foreground">
                <p className="text-sm">Click on a node to view memory details</p>
                <p className="text-xs mt-2">
                  Node size represents importance and access frequency
                </p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
