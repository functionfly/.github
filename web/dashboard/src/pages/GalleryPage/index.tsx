/**
 * Gallery Page
 * A production-ready showcase page for FunctionFly visual assets, workflows, and media
 * Designed with an editorial/magazine aesthetic - refined, asymmetrical, and visually striking
 */

import { useState, useRef, useEffect, useCallback } from 'react';
import { motion, AnimatePresence, useScroll, useTransform, useSpring } from 'framer-motion';
import {
  Search,
  X,
  Grid3X3,
  LayoutGrid,
  Maximize2,
  Download,
  Share2,
  ChevronLeft,
  ChevronRight,
  Filter,
  ArrowUpRight,
  Image as ImageIcon,
  Layers,
  Play,
  FileCode,
  Cpu,
  Sparkles,
  Zap,
  Box,
  Menu,
  ExternalLink,
  Heart,
  Eye,
  Clock,
  MoreHorizontal,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import { GlassmorphismCard } from '@/components/ui/GlassmorphismCard';
import { SpotlightCard } from '@/components/ui/SpotlightCard';
import { TextGradient } from '@/components/ui/TextGradient';
import { cn } from '@/lib/utils';

// =============================================================================
// Types & Interfaces
// =============================================================================

type GalleryCategory = 'all' | 'workflows' | 'graphs' | 'demos' | 'ui' | 'infra';

interface GalleryItem {
  id: string;
  title: string;
  description: string;
  category: GalleryCategory;
  type: 'image' | 'video' | 'interactive' | 'code';
  thumbnail: string;
  fullSize: string;
  author: string;
  authorAvatar: string;
  date: string;
  likes: number;
  views: number;
  tags: string[];
  featured?: boolean;
  aspectRatio: 'landscape' | 'portrait' | 'square' | 'wide';
}

// =============================================================================
// Mock Data - Gallery Items
// =============================================================================

const GALLERY_ITEMS: GalleryItem[] = [
  {
    id: '1',
    title: 'E-Commerce Checkout Flow',
    description: 'Complete serverless workflow for handling checkout processes with inventory management and payment processing.',
    category: 'workflows',
    type: 'image',
    thumbnail: 'https://images.unsplash.com/photo-1558494949-ef010cbdcc31?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1558494949-ef010cbdcc31?w=1600&q=80',
    author: 'Sarah Chen',
    authorAvatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=100&q=80',
    date: '2024-12-15',
    likes: 234,
    views: 1205,
    tags: ['e-commerce', 'checkout', 'payments'],
    featured: true,
    aspectRatio: 'landscape',
  },
  {
    id: '2',
    title: 'AI Pipeline Architecture',
    description: 'Multi-stage AI processing graph with model inference, data transformation, and result caching.',
    category: 'graphs',
    type: 'image',
    thumbnail: 'https://images.unsplash.com/photo-1620712943543-bcc4688e7485?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1620712943543-bcc4688e7485?w=1600&q=80',
    author: 'Marcus Johnson',
    authorAvatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=100&q=80',
    date: '2024-12-10',
    likes: 189,
    views: 892,
    tags: ['ai', 'ml', 'pipeline'],
    featured: true,
    aspectRatio: 'square',
  },
  {
    id: '3',
    title: 'Real-time Data Processing Demo',
    description: 'Live demonstration of streaming data processing with sub-millisecond latency.',
    category: 'demos',
    type: 'video',
    thumbnail: 'https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=1600&q=80',
    author: 'FunctionFly Team',
    authorAvatar: 'https://images.unsplash.com/photo-1560250097-0b93528c311a?w=100&q=80',
    date: '2024-12-08',
    likes: 456,
    views: 2341,
    tags: ['streaming', 'real-time', 'demo'],
    featured: true,
    aspectRatio: 'wide',
  },
  {
    id: '4',
    title: '3D Graph Visualization',
    description: 'Interactive 3D visualization of function call graphs with real-time execution tracing.',
    category: 'graphs',
    type: 'interactive',
    thumbnail: 'https://images.unsplash.com/photo-1633356122544-f134324a6cee?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1633356122544-f134324a6cee?w=1600&q=80',
    author: 'Alex Rivera',
    authorAvatar: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=100&q=80',
    date: '2024-12-05',
    likes: 312,
    views: 1567,
    tags: ['3d', 'visualization', 'interactive'],
    aspectRatio: 'landscape',
  },
  {
    id: '5',
    title: 'Dashboard UI Components',
    description: 'Comprehensive UI kit for building FunctionFly dashboard interfaces.',
    category: 'ui',
    type: 'image',
    thumbnail: 'https://images.unsplash.com/photo-1581291518633-83b4ebd1d83e?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1581291518633-83b4ebd1d83e?w=1600&q=80',
    author: 'Design Team',
    authorAvatar: 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=100&q=80',
    date: '2024-12-01',
    likes: 178,
    views: 923,
    tags: ['ui', 'components', 'design'],
    aspectRatio: 'portrait',
  },
  {
    id: '6',
    title: 'Multi-Region Deployment Map',
    description: 'Global infrastructure visualization showing FunctionFly edge locations worldwide.',
    category: 'infra',
    type: 'image',
    thumbnail: 'https://images.unsplash.com/photo-1451187580459-43490279c0fa?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1451187580459-43490279c0fa?w=1600&q=80',
    author: 'DevOps Team',
    authorAvatar: 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&q=80',
    date: '2024-11-28',
    likes: 267,
    views: 1345,
    tags: ['infrastructure', 'global', 'edge'],
    featured: true,
    aspectRatio: 'wide',
  },
  {
    id: '7',
    title: 'Serverless Workflow Composer',
    description: 'Visual workflow builder interface for creating complex function compositions.',
    category: 'workflows',
    type: 'image',
    thumbnail: 'https://images.unsplash.com/photo-1504868584819-f8e8b4b6d7e3?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1504868584819-f8e8b4b6d7e3?w=1600&q=80',
    author: 'Product Team',
    authorAvatar: 'https://images.unsplash.com/photo-1519345182560-3f2917c472ef?w=100&q=80',
    date: '2024-11-25',
    likes: 198,
    views: 876,
    tags: ['composer', 'builder', 'visual'],
    aspectRatio: 'landscape',
  },
  {
    id: '8',
    title: 'GraphQL API Explorer',
    description: 'Interactive GraphQL playground for testing FunctionFly API endpoints.',
    category: 'demos',
    type: 'interactive',
    thumbnail: 'https://images.unsplash.com/photo-1555949963-ff9fe0c870eb?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1555949963-ff9fe0c870eb?w=1600&q=80',
    author: 'API Team',
    authorAvatar: 'https://images.unsplash.com/photo-1463453091185-61582044d556?w=100&q=80',
    date: '2024-11-20',
    likes: 145,
    views: 678,
    tags: ['api', 'graphql', 'explorer'],
    aspectRatio: 'landscape',
  },
  {
    id: '9',
    title: 'Function Runtime Metrics',
    description: 'Performance analytics dashboard showing execution times and resource utilization.',
    category: 'infra',
    type: 'image',
    thumbnail: 'https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1551288049-bebda4e38f71?w=1600&q=80',
    author: 'Analytics Team',
    authorAvatar: 'https://images.unsplash.com/photo-1489424734100-6f45997a3410?w=100&q=80',
    date: '2024-11-18',
    likes: 234,
    views: 1123,
    tags: ['metrics', 'analytics', 'performance'],
    aspectRatio: 'square',
  },
  {
    id: '10',
    title: 'Dark Mode UI System',
    description: 'Complete dark mode design system with accessibility-first color palette.',
    category: 'ui',
    type: 'image',
    thumbnail: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=1600&q=80',
    author: 'UI Team',
    authorAvatar: 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=100&q=80',
    date: '2024-11-15',
    likes: 389,
    views: 2134,
    tags: ['dark-mode', 'design-system', 'accessibility'],
    aspectRatio: 'portrait',
  },
  {
    id: '11',
    title: 'Edge Function Deployment Flow',
    description: 'Step-by-step visualization of deploying functions to edge locations.',
    category: 'workflows',
    type: 'code',
    thumbnail: 'https://images.unsplash.com/photo-1516110833967-0b5716ca1387?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1516110833967-0b5716ca1387?w=1600&q=80',
    author: 'Platform Team',
    authorAvatar: 'https://images.unsplash.com/photo-1507591064344-4c6ce005b128?w=100&q=80',
    date: '2024-11-12',
    likes: 156,
    views: 789,
    tags: ['deployment', 'edge', 'ci/cd'],
    aspectRatio: 'landscape',
  },
  {
    id: '12',
    title: 'State Fabric Visualizer',
    description: '3D state visualization showing data flow through distributed state fabric.',
    category: 'graphs',
    type: 'interactive',
    thumbnail: 'https://images.unsplash.com/photo-1550751827-4bd374c3f58b?w=800&q=80',
    fullSize: 'https://images.unsplash.com/photo-1550751827-4bd374c3f58b?w=1600&q=80',
    author: 'Research Team',
    authorAvatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=100&q=80',
    date: '2024-11-10',
    likes: 278,
    views: 1432,
    tags: ['state', 'fabric', 'distributed'],
    aspectRatio: 'wide',
  },
];

const CATEGORIES: { value: GalleryCategory; label: string; icon: React.ElementType; color: string }[] = [
  { value: 'all', label: 'All Items', icon: LayoutGrid, color: '#6366f1' },
  { value: 'workflows', label: 'Workflows', icon: Layers, color: '#3b82f6' },
  { value: 'graphs', label: 'Graphs', icon: Box, color: '#8b5cf6' },
  { value: 'demos', label: 'Demos', icon: Play, color: '#f59e0b' },
  { value: 'ui', label: 'UI/UX', icon: ImageIcon, color: '#ec4899' },
  { value: 'infra', label: 'Infrastructure', icon: Cpu, color: '#10b981' },
];

// =============================================================================
// Utility Components
// =============================================================================

function AnimatedCounter({ value, suffix = '' }: { value: number; suffix?: string }) {
  const [count, setCount] = useState(0);
  const ref = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          const duration = 1500;
          const steps = 30;
          const increment = value / steps;
          let current = 0;
          let step = 0;

          const timer = setInterval(() => {
            step++;
            current = Math.min(value, increment * step);
            setCount(Math.floor(current));
            if (step >= steps) {
              clearInterval(timer);
              setCount(value);
            }
          }, duration / steps);

          observer.disconnect();
        }
      },
      { threshold: 0.5 }
    );

    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, [value]);

  return (
    <span ref={ref}>
      {count.toLocaleString()}{suffix}
    </span>
  );
}

