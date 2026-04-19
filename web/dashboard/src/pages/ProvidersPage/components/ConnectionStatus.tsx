import { Badge } from '@/components/ui/badge';

interface ConnectionStatusProps {
  connected: boolean;
  status?: string;
  animate?: boolean;
}

export function ConnectionStatus({ connected, status, animate = true }: ConnectionStatusProps) {
  // Map status to aviation-themed glow classes
  const getStatusDotClass = () => {
    if (!connected) return 'status-dot-runway pending';

    switch (status) {
      case 'online':
        return 'status-dot-runway online';
      case 'degraded':
        return 'status-dot-runway degraded';
      case 'offline':
        return 'status-dot-runway offline';
      default:
        return 'status-dot-runway pending';
    }
  };

  // Get status border accent class
  const getStatusBorderClass = () => {
    if (!connected) return 'border-beacon';

    switch (status) {
      case 'online':
        return 'border-taxiway';
      case 'degraded':
        return 'border-beacon';
      case 'offline':
        return 'border-afterburner';
      default:
        return 'border-beacon';
    }
  };

  if (connected) {
    return (
      <div className="flex items-center gap-2">
        <Badge
          variant="outline"
          className={`${getStatusBorderClass()} bg-bg-secondary/50 text-taxiway font-medium gap-1.5 transition-all duration-300`}
        >
          <span className={animate ? getStatusDotClass() : 'w-2 h-2 rounded-full bg-current'} />
          Connected
        </Badge>
        {status && (
          <Badge
            variant="outline"
            className={`
              ${status === 'online' ? 'border-taxiway/30 text-taxiway' : ''}
              ${status === 'degraded' ? 'border-beacon/30 text-beacon' : ''}
              ${status === 'offline' ? 'border-afterburner/30 text-afterburner' : ''}
              transition-all duration-300
            `}
          >
            {status}
          </Badge>
        )}
      </div>
    );
  }

  return (
    <Badge
      variant="outline"
      className={`${getStatusBorderClass()} bg-bg-secondary/50 text-beacon font-medium gap-1.5 transition-all duration-300`}
    >
      <span className={animate ? getStatusDotClass() : 'w-2 h-2 rounded-full bg-current animate-pulse'} />
      Not Connected
    </Badge>
  );
}
