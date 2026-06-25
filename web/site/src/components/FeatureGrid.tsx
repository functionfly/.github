import React from "react";
import { Chamber, CornerBrace } from "./sc";

interface Feature {
  icon: string;
  title: string;
  desc: string;
  tag: string;
  color: "flame" | "cyan" | "strat" | "taxiway" | "beacon" | "afterburner";
}

interface FeatureGridProps {
  features: Feature[];
}

const COLOR_MAP: Record<Feature["color"], { bg: string; border: string; glow: string; text: string }> = {
  flame: { bg: "rgba(255, 107, 53, 0.1)", border: "rgba(255, 107, 53, 0.3)", glow: "rgba(255, 107, 53, 0.4)", text: "#FF6B35" },
  cyan: { bg: "rgba(0, 212, 255, 0.1)", border: "rgba(0, 212, 255, 0.3)", glow: "rgba(0, 212, 255, 0.4)", text: "#00D4FF" },
  strat: { bg: "rgba(91, 124, 245, 0.1)", border: "rgba(91, 124, 245, 0.3)", glow: "rgba(91, 124, 245, 0.4)", text: "#5B7CF5" },
  taxiway: { bg: "rgba(0, 255, 157, 0.1)", border: "rgba(0, 255, 157, 0.3)", glow: "rgba(0, 255, 157, 0.4)", text: "#00FF9D" },
  beacon: { bg: "rgba(255, 184, 0, 0.1)", border: "rgba(255, 184, 0, 0.3)", glow: "rgba(255, 184, 0, 0.4)", text: "#FFB800" },
  afterburner: { bg: "rgba(255, 79, 94, 0.1)", border: "rgba(255, 79, 94, 0.3)", glow: "rgba(255, 79, 94, 0.4)", text: "#FF4F5E" },
};

export const FeatureGrid: React.FC<FeatureGridProps> = ({ features }) => {
  return (
    <div className="ff-feature-grid">
      {features.map((feature, index) => {
        const colors = COLOR_MAP[feature.color];
        return (
          <Chamber key={feature.title} ribs className="ff-feature-card">
            <CornerBrace position="tl" />
            <CornerBrace position="br" />
            <div
              className="ff-feature-icon"
              style={{
                background: colors.bg,
                borderColor: colors.border,
                boxShadow: `0 0 20px ${colors.glow}`,
              }}
            >
              {feature.icon}
            </div>
            <h3 className="ff-feature-title">{feature.title}</h3>
            <p className="ff-feature-desc">{feature.desc}</p>
            <span
              className="ff-feature-tag"
              style={{
                background: colors.bg,
                borderColor: colors.border,
                color: colors.text,
              }}
            >
              {feature.tag}
            </span>
          </Chamber>
        );
      })}
    </div>
  );
};

export default FeatureGrid;
