import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Calendar } from 'lucide-react';
import { ChangelogEntry as ChangelogEntryType, ChangelogChange } from '@/api/content';
import { getBadgeVariant, getTypeColor, getIconComponent } from '../utils/changelogUtils';

interface ChangelogEntryProps {
  entry: ChangelogEntryType;
  variant?: 'latest' | 'regular' | 'whatsnew';
  comparisonMode?: boolean;
  isNewCategory?: boolean;
  isRemovedCategory?: boolean;
  hasChanges?: boolean;
}

const ChangelogEntry = ({
  entry,
  variant = 'regular',
  comparisonMode = false,
  isNewCategory = false,
  isRemovedCategory = false,
  hasChanges = false
}: ChangelogEntryProps) => {
  const isLatest = variant === 'latest';
  const isWhatsNew = variant === 'whatsnew';

  const cardClassName = isLatest || isWhatsNew
    ? "glass-card glow-lg animate-float border-brand-500/20 bg-gradient-to-br from-brand-500/5 via-purple-500/5 to-transparent hover:from-brand-500/10 hover:via-purple-500/10 hover:to-pink-500/5 transition-all duration-500 hover:scale-[1.02] group"
    : comparisonMode
    ? `glass-card glow animate-float border-2 transition-all duration-300 ${
        isNewCategory
          ? 'bg-gradient-to-br from-green-500/10 to-emerald-500/10 border-green-500/30 hover:from-green-500/15 hover:to-emerald-500/15'
          : isRemovedCategory
          ? 'bg-gradient-to-br from-red-500/10 to-rose-500/10 border-red-500/30 hover:from-red-500/15 hover:to-rose-500/15'
          : hasChanges
          ? 'bg-gradient-to-br from-yellow-500/10 to-orange-500/10 border-yellow-500/30 hover:from-yellow-500/15 hover:to-orange-500/15'
          : 'bg-bg-glass/50 border-border-subtle/50 hover:bg-bg-glass/70 hover:border-border-default/50'
      }`
    : "glass-card glow hover:glow-lg transition-all duration-300 hover:scale-[1.01] group";

  const headerClassName = isLatest ? "text-3xl lg:text-4xl" : isWhatsNew ? "text-2xl lg:text-3xl" : "text-xl lg:text-2xl";
  const titleClassName = isLatest ? "text-2xl lg:text-3xl" : "text-xl lg:text-2xl";

  return (
    <Card className={cardClassName}>
      <CardHeader className="pb-6">
        <div className="flex items-start justify-between mb-4">
          <div className="flex items-center gap-4 flex-wrap">
            {isLatest && (
              <Badge className="bg-gradient-to-r from-brand-500 to-purple-500 text-white border-0 animate-pulse-glow shadow-lg">
                ✨ Latest Release
              </Badge>
            )}
            <h3 className={`font-bold text-gradient ${headerClassName} group-hover:scale-105 transition-transform duration-300`}>
              {entry.version}
            </h3>
            <div className="flex items-center gap-2 text-text-secondary bg-bg-glass/50 px-3 py-1 rounded-full border border-border-subtle/30">
              <Calendar className="h-4 w-4 text-brand-500 animate-pulse" />
              <span className="text-sm font-medium">{new Date(entry.date).toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'short',
                day: 'numeric'
              })}</span>
            </div>
          </div>
          <Badge className={`font-semibold text-xs px-3 py-1 rounded-full border-2 shadow-sm ${getTypeColor(entry.type)}`}>
            {entry.type.toUpperCase()}
          </Badge>
        </div>
        <CardTitle className={`${titleClassName} text-text-primary leading-tight group-hover:text-brand-500 transition-colors duration-300`}>
          {entry.title}
        </CardTitle>
      </CardHeader>
      <CardContent className="pt-0">
        <p className={`text-text-secondary text-lg leading-relaxed ${isLatest ? 'mb-8' : 'mb-6'} animate-fade-in-delayed`}>
          {entry.description}
        </p>
        <div className={`space-y-${isLatest || isWhatsNew ? '8' : '6'} animate-fade-in-delayed`}>
          {entry.changes.map((change: ChangelogChange, changeIndex: number) => {
            const IconComponent = getIconComponent(change.icon);
            const iconSize = isLatest || isWhatsNew ? 'h-6 w-6' : comparisonMode ? 'h-5 w-5' : 'h-5 w-5';
            const titleSize = isLatest || isWhatsNew ? 'text-lg font-semibold' : 'text-base font-semibold';
            const listMargin = isLatest || isWhatsNew ? 'ml-8' : comparisonMode ? 'ml-7' : 'ml-7';
            const listSpacing = 'space-y-2';

            return (
              <div
                key={change.id}
                className={`space-y-4 animate-fade-in`}
                style={{ animationDelay: `${changeIndex * 100}ms` }}
              >
                <div className="flex items-center gap-3">
                  <div className={`p-2 rounded-lg bg-gradient-to-br from-brand-500/10 to-purple-500/10 border border-brand-500/20 group-hover:from-brand-500/20 group-hover:to-purple-500/20 transition-all duration-300`}>
                    <IconComponent className={`${iconSize} text-brand-500 animate-pulse-glow`} />
                  </div>
                  <h4 className={`${titleSize} text-text-primary group-hover:text-brand-500 transition-colors duration-300`}>
                    {change.category}
                  </h4>
                  {comparisonMode && (
                    <div className="flex gap-2">
                      {isNewCategory && (
                        <Badge className="text-xs bg-gradient-to-r from-green-500 to-emerald-500 text-white border-0 animate-bounce-subtle">
                          ➕ New
                        </Badge>
                      )}
                      {isRemovedCategory && (
                        <Badge className="text-xs bg-gradient-to-r from-red-500 to-rose-500 text-white border-0 animate-shake">
                          ➖ Removed
                        </Badge>
                      )}
                    </div>
                  )}
                </div>
                <ul className={`list-none text-text-secondary ${listSpacing} ${listMargin} space-y-3`}>
                  {change.items.map((item: string, itemIndex: number) => (
                    <li key={itemIndex} className="flex items-start gap-3 group/item">
                      <div className="w-1.5 h-1.5 bg-brand-500 rounded-full mt-2.5 flex-shrink-0 group-hover/item:bg-purple-500 transition-colors duration-300 animate-pulse-glow"></div>
                      <span className="leading-relaxed group-hover/item:text-text-primary transition-colors duration-300">{item}</span>
                    </li>
                  ))}
                </ul>
              </div>
            );
          })}
        </div>
      </CardContent>
    </Card>
  );
};

export default ChangelogEntry;