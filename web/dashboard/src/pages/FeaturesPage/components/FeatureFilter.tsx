import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

interface FeatureFilterProps {
  categories: string[];
  features: Array<{ category: string }>;
  activeFilter: string | null;
  onFilterChange: (category: string | null) => void;
  categoryColors?: Record<string, { color: string; bgGradient: string; borderColor: string }>;
}

export function FeatureFilter({
  categories,
  features,
  activeFilter,
  onFilterChange,
  categoryColors = {}
}: FeatureFilterProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6, delay: 0.2 }}
      className="feature-filters"
    >
      <button
        onClick={() => onFilterChange(null)}
        className={`feature-filter-button ${activeFilter === null ? 'active' : ''}`}
      >
        All Features
        <span className="feature-filter-badge">
          {features.length}
        </span>
      </button>
      {categories.map((category) => {
        const categoryFeatures = features.filter(f => f.category === category);
        const isActive = activeFilter === category;
        const categoryColor = categoryColors[category];

        return (
          <button
            key={category}
            onClick={() => onFilterChange(isActive ? null : category)}
            className={`feature-filter-button ${isActive ? 'active' : ''}`}
            style={isActive && categoryColor ? {
              background: `linear-gradient(135deg, ${categoryColor.color} 0%, ${categoryColor.color}dd 100%)`,
              borderColor: categoryColor.color,
            } : undefined}
          >
            {category}
            <span className="feature-filter-badge">
              {categoryFeatures.length}
            </span>
          </button>
        );
      })}
    </motion.div>
  );
}