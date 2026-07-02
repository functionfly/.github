import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import {
  Shield,
  FileText,
  AlertTriangle,
  Scale,
  Gavel,
  ChevronDown,
  ChevronRight,
  Info,
  AlertCircle,
  Trash2,
  Ban,
} from 'lucide-react';
import { communityApi, type CommunityRule } from '@/api/community';
import { Card } from '@/components/containment';

const CATEGORY_META: Record<
  string,
  { label: string; icon: React.ReactNode; color: string }
> = {
  conduct: { label: 'Conduct', icon: <Shield size={12} />, color: 'var(--status-ok)' },
  content: { label: 'Content', icon: <FileText size={12} />, color: 'var(--text-dim)' },
  safety: { label: 'Safety', icon: <AlertTriangle size={12} />, color: 'var(--status-revoked)' },
  legal: { label: 'Legal', icon: <Scale size={12} />, color: 'var(--text-faint)' },
  moderation: { label: 'Moderation', icon: <Gavel size={12} />, color: 'var(--text-dim)' },
};

const ENFORCEMENT_META: Record<string, { label: string; icon: React.ReactNode }> = {
  info: { label: 'Info', icon: <Info size={10} /> },
  warning: { label: 'Warning', icon: <AlertCircle size={10} /> },
  deletion: { label: 'Deletion', icon: <Trash2 size={10} /> },
  suspension: { label: 'Suspension', icon: <Ban size={10} /> },
};

export function CommunityRules() {
  const [expanded, setExpanded] = useState(true);
  const [expandedRule, setExpandedRule] = useState<string | null>(null);

  const { data } = useQuery({
    queryKey: ['community-rules'],
    queryFn: () => communityApi.listRules(),
    staleTime: 5 * 60 * 1000,
  });

  const rules = data?.rules ?? [];
  const grouped = rules.reduce(
    (acc, rule) => {
      (acc[rule.category] = acc[rule.category] || []).push(rule);
      return acc;
    },
    {} as Record<string, CommunityRule[]>
  );

  const categoryOrder = ['conduct', 'content', 'safety', 'legal', 'moderation'];

  return (
    <Card className="sc-community-sidebar-card sc-community-rules-card">
      <button
        className="sc-community-rules-header"
        onClick={() => setExpanded(!expanded)}
        type="button"
      >
        <div className="sc-community-sidebar-title" style={{ padding: 0, marginBottom: 0 }}>
          <Gavel size={12} />
          Community Guidelines
          <span className="sc-community-rules-count">{rules.length}</span>
        </div>
        {expanded ? <ChevronDown size={14} /> : <ChevronRight size={14} />}
      </button>

      {expanded && (
        <div className="sc-community-rules-body">
          {categoryOrder
            .filter((cat) => grouped[cat]?.length)
            .map((cat) => {
              const meta = CATEGORY_META[cat] ?? CATEGORY_META.content;
              return (
                <div key={cat} className="sc-community-rules-group">
                  <div className="sc-community-rules-group-header">
                    <span className="sc-community-rules-group-icon" style={{ color: meta.color }}>
                      {meta.icon}
                    </span>
                    <span className="sc-community-rules-group-label">{meta.label}</span>
                  </div>
                  <ol className="sc-community-rules-list">
                    {grouped[cat].map((rule) => {
                      const isOpen = expandedRule === rule.id;
                      const enforce = ENFORCEMENT_META[rule.enforcement] ?? ENFORCEMENT_META.info;
                      return (
                        <li key={rule.id} className="sc-community-rule-item">
                          <button
                            className={`sc-community-rule-trigger ${isOpen ? 'open' : ''}`}
                            onClick={() => setExpandedRule(isOpen ? null : rule.id)}
                            type="button"
                          >
                            <span className="sc-community-rule-title">{rule.title}</span>
                            <span
                              className="sc-community-rule-enforcement"
                              title={`Enforcement: ${enforce.label}`}
                            >
                              {enforce.icon}
                            </span>
                          </button>
                          {isOpen && rule.description && (
                            <p className="sc-community-rule-desc">{rule.description}</p>
                          )}
                        </li>
                      );
                    })}
                  </ol>
                </div>
              );
            })}
        </div>
      )}
    </Card>
  );
}
