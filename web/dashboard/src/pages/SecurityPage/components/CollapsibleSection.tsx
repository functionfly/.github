import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { ChevronDown, ChevronUp } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

interface CollapsibleSectionProps {
  title: string;
  description?: string;
  icon?: React.ReactNode;
  children: React.ReactNode;
  defaultExpanded?: boolean;
  className?: string;
}

export function CollapsibleSection({
  title,
  description,
  icon,
  children,
  defaultExpanded = true,
  className = ''
}: CollapsibleSectionProps) {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);
  const [isHovered, setIsHovered] = useState(false);

  // Get color scheme based on title for consistent theming
  const getSectionColors = (title: string) => {
    const colors = {
      'Real-Time Security Status': { bg: 'from-blue-500/20 to-cyan-500/20', border: 'border-blue-500/30', icon: 'text-blue-400' },
      'Compliance Certifications': { bg: 'from-purple-500/20 to-violet-500/20', border: 'border-purple-500/30', icon: 'text-purple-400' },
      'Security Measures': { bg: 'from-green-500/20 to-emerald-500/20', border: 'border-green-500/30', icon: 'text-green-400' },
      'Incident Response': { bg: 'from-orange-500/20 to-red-500/20', border: 'border-orange-500/30', icon: 'text-orange-400' },
      'Security FAQ': { bg: 'from-indigo-500/20 to-blue-500/20', border: 'border-indigo-500/30', icon: 'text-indigo-400' },
      'Security Resources': { bg: 'from-teal-500/20 to-cyan-500/20', border: 'border-teal-500/30', icon: 'text-teal-400' },
      'Contact Information': { bg: 'from-cyan-500/20 to-blue-500/20', border: 'border-cyan-500/30', icon: 'text-cyan-400' },
    };
    return colors[title as keyof typeof colors] || { bg: 'from-white/10 to-white/5', border: 'border-white/20', icon: 'text-white' };
  };

  const colors = getSectionColors(title);

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6 }}
      className={className}
      onHoverStart={() => setIsHovered(true)}
      onHoverEnd={() => setIsHovered(false)}
    >
      <Card className="w-full overflow-hidden bg-gradient-to-br from-white/10 to-white/5 backdrop-blur-xl border border-white/20 hover:border-white/30 shadow-lg hover:shadow-xl transition-all duration-300 group">
        {/* Card background gradient overlay */}
        <div className={`absolute inset-0 bg-gradient-to-br ${colors.bg} opacity-0 group-hover:opacity-10 transition-opacity duration-300`} />

        <CardHeader className="pb-4 relative z-10">
          <CardTitle className="flex items-start justify-between">
            <div className="flex items-start gap-4 flex-1">
              {icon && (
                <div className={`relative p-3 rounded-xl bg-gradient-to-br ${colors.bg} border ${colors.border} backdrop-blur-sm transition-all duration-300 group-hover:scale-110`}>
                  <div className={colors.icon}>
                    {icon}
                  </div>
                  <div className={`absolute inset-0 bg-gradient-to-br ${colors.bg} rounded-xl blur-lg opacity-50 -z-10`} />
                </div>
              )}
              <div className="flex-1">
                <h3 className="text-xl md:text-2xl font-bold text-white mb-2 group-hover:text-white/90 transition-colors">
                  {title}
                </h3>
                {description && (
                  <p className="text-base text-text-secondary leading-relaxed group-hover:text-text-secondary/90 transition-colors">
                    {description}
                  </p>
                )}
              </div>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setIsExpanded(!isExpanded)}
              className="md:hidden p-3 hover:bg-white/10 transition-all duration-300"
              aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
            >
              <motion.div
                animate={{ rotate: isExpanded ? 180 : 0 }}
                transition={{ duration: 0.3 }}
              >
                <ChevronDown className="h-5 w-5" />
              </motion.div>
            </Button>
          </CardTitle>
        </CardHeader>

        <AnimatePresence>
          {isExpanded && (
            <motion.div
              initial={{ height: 0, opacity: 0 }}
              animate={{ height: 'auto', opacity: 1 }}
              exit={{ height: 0, opacity: 0 }}
              transition={{ duration: 0.3, ease: 'easeInOut' }}
            >
              <CardContent className="pt-0 relative z-10">
                {children}
              </CardContent>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Hover indicator line */}
        <motion.div
          className={`absolute bottom-0 left-0 h-0.5 bg-gradient-to-r ${colors.bg.replace('/20', '')}`}
          initial={{ width: 0 }}
          animate={{ width: isHovered ? '100%' : '0%' }}
          transition={{ duration: 0.3 }}
        />
      </Card>
    </motion.div>
  );
}