import { Button } from '@/components/ui/button';
import { Card, CardContent } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Textarea } from '@/components/ui/textarea';
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { publicAppUrl } from '@/lib/constants';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import { Icon } from '@iconify/react';
import {
  AlertCircle,
  ArrowLeft,
  Box,
  CheckCircle2,
  ChevronRight,
  Clock,
  Cloud,
  Code2,
  Globe,
  Info,
  Key,
  Layers,
  Loader2,
  Lock,
  Rocket,
  Save,
  Smile,
  Sparkles,
  Tag,
  X,
  Zap,
} from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { ICON_COLORS, ICON_OPTIONS, useCreateApp, type Environment } from './hooks/useCreateApp';

type IconPickerTab = 'emoji' | 'icons';

// Animation keyframes as style element
const animationStyles = `
@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes slideInRight {
  from {
    opacity: 0;
    transform: translateX(30px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

@keyframes pulse-soft {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

.animate-fade-in-up {
  animation: fadeInUp 0.5s ease-out forwards;
}

.animate-slide-in-right {
  animation: slideInRight 0.5s ease-out forwards;
}

.animate-pulse-soft {
  animation: pulse-soft 2s ease-in-out infinite;
}

.animation-delay-100 {
  animation-delay: 100ms;
}

.animation-delay-200 {
  animation-delay: 200ms;
}

.animation-delay-300 {
  animation-delay: 300ms;
}

.animation-delay-400 {
  animation-delay: 400ms;
}
`;

