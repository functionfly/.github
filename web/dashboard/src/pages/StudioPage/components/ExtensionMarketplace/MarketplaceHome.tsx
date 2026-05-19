import { useState } from "react";
import { GlassCard, Badge, Button, Spinner, Input } from "@functionfly/ui-core";
import { Search, Download, Star, Shield, Clock, TrendingUp, Filter, Grid, List, ExternalLink } from "lucide-react";

interface MarketplaceExtension {
  id: string;
  name: string;
  version: string;
  description: string;
  author: string;
  category: string;
  downloads: number;
  rating: number;
  verified: boolean;
  trust_score: number;
}

interface MarketplaceHomeProps {
  onInstall?: (extension: MarketplaceExtension) => void;
}

const featuredExtensions: MarketplaceExtension[] = [
  {
    id: "ext-1",
    name: "GitHub Integration Pack",
    version: "2.0.0",
    description: "Complete GitHub workflow integration with PR reviews, issue management, automated deployments, and AI-powered code review.",
    author: "FunctionFly Labs",
    category: "integrations",
    downloads: 45200,
    rating: 4.9,
    verified: true,
    trust_score: 98,
  },
  {
    id: "ext-2",
    name: "Neural Process Visualizer",
    version: "1.5.0",
    description: "Real-time visualization of AI agent thought processes and neural network activations during execution.",
    author: "AI Visualizers Inc",
    category: "visualization",
    downloads: 28900,
    rating: 4.7,
    verified: true,
    trust_score: 95,
  },
  {
    id: "ext-3",
    name: "CloudOps Commander",
    version: "3.2.0",
    description: "Multi-cloud infrastructure management with automated scaling, monitoring, and incident response.",
    author: "CloudOps Pro",
    category: "infrastructure",
    downloads: 62100,
    rating: 4.8,
    verified: true,
    trust_score: 99,
  },
  {
    id: "ext-4",
    name: "DataFlow Pipeline",
    version: "1.8.0",
    description: "High-performance data processing pipelines with support for streaming, batch, and real-time analytics.",
    author: "DataEng Solutions",
    category: "data",
    downloads: 34500,
    rating: 4.6,
    verified: true,
    trust_score: 92,
  },
  {
    id: "ext-5",
    name: "Security Sentinel",
    version: "2.1.0",
    description: "Enterprise-grade security monitoring with threat detection, vulnerability scanning, and compliance reporting.",
    author: "SecureStack",
    category: "security",
    downloads: 41200,
    rating: 4.9,
    verified: true,
    trust_score: 99,
  },
  {
    id: "ext-6",
    name: "Slack Bridge",
    version: "1.3.0",
    description: "Seamless Slack integration for notifications, commands, and collaborative workflows.",
    author: "DevTools Co",
    category: "integrations",
    downloads: 19800,
    rating: 4.5,
    verified: false,
    trust_score: 85,
  },
];

const categories = [
  { id: "all", label: "All", icon: "📦", count: 156 },
  { id: "integrations", label: "Integrations", icon: "🔗", count: 42 },
  { id: "visualization", label: "Visualization", icon: "📊", count: 28 },
  { id: "infrastructure", label: "Infrastructure", icon: "🏗️", count: 35 },
  { id: "data", label: "Data & Analytics", icon: "📈", count: 31 },
  { id: "security", label: "Security", icon: "🔒", count: 24 },
  { id: "ai_agents", label: "AI Agents", icon: "🤖", count: 48 },
  { id: "workflows", label: "Workflows", icon: "⚡", count: 39 },
];

