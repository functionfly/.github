'use client';

import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Play, Pause, SkipForward } from 'lucide-react';

interface ReplayControlsProps {
  speed: number;
  onSpeedChange: (speed: number) => void;
  isPlaying: boolean;
  onPlayPause: () => void;
}

export default function ReplayControls({
  speed,
  onSpeedChange,
  isPlaying,
  onPlayPause,
}: ReplayControlsProps) {
  return (
    <div className="flex items-center gap-2">
      <Button variant="outline" size="sm" onClick={onPlayPause}>
        {isPlaying ? (
          <Pause className="h-4 w-4" />
        ) : (
          <Play className="h-4 w-4" />
        )}
      </Button>

      <Select value={String(speed)} onValueChange={(v) => onSpeedChange(Number(v))}>
        <SelectTrigger className="w-[80px]">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="0.5">0.5x</SelectItem>
          <SelectItem value="1">1x</SelectItem>
          <SelectItem value="2">2x</SelectItem>
          <SelectItem value="4">4x</SelectItem>
        </SelectContent>
      </Select>
    </div>
  );
}
