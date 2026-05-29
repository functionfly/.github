import {
  ChevronLeft,
  Code2,
  FileCode2,
  GitBranch,
  History,
  Undo2,
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { createPlaygroundUrl } from '../../StandalonePlaygroundPage';
import type { GenerationHistoryItem } from '../useGenerationHistory';

interface GenerationHistorySidebarProps {
  history: GenerationHistoryItem[];
  onClose: () => void;
  onRevert: (item: GenerationHistoryItem) => void;
  onFork: (item: GenerationHistoryItem) => void;
  onCompare: (item: GenerationHistoryItem, olderItem: GenerationHistoryItem | null) => void;
}

export function GenerationHistorySidebar({
  history,
  onClose,
  onRevert,
  onFork,
  onCompare,
}: GenerationHistorySidebarProps) {
  return (
    <Card className="border-border/50 shadow-sm lg:order-first">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-base">
            <History className="h-4 w-4" />
            Generation History
          </CardTitle>
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={onClose}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent className="p-0">
        <ScrollArea className="h-[calc(100vh-300px)]">
          {history.length === 0 ? (
            <div className="p-4 text-center text-sm text-muted-foreground">
              <History className="h-8 w-8 mx-auto mb-2 opacity-50" />
              <p>No generations yet.</p>
              <p className="text-xs">Generated functions will appear here.</p>
            </div>
          ) : (
            <div className="space-y-2 p-4">
              {history.map((item, index) => (
                <div
                  key={item.id}
                  className="group relative bg-muted/50 rounded-lg p-3 hover:bg-muted transition-colors"
                >
                  <div className="flex items-start justify-between gap-2">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <Badge variant="outline" className="text-[10px] shrink-0">
                          v{history.length - index}
                        </Badge>
                        <Badge variant="secondary" className="text-[10px] shrink-0">
                          {item.runtime}
                        </Badge>
                      </div>
                      <p className="text-xs text-muted-foreground line-clamp-2 mb-1">
                        {item.description}
                      </p>
                      <p className="text-[10px] text-muted-foreground/60">
                        {new Date(item.timestamp).toLocaleString(undefined, {
                          month: 'short',
                          day: 'numeric',
                          hour: '2-digit',
                          minute: '2-digit',
                        })}
                      </p>
                    </div>
                  </div>

                  <div className="flex items-center gap-1 mt-2 pt-2 border-t border-border/50">
                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => onRevert(item)}
                          >
                            <Undo2 className="h-3.5 w-3.5" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Revert to this version</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>

                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => onFork(item)}
                          >
                            <GitBranch className="h-3.5 w-3.5" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Fork from this version</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>

                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => {
                              const playgroundUrl = createPlaygroundUrl(
                                item.result.code,
                                item.runtime
                              );
                              window.open(playgroundUrl, '_blank');
                            }}
                          >
                            <FileCode2 className="h-3.5 w-3.5" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Open in Playground</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>

                    <TooltipProvider>
                      <Tooltip>
                        <TooltipTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => {
                              const currentIndex = history.findIndex((h) => h.id === item.id);
                              onCompare(item, history[currentIndex + 1] || null);
                            }}
                          >
                            <Code2 className="h-3.5 w-3.5" />
                          </Button>
                        </TooltipTrigger>
                        <TooltipContent>
                          <p>Compare with previous</p>
                        </TooltipContent>
                      </Tooltip>
                    </TooltipProvider>
                  </div>
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
