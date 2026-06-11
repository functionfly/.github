/**
 * Search Page
 * Agent search workspace for research, fact-finding, and documentation lookup
 */

import { SearchPanel } from '@/components/search';
import { usePageTitle } from '@/hooks';

export function SearchPage() {
  usePageTitle('Agent Search');

  return (
    <div className="container py-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold tracking-tight text-text-primary">
          Agent Search
        </h1>
        <p className="mt-1 text-text-secondary">
          Research, fact-finding, and documentation lookup for agents
        </p>
      </div>
      <SearchPanel />
    </div>
  );
}

export default SearchPage;