// =============================================================================
// Lightbox Component
// =============================================================================

function Lightbox({
  item,
  onClose,
  onNext,
  onPrev,
  hasNext,
  hasPrev,
}: {
  item: GalleryItem;
  onClose: () => void;
  onNext: () => void;
  onPrev: () => void;
  hasNext: boolean;
  hasPrev: boolean;
}) {
  const [isLoaded, setIsLoaded] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
      if (e.key === 'ArrowRight' && hasNext) onNext();
      if (e.key === 'ArrowLeft' && hasPrev) onPrev();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [onClose, onNext, onPrev, hasNext, hasPrev]);

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="fixed inset-0 z-[100] flex items-center justify-center"
      onClick={onClose}
    >
      {/* Backdrop */}
      <div className="absolute inset-0 bg-black/95 backdrop-blur-xl" />

      {/* Content */}
      <motion.div
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.9, opacity: 0 }}
        transition={{ type: 'spring', damping: 25, stiffness: 300 }}
        className="relative z-10 w-full h-full max-w-7xl mx-auto p-4 md:p-8 flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-4">
            <img
              src={item.authorAvatar}
              alt={item.author}
              className="w-10 h-10 rounded-full border-2 border-white/20"
            />
            <div>
              <h3 className="text-white font-semibold">{item.title}</h3>
              <p className="text-white/60 text-sm">by {item.author}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon" className="text-white/60 hover:text-white hover:bg-white/10">
              <Download className="w-5 h-5" />
            </Button>
            <Button variant="ghost" size="icon" className="text-white/60 hover:text-white hover:bg-white/10">
              <Share2 className="w-5 h-5" />
            </Button>
            <Button variant="ghost" size="icon" onClick={onClose} className="text-white/60 hover:text-white hover:bg-white/10">
              <X className="w-5 h-5" />
            </Button>
          </div>
        </div>

        {/* Main Image */}
        <div className="flex-1 relative flex items-center justify-center overflow-hidden rounded-2xl bg-black/50">
          {!isLoaded && (
            <div className="absolute inset-0 flex items-center justify-center">
              <div className="w-8 h-8 border-2 border-white/20 border-t-white rounded-full animate-spin" />
            </div>
          )}
          <motion.img
            src={item.fullSize}
            alt={item.title}
            className={cn(
              "max-w-full max-h-full object-contain rounded-2xl",
              !isLoaded && "opacity-0"
            )}
            initial={{ opacity: 0 }}
            animate={{ opacity: isLoaded ? 1 : 0 }}
            onLoad={() => setIsLoaded(true)}
          />

          {/* Navigation Arrows */}
          {hasPrev && (
            <button
              onClick={onPrev}
              className="absolute left-4 p-3 rounded-full bg-black/50 text-white/80 hover:bg-black/70 hover:text-white transition-all"
            >
              <ChevronLeft className="w-6 h-6" />
            </button>
          )}
          {hasNext && (
            <button
              onClick={onNext}
              className="absolute right-4 p-3 rounded-full bg-black/50 text-white/80 hover:bg-black/70 hover:text-white transition-all"
            >
              <ChevronRight className="w-6 h-6" />
            </button>
          )}
        </div>

        {/* Footer */}
        <div className="mt-4 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <span className="flex items-center gap-2 text-white/60">
              <Heart className="w-4 h-4" />
              {item.likes.toLocaleString()}
            </span>
            <span className="flex items-center gap-2 text-white/60">
              <Eye className="w-4 h-4" />
              {item.views.toLocaleString()}
            </span>
            <span className="flex items-center gap-2 text-white/60">
              <Clock className="w-4 h-4" />
              {new Date(item.date).toLocaleDateString()}
            </span>
          </div>
          <div className="flex gap-2">
            {item.tags.map((tag) => (
              <Badge key={tag} variant="outline" className="border-white/20 text-white/60">
                {tag}
              </Badge>
            ))}
          </div>
        </div>
      </motion.div>
    </motion.div>
  );
}

