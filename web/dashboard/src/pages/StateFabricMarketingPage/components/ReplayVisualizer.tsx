import { motion } from "framer-motion";
import { Play, Pause, RotateCcw, SkipBack, SkipForward } from "lucide-react";
import { useState, useEffect } from "react";

interface ReplayStep {
  id: string;
  timestamp: string;
  action: string;
  state: Record<string, any>;
  cost?: number;
}

interface ReplayVisualizerProps {
  steps: ReplayStep[];
  className?: string;
}

export function ReplayVisualizer({ steps, className = "" }: ReplayVisualizerProps) {
  const [currentStep, setCurrentStep] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [speed, setSpeed] = useState(1000);

  useEffect(() => {
    if (!isPlaying) return;

    const interval = setInterval(() => {
      setCurrentStep((prev) => {
        if (prev >= steps.length - 1) {
          setIsPlaying(false);
          return prev;
        }
        return prev + 1;
      });
    }, speed);

    return () => clearInterval(interval);
  }, [isPlaying, speed, steps.length]);

  const handlePlay = () => setIsPlaying(!isPlaying);
  const handleReset = () => {
    setIsPlaying(false);
    setCurrentStep(0);
  };
  const handlePrev = () => {
    setCurrentStep(Math.max(0, currentStep - 1));
    setIsPlaying(false);
  };
  const handleNext = () => {
    setCurrentStep(Math.min(steps.length - 1, currentStep + 1));
    setIsPlaying(false);
  };

  const currentStepData = steps[currentStep];
  const progress = ((currentStep + 1) / steps.length) * 100;

  return (
    <motion.div
      className={`bg-bg-secondary rounded-lg border border-border p-6 ${className}`}
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6 }}
      viewport={{ once: true }}
    >
      <div className="flex items-center justify-between mb-6">
        <h3 className="text-lg font-semibold text-slate-900 dark:text-white">
          Execution Replay
        </h3>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setSpeed(2000)}
            className={`px-2 py-1 text-xs rounded ${speed === 2000 ? 'bg-blue-100 text-blue-700' : 'text-slate-500'}`}
          >
            0.5x
          </button>
          <button
            onClick={() => setSpeed(1000)}
            className={`px-2 py-1 text-xs rounded ${speed === 1000 ? 'bg-blue-100 text-blue-700' : 'text-slate-500'}`}
          >
            1x
          </button>
          <button
            onClick={() => setSpeed(500)}
            className={`px-2 py-1 text-xs rounded ${speed === 500 ? 'bg-blue-100 text-blue-700' : 'text-slate-500'}`}
          >
            2x
          </button>
        </div>
      </div>

      {/* Progress Bar */}
      <div className="mb-6">
        <div className="w-full bg-bg-primary rounded-full h-2 mb-2">
          <motion.div
            className="bg-blue-500 h-2 rounded-full"
            style={{ width: `${progress}%` }}
            transition={{ duration: 0.3 }}
          />
        </div>
        <div className="flex justify-between text-xs text-slate-500">
          <span>Step {currentStep + 1} of {steps.length}</span>
          <span>{currentStepData?.timestamp}</span>
        </div>
      </div>

      {/* Controls */}
      <div className="flex items-center justify-center gap-4 mb-6">
        <motion.button
          onClick={handleReset}
          className="p-2 rounded-lg bg-bg-primary hover:bg-border transition-colors"
          whileHover={{ scale: 1.1 }}
          whileTap={{ scale: 0.95 }}
        >
          <RotateCcw className="w-4 h-4 text-slate-600" />
        </motion.button>

        <motion.button
          onClick={handlePrev}
          className="p-2 rounded-lg bg-bg-primary hover:bg-border transition-colors"
          whileHover={{ scale: 1.1 }}
          whileTap={{ scale: 0.95 }}
        >
          <SkipBack className="w-4 h-4 text-slate-600" />
        </motion.button>

        <motion.button
          onClick={handlePlay}
          className="p-3 rounded-lg bg-blue-500 hover:bg-blue-600 text-white transition-colors"
          whileHover={{ scale: 1.1 }}
          whileTap={{ scale: 0.95 }}
        >
          {isPlaying ? (
            <Pause className="w-5 h-5" />
          ) : (
            <Play className="w-5 h-5" />
          )}
        </motion.button>

        <motion.button
          onClick={handleNext}
          className="p-2 rounded-lg bg-bg-primary hover:bg-border transition-colors"
          whileHover={{ scale: 1.1 }}
          whileTap={{ scale: 0.95 }}
        >
          <SkipForward className="w-4 h-4 text-slate-600" />
        </motion.button>
      </div>

      {/* Current Step Display */}
      {currentStepData && (
        <motion.div
          className="bg-bg-primary rounded-lg p-4 border border-border"
          key={currentStep}
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.3 }}
        >
          <div className="flex justify-between items-start mb-3">
            <h4 className="font-semibold text-slate-900 dark:text-white">
              {currentStepData.action}
            </h4>
            {currentStepData.cost && (
              <span className="text-sm text-green-600 dark:text-green-400 font-medium">
                ${currentStepData.cost.toFixed(4)}
              </span>
            )}
          </div>

          <div className="text-sm text-slate-600 dark:text-text-secondary">
            <strong>State:</strong>
            <pre className="mt-1 bg-bg-secondary p-2 rounded text-xs overflow-x-auto">
              {JSON.stringify(currentStepData.state, null, 2)}
            </pre>
          </div>
        </motion.div>
      )}
    </motion.div>
  );
}