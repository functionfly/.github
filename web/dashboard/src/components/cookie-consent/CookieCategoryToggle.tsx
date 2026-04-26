'use client';

import { useState } from 'react';
import { motion } from 'framer-motion';
import { Checkbox } from '@/components/ui/checkbox';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Lock, CheckCircle, Circle } from 'lucide-react';

interface CookieCategoryToggleProps {
  category: string;
  title: string;
  description: string;
  enabled: boolean;
  readOnly?: boolean;
  onChange: (enabled: boolean) => void;
}

export function CookieCategoryToggle({
  category,
  title,
  description,
  enabled,
  readOnly = false,
  onChange,
}: CookieCategoryToggleProps) {
  const [isHovered, setIsHovered] = useState(false);

  const getCategoryColor = (cat: string) => {
    switch (cat) {
      case 'necessary': return '#10b981'; // green
      case 'analytics': return '#f59e0b'; // amber
      case 'marketing': return '#ef4444'; // red
      case 'functionality': return '#8b5cf6'; // violet
      default: return '#6366f1'; // indigo
    }
  };

  const categoryColor = getCategoryColor(category);

  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      whileHover={{ scale: 1.02 }}
      transition={{ duration: 0.2 }}
      onHoverStart={() => setIsHovered(true)}
      onHoverEnd={() => setIsHovered(false)}
    >
      <Card
        className={`relative overflow-hidden transition-all duration-300 border cursor-pointer ${
          enabled
            ? 'border-[var(--border-subtle)] bg-[var(--bg-tertiary)] shadow-lg'
            : 'border-[var(--border-subtle)] bg-[var(--bg-secondary)] hover:border-[var(--border-default)] hover:shadow-md'
        }`}
      >
        {/* Background gradient overlay */}
        <div
          className={`absolute inset-0 opacity-0 transition-opacity duration-300 ${
            isHovered ? 'opacity-10' : ''
          }`}
          style={{
            background: `linear-gradient(135deg, ${categoryColor}20, transparent)`,
          }}
        />

        <CardHeader className="pb-3 relative z-10">
          <div className="flex items-start justify-between">
            <div className="flex-1">
              <CardTitle className="text-base flex items-center gap-3 mb-1">
                <motion.div
                  className="w-8 h-8 rounded-lg flex items-center justify-center"
                  style={{
                    backgroundColor: enabled ? `${categoryColor}20` : 'rgba(255,255,255,0.1)',
                  }}
                  whileHover={{ rotate: enabled ? 0 : 360 }}
                  transition={{ duration: 0.3 }}
                >
                  {readOnly ? (
                    <Lock className="h-4 w-4 text-green-400 dark:text-green-400" />
                  ) : enabled ? (
                    <CheckCircle className="h-4 w-4" style={{ color: categoryColor }} />
                  ) : (
                    <Circle className="h-4 w-4 text-white/40 dark:text-white/40 text-slate-400" />
                  )}
                </motion.div>
                <span className="text-[var(--text-primary)] font-medium">{title}</span>
                {readOnly && (
                  <motion.div
                    initial={{ scale: 0 }}
                    animate={{ scale: 1 }}
                    className="px-2 py-0.5 rounded-full bg-green-500/20 dark:bg-green-500/20 border border-green-500/30 dark:border-green-500/30"
                  >
                    <span className="text-xs font-medium text-green-400 dark:text-green-400">Required</span>
                  </motion.div>
                )}
              </CardTitle>
            </div>

            {!readOnly && (
              <motion.div
                whileHover={{ scale: 1.05 }}
                whileTap={{ scale: 0.95 }}
                className="relative"
              >
                <Checkbox
                  id={`cookie-${category}`}
                  checked={enabled}
                  onCheckedChange={onChange}
                  disabled={readOnly}
                  className="data-[state=checked]:bg-current data-[state=checked]:border-current"
                  style={{
                    color: enabled ? categoryColor : undefined,
                  }}
                />
                {/* Custom glow effect for enabled state */}
                {enabled && (
                  <motion.div
                    initial={{ scale: 0, opacity: 0 }}
                    animate={{ scale: 1, opacity: 1 }}
                    className="absolute inset-0 rounded-full"
                    style={{
                      background: `radial-gradient(circle, ${categoryColor}40 0%, transparent 70%)`,
                      filter: 'blur(8px)',
                    }}
                  />
                )}
              </motion.div>
            )}

            {readOnly && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="flex items-center gap-2 text-sm text-green-400 dark:text-green-400 font-medium"
              >
                <CheckCircle className="h-4 w-4" />
                <span>Always Active</span>
              </motion.div>
            )}
          </div>
        </CardHeader>

        <CardContent className="pt-0 relative z-10">
          <p
            className={`text-sm leading-relaxed transition-colors duration-200 ${
              isHovered
                ? 'text-[var(--text-primary)]'
                : 'text-[var(--text-secondary)]'
            }`}
          >
            {description}
          </p>

          {/* Category badge */}
          <motion.div
            className="mt-3 flex items-center gap-2"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.2 }}
          >
            <div
              className="px-2 py-1 rounded-full text-xs font-medium"
              style={{
                backgroundColor: `${categoryColor}20`,
                color: categoryColor,
                border: `1px solid ${categoryColor}30`,
              }}
            >
              {category.charAt(0).toUpperCase() + category.slice(1)}
            </div>
            {enabled && !readOnly && (
              <motion.div
                initial={{ scale: 0 }}
                animate={{ scale: 1 }}
                className="flex items-center gap-1 text-xs text-green-400 dark:text-green-400"
              >
                <motion.div
                  animate={{ rotate: 360 }}
                  transition={{ duration: 2, repeat: Infinity, ease: "linear" }}
                >
                  ✓
                </motion.div>
                Enabled
              </motion.div>
            )}
          </motion.div>
        </CardContent>

        {/* Hover indicator line */}
        <motion.div
          className="absolute bottom-0 left-0 h-0.5 bg-gradient-to-r"
          style={{
            background: `linear-gradient(to right, ${categoryColor}, ${categoryColor}80)`,
          }}
          initial={{ width: 0 }}
          animate={{ width: isHovered ? '100%' : '0%' }}
          transition={{ duration: 0.3 }}
        />
      </Card>
    </motion.div>
  );
}