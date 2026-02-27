import { motion } from "framer-motion";
import { Integration } from "../../data/integrations";
import IntegrationCard from "./IntegrationCard";

interface IntegrationGridProps {
  integrations: Integration[];
  expandedIntegrations: Set<string>;
  onToggleExpansion: (integrationId: string) => void;
}

const IntegrationGrid = ({
  integrations,
  expandedIntegrations,
  onToggleExpansion
}: IntegrationGridProps) => {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
      className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 relative"
    >
      {/* Grid background pattern */}
      <div className="absolute inset-0 bg-[linear-gradient(45deg,transparent_25%,rgba(255,255,255,0.01)_25%,rgba(255,255,255,0.01)_50%,transparent_50%,transparent_75%,rgba(255,255,255,0.01)_75%)] bg-[length:30px_30px] opacity-20 pointer-events-none rounded-lg" />

      {integrations.map((integration, index) => (
        <motion.div
          key={integration.id}
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{
            duration: 0.6,
            delay: index * 0.1,
            ease: [0.25, 0.46, 0.45, 0.94],
          }}
          className="relative"
        >
          <IntegrationCard
            integration={integration}
            index={index}
            isExpanded={expandedIntegrations.has(integration.id)}
            onToggleExpansion={onToggleExpansion}
          />
        </motion.div>
      ))}
    </motion.div>
  );
};

export default IntegrationGrid;