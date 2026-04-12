/**
 * LiveCursors Component
 * Show fake collaborative cursors for demo purposes
 * Cursors move around randomly with smooth animation
 */

import { useEffect, useState, useRef } from 'react';
import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';

interface Cursor {
  id: string;
  name: string;
  color: string;
  x: number;
  y: number;
  isActive: boolean;
}

interface LiveCursorsProps {
  containerRef?: React.RefObject<HTMLElement | null>;
  className?: string;
}

const MOCK_CURSORS: Omit<Cursor, 'x' | 'y' | 'isActive'>[] = [
  {
    id: '1',
    name: 'Alex',
    color: '#6366f1', // brand-500
  },
  {
    id: '2',
    name: 'Sam',
    color: '#10b981', // success
  },
  {
    id: '3',
    name: 'Jordan',
    color: '#f59e0b', // warning
  },
];

// Generate random position within bounds
function getRandomPosition(width: number, height: number) {
  const padding = 100;
  return {
    x: padding + Math.random() * (width - padding * 2),
    y: padding + Math.random() * (height - padding * 2),
  };
}

function CursorComponent({ cursor }: { cursor: Cursor }) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0 }}
      animate={{ 
        opacity: cursor.isActive ? 1 : 0.3, 
        scale: 1,
        x: cursor.x,
        y: cursor.y,
      }}
      transition={{
        type: 'spring',
        damping: 30,
        stiffness: 200,
        mass: 0.8,
      }}
      className="absolute pointer-events-none z-50"
      style={{ left: 0, top: 0 }}
    >
      {/* Cursor Icon */}
      <svg
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        style={{
          filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.2))',
        }}
      >
        <path
          d="M5.5 3.21V20.8c0 .45.56.67.85.35l4.86-5.52a.5.5 0 0 1 .38-.15h6.87c.44 0 .66-.53.35-.85L6.35 2.85a.5.5 0 0 0-.85.36Z"
          fill={cursor.color}
          stroke="white"
          strokeWidth="1.5"
        />
      </svg>

      {/* Name Label */}
      <div
        className={cn(
          "absolute left-4 top-4 px-2 py-1 rounded-md text-xs font-medium whitespace-nowrap",
          "text-white shadow-sm"
        )}
        style={{ backgroundColor: cursor.color }}
      >
        {cursor.name}
      </div>
    </motion.div>
  );
}

export function LiveCursors({ containerRef, className }: LiveCursorsProps) {
  const [cursors, setCursors] = useState<Cursor[]>([]);
  const [dimensions, setDimensions] = useState({ width: 1000, height: 600 });
  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  useEffect(() => {
    // Initialize cursors with random positions
    const initialCursors = MOCK_CURSORS.map((mock) => ({
      ...mock,
      ...getRandomPosition(1000, 600),
      isActive: true,
    }));
    setCursors(initialCursors);

    // Update dimensions based on container
    const updateDimensions = () => {
      const ref = containerRef?.current;
      if (ref) {
        const rect = ref.getBoundingClientRect();
        setDimensions({ width: rect.width, height: rect.height });
      } else {
        setDimensions({ width: window.innerWidth, height: window.innerHeight });
      }
    };

    updateDimensions();
    window.addEventListener('resize', updateDimensions);

    // Move cursors randomly every 2-4 seconds
    intervalRef.current = setInterval(() => {
      setCursors((prev) =>
        prev.map((cursor) => {
          // 80% chance to move, 20% chance to stay idle
          if (Math.random() > 0.2) {
            const newPos = getRandomPosition(dimensions.width, dimensions.height);
            return {
              ...cursor,
              x: newPos.x,
              y: newPos.y,
              isActive: true,
            };
          }
          return { ...cursor, isActive: false };
        })
      );
    }, 2000 + Math.random() * 2000);

    return () => {
      window.removeEventListener('resize', updateDimensions);
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [containerRef, dimensions.width, dimensions.height]);

  // Periodically make cursors active again
  useEffect(() => {
    const activationInterval = setInterval(() => {
      setCursors((prev) =>
        prev.map((cursor) => ({ ...cursor, isActive: true }))
      );
    }, 5000);

    return () => clearInterval(activationInterval);
  }, []);

  return (
    <div className={cn("absolute inset-0 pointer-events-none overflow-hidden", className)}>
      {cursors.map((cursor) => (
        <CursorComponent key={cursor.id} cursor={cursor} />
      ))}
    </div>
  );
}

export default LiveCursors;
