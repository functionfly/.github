import { RUNTIME_COLORS } from '../constants';

interface TrustGaugeProps {
  score: number;
  runtime?: string;
  size?: number;
}

export function TrustGauge({ score, runtime = 'python', size = 44 }: TrustGaugeProps) {
  const colors = RUNTIME_COLORS[runtime] || RUNTIME_COLORS.python;
  const radius = (size - 6) / 2;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference - (score / 100) * circumference;
  const color = score >= 90 ? '#00ff9d' : score >= 75 ? colors.glow : '#ffb800';

  return (
    <div className="flyway-trust-ring" style={{ width: size, height: size }}>
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="rgba(255,255,255,0.08)"
          strokeWidth={3}
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke={color}
          strokeWidth={3}
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          strokeLinecap="round"
          style={{ filter: `drop-shadow(0 0 4px ${color})` }}
        />
      </svg>
      <span className="flyway-trust-ring-value" style={{ color }}>
        {Math.round(score)}
      </span>
    </div>
  );
}
