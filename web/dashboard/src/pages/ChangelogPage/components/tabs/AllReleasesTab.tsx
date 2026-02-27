import { ChangelogEntry as ChangelogEntryType } from '@/api/content';
import ChangelogEntry from '../ChangelogEntry';
import SubscriptionCard from '../SubscriptionCard';

interface AllReleasesTabProps {
  filteredEntries: ChangelogEntryType[];
}

const AllReleasesTab = ({ filteredEntries }: AllReleasesTabProps) => {
  return (
    <div className="space-y-12">
      {/* Latest Version Highlight */}
      {filteredEntries.length > 0 && (
        <div className="relative">
          <div className="absolute -inset-4 bg-gradient-to-r from-brand-500/20 via-purple-500/20 to-pink-500/20 rounded-2xl blur-2xl animate-pulse-glow"></div>
          <div className="relative">
            <ChangelogEntry entry={filteredEntries[0]} variant="latest" />
          </div>
        </div>
      )}

      {/* All Versions */}
      <div className="space-y-8">
        {filteredEntries.slice(1).map((entry: ChangelogEntryType, index: number) => (
          <div
            key={entry.id}
            className="animate-fade-in"
            style={{ animationDelay: `${index * 150}ms` }}
          >
            <ChangelogEntry entry={entry} variant="regular" />
          </div>
        ))}
      </div>

      {/* Subscribe to Updates */}
      <div className="animate-fade-in" style={{ animationDelay: '500ms' }}>
        <SubscriptionCard />
      </div>
    </div>
  );
};

export default AllReleasesTab;