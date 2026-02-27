import { motion } from "framer-motion";
import { Play, Code, Zap } from "lucide-react";
import { useState } from "react";

interface InteractiveDemoMockProps {
  className?: string;
}

export function InteractiveDemoMock({ className = "" }: InteractiveDemoMockProps) {
  const [currentView, setCurrentView] = useState<"code" | "visual" | "metrics">("code");

  const views = {
    code: {
      icon: <Code className="w-5 h-5" />,
      title: "Code Execution",
      content: (
        <div className="space-y-4">
          <div className="bg-bg-secondary p-4 rounded-lg border border-border">
            <div className="flex items-center gap-2 mb-3">
              <div className="w-2 h-2 bg-red-500 rounded-full"></div>
              <div className="w-2 h-2 bg-yellow-500 rounded-full"></div>
              <div className="w-2 h-2 bg-green-500 rounded-full"></div>
            </div>
            <pre className="text-sm text-slate-800 dark:text-slate-200 overflow-x-auto">
{`// Agent execution with StateFabric
const agent = new StateFabricAgent({
  memory: "durable",
  replay: true,
  attribution: true
});

const result = await agent.run({
  task: "analyze customer sentiment",
  data: customerReviews
});

// Automatic state persistence
console.log("Execution saved:", result.id);`}
            </pre>
          </div>
          <motion.button
            className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
            whileHover={{ scale: 1.05 }}
            whileTap={{ scale: 0.95 }}
          >
            <Play className="w-4 h-4" />
            Run Code
          </motion.button>
        </div>
      )
    },
    visual: {
      icon: <Zap className="w-5 h-5" />,
      title: "Visual State Flow",
      content: (
        <div className="space-y-4">
          <div className="bg-bg-secondary p-6 rounded-lg border border-border">
            <div className="flex items-center justify-center space-x-4">
              <motion.div
                className="w-16 h-16 bg-blue-100 dark:bg-blue-900/20 rounded-lg flex items-center justify-center"
                animate={{ scale: [1, 1.1, 1] }}
                transition={{ duration: 2, repeat: Infinity }}
              >
                <span className="text-2xl">🧠</span>
              </motion.div>
              <motion.div
                className="w-8 h-0.5 bg-blue-500"
                animate={{ opacity: [0.5, 1, 0.5] }}
                transition={{ duration: 1.5, repeat: Infinity }}
              />
              <motion.div
                className="w-16 h-16 bg-green-100 dark:bg-green-900/20 rounded-lg flex items-center justify-center"
                animate={{ scale: [1, 1.1, 1] }}
                transition={{ duration: 2, repeat: Infinity, delay: 0.5 }}
              >
                <span className="text-2xl">💾</span>
              </motion.div>
              <motion.div
                className="w-8 h-0.5 bg-green-500"
                animate={{ opacity: [0.5, 1, 0.5] }}
                transition={{ duration: 1.5, repeat: Infinity, delay: 0.5 }}
              />
              <motion.div
                className="w-16 h-16 bg-purple-100 dark:bg-purple-900/20 rounded-lg flex items-center justify-center"
                animate={{ scale: [1, 1.1, 1] }}
                transition={{ duration: 2, repeat: Infinity, delay: 1 }}
              >
                <span className="text-2xl">🔄</span>
              </motion.div>
            </div>
            <div className="text-center mt-4 text-sm text-slate-600 dark:text-text-secondary">
              Memory → Persistence → Replay
            </div>
          </div>
        </div>
      )
    },
    metrics: {
      icon: <Code className="w-5 h-5" />,
      title: "Live Metrics",
      content: (
        <div className="grid grid-cols-2 gap-4">
          <motion.div
            className="bg-green-50 dark:bg-green-900/10 p-4 rounded-lg border border-green-200 dark:border-green-800"
            animate={{ opacity: [0.8, 1, 0.8] }}
            transition={{ duration: 2, repeat: Infinity }}
          >
            <div className="text-2xl font-bold text-green-700 dark:text-green-400">
              99.9%
            </div>
            <div className="text-sm text-green-600 dark:text-green-500">
              Uptime
            </div>
          </motion.div>
          <motion.div
            className="bg-blue-50 dark:bg-blue-900/10 p-4 rounded-lg border border-blue-200 dark:border-blue-800"
            animate={{ opacity: [0.8, 1, 0.8] }}
            transition={{ duration: 2, repeat: Infinity, delay: 0.5 }}
          >
            <div className="text-2xl font-bold text-blue-700 dark:text-blue-400">
              0.1ms
            </div>
            <div className="text-sm text-blue-600 dark:text-blue-500">
              Replay Latency
            </div>
          </motion.div>
          <motion.div
            className="bg-purple-50 dark:bg-purple-900/10 p-4 rounded-lg border border-purple-200 dark:border-purple-800"
            animate={{ opacity: [0.8, 1, 0.8] }}
            transition={{ duration: 2, repeat: Infinity, delay: 1 }}
          >
            <div className="text-2xl font-bold text-purple-700 dark:text-purple-400">
              $0.001
            </div>
            <div className="text-sm text-purple-600 dark:text-purple-500">
              Per Tool Call
            </div>
          </motion.div>
          <motion.div
            className="bg-orange-50 dark:bg-orange-900/10 p-4 rounded-lg border border-orange-200 dark:border-orange-800"
            animate={{ opacity: [0.8, 1, 0.8] }}
            transition={{ duration: 2, repeat: Infinity, delay: 1.5 }}
          >
            <div className="text-2xl font-bold text-orange-700 dark:text-orange-400">
              ∞
            </div>
            <div className="text-sm text-orange-600 dark:text-orange-500">
              Retention
            </div>
          </motion.div>
        </div>
      )
    }
  };

  return (
    <motion.div
      className={`bg-bg-secondary rounded-lg border border-border overflow-hidden ${className}`}
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6 }}
      viewport={{ once: true }}
    >
      {/* Header with tabs */}
      <div className="border-b border-border p-4">
        <div className="flex gap-2">
          {Object.entries(views).map(([key, view]) => (
            <motion.button
              key={key}
              onClick={() => setCurrentView(key as typeof currentView)}
              className={`
                flex items-center gap-2 px-4 py-2 rounded-lg text-sm font-medium transition-colors
                ${currentView === key
                  ? 'interactive-demo-tab-active bg-blue-100 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300'
                  : 'interactive-demo-tab-inactive text-slate-600 dark:text-slate-400 hover:bg-bg-primary'
                }
              `}
              whileHover={{ scale: 1.02 }}
              whileTap={{ scale: 0.98 }}
            >
              {view.icon}
              {view.title}
            </motion.button>
          ))}
        </div>
      </div>

      {/* Content */}
      <div className="p-6">
        {views[currentView].content}
      </div>
    </motion.div>
  );
}