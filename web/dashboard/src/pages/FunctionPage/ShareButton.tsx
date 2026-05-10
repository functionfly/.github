import { Button } from '@/components/ui/button';
import { Loader2, Share2 } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';
import type { FunctionInfo } from './types';

export function ShareButton({ functionInfo }: { functionInfo: FunctionInfo }) {
  const [isSharing, setIsSharing] = useState(false);

  const handleShare = async () => {
    setIsSharing(true);
    const shareUrl = window.location.href;
    const shareData = {
      title: `${functionInfo.author}/${functionInfo.name}`,
      text: functionInfo.description || `Check out ${functionInfo.name} on FunctionFly`,
      url: shareUrl,
    };

    try {
      if (navigator.share) {
        await navigator.share(shareData);
        toast.success('Shared successfully');
      } else {
        await navigator.clipboard.writeText(shareUrl);
        toast.success('Link copied to clipboard');
      }
    } catch (err) {
      if ((err as Error).name !== 'AbortError') {
        toast.error('Failed to share');
      }
    } finally {
      setIsSharing(false);
    }
  };

  return (
    <Button
      variant="outline"
      size="lg"
      className="gap-2"
      onClick={handleShare}
      disabled={isSharing}
    >
      {isSharing ? <Loader2 className="w-4 h-4 animate-spin" /> : <Share2 className="w-4 h-4" />}
      Share
    </Button>
  );
}