// =============================================================================
// Gallery Card Component
// =============================================================================

function GalleryCard({
  item,
  index,
  onClick,
  layoutMode,
}: {
  item: GalleryItem;
  index: number;
  onClick: () => void;
  layoutMode: 'masonry' | 'grid';
}) {
  const [isHovered, setIsHovered] = useState(false);
  const [imageLoaded, setImageLoaded] = useState(false);

  const aspectClasses = {
    landscape: 'aspect-[16/10]',
    portrait: 'aspect-[3/4]',
    square: 'aspect-square',
    wide: 'aspect-[21/9]',
  };

  const getTypeIcon = () => {
    switch (item.type) {
      case 'video': return <Play className="w-4 h-4" />;
      case 'interactive': return <Sparkles className="w-4 h-4" />;
      case 'code': return <FileCode className="w-4 h-4" />;
      default: return <ImageIcon className="w-4 h-4" />;
    }
  };

  const categoryColor = CATEGORIES.find(c => c.value === item.category)?.color || '#6366f1';

  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05, duration: 0.5, ease: [0.23, 1, 0.32, 1] }}
      layout
      className={cn(
        "group relative cursor-pointer",
        layoutMode === 'grid' && aspectClasses[item.aspectRatio]
      )}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
      onClick={onClick}
    >
      <SpotlightCard
        className={cn(
          "h-full overflow-hidden rounded-2xl border-0 bg-transparent",
          layoutMode === 'masonry' && aspectClasses[item.aspectRatio]
        )}
        spotlightColor={categoryColor}
        spotlightSize={300}
      >
        {/* Image Container */}
        <div className="relative h-full overflow-hidden rounded-2xl">
          {!imageLoaded && (
            <div className="absolute inset-0 bg-slate-800/50 animate-pulse" />
          )}
          <img
            src={item.thumbnail}
            alt={item.title}
            className={cn(
              "h-full w-full object-cover transition-all duration-700",
              isHovered ? "scale-110" : "scale-100",
              !imageLoaded && "opacity-0"
            )}
            onLoad={() => setImageLoaded(true)}
          />

          {/* Gradient Overlay */}
          <div className={cn(
            "absolute inset-0 bg-gradient-to-t transition-opacity duration-500",
            isHovered
              ? "from-black/90 via-black/40 to-transparent opacity-100"
              : "from-black/60 via-black/20 to-transparent opacity-60"
          )} />

          {/* Type Badge */}
          <div className="absolute top-4 left-4">
            <Badge
              className="bg-black/60 backdrop-blur-sm text-white border-0"
              style={{ backgroundColor: `${categoryColor}40` }}
            >
              {getTypeIcon()}
              <span className="ml-1 capitalize">{item.type}</span>
            </Badge>
          </div>

          {/* Featured Badge */}
          {item.featured && (
            <div className="absolute top-4 right-4">
              <Badge className="bg-brand-500/80 text-white border-0">
                <Zap className="w-3 h-3 mr-1" />
                Featured
              </Badge>
            </div>
          )}

          {/* Content Overlay */}
          <div className={cn(
            "absolute inset-x-0 bottom-0 p-4 transition-all duration-500",
            isHovered ? "translate-y-0 opacity-100" : "translate-y-4 opacity-0"
          )}>
            <h3 className="text-white font-semibold text-lg mb-1 line-clamp-1">{item.title}</h3>
            <p className="text-white/70 text-sm line-clamp-2 mb-3">{item.description}</p>

            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <img
                  src={item.authorAvatar}
                  alt={item.author}
                  className="w-6 h-6 rounded-full border border-white/20"
                />
                <span className="text-white/80 text-sm">{item.author}</span>
              </div>
              <div className="flex items-center gap-3 text-white/60 text-sm">
                <span className="flex items-center gap-1">
                  <Heart className="w-4 h-4" />
                  {item.likes}
                </span>
                <span className="flex items-center gap-1">
                  <Eye className="w-4 h-4" />
                  {item.views}
                </span>
              </div>
            </div>
          </div>

          {/* Hover Actions */}
          <div className={cn(
            "absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 transition-all duration-300",
            isHovered ? "opacity-100 scale-100" : "opacity-0 scale-90"
          )}>
            <div className="flex gap-2">
              <Button size="sm" className="bg-white text-black hover:bg-white/90">
                <Maximize2 className="w-4 h-4 mr-1" />
                View
              </Button>
            </div>
          </div>
        </div>
      </SpotlightCard>
    </motion.div>
  );
}

