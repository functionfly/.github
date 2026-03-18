/**
 * AvatarPicker – change profile picture: upload or choose a default FunctionFly avatar.
 */

import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { cn } from '@/lib/utils';
import { Loader2, Upload, User } from 'lucide-react';
import { useRef } from 'react';
import { toast } from 'sonner';

const DEFAULT_AVATARS = [
  { id: 'default-1', url: '/avatars/default-1.svg', label: 'FunctionFly FF' },
  { id: 'default-2', url: '/avatars/default-2.svg', label: 'FunctionFly Bolt' },
] as const;

const MAX_FILE_SIZE_MB = 5;
const ACCEPT_IMAGES = 'image/jpeg,image/png,image/webp,image/gif';

export interface AvatarPickerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  currentAvatar?: string | null;
  onSelect: (avatarUrl: string) => void;
  isLoading?: boolean;
}

export function AvatarPicker({
  open,
  onOpenChange,
  currentAvatar,
  onSelect,
  isLoading = false,
}: AvatarPickerProps) {
  const inputRef = useRef<HTMLInputElement>(null);

  const handleUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.size > MAX_FILE_SIZE_MB * 1024 * 1024) {
      toast.error(`File is too large. Maximum size is ${MAX_FILE_SIZE_MB} MB.`);
      return;
    }
    const reader = new FileReader();
    reader.onload = () => {
      const dataUrl = reader.result as string;
      onSelect(dataUrl);
      onOpenChange(false);
    };
    reader.readAsDataURL(file);
    e.target.value = '';
  };

  const handleDefault = (url: string) => {
    onSelect(url);
    onOpenChange(false);
  };

  const handleClear = () => {
    onSelect('');
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Profile picture</DialogTitle>
          <DialogDescription>
            Upload a photo or choose a default FunctionFly avatar.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-6 py-2">
          <input
            ref={inputRef}
            type="file"
            accept={ACCEPT_IMAGES}
            className="hidden"
            onChange={handleUpload}
          />
          <Button
            type="button"
            variant="outline"
            className="w-full gap-2"
            onClick={() => inputRef.current?.click()}
            disabled={isLoading}
          >
            {isLoading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <Upload className="h-4 w-4" />
            )}
            Upload photo
          </Button>

          <div>
            <p className="text-sm font-medium text-text-secondary mb-2">Default avatars</p>
            <div className="flex gap-4">
              {DEFAULT_AVATARS.map(({ id, url, label }) => (
                <button
                  key={id}
                  type="button"
                  onClick={() => handleDefault(url)}
                  disabled={isLoading}
                  className={cn(
                    'rounded-full ring-2 ring-transparent transition focus:outline-none focus:ring-2 focus:ring-brand-500',
                    currentAvatar === url && 'ring-brand-500'
                  )}
                  title={label}
                >
                  <Avatar className="h-16 w-16 md:h-20 md:w-20">
                    <AvatarImage src={url} alt={label} />
                    <AvatarFallback className="bg-muted">
                      <User className="h-8 w-8" />
                    </AvatarFallback>
                  </Avatar>
                </button>
              ))}
            </div>
          </div>

          {currentAvatar && (
            <Button
              type="button"
              variant="ghost"
              className="w-full text-muted-foreground"
              onClick={handleClear}
              disabled={isLoading}
            >
              Remove photo
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
