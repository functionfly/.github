import { useEffect, useState } from 'react';
import { Search } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { GlobalCommandPalette } from './GlobalCommandPalette';

interface SearchButtonProps {
  className?: string;
}

export function SearchButton({ className }: SearchButtonProps) {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setOpen(true);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, []);

  return (
    <>
      <Button
        variant="ghost"
        size="icon"
        className={cn(
          'text-text-secondary hover:text-text-primary hover:bg-bg-hover transition-colors',
          className,
        )}
        onClick={() => setOpen(true)}
        title="Search (\u2318K)"
      >
        <Search className="w-5 h-5" />
      </Button>

      <GlobalCommandPalette open={open} onOpenChange={setOpen} />
    </>
  );
}