// =============================================================================
// Featured Hero Section
// =============================================================================

function FeaturedSection({ items, onItemClick }: { items: GalleryItem[]; onItemClick: (item: GalleryItem) => void }) {
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end start"]
  });
  const y = useTransform(scrollYProgress, [0, 1], ["0%", "30%"]);

  const featured = items.filter(i => i.featured).slice(0, 3);

  return (
    <section ref={containerRef} className="relative min-h-[70vh] overflow-hidden">
      {/* Parallax Background */}
      <motion.div style={{ y }} className="absolute inset-0">
        <div className="absolute inset-0 bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900" />
        <div className="absolute inset-0 opacity-30">
          <div className="absolute top-0 left-1/4 w-96 h-96 bg-brand-500/20 rounded-full blur-[128px]" />
          <div className="absolute bottom-0 right-1/4 w-96 h-96 bg-purple-500/20 rounded-full blur-[128px]" />
        </div>
        {/* Grid Pattern */}
        <div
          className="absolute inset-0 opacity-[0.03]"
          style={{
            backgroundImage: `linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px),
                             linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px)`,
            backgroundSize: '50px 50px',
          }}
        />
      </motion.div>

      <div className="relative z-10 max-w-7xl mx-auto px-4 lg:px-6 py-20">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8 }}
          className="text-center mb-16"
        >
          <Badge variant="outline" className="mb-4 border-brand-500/50 text-brand-400">
            <Sparkles className="w-3 h-3 mr-1" />
            Curated Collection
          </Badge>
          <h1 className="text-4xl md:text-6xl lg:text-7xl font-bold mb-4">
            <TextGradient
              text="Visual Gallery"
              colors={['#6366f1', '#8b5cf6', '#ec4899']}
              className="font-bold"
            />
          </h1>
          <p className="text-lg md:text-xl text-slate-400 max-w-2xl mx-auto">
            Explore workflows, visualizations, and design systems from the FunctionFly ecosystem.
          </p>
        </motion.div>

        {/* Featured Cards - Asymmetrical Layout */}
        <div className="grid md:grid-cols-12 gap-4 md:gap-6">
          {featured.map((item, index) => {
            const spanClass = index === 0
              ? 'md:col-span-7 md:row-span-2'
              : index === 1
                ? 'md:col-span-5'
                : 'md:col-span-5';

            return (
              <motion.div
                key={item.id}
                className={spanClass}
                initial={{ opacity: 0, y: 40 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.2 + index * 0.15, duration: 0.6 }}
              >
                <GlassmorphismCard
                  className="h-full cursor-pointer group overflow-hidden rounded-2xl"
                  onClick={() => onItemClick(item)}
                >
                  <div className="relative h-full min-h-[250px] md:min-h-0">
                    <img
                      src={item.thumbnail}
                      alt={item.title}
                      className="absolute inset-0 w-full h-full object-cover transition-transform duration-700 group-hover:scale-105"
                    />
                    <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/20 to-transparent" />
                    <div className="absolute inset-0 p-6 flex flex-col justify-end">
                      <Badge className="w-fit mb-2 bg-white/20 text-white border-0">
                        {CATEGORIES.find(c => c.value === item.category)?.label}
                      </Badge>
                      <h3 className="text-xl md:text-2xl font-bold text-white mb-2">{item.title}</h3>
                      <p className="text-white/70 text-sm line-clamp-2">{item.description}</p>
                    </div>
                    <div className="absolute top-4 right-4 opacity-0 group-hover:opacity-100 transition-opacity">
                      <div className="w-10 h-10 rounded-full bg-white/20 backdrop-blur-sm flex items-center justify-center">
                        <ArrowUpRight className="w-5 h-5 text-white" />
                      </div>
                    </div>
                  </div>
                </GlassmorphismCard>
              </motion.div>
            );
          })}
        </div>

        {/* Stats Row */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.8 }}
          className="mt-12 grid grid-cols-2 md:grid-cols-4 gap-8"
        >
          {[
            { label: 'Total Items', value: GALLERY_ITEMS.length },
            { label: 'Contributors', value: 24 },
            { label: 'Categories', value: CATEGORIES.length - 1 },
            { label: 'Views', value: 12500, suffix: '+' },
          ].map((stat, i) => (
            <div key={i} className="text-center">
              <div className="text-3xl md:text-4xl font-bold text-white mb-1">
                <AnimatedCounter value={stat.value} suffix={stat.suffix} />
              </div>
              <div className="text-slate-400 text-sm">{stat.label}</div>
            </div>
          ))}
        </motion.div>
      </div>
    </section>
  );
}

