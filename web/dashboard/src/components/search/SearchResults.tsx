import { ExternalLink, FileText, Building2, Newspaper, Globe } from 'lucide-react';

interface SearchResultsProps {
  toolName: 'search.web' | 'search.news' | 'search.docs' | 'search.company';
  results: unknown[];
  loading?: boolean;
}

const toolIcons = {
  'search.web': Globe,
  'search.news': Newspaper,
  'search.docs': FileText,
  'search.company': Building2,
};

export function SearchResults({ toolName, results, loading }: SearchResultsProps) {
  if (loading) {
    return <div className="p-4 text-muted-foreground">Searching...</div>;
  }

  if (!results.length) {
    return <div className="p-4 text-muted-foreground">No results found</div>;
  }

  const Icon = toolIcons[toolName];

  return (
    <div className="space-y-2">
      {results.map((result, index) => (
        <SearchResultItem key={index} result={result} Icon={Icon} toolName={toolName} />
      ))}
    </div>
  );
}

function SearchResultItem({
  result,
  Icon,
  toolName,
}: {
  result: unknown;
  Icon: React.ElementType;
  toolName: string;
}) {
  const item = result as Record<string, unknown>;
  const title = String(item.title ?? '');
  const url = String(item.url ?? '#');
  const snippet = String(item.snippet ?? item.description ?? '');
  const source = item.source ? String(item.source) : undefined;

  const IconComponent = Icon as React.ComponentType<{ className?: string }>;

  return (
    <div className="p-3 rounded-lg border bg-card hover:bg-accent transition-colors">
      <div className="flex items-start gap-3">
        <IconComponent className="h-5 w-5 text-muted-foreground mt-0.5" />
        <div className="flex-1 min-w-0">
          <a
            href={url}
            target="_blank"
            rel="noopener noreferrer"
            className="font-medium hover:underline flex items-center gap-1"
          >
            {title}
            <ExternalLink className="h-3 w-3" />
          </a>
          <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
            {snippet}
          </p>
          {source && (
            <p className="text-xs text-muted-foreground mt-1">
              {source}
            </p>
          )}
        </div>
      </div>
    </div>
  );
}