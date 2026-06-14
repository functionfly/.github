import { useState, useEffect, useRef } from 'react';
import { changelogData as fallbackData, changeTypeLabels, changeTypeColors, type ChangeType, type Release } from './data/changelog';
import { API_ORIGIN } from '../../config';

const ALL_FILTERS: ChangeType[] = ['new', 'improved', 'fixed', 'security', 'breaking'];

interface BackendChangelogEntry {
  id: string;
  version: string;
  date: string;
  type: 'major' | 'minor' | 'patch';
  title: string;
  description: string;
  changes: BackendChangelogChange[];
  release_url?: string;
  is_published: boolean;
}

interface BackendChangelogChange {
  id: string;
  entry_id: string;
  category: string;
  icon: string;
  items: string[];
}

interface CacheEntry {
  data: Release[];
  timestamp: number;
}

let cache: CacheEntry | null = null;
const CACHE_TTL_MS = 5 * 60 * 1000;

function mapBackendToFrontend(entries: BackendChangelogEntry[]): Release[] {
  return entries.map((entry, idx) => ({
    version: entry.version,
    date: entry.date,
    isLatest: idx === 0,
    changes: entry.changes.flatMap((change) =>
      change.items.map((item) => ({
        type: change.category as ChangeType,
        description: item,
        details: undefined,
      }))
    ),
  }));
}

interface FilterButtonProps {
  type: ChangeType;
  active: boolean;
  count: number;
  onClick: () => void;
}

function FilterButton({ type, active, count, onClick }: FilterButtonProps) {
  const colors = changeTypeColors[type];
  return (
    <button
      onClick={onClick}
      className={`ff-changelog-filter-btn ${active ? 'active' : ''}`}
      style={{
        '--filter-bg': colors.bg,
        '--filter-border': colors.border,
        '--filter-text': colors.text,
      } as React.CSSProperties}
    >
      <span
        className="ff-changelog-filter-dot"
        style={{ background: colors.dot }}
      />
      <span>{changeTypeLabels[type]}</span>
      <span className="ff-changelog-filter-count">{count}</span>
    </button>
  );
}

interface ChangeItemProps {
  change: Release['changes'][0];
  index: number;
}

function ChangeItem({ change, index }: ChangeItemProps) {
  const colors = changeTypeColors[change.type];
  return (
    <div
      className="ff-changelog-change"
      style={{
        '--change-bg': colors.bg,
        '--change-border': colors.border,
        '--change-text': colors.text,
        animationDelay: `${index * 80}ms`,
      } as React.CSSProperties}
    >
      <div className="ff-changelog-change-type">
        <span
          className="ff-changelog-change-dot"
          style={{ background: colors.dot }}
        />
        <span style={{ color: colors.text }}>{changeTypeLabels[change.type]}</span>
      </div>
      <div className="ff-changelog-change-content">
        <p className="ff-changelog-change-desc">{change.description}</p>
        {change.details && (
          <p className="ff-changelog-change-details">{change.details}</p>
        )}
      </div>
    </div>
  );
}

interface ReleaseCardProps {
  release: Release;
  isFirst: boolean;
  index: number;
}

function ReleaseCard({ release, isFirst, index }: ReleaseCardProps) {
  const cardRef = useRef<HTMLDivElement>(null);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsVisible(true);
          observer.disconnect();
        }
      },
      { threshold: 0.1 }
    );

    if (cardRef.current) {
      observer.observe(cardRef.current);
    }

    return () => observer.disconnect();
  }, []);

  const formattedDate = new Date(release.date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });

  return (
    <div
      ref={cardRef}
      className={`ff-changelog-release ${isVisible ? 'visible' : ''} ${isFirst ? 'latest' : ''}`}
      style={{ '--delay': `${index * 100}ms` } as React.CSSProperties}
    >
      <div className="ff-changelog-release-marker">
        <div className="ff-changelog-release-dot" />
        {!isFirst && <div className="ff-changelog-release-line" />}
      </div>

      <div className="ff-changelog-release-content">
        <div className="ff-changelog-release-header">
          <div className="ff-changelog-release-info">
            <h3 className="ff-changelog-release-version">
              {release.version}
              {release.isLatest && <span className="ff-changelog-latest-badge">Latest</span>}
            </h3>
            <time className="ff-changelog-release-date" dateTime={release.date}>
              {formattedDate}
            </time>
          </div>
        </div>

        <div className="ff-changelog-changes">
          {release.changes.map((change, i) => (
            <ChangeItem key={i} change={change} index={i} />
          ))}
        </div>
      </div>
    </div>
  );
}

