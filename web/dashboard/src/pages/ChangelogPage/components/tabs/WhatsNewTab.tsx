import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Sparkles, CheckCircle } from 'lucide-react';
import ChangelogEntry from '../ChangelogEntry';
import { ChangelogEntry as ChangelogEntryType } from '@/api/content';

interface WhatsNewTabProps {
  whatsNewVersion: string;
  setWhatsNewVersion: (value: string) => void;
  availableVersions: Array<{ value: string; label: string }>;
  whatsNewData: ChangelogEntryType[];
}

const WhatsNewTab = ({
  whatsNewVersion,
  setWhatsNewVersion,
  availableVersions,
  whatsNewData
}: WhatsNewTabProps) => {
  return (
    <div className="space-y-8">
      {/* What's New Controls */}
      <Card className="glass-card glow animate-float border-brand-500/20 bg-gradient-to-br from-brand-500/5 via-purple-500/5 to-pink-500/5">
        <CardHeader className="text-center pb-6">
          <div className="flex justify-center mb-4">
            <div className="p-4 bg-gradient-to-br from-yellow-500/20 to-orange-500/20 rounded-2xl border border-yellow-500/30">
              <Sparkles className="h-8 w-8 text-yellow-500 animate-pulse-glow" />
            </div>
          </div>
          <CardTitle className="text-2xl lg:text-3xl text-gradient">
            What's New Since
          </CardTitle>
          <p className="text-text-secondary text-lg mt-2">
            Discover all the exciting features and improvements added since your last update
          </p>
        </CardHeader>
        <CardContent>
          <div className="max-w-lg mx-auto">
            <label className="text-sm font-semibold text-text-primary flex items-center gap-2 mb-4 justify-center">
              <div className="w-2 h-2 bg-yellow-500 rounded-full animate-pulse"></div>
              Select a version to see what's new
            </label>
            <Select value={whatsNewVersion} onValueChange={setWhatsNewVersion}>
              <SelectTrigger className="glass-card border-border-subtle/50 hover:border-yellow-500/50 transition-all duration-300 hover:glow-sm h-12 mx-auto max-w-md">
                <SelectValue placeholder="Choose a version" />
              </SelectTrigger>
              <SelectContent className="glass-card border-border-subtle/50">
                {availableVersions.map(version => (
                  <SelectItem key={version.value} value={version.value} className="hover:bg-yellow-500/10">
                    {version.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* What's New Results */}
      {whatsNewData.length > 0 && (
        <div className="space-y-8">
          <div className="text-center">
            <h3 className="text-xl font-semibold text-gradient mb-2">
              ✨ New Releases Since {whatsNewVersion}
            </h3>
            <p className="text-text-secondary">Catch up on all the latest improvements</p>
          </div>

          {whatsNewData.map((entry: ChangelogEntryType, index: number) => (
            <div
              key={entry.id}
              className="animate-fade-in"
              style={{ animationDelay: `${index * 200}ms` }}
            >
              <ChangelogEntry entry={entry} variant="whatsnew" />
            </div>
          ))}
        </div>
      )}

      {whatsNewVersion && whatsNewData.length === 0 && (
        <Card className="glass-card glow animate-float text-center py-12 border-success/30 bg-gradient-to-br from-success/5 to-emerald-500/5">
          <CardContent className="pt-8">
            <div className="flex justify-center mb-6">
              <div className="p-6 bg-gradient-to-br from-success/20 to-emerald-500/20 rounded-3xl border border-success/30">
                <CheckCircle className="h-12 w-12 text-success animate-pulse-glow" />
              </div>
            </div>
            <h3 className="text-2xl font-bold text-gradient mb-4">🎉 You're All Caught Up!</h3>
            <p className="text-text-secondary text-lg leading-relaxed">
              No new releases since version <span className="font-semibold text-success">{whatsNewVersion}</span>.
              You're running the latest and greatest version of FunctionFly!
            </p>
            <div className="mt-6 p-4 bg-success/10 rounded-lg border border-success/20">
              <p className="text-sm text-success font-medium">
                Keep an eye out for future updates - we're always working on something exciting! 🚀
              </p>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};

export default WhatsNewTab;