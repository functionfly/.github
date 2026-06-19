import { useEffect, useState } from "react";
import { useInView } from "react-intersection-observer";
import { useGesture } from "@use-gesture/react";
import toast from "react-hot-toast";

// Custom hook for scroll-triggered animations
export function useScrollAnimation(threshold = 0.1, triggerOnce = true) {
  const [ref, inView] = useInView({
    threshold,
    triggerOnce,
  });

  const [hasAnimated, setHasAnimated] = useState(false);

  useEffect(() => {
    if (inView && !hasAnimated) {
      setHasAnimated(true);
    }
  }, [inView, hasAnimated]);

  return { ref, inView: hasAnimated || inView };
}

// Custom hook for card gesture interactions
export function useCardGestures(planName?: string) {
  const [isHovered, setIsHovered] = useState(false);
  const [isDragging, setIsDragging] = useState(false);
  const [dragOffset, setDragOffset] = useState({ x: 0, y: 0 });
  const [scale, setScale] = useState(1);
  const [hasShownHoverToast, setHasShownHoverToast] = useState(false);

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const bind = useGesture({
    onHover: ({ hovering }: any) => {
      const wasHovered = isHovered;
      setIsHovered(hovering ?? false);

      // Show toast on first hover
      if (hovering && !wasHovered && !hasShownHoverToast && planName) {
        toast(`Exploring ${planName} plan...`, {
          duration: 2000,
          style: {
            background: '#1a1a1a',
            color: '#fff',
            border: '#1px solid #6366f1',
          },
          icon: '👀',
        });
        setHasShownHoverToast(true);
      }
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    onDrag: ({ active, movement }: any) => {
      setIsDragging(active);
      if (active) {
        setDragOffset({ x: movement[0], y: movement[1] });
        setScale(1.05);
        toast(`Dragging ${planName || 'plan'} card!`, {
          duration: 1500,
          style: {
            background: '#1a1a1a',
            color: '#fff',
            border: '#1px solid #06b6d4',
          },
          icon: '🎯',
        });
      } else {
        setDragOffset({ x: 0, y: 0 });
        setScale(isHovered ? 1.02 : 1);
      }
    },
    onDragEnd: () => {
      setScale(isHovered ? 1.02 : 1);
    }
  });

  useEffect(() => {
    setScale(isHovered && !isDragging ? 1.02 : 1);
  }, [isHovered, isDragging]);

  return {
    bind,
    isHovered,
    isDragging,
    dragOffset,
    scale,
    style: {
      transform: `translate(${dragOffset.x}px, ${dragOffset.y}px) scale(${scale})`,
      cursor: isDragging ? 'grabbing' : isHovered ? 'grab' : 'default',
      zIndex: isDragging ? 10 : isHovered ? 5 : 1,
    }
  };
}
