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
import { useTranslation } from 'react-i18next';
import { Link, useNavigate } from 'react-router-dom';
import '@/styles/aviation-create-app.css';
import { ICON_COLORS, ICON_OPTIONS, useCreateApp, type Environment } from './hooks/useCreateApp';

type IconPickerTab = 'emoji' | 'icons';

// Icon picker panel (used inside Radix Popover so it is not clipped by parent overflow-hidden)
function IconPickerPanel({
  value,
  onSelect,
}: {
  value: string;
  onSelect: (value: string) => void;
}) {
  const { t } = useTranslation();
  const [activeTab, setActiveTab] = useState<IconPickerTab>('emoji');

  return (
    <div className="icon-picker-panel">
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
          {t('appsPage.emoji')}
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
          {t('appsPage.icons')}
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
        {t('appsPage.clickToSelect')}{' '}
        {ICON_OPTIONS.find((o) => o.emoji === value)?.label || t('appsPage.custom')}
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
    <div className={cn('validation-message', styles[type])}>
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
        'section-card animate-fade-in-up opacity-0',
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
  const { t } = useTranslation();
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
        'tag-input-container',
        disabled && 'opacity-50 cursor-not-allowed'
      )}
      onClick={() => inputRef.current?.focus()}
    >
      {tags.map((tag) => (
        <span
          key={tag}
          className="tag-item"
        >
          {tag}
          <button
            type="button"
            onClick={(e) => {
              e.stopPropagation();
              onRemove(tag);
            }}
            className="tag-remove-btn"
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
        placeholder={tags.length === 0 ? t('appsPage.addTagsPlaceholder') : ''}
        className="tag-input-field"
        disabled={disabled}
      />
    </div>
  );
}

function AppPreviewCard({ formData }: { formData: ReturnType<typeof useCreateApp>['formData'] }) {
  const { t } = useTranslation();
  const selectedColor = ICON_COLORS.find((c) => c.value === formData.iconColor) || ICON_COLORS[0];

  return (
    <Card className="app-preview-card animate-slide-in-right opacity-0">
      <CardContent className="p-0">
        {/* Preview Header */}
        <div className="app-preview-header">
          <p className="app-preview-label">
            {t('appsPage.livePreview')}
          </p>
        </div>

        {/* Preview Content */}
        <div className="app-preview-content space-y-6">
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
                  <div className="app-preview-icon">
                    {formData.iconEmoji}
                  </div>
                  <span
                    className={cn(
                      'app-preview-visibility',
                      formData.visibility === 'public' ? 'app-preview-visibility-public' : 'app-preview-visibility-private'
                    )}
                  >
                    {formData.visibility === 'public' ? t('appsPage.public') : t('appsPage.private')}
                  </span>
                </div>
                <div className="mt-4">
                  <h3 className="app-preview-name">
                    {formData.name || t('appsPage.untitledApp')}
                  </h3>
                  <p className="app-preview-slug mt-0.5">
                    {formData.slug || t('appsPage.appSlugDefault')}
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
                    className="app-preview-tag"
                  >
                    {tag}
                  </span>
                ))}
                {formData.tags.length > 3 && (
                  <span className="app-preview-tag">
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
              <p className="text-xs text-muted-foreground mb-1.5">{t('appsPage.appUrl')}</p>
              <code className="app-preview-url">
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
  icon: React.ComponentType<{ className?: string }>;
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
        'getting-started-tip',
        'animate-fade-in-up opacity-0',
        delayClass
      )}
      style={{ animationFillMode: 'forwards' }}
    >
      <div className="flex items-start gap-4">
        <div className="getting-started-tip-icon">
          <Icon className="w-5 h-5 text-primary" />
        </div>
        <div className="flex-1">
          <h4 className="getting-started-tip-title">{title}</h4>
          <p className="getting-started-tip-description mt-1">{description}</p>
        </div>
      </div>
    </div>
  );
}

