import { useState, useEffect, useCallback } from "react";
import { useNavigate } from "react-router-dom";
import {
  Search,
  Star,
  Download,
  Code,
  SortAsc,
  SortDesc,
  Grid,
  List,
  Play,
  LogIn,
  Zap,
  Database,
  Shield,
  Mail,
  FileText,
  Image,
  Brain,
  CreditCard,
  Wrench,
  Globe,
  GitBranch,
  Layers,
  Activity,
  TrendingUp,
  Filter,
  X,
  ChevronRight,
  Boxes,
  Sparkles,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Navbar } from "@/components/common/Navbar";
import { Footer } from "@/pages/LandingPage/components";
import { useAuthStore } from "@/stores/authStore";
import { registryApi, type RegistryFunction } from "@/api/registry";
import { useRealtime } from "@/hooks/useRealtime";
import { MetaTags } from "@/components/seo/MetaTags";

/* ────────────────────────────────────────────────────────────── */
/*  Constants                                                      */
/* ────────────────────────────────────────────────────────────── */

const CATEGORIES = [
  "All Categories",
  "API Tools",
  "Authentication",
  "Database",
  "Email",
  "File Processing",
  "Image Processing",
  "Machine Learning",
  "Payment",
  "Utility",
  "Web Scraping",
  "Workflow",
];

const SORT_OPTIONS = [
  { value: "popularity", label: "Popularity" },
  { value: "rating", label: "Rating" },
  { value: "reliability", label: "Reliability" },
  { value: "newest", label: "Newest" },
  { value: "name", label: "Name" },
];

/** Map each category to a lucide icon + a CSS color class */
const CATEGORY_META: Record<string, { icon: React.ElementType; colorClass: string }> = {
  "All Categories":   { icon: Boxes,      colorClass: "registry-cat-color-indigo"  },
  "API Tools":        { icon: Zap,        colorClass: "registry-cat-color-indigo"  },
  "Authentication":   { icon: Shield,     colorClass: "registry-cat-color-violet"  },
  "Database":         { icon: Database,   colorClass: "registry-cat-color-sky"     },
  "Email":            { icon: Mail,       colorClass: "registry-cat-color-amber"   },
  "File Processing":  { icon: FileText,   colorClass: "registry-cat-color-teal"    },
  "Image Processing": { icon: Image,      colorClass: "registry-cat-color-fuchsia" },
  "Machine Learning": { icon: Brain,      colorClass: "registry-cat-color-violet"  },
  "Payment":          { icon: CreditCard, colorClass: "registry-cat-color-emerald" },
  "Utility":          { icon: Wrench,     colorClass: "registry-cat-color-cyan"    },
  "Web Scraping":     { icon: Globe,      colorClass: "registry-cat-color-rose"    },
  "Workflow":         { icon: GitBranch,  colorClass: "registry-cat-color-orange"  },
};

/* ────────────────────────────────────────────────────────────── */
/*  Helpers                                                        */
/* ────────────────────────────────────────────────────────────── */

function formatScore(score: number | undefined | null): string {
  return typeof score === "number" && !Number.isNaN(score)
    ? score.toFixed(1)
    : "0.0";
}

function reliabilityLabel(score: number): { label: string; cls: string } {
  if (score >= 0.9) return { label: "High", cls: "registry-reliability-high" };
  if (score >= 0.6) return { label: "Med",  cls: "registry-reliability-mid"  };
  return { label: "Low", cls: "registry-reliability-low" };
}

function getColorClass(category: string | undefined): string {
  return CATEGORY_META[category ?? ""]?.colorClass ?? "registry-cat-color-indigo";
}

/* ────────────────────────────────────────────────────────────── */
/*  Sub-components                                                 */
/* ────────────────────────────────────────────────────────────── */

