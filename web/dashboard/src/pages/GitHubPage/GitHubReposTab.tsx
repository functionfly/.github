import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Skeleton } from '@/components/ui/skeleton';
import { ToggleButtonGroup } from '@/components/ui';
import { useGitHubRepos, useRefreshGitHubRepos } from '@/hooks/useGitHubRepos';
import { useGitHubStore } from '@/stores/githubStore';
import { motion } from 'framer-motion';
import {
  ChevronLeft,
  ChevronRight,
  GitFork,
  Globe,
  Hash,
  LayoutGrid,
  List,
  Lock,
  RefreshCw,
  Search,
  Star,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';

const SUPPORTED_LANGUAGES = [
  'TypeScript',
  'JavaScript',
  'Python',
  'Go',
  'Rust',
];

const LANGUAGE_OPTIONS = ['All', ...SUPPORTED_LANGUAGES];

const SUPPORTED_RUNTIMES = new Set([
  'typescript',
  'javascript',
  'python',
  'python-wasm',
  'rust-wasm',
  'go',
  'deno',
  'bun',
  'browser-wasm',
]);

const VISIBILITY_OPTIONS = ['all', 'public', 'private'] as const;

export function GitHubReposTab() {
  const navigate = useNavigate();
  const setSelectedRepo = useGitHubStore((s) => s.setSelectedRepo);

  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [language, setLanguage] = useState('All');
  const [visibility, setVisibility] = useState<(typeof VISIBILITY_OPTIONS)[number]>('all');
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [page, setPage] = useState(1);
  const perPage = 12;

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(searchQuery);
      setPage(1);
    }, 300);
    return () => clearTimeout(timer);
  }, [searchQuery]);

  const reposParams = useMemo(
    () => ({
      page,
      per_page: perPage,
      search: debouncedSearch || undefined,
      sort: 'updated_at' as const,
      direction: 'desc' as const,
    }),
    [page, debouncedSearch]
  );

  const { data, isLoading, error } = useGitHubRepos(reposParams);
  const refreshMutation = useRefreshGitHubRepos();

  const repos = data?.repos ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / perPage);

  const filteredRepos = useMemo(() => {
    return repos.filter((repo) => {
      const isSupported =
        SUPPORTED_LANGUAGES.includes(repo.language ?? '') ||
        (repo.detected_runtime != null &&
          SUPPORTED_RUNTIMES.has(repo.detected_runtime));
      if (!isSupported) return false;
      if (language !== 'All' && repo.language !== language) return false;
      if (visibility === 'public' && repo.is_private) return false;
      if (visibility === 'private' && !repo.is_private) return false;
      return true;
    });
  }, [repos, language, visibility]);

  const handleRepoClick = useCallback(
    (repoId: string) => {
      navigate(`/github/import/${repoId}`);
    },
    [navigate]
  );

  const stats = useMemo(() => {
    if (!data) return null;
    return {
      total: filteredRepos.length,
      languages: new Set(filteredRepos.map((r) => r.language).filter(Boolean)).size,
      private: filteredRepos.filter((r) => r.is_private).length,
    };
  }, [data, filteredRepos]);

  if (error) {
    return (
      <div className="text-center py-12">
        <p className="text-text-secondary mb-4">Failed to load repositories</p>
        <Button
          variant="outline"
          onClick={() => refreshMutation.mutate()}
          disabled={refreshMutation.isPending}
        >
          <RefreshCw className={`w-4 h-4 mr-2 ${refreshMutation.isPending ? 'animate-spin' : ''}`} />
          Retry
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Stats Bar */}
      {stats && (
        <div className="flex items-center gap-6 text-sm">
          <div className="flex items-center gap-2">
            <Hash className="w-4 h-4 text-text-muted" />
            <span className="text-text-secondary">
              <span className="font-semibold text-text-primary">{stats.total}</span> repos
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Globe className="w-4 h-4 text-text-muted" />
            <span className="text-text-secondary">
              <span className="font-semibold text-text-primary">{stats.languages}</span> languages
            </span>
          </div>
          <div className="flex items-center gap-2">
            <Lock className="w-4 h-4 text-text-muted" />
            <span className="text-text-secondary">
              <span className="font-semibold text-text-primary">{stats.private}</span> private
            </span>
          </div>
        </div>
      )}

      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
          <Input
            placeholder="Search repositories..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>

        <Select value={language} onValueChange={setLanguage}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder="Language" />
          </SelectTrigger>
          <SelectContent>
            {LANGUAGE_OPTIONS.map((lang) => (
              <SelectItem key={lang} value={lang}>
                {lang}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <Select value={visibility} onValueChange={(v) => setVisibility(v as typeof visibility)}>
          <SelectTrigger className="w-[140px]">
            <SelectValue placeholder="Visibility" />
          </SelectTrigger>
          <SelectContent>
            {VISIBILITY_OPTIONS.map((v) => (
              <SelectItem key={v} value={v}>
                {v.charAt(0).toUpperCase() + v.slice(1)}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <ToggleButtonGroup
          value={viewMode}
          onValueChange={(v) => setViewMode(v as 'grid' | 'list')}
          options={[
            { value: 'grid', label: 'Grid', icon: <LayoutGrid className="h-4 w-4" /> },
            { value: 'list', label: 'List', icon: <List className="h-4 w-4" /> },
          ]}
          variant="outline"
          size="sm"
        />

        <Button
          variant="outline"
          size="sm"
          onClick={() => refreshMutation.mutate()}
          disabled={refreshMutation.isPending}
        >
          <RefreshCw className={`w-4 h-4 mr-2 ${refreshMutation.isPending ? 'animate-spin' : ''}`} />
          Refresh
        </Button>
      </div>

      {/* Loading */}
      {isLoading && (
        <div
          className={
            viewMode === 'grid'
              ? 'grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4'
              : 'space-y-2'
          }
        >
          {Array.from({ length: 6 }).map((_, i) => (
            <Skeleton key={i} className={viewMode === 'grid' ? 'h-40 rounded-lg' : 'h-16 rounded-lg'} />
          ))}
        </div>
      )}

      {/* Repos Grid/List */}
      {!isLoading && filteredRepos.length === 0 && (
        <div className="text-center py-12">
          <p className="text-text-secondary">No repositories found</p>
        </div>
      )}

      {!isLoading && filteredRepos.length > 0 && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          className={
            viewMode === 'grid'
              ? 'grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4'
              : 'space-y-2'
          }
        >
          {filteredRepos.map((repo) => (
            <button
              key={repo.id}
              onClick={() => handleRepoClick(repo.id)}
              className={
                viewMode === 'grid'
                  ? 'text-left p-4 rounded-lg border border-border-default bg-bg-secondary hover:border-brand-500/50 hover:bg-bg-tertiary transition-all group'
                  : 'text-left w-full flex items-center gap-4 p-3 rounded-lg border border-border-default bg-bg-secondary hover:border-brand-500/50 hover:bg-bg-tertiary transition-all group'
              }
            >
              <div className={viewMode === 'grid' ? 'space-y-3' : 'flex-1 min-w-0 flex items-center gap-4'}>
                <div className="flex items-center gap-2 min-w-0">
                  {repo.is_private ? (
                    <Lock className="w-4 h-4 text-text-muted shrink-0" />
                  ) : (
                    <Globe className="w-4 h-4 text-text-muted shrink-0" />
                  )}
                  <span className="font-medium text-text-primary truncate group-hover:text-brand-500 transition-colors">
                    {repo.full_name}
                  </span>
                </div>
                {repo.description && (
                  <p className="text-sm text-text-secondary line-clamp-2">{repo.description}</p>
                )}
                <div className="flex items-center gap-3 text-xs text-text-muted">
                  {repo.language && (
                    <span className="flex items-center gap-1">
                      <span className="w-2 h-2 rounded-full bg-brand-500" />
                      {repo.language}
                    </span>
                  )}
                  <span className="flex items-center gap-1">
                    <Star className="w-3 h-3" />
                    {repo.stars_count}
                  </span>
                  <span className="flex items-center gap-1">
                    <GitFork className="w-3 h-3" />
                    {repo.forks_count}
                  </span>
                </div>
              </div>
            </button>
          ))}
        </motion.div>
      )}

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between">
          <span className="text-sm text-text-muted">
            Page {page} of {totalPages}
          </span>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
            >
              <ChevronLeft className="w-4 h-4" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
            >
              <ChevronRight className="w-4 h-4" />
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