export function CreateAppPage() {
  const { t } = useTranslation();
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
    if (seconds < 5) return t('appsPage.justNow');
    if (seconds < 60) return t('appsPage.secondsAgo', { count: seconds });
    const minutes = Math.floor(seconds / 60);
    if (minutes < 60) return t('appsPage.minutesAgo', { count: minutes });
    return t('appsPage.hoursAgo', { count: Math.floor(minutes / 60) });
  };

  return (
    <TooltipProvider>
      <div className="aviation-create-app min-h-screen">
        {/* Sticky Action Bar */}
        <div className="action-bar sticky top-0 z-50 border-b">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 h-14 flex items-center justify-between gap-4">
            {/* Left: Breadcrumb */}
            <div className="flex items-center gap-3 min-w-0">
              <Button
                variant="ghost"
                size="icon"
                onClick={handleCancel}
                className="shrink-0 text-muted-foreground hover:text-foreground"
                aria-label={t('appsPage.backToApps')}
              >
                <ArrowLeft className="w-4 h-4" />
              </Button>
              <nav className="flex items-center gap-1.5 text-sm min-w-0" aria-label="Breadcrumb">
                <Link
                  to="/apps"
                  className="breadcrumb-link hover:text-foreground transition-colors truncate"
                >
                  {t('appsPage.breadcrumbApps')}
                </Link>
                <ChevronRight className="w-3.5 h-3.5 text-muted-foreground shrink-0" />
                <span className="breadcrumb-current font-medium truncate">{t('appsPage.breadcrumbCreate')}</span>
              </nav>
            </div>

            {/* Center: Auto-save indicator */}
            <div className="hidden sm:flex items-center gap-2 text-xs text-muted-foreground">
              {isSubmitting ? (
                <span className="autosave-indicator autosave-submitting flex items-center gap-1.5">
                  <Loader2 className="w-3.5 h-3.5 animate-spin text-primary" />
                  {t('appsPage.creatingApp')}
                </span>
              ) : isDirty ? (
                lastSaved ? (
                  <span className="autosave-indicator autosave-saved flex items-center gap-1.5">
                    <CheckCircle2 className="w-3.5 h-3.5" />
                    {t('appsPage.draftSaved', { time: formatRelativeTime(lastSaved) })}
                  </span>
                ) : (
                  <span className="autosave-indicator autosave-dirty flex items-center gap-1.5">
                    <Save className="w-3.5 h-3.5" />
                    {t('appsPage.unsavedChanges')}
                  </span>
                )
              ) : (
                <span className="autosave-indicator flex items-center gap-1.5">
                  <Clock className="w-3.5 h-3.5" />
                  {t('appsPage.readyToCreate')}
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
                {t('appsPage.cancel')}
              </Button>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    size="sm"
                    onClick={handleSubmit}
                    disabled={!isValid || isSubmitting}
                    className={cn(
                      'gap-2 font-semibold min-w-[120px]',
                      isValid ? 'create-btn create-btn-enabled' : 'create-btn-disabled'
                    )}
                  >
                    {isSubmitting ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin" />
                        {t('appsPage.creating')}
                      </>
                    ) : (
                      <>
                        <Sparkles className="w-4 h-4" />
                        {t('appsPage.createApp')}
                      </>
                    )}
                  </Button>
                </TooltipTrigger>
                {!isValid && !isSubmitting && (
                  <TooltipContent>
                    <p>{t('appsPage.fillRequiredFields')}</p>
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
            className="page-header mb-8 animate-fade-in-up opacity-0"
            style={{ animationFillMode: 'forwards' }}
          >
            <div className="flex items-start gap-4">
              <div className="page-header-icon animate-pulse-soft">
                <Rocket className="w-7 h-7" />
              </div>
              <div>
                <h1 className="page-header-title">{t('appsPage.createNewApp')}</h1>
                <p className="page-header-description mt-1.5 text-base">
                  {t('appsPage.appsDescription')}
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
                <div className="general-error animate-fade-in-up">
                  <AlertCircle className="w-4 h-4 mt-0.5 flex-shrink-0" />
                  <span>{errors.general}</span>
                </div>
              )}

              {/* App Identity Section */}
              <SectionCard delay={100}>
                <div className="section-header">
                  <div className="flex items-center gap-3">
                    <div className="section-header-icon">
                      <Box className="w-4 h-4 text-primary" />
                    </div>
                    <div>
                      <h2 className="section-header-title">{t('appsPage.appIdentity')}</h2>
                      <p className="section-header-description">
                        {t('appsPage.appIdentityDescription')}
                      </p>
                    </div>
                  </div>
                </div>

                <div className="section-content space-y-6">
                  {/* App Name */}
                  <div className="space-y-1.5">
                    <Label htmlFor="app-name" className="form-label">
                      {t('appsPage.appName')} <span className="text-destructive">*</span>
                    </Label>
                    <Input
                      id="app-name"
                      placeholder="My Awesome App"
                      value={formData.name}
                      onChange={(e) => setName(e.target.value)}
                      disabled={isSubmitting}
                      className={cn(
                        'form-input',
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
                          <ValidationMessage message={t('appsPage.looksGood')} type="success" />
                        ) : (
                          <ValidationMessage message={t('appsPage.nameCharLimit')} type="info" />
                        )}
                      </div>
                      <span className="char-counter">
                        {formData.name.length}/50
                      </span>
                    </div>
                  </div>

                  {/* App Slug */}
                  <div className="space-y-1.5">
                    <Label htmlFor="app-slug" className="form-label">
                      {t('appsPage.appSlug')} <span className="text-destructive">*</span>
                    </Label>
                    <div className="relative">
                      <Input
                        id="app-slug"
                        placeholder="my-awesome-app"
                        value={formData.slug}
                        onChange={(e) => setSlug(e.target.value)}
                        disabled={isSubmitting}
                        className={cn(
                          'form-input form-input-mono pr-10',
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
                              {t('appsPage.urlPrefix')}{' '}
                              <code className="font-mono text-foreground/80">{slugPreview}</code>
                            </span>
                          </div>
                        ) : (
                          <ValidationMessage
                            message={t('appsPage.slugHint')}
                            type="info"
                          />
                        )}
                      </div>
                      <span className="char-counter">
                        {formData.slug.length}/63
                      </span>
                    </div>
                  </div>

                  {/* Description */}
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <Label htmlFor="app-description" className="form-label">
                          {t('appsPage.description')}
                        </Label>
                        <button
                          type="button"
                          onClick={() => generateAIDescription()}
                          disabled={isGeneratingDescription || !formData.name.trim()}
                          className="ai-generate-btn"
                          title={
                            formData.name.trim()
                              ? t('appsPage.aiGenerateDescription')
                              : t('appsPage.enterAppNameFirst')
                          }
                        >
                          {isGeneratingDescription ? (
                            <>
                              <Loader2 className="w-3 h-3 animate-spin" />
                              {t('appsPage.generating')}
                            </>
                          ) : (
                            <>
                              <Sparkles className="w-3 h-3" />
                              {t('appsPage.aiGenerate')}
                            </>
                          )}
                        </button>
                      </div>
                      <span
                        className={cn(
                          'char-counter',
                          formData.description.length > 450
                            ? formData.description.length >= 500
                              ? 'char-counter-error'
                              : 'char-counter-warning'
                            : ''
                        )}
                      >
                        {formData.description.length}/500
                      </span>
                    </div>
                    <Textarea
                      id="app-description"
                      placeholder={t('appsPage.descriptionPlaceholder')}
                      value={formData.description}
                      onChange={(e) => setDescription(e.target.value)}
                      disabled={isSubmitting}
                      className={cn(
                        'form-textarea',
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
                    <Label className="form-label">{t('appsPage.appIcon')}</Label>
                    <div className="flex flex-wrap gap-2 pt-2">
                      {ICON_COLORS.map((color) => (
                        <button
                          key={color.value}
                          type="button"
                          onClick={() => setIconColor(color.value)}
                          className={cn(
                            'icon-color-btn',
                            formData.iconColor === color.value
                              ? 'icon-color-btn-selected'
                              : ''
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
                          className="icon-picker-trigger"
                        >
                          <span>{formData.iconEmoji}</span>
                          <span className="text-xs text-muted-foreground">{t('appsPage.changeIcon')}</span>
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
                <div className="section-header">
                  <div className="flex items-center gap-3">
                    <div className="section-header-icon">
                      <Key className="w-4 h-4 text-primary" />
                    </div>
                    <div>
                      <h2 className="section-header-title">{t('appsPage.visibilityAccess')}</h2>
                      <p className="section-header-description">
                        {t('appsPage.visibilityAccessDescription')}
                      </p>
                    </div>
                  </div>
                </div>

                <div className="section-content space-y-6">
                  {/* Visibility Toggle */}
                  <div className="visibility-toggle">
                    <div className="space-y-1">
                      <Label className="form-label">
                        {formData.visibility === 'public' ? t('appsPage.publicApp') : t('appsPage.privateApp')}
                      </Label>
                      <p className="text-xs text-muted-foreground">
                        {formData.visibility === 'public'
                          ? t('appsPage.publicAppDescription')
                          : t('appsPage.privateAppDescription')}
                      </p>
                    </div>
                    <div className="flex items-center gap-3">
                      {formData.visibility === 'public' ? (
                        <Globe className="visibility-icon-public w-4 h-4" />
                      ) : (
                        <Lock className="visibility-icon-private w-4 h-4" />
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
                    <div className="team-access-card">
                      <div className="team-access-avatar">
                        <span>
                          {(user.username || user.name || user.email)?.charAt(0).toUpperCase()}
                        </span>
                      </div>
                      <div>
                        <p className="team-access-name">
                          {user.username || user.name || user.email}
                        </p>
                        <p className="team-access-role">{t('appsPage.owner')}</p>
                      </div>
                    </div>
                  )}
                </div>
              </SectionCard>

              {/* Initial Configuration Section */}
              <SectionCard delay={300}>
                <div className="section-header">
                  <div className="flex items-center gap-3">
                    <div className="section-header-icon">
                      <Zap className="w-4 h-4 text-primary" />
                    </div>
                    <div>
                      <h2 className="section-header-title">{t('appsPage.initialConfiguration')}</h2>
                      <p className="section-header-description">
                        {t('appsPage.initialConfigurationDescription')}
                      </p>
                    </div>
                  </div>
                </div>

                <div className="section-content space-y-6">
                  {/* Tags */}
                  <div className="space-y-2">
                    <Label className="form-label flex items-center gap-2">
                      <Tag className="w-3.5 h-3.5" />
                      {t('appsPage.tagsLabels')}
                    </Label>
                    <TagInput
                      tags={formData.tags}
                      onAdd={addTag}
                      onRemove={removeTag}
                      disabled={isSubmitting}
                    />
                    <p className="text-xs text-muted-foreground">
                      {t('appsPage.tagsHint')}
                    </p>
                  </div>

                  {/* Environment */}
                  <div className="space-y-2">
                    <Label htmlFor="environment" className="form-label">
                      {t('appsPage.environment')}
                    </Label>
                    <Select
                      value={formData.environment}
                      onValueChange={(value) => setEnvironment(value as Environment)}
                      disabled={isSubmitting}
                    >
                      <SelectTrigger id="environment" className="environment-select h-11">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="development">
                          <div className="environment-option">
                            <Code2 className="w-4 h-4 text-blue-500" />
                            {t('appsPage.development')}
                          </div>
                        </SelectItem>
                        <SelectItem value="staging">
                          <div className="environment-option">
                            <Layers className="w-4 h-4 text-amber-500" />
                            {t('appsPage.staging')}
                          </div>
                        </SelectItem>
                        <SelectItem value="production">
                          <div className="environment-option">
                            <Rocket className="w-4 h-4 text-emerald-500" />
                            {t('appsPage.production')}
                          </div>
                        </SelectItem>
                      </SelectContent>
                    </Select>
                    <p className="text-xs text-muted-foreground">
                      {t('appsPage.environmentHint')}
                    </p>
                  </div>
                </div>
              </SectionCard>

              {/* Getting Started Tips */}
              <div className="getting-started-section">
                <h3 className="getting-started-title">
                  <Sparkles className="w-4 h-4 text-primary" />
                  {t('appsPage.gettingStarted')}
                </h3>
                <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                  <GettingStartedTip
                    icon={Code2}
                    title={t('appsPage.tipAddFunctions')}
                    description={t('appsPage.tipAddFunctionsDesc')}
                    delay={100}
                  />
                  <GettingStartedTip
                    icon={Cloud}
                    title={t('appsPage.tipConfigureBackends')}
                    description={t('appsPage.tipConfigureBackendsDesc')}
                    delay={200}
                  />
                  <GettingStartedTip
                    icon={Rocket}
                    title={t('appsPage.tipDeploy')}
                    description={t('appsPage.tipDeployDesc')}
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
