import { Feature } from "../../types";
import { FeatureCard } from "./FeatureCard";

interface FeatureGridProps {
  features: Feature[];
  expandedFeatures: string[];
  onToggleFeatureExpansion: (title: string) => void;
}

export const FeatureGrid = ({ features, expandedFeatures, onToggleFeatureExpansion }: FeatureGridProps) => {
  return (
    <div className="feature-grid">
      {features.map((feature, index) => (
        <FeatureCard
          key={feature.title}
          feature={feature}
          index={index}
          isExpanded={expandedFeatures.includes(feature.title)}
          onToggleExpansion={onToggleFeatureExpansion}
        />
      ))}
    </div>
  );
};