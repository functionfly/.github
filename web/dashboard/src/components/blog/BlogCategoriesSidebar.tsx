'use client';

import { motion } from 'framer-motion';
import {
  Layers,
  Shield,
  Code2,
  Megaphone,
  GraduationCap,
  TrendingUp,
  Gamepad2,
  Zap,
  FileText,
} from 'lucide-react';

export interface BlogCategory {
  id: string;
  title: string;
  slug: string;
  order?: number;
}

const CATEGORY_PALETTE = [
  'from-violet-500/20 to-fuchsia-500/20 border-violet-400/30 hover:shadow-violet-500/20',
  'from-emerald-500/20 to-teal-500/20 border-emerald-400/30 hover:shadow-emerald-500/20',
  'from-amber-500/20 to-orange-500/20 border-amber-400/30 hover:shadow-amber-500/20',
  'from-rose-500/20 to-pink-500/20 border-rose-400/30 hover:shadow-rose-500/20',
  'from-blue-500/20 to-cyan-500/20 border-blue-400/30 hover:shadow-blue-500/20',
  'from-indigo-500/20 to-violet-500/20 border-indigo-400/30 hover:shadow-indigo-500/20',
  'from-lime-500/20 to-green-500/20 border-lime-400/30 hover:shadow-lime-500/20',
  'from-sky-500/20 to-blue-500/20 border-sky-400/30 hover:shadow-sky-500/20',
  'from-orange-500/20 to-red-500/20 border-orange-400/30 hover:shadow-orange-500/20',
  'from-purple-500/20 to-indigo-500/20 border-purple-400/30 hover:shadow-purple-500/20',
];

const CATEGORY_ICON_MAP: Record<string, React.ReactNode> = {
  ai: <Zap className="h-4 w-4" />,
  automation: <Zap className="h-4 w-4" />,
  platform: <Layers className="h-4 w-4" />,
  architecture: <Layers className="h-4 w-4" />,
  infrastructure: <Layers className="h-4 w-4" />,
  security: <Shield className="h-4 w-4" />,
  trust: <Shield className="h-4 w-4" />,
  product: <Code2 className="h-4 w-4" />,
  tutorial: <GraduationCap className="h-4 w-4" />,
  'how-to': <GraduationCap className="h-4 w-4" />,
  builder: <TrendingUp className="h-4 w-4" />,
  case: <FileText className="h-4 w-4" />,
  gamification: <Gamepad2 className="h-4 w-4" />,
  reputation: <Gamepad2 className="h-4 w-4" />,
  growth: <TrendingUp className="h-4 w-4" />,
  monetization: <TrendingUp className="h-4 w-4" />,
  open: <Code2 className="h-4 w-4" />,
  protocol: <Code2 className="h-4 w-4" />,
  compute: <Layers className="h-4 w-4" />,
  announcement: <Megaphone className="h-4 w-4" />,
  roadmap: <Megaphone className="h-4 w-4" />,
};

function getCategoryIcon(title: string, slug: string): React.ReactNode {
  const keys = [
    ...slug.toLowerCase().split('-'),
    ...title.toLowerCase().split(/\s+/),
  ];
  const key = keys.find((k) => CATEGORY_ICON_MAP[k]);
  return key ? CATEGORY_ICON_MAP[key] : <Layers className="h-4 w-4" />;
}

export interface BlogCategoriesSidebarProps {
  categories: BlogCategory[];
}

export function BlogCategoriesSidebar({ categories }: BlogCategoriesSidebarProps) {
  const sorted = [...categories].sort((a, b) => (a.order ?? 0) - (b.order ?? 0));

  if (sorted.length === 0) return null;

  return (
    <aside
      className="w-full lg:w-72 shrink-0"
      aria-label="Blog categories"
    >
      <div className="sticky top-24 space-y-4">
        <h3 className="text-sm font-semibold text-foreground tracking-wide uppercase">
          Categories
        </h3>
        <nav className="flex flex-col gap-2">
          {sorted.map((cat, i) => {
            const palette = CATEGORY_PALETTE[i % CATEGORY_PALETTE.length];
            const icon = getCategoryIcon(cat.title, cat.slug);
            return (
              <motion.div
                key={cat.id}
                initial={{ opacity: 0, x: -8 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{
                  duration: 0.3,
                  delay: 0.03 * i,
                  ease: [0.25, 0.46, 0.45, 0.94],
                }}
                whileHover={{ x: 4 }}
                whileTap={{ scale: 0.98 }}
                className={`
                  group relative overflow-hidden rounded-xl border bg-linear-to-br ${palette}
                  shadow-sm hover:shadow-md transition-all duration-300
                  cursor-default
                `}
              >
                <div className="absolute inset-0 bg-background/40 dark:bg-background/20 group-hover:bg-background/20 transition-colors duration-300" />
                <div className="relative flex items-center gap-3 p-3">
                  <div className="shrink-0 w-9 h-9 rounded-lg bg-background/60 dark:bg-background/40 border border-border/50 flex items-center justify-center text-foreground/80 group-hover:text-foreground transition-colors">
                    {icon}
                  </div>
                  <span className="text-sm font-medium text-foreground/90 leading-tight line-clamp-2 min-w-0">
                    {cat.title}
                  </span>
                </div>
              </motion.div>
            );
          })}
        </nav>
      </div>
    </aside>
  );
}