/** Animated stat counter */
function StatItem({
  value,
  label,
  icon: Icon,
}: {
  value: string | number;
  label: string;
  icon: React.ElementType;
}) {
  return (
    <div className="registry-stat-item">
      <div className="flex items-center gap-1.5">
        <Icon className="h-3.5 w-3.5 text-brand-500 opacity-70 shrink-0" />
        <span className="registry-stat-value">{value}</span>
      </div>
      <span className="registry-stat-label">{label}</span>
    </div>
  );
}

/** Category sidebar item */
function CategoryItem({
  category,
  count,
  active,
  onClick,
}: {
  category: string;
  count: number;
  active: boolean;
  onClick: () => void;
}) {
  const meta = CATEGORY_META[category] ?? { icon: Code, colorClass: "registry-cat-color-indigo" };
  const Icon = meta.icon;
  return (
    <button
      type="button"
      className={`registry-cat-item ${meta.colorClass} ${active ? "active" : ""}`}
      onClick={onClick}
    >
      <span className="registry-cat-icon">
        <Icon className="h-3.5 w-3.5" />
      </span>
      <span className="registry-cat-label">{category}</span>
      <span className="registry-cat-count">{count}</span>
    </button>
  );
}

/** Mobile horizontal pill for categories */
function MobileChip({
  category,
  active,
  onClick,
}: {
  category: string;
  active: boolean;
  onClick: () => void;
}) {
  const meta = CATEGORY_META[category] ?? { icon: Code, colorClass: "registry-cat-color-indigo" };
  const Icon = meta.icon;
  return (
    <button
      type="button"
      className={`registry-mobile-chip ${active ? "active" : ""}`}
      onClick={onClick}
    >
      <Icon className="h-3 w-3 shrink-0" />
      {category === "All Categories" ? "All" : category}
    </button>
  );
}

