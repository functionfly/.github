'use client';

import { Button } from '@/components/ui/button';
import { WifiOff, RefreshCw } from 'lucide-react';

interface ReconnectButtonProps {
  onReconnect: () => void;
  loading?: boolean;
}

export default function ReconnectButton({ onReconnect, loading }: ReconnectButtonProps) {
  return (
    <Button variant="outline" size="sm" onClick={onReconnect} disabled={loading} className="gap-2">
      {loading ? (
        <>
          <RefreshCw className="h-4 w-4 animate-spin" />
          Reconnecting...
        </>
      ) : (
        <>
          <WifiOff className="h-4 w-4" />
          Reconnect
        </>
      )}
    </Button>
  );
}
