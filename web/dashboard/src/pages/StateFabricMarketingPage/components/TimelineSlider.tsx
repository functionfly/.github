import { motion } from "framer-motion";
import { useState } from "react";

interface TimelineEvent {
  id: string;
  timestamp: string;
  title: string;
  description: string;
  type: "action" | "snapshot" | "error" | "cost";
}

interface TimelineSliderProps {
  events: TimelineEvent[];
  className?: string;
}

const eventColors = {
  action: "bg-blue-500",
  snapshot: "bg-green-500",
  error: "bg-red-500",
  cost: "bg-yellow-500"
};

export function TimelineSlider({ events, className = "" }: TimelineSliderProps) {
  const [selectedEvent, setSelectedEvent] = useState(0);

  return (
    <motion.div
      className={`bg-bg-secondary rounded-lg border border-border p-6 ${className}`}
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6 }}
      viewport={{ once: true }}
    >
      <h3 className="text-lg font-semibold text-slate-900 dark:text-white mb-6">
        Execution Timeline
      </h3>

      {/* Timeline visualization */}
      <div className="relative mb-8">
        <div className="flex justify-between items-center">
          {events.map((event, index) => (
            <motion.button
              key={event.id}
              onClick={() => setSelectedEvent(index)}
              className={`
                relative flex flex-col items-center
                ${selectedEvent === index ? 'z-10' : 'z-0'}
              `}
              whileHover={{ scale: 1.1 }}
              whileTap={{ scale: 0.95 }}
            >
              <motion.div
                className={`
                  w-4 h-4 rounded-full border-2 border-white shadow-lg
                  ${eventColors[event.type]}
                  ${selectedEvent === index ? 'ring-4 ring-blue-200' : ''}
                `}
                animate={{
                  scale: selectedEvent === index ? 1.2 : 1,
                }}
                transition={{ duration: 0.2 }}
              />
              <div className="text-xs text-slate-500 mt-2 text-center">
                {event.timestamp.split('T')[1]?.split('.')[0] || event.timestamp}
              </div>
            </motion.button>
          ))}
        </div>

        {/* Connection line */}
        <div className="absolute top-2 left-2 right-2 h-0.5 bg-border -z-10">
          <motion.div
            className="h-full bg-blue-500"
            style={{
              width: `${((selectedEvent + 1) / events.length) * 100}%`
            }}
            transition={{ duration: 0.3 }}
          />
        </div>
      </div>

      {/* Event details */}
      {events[selectedEvent] && (
        <motion.div
          className="bg-bg-primary rounded-lg p-4 border border-border"
          key={selectedEvent}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
        >
          <div className="flex items-start gap-3">
            <div className={`w-3 h-3 rounded-full mt-1.5 ${eventColors[events[selectedEvent].type]}`} />
            <div className="flex-1">
              <h4 className="font-semibold text-slate-900 dark:text-white mb-1">
                {events[selectedEvent].title}
              </h4>
              <p className="text-sm text-slate-600 dark:text-text-secondary mb-2">
                {events[selectedEvent].description}
              </p>
              <div className="text-xs text-slate-500">
                {events[selectedEvent].timestamp}
              </div>
            </div>
          </div>
        </motion.div>
      )}

      {/* Timeline scrubber */}
      <div className="mt-4">
        <input
          type="range"
          min="0"
          max={events.length - 1}
          value={selectedEvent}
          onChange={(e) => setSelectedEvent(Number(e.target.value))}
          className="w-full h-2 bg-border rounded-lg appearance-none cursor-pointer slider"
        />
        <style jsx>{`
          .slider::-webkit-slider-thumb {
            appearance: none;
            height: 16px;
            width: 16px;
            border-radius: 50%;
            background: #3b82f6;
            cursor: pointer;
          }
          .slider::-moz-range-thumb {
            height: 16px;
            width: 16px;
            border-radius: 50%;
            background: #3b82f6;
            cursor: pointer;
            border: none;
          }
        `}</style>
      </div>
    </motion.div>
  );
}