/**
 * Enterprise Badge Component
 *
 * A stunning animated badge for enterprise members featuring:
 * - Pulsing aurora glow effect
 * - Animated gradient border
 * - Floating particle sparkles
 * - 3D hover tilt
 * - Premium shimmer effect
 */

import { cn } from '@/lib/utils';
import { motion, useAnimation, useMotionValue, useTransform } from 'framer-motion';
import { Building2, Crown, Sparkles, Zap } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

export interface EnterpriseBadgeProps {
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  showParticles?: boolean;
}

const sizeConfig = {
  sm: {
    badge: 'h-6 px-2 text-xs gap-1',
    icon: 'w-3 h-3',
    crown: 'w-2.5 h-2.5',
    glow: 'blur-md',
  },
  md: {
    badge: 'h-7 px-2.5 text-sm gap-1.5',
    icon: 'w-3.5 h-3.5',
    crown: 'w-3 h-3',
    glow: 'blur-lg',
  },
  lg: {
    badge: 'h-8 px-3 text-sm gap-2',
    icon: 'w-4 h-4',
    crown: 'w-3.5 h-3.5',
    glow: 'blur-xl',
  },
};

// Floating particle component
function Particle({ delay, duration, x }: { delay: number; duration: number; x: number }) {
  return (
    <motion.div
      className="absolute w-1 h-1 rounded-full bg-gradient-to-r from-amber-300 via-yellow-200 to-amber-400"
      style={{ left: `${x}%`, bottom: 0 }}
      animate={{
        y: [-4, -24, -4],
        x: [0, Math.sin(x) * 8, 0],
        opacity: [0, 1, 0],
        scale: [0.5, 1.2, 0.5],
      }}
      transition={{
        duration,
        delay,
        repeat: Infinity,
        ease: 'easeInOut',
      }}
    />
  );
}

