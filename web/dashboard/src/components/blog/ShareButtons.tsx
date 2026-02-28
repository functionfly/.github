'use client';

import { useState, useCallback } from 'react';
import { Twitter, Linkedin, Link2, Check, Share2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip';

interface ShareButtonsProps {
  url?: string;
  title: string;
  description?: string;
  size?: 'sm' | 'md' | 'lg';
}

export function ShareButtons({
  url,
  title,
  description,
  size = 'md',
}: ShareButtonsProps) {
  const [copied, setCopied] = useState(false);
  
  // Get current URL if not provided
  const shareUrl = url || (typeof window !== 'undefined' ? window.location.href : '');
  const encodedUrl = encodeURIComponent(shareUrl);
  const encodedTitle = encodeURIComponent(title);
  const encodedDescription = encodeURIComponent(description || '');

  const handleCopyLink = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(shareUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy link:', err);
      // Fallback for older browsers
      const textArea = document.createElement('textarea');
      textArea.value = shareUrl;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  }, [shareUrl]);

  const twitterUrl = `https://twitter.com/intent/tweet?text=${encodedTitle}&url=${encodedUrl}`;
  const linkedInUrl = `https://www.linkedin.com/sharing/share-offsite/?url=${encodedUrl}`;

  const sizeClasses = {
    sm: 'h-8 w-8',
    md: 'h-10 w-10',
    lg: 'h-12 w-12',
  };

  const iconSizes = {
    sm: 'h-4 w-4',
    md: 'h-5 w-5',
    lg: 'h-6 w-6',
  };

  return (
    <TooltipProvider>
      <div className="flex items-center gap-2" role="group" aria-label="Share this article">
        <span className="text-sm text-muted-foreground mr-2">
          <Share2 className="h-4 w-4 inline mr-1" />
          Share
        </span>
        
        {/* Twitter/X */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              className={`${sizeClasses[size]} rounded-full border border-border/60 hover:bg-[#1DA1F2]/10 hover:border-[#1DA1F2]/30 hover:text-[#1DA1F2]`}
              asChild
            >
              <a
                href={twitterUrl}
                target="_blank"
                rel="noopener noreferrer"
                aria-label="Share on Twitter"
              >
                <Twitter className={iconSizes[size]} />
              </a>
            </Button>
          </TooltipTrigger>
          <TooltipContent>Share on Twitter</TooltipContent>
        </Tooltip>

        {/* LinkedIn */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              className={`${sizeClasses[size]} rounded-full border border-border/60 hover:bg-[#0A66C2]/10 hover:border-[#0A66C2]/30 hover:text-[#0A66C2]`}
              asChild
            >
              <a
                href={linkedInUrl}
                target="_blank"
                rel="noopener noreferrer"
                aria-label="Share on LinkedIn"
              >
                <Linkedin className={iconSizes[size]} />
              </a>
            </Button>
          </TooltipTrigger>
          <TooltipContent>Share on LinkedIn</TooltipContent>
        </Tooltip>

        {/* Copy Link */}
        <Tooltip>
          <TooltipTrigger asChild>
            <Button
              variant="outline"
              size="icon"
              className={`${sizeClasses[size]} rounded-full border border-border/60 hover:bg-muted/60`}
              onClick={handleCopyLink}
              aria-label={copied ? 'Link copied' : 'Copy link'}
            >
              {copied ? (
                <Check className={`${iconSizes[size]} text-green-500`} />
              ) : (
                <Link2 className={iconSizes[size]} />
              )}
            </Button>
          </TooltipTrigger>
          <TooltipContent>{copied ? 'Link copied!' : 'Copy link'}</TooltipContent>
        </Tooltip>
      </div>
    </TooltipProvider>
  );
}
