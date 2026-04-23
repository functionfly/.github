import { type MessageAttachment, conversationsApi } from '@/api/conversations';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import {
  File,
  FileAudio,
  FileCode,
  FileImage,
  FileText,
  FileVideo,
  Trash2,
} from 'lucide-react';
import { toast } from 'sonner';

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function getIcon(contentType: string) {
  if (contentType.startsWith('image/')) return FileImage;
  if (contentType.startsWith('video/')) return FileVideo;
  if (contentType.startsWith('audio/')) return FileAudio;
  if (
    contentType.startsWith('text/') ||
    contentType.includes('json') ||
    contentType.includes('xml') ||
    contentType.includes('javascript') ||
    contentType.includes('typescript')
  )
    return FileCode;
  if (contentType.includes('pdf') || contentType.includes('document')) return FileText;
  return File;
}

export interface MessageAttachmentsProps {
  attachments: MessageAttachment[];
  username: string;
  conversationId: string;
  messageId: string;
  currentUserId?: string;
  className?: string;
}

export function MessageAttachments({
  attachments,
  username,
  conversationId,
  messageId,
  currentUserId,
  className,
}: MessageAttachmentsProps) {
  const queryClient = useQueryClient();

  const deleteMutation = useMutation({
    mutationFn: (attachmentId: string) =>
      conversationsApi.deleteAttachment(username, conversationId, messageId, attachmentId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['conversation-messages', username, conversationId] });
      toast.success('Attachment removed');
    },
    onError: (err: Error) => toast.error(err.message || 'Failed to remove attachment'),
  });

  if (!attachments || attachments.length === 0) return null;

  return (
    <div className={cn('flex flex-wrap gap-2 mt-1.5', className)}>
      {attachments.map((att) => {
        const Icon = getIcon(att.content_type);
        const isOwner = att.uploaded_by === currentUserId;

        return (
          <div
            key={att.id}
            className={cn(
              'flex items-center gap-2 rounded-md border border-border/60 bg-muted/40 px-2.5 py-1.5 text-xs',
              'hover:bg-muted/60 transition-colors group'
            )}
          >
            {att.thumbnail_url ? (
              <img
                src={att.thumbnail_url}
                alt={att.filename}
                className="h-8 w-8 rounded object-cover"
              />
            ) : (
              <Icon className="h-4 w-4 text-muted-foreground shrink-0" />
            )}
            <div className="flex flex-col min-w-0">
              <a
                href={att.storage_url}
                target="_blank"
                rel="noopener noreferrer"
                className="font-medium truncate hover:underline"
              >
                {att.filename}
              </a>
              <span className="text-muted-foreground">{formatSize(att.size_bytes)}</span>
            </div>
            {isOwner && (
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6 opacity-0 group-hover:opacity-100 transition-opacity shrink-0"
                onClick={() => deleteMutation.mutate(att.id)}
                disabled={deleteMutation.isPending}
              >
                <Trash2 className="h-3 w-3 text-destructive" />
              </Button>
            )}
          </div>
        );
      })}
    </div>
  );
}