// =============================================================================
// Main Gallery Page
// =============================================================================

export default function GalleryPage() {
  const [activeCategory, setActiveCategory] = useState<GalleryCategory>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [layoutMode, setLayoutMode] = useState<'masonry' | 'grid'>('masonry');
  const [lightboxItem, setLightboxItem] = useState<GalleryItem | null>(null);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  // Filter items
  const filteredItems = GALLERY_ITEMS.filter(item => {
    const matchesCategory = activeCategory === 'all' || item.category === activeCategory;
    const matchesSearch =
      searchQuery === '' ||
      item.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
      item.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase()));
    return matchesCategory && matchesSearch;
  });

  // Lightbox navigation
  const currentIndex = lightboxItem ? filteredItems.findIndex(i => i.id === lightboxItem.id) : -1;
  const hasNext = currentIndex < filteredItems.length - 1;
  const hasPrev = currentIndex > 0;

  const handleNext = useCallback(() => {
    if (hasNext) setLightboxItem(filteredItems[currentIndex + 1]);
  }, [currentIndex, filteredItems, hasNext]);

  const handlePrev = useCallback(() => {
    if (hasPrev) setLightboxItem(filteredItems[currentIndex - 1]);
  }, [currentIndex, filteredItems, hasPrev]);

  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      {/* Navigation */}
      <nav className="fixed top-0 left-0 right-0 z-50 border-b border-white/5 bg-slate-950/80 backdrop-blur-xl">
        <div className="max-w-7xl mx-auto px-4 lg:px-6 h-16 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-xl bg-gradient-to-br from-brand-500 to-purple-600 flex items-center justify-center">
              <LayoutGrid className="w-5 h-5 text-white" />
            </div>
            <span className="font-bold text-lg text-white">Gallery</span>
            <Badge variant="secondary" className="hidden sm:inline text-xs bg-slate-800 text-slate-400">
              FunctionFly
            </Badge>
          </div>

          <div className="hidden md:flex items-center gap-6">
            {CATEGORIES.map(cat => (
              <button
                key={cat.value}
                onClick={() => setActiveCategory(cat.value)}
                className={cn(
                  "text-sm transition-colors",
                  activeCategory === cat.value
                    ? "text-white font-medium"
                    : "text-slate-400 hover:text-slate-200"
                )}
              >
                {cat.label}
              </button>
            ))}
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="ghost"
              size="sm"
              className="hidden sm:flex text-slate-400 hover:text-white"
              onClick={() => window.open('/docs', '_blank')}
            >
              Documentation
            </Button>
            <Button
              size="sm"
              className="bg-brand-500 hover:bg-brand-600"
              onClick={() => window.location.href = '/overview'}
            >
              Dashboard
              <ExternalLink className="w-4 h-4 ml-1" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="md:hidden text-slate-400"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            >
              {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </Button>
          </div>
        </div>

        {/* Mobile Menu */}
        <AnimatePresence>
          {mobileMenuOpen && (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              exit={{ opacity: 0, height: 0 }}
              className="md:hidden border-t border-white/5 bg-slate-950"
            >
              <div className="px-4 py-4 space-y-2">
                {CATEGORIES.map(cat => (
                  <button
                    key={cat.value}
                    onClick={() => {
                      setActiveCategory(cat.value);
                      setMobileMenuOpen(false);
                    }}
                    className={cn(
                      "w-full text-left py-2 px-3 rounded-lg transition-colors",
                      activeCategory === cat.value
                        ? "bg-slate-800 text-white"
                        : "text-slate-400 hover:text-white hover:bg-slate-800/50"
                    )}
                  >
                    <cat.icon className="w-4 h-4 inline mr-2" style={{ color: cat.color }} />
                    {cat.label}
                  </button>
                ))}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </nav>

      {/* Hero Section */}
      <div className="pt-16">
        <FeaturedSection items={GALLERY_ITEMS} onItemClick={setLightboxItem} />
      </div>

      {/* Filter & Search Bar */}
      <div className="sticky top-16 z-40 bg-slate-950/95 backdrop-blur-xl border-b border-white/5">
        <div className="max-w-7xl mx-auto px-4 lg:px-6 py-4">
          <div className="flex flex-col md:flex-row gap-4 items-center justify-between">
            {/* Category Pills - Desktop */}
            <div className="hidden md:flex items-center gap-2 overflow-x-auto pb-2 md:pb-0">
              {CATEGORIES.map(cat => {
                const Icon = cat.icon;
                const isActive = activeCategory === cat.value;
                return (
                  <button
                    key={cat.value}
                    onClick={() => setActiveCategory(cat.value)}
                    className={cn(
                      "flex items-center gap-2 px-4 py-2 rounded-full text-sm font-medium transition-all",
                      isActive
                        ? "bg-white text-slate-950"
                        : "bg-slate-900 text-slate-400 hover:bg-slate-800 hover:text-slate-200"
                    )}
                  >
                    <Icon className="w-4 h-4" style={{ color: isActive ? cat.color : undefined }} />
                    {cat.label}
                  </button>
                );
              })}
            </div>

            {/* Search & Layout Controls */}
            <div className="flex items-center gap-3 w-full md:w-auto">
              <div className="relative flex-1 md:w-64">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
                <Input
                  placeholder="Search gallery..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 bg-slate-900 border-slate-800 text-slate-200 placeholder:text-slate-500 focus-visible:ring-brand-500"
                />
                {searchQuery && (
                  <button
                    onClick={() => setSearchQuery('')}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
                  >
                    <X className="w-4 h-4" />
                  </button>
                )}
              </div>

              <div className="flex items-center bg-slate-900 rounded-lg p-1">
                <button
                  onClick={() => setLayoutMode('masonry')}
                  className={cn(
                    "p-2 rounded-md transition-colors",
                    layoutMode === 'masonry' ? "bg-slate-800 text-white" : "text-slate-500 hover:text-slate-300"
                  )}
                  title="Masonry Layout"
                >
                  <LayoutGrid className="w-4 h-4" />
                </button>
                <button
                  onClick={() => setLayoutMode('grid')}
                  className={cn(
                    "p-2 rounded-md transition-colors",
                    layoutMode === 'grid' ? "bg-slate-800 text-white" : "text-slate-500 hover:text-slate-300"
                  )}
                  title="Grid Layout"
                >
                  <Grid3X3 className="w-4 h-4" />
                </button>
              </div>

              <Button variant="outline" size="icon" className="border-slate-800 text-slate-400 hover:text-white">
                <Filter className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* Gallery Grid */}
      <section className="max-w-7xl mx-auto px-4 lg:px-6 py-8">
        <AnimatePresence mode="wait">
          {filteredItems.length > 0 ? (
            <motion.div
              key={`${activeCategory}-${layoutMode}`}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.3 }}
              className={cn(
                "grid gap-4 md:gap-6",
                layoutMode === 'masonry'
                  ? "grid-cols-2 md:grid-cols-3 lg:grid-cols-4 auto-rows-[200px]"
                  : "grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4"
              )}
            >
              {filteredItems.map((item, index) => (
                <GalleryCard
                  key={item.id}
                  item={item}
                  index={index}
                  onClick={() => setLightboxItem(item)}
                  layoutMode={layoutMode}
                />
              ))}
            </motion.div>
          ) : (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-center py-20"
            >
              <div className="w-16 h-16 mx-auto mb-4 rounded-2xl bg-slate-900 flex items-center justify-center">
                <Search className="w-8 h-8 text-slate-600" />
              </div>
              <h3 className="text-xl font-semibold text-slate-300 mb-2">No items found</h3>
              <p className="text-slate-500">Try adjusting your search or filters</p>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Load More */}
        {filteredItems.length > 0 && (
          <div className="mt-12 text-center">
            <Button variant="outline" size="lg" className="border-slate-800 text-slate-400 hover:text-white hover:bg-slate-900">
              <MoreHorizontal className="w-4 h-4 mr-2" />
              Load More
            </Button>
          </div>
        )}
      </section>

      {/* Footer */}
      <footer className="border-t border-white/5 py-12 px-4 lg:px-6">
        <div className="max-w-7xl mx-auto">
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-brand-500 to-purple-600 flex items-center justify-center">
                <LayoutGrid className="w-4 h-4 text-white" />
              </div>
              <span className="font-semibold text-white">FunctionFly Gallery</span>
            </div>
            <p className="text-slate-500 text-sm">
              © 2024 FunctionFly. All rights reserved.
            </p>
            <div className="flex gap-4">
              <Button variant="ghost" size="sm" className="text-slate-500 hover:text-white">
                Submit
              </Button>
              <Button variant="ghost" size="sm" className="text-slate-500 hover:text-white">
                Guidelines
              </Button>
              <Button variant="ghost" size="sm" className="text-slate-500 hover:text-white">
                License
              </Button>
            </div>
          </div>
        </div>
      </footer>

      {/* Lightbox */}
      <AnimatePresence>
        {lightboxItem && (
          <Lightbox
            item={lightboxItem}
            onClose={() => setLightboxItem(null)}
            onNext={handleNext}
            onPrev={handlePrev}
            hasNext={hasNext}
            hasPrev={hasPrev}
          />
        )}
      </AnimatePresence>
    </div>
  );
}
