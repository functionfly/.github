import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { GitCompare } from 'lucide-react';
import ChangelogEntry from '../ChangelogEntry';
import { ChangelogEntry as ChangelogEntryType } from '@/api/content';

interface CompareTabProps {
  compareVersion1: string;
  setCompareVersion1: (value: string) => void;
  compareVersion2: string;
  setCompareVersion2: (value: string) => void;
  availableVersions: Array<{ value: string; label: string }>;
  comparisonData: { entry1: ChangelogEntryType; entry2: ChangelogEntryType } | null;
  differences: {
    newCategories: string[];
    removedCategories: string[];
    changedCategories: string[];
    categoryDetails: Record<string, { added: string[]; removed: string[]; unchanged: string[] }>;
  } | null;
}

const CompareTab = ({
  compareVersion1,
  setCompareVersion1,
  compareVersion2,
  setCompareVersion2,
  availableVersions,
  comparisonData,
  differences
}: CompareTabProps) => {
  return (
    <div className="space-y-8">
      {/* Version Comparison Controls */}
      <Card className="glass-card glow animate-float border-brand-500/20 bg-gradient-to-br from-brand-500/5 via-purple-500/5 to-transparent">
        <CardHeader className="text-center pb-6">
          <div className="flex justify-center mb-4">
            <div className="p-4 bg-gradient-to-br from-brand-500/20 to-purple-500/20 rounded-2xl border border-brand-500/30">
              <GitCompare className="h-8 w-8 text-brand-500 animate-pulse-glow" />
            </div>
          </div>
          <CardTitle className="text-2xl lg:text-3xl text-gradient">
            Compare Versions
          </CardTitle>
          <p className="text-text-secondary text-lg mt-2">
            Select two versions to compare their changes side by side and see what's new or different
          </p>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div className="space-y-3">
              <label className="text-sm font-semibold text-text-primary flex items-center gap-2">
                <div className="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></div>
                Version 1 (Older)
              </label>
              <Select value={compareVersion1} onValueChange={setCompareVersion1}>
                <SelectTrigger className="glass-card border-border-subtle/50 hover:border-brand-500/50 transition-all duration-300 hover:glow-sm h-12">
                  <SelectValue placeholder="Select version" />
                </SelectTrigger>
                <SelectContent className="glass-card border-border-subtle/50">
                  {availableVersions.map(version => (
                    <SelectItem key={version.value} value={version.value} className="hover:bg-brand-500/10">
                      {version.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="space-y-3">
              <label className="text-sm font-semibold text-text-primary flex items-center gap-2">
                <div className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                Version 2 (Newer)
              </label>
              <Select value={compareVersion2} onValueChange={setCompareVersion2}>
                <SelectTrigger className="glass-card border-border-subtle/50 hover:border-brand-500/50 transition-all duration-300 hover:glow-sm h-12">
                  <SelectValue placeholder="Select version" />
                </SelectTrigger>
                <SelectContent className="glass-card border-border-subtle/50">
                  {availableVersions.map(version => (
                    <SelectItem key={version.value} value={version.value} className="hover:bg-brand-500/10">
                      {version.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Comparison Results */}
      {comparisonData && differences && (() => {
        const { entry1, entry2 } = comparisonData;

        return (
          <div className="space-y-6">
            <div className="text-center">
              <h3 className="text-xl font-semibold text-gradient mb-2">Version Comparison</h3>
              <p className="text-text-secondary">Changes between {entry1.version} and {entry2.version}</p>
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
              {/* Version 1 */}
              <div className="relative">
                <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                  <div className="bg-gradient-to-r from-blue-500 to-cyan-500 text-white px-4 py-2 rounded-full text-sm font-semibold shadow-lg">
                    {entry1.version}
                  </div>
                </div>
                <ChangelogEntry
                  entry={entry1}
                  variant="regular"
                  comparisonMode={true}
                />
              </div>

              {/* Version 2 */}
              <div className="relative">
                <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                  <div className="bg-gradient-to-r from-green-500 to-emerald-500 text-white px-4 py-2 rounded-full text-sm font-semibold shadow-lg">
                    {entry2.version}
                  </div>
                </div>
                <ChangelogEntry
                  entry={entry2}
                  variant="regular"
                  comparisonMode={true}
                />
              </div>
            </div>
          </div>
        );
      })()}
    </div>
  );
};

export default CompareTab;