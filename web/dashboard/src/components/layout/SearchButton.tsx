import { useEffect, useState } from "react";
import { Search } from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { GlobalSearch } from "./GlobalSearch";

interface SearchButtonProps {
  className?: string;
}

export function SearchButton({ className }: SearchButtonProps) {
  const [isSearchOpen, setIsSearchOpen] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Cmd+K or Ctrl+K
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        setIsSearchOpen(true);
      }

      // Escape to close
      if (e.key === 'Escape' && isSearchOpen) {
        setIsSearchOpen(false);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isSearchOpen]);

  return (
    <>
      <Button
        variant="ghost"
        size="icon"
        className={cn(
          "text-text-secondary hover:text-text-primary hover:bg-bg-hover transition-colors",
          className
        )}
        onClick={() => setIsSearchOpen(true)}
        title="Search (⌘K)"
      >
        <Search className="w-5 h-5" />
      </Button>

      <GlobalSearch
        isOpen={isSearchOpen}
        onClose={() => setIsSearchOpen(false)}
      />
    </>
  );
}