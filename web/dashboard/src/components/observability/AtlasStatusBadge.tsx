'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Wifi, WifiOff } from 'lucide-react';

interface AtlasStatusBadgeProps {
  connected: boolean;
}

export default function AtlasStatusBadge({ connected }: AtlasStatusBadgeProps) {
  const [lastSync, setLastSync] = useState<Date | null>(null);

  useEffect(() => {
    if (connected) {
      setLastSync(new Date());
    }
  }, [connected]);

  return (
    <Badge
      variant={connected ? 'default' : 'destructive'}
      className="flex items-center gap-1"
    >
      {connected ? (
        <>
          <Wifi className="h-3 w-3" />
          Connected
          {lastSync && (
            <span className="text-xs opacity-70">
              {lastSync.toLocaleTimeString()}
            </span>
          )}
        </>
      ) : (
        <>
          <WifiOff className="h-3 w-3" />
          Disconnected
        </>
      )}
    </Badge>
  );
}