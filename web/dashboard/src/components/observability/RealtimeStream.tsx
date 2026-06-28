'use client';

import { useEffect, useState } from 'react';
import { Badge } from '@/components/ui/badge';
import { Wifi, WifiOff, Pause, Play } from 'lucide-react';
import { tokenVault } from '@/utils/token-vault';

interface RealtimeStreamProps {
  runId: string | null;
  onEvent?: (event: any) => void;
}

export default function RealtimeStream({ runId, onEvent }: RealtimeStreamProps) {
  const [connected, setConnected] = useState(false);
  const [paused, setPaused] = useState(false);
  const [eventCount, setEventCount] = useState(0);
  const [buffer, setBuffer] = useState<any[]>([]);

  useEffect(() => {
    if (!runId || paused) {
      setConnected(false);
      return;
    }

    let ws: WebSocket | null = null;

    (async () => {
      const API_BASE = import.meta.env.VITE_API_URL || '';
      const wsUrl = `${API_BASE.replace('http', 'ws')}/v1/agent-observability/runs/${runId}/stream`;
      await tokenVault.initialize();
      const token = await tokenVault.getAccessToken();

      ws = new WebSocket(wsUrl);

      ws.onopen = () => {
        setConnected(true);
        if (token) {
          ws.send(JSON.stringify({ headers: { Authorization: `Bearer ${token}` } }));
        }
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          setEventCount((prev) => prev + 1);
          setBuffer((prev) => [...prev.slice(-99), data]);
          onEvent?.(data);
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e);
        }
      };

      ws.onclose = () => {
        setConnected(false);
      };
    })();

    return () => {
      if (ws) ws.close();
    };
  }, [runId, paused, onEvent]);

  return (
    <div className="flex items-center gap-4">
      <div className="flex items-center gap-2">
        {connected ? (
          <Wifi className="h-4 w-4 text-green-500" />
        ) : (
          <WifiOff className="h-4 w-4 text-red-500" />
        )}
        <Badge variant={connected ? 'default' : 'destructive'}>
          {connected ? 'Live' : 'Disconnected'}
        </Badge>
      </div>

      <Badge variant="outline">
        {eventCount} events
      </Badge>

      <button
        onClick={() => setPaused(!paused)}
        className="p-1 rounded hover:bg-muted"
      >
        {paused ? (
          <Play className="h-4 w-4" />
        ) : (
          <Pause className="h-4 w-4" />
        )}
      </button>
    </div>
  );
}