import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { IntegrationCategory } from "../../data/integrations";
import { Integration } from "../../data/integrations";
import { CategoryColors } from "../../types";

interface IntegrationFilterProps {
  categories: IntegrationCategory[];
  integrations: Integration[];
  activeFilter: string | null;
  onFilterChange: (category: string | null) => void;
  categoryColors: CategoryColors;
}

const IntegrationFilter = ({
  categories,
  integrations,
  activeFilter,
  onFilterChange,
  categoryColors
}: IntegrationFilterProps) => {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6, delay: 0.2 }}
      className="relative"
    >
      {/* Background container with subtle gradient */}
      <div className="absolute inset-0 bg-gradient-to-r from-white/5 via-white/10 to-white/5 rounded-2xl blur-xl -mx-4 -my-2" />
      <div className="relative bg-white/5 backdrop-blur-sm border border-white/10 rounded-2xl p-6">
        <div className="flex flex-wrap justify-center gap-3">
          <Button
            variant={activeFilter === null ? "default" : "outline"}
            onClick={() => onFilterChange(null)}
            className={`transition-all duration-200 ${
              activeFilter === null
                ? "bg-[#6366f1] hover:bg-[#6366f1]/90"
                : "hover:bg-white/10"
            }`}
          >
            All Integrations
            <Badge variant="secondary" className="ml-2 bg-white/20 text-white">
              {integrations.length}
            </Badge>
          </Button>
          {categories.map((category) => {
            const categoryIntegrations = integrations.filter(i => i.category === category.name);
            const isActive = activeFilter === category.name;
            const categoryColor = categoryColors[category.name];

            return (
              <Button
                key={category.name}
                variant={isActive ? "default" : "outline"}
                onClick={() => onFilterChange(isActive ? null : category.name)}
                className={`transition-all duration-200 ${
                  isActive
                    ? "hover:opacity-90"
                    : "hover:bg-white/10"
                }`}
                style={isActive ? {
                  backgroundColor: categoryColor,
                  borderColor: categoryColor,
                } : undefined}
                onMouseEnter={(e) => {
                  if (isActive) {
                    e.currentTarget.style.backgroundColor = categoryColor + "dd";
                  }
                }}
                onMouseLeave={(e) => {
                  if (isActive) {
                    e.currentTarget.style.backgroundColor = categoryColor;
                  }
                }}
              >
                {category.name}
                <Badge variant="secondary" className="ml-2 bg-white/20 text-white">
                  {categoryIntegrations.length}
                </Badge>
              </Button>
            );
          })}
        </div>
      </div>
    </motion.div>
  );
};

export default IntegrationFilter;