// Icon picker panel (used inside Radix Popover so it is not clipped by parent overflow-hidden)
function IconPickerPanel({
  value,
  onSelect,
}: {
  value: string;
  onSelect: (value: string) => void;
}) {
  const [activeTab, setActiveTab] = useState<IconPickerTab>('emoji');

  return (
    <div className="min-w-[280px]">
      {/* Tabs */}
      <div className="flex items-center gap-1 p-1 rounded-lg bg-muted/50 mb-3">
        <button
          type="button"
          onClick={() => setActiveTab('emoji')}
          className={cn(
            'flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all flex-1',
            activeTab === 'emoji'
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <Smile className="w-4 h-4" />
          Emoji
        </button>
        <button
          type="button"
          onClick={() => setActiveTab('icons')}
          className={cn(
            'flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm font-medium transition-all flex-1',
            activeTab === 'icons'
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          )}
        >
          <Icon icon="fluent:icons-24-filled" className="w-4 h-4" />
          Icons
        </button>
      </div>

      {/* Content */}
      {activeTab === 'emoji' ? (
        <div className="grid grid-cols-5 gap-2">
          {ICON_OPTIONS.map((option) => (
            <button
              key={option.emoji}
              type="button"
              onClick={() => onSelect(option.emoji)}
              className={cn(
                'w-10 h-10 rounded-lg text-2xl flex items-center justify-center transition-all',
                value === option.emoji
                  ? 'bg-primary/20 ring-2 ring-primary scale-105'
                  : 'hover:bg-muted hover:scale-105'
              )}
              title={option.label}
            >
              {option.emoji}
            </button>
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-4 gap-2">
          {ICON_OPTIONS.map((option) => (
            <button
              key={option.icon}
              type="button"
              onClick={() => onSelect(option.emoji)}
              className={cn(
                'w-10 h-10 rounded-lg flex items-center justify-center transition-all text-primary',
                value === option.emoji
                  ? 'bg-primary/20 ring-2 ring-primary scale-105'
                  : 'hover:bg-muted hover:scale-105'
              )}
              title={option.label}
            >
              <Icon icon={option.icon} className="w-6 h-6" />
            </button>
          ))}
        </div>
      )}

      <p className="text-xs text-muted-foreground mt-3 text-center">
        Click to select • Currently:{' '}
        {ICON_OPTIONS.find((o) => o.emoji === value)?.label || 'Custom'}
      </p>
    </div>
  );
}

function ValidationMessage({
  message,
  type,
}: {
  message: string;
  type: 'error' | 'success' | 'info';
}) {
  const styles = {
    error: 'text-destructive',
    success: 'text-emerald-500',
    info: 'text-muted-foreground',
  };
  const icons = {
    error: AlertCircle,
    success: CheckCircle2,
    info: Info,
  };
  const Icon = icons[type];

  return (
    <div className={cn('flex items-center gap-1.5 text-xs mt-1.5', styles[type])}>
      <Icon className="w-3 h-3 flex-shrink-0" />
      {message}
    </div>
  );
}

function SectionCard({
  children,
  className,
  delay = 0,
}: {
  children: React.ReactNode;
  className?: string;
  delay?: number;
}) {
  const delayClass =
    delay === 100
      ? 'animation-delay-100'
      : delay === 200
        ? 'animation-delay-200'
        : delay === 300
          ? 'animation-delay-300'
          : delay === 400
            ? 'animation-delay-400'
            : '';

  return (
    <div
      className={cn(
        'rounded-2xl border border-border/50 bg-card overflow-hidden animate-fade-in-up opacity-0',
        delayClass,
        className
      )}
      style={{ animationFillMode: 'forwards' }}
    >
      {children}
    </div>
  );
}

function TagInput({
  tags,
  onAdd,
  onRemove,
  disabled,
}: {
  tags: string[];
  onAdd: (tag: string) => void;
  onRemove: (tag: string) => void;
  disabled?: boolean;
}) {
  const [input, setInput] = useState('');
  const inputRef = useRef<HTMLInputElement>(null);

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault();
      if (input.trim()) {
        onAdd(input);
        setInput('');
      }
    }
    if (e.key === 'Backspace' && !input && tags.length > 0) {
      onRemove(tags[tags.length - 1]);
    }
  };

  return (
    <div
      className={cn(
        'flex flex-wrap items-center gap-2 p-2 rounded-lg border bg-secondary/50 min-h-[44px]',
        'focus-within:ring-2 focus-within:ring-ring focus-within:border-input',
        disabled && 'opacity-50 cursor-not-allowed'
      )}
      onClick={() => inputRef.current?.focus()}
    >
      {tags.map((tag) => (
        <span
          key={tag}
          className="inline-flex items-center gap-1 px-2 py-1 rounded-md bg-primary/10 text-primary text-sm"
        >
          {tag}
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onRemove(tag);
            }}
            className="hover:text-destructive transition-colors"
            disabled={disabled}
          >
            <X className="w-3 h-3" />
          </button>
        </span>
      ))}
      <input
        ref={inputRef}
        type="text"
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyDown={handleKeyDown}
        placeholder={tags.length === 0 ? 'Add tags (press Enter)' : ''}
        className="flex-1 min-w-[100px] bg-transparent outline-none text-sm placeholder:text-muted-foreground"
        disabled={disabled}
      />
    </div>
  );
}

function AppPreviewCard({ formData }: { formData: ReturnType<typeof useCreateApp>['formData'] }) {
  const selectedColor = ICON_COLORS.find((c) => c.value === formData.iconColor) || ICON_COLORS[0];

  return (
    <Card className="overflow-hidden border border-border/50 bg-card dark:bg-card/50 backdrop-blur-sm animate-slide-in-right opacity-0 sticky top-20 shadow-sm">
      <CardContent className="p-0">
        {/* Preview Header */}
        <div className="px-4 py-3 border-b border-border/50 bg-muted/50 dark:bg-muted/30">
          <p className="text-xs font-medium text-muted-foreground uppercase tracking-wider">
            Live Preview
          </p>
        </div>

        {/* Preview Content */}
        <div className="p-6 space-y-6">
          {/* App Card */}
          <div className="relative">
            <div
              className="w-full rounded-xl border border-border/50 p-5 relative overflow-hidden"
              style={{
                background: `linear-gradient(135deg, ${selectedColor.from} 0%, ${selectedColor.to} 100%)`,
              }}
            >
              <div className="absolute inset-0 rounded-xl bg-gradient-to-br from-white/10 to-transparent pointer-events-none" />
              <div className="relative z-10">
                <div className="flex items-start justify-between">
                  <div
                    className={cn(
                      'w-14 h-14 rounded-xl bg-white/20 backdrop-blur-sm flex items-center justify-center text-3xl shadow-lg',
                      'border border-white/30'
                    )}
                  >
                    {formData.iconEmoji}
                  </div>
                  <span
                    className={cn(
                      'px-2.5 py-1 rounded-full text-xs font-medium backdrop-blur-sm text-white',
                      'border border-white/30 shadow-sm',
                      formData.visibility === 'public' ? 'bg-emerald-500/40' : 'bg-amber-500/40'
                    )}
                  >
                    {formData.visibility === 'public' ? 'Public' : 'Private'}
                  </span>
                </div>
                <div className="mt-4">
                  <h3 className="text-lg font-semibold text-white drop-shadow-md">
                    {formData.name || 'Untitled App'}
                  </h3>
                  <p className="text-sm text-white/90 font-mono mt-0.5 drop-shadow-sm">
                    {formData.slug || 'app-slug'}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* Details */}
          <div className="space-y-3">
            {formData.description && (
              <p className="text-sm text-muted-foreground line-clamp-3">{formData.description}</p>
            )}

            {formData.tags.length > 0 && (
              <div className="flex flex-wrap gap-1.5">
                {formData.tags.slice(0, 3).map((tag) => (
                  <span
                    key={tag}
                    className="px-2 py-0.5 rounded-full bg-secondary text-secondary-foreground text-xs"
                  >
                    {tag}
                  </span>
                ))}
                {formData.tags.length > 3 && (
                  <span className="px-2 py-0.5 rounded-full bg-secondary text-secondary-foreground text-xs">
                    +{formData.tags.length - 3}
                  </span>
                )}
              </div>
            )}

            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <Layers className="w-3.5 h-3.5" />
              <span className="capitalize">{formData.environment}</span>
            </div>
          </div>

          {/* URL Preview */}
          {formData.slug && (
            <div className="pt-4 border-t border-border/50">
              <p className="text-xs text-muted-foreground mb-1.5">App URL</p>
              <code className="block text-xs font-mono text-primary break-all bg-muted/50 dark:bg-muted/30 px-3 py-2 rounded-lg">
                {publicAppUrl(formData.slug)}
              </code>
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  );
}

function GettingStartedTip({
  icon: Icon,
  title,
  description,
  delay,
}: {
  icon: React.ElementType;
  title: string;
  description: string;
  delay: number;
}) {
  const delayClass =
    delay === 100
      ? 'animation-delay-100'
      : delay === 200
        ? 'animation-delay-200'
        : delay === 300
          ? 'animation-delay-300'
          : '';

  return (
    <div
      className={cn(
        'group p-4 rounded-xl border border-border bg-card hover:bg-accent/50 transition-all duration-300',
        'hover:border-border hover:shadow-md hover:-translate-y-1',
        'animate-fade-in-up opacity-0',
        delayClass
      )}
      style={{ animationFillMode: 'forwards' }}
    >
      <div className="flex items-start gap-4">
        <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center group-hover:bg-primary/20 transition-colors">
          <Icon className="w-5 h-5 text-primary" />
        </div>
        <div className="flex-1">
          <h4 className="font-medium text-sm">{title}</h4>
          <p className="text-xs text-muted-foreground mt-1">{description}</p>
        </div>
      </div>
    </div>
  );
}

export function CreateAppPage() {
  const navigate = useNavigate();
  const user = useAuthStore((state) => state.user);
  const [iconPickerOpen, setIconPickerOpen] = useState(false);
  const [isGeneratingDescription, setIsGeneratingDescription] = useState(false);

  const {
    formData,
    errors,
    isDirty,
    isSubmitting,
    isValid,
    lastSaved,
    setName,
    setSlug,
    setDescription,
    setIconColor,
    setIconEmoji,
    setVisibility,
    setEnvironment,
    addTag,
    removeTag,
    handleSubmit,
    handleCancel,
    slugPreview,
    nameValidation,
    slugValidation,
    descriptionValidation,
  } = useCreateApp();

  // AI Generate Description
  const generateAIDescription = async () => {
    if (!formData.name.trim()) return;

    setIsGeneratingDescription(true);
    try {
      // Simulate AI generation - replace with actual API call
      await new Promise((resolve) => setTimeout(resolve, 1500));

      const descriptions = [
        `${formData.name} is a powerful application designed to streamline your workflow and boost productivity.`,
        `A cutting-edge solution for modern teams. ${formData.name} helps you manage resources efficiently and scale your operations.`,
        `${formData.name} provides an intuitive platform for organizing, deploying, and monitoring your applications with ease.`,
        `Transform your development process with ${formData.name}. Built for speed, reliability, and seamless collaboration.`,
      ];

      const randomDescription = descriptions[Math.floor(Math.random() * descriptions.length)];
      setDescription(randomDescription);
    } catch (error) {
      console.error('Failed to generate description:', error);
    } finally {
      setIsGeneratingDescription(false);
    }
  };

  // Keyboard shortcuts
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') {
        handleCancel();
      }
      if ((e.ctrlKey || e.metaKey) && e.key === 'Enter' && isValid && !isSubmitting) {
        e.preventDefault();
        handleSubmit();
      }
    }
    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [handleCancel, handleSubmit, isValid, isSubmitting]);

  const formatRelativeTime = (date: Date): string => {
    const seconds = Math.floor((Date.now() - date.getTime()) / 1000);
    if (seconds < 5) return 'just now';
    if (seconds < 60) return `${seconds}s ago`;
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return `${minutes}m ago`;
    return `${Math.floor(minutes / 60)}h ago`;
  };

  return (
    <TooltipProvider>
      <style>{animationStyles}</style>

      <div className="min-h-screen bg-background">
        {/* Sticky Action Bar */}
        <div className="sticky top-0 z-50 border-b border-border/50 bg-background/85 backdrop-blur-xl">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between gap-4">
            {/* Left: Breadcrumb */}
            <div className="flex items-center gap-3 min-w-0">
              <Button
                variant="ghost"
                size="icon"
                onClick={handleCancel}
                className="shrink-0 text-muted-foreground hover:text-foreground"
                aria-label="Back to apps"
              >
                <ArrowLeft className="w-4 h-4" />
              </Button>
              <nav className="flex items-center gap-1.5 text-sm min-w-0" aria-label="Breadcrumb">
                <Link
                  to="/apps"
                  className="text-muted-foreground hover:text-foreground transition-colors truncate"
                >
                  Apps
                </Link>
                <ChevronRight className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                <span className="text-foreground font-medium truncate">Create New App</span>
              </nav>
            </div>

            {/* Center: Auto-save indicator */}
            <div className="hidden sm:flex items-center gap-2 text-xs text-muted-foreground">
              {isSubmitting ? (
                <span className="flex items-center gap-1.5">
                  <Loader2 className="w-3.5 h-3.5 animate-spin text-primary" />
                  Creating app...
                </span>
              ) : isDirty ? (
                lastSaved ? (
                  <span className="flex items-center gap-1.5">
                    <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                    Draft saved {formatRelativeTime(lastSaved)}
                  </span>
                ) : (
                  <span className="flex items-center gap-1.5">
                    <Save className="w-3.5 h-3.5 text-amber-400" />
                    Unsaved changes
                  </span>
                )
              ) : (
                <span className="flex items-center gap-1.5">
                  <Clock className="w-3.5 h-3.5" />
                  Ready to create
                </span>
              )}
            </div>

            {/* Right: Actions */}
            <div className="flex items-center gap-2 shrink-0">
              <Button
                variant="ghost"
                size="sm"
                onClick={handleCancel}
                className="text-muted-foreground hover:text-foreground hidden sm:flex"
              >
                Cancel
              </Button>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    size="sm"
                    onClick={handleSubmit}
                    disabled={!isValid || isSubmitting}
                    className={cn(
                      'gap-2 font-semibold min-w-[120px]',
                      isValid && 'gradient-primary text-primary-foreground border-0'
                    )}
                  >
                    {isSubmitting ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        Creating...
                      </>
                    ) : (
                      <>
                        <Sparkles className="w-4 h-4" />
                        Create App
                      </>
                    )}
                  </Button>
                </TooltipTrigger>
                {!isValid && !isSubmitting && (
                  <TooltipContent>
                    <p>Please fill in required fields</p>
                  </TooltipContent>
                )}
              </Tooltip>
            </div>
          </div>
        </div>

        {/* Main Content */}
        <div className="max-w-7xl mx-auto px-4 sm:px-6 py-8">
          {/* Page Header */}
          <div
            className="mb-8 animate-fade-in-up opacity-0"
            style={{ animationFillMode: 'forwards' }}
          >
            <div className="flex items-start gap-4">
              <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-indigo-500/20 to-purple-500/20 border border-indigo-500/30 flex items-center justify-center animate-pulse-soft">
                <Rocket className="w-7 h-7 text-indigo-400" />
              </div>
              <div>
                <h1 className="text-3xl font-bold tracking-tight">Create New App</h1>
                <p className="text-muted-foreground mt-1.5 text-base">
                  Apps organize your functions into deployable units
                </p>
              </div>
            </div>
          </div>

          {/* Two Column Layout */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Left Column: Form */}
            <div className="lg:col-span-2 space-y-6">
              {/* General Error */}
              {errors.general && (
                <div className="flex items-start gap-2.5 p-4 text-sm text-destructive bg-destructive/10 border border-destructive/20 rounded-xl animate-fade-in-up">
                  <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
                  <span>{errors.general}</span>
                </div>
              )}

              {/* App Identity Section */}
              <SectionCard delay={100}>
                <div className="px-6 py-4 border-b border-border/50 bg-muted/20">
                  <div className="flex items-center gap-3">
                    <div className="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center">
                      <Box className="w-4 h-4 text-primary" />
                    </div>
                    <div>
                      <h2 className="font-semibold text-sm">App Identity</h2>
                      <p className="text-xs text-muted-foreground">
                        Define your app's name, identifier, and appearance
                      </p>
                    </div>
                  </div>
                </div>

                <div className="p-6 space-y-6">
                  {/* App Name */}
                  <div className="space-y-1.5">
                    <Label htmlFor="app-name" className="text-sm font-medium">
                      App Name <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="app-name"
                      placeholder="My Awesome App"
                      value={formData.name}
                      onChange={(e) => setName(e.target.value)}
                      disabled={isSubmitting}
                      className={cn(
                        'h-12 text-base transition-colors',
                        errors.name && 'border-destructive focus-visible:ring-destructive/30'
                      )}
                      autoFocus
                      maxLength={50}
                      aria-describedby={errors.name ? 'name-error' : 'name-hint'}
                    />
                    <div className="flex items-center justify-between">
                      <div>
                        {errors.name ? (
                          <ValidationMessage message={errors.name} type="error" />
                        ) : nameValidation.valid ? (
                          <ValidationMessage message="Looks good!" type="success" />
                        ) : (
                          <ValidationMessage message="3–50 characters" type="info" />
                        )}
                      </div>
                      <span className="text-xs text-muted-foreground tabular-nums">
                        {formData.name.length}/50
                      </span>
                    </div>
                  </div>

                  {/* App Slug */}
                  <div className="space-y-1.5">
                    <Label htmlFor="app-slug" className="text-sm font-medium">
                      App Slug / Identifier <span className="text-destructive">*</span>
                    </Label>
                    <div className="relative">
                      <Input
                        id="app-slug"
                        placeholder="my-awesome-app"
                        value={formData.slug}
                        onChange={(e) => setSlug(e.target.value)}
                        disabled={isSubmitting}
                        className={cn(
                          'h-12 font-mono transition-colors pr-10',
                          errors.slug && 'border-destructive focus-visible:ring-destructive/30',
                          slugValidation.valid &&
                            !errors.slug &&
                            'border-emerald-500/50 focus-visible:ring-emerald-500/30'
                        )}
                        maxLength={63}
                        aria-describedby={errors.slug ? 'slug-error' : 'slug-hint'}
                      />
                      {slugValidation.valid && !errors.slug && (
                        <CheckCircle2 className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-emerald-500" />
                      )}
                    </div>
                    <div className="flex items-center justify-between">
                      <div>
                        {errors.slug ? (
                          <ValidationMessage message={errors.slug} type="error" />
                        ) : slugPreview ? (
                          <div className="flex items-center gap-1.5 text-xs text-muted-foreground mt-1.5">
                            <Globe className="w-3 h-3 flex-shrink-0" />
                            <span>
                              URL:{' '}
                              <code className="font-mono text-foreground/80">{slugPreview}</code>
                            </span>
                          </div>
                        ) : (
                          <ValidationMessage
                            message="Lowercase letters, numbers, and hyphens only"
                            type="info"
                          />
                        )}
                      </div>
                      <span className="text-xs text-muted-foreground tabular-nums">
                        {formData.slug.length}/63
                      </span>
                    </div>
                  </div>

                  {/* Description */}
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Label htmlFor="app-description" className="text-sm font-medium">
                          Description
                        </Label>
                        <button
                          type="button"
                          onClick={() => generateAIDescription()}
                          disabled={isGeneratingDescription || !formData.name.trim()}
                          className={cn(
                            'inline-flex items-center gap-1 px-2 py-0.5 rounded-md text-xs font-medium',
                            'bg-gradient-to-r from-violet-500/10 to-fuchsia-500/10',
                            'text-violet-600 dark:text-violet-400',
                            'border border-violet-500/20',
                            'hover:from-violet-500/20 hover:to-fuchsia-500/20',
                            'transition-all duration-200',
                            'disabled:opacity-50 disabled:cursor-not-allowed'
                          )}
                          title={
                            formData.name.trim()
                              ? 'AI Generate Description'
                              : 'Enter app name first'
                          }
                        >
                          {isGeneratingDescription ? (
                            <>
                              <Loader2 className="w-3 h-3 animate-spin" />
                              Generating...
                            </>
                          ) : (
                            <>
                              <Sparkles className="w-3 h-3" />
                              AI Generate
                            </>
                          )}
                        </button>
                      </div>
                      <span
                        className={cn(
                          'text-xs tabular-nums',
                          formData.description.length > 450
                            ? formData.description.length >= 500
                              ? 'text-destructive'
                              : 'text-amber-500'
                            : 'text-muted-foreground'
                        )}
                      >
                        {formData.description.length}/500
                      </span>
                    </div>
                    <Textarea
                      id="app-description"
                      placeholder="What does this app do? Describe its purpose and key features..."
                      value={formData.description}
                      onChange={(e) => setDescription(e.target.value)}
                      disabled={isSubmitting}
                      className={cn(
                        'min-h-[100px] resize-none transition-colors',
                        errors.description && 'border-destructive focus-visible:ring-destructive/30'
                      )}
                      maxLength={500}
                    />
                    {errors.description && (
                      <ValidationMessage message={errors.description} type="error" />
                    )}
                  </div>

                  {/* Icon Selection */}
                  <div className="space-y-3">
                    <Label className="text-sm font-medium">App Icon</Label>
                    <div className="flex flex-wrap gap-2 pt-2">
                      {ICON_COLORS.map((color) => (
                        <button
                          key={color.value}
                          type="button"
                          onClick={() => setIconColor(color.value)}
                          className={cn(
                            'w-10 h-10 rounded-lg transition-all duration-200',
                            formData.iconColor === color.value
                              ? 'ring-2 ring-offset-2 ring-offset-background ring-primary scale-110'
                              : 'hover:scale-105 opacity-80 hover:opacity-100'
                          )}
                          style={{
                            background: `linear-gradient(135deg, ${color.from} 0%, ${color.to} 100%)`,
                          }}
                          aria-label={`Select ${color.name} color`}
                        />
                      ))}
                    </div>

                    {/* Icon picker: Radix Popover portals outside SectionCard (overflow-hidden) */}
                    <Popover open={iconPickerOpen} onOpenChange={setIconPickerOpen}>
                      <PopoverTrigger asChild>
                        <button
                          type="button"
                          disabled={isSubmitting}
                          className={cn(
                            'inline-flex items-center gap-2 px-4 py-2 rounded-lg border border-border/50 bg-secondary/50',
                            'hover:bg-secondary transition-colors text-2xl',
                            'disabled:opacity-50 disabled:pointer-events-none'
                          )}
                        >
                          <span>{formData.iconEmoji}</span>
                          <span className="text-xs text-muted-foreground">Change icon</span>
                        </button>
                      </PopoverTrigger>
                      <PopoverContent
                        align="start"
                        side="bottom"
                        className="w-auto p-3"
                        collisionPadding={16}
                      >
                        <IconPickerPanel
                          value={formData.iconEmoji}
                          onSelect={(emoji) => {
                            setIconEmoji(emoji);
                            setIconPickerOpen(false);
                          }}
                        />
                      </PopoverContent>
                    </Popover>
                  </div>
                </div>
              </SectionCard>

              {/* Visibility & Access Section */}
              <SectionCard delay={200}>
                <div className="px-6 py-4 border-b border-border/50 bg-muted/20">
                  <div className="flex items-center gap-3">
                    <div className="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center">
                      <Key className="w-4 h-4 text-primary" />
                    </div>
                    <div>
                      <h2 className="font-semibold text-sm">Visibility & Access</h2>
                      <p className="text-xs text-muted-foreground">
                        Control who can see and use your app
                      </p>
                    </div>
                  </div>
                </div>

                <div className="p-6 space-y-6">
                  {/* Visibility Toggle */}
                  <div className="flex items-center justify-between">
                    <div className="space-y-1">
                      <Label className="text-sm font-medium">
                        {formData.visibility === 'public' ? 'Public App' : 'Private App'}
                      </Label>
                      <p className="text-xs text-muted-foreground">
                        {formData.visibility === 'public'
                          ? "Anyone can discover and use your app's functions"
                          : 'Only you and your team can access this app'}
                      </p>
                    </div>
                    <div className="flex items-center gap-3">
                      {formData.visibility === 'public' ? (
                        <Globe className="w-4 h-4 text-emerald-500" />
                      ) : (
                        <Lock className="w-4 h-4 text-amber-500" />
                      )}
                      <Switch
                        checked={formData.visibility === 'public'}
                        onCheckedChange={(checked) => setVisibility(checked ? 'public' : 'private')}
                        disabled={isSubmitting}
                        aria-label="Toggle app visibility"
                      />
                    </div>
                  </div>

                  {/* Team Access Info */}
                  {user && (
                    <div className="flex items-center gap-3 p-3 rounded-lg bg-muted/30">
                      <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center">
                        <span className="text-sm font-medium text-primary">
                          {(user.username || user.name || user.email)?.charAt(0).toUpperCase()}
                        </span>
                      </div>
                      <div>
                        <p className="text-sm font-medium">
                          {user.username || user.name || user.email}
                        </p>
                        <p className="text-xs text-muted-foreground">Owner</p>
                      </div>
                    </div>
                  )}
                </div>
              </SectionCard>

              {/* Initial Configuration Section */}
              <SectionCard delay={300}>
                <div className="px-6 py-4 border-b border-border/50 bg-muted/20">
                  <div className="flex items-center gap-3">
                    <div className="w-9 h-9 rounded-lg bg-primary/10 flex items-center justify-center">
                      <Zap className="w-4 h-4 text-primary" />
                    </div>
                    <div>
                      <h2 className="font-semibold text-sm">Initial Configuration</h2>
                      <p className="text-xs text-muted-foreground">
                        Set up tags and environment for your app
                      </p>
                    </div>
                  </div>
                </div>

                <div className="p-6 space-y-6">
                  {/* Tags */}
                  <div className="space-y-2">
                    <Label className="text-sm font-medium flex items-center gap-2">
                      <Tag className="w-3.5 h-3.5" />
                      Tags / Labels
                    </Label>
                    <TagInput
                      tags={formData.tags}
                      onAdd={addTag}
                      onRemove={removeTag}
                      disabled={isSubmitting}
                    />
                    <p className="text-xs text-muted-foreground">
                      Press Enter or comma to add tags (max 10)
                    </p>
                  </div>

                  {/* Environment */}
                  <div className="space-y-2">
                    <Label htmlFor="environment" className="text-sm font-medium">
                      Environment
                    </Label>
                    <Select
                      value={formData.environment}
                      onValueChange={(value) => setEnvironment(value as Environment)}
                      disabled={isSubmitting}
                    >
                      <SelectTrigger id="environment" className="h-11">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="development">
                          <div className="flex items-center gap-2">
                            <Code2 className="w-4 h-4 text-blue-500" />
                            Development
                          </div>
                        </SelectItem>
                        <SelectItem value="staging">
                          <div className="flex items-center gap-2">
                            <Layers className="w-4 h-4 text-amber-500" />
                            Staging
                          </div>
                        </SelectItem>
                        <SelectItem value="production">
                          <div className="flex items-center gap-2">
                            <Rocket className="w-4 h-4 text-emerald-500" />
                            Production
                          </div>
                        </SelectItem>
                      </SelectContent>
                    </Select>
                    <p className="text-xs text-muted-foreground">
                      Choose the initial environment for your app
                    </p>
                  </div>
                </div>
              </SectionCard>

              {/* Getting Started Tips */}
              <div className="pt-4">
                <h3 className="text-sm font-semibold mb-4 flex items-center gap-2">
                  <Sparkles className="w-4 h-4 text-primary" />
                  Getting Started
                </h3>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <GettingStartedTip
                    icon={Code2}
                    title="Add Functions"
                    description="Create and deploy serverless functions to power your app"
                    delay={100}
                  />
                  <GettingStartedTip
                    icon={Cloud}
                    title="Configure Backends"
                    description="Connect cloud providers for scalable infrastructure"
                    delay={200}
                  />
                  <GettingStartedTip
                    icon={Rocket}
                    title="Deploy"
                    description="Deploy your app with a single click to the edge"
                    delay={300}
                  />
                </div>
              </div>
            </div>

            {/* Right Column: Preview */}
            <div className="lg:col-span-1">
              <AppPreviewCard formData={formData} />
            </div>
          </div>
        </div>
      </div>
    </TooltipProvider>
  );
}
