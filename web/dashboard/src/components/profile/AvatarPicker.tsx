import { cn } from '@/lib/utils';
import { Loader2, Upload, User, X } from 'lucide-react';
import { useRef } from 'react';
import { toast } from 'sonner';

const DEFAULT_AVATARS = [
  { id: 'default-1', url: '/avatars/default-1.svg', label: 'FunctionFly FF' },
  { id: 'default-2', url: '/avatars/default-2.svg', label: 'FunctionFly Bolt' },
] as const;

const MAX_FILE_SIZE_MB = 5;
const ACCEPT_IMAGES = 'image/jpeg,image/png,image/webp,image/gif';
const AVATAR_MAX_DIM = 256;
const AVATAR_QUALITY = 0.7;

function compressImage(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error('Failed to read file'));
    reader.onload = () => {
      const img = new Image();
      img.onerror = () => reject(new Error('Failed to load image'));
      img.onload = () => {
        let { width, height } = img;
        if (width > AVATAR_MAX_DIM || height > AVATAR_MAX_DIM) {
          const ratio = Math.min(AVATAR_MAX_DIM / width, AVATAR_MAX_DIM / height);
          width = Math.round(width * ratio);
          height = Math.round(height * ratio);
        }
        const canvas = document.createElement('canvas');
        canvas.width = width;
        canvas.height = height;
        const ctx = canvas.getContext('2d');
        if (!ctx) { reject(new Error('Canvas not supported')); return; }
        ctx.drawImage(img, 0, 0, width, height);
        resolve(canvas.toDataURL('image/jpeg', AVATAR_QUALITY));
      };
      img.src = reader.result as string;
    };
    reader.readAsDataURL(file);
  });
}

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

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (file.size > MAX_FILE_SIZE_MB * 1024 * 1024) {
      toast.error(`File is too large. Maximum size is ${MAX_FILE_SIZE_MB} MB.`);
      return;
    }
    e.target.value = '';
    try {
      const dataUrl = await compressImage(file);
      await onSelect(dataUrl);
      onOpenChange(false);
    } catch {
      toast.error('Failed to upload profile picture');
    }
  };

  const handleDefault = (url: string) => {
    onSelect(url);
    onOpenChange(false);
  };

  const handleClear = () => {
    onSelect('');
    onOpenChange(false);
  };

  if (!open) return null;

  return (
    <div className="avatar-modal-overlay" onClick={() => onOpenChange(false)}>
      <div className="avatar-modal" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="avatar-modal__header">
          <div>
            <h2 className="avatar-modal__title">Profile picture</h2>
            <p className="avatar-modal__desc">Upload a photo or choose a default FunctionFly avatar.</p>
          </div>
          <button className="avatar-modal__close" onClick={() => onOpenChange(false)} aria-label="Close">
            <X className="avatar-modal__close-icon" />
          </button>
        </div>

        {/* Content */}
        <div className="avatar-modal__body">
          <input
            ref={inputRef}
            type="file"
            accept={ACCEPT_IMAGES}
            className="hidden"
            onChange={handleUpload}
          />

          {/* Upload button */}
          <button
            type="button"
            className="avatar-modal__upload-btn"
            onClick={() => inputRef.current?.click()}
            disabled={isLoading}
          >
            {isLoading ? (
              <Loader2 className="avatar-modal__upload-icon avatar-modal__upload-icon--spin" />
            ) : (
              <Upload className="avatar-modal__upload-icon" />
            )}
            Upload photo
          </button>

          {/* Default avatars */}
          <div className="avatar-modal__defaults">
            <p className="avatar-modal__defaults-label">Default avatars</p>
            <div className="avatar-modal__defaults-grid">
              {DEFAULT_AVATARS.map(({ id, url, label }) => (
                <button
                  key={id}
                  type="button"
                  onClick={() => handleDefault(url)}
                  disabled={isLoading}
                  className={cn(
                    'avatar-modal__avatar-btn',
                    currentAvatar === url && 'avatar-modal__avatar-btn--selected'
                  )}
                  title={label}
                >
                  <div className="avatar-modal__avatar-circle">
                    <img src={url} alt={label} className="avatar-modal__avatar-img" />
                    <div className="avatar-modal__avatar-fallback">
                      <User className="avatar-modal__avatar-fallback-icon" />
                    </div>
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Remove photo */}
          {currentAvatar && (
            <button
              type="button"
              className="avatar-modal__remove-btn"
              onClick={handleClear}
              disabled={isLoading}
            >
              Remove photo
            </button>
          )}
        </div>
      </div>
    </div>
  );
}
