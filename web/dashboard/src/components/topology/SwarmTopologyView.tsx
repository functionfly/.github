import { useState, useCallback, useEffect, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { Group } from '@visx/group';
import { Zoom } from '@visx/zoom';
import { localPoint } from '@visx/event';
import { Text } from '@visx/text';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { agentApi } from '@/api/agent';
import type { AgentIdentity } from '@/api/agent';
import {
  ZoomIn,
  ZoomOut,
  RotateCcw,
  Network,
  Loader2,
  Circle,
  Triangle,
  Square,
  Users,
  Shield,
  Server,
} from 'lucide-react';

interface TopologyNode {
  id: string;
  name: string;
  status: 'active' | 'suspended' | 'pending';
  swarmRole: 'worker' | 'manager' | 'infrastructure';
  trustScore: number;
  agentId: string;
  x: number;
  y: number;
}

interface TopologyLink {
  source: TopologyNode;
  target: TopologyNode;
}

interface SwarmTopologyViewProps {
  agentId: string;
  className?: string;
}

const roleIcons = {
  worker: Circle,
  manager: Triangle,
  infrastructure: Square,
};

const roleColors = {
  worker: '#3b82f6',
  manager: '#a855f7',
  infrastructure: '#f97316',
};

const statusColors = {
  active: '#22c55e',
  suspended: '#ef4444',
  pending: '#eab308',
};

const roleBadgeColors = {
  worker: 'bg-blue-500/20 text-blue-500 border-blue-500/30',
  manager: 'bg-purple-500/20 text-purple-500 border-purple-500/30',
  infrastructure: 'bg-orange-500/20 text-orange-500 border-orange-500/30',
};

function NodeCircle({
  node,
  isSelected,
  onClick,
}: {
  node: TopologyNode;
  isSelected: boolean;
  onClick: (node: TopologyNode, event: React.MouseEvent) => void;
}) {
  const RoleIcon = roleIcons[node.swarmRole] || Circle;
  const nodeSize = isSelected ? 40 : 32;

  return (
    <Group
      onClick={(event) => onClick(node, event)}
      style={{ cursor: 'pointer' }}
    >
      <circle
        cx={node.x}
        cy={node.y}
        r={nodeSize + 8}
        fill={isSelected ? 'rgba(99, 102, 241, 0.2)' : 'transparent'}
        stroke={isSelected ? '#6366f1' : 'transparent'}
        strokeWidth={2}
      />

      <circle
        cx={node.x}
        cy={node.y}
        r={nodeSize}
        fill="#ffffff"
        stroke={statusColors[node.status]}
        strokeWidth={3}
        filter="url(#shadow)"
      />

      <circle
        cx={node.x - nodeSize * 0.5}
        cy={node.y - nodeSize * 0.5}
        r={8}
        fill={roleColors[node.swarmRole]}
      />

      <foreignObject
        x={node.x - nodeSize}
        y={node.y - nodeSize * 0.3}
        width={nodeSize * 2}
        height={nodeSize * 2}
        style={{ pointerEvents: 'none' }}
      >
        <div className="flex flex-col items-center justify-center h-full">
          <RoleIcon
            className="h-5 w-5 mb-1"
            style={{ color: roleColors[node.swarmRole] }}
          />
        </div>
      </foreignObject>

      <Text
        x={node.x}
        y={node.y + nodeSize + 12}
        textAnchor="middle"
        fontSize={11}
        fill="#1e293b"
        fontWeight={isSelected ? 'bold' : 'normal'}
      >
        {node.name.length > 14 ? node.name.slice(0, 12) + '...' : node.name}
      </Text>

      <Text
        x={node.x}
        y={node.y + nodeSize + 26}
        textAnchor="middle"
        fontSize={9}
        fill="#64748b"
      >
        {node.trustScore.toFixed(2)}
      </Text>
    </Group>
  );
}

function LinkLine({ link }: { link: TopologyLink }) {
  return (
    <line
      x1={link.source.x}
      y1={link.source.y}
      x2={link.target.x}
      y2={link.target.y}
      stroke="#64748b"
      strokeWidth={1.5}
      strokeOpacity={0.6}
    />
  );
}

export function SwarmTopologyView({ agentId, className }: SwarmTopologyViewProps) {
  const { t } = useTranslation();
  const [rootAgent, setRootAgent] = useState<AgentIdentity | null>(null);
  const [nodes, setNodes] = useState<TopologyNode[]>([]);
  const [links, setLinks] = useState<TopologyLink[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedNode, setSelectedNode] = useState<TopologyNode | null>(null);

  const width = 800;
  const height = 500;

  const loadTopology = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);

      const { agent } = await agentApi.getAgent(agentId);
      setRootAgent(agent);

      const { children } = await agentApi.getChildren(agentId);

      const allNodes: TopologyNode[] = [];
      const allLinks: TopologyLink[] = [];

      const rootNode: TopologyNode = {
        id: agent.id,
        name: agent.name,
        status: (agent.status as TopologyNode['status']) || 'active',
        swarmRole: (agent.swarmRole as TopologyNode['swarmRole']) || 'worker',
        trustScore: 0.85,
        agentId: agent.agentId,
        x: width / 2,
        y: 80,
      };
      allNodes.push(rootNode);

      const childNodes: TopologyNode[] = children.map((child, index) => {
        const angle = (2 * Math.PI * index) / Math.max(children.length, 1);
        const radius = 200;
        const level = 280;

        return {
          id: child.id,
          name: child.name,
          status: child.status,
          swarmRole: child.swarmRole,
          trustScore: child.trustScore,
          agentId: child.id,
          x: width / 2 + radius * Math.cos(angle - Math.PI / 2),
          y: 80 + level,
        };
      });

      allNodes.push(...childNodes);

      for (const child of childNodes) {
        allLinks.push({
          source: rootNode,
          target: child,
        });
      }

      setNodes(allNodes);
      setLinks(allLinks);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load topology');
    } finally {
      setLoading(false);
    }
  }, [agentId]);

  useEffect(() => {
    loadTopology();
  }, [loadTopology]);

  const handleNodeClick = useCallback((node: TopologyNode, event: React.MouseEvent) => {
    event.stopPropagation();
    setSelectedNode(node);
  }, []);

  const handleBackgroundClick = useCallback(() => {
    setSelectedNode(null);
  }, []);

  const initialTransform = useMemo(
    () => ({
      scaleX: 1,
      scaleY: 1,
      translateX: 0,
      translateY: 0,
      skewX: 0,
      skewY: 0,
    }),
    []
  );

  if (loading) {
    return (
      <Card className={className}>
        <CardContent className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card className={className}>
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
    <Card className={className}>
      <CardHeader className="pb-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-brand-500/10 rounded-lg">
              <Network className="h-5 w-5 text-brand-500" />
            </div>
            <div>
              <CardTitle className="flex items-center gap-2">
                {t('swarm.topology', 'Swarm Topology')}
              </CardTitle>
              <p className="text-sm text-muted-foreground">
                {nodes.length} {t('swarm.agents', 'agents')} · {links.length}{' '}
                {t('swarm.connections', 'connections')}
              </p>
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent className="pt-0">
        <div className="grid gap-6 lg:grid-cols-[1fr,300px]">
          <div className="relative">
            <Zoom<SVGSVGElement>
              width={width}
              height={height}
              scaleXMin={0.5}
              scaleXMax={3}
              scaleYMin={0.5}
              scaleYMax={3}
              initialTransformMatrix={initialTransform}
            >
              {(zoom) => (
                <svg
                  width="100%"
                  height={height}
                  viewBox={`0 0 ${width} ${height}`}
                  className="bg-bg-secondary rounded-lg cursor-grab active:cursor-grabbing"
                  onMouseDown={zoom.dragStart}
                  onMouseMove={zoom.dragMove}
                  onMouseUp={zoom.dragEnd}
                  onMouseLeave={() => {
                    if (zoom.isDragging) zoom.dragEnd();
                  }}
                  onClick={handleBackgroundClick}
                  onDoubleClick={(event) => {
                    const point = localPoint(event);
                    if (point) {
                      zoom.scale({ scaleX: 1.5, scaleY: 1.5, point });
                    }
                  }}
                >
                  <rect width={width} height={height} fill="transparent" />
                  <g transform={zoom.toString()}>
                    {links.map((link, i) => (
                      <LinkLine key={i} link={link} />
                    ))}

                    {nodes.map((node) => (
                      <NodeCircle
                        key={node.id}
                        node={node}
                        isSelected={selectedNode?.id === node.id}
                        onClick={handleNodeClick}
                      />
                    ))}
                  </g>

                  <defs>
                    <filter id="shadow" x="-20%" y="-20%" width="140%" height="140%">
                      <feDropShadow dx="0" dy="2" stdDeviation="3" floodOpacity="0.15" />
                    </filter>
                  </defs>

                  <g transform={`translate(${width - 100}, ${height - 80})`}>
                    <foreignObject width={90} height={70}>
                      <div className="flex flex-col gap-1 bg-bg-primary/90 p-2 rounded-lg shadow-lg border">
                        <button
                          onClick={() => zoom.scale({ scaleX: 1.2, scaleY: 1.2 })}
                          className="text-xs px-2 py-1 bg-bg-secondary rounded hover:bg-bg-tertiary transition-colors flex items-center gap-1"
                        >
                          <ZoomIn className="h-3 w-3" />
                          {t('common.zoomIn', 'Zoom In')}
                        </button>
                        <button
                          onClick={() => zoom.scale({ scaleX: 0.8, scaleY: 0.8 })}
                          className="text-xs px-2 py-1 bg-bg-secondary rounded hover:bg-bg-tertiary transition-colors flex items-center gap-1"
                        >
                          <ZoomOut className="h-3 w-3" />
                          {t('common.zoomOut', 'Zoom Out')}
                        </button>
                        <button
                          onClick={zoom.reset}
                          className="text-xs px-2 py-1 bg-bg-secondary rounded hover:bg-bg-tertiary transition-colors flex items-center gap-1"
                        >
                          <RotateCcw className="h-3 w-3" />
                          {t('common.reset', 'Reset')}
                        </button>
                      </div>
                    </foreignObject>
                  </g>
                </svg>
              )}
            </Zoom>

            <div className="absolute bottom-4 left-4 bg-bg-primary/90 backdrop-blur rounded-lg p-3 border shadow-sm">
              <p className="text-xs font-medium mb-2">{t('swarm.roles', 'Roles')}</p>
              <div className="space-y-1">
                {Object.entries(roleColors).map(([role, color]) => {
                  const Icon = roleIcons[role as keyof typeof roleIcons];
                  return (
                    <div key={role} className="flex items-center gap-2 text-xs">
                      <div className="w-2 h-2 rounded-full" style={{ backgroundColor: color }} />
                      <span className="capitalize flex items-center gap-1">
                        <Icon className="h-3 w-3" style={{ color }} />
                        {role}
                      </span>
                    </div>
                  );
                })}
              </div>
            </div>

            <div className="absolute bottom-4 left-4 ml-32 bg-bg-primary/90 backdrop-blur rounded-lg p-3 border shadow-sm">
              <p className="text-xs font-medium mb-2">{t('swarm.status', 'Status')}</p>
              <div className="space-y-1">
                {Object.entries(statusColors).map(([status, color]) => (
                  <div key={status} className="flex items-center gap-2 text-xs">
                    <div
                      className="w-2 h-2 rounded-full border-2"
                      style={{ borderColor: color, backgroundColor: 'transparent' }}
                    />
                    <span className="capitalize">{status}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <Card className="h-fit">
            <CardHeader className="pb-3">
              <CardTitle className="text-base">
                {selectedNode
                  ? t('swarm.agentDetails', 'Agent Details')
                  : t('swarm.topologyInfo', 'Topology Info')}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {selectedNode ? (
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <h4 className="font-semibold">{selectedNode.name}</h4>
                    <Badge variant="outline" className={roleBadgeColors[selectedNode.swarmRole]}>
                      {selectedNode.swarmRole}
                    </Badge>
                  </div>

                  <div className="space-y-3">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground flex items-center gap-2">
                        <Circle
                          className="h-3 w-3"
                          style={{ color: statusColors[selectedNode.status] }}
                        />
                        Status
                      </span>
                      <span className="font-medium capitalize">{selectedNode.status}</span>
                    </div>

                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground flex items-center gap-2">
                        <Shield className="h-3 w-3" />
                        Trust Score
                      </span>
                      <span className="font-medium">{(selectedNode.trustScore * 100).toFixed(0)}%</span>
                    </div>

                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground flex items-center gap-2">
                        <Users className="h-3 w-3" />
                        Agent ID
                      </span>
                      <span className="font-medium text-xs">{selectedNode.agentId.slice(0, 8)}...</span>
                    </div>
                  </div>

                  <div className="pt-3 border-t">
                    <Button
                      variant="outline"
                      size="sm"
                      className="w-full"
                      onClick={() => (window.location.href = `/agents/${selectedNode.id}`)}
                    >
                      {t('swarm.viewDetails', 'View Full Details')}
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="space-y-4">
                  <div className="flex items-center gap-3 p-3 bg-bg-secondary rounded-lg">
                    <div className="p-2 bg-brand-500/10 rounded-lg">
                      <Server className="h-4 w-4 text-brand-500" />
                    </div>
                    <div>
                      <p className="font-medium text-sm">{rootAgent?.name || 'Root Agent'}</p>
                      <p className="text-xs text-muted-foreground">{t('swarm.rootAgent', 'Root Agent')}</p>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div className="p-3 bg-bg-secondary rounded-lg text-center">
                      <p className="text-2xl font-bold">{nodes.length}</p>
                      <p className="text-xs text-muted-foreground">{t('swarm.totalAgents', 'Total Agents')}</p>
                    </div>
                    <div className="p-3 bg-bg-secondary rounded-lg text-center">
                      <p className="text-2xl font-bold">{links.length}</p>
                      <p className="text-xs text-muted-foreground">{t('swarm.connections', 'Connections')}</p>
                    </div>
                  </div>

                  <p className="text-xs text-muted-foreground text-center">
                    {t('swarm.clickNode', 'Click on a node to view agent details')}
                  </p>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      </CardContent>
    </Card>
  );
}

export default SwarmTopologyView;
