'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Wifi, WifiOff, Loader2, CloudOff } from 'lucide-react';
import { useAtlasHealth } from '@/hooks/useAtlasTraces';

interface AtlasStatusBadgeProps {
  connected: boolean;
}

export default function AtlasStatusBadge({ connected }: AtlasStatusBadgeProps) {
  const [lastSync, setLastSync] = useState<Date | null>(null);
  const { data: health, isLoading } = useAtlasHealth();

  useEffect(() => {
    if (connected) {
      setLastSync(new Date());
    }
  }, [connected]);

  if (isLoading) {
    return (
      <Badge variant="secondary" className="flex items-center gap-1">
        <Loader2 className="h-3 w-3 animate-spin" />
        Checking...
      </Badge>
    );
  }

  const atlasReachable = health?.status === 'ok';

  if (atlasReachable) {
    return (
      <Badge variant="default" className="flex items-center gap-1 bg-emerald-600 hover:bg-emerald-700">
        <Wifi className="h-3 w-3" />
        {connected ? 'Connected' : 'Atlas Ready'}
        {connected && lastSync && (
          <span className="text-xs opacity-70">
            {lastSync.toLocaleTimeString()}
          </span>
        )}
      </Badge>
    );
  }

  if (health?.status === 'disabled') {
    return (
      <Badge variant="outline" className="flex items-center gap-1 opacity-60">
        <CloudOff className="h-3 w-3" />
        Atlas Off
      </Badge>
    );
  }

  return (
    <Badge variant="destructive" className="flex items-center gap-1">
      <WifiOff className="h-3 w-3" />
      Disconnected
    </Badge>
  );
}
