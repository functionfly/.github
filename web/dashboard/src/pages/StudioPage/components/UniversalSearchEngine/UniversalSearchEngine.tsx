import { useState, useEffect, useMemo } from "react";
import { GlassCard, Badge, Button } from "@functionfly/ui-core";
import { Input } from "@/components/ui/input";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { cn } from "@/lib/utils";
import {
  Search, FileText, Code, Settings, Hash, Clock, Star, StarOff,
  ArrowUp, ArrowDown, ChevronRight, Filter, History, Sparkles, X
} from "lucide-react";

interface SearchResult {
  id: string;
  type: "graph" | "node" | "plugin" | "setting" | "doc";
  title: string;
  description: string;
  path?: string;
  relevance: number;
  recent?: boolean;
}

export interface UniversalSearchEngineProps {
  onClose?: () => void;
  extraResults?: SearchResult[];
  onResultSelect?: (result: SearchResult) => void;
}

interface SearchCategory {
  id: string;
  name: string;
  icon: React.ReactNode;
}

const searchCategories: SearchCategory[] = [
  { id: "all", name: "All", icon: <Search className="w-4 h-4" /> },
  { id: "graphs", name: "Graphs", icon: <Sparkles className="w-4 h-4" /> },
  { id: "nodes", name: "Nodes", icon: <Code className="w-4 h-4" /> },
  { id: "plugins", name: "Plugins", icon: <Star className="w-4 h-4" /> },
  { id: "settings", name: "Settings", icon: <Settings className="w-4 h-4" /> },
  { id: "docs", name: "Docs", icon: <FileText className="w-4 h-4" /> },
];

const mockResults: SearchResult[] = [
  { id: "r1", type: "graph", title: "Customer Churn Prediction", description: "ML workflow for predicting customer churn", path: "Graphs/ML/customer-churn.flow", relevance: 0.95, recent: true },
  { id: "r2", type: "node", title: "HTTP Request Node", description: "Make HTTP requests to external APIs", path: "Nodes/http-request.ts", relevance: 0.9 },
  { id: "r3", type: "graph", title: "Image Classification Pipeline", description: "CNN-based image classification workflow", path: "Graphs/ML/image-classifier.flow", relevance: 0.85, recent: true },
  { id: "r4", type: "plugin", title: "GitHub Integration", description: "Connect to GitHub repositories and workflows", relevance: 0.8 },
  { id: "r5", type: "setting", title: "Theme Settings", description: "Customize studio appearance", path: "Settings/Appearance/Theme", relevance: 0.75 },
  { id: "r6", type: "doc", title: "Getting Started with Graphs", description: "Learn how to create and manage graphs", path: "Documentation/guides/graphs.md", relevance: 0.7 },
  { id: "r7", type: "node", title: "Data Transform Node", description: "Transform and reshape data", path: "Nodes/data-transform.ts", relevance: 0.65 },
  { id: "r8", type: "setting", title: "Keyboard Shortcuts", description: "View all keyboard shortcuts", path: "Settings/Shortcuts", relevance: 0.6, recent: true },
];

const recentSearches = ["customer churn", "GitHub integration", "theme settings", "HTTP request"];

