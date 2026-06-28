import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { usePageTitle } from "@/hooks";
import { Database, Search, Plus, ChevronRight, Clock, Trash2, MoreVertical } from "lucide-react";
import { useStates, useDeleteState } from "@/hooks/useState";
import type { SimpleState } from "@/types";
import {
  PageGrid, Chamber, CornerBrace, TrustSeal,
  SealedButton, FrameButton, StatusPill, AnnotationTag,
} from "@/components/containment";
import "./styles.css";

const valueTypeToPill = (value: unknown): 'live' | 'pending' | 'revoked' => {
  if (value === null) return 'revoked';
  if (typeof value === 'boolean') return 'pending';
  return 'live';
};

const getValueType = (value: unknown): string => {
  if (value === null) return "null";
  if (Array.isArray(value)) return "array";
  return typeof value;
};

const formatValue = (value: unknown): string => {
  if (value === null) return "null";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
};

export function StatePage() {
  usePageTitle('State');
  const { t } = useTranslation();
  const navigate = useNavigate();
  const [searchQuery, setSearchQuery] = useState("");
  const [prefixFilter, setPrefixFilter] = useState("");
  const [openMenu, setOpenMenu] = useState<string | null>(null);

  const { data: states, isLoading, error } = useStates({ prefix: prefixFilter || undefined });
  const deleteState = useDeleteState();

  const filteredStates = states?.filter((state) =>
    state.path.toLowerCase().includes(searchQuery.toLowerCase()) ||
    state.key.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleCreate = () => navigate("/state/new");
  const handleView = (path: string) => navigate(`/state/${encodeURIComponent(path)}`);

  const handleDelete = async (path: string) => {
    if (window.confirm(`Are you sure you want to delete "${path}"?`)) {
      await deleteState.mutateAsync(path);
    }
  };

  return (
    <div className="st-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="st-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE ST-01" secondary="Simple State" position="top-right" />

        <div className="st-hero__header">
          <div className="st-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="st-hero__title">Simple State</h1>
          </div>
          <p className="st-hero__subtitle">Manage your key-value state storage</p>
          <div className="st-hero__actions">
            <SealedButton onClick={handleCreate} iconLeft={<Plus className="st-icon-sm" />}>
              {t("Create State")}
            </SealedButton>
          </div>
        </div>
      </Chamber>

      {/* Search + Filter */}
      <div className="st-controls">
        <div className="st-search">
          <Search className="st-search__icon" />
          <input className="st-input st-search__input" placeholder="Search by path or key..."
            value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)} />
        </div>
        <input className="st-input st-filter-input" placeholder="Filter by prefix..."
          value={prefixFilter} onChange={(e) => setPrefixFilter(e.target.value)} />
      </div>

      {/* Table */}
      <Chamber className="st-table-chamber">
        <CornerBrace position="tr" />
        <CornerBrace position="bl" />

        <div className="st-table-header">
          <h2 className="st-table-title">States ({filteredStates?.length || 0})</h2>
        </div>

        {isLoading ? (
          <div className="st-loading">
            {[...Array(5)].map((_, i) => <div key={i} className="st-skeleton" />)}
          </div>
        ) : error ? (
          <div className="st-error">
            <p>Failed to load states: {(error as Error).message}</p>
          </div>
        ) : filteredStates?.length === 0 ? (
          <div className="st-empty">
            <Database className="st-empty__icon" />
            <p className="st-empty__desc">No states found</p>
          </div>
        ) : (
          <div className="st-table-wrapper">
            <table className="st-table">
              <thead>
                <tr>
                  <th>Path</th>
                  <th>Key</th>
                  <th>Value</th>
                  <th>Version</th>
                  <th>Updated</th>
                  <th className="st-th-actions"></th>
                </tr>
              </thead>
              <tbody>
                {filteredStates?.map((state) => (
                  <tr key={state.path} className="st-table-row" onClick={() => handleView(state.path)}>
                    <td className="st-td-path">{state.path}</td>
                    <td className="st-td-key">{state.key}</td>
                    <td>
                      <div className="st-td-value">
                        <span className="st-td-value-text">{formatValue(state.value)}</span>
                        <span className={`st-type-badge st-type-badge--${getValueType(state.value)}`}>
                          {getValueType(state.value)}
                        </span>
                      </div>
                    </td>
                    <td><span className="st-version-badge">v{state.version}</span></td>
                    <td>
                      <span className="st-td-updated">
                        <Clock className="st-icon-xs" />
                        {new Date(state.updatedAt).toLocaleDateString()}
                      </span>
                    </td>
                    <td className="st-td-actions" onClick={(e) => e.stopPropagation()}>
                      <div className="st-menu-wrap">
                        <button className="st-icon-btn" onClick={() => setOpenMenu(openMenu === state.path ? null : state.path)}>
                          <MoreVertical className="st-icon-sm" />
                        </button>
                        {openMenu === state.path && (
                          <div className="st-menu">
                            <button className="st-menu__item" onClick={() => { handleView(state.path); setOpenMenu(null); }}>
                              <ChevronRight className="st-icon-xs" /> View Details
                            </button>
                            <button className="st-menu__item st-menu__item--danger" onClick={() => { handleDelete(state.path); setOpenMenu(null); }}>
                              <Trash2 className="st-icon-xs" /> Delete
                            </button>
                          </div>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Chamber>
    </div>
  );
}

export default StatePage;
