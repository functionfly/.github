import { useState, useEffect, useRef, useMemo } from "react";
import { Search, X, FileText, ArrowRight, CornerDownLeft } from "lucide-react";
import { searchPages, getAllPages } from "../data/docs";

interface DocsSearchProps {
  onClose: () => void;
  onSelect: (slug: string) => void;
}

export function DocsSearch({ onClose, onSelect }: DocsSearchProps) {
  const [query, setQuery] = useState("");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const inputRef = useRef<HTMLInputElement>(null);

  const results = useMemo(() => {
    if (!query.trim()) return [];
    return searchPages(query).slice(0, 8);
  }, [query]);

  // Focus input on mount
  useEffect(() => {
    inputRef.current?.focus();
  }, []);

  // Reset selection when results change
  useEffect(() => {
    setSelectedIndex(0);
  }, [results.length]);

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "ArrowDown") {
        e.preventDefault();
        setSelectedIndex((prev) =>
          prev < results.length - 1 ? prev + 1 : prev
        );
      } else if (e.key === "ArrowUp") {
        e.preventDefault();
        setSelectedIndex((prev) => (prev > 0 ? prev - 1 : 0));
      } else if (e.key === "Enter") {
        e.preventDefault();
        const selected = results[selectedIndex];
        if (selected) {
          onSelect(selected.slug);
          onClose();
        }
      } else if (e.key === "Escape") {
        onClose();
      }
    };

    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [results, selectedIndex, onSelect, onClose]);

  const handleSelect = (slug: string) => {
    onSelect(slug);
    onClose();
  };

  return (
    <div
      className="fixed inset-0 z-50 flex items-start justify-center pt-[20vh] p-4"
      onClick={onClose}
    >
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />

      {/* Modal */}
      <div
        className="relative w-full max-w-2xl bg-bg-secondary rounded-xl shadow-2xl border border-border-default overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Search Input */}
        <div className="flex items-center gap-3 px-4 py-4 border-b border-border-subtle">
          <Search className="w-5 h-5 text-text-muted" />
          <input
            ref={inputRef}
            type="text"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="Search documentation..."
            className="flex-1 bg-transparent text-text-primary placeholder:text-text-muted text-lg outline-none"
          />
          {query && (
            <button
              onClick={() => setQuery("")}
              className="p-1 rounded-md hover:bg-bg-tertiary text-text-muted transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          )}
          <kbd className="hidden sm:flex items-center gap-1 px-2 py-1 rounded text-xs font-mono bg-bg-tertiary text-text-muted border border-border-subtle">
            ESC
          </kbd>
        </div>

        {/* Results */}
        <div className="max-h-[60vh] overflow-y-auto">
          {results.length > 0 ? (
            <div className="py-2">
              <div className="px-4 py-2 text-xs font-semibold text-text-muted uppercase tracking-wider">
                Results
              </div>
              {results.map((page, index) => {
                const isSelected = index === selectedIndex;

                return (
                  <button
                    key={page.slug}
                    onClick={() => handleSelect(page.slug)}
                    onMouseEnter={() => setSelectedIndex(index)}
                    className={`
                      w-full flex items-start gap-3 px-4 py-3 text-left
                      transition-colors duration-150
                      ${isSelected
                        ? "bg-brand-500/10 border-l-2 border-brand-500"
                        : "hover:bg-bg-tertiary border-l-2 border-transparent"
                      }
                    `}
                  >
                    <FileText className={`
                      w-5 h-5 mt-0.5 flex-shrink-0
                      ${isSelected ? "text-brand-500" : "text-text-muted"}
                    `} />
                    <div className="flex-1 min-w-0">
                      <div className={`
                        font-medium truncate
                        ${isSelected ? "text-brand-500" : "text-text-primary"}
                      `}>
                        {page.title}
                      </div>
                      <div className="text-sm text-text-muted line-clamp-1">
                        {page.description}
                      </div>
                    </div>
                    {isSelected && (
                      <CornerDownLeft className="w-4 h-4 text-brand-500 flex-shrink-0" />
                    )}
                  </button>
                );
              })}
            </div>
          ) : query ? (
            <div className="py-12 text-center">
              <div className="w-12 h-12 mx-auto mb-4 rounded-full bg-bg-tertiary flex items-center justify-center">
                <Search className="w-6 h-6 text-text-muted" />
              </div>
              <p className="text-text-secondary">No results found for "{query}"</p>
              <p className="text-sm text-text-muted mt-1">
                Try a different search term
              </p>
            </div>
          ) : (
            <div className="py-8">
              <div className="px-4 py-2 text-xs font-semibold text-text-muted uppercase tracking-wider">
                Popular Pages
              </div>
              {getAllPages().slice(0, 5).map((page, index) => (
                <button
                  key={page.slug}
                  onClick={() => handleSelect(page.slug)}
                  onMouseEnter={() => setSelectedIndex(index)}
                  className={`
                    w-full flex items-start gap-3 px-4 py-3 text-left
                    transition-colors duration-150
                    ${index === selectedIndex
                      ? "bg-brand-500/10 border-l-2 border-brand-500"
                      : "hover:bg-bg-tertiary border-l-2 border-transparent"
                    }
                  `}
                >
                  <FileText className="w-5 h-5 mt-0.5 text-text-muted flex-shrink-0" />
                  <div className="flex-1 min-w-0">
                    <div className="font-medium text-text-primary truncate">
                      {page.title}
                    </div>
                    <div className="text-sm text-text-muted line-clamp-1">
                      {page.description}
                    </div>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-4 py-3 border-t border-border-subtle bg-bg-primary/50">
          <div className="flex items-center gap-4 text-xs text-text-muted">
            <div className="flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 rounded bg-bg-tertiary border border-border-subtle">
                ↑
              </kbd>
              <kbd className="px-1.5 py-0.5 rounded bg-bg-tertiary border border-border-subtle">
                ↓
              </kbd>
              <span>to navigate</span>
            </div>
            <div className="flex items-center gap-1">
              <kbd className="px-1.5 py-0.5 rounded bg-bg-tertiary border border-border-subtle">
                ↵
              </kbd>
              <span>to select</span>
            </div>
          </div>
          <div className="text-xs text-text-muted">
            {results.length} result{results.length !== 1 ? 's' : ''}
          </div>
        </div>
      </div>
    </div>
  );
}