/** Premium grid card */
function GridCard({
  fn,
  isAuthenticated,
  onView,
  onDeploy,
  onTry,
  flashId,
}: {
  fn: RegistryFunction;
  isAuthenticated: boolean;
  onView: () => void;
  onDeploy: () => void;
  onTry: () => void;
  flashId: string | null;
}) {
  const colorClass = getColorClass(fn.category);
  const rating = Number(fn.overall_score ?? 0);
  const ratingPct = Math.min(100, (rating / 5) * 100);
  const reliability = fn.reliability_score ?? 0;
  const rel = reliabilityLabel(reliability);
  const isFlashing = flashId === fn.id;

  return (
    <div className={`registry-card ${isFlashing ? "registry-card-flash" : ""}`}>
      <div className="registry-card-accent" />

      {/* Card body */}
      <div className="p-4 flex flex-col h-full gap-3">

        {/* Header row */}
        <div className="flex items-start justify-between gap-2">
          <div className="flex items-center gap-2.5 min-w-0">
            <div
              className={`registry-card-icon ${colorClass}`}
              style={{
                background: "var(--cat-bg)",
                border: "1px solid color-mix(in srgb, var(--cat-color) 25%, transparent)",
              }}
            >
              <Code className="h-5 w-5" style={{ color: "var(--cat-color)" }} />
            </div>
            <div className="min-w-0">
              <h3 className="text-sm font-bold text-text-primary truncate leading-tight">
                {fn.name}
              </h3>
              <p className="text-xs text-text-muted truncate">@{fn.author}</p>
            </div>
          </div>

          {fn.category && (
            <span className={`registry-cat-badge ${colorClass} shrink-0`}>
              <span className="registry-cat-dot" />
              {fn.category}
            </span>
          )}
        </div>

        {/* Description */}
        <p className="text-xs text-text-secondary line-clamp-2 leading-relaxed flex-grow">
          {fn.description || fn.title || "No description available."}
        </p>

        {/* Tags */}
        {fn.tags && fn.tags.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {fn.tags.slice(0, 3).map((tag, i) => (
              <span
                key={i}
                className="inline-flex items-center px-1.5 py-0.5 rounded-md text-[0.62rem] font-medium
                           bg-white/[0.04] border border-white/[0.07] text-text-muted"
              >
                {tag}
              </span>
            ))}
            {fn.tags.length > 3 && (
              <span
                className="inline-flex items-center px-1.5 py-0.5 rounded-md text-[0.62rem] font-medium
                           bg-white/[0.04] border border-white/[0.07] text-text-muted"
              >
                +{fn.tags.length - 3}
              </span>
            )}
          </div>
        )}

        {/* Stats row */}
        <div className="flex items-center justify-between pt-1 border-t border-white/[0.05]">
          <div className="flex items-center gap-1.5">
            <Star className="h-3.5 w-3.5 fill-amber-400 text-amber-400 shrink-0" />
            <span className="text-xs font-bold text-text-primary">{formatScore(fn.overall_score)}</span>
            <div className="registry-rating-bar">
              <div className="registry-rating-fill" style={{ width: `${ratingPct}%` }} />
            </div>
            <span className="text-[0.62rem] text-text-muted">({fn.total_ratings ?? 0})</span>
          </div>

          <div className="flex items-center gap-2">
            <span className={`registry-reliability-badge ${rel.cls}`}>
              <Activity className="h-2.5 w-2.5" />
              {rel.label}
            </span>
            <span className="flex items-center gap-1 text-[0.65rem] text-text-muted">
              <Download className="h-3 w-3" />
              {Math.floor(Number(fn.popularity_score) || 0)}
            </span>
            {Number(fn.price_per_call) > 0 && (
              <span className="registry-price-badge">${fn.price_per_call}/call</span>
            )}
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-1.5 mt-auto pt-1">
          <button
            type="button"
            onClick={onView}
            className="flex-1 h-8 rounded-lg text-xs font-semibold border border-white/10
                       bg-white/[0.04] text-text-secondary hover:bg-white/[0.07] hover:text-text-primary
                       transition-all duration-150 flex items-center justify-center gap-1"
          >
            <ChevronRight className="h-3 w-3" />
            View
          </button>
          <button
            type="button"
            onClick={onDeploy}
            className="flex-1 h-8 rounded-lg text-xs font-semibold
                       bg-gradient-to-r from-brand-500 via-purple-500 to-pink-500
                       text-white hover:brightness-110 hover:scale-[1.02]
                       transition-all duration-150 flex items-center justify-center gap-1
                       shadow-lg shadow-brand-500/20"
          >
            {!isAuthenticated && <LogIn className="h-3 w-3" />}
            <Layers className="h-3 w-3" />
            Deploy
          </button>
          <button
            type="button"
            onClick={onTry}
            className="flex-1 h-8 rounded-lg text-xs font-semibold border border-brand-500/25
                       bg-brand-500/10 text-brand-400 hover:bg-brand-500/[0.18] hover:border-brand-500/40
                       transition-all duration-150 flex items-center justify-center gap-1"
          >
            {isAuthenticated ? (
              <Play className="h-3 w-3" />
            ) : (
              <LogIn className="h-3 w-3" />
            )}
            Try
          </button>
        </div>
      </div>
    </div>
  );
}

