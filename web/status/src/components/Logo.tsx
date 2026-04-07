import { motion } from 'framer-motion';
import { cn } from '@/lib/utils';

interface LogoProps {
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  showText?: boolean;
  className?: string;
  animated?: boolean;
}

const sizeConfig = {
  xs: { icon: 16, text: 'text-sm' },
  sm: { icon: 20, text: 'text-lg' },
  md: { icon: 28, text: 'text-xl' },
  lg: { icon: 36, text: 'text-2xl' },
  xl: { icon: 48, text: 'text-3xl' },
};

export function Logo({ 
  size = 'md', 
  showText = true, 
  className,
  animated = false 
}: LogoProps) {
  const config = sizeConfig[size];

  return (
    <motion.div 
      className={cn('flex items-center gap-3', className)}
      whileHover={animated ? { scale: 1.02 } : undefined}
      transition={{ duration: 0.2 }}
    >
      <div className={cn(
        'relative flex items-center justify-center',
        animated && 'animate-float'
      )}>
        {/* Glow effect behind logo */}
        <motion.div 
          className={cn(
            'absolute inset-0 rounded-xl blur-xl opacity-50',
            'bg-gradient-to-r from-brand-500 via-purple-500 to-pink-500'
          )}
          animate={animated ? {
            opacity: [0.3, 0.6, 0.3],
            scale: [1, 1.1, 1],
          } : undefined}
          transition={animated ? {
            duration: 3,
            repeat: Infinity,
            ease: 'easeInOut',
          } : undefined}
        />
        
        {/* Logo SVG */}
        <motion.svg
          width={config.icon}
          height={config.icon}
          viewBox="0 0 32 32"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className="relative z-10"
          whileHover={animated ? { rotate: 360 } : undefined}
          transition={animated ? { duration: 0.6, ease: 'easeInOut' } : undefined}
        >
          <defs>
            <linearGradient id="logoGradient" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stopColor="#6366f1" />
              <stop offset="50%" stopColor="#8b5cf6" />
              <stop offset="100%" stopColor="#d946ef" />
            </linearGradient>
            <linearGradient id="logoGradient2" x1="100%" y1="0%" x2="0%" y2="100%">
              <stop offset="0%" stopColor="#10b981" />
              <stop offset="100%" stopColor="#34d399" />
            </linearGradient>
            <filter id="glow">
              <feGaussianBlur stdDeviation="2" result="coloredBlur"/>
              <feMerge>
                <feMergeNode in="coloredBlur"/>
                <feMergeNode in="SourceGraphic"/>
              </feMerge>
            </filter>
          </defs>
          
          {/* Hexagon background */}
          <motion.path
            d="M16 2L28 8.5V23.5L16 30L4 23.5V8.5L16 2Z"
            fill="url(#logoGradient)"
            opacity="0.2"
            stroke="url(#logoGradient)"
            strokeWidth="1"
            animate={animated ? {
              opacity: [0.2, 0.4, 0.2],
            } : undefined}
            transition={animated ? {
              duration: 2,
              repeat: Infinity,
              ease: 'easeInOut',
            } : undefined}
          />
          
          {/* Inner shape - flight path style */}
          <path
            d="M16 8L22 11.5V18.5L16 22L10 18.5V11.5L16 8Z"
            fill="url(#logoGradient)"
            opacity="0.6"
            filter={animated ? 'url(#glow)' : undefined}
          />
          
          {/* Center dot with pulse */}
          <motion.circle 
            cx="16" 
            cy="15" 
            r="3" 
            fill="url(#logoGradient2)" 
            animate={animated ? {
              r: [3, 4, 3],
              opacity: [1, 0.8, 1],
            } : undefined}
            transition={animated ? {
              duration: 2,
              repeat: Infinity,
              ease: 'easeInOut',
            } : undefined}
          />
          
          {/* Outer orbital ring */}
          <motion.circle 
            cx="16" 
            cy="15" 
            r="6" 
            stroke="url(#logoGradient)" 
            strokeWidth="0.5"
            fill="none"
            opacity="0.4"
            animate={animated ? {
              rotate: [0, 360],
            } : undefined}
            transition={animated ? {
              duration: 10,
              repeat: Infinity,
              ease: 'linear',
            } : undefined}
            style={{ transformOrigin: 'center' }}
          />
          
          {/* Small orbital dot */}
          <motion.circle 
            cx="24" 
            cy="15" 
            r="1.5" 
            fill="white" 
            opacity="0.8"
            animate={animated ? {
              rotate: [0, 360],
            } : undefined}
            transition={animated ? {
              duration: 10,
              repeat: Infinity,
              ease: 'linear',
            } : undefined}
            style={{ transformOrigin: '16px 15px' }}
          />
        </motion.svg>
      </div>
      
      {showText && (
        <span className={cn(
          'font-bold tracking-tight gradient-text',
          config.text
        )}>
          FunctionFly
        </span>
      )}
    </motion.div>
  );
}

export function LogoSimple({ size = 'md', className }: { size?: 'sm' | 'md' | 'lg'; className?: string }) {
  const sizeMap = { sm: 20, md: 28, lg: 36 };
  
  return (
    <svg
      width={sizeMap[size]}
      height={sizeMap[size]}
      viewBox="0 0 32 32"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      className={className}
    >
      <defs>
        <linearGradient id="logoGradientSimple" x1="4" y1="2" x2="28" y2="30">
          <stop stopColor="#6366f1" />
          <stop offset="0.5" stopColor="#8b5cf6" />
          <stop offset="1" stopColor="#d946ef" />
        </linearGradient>
      </defs>
      <path
        d="M16 2L28 8.5V23.5L16 30L4 23.5V8.5L16 2Z"
        fill="url(#logoGradientSimple)"
        opacity="0.3"
      />
      <path
        d="M16 8L22 11.5V18.5L16 22L10 18.5V11.5L16 8Z"
        fill="url(#logoGradientSimple)"
      />
      <circle cx="16" cy="15" r="3" fill="#10b981" />
    </svg>
  );
}

export function LogoLoading({ size = 'lg' }: { size?: 'sm' | 'md' | 'lg' | 'xl' }) {
  const sizes = { sm: 32, md: 48, lg: 64, xl: 96 };
  const dim = sizes[size];
  
  return (
    <div className="relative flex items-center justify-center">
      {/* Outer spinning ring */}
      <motion.div
        className="absolute rounded-full border-2 border-transparent border-t-brand-500 border-r-purple-500"
        style={{ width: dim, height: dim }}
        animate={{ rotate: 360 }}
        transition={{ duration: 2, repeat: Infinity, ease: 'linear' }}
      />
      
      {/* Inner spinning ring */}
      <motion.div
        className="absolute rounded-full border-2 border-transparent border-b-brand-500 border-l-purple-500"
        style={{ width: dim * 0.7, height: dim * 0.7 }}
        animate={{ rotate: -360 }}
        transition={{ duration: 1.5, repeat: Infinity, ease: 'linear' }}
      />
      
      {/* Center logo */}
      <div style={{ width: dim * 0.5, height: dim * 0.5 }}>
        <LogoSimple size="md" className="w-full h-full" />
      </div>
    </div>
  );
}
