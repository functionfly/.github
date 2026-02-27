import { motion } from "framer-motion";
import { Card, CardContent } from "@/components/ui/card";
import { Feature } from "../../types";

interface CategorySectionProps {
  category: string;
  categoryIndex: number;
  features: Feature[];
}

export const CategorySection = ({ category, categoryIndex, features }: CategorySectionProps) => {
  const getCategoryDescription = (category: string) => {
    const descriptions = {
      "Reliability": "Ensure your applications stay online with our advanced reliability features.",
      "Deployment": "Deploy anywhere with flexible, multi-provider deployment options.",
      "Intelligence": "Smart routing and predictive analytics keep your apps performing optimally.",
      "Developer Tools": "Developer-first tools that make deployment and management effortless.",
      "Collaboration": "Work together seamlessly with powerful team collaboration features.",
      "Monitoring": "Comprehensive analytics and monitoring for complete visibility.",
      "Configuration": "Fine-tune your deployments with granular configuration options.",
      "Infrastructure": "Leverage our global edge network for optimal performance.",
    };
    return descriptions[category as keyof typeof descriptions] || "";
  };

  return (
    <motion.section
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.6, delay: categoryIndex * 0.2 }}
      className="category-section"
    >
      <div className="category-section-header">
        <h2>
          {category}
        </h2>
        <p>
          {getCategoryDescription(category)}
        </p>
      </div>

      <div className="category-grid">
        {features.map((feature, index) => (
          <motion.div
            key={feature.title}
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: index * 0.1 }}
          >
            <div className="category-card">
              <div className="category-card-icon">
                {feature.illustration ? (
                  <feature.illustration />
                ) : (
                      <span style={{ color: feature.color }}>
                        <feature.icon
                          className="w-5 h-5"
                          aria-hidden="true"
                        />
                      </span>
                )}
              </div>
              <h3>
                {feature.title}
              </h3>
              <p>
                {feature.description}
              </p>
            </div>
          </motion.div>
        ))}
      </div>
    </motion.section>
  );
};