/** Enhanced list row */
function ListCard({
  fn,
  isAuthenticated,
  onView,
  onDeploy,
  onTry,
  flashId,
}: {
  fn: RegistryFunction;
  isAuthenticated: boolean;
  onView: () => void;
  onDeploy: () => void;
  onTry: () => void;
  flashId: string | null;
}) {
  const colorClass = getColorClass(fn.category);
  const isFlashing = flashId === fn.id;

  return (
    <div className={`registry-list-card ${colorClass} ${isFlashing ? "registry-card-flash" : ""}`}>
      {/* Icon */}
      <div
        className="w-10 h-10 rounded-xl flex items-center justify-center shrink-0"
        style={{
          background: "var(--cat-bg)",
          border: "1px solid color-mix(in srgb, var(--cat-color) 25%, transparent)",
        }}
      >
        <Code className="h-4 w-4" style={{ color: "var(--cat-color)" }} />
      </div>

      {/* Main info */}
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <span className="text-sm font-bold text-text-primary truncate">
            {fn.author}/{fn.name}
          </span>
          {fn.category && (
            <span className={`registry-cat-badge ${colorClass}`}>
              <span className="registry-cat-dot" />
              {fn.category}
            </span>
          )}
        </div>
        <p className="text-xs text-text-secondary line-clamp-1 mt-0.5">
          {fn.title || fn.description || "No description"}
        </p>
      </div>

      {/* Stats */}
      <div className="hidden sm:flex items-center gap-4 shrink-0">
        <div className="flex items-center gap-1 text-xs">
          <Star className="h-3.5 w-3.5 fill-amber-400 text-amber-400" />
          <span className="font-bold text-text-primary">{formatScore(fn.overall_score)}</span>
          <span className="text-text-muted">({fn.total_ratings ?? 0})</span>
        </div>
        <div className="flex items-center gap-1 text-xs text-text-muted">
          <Download className="h-3.5 w-3.5" />
          <span>{Math.floor(Number(fn.popularity_score) || 0)}</span>
        </div>
        {Number(fn.price_per_call) > 0 && (
          <span className="registry-price-badge">${fn.price_per_call}/call</span>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-1.5 shrink-0">
        <button
          type="button"
          onClick={onView}
          className="h-7 px-3 rounded-lg text-xs font-semibold border border-white/10
                     bg-white/[0.04] text-text-secondary hover:bg-white/[0.07] hover:text-text-primary
                     transition-all duration-150 flex items-center gap-1"
        >
          View
        </button>
        <button
          type="button"
          onClick={onDeploy}
          className="h-7 px-3 rounded-lg text-xs font-semibold
                     bg-gradient-to-r from-brand-500 to-purple-500 text-white
                     hover:brightness-110 transition-all duration-150 flex items-center gap-1"
        >
          {!isAuthenticated && <LogIn className="h-3 w-3" />}
          Deploy
        </button>
        <button
          type="button"
          onClick={onTry}
          className="h-7 px-3 rounded-lg text-xs font-semibold border border-brand-500/25
                     bg-brand-500/10 text-brand-400 hover:bg-brand-500/[0.18]
                     transition-all duration-150 flex items-center gap-1"
        >
          {isAuthenticated ? <Play className="h-3 w-3" /> : <LogIn className="h-3 w-3" />}
          Try
        </button>
      </div>
    </div>
  );
}

/** Loading skeleton grid */
function SkeletonGrid({ count = 6, mode }: { count?: number; mode: "grid" | "list" }) {
  if (mode === "list") {
    return (
      <div className="space-y-2">
        {Array.from({ length: count }).map((_, i) => (
          <div
            key={i}
            className="registry-list-card animate-fade-in-up"
            style={{ animationDelay: `${i * 60}ms`, opacity: 0 }}
          >
            <div className="w-10 h-10 rounded-xl bg-white/[0.04] shrink-0" />
            <div className="flex-1 space-y-2">
              <div className="h-3.5 w-40 rounded-md bg-white/[0.04]" />
              <div className="h-2.5 w-64 rounded-md bg-white/[0.03]" />
            </div>
            <div className="hidden sm:flex gap-3">
              <div className="h-4 w-16 rounded-md bg-white/[0.04]" />
              <div className="h-4 w-12 rounded-md bg-white/[0.03]" />
            </div>
            <div className="flex gap-1.5">
              <div className="h-7 w-14 rounded-lg bg-white/[0.04]" />
              <div className="h-7 w-16 rounded-lg bg-white/[0.04]" />
              <div className="h-7 w-12 rounded-lg bg-white/[0.04]" />
            </div>
          </div>
        ))}
      </div>
    );
  }
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="registry-skeleton-card animate-fade-in-up"
          style={{ animationDelay: `${i * 80}ms`, opacity: 0 }}
        />
      ))}
    </div>
  );
}

/* ────────────────────────────────────────────────────────────── */
/*  Main Page                                                      */
/* ────────────────────────────────────────────────────────────── */

