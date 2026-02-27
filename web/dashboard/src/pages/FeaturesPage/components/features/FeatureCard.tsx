import { motion } from "framer-motion";
import {
  ArrowRightLeft,
  ChevronDown,
  ChevronUp,
  Settings,
  Zap as ZapIcon,
  Shield,
  TrendingUp,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Link } from "react-router-dom";
import { Feature } from "../../types";

interface FeatureCardProps {
  feature: Feature;
  index: number;
  isExpanded: boolean;
  onToggleExpansion: (title: string) => void;
}

export const FeatureCard = ({ feature, index, isExpanded, onToggleExpansion }: FeatureCardProps) => {
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
      <div className="feature-card">
        <div className="feature-card-icon">
          {feature.illustration ? (
            <feature.illustration />
          ) : (
            <span style={{ color: feature.color }}>
              <feature.icon
                className="w-6 h-6"
                aria-hidden="true"
              />
            </span>
          )}
        </div>
        <div className="feature-card-category">
          {feature.category}
        </div>
        <h3 className="text-lg font-semibold text-white mb-2">
          {feature.title}
        </h3>
        <p className="text-text-secondary text-sm mb-4">
          {feature.description}
        </p>
        <ul className="feature-card-highlights">
          {feature.highlights.map((highlight, idx) => (
            <li
              key={idx}
              className="feature-card-highlight"
              style={{ color: feature.color }}
            >
              {highlight}
            </li>
          ))}
        </ul>

          {/* Feature-specific CTA */}
          {feature.cta && (
            <div className="mt-4 pt-3 border-t border-white/10">
              {feature.cta.action === 'demo' ? (
                <a
                  href={feature.cta.link}
                  className="feature-card-cta"
                  style={{
                    color: feature.color,
                    backgroundColor: `${feature.color}1a`,
                    borderColor: `${feature.color}33`,
                  }}
                >
                  {feature.cta.text}
                  <ArrowRightLeft className="w-3 h-3" />
                </a>
              ) : (
                <Link to={feature.cta.link!}>
                  <button
                    className="feature-card-cta"
                    style={{
                      color: feature.color,
                      backgroundColor: `${feature.color}1a`,
                      borderColor: `${feature.color}33`,
                    }}
                  >
                    {feature.cta.text}
                  </button>
                </Link>
              )}
            </div>
          )}

          {feature.detailedContent && (
            <button
              type="button"
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onToggleExpansion(feature.title);
              }}
              className="feature-card-expand"
            >
              {isExpanded ? "Show Less" : "Learn More"}
              {isExpanded ? (
                <ChevronUp className="w-3 h-3" />
              ) : (
                <ChevronDown className="w-3 h-3" />
              )}
            </button>
          )}
        </div>

      {/* Expandable detailed content */}
      {isExpanded && feature.detailedContent && (
        <div className="feature-details feature-details-expanded">
          <div className="feature-details-section">
            <div className="feature-details-header">
              <ZapIcon
                className="w-4 h-4"
                style={{ color: feature.color }}
              />
              <h4>Overview</h4>
            </div>
            <p className="text-sm text-text-secondary">
              {feature.detailedContent!.overview}
            </p>
          </div>

          <div className="feature-details-section">
            <div className="feature-details-header">
              <Settings
                className="w-4 h-4"
                style={{ color: feature.color }}
              />
              <h4>Technical Details</h4>
            </div>
            <ul className="feature-details-list">
              {feature.detailedContent!.technicalDetails.map(
                (detail, idx) => (
                  <li
                    key={idx}
                    className="feature-details-item"
                  >
                    {detail}
                  </li>
                ),
              )}
            </ul>
          </div>

          <div className="feature-details-section">
            <div className="feature-details-header">
              <TrendingUp
                className="w-4 h-4"
                style={{ color: feature.color }}
              />
              <h4>Key Benefits</h4>
            </div>
            <ul className="feature-details-list">
              {feature.detailedContent!.benefits.map(
                (benefit, idx) => (
                  <li
                    key={idx}
                    className="feature-details-item"
                  >
                    {benefit}
                  </li>
                ),
              )}
            </ul>
          </div>

          <div className="feature-details-implementation">
            <div className="feature-details-header">
              <Shield
                className="w-4 h-4"
                style={{ color: feature.color }}
              />
              <h4>Implementation</h4>
            </div>
            <p className="text-sm text-text-secondary">
              {feature.detailedContent!.implementation}
            </p>
          </div>
        </div>
      )}
    </motion.div>
  );
};