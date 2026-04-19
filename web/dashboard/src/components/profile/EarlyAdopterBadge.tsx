/**
 * Early Adopter Badge Component
 *
 * A distinctive animated badge for early community members featuring:
 * - Heritage gold/amber glow effect
 * - Rocket icon with launch animation
 * - Floating spark particles
 * - 3D hover tilt
 * - Member number display (e.g., "Member #123")
 * - Special styling for single-digit member numbers (founding members)
 */

import { cn } from '@/lib/utils';
import { motion, useAnimation, useMotionValue, useTransform } from 'framer-motion';
import { Rocket, Sparkles, Star, Zap } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

export interface EarlyAdopterBadgeProps {
  profileNumber: number;
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  showParticles?: boolean;
}

const sizeConfig = {
  sm: {
    badge: 'h-6 px-2 text-xs gap-1',
    icon: 'w-3 h-3',
    number: 'text-[10px]',
    glow: 'blur-md',
  },
  md: {
    badge: 'h-7 px-2.5 text-sm gap-1.5',
    icon: 'w-3.5 h-3.5',
    number: 'text-xs',
    glow: 'blur-lg',
  },
  lg: {
    badge: 'h-8 px-3 text-sm gap-2',
    icon: 'w-4 h-4',
    number: 'text-sm',
    glow: 'blur-xl',
  },
};

// Special styling for founding members (single digit)
const isFoundingMember = (num: number) => num >= 1 && num <= 9;

// Special styling for early batch (two digits)
const isEarlyBatch = (num: number) => num >= 10 && num <= 99;

