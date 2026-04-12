/**
 * CanvasControls Component
 * Floating control panel for the canvas with zoom, fit view, grid snap, etc.
 */

import { useState, useCallback } from 'react';
import { useReactFlow, useViewport } from '@xyflow/react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Plus,
  Minus,
  Maximize,
  Grid3X3,
  Presentation,
  Map,
  Settings,
  ChevronRight,
  ChevronLeft,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';
import { cn } from '@/lib/utils';

interface CanvasControlsProps {
  zoom?: number;
  onZoomIn?: () => void;
  onZoomOut?: () => void;
  onFitView?: () => void;
  gridSnap?: boolean;
  onToggleGridSnap?: () => void;
  presentationMode?: boolean;
  onTogglePresentationMode?: () => void;
  onPresentationModeToggle?: () => void;  // Alternative prop name
  miniMapVisible?: boolean;
  onToggleMiniMap?: () => void;
  className?: string;
}

export function CanvasControls({
  zoom: zoomProp,
  onZoomIn: onZoomInProp,
  onZoomOut: onZoomOutProp,
  onFitView: onFitViewProp,
  gridSnap: gridSnapProp,
  onToggleGridSnap,
  presentationMode = false,
  onTogglePresentationMode,
  onPresentationModeToggle,
  miniMapVisible: miniMapVisibleProp,
  onToggleMiniMap,
  className,
}: CanvasControlsProps) {
  const { zoomIn, zoomOut, fitView } = useReactFlow();
  const { zoom } = useViewport();
  const [isExpanded, setIsExpanded] = useState(true);
  const [gridSnap, setGridSnap] = useState(gridSnapProp ?? false);
  const [miniMapVisible, setMiniMapVisible] = useState(miniMapVisibleProp ?? true);
  
  const zoomPercent = Math.round((zoomProp ?? zoom) * 100);
  
  const onZoomIn = useCallback(() => {
    onZoomInProp ? onZoomInProp() : zoomIn();
  }, [onZoomInProp, zoomIn]);
  
  const onZoomOut = useCallback(() => {
    onZoomOutProp ? onZoomOutProp() : zoomOut();
  }, [onZoomOutProp, zoomOut]);
  
  const onFitView = useCallback(() => {
    onFitViewProp ? onFitViewProp() : fitView({ padding: 0.2 });
  }, [onFitViewProp, fitView]);
  
  const handleToggleGridSnap = useCallback(() => {
    if (onToggleGridSnap) {
      onToggleGridSnap();
    } else {
      setGridSnap(prev => !prev);
    }
  }, [onToggleGridSnap]);
  
  const handleToggleMiniMap = useCallback(() => {
    if (onToggleMiniMap) {
      onToggleMiniMap();
    } else {
      setMiniMapVisible(prev => !prev);
    }
  }, [onToggleMiniMap]);
  
  const handleTogglePresentationMode = useCallback(() => {
    if (onTogglePresentationMode) {
      onTogglePresentationMode();
    } else if (onPresentationModeToggle) {
      onPresentationModeToggle();
    }
  }, [onTogglePresentationMode, onPresentationModeToggle]);

  return (
    <TooltipProvider>
      <motion.div
        initial={{ opacity: 0, x: 20 }}
        animate={{ opacity: 1, x: 0 }}
        className={cn(
          "absolute top-4 right-4 z-10 flex flex-col items-end gap-2",
          className
        )}
      >
        {/* Collapse Toggle */}
        <Button
          variant="ghost"
          size="icon"
          className="h-6 w-6 opacity-50 hover:opacity-100"
          onClick={() => setIsExpanded(!isExpanded)}
        >
          {isExpanded ? (
            <ChevronRight className="w-4 h-4" />
          ) : (
            <ChevronLeft className="w-4 h-4" />
          )}
        </Button>

        <AnimatePresence>
          {isExpanded && (
            <motion.div
              initial={{ opacity: 0, scale: 0.9, x: 10 }}
              animate={{ opacity: 1, scale: 1, x: 0 }}
              exit={{ opacity: 0, scale: 0.9, x: 10 }}
              className="bg-[var(--bg-secondary)]/90 backdrop-blur-md border border-[var(--border-subtle)] rounded-lg shadow-xl overflow-hidden"
            >
              {/* Zoom Controls */}
              <div className="flex items-center p-1.5 border-b border-[var(--border-subtle)]">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={onZoomOut}
                    >
                      <Minus className="w-4 h-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="left">Zoom out</TooltipContent>
                </Tooltip>

                <div className="w-14 text-center">
                  <span className="text-xs font-medium text-[var(--text-primary)]">
                    {zoomPercent}%
                  </span>
                </div>

                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={onZoomIn}
                    >
                      <Plus className="w-4 h-4" />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="left">Zoom in</TooltipContent>
                </Tooltip>
              </div>

              {/* Action Buttons */}
              <div className="flex flex-col p-1 gap-0.5">
                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-8 justify-start gap-2 text-xs"
                      onClick={onFitView}
                    >
                      <Maximize className="w-4 h-4" />
                      <span className="flex-1">Fit View</span>
                      <kbd className="hidden sm:inline-block px-1.5 py-0.5 text-[10px] bg-[var(--bg-tertiary)] rounded text-[var(--text-muted)]">
                        F
                      </kbd>
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="left">Fit all nodes in view</TooltipContent>
                </Tooltip>

                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant={miniMapVisible ? 'secondary' : 'ghost'}
                      size="sm"
                      className="h-8 justify-start gap-2 text-xs"
                      onClick={onToggleMiniMap}
                    >
                      <Map className="w-4 h-4" />
                      <span className="flex-1">Mini-map</span>
                      <Switch
                        checked={miniMapVisible}
                        className="scale-75"
                        onClick={(e) => e.stopPropagation()}
                      />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="left">Toggle mini-map visibility</TooltipContent>
                </Tooltip>

                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant={gridSnap ? 'secondary' : 'ghost'}
                      size="sm"
                      className="h-8 justify-start gap-2 text-xs"
                      onClick={onToggleGridSnap}
                    >
                      <Grid3X3 className="w-4 h-4" />
                      <span className="flex-1">Grid Snap</span>
                      <Switch
                        checked={gridSnap}
                        className="scale-75"
                        onClick={(e) => e.stopPropagation()}
                      />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="left">Snap nodes to grid</TooltipContent>
                </Tooltip>

                <Tooltip>
                  <TooltipTrigger asChild>
                    <Button
                      variant={presentationMode ? 'secondary' : 'ghost'}
                      size="sm"
                      className="h-8 justify-start gap-2 text-xs"
                      onClick={handleTogglePresentationMode}
                    >
                      <Presentation className="w-4 h-4" />
                      <span className="flex-1">Present</span>
                      <Switch
                        checked={presentationMode}
                        className="scale-75"
                        onClick={(e) => e.stopPropagation()}
                      />
                    </Button>
                  </TooltipTrigger>
                  <TooltipContent side="left">Presentation mode (hides UI)</TooltipContent>
                </Tooltip>
              </div>
            </motion.div>
          )}
        </AnimatePresence>

        {/* Collapsed State - Just zoom percentage */}
        {!isExpanded && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="bg-[var(--bg-secondary)]/90 backdrop-blur-md border border-[var(--border-subtle)] rounded-lg px-2 py-1"
          >
            <span className="text-xs font-medium text-[var(--text-primary)]">
              {zoomPercent}%
            </span>
          </motion.div>
        )}
      </motion.div>
    </TooltipProvider>
  );
}

export default CanvasControls;
