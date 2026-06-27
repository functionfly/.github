import { type ReactNode } from 'react';
import { motion, useReducedMotion } from 'framer-motion';

interface CardProps {
  children: ReactNode;
  className?: string;
  style?: React.CSSProperties;
  animate?: boolean;
  staggerDelay?: number;
}

const cardVariants = {
  hidden: {
    opacity: 0,
    y: 12,
  },
  visible: (custom: number) => ({
    opacity: 1,
    y: 0,
    transition: {
      duration: 0.2,
      delay: custom,
      ease: [0.16, 1, 0.3, 1],
    },
  }),
};

const flareVariants = {
  rest: { x: '-100%', opacity: 0 },
  hover: {
    x: '200%',
    opacity: [0, 1, 0],
    transition: {
      duration: 0.8,
      ease: [0.16, 1, 0.3, 1],
    },
  },
};

const cornerAccentVariants = {
  rest: { scaleX: 0, opacity: 0 },
  hover: (direction: 'left' | 'right') => ({
    scaleX: 1,
    opacity: 1,
    transition: {
      duration: 0.25,
      ease: [0.16, 1, 0.3, 1],
    },
  }),
};

export function Card({ children, className = '', style = {}, animate = true, staggerDelay = 0 }: CardProps) {
  const prefersReducedMotion = useReducedMotion();

  if (prefersReducedMotion || !animate) {
    return (
      <div
        className={`card ${className}`}
        style={style}
      >
        {children}
      </div>
    );
  }

  return (
    <motion.div
      className={`card ${className}`}
      style={style}
      variants={cardVariants}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true, margin: '-40px' }}
      whileHover="hover"
      custom={staggerDelay}
    >
      {/* Scanning flare line */}
      <motion.div
        className="card-flare"
        variants={flareVariants}
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '60%',
          height: '1px',
          background: 'linear-gradient(90deg, transparent, var(--foil-a), var(--foil-b), var(--foil-c), var(--foil-d), transparent)',
          pointerEvents: 'none',
          zIndex: 10,
          filter: 'blur(0.5px)',
        }}
      />

      {/* Top left corner accent */}
      <motion.div
        variants={cornerAccentVariants}
        custom="left"
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '24px',
          height: '2px',
          background: 'linear-gradient(90deg, var(--foil-a), var(--foil-b))',
          pointerEvents: 'none',
          zIndex: 10,
          transformOrigin: 'left',
        }}
      />
      <motion.div
        variants={cornerAccentVariants}
        custom="left"
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          width: '2px',
          height: '16px',
          background: 'linear-gradient(180deg, var(--foil-b), var(--foil-c))',
          pointerEvents: 'none',
          zIndex: 10,
          transformOrigin: 'top',
        }}
      />

      {/* Top right corner accent */}
      <motion.div
        variants={cornerAccentVariants}
        custom="right"
        style={{
          position: 'absolute',
          top: 0,
          right: 0,
          width: '24px',
          height: '2px',
          background: 'linear-gradient(270deg, var(--foil-c), var(--foil-d))',
          pointerEvents: 'none',
          zIndex: 10,
          transformOrigin: 'right',
        }}
      />
      <motion.div
        variants={cornerAccentVariants}
        custom="right"
        style={{
          position: 'absolute',
          top: 0,
          right: 0,
          width: '2px',
          height: '16px',
          background: 'linear-gradient(180deg, var(--foil-c), var(--foil-d))',
          pointerEvents: 'none',
          zIndex: 10,
          transformOrigin: 'top',
        }}
      />

      {/* Subtle top inner glow */}
      <div
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          height: '60px',
          background: 'linear-gradient(180deg, rgba(159, 216, 255, 0.06) 0%, transparent 100%)',
          pointerEvents: 'none',
          zIndex: 1,
        }}
      />

      <div style={{ position: 'relative', zIndex: 2 }}>
        {children}
      </div>
    </motion.div>
  );
}
