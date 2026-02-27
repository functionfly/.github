import { motion } from "framer-motion";
import {
  ChevronDown,
  ChevronUp,
  ExternalLink,
  Clock,
  Star,
  BookOpen,
  Code,
  PlayCircle
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Integration } from "../../data/integrations";
import { categoryColors } from "../../types";

interface IntegrationCardProps {
  integration: Integration;
  index: number;
  isExpanded: boolean;
  onToggleExpansion: (integrationId: string) => void;
}

const IntegrationCard = ({ integration, index, isExpanded, onToggleExpansion }: IntegrationCardProps) => {
  const categoryColor = categoryColors[integration.category] || "#6366f1";

  const getStatusBadge = (status: string) => {
    switch (status) {
      case 'available':
        return <Badge variant="secondary" className="text-xs bg-green-500/20 text-green-400 border-green-500/30">Available</Badge>;
      case 'beta':
        return <Badge variant="secondary" className="text-xs bg-yellow-500/20 text-yellow-400 border-yellow-500/30">Beta</Badge>;
      case 'coming-soon':
        return <Badge variant="secondary" className="text-xs bg-gray-500/20 text-gray-400 border-gray-500/30">Coming Soon</Badge>;
      default:
        return null;
    }
  };

  const getPopularityIcon = (popularity: string) => {
    switch (popularity) {
      case 'high':
        return <Star className="w-3 h-3 fill-yellow-400 text-yellow-400" />;
      case 'medium':
        return <Star className="w-3 h-3 text-yellow-400" />;
      case 'low':
        return <Star className="w-3 h-3 text-gray-400" />;
      default:
        return null;
    }
  };

  const IconComponent = integration.icon;

  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{
        duration: 0.6,
        delay: index * 0.1,
        ease: [0.25, 0.46, 0.45, 0.94],
      }}
    >
      <Card
        className="h-full transition-colors group hover:border-[#6366f1]/30"
        style={{
          borderColor: `${categoryColor}1a`,
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.borderColor = `${categoryColor}4d`;
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.borderColor = `${categoryColor}1a`;
        }}
      >
        <CardContent className="p-6">
          {/* Header with icon and badges */}
          <div className="flex items-start justify-between mb-4">
            <div
              className="w-12 h-12 rounded-xl bg-linear-to-br border flex items-center justify-center group-hover:scale-110 transition-transform will-change-transform"
              style={{
                background: `linear-gradient(135deg, ${categoryColor}33, ${categoryColor}1a)`,
                borderColor: `${categoryColor}33`,
              }}
            >
              <IconComponent className="w-6 h-6" />
            </div>
            <div className="flex items-center gap-2">
              {getPopularityIcon(integration.popularity)}
              {getStatusBadge(integration.status)}
            </div>
          </div>

          {/* Category and title */}
          <div className="mb-3">
            <span
              className="text-xs font-medium px-2 py-1 rounded-full"
              style={{
                color: categoryColor,
                backgroundColor: `${categoryColor}1a`,
              }}
            >
              {integration.category}
            </span>
          </div>

          <h3 className="text-lg font-semibold text-white mb-2">
            {integration.name}
          </h3>

          <p className="text-text-secondary text-sm mb-4">
            {integration.description}
          </p>

          {/* Setup time */}
          <div className="flex items-center gap-2 mb-4">
            <Clock className="w-3 h-3 text-text-secondary" />
            <span className="text-xs text-text-secondary">
              Setup: {integration.setupTime}
            </span>
          </div>

          {/* Features list */}
          <ul className="space-y-2 mb-4">
            {integration.features.slice(0, 3).map((feature, idx) => (
              <li
                key={idx}
                className="flex items-center gap-2 text-xs text-text-secondary"
              >
                <div
                  className="w-1.5 h-1.5 rounded-full"
                  style={{ backgroundColor: categoryColor }}
                />
                {feature}
              </li>
            ))}
            {integration.features.length > 3 && (
              <li className="text-xs text-text-secondary pl-3.5">
                +{integration.features.length - 3} more features
              </li>
            )}
          </ul>

          {/* Documentation links */}
          <div className="flex flex-wrap gap-2 mb-4">
            <a
              href={integration.documentation.guide}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-xs px-2 py-1 rounded-md transition-colors hover:bg-white/5"
              style={{ color: categoryColor }}
            >
              <BookOpen className="w-3 h-3" />
              Guide
            </a>
            <a
              href={integration.documentation.api}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-xs px-2 py-1 rounded-md transition-colors hover:bg-white/5"
              style={{ color: categoryColor }}
            >
              <Code className="w-3 h-3" />
              API
            </a>
            <a
              href={integration.documentation.examples}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1 text-xs px-2 py-1 rounded-md transition-colors hover:bg-white/5"
              style={{ color: categoryColor }}
            >
              <PlayCircle className="w-3 h-3" />
              Examples
            </a>
          </div>

          {/* Expand button */}
          <div className="pt-4 border-t border-white/10">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onToggleExpansion(integration.id)}
              className="w-full justify-between text-xs hover:bg-white/5"
              style={{ color: categoryColor }}
            >
              {isExpanded ? "Show Less" : "Learn More"}
              {isExpanded ? (
                <ChevronUp className="w-3 h-3" />
              ) : (
                <ChevronDown className="w-3 h-3" />
              )}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Expandable detailed content */}
      {isExpanded && (
        <motion.div
          initial={{ opacity: 0, height: 0 }}
          animate={{ opacity: 1, height: "auto" }}
          exit={{ opacity: 0, height: 0 }}
          transition={{ duration: 0.3 }}
          className="mt-4 overflow-hidden"
        >
          <Card className="border-white/10">
            <CardContent className="p-6">
              <div className="space-y-6">
                {/* All Features */}
                <div>
                  <h4 className="text-sm font-semibold text-white mb-3 flex items-center gap-2">
                    <Star className="w-4 h-4" style={{ color: categoryColor }} />
                    All Features
                  </h4>
                  <ul className="space-y-2">
                    {integration.features.map((feature, idx) => (
                      <li
                        key={idx}
                        className="flex items-start gap-2 text-sm text-text-secondary"
                      >
                        <div
                          className="w-1.5 h-1.5 rounded-full mt-2 shrink-0"
                          style={{ backgroundColor: categoryColor }}
                        />
                        {feature}
                      </li>
                    ))}
                  </ul>
                </div>

                {/* Setup Instructions */}
                <div className="bg-white/5 rounded-lg p-4 border border-white/10">
                  <h4 className="text-sm font-semibold text-white mb-2 flex items-center gap-2">
                    <Clock className="w-4 h-4" style={{ color: categoryColor }} />
                    Quick Setup
                  </h4>
                  <p className="text-sm text-text-secondary mb-3">
                    Get started with {integration.name} in just {integration.setupTime}.
                  </p>
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      className="text-xs h-7 px-3"
                      style={{
                        color: categoryColor,
                        borderColor: `${categoryColor}33`,
                      }}
                    >
                      <BookOpen className="w-3 h-3 mr-1" />
                      Setup Guide
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      className="text-xs h-7 px-3"
                      style={{
                        color: categoryColor,
                        borderColor: `${categoryColor}33`,
                      }}
                    >
                      <PlayCircle className="w-3 h-3 mr-1" />
                      Try Demo
                    </Button>
                  </div>
                </div>

                {/* Documentation Links */}
                <div>
                  <h4 className="text-sm font-semibold text-white mb-3 flex items-center gap-2">
                    <ExternalLink className="w-4 h-4" style={{ color: categoryColor }} />
                    Documentation
                  </h4>
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                    <a
                      href={integration.documentation.guide}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-2 p-3 rounded-lg border border-white/10 hover:border-white/20 transition-colors group"
                    >
                      <BookOpen className="w-4 h-4 text-text-secondary group-hover:text-white" />
                      <div>
                        <div className="text-sm font-medium text-white">Setup Guide</div>
                        <div className="text-xs text-text-secondary">Step-by-step instructions</div>
                      </div>
                    </a>
                    <a
                      href={integration.documentation.api}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-2 p-3 rounded-lg border border-white/10 hover:border-white/20 transition-colors group"
                    >
                      <Code className="w-4 h-4 text-text-secondary group-hover:text-white" />
                      <div>
                        <div className="text-sm font-medium text-white">API Reference</div>
                        <div className="text-xs text-text-secondary">Complete API docs</div>
                      </div>
                    </a>
                    <a
                      href={integration.documentation.examples}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="flex items-center gap-2 p-3 rounded-lg border border-white/10 hover:border-white/20 transition-colors group"
                    >
                      <PlayCircle className="w-4 h-4 text-text-secondary group-hover:text-white" />
                      <div>
                        <div className="text-sm font-medium text-white">Examples</div>
                        <div className="text-xs text-text-secondary">Code samples & demos</div>
                      </div>
                    </a>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        </motion.div>
      )}
    </motion.div>
  );
};

export default IntegrationCard;