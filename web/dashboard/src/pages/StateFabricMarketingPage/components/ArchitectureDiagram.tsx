import { motion } from "framer-motion";
import { Zap, FileText, Camera, Database, RotateCcw } from "lucide-react";
import { LayerCard } from "./LayerCard";
import { AnimatedFlowArrow } from "./AnimatedFlowArrow";

const flowSteps = [
  {
    title: "Function Execution",
    icon: <Zap className="w-6 h-6 text-yellow-500" />,
    delay: 0.1
  },
  {
    title: "Event Log",
    icon: <FileText className="w-6 h-6 text-blue-500" />,
    delay: 0.2
  },
  {
    title: "Snapshot",
    icon: <Camera className="w-6 h-6 text-green-500" />,
    delay: 0.3
  },
  {
    title: "Object Storage",
    icon: <Database className="w-6 h-6 text-purple-500" />,
    delay: 0.4
  },
  {
    title: "Replay Engine",
    icon: <RotateCcw className="w-6 h-6 text-orange-500" />,
    delay: 0.5
  }
];

export function ArchitectureDiagram() {
  return (
    <div className="w-full architecture-diagram">
      {/* Main Flow */}
      <div className="flex flex-col lg:flex-row items-center justify-center gap-4 lg:gap-8 mb-8">
        {flowSteps.map((step, index) => (
          <div key={index} className="flex items-center">
            <LayerCard
              title={step.title}
              icon={step.icon}
              delay={step.delay}
            />
            {index < flowSteps.length - 1 && (
              <div className="hidden lg:block">
                <AnimatedFlowArrow delay={step.delay + 0.2} />
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Mobile arrows */}
      <div className="flex lg:hidden justify-center gap-4 mb-8">
        {Array.from({ length: 4 }).map((_, index) => (
          <AnimatedFlowArrow key={index} delay={0.6 + index * 0.1} />
        ))}
      </div>

      {/* Supporting Infrastructure */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.7, duration: 0.6 }}
          viewport={{ once: true }}
        >
          <LayerCard
            title="Metadata Index"
            subtitle="Postgres"
            className="h-full"
          />
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.8, duration: 0.6 }}
          viewport={{ once: true }}
        >
          <LayerCard
            title="Object Storage"
            subtitle="Events + Snapshots"
            className="h-full"
          />
        </motion.div>
      </div>
    </div>
  );
}