export function EnterpriseBadge({
  className,
  size = 'md',
  showParticles = true,
}: EnterpriseBadgeProps) {
  const controls = useAnimation();
  const badgeRef = useRef<HTMLDivElement>(null);
  const [isHovered, setIsHovered] = useState(false);

  // Mouse position for 3D tilt effect
  const mouseX = useMotionValue(0);
  const mouseY = useMotionValue(0);

  const rotateX = useTransform(mouseY, [-0.5, 0.5], [8, -8]);
  const rotateY = useTransform(mouseX, [-0.5, 0.5], [-8, 8]);

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!badgeRef.current) return;
    const rect = badgeRef.current.getBoundingClientRect();
    const x = (e.clientX - rect.left) / rect.width - 0.5;
    const y = (e.clientY - rect.top) / rect.height - 0.5;
    mouseX.set(x);
    mouseY.set(y);
  };

  const handleMouseLeave = () => {
    mouseX.set(0);
    mouseY.set(0);
    setIsHovered(false);
  };

  // Entrance animation
  useEffect(() => {
    controls.start({
      opacity: 1,
      scale: 1,
      y: 0,
      transition: {
        type: 'spring',
        stiffness: 400,
        damping: 20,
        delay: 0.2,
      },
    });
  }, [controls]);

  const config = sizeConfig[size];

  // Generate random particles
  const particles = showParticles
    ? Array.from({ length: 6 }, (_, i) => ({
        id: i,
        delay: i * 0.3,
        duration: 2 + Math.random() * 1.5,
        x: 15 + i * 12,
      }))
    : [];

  return (
    <motion.div
      ref={badgeRef}
      className={cn('relative inline-block', className)}
      initial={{ opacity: 0, scale: 0.8, y: 10 }}
      animate={controls}
      style={{
        perspective: 1000,
        transformStyle: 'preserve-3d',
      }}
      onMouseMove={handleMouseMove}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={handleMouseLeave}
    >
      {/* Aurora glow layers */}
      <motion.div
        className={cn(
          'absolute inset-0 rounded-full opacity-60',
          'bg-gradient-to-r from-indigo-500 via-purple-500 via-amber-400 to-indigo-500',
          'bg-[length:200%_100%]',
          config.glow
        )}
        animate={{
          backgroundPosition: ['0% 50%', '200% 50%', '0% 50%'],
          scale: isHovered ? [1, 1.15, 1.1] : [1, 1.08, 1],
          opacity: isHovered ? [0.6, 0.9, 0.7] : [0.5, 0.7, 0.5],
        }}
        transition={{
          backgroundPosition: {
            duration: 4,
            repeat: Infinity,
            ease: 'linear',
          },
          scale: {
            duration: 2,
            repeat: Infinity,
            ease: 'easeInOut',
          },
          opacity: {
            duration: 2,
            repeat: Infinity,
            ease: 'easeInOut',
          },
        }}
        style={{ transform: 'translateZ(-10px)' }}
      />

      {/* Secondary pulse glow */}
      <motion.div
        className={cn(
          'absolute -inset-1 rounded-full opacity-40',
          'bg-gradient-to-r from-amber-400/50 via-purple-500/50 to-indigo-500/50',
          'blur-md'
        )}
        animate={{
          scale: [1, 1.2, 1],
          opacity: [0.3, 0.5, 0.3],
        }}
        transition={{
          duration: 3,
          repeat: Infinity,
          ease: 'easeInOut',
          delay: 0.5,
        }}
      />

      {/* Rotating border gradient */}
      <motion.div
        className="absolute inset-0 rounded-full p-[1px] overflow-hidden"
        style={{
          rotateX,
          rotateY,
          transformStyle: 'preserve-3d',
        }}
      >
        <motion.div
          className="absolute inset-0 rounded-full"
          style={{
            background: 'conic-gradient(from 0deg, #6366f1, #a855f7, #fbbf24, #6366f1)',
          }}
          animate={{ rotate: 360 }}
          transition={{
            duration: 8,
            repeat: Infinity,
            ease: 'linear',
          }}
        />
        <div className="absolute inset-[1px] rounded-full bg-gradient-to-r from-indigo-950 via-purple-950 to-indigo-950" />
      </motion.div>

      {/* Main badge content */}
      <motion.div
        className={cn(
          'relative flex items-center rounded-full font-semibold',
          'bg-gradient-to-r from-indigo-900/90 via-purple-900/90 to-indigo-900/90',
          'backdrop-blur-sm border border-white/10',
          'text-white shadow-lg overflow-hidden',
          config.badge
        )}
        style={{
          rotateX,
          rotateY,
          transformStyle: 'preserve-3d',
        }}
        whileHover={{ scale: 1.05 }}
        whileTap={{ scale: 0.98 }}
      >
        {/* Shimmer overlay */}
        <motion.div
          className="absolute inset-0 bg-gradient-to-r from-transparent via-white/20 to-transparent"
          initial={{ x: '-100%' }}
          animate={{ x: '200%' }}
          transition={{
            duration: 2.5,
            repeat: Infinity,
            repeatDelay: 3,
            ease: 'easeInOut',
          }}
        />

        {/* Crown icon with bounce animation */}
        <motion.div
          className="relative"
          animate={{
            y: isHovered ? [0, -3, 0] : [0, -1, 0],
            rotate: isHovered ? [0, -10, 10, 0] : 0,
          }}
          transition={{
            duration: isHovered ? 0.5 : 2,
            repeat: isHovered ? 0 : Infinity,
            ease: 'easeInOut',
          }}
        >
          <Crown className={cn('text-amber-400', config.crown)} />

          {/* Crown sparkle */}
          <motion.div
            className="absolute -top-1 -right-1"
            animate={{
              scale: [0, 1, 0],
              opacity: [0, 1, 0],
              rotate: [0, 45, 90],
            }}
            transition={{
              duration: 1.5,
              repeat: Infinity,
              repeatDelay: 1,
              ease: 'easeOut',
            }}
          >
            <Sparkles className="w-2 h-2 text-yellow-300" />
          </motion.div>
        </motion.div>

        {/* Building icon */}
        <motion.div
          animate={{
            scale: isHovered ? [1, 1.1, 1] : 1,
          }}
          transition={{ duration: 0.3 }}
        >
          <Building2 className={cn('text-indigo-300', config.icon)} />
        </motion.div>

        {/* Text */}
        <span className="relative bg-gradient-to-r from-white via-indigo-100 to-white bg-clip-text">
          Enterprise
        </span>

        {/* Zap icon on hover */}
        <motion.div
          initial={{ opacity: 0, scale: 0, x: -5 }}
          animate={{
            opacity: isHovered ? 1 : 0,
            scale: isHovered ? 1 : 0,
            x: isHovered ? 0 : -5,
            rotate: isHovered ? [0, -15, 15, 0] : 0,
          }}
          transition={{ duration: 0.2 }}
        >
          <Zap className="w-3 h-3 text-amber-400 fill-amber-400" />
        </motion.div>

        {/* Floating particles container */}
        <div className="absolute inset-0 overflow-hidden rounded-full pointer-events-none">
          {particles.map((particle) => (
            <Particle
              key={particle.id}
              delay={particle.delay}
              duration={particle.duration}
              x={particle.x}
            />
          ))}
        </div>
      </motion.div>

      {/* Tooltip on hover */}
      <motion.div
        className="absolute -bottom-8 left-1/2 -translate-x-1/2 whitespace-nowrap"
        initial={{ opacity: 0, y: -5 }}
        animate={{
          opacity: isHovered ? 1 : 0,
          y: isHovered ? 0 : -5,
        }}
        transition={{ duration: 0.2 }}
      >
        <div
          className="px-2 py-1 text-xs font-medium rounded-md border border-indigo-400/40 bg-indigo-950 shadow-xl backdrop-blur-sm"
          style={{ color: '#f8fafc' }}
        >
          Premium Enterprise Member
        </div>
      </motion.div>
    </motion.div>
  );
}

// Compact version for inline usage
export function EnterpriseBadgeCompact({ className }: { className?: string }) {
  return (
    <motion.span
      className={cn(
        'inline-flex items-center gap-1 px-1.5 py-0.5 text-xs font-semibold',
        'rounded-full bg-gradient-to-r from-indigo-600 to-purple-600',
        'text-white border border-indigo-400/30 shadow-sm',
        className
      )}
      whileHover={{ scale: 1.05 }}
      animate={{
        boxShadow: [
          '0 0 0 0 rgba(99, 102, 241, 0.4)',
          '0 0 0 4px rgba(99, 102, 241, 0)',
          '0 0 0 0 rgba(99, 102, 241, 0.4)',
        ],
      }}
      transition={{
        boxShadow: {
          duration: 2,
          repeat: Infinity,
          ease: 'easeOut',
        },
      }}
    >
      <Crown className="w-3 h-3 text-amber-300" />
      <span>ENT</span>
    </motion.span>
  );
}

export default EnterpriseBadge;
