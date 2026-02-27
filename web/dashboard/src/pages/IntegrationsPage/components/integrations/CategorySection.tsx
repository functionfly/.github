import { motion } from "framer-motion";
import { Integration } from "../../data/integrations";
import IntegrationGrid from "./IntegrationGrid";
import { categoryColors } from "../../types";

const getCategoryDescription = (category: string) => {
  const descriptions = {
    "Cloud Providers": "Deploy to multiple cloud platforms simultaneously with FunctionFly's cloud provider integrations.",
    "Frameworks": "Native support for popular web frameworks and runtimes to streamline your development workflow.",
    "Deployment Platforms": "One-click deployment to your preferred hosting platforms with seamless integration.",
    "Databases": "Connect to popular databases and data stores with built-in connection management.",
    "APIs & Services": "Integrate with third-party APIs and microservices to extend your functionality.",
    "Monitoring & Analytics": "Track performance, errors, and user behavior with comprehensive monitoring tools.",
  };
  return descriptions[category as keyof typeof descriptions] || "";
};

interface CategorySectionProps {
  category: string;
  categoryIndex: number;
  integrations: Integration[];
  expandedIntegrations: Set<string>;
  onToggleExpansion: (integrationId: string) => void;
}

const CategorySection = ({
  category,
  categoryIndex,
  integrations,
  expandedIntegrations,
  onToggleExpansion
}: CategorySectionProps) => {
  const categoryColor = categoryColors[category] || "#6366f1";

  return (
    <motion.section
      initial={{ opacity: 0, y: 50 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{
        duration: 0.6,
        delay: categoryIndex * 0.1
      }}
      className="py-20 relative"
    >
      {/* Section background with subtle gradient */}
      <div
        className="absolute inset-0 rounded-3xl opacity-30 blur-3xl -mx-8"
        style={{
          background: `linear-gradient(135deg, ${categoryColor}08, transparent, ${categoryColor}05)`,
        }}
      />

      <div className="relative">
        <div className="mb-12">
          <div className="flex items-center gap-4 mb-4">
            <div
              className="w-12 h-12 rounded-xl flex items-center justify-center shadow-lg backdrop-blur-sm"
              style={{
                background: `linear-gradient(135deg, ${categoryColor}20, ${categoryColor}10)`,
                boxShadow: `0 8px 32px ${categoryColor}15`
              }}
            >
              <div
                className="w-6 h-6 rounded-lg"
                style={{ backgroundColor: categoryColor }}
              />
            </div>
            <div>
              <h2
                className="text-3xl md:text-4xl font-bold mb-2"
                style={{ color: categoryColor }}
              >
                {category}
              </h2>
              <div
                className="w-16 h-1 rounded-full"
                style={{ backgroundColor: categoryColor }}
              />
            </div>
          </div>
          <p className="text-text-secondary text-lg max-w-2xl">
            {getCategoryDescription(category)}
          </p>
        </div>

        <IntegrationGrid
          integrations={integrations}
          expandedIntegrations={expandedIntegrations}
          onToggleExpansion={onToggleExpansion}
        />
      </div>
    </motion.section>
  );
};

export default CategorySection;