export function BrowseFunctionsPage() {
  const navigate = useNavigate();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const { subscribe, unsubscribe } = useRealtime();

  const [functions, setFunctions] = useState<RegistryFunction[]>([]);
  const [filteredFunctions, setFilteredFunctions] = useState<RegistryFunction[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState("All Categories");
  const [sortBy, setSortBy] = useState("popularity");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("desc");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [showMobileFilters, setShowMobileFilters] = useState(false);
  const [flashId, setFlashId] = useState<string | null>(null);

  /* Realtime flash on rating update */
  useEffect(() => {
    const handleRegistryUpdate = (data: {
      event?: string;
      update_type?: string;
      function_id?: string;
      overall_score?: number;
      total_ratings?: number;
    }) => {
      if (data.event === "registry_update" && data.update_type === "rating") {
        setFunctions((prev) =>
          prev.map((fn) =>
            fn.id === data.function_id
              ? {
                  ...fn,
                  overall_score: data.overall_score ?? fn.overall_score,
                  total_ratings: data.total_ratings ?? fn.total_ratings,
                }
              : fn
          )
        );
        if (data.function_id) {
          setFlashId(data.function_id);
          setTimeout(() => setFlashId(null), 700);
        }
      }
    };
    subscribe("registry_updates", handleRegistryUpdate);
    return () => unsubscribe("registry_updates", handleRegistryUpdate);
  }, [subscribe, unsubscribe]);

  useEffect(() => {
    loadFunctions();
  }, []);

  /* Filter + sort */
  useEffect(() => {
    const q = searchQuery.toLowerCase();
    let filtered = functions.filter((fn) => {
      const matchesSearch =
        q === "" ||
        fn.name.toLowerCase().includes(q) ||
        fn.author.toLowerCase().includes(q) ||
        (fn.description?.toLowerCase().includes(q) ?? false);
      const matchesCategory =
        selectedCategory === "All Categories" || fn.category === selectedCategory;
      return matchesSearch && matchesCategory;
    });

    filtered.sort((a, b) => {
      let aVal: number | string, bVal: number | string;
      switch (sortBy) {
        case "popularity":  aVal = a.popularity_score; bVal = b.popularity_score; break;
        case "rating":      aVal = a.overall_score;    bVal = b.overall_score;    break;
        case "reliability": aVal = a.reliability_score; bVal = b.reliability_score; break;
        case "newest":
          aVal = new Date(a.created_at).getTime();
          bVal = new Date(b.created_at).getTime();
          break;
        case "name": aVal = a.name.toLowerCase(); bVal = b.name.toLowerCase(); break;
        default: return 0;
      }
      if (sortOrder === "asc") return aVal < bVal ? -1 : aVal > bVal ? 1 : 0;
      return aVal > bVal ? -1 : aVal < bVal ? 1 : 0;
    });

    setFilteredFunctions(filtered);
  }, [functions, searchQuery, selectedCategory, sortBy, sortOrder]);

  const loadFunctions = async () => {
    try {
      setLoading(true);
      const res = await registryApi.getFunctions({ limit: 500 });
      setFunctions(res.functions ?? []);
    } catch (e) {
      console.error("Failed to load registry functions:", e);
    } finally {
      setLoading(false);
    }
  };

  /* Category counts */
  const categoryCounts = useCallback((): Record<string, number> => {
    const counts: Record<string, number> = { "All Categories": functions.length };
    for (const fn of functions) {
      if (fn.category) counts[fn.category] = (counts[fn.category] ?? 0) + 1;
    }
    return counts;
  }, [functions]);
  const counts = categoryCounts();

  /* Nav helpers */
  const getLoginRedirect = (path: string) =>
    `/login?redirect=${encodeURIComponent(path)}`;

  const handleDeploy = (fn: RegistryFunction) => {
    const path = `/functions/deploy?registry=${fn.author}/${fn.name}`;
    if (!isAuthenticated) { navigate(getLoginRedirect(path)); return; }
    navigate(path);
  };

  const handleTry = (fn: RegistryFunction) => {
    const path = `/run/${fn.author}/${fn.name}`;
    if (!isAuthenticated) { navigate(getLoginRedirect(path)); return; }
    navigate(path);
  };

  const handleView = (fn: RegistryFunction) =>
    navigate(`/fx/${fn.author}/${fn.name}`);

  /* Derived values */
  const totalFunctions = functions.length;
  const activeCategories = Object.keys(counts).length - 1; // minus "All"
  const hasActiveFilters =
    searchQuery !== "" || selectedCategory !== "All Categories";

  /* ─── Render ─── */
  return (
    <div className="min-h-screen bg-bg-primary flex flex-col">
      <MetaTags
        title="Browse Functions | Registry"
        description="Discover and explore premium serverless functions. Browse the registry, deploy instantly, or try live in the playground."
        keywords={["function registry", "serverless", "browse functions", "deploy functions"]}
      />
      <Navbar variant="landing" />

      <main className="flex-1 pt-16">

        {/* ══════════════════════════════════════════════════════
            HERO
        ══════════════════════════════════════════════════════ */}
        <section className="registry-hero section-padding">
          {/* Decorative orbs */}
          <div className="registry-orb registry-orb-1" aria-hidden="true" />
          <div className="registry-orb registry-orb-2" aria-hidden="true" />
          <div className="registry-orb registry-orb-3" aria-hidden="true" />

          <div className="container-wide px-4 lg:px-6">
            <div className="max-w-3xl animate-stagger">

              {/* Label pill + live badge */}
              <div className="flex items-center gap-3 mb-5">
                <span className="registry-label-pill">
                  <Sparkles className="h-3 w-3" />
                  Registry
                </span>
                <span className="registry-live-badge">
                  <span className="registry-live-dot" />
                  Live
                </span>
              </div>

              {/* Title */}
              <h1 className="registry-hero-title mb-4">
                Discover<br />
                <span style={{ opacity: 0.7 }}>Functions</span>
              </h1>

              {/* Subtitle */}
              <p className="text-base md:text-lg text-text-secondary mb-7 max-w-xl leading-relaxed">
                Browse community-built serverless functions. View docs and
                examples freely — sign in to deploy or run them live.
              </p>

              {/* Auth banner */}
              {!isAuthenticated && (
                <div className="registry-auth-banner mb-8">
                  <LogIn className="h-4 w-4 text-amber-400 shrink-0" />
                  <span className="text-sm text-amber-400/90">
                    Sign in to deploy or run functions.
                  </span>
                  <div className="flex items-center gap-2 ml-auto">
                    <Button
                      size="sm"
                      variant="outline"
                      className="h-7 text-xs border-amber-500/30 text-amber-400 hover:bg-amber-500/10"
                      onClick={() => navigate("/login")}
                    >
                      Sign in
                    </Button>
                    <Button
                      size="sm"
                      className="h-7 text-xs bg-amber-500 hover:bg-amber-400 text-black font-semibold"
                      onClick={() => navigate("/signup")}
                    >
                      Sign up free
                    </Button>
                  </div>
                </div>
              )}

              {/* Stats bar */}
              <div className="registry-stats-bar max-w-md">
                <StatItem
                  value={loading ? "—" : totalFunctions.toLocaleString()}
                  label="Functions"
                  icon={Code}
                />
                <StatItem
                  value={loading ? "—" : activeCategories}
                  label="Categories"
                  icon={Layers}
                />
                <StatItem
                  value={loading ? "—" : filteredFunctions.length.toLocaleString()}
                  label="Visible"
                  icon={TrendingUp}
                />
              </div>
            </div>
          </div>
        </section>

        {/* ══════════════════════════════════════════════════════
            CONTENT — Sidebar + Main
        ══════════════════════════════════════════════════════ */}
        <section className="py-8">
          <div className="container-wide px-4 lg:px-6">
            <div className="registry-layout">

              {/* ─── Sidebar (desktop) ──────────────────────────── */}
              <aside className="hidden lg:block">
                <div className="registry-sidebar">
                  <div className="registry-sidebar-header">
                    <Filter className="h-3.5 w-3.5 text-text-muted" />
                    <span className="registry-sidebar-title">Categories</span>
                  </div>
                  <div className="registry-sidebar-list">
                    {CATEGORIES.map((cat) => (
                      <CategoryItem
                        key={cat}
                        category={cat}
                        count={counts[cat] ?? 0}
                        active={selectedCategory === cat}
                        onClick={() => setSelectedCategory(cat)}
                      />
                    ))}
                  </div>
                </div>
              </aside>

              {/* ─── Main content ───────────────────────────────── */}
              <div className="min-w-0 space-y-4">

                {/* Mobile category chips */}
                <div className="lg:hidden registry-mobile-cats">
                  {CATEGORIES.map((cat) => (
                    <MobileChip
                      key={cat}
                      category={cat}
                      active={selectedCategory === cat}
                      onClick={() => setSelectedCategory(cat)}
                    />
                  ))}
                </div>

                {/* Search + controls */}
                <div className="space-y-3">
                  <div className="flex gap-3">
                    {/* Search */}
                    <div className="registry-search-wrapper flex-1">
                      <div className="relative registry-search-input">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-text-muted pointer-events-none" />
                        <Input
                          placeholder="Search by name, author, or description…"
                          value={searchQuery}
                          onChange={(e) => setSearchQuery(e.target.value)}
                          className="pl-10 pr-16 bg-bg-secondary border-border-subtle h-10 rounded-xl focus-visible:ring-0"
                        />
                        <div className="absolute right-3 top-1/2 -translate-y-1/2 flex items-center pointer-events-none">
                          <span className="registry-search-shortcut">⌘K</span>
                        </div>
                      </div>
                    </div>

                    {/* Mobile sort toggle */}
                    <button
                      type="button"
                      className="lg:hidden flex items-center gap-1.5 h-10 px-3 rounded-xl text-sm font-medium
                                 border border-border-subtle bg-bg-secondary text-text-secondary
                                 hover:text-text-primary hover:border-brand-500/30 transition-all"
                      onClick={() => setShowMobileFilters(!showMobileFilters)}
                    >
                      <Filter className="h-4 w-4" />
                      Sort
                    </button>

                    {/* View toggle */}
                    <div className="registry-view-toggle">
                      <button
                        type="button"
                        className={`registry-view-btn ${viewMode === "grid" ? "active" : ""}`}
                        onClick={() => setViewMode("grid")}
                        aria-label="Grid view"
                      >
                        <Grid className="h-4 w-4" />
                      </button>
                      <button
                        type="button"
                        className={`registry-view-btn ${viewMode === "list" ? "active" : ""}`}
                        onClick={() => setViewMode("list")}
                        aria-label="List view"
                      >
                        <List className="h-4 w-4" />
                      </button>
                    </div>
                  </div>

                  {/* Sort controls */}
                  <div className={`${showMobileFilters ? "block" : "hidden"} lg:block`}>
                    <div className="registry-controls">
                      <span className="text-xs font-semibold text-text-muted whitespace-nowrap">
                        Sort by
                      </span>
                      <Select value={sortBy} onValueChange={setSortBy}>
                        <SelectTrigger className="flex-1 h-8 text-xs bg-transparent border-white/[0.07] min-w-[120px] max-w-[160px]">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          {SORT_OPTIONS.map((o) => (
                            <SelectItem key={o.value} value={o.value} className="text-xs">
                              {o.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <button
                        type="button"
                        onClick={() =>
                          setSortOrder(sortOrder === "asc" ? "desc" : "asc")
                        }
                        className="h-8 w-8 flex items-center justify-center rounded-lg border border-white/[0.07]
                                   bg-white/[0.03] text-text-muted hover:text-text-primary hover:bg-white/[0.06]
                                   transition-all"
                        aria-label={sortOrder === "asc" ? "Ascending" : "Descending"}
                      >
                        {sortOrder === "asc" ? (
                          <SortAsc className="h-3.5 w-3.5" />
                        ) : (
                          <SortDesc className="h-3.5 w-3.5" />
                        )}
                      </button>
                      <div className="flex-1" />
                      {/* Results count */}
                      <div className="registry-results-count">
                        <strong>
                          {loading ? "…" : filteredFunctions.length}
                        </strong>
                        &nbsp;
                        {filteredFunctions.length === 1 ? "function" : "functions"}
                      </div>
                    </div>
                  </div>

                  {/* Active filter chips */}
                  {hasActiveFilters && (
                    <div className="flex flex-wrap gap-2 items-center">
                      <span className="text-xs text-text-muted font-medium">
                        Filters:
                      </span>
                      {searchQuery && (
                        <span className="registry-filter-chip">
                          <Search className="h-2.5 w-2.5" />
                          "
                          {searchQuery.length > 16
                            ? searchQuery.slice(0, 16) + "…"
                            : searchQuery}
                          "
                          <button
                            type="button"
                            className="registry-filter-chip-close"
                            onClick={() => setSearchQuery("")}
                            aria-label="Clear search"
                          >
                            <X className="h-2 w-2" />
                          </button>
                        </span>
                      )}
                      {selectedCategory !== "All Categories" && (
                        <span className="registry-filter-chip">
                          <Layers className="h-2.5 w-2.5" />
                          {selectedCategory}
                          <button
                            type="button"
                            className="registry-filter-chip-close"
                            onClick={() => setSelectedCategory("All Categories")}
                            aria-label="Clear category"
                          >
                            <X className="h-2 w-2" />
                          </button>
                        </span>
                      )}
                      <button
                        type="button"
                        onClick={() => {
                          setSearchQuery("");
                          setSelectedCategory("All Categories");
                        }}
                        className="text-xs text-text-muted hover:text-text-secondary transition-colors
                                   underline underline-offset-2"
                      >
                        Clear all
                      </button>
                    </div>
                  )}
                </div>

                {/* ─── Results ──────────────────────────────────── */}
                {loading ? (
                  <SkeletonGrid mode={viewMode} count={6} />
                ) : filteredFunctions.length === 0 ? (
                  <div className="registry-empty-state">
                    <div className="registry-empty-icon">
                      <Code className="h-8 w-8 text-brand-500" />
                    </div>
                    <h3 className="text-xl font-bold text-text-primary mb-2">
                      No functions found
                    </h3>
                    <p className="text-sm text-text-secondary max-w-sm mb-6">
                      Try adjusting your search query or category filter to find
                      what you're looking for.
                    </p>
                    {hasActiveFilters && (
                      <button
                        type="button"
                        onClick={() => {
                          setSearchQuery("");
                          setSelectedCategory("All Categories");
                        }}
                        className="h-9 px-5 rounded-xl text-sm font-semibold
                                   bg-gradient-to-r from-brand-500 to-purple-500 text-white
                                   hover:brightness-110 transition-all shadow-lg shadow-brand-500/20"
                      >
                        Reset filters
                      </button>
                    )}
                  </div>
                ) : viewMode === "grid" ? (
                  <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4 animate-stagger">
                    {filteredFunctions.map((fn) => (
                      <GridCard
                        key={fn.id}
                        fn={fn}
                        isAuthenticated={isAuthenticated}
                        onView={() => handleView(fn)}
                        onDeploy={() => handleDeploy(fn)}
                        onTry={() => handleTry(fn)}
                        flashId={flashId}
                      />
                    ))}
                  </div>
                ) : (
                  <div className="space-y-2 animate-stagger">
                    {filteredFunctions.map((fn) => (
                      <ListCard
                        key={fn.id}
                        fn={fn}
                        isAuthenticated={isAuthenticated}
                        onView={() => handleView(fn)}
                        onDeploy={() => handleDeploy(fn)}
                        onTry={() => handleTry(fn)}
                        flashId={flashId}
                      />
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        </section>
      </main>

      <Footer />
    </div>
  );
}
