import {
  GlobalNotificationCenter,
  UniversalSearchEngine,
} from '@/pages/StudioPage/components';
import { Sheet, SheetContent } from '@functionfly/ui-core';

interface SearchResult {
  id: string;
  type: "graph" | "node" | "plugin" | "setting" | "doc";
  title: string;
  description: string;
  path?: string;
  relevance: number;
  recent?: boolean;
}

interface StudioShellOverlaysProps {
  searchOpen: boolean;
  onSearchOpenChange: (open: boolean) => void;
  notificationsOpen: boolean;
  onNotificationsOpenChange: (open: boolean) => void;
  searchExtraResults?: SearchResult[];
  onSearchResultSelect?: (result: SearchResult) => void;
}

export function StudioShellOverlays({
  searchOpen,
  onSearchOpenChange,
  notificationsOpen,
  onNotificationsOpenChange,
  searchExtraResults = [],
  onSearchResultSelect,
}: StudioShellOverlaysProps) {
  return (
    <>
      <Sheet open={searchOpen} onOpenChange={onSearchOpenChange}>
        <SheetContent
          side="top"
          className="h-[85vh] max-h-[720px] w-full max-w-4xl mx-auto mt-4 rounded-xl border border-border-subtle bg-bg-primary p-0 overflow-hidden"
        >
          <UniversalSearchEngine
            onClose={() => onSearchOpenChange(false)}
            extraResults={searchExtraResults}
            onResultSelect={onSearchResultSelect}
          />
        </SheetContent>
      </Sheet>

      <Sheet open={notificationsOpen} onOpenChange={onNotificationsOpenChange}>
        <SheetContent
          side="right"
          className="w-full max-w-md sm:max-w-lg h-full flex flex-col bg-bg-primary p-0 overflow-hidden"
        >
          <GlobalNotificationCenter />
        </SheetContent>
      </Sheet>
    </>
  );
}
