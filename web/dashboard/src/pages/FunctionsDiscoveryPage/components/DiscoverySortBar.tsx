import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { cn } from '@/lib/utils';
import { ArrowUpDown } from 'lucide-react';
import { DISCOVERY_SORT_OPTIONS, type DiscoverySortBy } from '../utils';

interface DiscoverySortBarProps {
  value: DiscoverySortBy;
  onChange: (sort: DiscoverySortBy) => void;
  isNewStyle?: boolean;
  className?: string;
}

export function DiscoverySortBar({
  value,
  onChange,
  isNewStyle = false,
  className,
}: DiscoverySortBarProps) {
  return (
    <div className={cn('discovery-sort-bar', isNewStyle && 'discovery-sort-bar-new', className)}>
      <ArrowUpDown className="discovery-sort-icon" aria-hidden />
      <span className="discovery-sort-label">Sort by</span>
      <Select value={value} onValueChange={(next) => onChange(next as DiscoverySortBy)}>
        <SelectTrigger className="discovery-sort-select">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {DISCOVERY_SORT_OPTIONS.map((option) => (
            <SelectItem key={option.value} value={option.value}>
              {option.label}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}
