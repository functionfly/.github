/**
 * Admin Badge Component
 *
 * A powerful animated badge for platform admin users featuring:
 * - Pulsing authority glow effect (red/amber)
 * - Animated shield icon with bounce
 * - Floating spark particles
 * - 3D hover tilt
 * - Authority shimmer effect
 */

import { cn } from '@/lib/utils';
import { motion, useAnimation, useMotionValue, useTransform } from 'framer-motion';
import { Lock, Shield, ShieldCheck, Sparkles, Zap } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';

export interface AdminBadgeProps {
  className?: string;
  size?: 'sm' | 'md' | 'lg';
  showParticles?: boolean;
  variant?: 'super_admin' | 'admin' | 'support';
}

const sizeConfig = {
  sm: {
    badge: 'h-6 px-2 text-xs gap-1',
    icon: 'w-3 h-3',
    shield: 'w-2.5 h-2.5',
    glow: 'blur-md',
  },
  md: {
    badge: 'h-7 px-2.5 text-sm gap-1.5',
    icon: 'w-3.5 h-3.5',
    shield: 'w-3 h-3',
    glow: 'blur-lg',
  },
  lg: {
    badge: 'h-8 px-3 text-sm gap-2',
    icon: 'w-4 h-4',
    shield: 'w-3.5 h-3.5',
    glow: 'blur-xl',
  },
};

const variantConfig = {
  super_admin: {
    label: 'Super Admin',
    gradient: 'from-red-600 via-rose-500 to-amber-500',
    glow: 'from-red-500 via-rose-400 to-amber-400',
    bg: 'from-red-950/90 via-rose-900/90 to-red-950/90',
    text: 'text-amber-100',
    iconColor: 'text-amber-400',
  },
  admin: {
    label: 'Admin',
    gradient: 'from-amber-500 via-orange-500 to-red-500',
    glow: 'from-amber-400 via-orange-400 to-red-400',
    bg: 'from-amber-950/90 via-orange-900/90 to-amber-950/90',
    text: 'text-amber-100',
    iconColor: 'text-amber-400',
  },
  support: {
    label: 'Support',
    gradient: 'from-blue-500 via-cyan-500 to-teal-500',
    glow: 'from-blue-400 via-cyan-400 to-teal-400',
    bg: 'from-blue-950/90 via-cyan-900/90 to-blue-950/90',
    text: 'text-cyan-100',
    iconColor: 'text-cyan-400',
  },
};

// Floating particle component
function Particle({
  delay,
  duration,
  x,
  color,
}: {
  delay: number;
  duration: number;
  x: number;
  color: string;
}) {
  return (
    <motion.div
      className={cn('absolute w-1 h-1 rounded-full', color)}
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

export function AdminBadge({
  className,
  size = 'md',
  showParticles = true,
  variant = 'admin',
}: AdminBadgeProps) {
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
  const vConfig = variantConfig[variant];

  // Generate random particles
  const particles = showParticles
    ? Array.from({ length: 6 }, (_, i) => ({
        id: i,
        delay: i * 0.3,
        duration: 2 + Math.random() * 1.5,
        x: 15 + i * 12,
      }))
    : [];

  const particleColor =
    variant === 'support'
      ? 'bg-gradient-to-r from-cyan-300 via-blue-200 to-cyan-400'
      : 'bg-gradient-to-r from-amber-300 via-yellow-200 to-amber-400';

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
      {/* Authority glow layers */}
      <motion.div
        className={cn(
          'absolute inset-0 rounded-full opacity-60',
          'bg-gradient-to-r',
          vConfig.glow,
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
          vConfig.glow.replace(/-400/g, '-400/50').replace(/-500/g, '-500/50'),
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
            background:
              variant === 'support'
                ? 'conic-gradient(from 0deg, #3b82f6, #06b6d4, #14b8a6, #3b82f6)'
                : 'conic-gradient(from 0deg, #ef4444, #f97316, #fbbf24, #ef4444)',
          }}
          animate={{ rotate: 360 }}
          transition={{
            duration: 8,
            repeat: Infinity,
            ease: 'linear',
          }}
        />
        <div className={cn('absolute inset-[1px] rounded-full bg-gradient-to-r', vConfig.bg)} />
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

        {/* Shield icon with bounce animation */}
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
          {variant === 'super_admin' ? (
            <ShieldCheck className={cn(vConfig.iconColor, config.shield)} />
          ) : (
            <Shield className={cn(vConfig.iconColor, config.shield)} />
          )}

          {/* Shield sparkle */}
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

        {/* Lock icon for super_admin */}
        {variant === 'super_admin' && (
          <motion.div
            animate={{
              scale: isHovered ? [1, 1.1, 1] : 1,
            }}
            transition={{ duration: 0.3 }}
          >
            <Lock className={cn('text-red-400', config.icon)} />
          </motion.div>
        )}

        {/* Text */}
        <span
          className={cn(
            'relative bg-gradient-to-r from-white via-amber-100 to-white bg-clip-text',
            vConfig.text
          )}
        >
          {vConfig.label}
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
          <Zap className={cn('w-3 h-3 fill-amber-400', vConfig.iconColor)} />
        </motion.div>

        {/* Floating particles container */}
        <div className="absolute inset-0 overflow-hidden rounded-full pointer-events-none">
          {particles.map((particle) => (
            <Particle
              key={particle.id}
              delay={particle.delay}
              duration={particle.duration}
              x={particle.x}
              color={particleColor}
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
            variant === 'support'
              ? 'bg-blue-950 border-blue-400/40'
              : 'bg-red-950 border-red-400/40'
          )}
          style={{ color: '#f8fafc' }}
        >
          {variant === 'super_admin' && 'Platform Super Administrator'}
          {variant === 'admin' && 'Platform Administrator'}
          {variant === 'support' && 'Platform Support Staff'}
        </div>
      </motion.div>
    </motion.div>
  );
}

// Compact version for inline usage
export function AdminBadgeCompact({
  className,
  variant = 'admin',
}: {
  className?: string;
  variant?: AdminBadgeProps['variant'];
}) {
  const vConfig = variantConfig[variant];

  return (
    <motion.span
      className={cn(
        'inline-flex items-center gap-1 px-1.5 py-0.5 text-xs font-semibold',
        'rounded-full border shadow-sm',
        'text-white',
        className
      )}
      style={{
        background:
          variant === 'support'
            ? 'linear-gradient(to right, #3b82f6, #06b6d4)'
            : 'linear-gradient(to right, #ef4444, #f97316)',
        borderColor: variant === 'support' ? 'rgba(59, 130, 246, 0.3)' : 'rgba(239, 68, 68, 0.3)',
      }}
      whileHover={{ scale: 1.05 }}
      animate={{
        boxShadow: [
          '0 0 0 0 rgba(239, 68, 68, 0.4)',
          '0 0 0 4px rgba(239, 68, 68, 0)',
          '0 0 0 0 rgba(239, 68, 68, 0.4)',
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
      {variant === 'super_admin' ? (
        <ShieldCheck className="w-3 h-3 text-amber-300" />
      ) : (
        <Shield className="w-3 h-3 text-amber-300" />
      )}
      <span>{variant === 'super_admin' ? 'SA' : variant === 'support' ? 'SPT' : 'ADM'}</span>
    </motion.span>
  );
}

export default AdminBadge;