export function MarketplaceHome({ onInstall }: MarketplaceHomeProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("all");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [installingId, setInstallingId] = useState<string | null>(null);

  const filteredExtensions = featuredExtensions.filter((ext) => {
    const matchesSearch =
      ext.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      ext.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      ext.author.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesCategory = selectedCategory === "all" || ext.category === selectedCategory;
    return matchesSearch && matchesCategory;
  });

  const handleInstall = async (ext: MarketplaceExtension) => {
    setInstallingId(ext.id);
    try {
      await onInstall?.(ext);
    } finally {
      setInstallingId(null);
    }
  };

  const formatNumber = (n: number) => {
    if (n >= 1000) return (n / 1000).toFixed(1) + "k";
    return n.toString();
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-white/40" />
          <Input
            placeholder="Search extensions..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9 bg-white/5 border-white/10"
          />
        </div>
        <div className="flex items-center gap-1">
          <Button size="icon" variant={viewMode === "grid" ? "secondary" : "ghost"} onClick={() => setViewMode("grid")}>
            <Grid className="w-4 h-4" />
          </Button>
          <Button size="icon" variant={viewMode === "list" ? "secondary" : "ghost"} onClick={() => setViewMode("list")}>
            <List className="w-4 h-4" />
          </Button>
        </div>
      </div>

      <div className="flex gap-2 overflow-x-auto pb-2">
        {categories.map((cat) => (
          <button
            key={cat.id}
            onClick={() => setSelectedCategory(cat.id)}
            className={`flex items-center gap-2 px-4 py-2 rounded-lg border transition-colors whitespace-nowrap ${
              selectedCategory === cat.id
                ? "bg-white/10 border-white/30 text-white"
                : "bg-white/5 border-white/10 text-white/60 hover:bg-white/10"
            }`}
          >
            <span>{cat.icon}</span>
            <span>{cat.label}</span>
            <Badge variant="outline" className="ml-1 text-xs">
              {cat.count}
            </Badge>
          </button>
        ))}
      </div>

      <div className="flex items-center justify-between">
        <p className="text-sm text-white/60">
          {filteredExtensions.length} extensions{selectedCategory !== "all" && ` in ${categories.find((c) => c.id === selectedCategory)?.label}`}
        </p>
        <div className="flex items-center gap-2">
          <Badge variant="outline" className="text-white/60 border-white/20">
            <TrendingUp className="w-3 h-3 mr-1" />
            Trending
          </Badge>
        </div>
      </div>

      <div className={viewMode === "grid" ? "grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4" : "space-y-4"}>
        {filteredExtensions.map((ext) => (
          <GlassCard key={ext.id} className={viewMode === "grid" ? "p-4" : "p-4"}>
            {viewMode === "grid" ? (
              <div className="flex flex-col h-full">
                <div className="flex items-start justify-between">
                  <div className="flex items-center gap-3">
                    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-purple-500/30 to-blue-500/30 flex items-center justify-center text-xl font-bold text-white/80">
                      {ext.name.charAt(0)}
                    </div>
                    <div>
                      <div className="flex items-center gap-2">
                        <h3 className="font-semibold text-white">{ext.name}</h3>
                        {ext.verified && <Shield className="w-4 h-4 text-green-400" />}
                      </div>
                      <p className="text-xs text-white/60">by {ext.author}</p>
                    </div>
                  </div>
                </div>

                <p className="mt-3 text-sm text-white/60 line-clamp-2 flex-1">{ext.description}</p>

                <div className="mt-4 flex items-center gap-2 flex-wrap">
                  <Badge className="text-xs bg-white/10 text-white/80 border-white/20">{ext.category}</Badge>
                  <div className="flex items-center gap-1 text-xs text-white/60">
                    <Download className="w-3 h-3" />
                    {formatNumber(ext.downloads)}
                  </div>
                  <div className="flex items-center gap-1 text-xs text-yellow-400">
                    <Star className="w-3 h-3 fill-current" />
                    {ext.rating}
                  </div>
                  <div className="ml-auto text-xs text-white/40">v{ext.version}</div>
                </div>

                <Button
                  size="sm"
                  className="w-full mt-4"
                  variant="default"
                  disabled={installingId === ext.id}
                  onClick={() => handleInstall(ext)}
                >
                  {installingId === ext.id ? (
                    <Spinner className="w-4 h-4" />
                  ) : (
                    <>
                      <Download className="w-4 h-4 mr-1" />
                      Install
                    </>
                  )}
                </Button>
              </div>
            ) : (
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-purple-500/30 to-blue-500/30 flex items-center justify-center text-2xl font-bold text-white/80">
                    {ext.name.charAt(0)}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="font-semibold text-white">{ext.name}</h3>
                      {ext.verified && <Shield className="w-4 h-4 text-green-400" />}
                      <Badge className="text-xs bg-white/10 text-white/80 border-white/20">{ext.category}</Badge>
                    </div>
                    <p className="text-sm text-white/60 mt-1">{ext.description}</p>
                    <div className="flex items-center gap-4 mt-2">
                      <span className="text-xs text-white/40">by {ext.author}</span>
                      <span className="text-xs text-white/60">
                        <Download className="w-3 h-3 inline mr-1" />
                        {formatNumber(ext.downloads)}
                      </span>
                      <span className="text-xs text-yellow-400">
                        <Star className="w-3 h-3 inline mr-1 fill-current" />
                        {ext.rating}
                      </span>
                      <span className="text-xs text-white/40">Trust: {ext.trust_score}%</span>
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    size="sm"
                    variant="default"
                    disabled={installingId === ext.id}
                    onClick={() => handleInstall(ext)}
                  >
                    {installingId === ext.id ? <Spinner className="w-4 h-4" /> : "Install"}
                  </Button>
                  <Button size="icon" variant="ghost">
                    <ExternalLink className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            )}
          </GlassCard>
        ))}
      </div>

      {filteredExtensions.length === 0 && (
        <div className="flex flex-col items-center justify-center h-64 text-white/40">
          <p>No extensions found matching your criteria</p>
          <p className="text-sm mt-1">Try adjusting your search or filters</p>
        </div>
      )}
    </div>
  );
}