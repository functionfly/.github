/**
 * FRGGraphsPage - Function Runtime Graphs Gallery
 */

import './styles.css';

import { useState, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { usePageTitle } from '@/hooks';
import {
  Plus, Upload, Search, Filter, MoreHorizontal, Play, Copy, Trash2,
  Edit3, Clock, Activity, Zap, GitBranch, Webhook, Database, FileJson,
  Sparkles, ChevronRight, AlertCircle, Loader2,
} from 'lucide-react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { frgApi } from '@/api/frg';
import type { GraphDefinition } from '@/types/frg';
import {
  PageGrid, Chamber, CornerBrace, TrustSeal,
  SealedButton, FrameButton, StatusPill, AnnotationTag, Card,
} from '@/components/containment';

const quickStartTemplates = [
  { id: 'ai-pipeline', name: 'AI Pipeline', description: 'Input → GPT-4 → Output processing chain', icon: Sparkles, nodeCount: 3 },
  { id: 'webhook-handler', name: 'Webhook Handler', description: 'API endpoint with validation and response', icon: Webhook, nodeCount: 4 },
  { id: 'data-processing', name: 'Data Processing', description: 'Upload → Transform → Store workflow', icon: Database, nodeCount: 5 },
  { id: 'etl-workflow', name: 'ETL Workflow', description: 'Extract, Transform, Load pipeline', icon: FileJson, nodeCount: 6 },
];

function GraphThumbnail() {
  return (
    <div className="frg-thumb">
      <svg className="frg-thumb__grid" viewBox="0 0 100 60">
        <defs>
          <pattern id="frg-grid" width="10" height="10" patternUnits="userSpaceOnUse">
            <path d="M 10 0 L 0 0 0 10" fill="none" stroke="currentColor" strokeWidth="0.5" />
          </pattern>
        </defs>
        <rect width="100" height="60" fill="url(#frg-grid)" />
        <circle cx="20" cy="30" r="4" fill="currentColor" className="frg-thumb__node frg-thumb__node--a" />
        <circle cx="50" cy="20" r="4" fill="currentColor" className="frg-thumb__node frg-thumb__node--b" />
        <circle cx="50" cy="40" r="4" fill="currentColor" className="frg-thumb__node frg-thumb__node--b" />
        <circle cx="80" cy="30" r="4" fill="currentColor" className="frg-thumb__node frg-thumb__node--c" />
        <line x1="24" y1="30" x2="46" y2="20" stroke="currentColor" strokeWidth="1" className="frg-thumb__edge" />
        <line x1="24" y1="30" x2="46" y2="40" stroke="currentColor" strokeWidth="1" className="frg-thumb__edge" />
        <line x1="54" y1="20" x2="76" y2="30" stroke="currentColor" strokeWidth="1" className="frg-thumb__edge" />
        <line x1="54" y1="40" x2="76" y2="30" stroke="currentColor" strokeWidth="1" className="frg-thumb__edge" />
      </svg>
    </div>
  );
}

function GraphCard({ graph, onEdit, onRun, onDuplicate, onDelete }: {
  graph: GraphDefinition;
  onEdit: (g: GraphDefinition) => void;
  onRun: (g: GraphDefinition) => void;
  onDuplicate: (g: GraphDefinition) => void;
  onDelete: (g: GraphDefinition) => void;
}) {
  const isPublished = !!graph.publishedAt;
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <div className="frg-card">
      <div className="frg-card__thumb">
        <GraphThumbnail />
        <div className="frg-card__badge-wrap">
          <StatusPill status={isPublished ? 'live' : 'pending'} label={isPublished ? 'Published' : 'Draft'} />
        </div>
      </div>

      <div className="frg-card__body">
        <div className="frg-card__header">
          <div className="frg-card__title-group">
            <h3 className="frg-card__title">{graph.name}</h3>
            <p className="frg-card__meta">v{graph.version} · {graph.author}</p>
          </div>
          <div className="frg-card__menu-wrap">
            <button className="frg-card__menu-btn" onClick={() => setMenuOpen(!menuOpen)}>
              <MoreHorizontal className="frg-icon-sm" />
            </button>
            {menuOpen && (
              <div className="frg-card__menu">
                <button className="frg-card__menu-item" onClick={() => { onEdit(graph); setMenuOpen(false); }}>
                  <Edit3 className="frg-icon-xs" /> Edit
                </button>
                <button className="frg-card__menu-item" onClick={() => { onRun(graph); setMenuOpen(false); }}>
                  <Play className="frg-icon-xs" /> Run
                </button>
                <button className="frg-card__menu-item" onClick={() => { onDuplicate(graph); setMenuOpen(false); }}>
                  <Copy className="frg-icon-xs" /> Duplicate
                </button>
                <div className="frg-card__menu-sep" />
                <button className="frg-card__menu-item frg-card__menu-item--danger" onClick={() => { onDelete(graph); setMenuOpen(false); }}>
                  <Trash2 className="frg-icon-xs" /> Delete
                </button>
              </div>
            )}
          </div>
        </div>

        <div className="frg-card__tags">
          {graph.visibility && <span className="frg-card__tag">{graph.visibility}</span>}
          {graph.executionMode && <span className="frg-card__tag">{graph.executionMode}</span>}
        </div>

        <div className="frg-card__stats">
          <span className="frg-card__stat"><Clock className="frg-icon-xs" /> {new Date(graph.updatedAt).toLocaleDateString()}</span>
          <span className="frg-card__stat"><Activity className="frg-icon-xs" /> {graph.nodeRefs?.length || 0} nodes</span>
        </div>

        <div className="frg-card__actions">
          <button className="frg-card__action" onClick={() => onEdit(graph)}><Edit3 className="frg-icon-xs" /> Edit</button>
          <button className="frg-card__action" onClick={() => onRun(graph)}><Play className="frg-icon-xs" /> Run</button>
          <button className="frg-card__action" onClick={() => onDuplicate(graph)}><Copy className="frg-icon-xs" /> Copy</button>
        </div>
      </div>
    </div>
  );
}