export function UniversalSearchEngine({
  onClose: _onClose,
  extraResults: _extraResults,
  onResultSelect: _onResultSelect,
}: UniversalSearchEngineProps = {}) {
  const [query, setQuery] = useState("");
  const [activeCategory, setActiveCategory] = useState("all");
  const [showRecents, setShowRecents] = useState(true);
  const [favorites, setFavorites] = useState<Set<string>>(new Set(["r1", "r4"]));
  const [selectedResult, setSelectedResult] = useState<string | null>(null);
  const [searchHistory, setSearchHistory] = useState<string[]>(recentSearches);

  const results = useMemo(() => {
    if (!query.trim()) return [];
    return mockResults.filter((r) => {
      const matchesQuery =
        r.title.toLowerCase().includes(query.toLowerCase()) ||
        r.description.toLowerCase().includes(query.toLowerCase());
      const matchesCategory = activeCategory === "all" || r.type === activeCategory.slice(0, -1);
      return matchesQuery && matchesCategory;
    }).sort((a, b) => b.relevance - a.relevance);
  }, [query, activeCategory]);

  useEffect(() => {
    if (query.trim() && !searchHistory.includes(query)) {
      setSearchHistory((prev) => [query, ...prev.slice(0, 4)]);
    }
  }, [query]);

  const toggleFavorite = (id: string) => {
    setFavorites((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const getTypeIcon = (type: SearchResult["type"]) => {
    switch (type) {
      case "graph":
        return <Sparkles className="w-4 h-4 text-orange-400" />;
      case "node":
        return <Code className="w-4 h-4 text-blue-400" />;
      case "plugin":
        return <Star className="w-4 h-4 text-purple-400" />;
      case "setting":
        return <Settings className="w-4 h-4 text-emerald-400" />;
      case "doc":
        return <FileText className="w-4 h-4 text-yellow-400" />;
    }
  };

  const handleSearch = (searchQuery: string) => {
    setQuery(searchQuery);
    setShowRecents(false);
  };

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between p-5 border-b border-white/10">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-blue-500/20 flex items-center justify-center">
            <Search className="w-5 h-5 text-blue-400" />
          </div>
          <div>
            <h2 className="text-xl font-semibold text-white">Universal Search</h2>
            <p className="text-sm text-white/60">Find anything in your workspace</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <kbd className="px-2 py-1 text-xs text-white/60 bg-white/5 border border-white/10 rounded">
            ⌘K
          </kbd>
        </div>
      </div>

      <div className="px-5 pt-5">
        <div className="relative">
          <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-white/40" />
          <Input
            placeholder="Search graphs, nodes, plugins, settings..."
            value={query}
            onChange={(e) => {
              setQuery(e.target.value);
              setShowRecents(e.target.value === "");
            }}
            className="pl-12 pr-12 h-12 bg-white/5 border-white/10 text-white text-lg"
          />
          {query && (
            <button
              onClick={() => {
                setQuery("");
                setShowRecents(true);
              }}
              className="absolute right-4 top-1/2 -translate-y-1/2 text-white/40 hover:text-white/60"
            >
              <X className="w-4 h-4" />
            </button>
          )}
        </div>

        <div className="flex items-center gap-2 mt-4 overflow-x-auto pb-2">
          {searchCategories.map((cat) => (
            <button
              key={cat.id}
              onClick={() => setActiveCategory(cat.id)}
              className={cn(
                "flex items-center gap-2 px-3 py-1.5 rounded-lg border transition-all duration-200 whitespace-nowrap text-sm",
                activeCategory === cat.id
                  ? "bg-white/10 border-white/20 text-white"
                  : "bg-white/5 border-white/10 text-white/60 hover:bg-white/10 hover:text-white"
              )}
            >
              {cat.icon}
              {cat.name}
            </button>
          ))}
        </div>
      </div>

      <div className="flex-1 overflow-auto p-4">
        {showRecents && !query ? (
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-sm font-semibold text-white/60 flex items-center gap-2">
                <History className="w-4 h-4" />
                Recent Searches
              </h3>
            </div>
            <div className="space-y-1">
              {searchHistory.map((term, i) => (
                <button
                  key={i}
                  onClick={() => handleSearch(term)}
                  className="flex items-center gap-3 w-full p-3 rounded-lg text-left text-white/80 hover:bg-white/5 transition-colors"
                >
                  <Clock className="w-4 h-4 text-white/30" />
                  <span>{term}</span>
                </button>
              ))}
            </div>

            <h3 className="text-sm font-semibold text-white/60 mt-6 flex items-center gap-2">
              <Sparkles className="w-4 h-4" />
              Suggested
            </h3>
            <div className="grid grid-cols-2 gap-3">
              {mockResults.filter(r => r.recent).map((result) => (
                <GlassCard
                  key={result.id}
                  className="p-4 cursor-pointer hover:bg-white/10 transition-colors"
                  onClick={() => setSelectedResult(result.id)}
                >
                  <div className="flex items-start gap-3">
                    {getTypeIcon(result.type)}
                    <div className="flex-1 min-w-0">
                      <h4 className="font-medium text-white truncate">{result.title}</h4>
                      <p className="text-xs text-white/60 line-clamp-1">{result.description}</p>
                    </div>
                  </div>
                </GlassCard>
              ))}
            </div>
          </div>
        ) : results.length > 0 ? (
          <div className="space-y-2">
            <p className="text-sm text-white/60 mb-3">{results.length} results found</p>
            {results.map((result, index) => (
              <div
                key={result.id}
                className={cn(
                  "group p-4 rounded-xl border transition-all duration-200 cursor-pointer",
                  selectedResult === result.id
                    ? "bg-white/10 border-orange-500/30"
                    : "bg-white/5 border-white/10 hover:bg-white/10 hover:border-white/20"
                )}
                onClick={() => setSelectedResult(result.id)}
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex items-start gap-3 flex-1 min-w-0">
                    <div className="w-8 h-8 rounded-lg bg-white/5 flex items-center justify-center shrink-0">
                      {getTypeIcon(result.type)}
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <h3 className="font-medium text-white">{result.title}</h3>
                        <Badge variant="outline" className="text-[10px] text-white/50 border-white/20">
                          {result.type}
                        </Badge>
                        {result.recent && (
                          <Badge className="text-[10px] bg-blue-500/20 text-blue-400 border-blue-500/30">
                            Recent
                          </Badge>
                        )}
                      </div>
                      <p className="text-sm text-white/60 line-clamp-1">{result.description}</p>
                      {result.path && (
                        <p className="text-xs text-white/40 font-mono mt-1 truncate">{result.path}</p>
                      )}
                    </div>
                  </div>
                  <div className="flex items-center gap-2 shrink-0">
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        toggleFavorite(result.id);
                      }}
                      className="p-2 rounded-lg text-white/30 hover:text-yellow-400 hover:bg-white/5 transition-colors"
                    >
                      {favorites.has(result.id) ? (
                        <Star className="w-4 h-4 fill-yellow-400 text-yellow-400" />
                      ) : (
                        <StarOff className="w-4 h-4" />
                      )}
                    </button>
                    <ChevronRight className="w-4 h-4 text-white/30 group-hover:text-white/60" />
                  </div>
                </div>
              </div>
            ))}
          </div>
        ) : query ? (
          <GlassCard className="flex flex-col items-center justify-center h-48">
            <Search className="w-10 h-10 text-white/30 mb-3" />
            <p className="text-white/60">No results found for "{query}"</p>
            <p className="text-sm text-white/40">Try a different search term</p>
          </GlassCard>
        ) : null}
      </div>
    </div>
  );
}