function LoadingSkeleton() {
  return (
    <div className="ff-changelog-loading">
      <div className="ff-changelog-skeleton-hero">
        <div className="skeleton-badge" />
        <div className="skeleton-title" />
        <div className="skeleton-lead" />
      </div>
      <div className="ff-changelog-skeleton-filters">
        <div className="skeleton-filter" />
        <div className="skeleton-filter" />
        <div className="skeleton-filter" />
      </div>
      <div className="ff-changelog-skeleton-timeline">
        <div className="skeleton-card" />
        <div className="skeleton-card" />
        <div className="skeleton-card" />
      </div>
    </div>
  );
}

export default function ChangelogPage() {
  const [releases, setReleases] = useState<Release[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeFilters, setActiveFilters] = useState<Set<ChangeType>>(new Set(ALL_FILTERS));
  const [isFilterVisible, setIsFilterVisible] = useState(false);
  const filterRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const fetchChangelog = async () => {
      const now = Date.now();
      if (cache && now - cache.timestamp < CACHE_TTL_MS) {
        setReleases(cache.data);
        setLoading(false);
        return;
      }

      try {
        const response = await fetch(`${API_ORIGIN}/v1/content/changelog?limit=20`);
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }
        const json = await response.json();
        const entries = json.entries as BackendChangelogEntry[];
        const mapped = mapBackendToFrontend(entries);
        if (mapped.length === 0) {
          throw new Error('No changelog entries returned');
        }
        cache = { data: mapped, timestamp: now };
        setReleases(mapped);
        setError(null);
      } catch (err) {
        console.warn('Failed to fetch changelog from API, using fallback:', err);
        setReleases(fallbackData);
        setError('Using cached data');
      } finally {
        setLoading(false);
      }
    };

    fetchChangelog();
  }, []);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          setIsFilterVisible(true);
          observer.disconnect();
        }
      },
      { threshold: 0.1 }
    );

    if (filterRef.current) {
      observer.observe(filterRef.current);
    }

    return () => observer.disconnect();
  }, []);

  const toggleFilter = (type: ChangeType) => {
    setActiveFilters((prev) => {
      const next = new Set(prev);
      if (next.has(type)) {
        if (next.size > 1) next.delete(type);
      } else {
        next.add(type);
      }
      return next;
    });
  };

  const getFilterCount = (type: ChangeType) => {
    return releases.reduce((acc, release) => {
      return acc + release.changes.filter((c) => c.type === type).length;
    }, 0);
  };

  const filteredReleases = releases.filter((release) =>
    release.changes.some((change) => activeFilters.has(change.type))
  );

  const allCounts = ALL_FILTERS.reduce(
    (acc, type) => {
      acc[type] = getFilterCount(type);
      return acc;
    },
    {} as Record<ChangeType, number>
  );

  if (loading) {
    return (
      <div className="ff-homepage">
        <section className="ff-hero-section">
          <div className="ff-hero-inner">
            <div className="ff-hero-eyebrow">
              <span className="ff-pulse-dot" />
              <span>Product Updates</span>
            </div>
            <h1 className="ff-hero-headline">Changelog</h1>
            <p className="ff-hero-sub">
              High-level product and API changes. Detailed per-service changelogs are in the docs.
            </p>
          </div>
        </section>
        <LoadingSkeleton />
      </div>
    );
  }

  return (
    <div className="ff-homepage">
      <section className="ff-hero-section">
        <div className="ff-hero-inner">
          <div className="ff-hero-eyebrow">
            <span className="ff-pulse-dot" />
            <span>Product Updates</span>
          </div>
          <h1 className="ff-hero-headline">Changelog</h1>
          <p className="ff-hero-sub">
            High-level product and API changes. Detailed per-service changelogs are in the docs.
          </p>
        </div>
      </section>

      <div className="ff-changelog-content">
        <div
          ref={filterRef}
          className={`ff-changelog-filters ${isFilterVisible ? 'visible' : ''}`}
        >
          <div className="ff-changelog-filters-inner">
            <span className="ff-changelog-filters-label">Filter by type:</span>
            <div className="ff-changelog-filter-buttons">
              {ALL_FILTERS.map((type) => (
                <FilterButton
                  key={type}
                  type={type}
                  active={activeFilters.has(type)}
                  count={allCounts[type]}
                  onClick={() => toggleFilter(type)}
                />
              ))}
            </div>
          </div>
        </div>

        <div className="ff-changelog-timeline">
          {filteredReleases.map((release, index) => (
            <ReleaseCard
              key={release.version}
              release={release}
              isFirst={index === 0}
              index={index}
            />
          ))}
        </div>

        {filteredReleases.length === 0 && (
          <div className="ff-changelog-empty">
            <p>No changes match the selected filters.</p>
          </div>
        )}
      </div>
    </div>
  );
}