export function FRGGraphsPage() {
  usePageTitle('FRG');
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [visibilityFilter, setVisibilityFilter] = useState('all');
  const [executionModeFilter, setExecutionModeFilter] = useState('all');

  const { data, isLoading, isError, error, refetch } = useQuery({
    queryKey: ['frg', 'graphs', visibilityFilter, executionModeFilter],
    queryFn: () => frgApi.listGraphs({
      visibility: visibilityFilter !== 'all' ? visibilityFilter : undefined,
      executionMode: executionModeFilter !== 'all' ? executionModeFilter : undefined,
    }),
  });

  const deleteMutation = useMutation({
    mutationFn: ({ author, name }: { author: string; name: string }) => frgApi.deleteGraph(author, name),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['frg', 'graphs'] }),
  });

  const remixMutation = useMutation({
    mutationFn: ({ author, name, newName }: { author: string; name: string; newName: string }) => frgApi.remixGraph(author, name, newName),
    onSuccess: (newGraph) => { queryClient.invalidateQueries({ queryKey: ['frg', 'graphs'] }); navigate(`/frg/${newGraph.author}/${newGraph.name}`); },
  });

  const graphs = data?.graphs || [];
  const filteredGraphs = useMemo(() => graphs.filter((g) =>
    g.name.toLowerCase().includes(searchQuery.toLowerCase()) || g.author.toLowerCase().includes(searchQuery.toLowerCase())
  ), [graphs, searchQuery]);

  const handleCreate = () => navigate('/frg/new');
  const handleImport = () => console.log('Import graph');
  const handleEdit = (g: GraphDefinition) => navigate(`/frg/${g.author}/${g.name}`);
  const handleRun = async (g: GraphDefinition) => { try { const r = await frgApi.executeGraph(g.author, g.name); if (r.instanceId) navigate(`/frg/${g.author}/${g.name}?instance=${r.instanceId}`); } catch (e) { console.error(e); } };
  const handleDuplicate = (g: GraphDefinition) => remixMutation.mutate({ author: g.author, name: g.name, newName: `${g.name}-copy` });
  const handleDelete = (g: GraphDefinition) => { if (confirm(`Delete "${g.name}"? This cannot be undone.`)) deleteMutation.mutate({ author: g.author, name: g.name }); };

  return (
    <div className="frg-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="frg-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE FRG-01" secondary="Function Runtime Graphs" position="top-right" />

        <div className="frg-hero__header">
          <div className="frg-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="frg-hero__title">Function Runtime Graphs</h1>
          </div>
          <p className="frg-hero__subtitle">Build, deploy, and monitor powerful function workflows</p>
          <div className="frg-hero__actions">
            <FrameButton size="sm" onClick={handleImport} iconLeft={<Upload className="frg-icon-sm" />}>Import</FrameButton>
            <SealedButton size="sm" onClick={handleCreate} iconLeft={<Plus className="frg-icon-sm" />}>Create Graph</SealedButton>
          </div>
        </div>
      </Chamber>

      {/* Search + Filters */}
      <div className="frg-controls">
        <div className="frg-search">
          <Search className="frg-search__icon" />
          <input className="frg-input frg-search__input" placeholder="Search graphs by name or author..."
            value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} />
        </div>
        <div className="frg-filters">
          <select className="frg-select" value={visibilityFilter} onChange={(e) => setVisibilityFilter(e.target.value)}>
            <option value="all">All Visibility</option>
            <option value="public">Public</option>
            <option value="private">Private</option>
            <option value="team">Team</option>
          </select>
          <select className="frg-select" value={executionModeFilter} onChange={(e) => setExecutionModeFilter(e.target.value)}>
            <option value="all">All Modes</option>
            <option value="sync">Sync</option>
            <option value="async">Async</option>
            <option value="streaming">Streaming</option>
            <option value="event_driven">Event Driven</option>
          </select>
        </div>
      </div>

      {/* Content */}
      <div className="frg-content-grid">
        <div>
          {isLoading ? (
            <Chamber className="frg-loading">
              <Loader2 className="frg-loading__spinner" />
            </Chamber>
          ) : isError ? (
            <Chamber className="frg-error">
              <AlertCircle className="frg-error__icon" />
              <h3 className="frg-error__title">Failed to load graphs</h3>
              <p className="frg-error__desc">{error instanceof Error ? error.message : 'Unknown error'}</p>
              <FrameButton onClick={() => refetch()}>Try Again</FrameButton>
            </Chamber>
          ) : filteredGraphs.length === 0 ? (
            <Chamber className="frg-empty">
              <GitBranch className="frg-empty__icon" />
              <h3 className="frg-empty__title">No graphs yet</h3>
              <p className="frg-empty__desc">Create your first Function Runtime Graph to start building powerful workflows</p>
              <SealedButton onClick={handleCreate} iconLeft={<Plus className="frg-icon-sm" />}>Create Your First Graph</SealedButton>
            </Chamber>
          ) : (
            <div className="frg-grid">
              {filteredGraphs.map((graph) => (
                <GraphCard key={`${graph.author}/${graph.name}@${graph.version}`} graph={graph}
                  onEdit={handleEdit} onRun={handleRun} onDuplicate={handleDuplicate} onDelete={handleDelete} />
              ))}
            </div>
          )}
        </div>

        {/* Templates Sidebar */}
        <div className="frg-templates">
          <Chamber className="frg-templates__chamber">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div className="frg-templates__header">
              <Zap className="frg-icon-sm frg-icon-accent" />
              <span className="frg-templates__title">Quick Start Templates</span>
            </div>
            <div className="frg-templates__list">
              {quickStartTemplates.map((tpl) => {
                const Icon = tpl.icon;
                return (
                  <button key={tpl.id} className="frg-template" onClick={() => navigate(`/frg/new?template=${tpl.id}`)}>
                    <div className="frg-template__icon-wrap"><Icon className="frg-template__icon" /></div>
                    <div className="frg-template__info">
                      <h4 className="frg-template__name">{tpl.name}</h4>
                      <p className="frg-template__desc">{tpl.description}</p>
                      <span className="frg-template__nodes"><Zap className="frg-icon-xs" /> {tpl.nodeCount} nodes</span>
                    </div>
                    <ChevronRight className="frg-template__arrow" />
                  </button>
                );
              })}
            </div>
          </Chamber>
        </div>
      </div>
    </div>
  );
}

export default FRGGraphsPage;