// Floating particle component
function Particle({
  delay,
  duration,
  x,
}: {
  delay: number;
  duration: number;
  x: number;
}) {
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

export function EarlyAdopterBadge({
  profileNumber,
  className,
  size = 'md',
  showParticles = true,
}: EarlyAdopterBadgeProps) {
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
  const founding = isFoundingMember(profileNumber);
  const early = isEarlyBatch(profileNumber);

  // Generate random particles
  const particles = showParticles
    ? Array.from({ length: 5 }, (_, i) => ({
        id: i,
        delay: i * 0.4,
        duration: 2 + Math.random() * 1.5,
        x: 20 + i * 14,
      }))
    : [];

  // Gradient based on member status
  const gradient = founding
    ? 'from-rose-500 via-amber-500 to-yellow-400'
    : early
      ? 'from-amber-500 via-orange-500 to-yellow-400'
      : 'from-amber-400 via-yellow-400 to-amber-500';

  const glowGradient = founding
    ? 'from-rose-400 via-amber-400 to-yellow-300'
    : early
      ? 'from-amber-400 via-orange-400 to-yellow-300'
      : 'from-amber-300 via-yellow-300 to-amber-400';

  const bgGradient = founding
    ? 'from-amber-950/95 via-orange-900/95 to-amber-950/95'
    : 'from-amber-950/90 via-orange-900/90 to-amber-950/90';

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
      {/* Heritage glow layers */}
      <motion.div
        className={cn(
          'absolute inset-0 rounded-full opacity-60',
          'bg-gradient-to-r',
          glowGradient,
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
          'bg-gradient-to-r',
          glowGradient.replace(/-400/g, '-400/50').replace(/-300/g, '-300/50'),
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
            background: founding
              ? 'conic-gradient(from 0deg, #f43f5e, #f97316, #fbbf24, #f43f5e)'
              : 'conic-gradient(from 0deg, #f59e0b, #f97316, #fbbf24, #f59e0b)',
          }}
          animate={{ rotate: 360 }}
          transition={{
            duration: founding ? 6 : 8,
            repeat: Infinity,
            ease: 'linear',
          }}
        />
        <div className={cn('absolute inset-[1px] rounded-full bg-gradient-to-r', bgGradient)} />
      </motion.div>

      {/* Main badge content */}
      <motion.div
        className={cn(
          'relative flex items-center rounded-full font-semibold',
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

        {/* Rocket icon with launch animation */}
        <motion.div
          className="relative"
          animate={{
            y: isHovered ? [0, -4, 0] : [0, -1, 0],
            rotate: isHovered ? [0, -15, -30, 0] : [0, -5, 0],
          }}
          transition={{
            duration: isHovered ? 0.6 : 2,
            repeat: isHovered ? 0 : Infinity,
            ease: 'easeInOut',
          }}
        >
          <Rocket className={cn('text-amber-300', config.icon)} />

          {/* Launch sparkle */}
          <motion.div
            className="absolute -bottom-2 left-1/2 -translate-x-1/2"
            animate={{
              scale: isHovered ? [0, 1, 0] : [0.5, 0.8, 0.5],
              opacity: isHovered ? [0, 1, 0] : [0.3, 0.5, 0.3],
              y: isHovered ? [0, 8, 12] : [0, 2, 4],
            }}
            transition={{
              duration: isHovered ? 0.8 : 1.5,
              repeat: isHovered ? 0 : Infinity,
              ease: 'easeOut',
            }}
          >
            <Zap className="w-2 h-2 text-orange-400 fill-orange-400" />
          </motion.div>
        </motion.div>

        {/* Founding member star */}
        {founding && (
          <motion.div
            initial={{ scale: 0, rotate: -180 }}
            animate={{ scale: 1, rotate: 0 }}
            transition={{
              type: 'spring',
              stiffness: 500,
              damping: 15,
              delay: 0.5,
            }}
          >
            <Star className="w-3 h-3 text-yellow-300 fill-yellow-300" />
          </motion.div>
        )}

        {/* Member number text */}
        <span
          className={cn(
            'relative bg-gradient-to-r from-white via-amber-100 to-white bg-clip-text',
            'text-amber-100',
            config.number
          )}
        >
          #{profileNumber}
        </span>

        {/* Sparkle on hover */}
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
          <Sparkles className="w-3 h-3 fill-amber-400 text-amber-400" />
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
          className={cn(
            'px-2 py-1 text-xs font-medium rounded-md border backdrop-blur-sm shadow-xl',
            'bg-amber-950 border-amber-400/40 text-amber-100'
          )}
        >
          {founding && 'Founding Member - Early Adopter'}
          {early && 'Early Adopter - Beta Member'}
          {!founding && !early && 'Early Adopter'}
        </div>
      </motion.div>
    </motion.div>
  );
}

// Compact version for inline usage
export function EarlyAdopterBadgeCompact({
  profileNumber,
  className,
}: {
  profileNumber: number;
  className?: string;
}) {
  const founding = isFoundingMember(profileNumber);
  const early = isEarlyBatch(profileNumber);

  return (
    <motion.span
      className={cn(
        'inline-flex items-center gap-1 px-1.5 py-0.5 text-xs font-semibold',
        'rounded-full border shadow-sm',
        'text-white',
        className
      )}
      style={{
        background: founding
          ? 'linear-gradient(to right, #f43f5e, #f97316, #fbbf24)'
          : early
            ? 'linear-gradient(to right, #f59e0b, #f97316)'
            : 'linear-gradient(to right, #f59e0b, #fbbf24)',
        borderColor: founding
          ? 'rgba(251, 191, 36, 0.5)'
          : early
            ? 'rgba(249, 115, 22, 0.3)'
            : 'rgba(245, 158, 11, 0.3)',
      }}
      whileHover={{ scale: 1.05 }}
      animate={{
        boxShadow: [
          '0 0 0 0 rgba(251, 191, 36, 0.4)',
          '0 0 0 4px rgba(251, 191, 36, 0)',
          '0 0 0 0 rgba(251, 191, 36, 0.4)',
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
      <Rocket className="w-3 h-3 text-amber-100" />
      {founding && <Star className="w-2 h-2 text-yellow-200 fill-yellow-200" />}
      <span>#{profileNumber}</span>
    </motion.span>
  );
}

export default EarlyAdopterBadge;
