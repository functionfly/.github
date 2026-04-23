import { useState } from 'react';
import { motion } from 'framer-motion';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import {
  Play,
  RotateCcw,
  Code2,
  Link2,
  Settings2,
  Check,
  Loader2,
  PanelRightOpen,
  PanelRightClose,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { usePlaygroundStore } from '../store/playgroundStore';
import { usePlaygroundState } from '../hooks/usePlaygroundState';

interface PlaygroundToolbarProps {
  author: string;
  name: string;
}

export function PlaygroundToolbar({ author, name }: PlaygroundToolbarProps) {
  const { t } = useTranslation();
  const {
    execute,
    formatJson,
    resetPlayground,
    settings,
    updateSettings,
    sidebarOpen,
    setSidebarOpen,
    isExecuting,
  } = usePlaygroundStore();

  const { shareableUrl, isInputValid } = usePlaygroundState();
  const [copiedLink, setCopiedLink] = useState(false);

  const handleRun = () => execute(author, name);

  const handleCopyLink = async () => {
    try {
      await navigator.clipboard.writeText(shareableUrl);
      setCopiedLink(true);
      setTimeout(() => setCopiedLink(false), 2000);
    } catch {
      // ignore
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.2, delay: 0.1 }}
      className="flex items-center gap-2 px-4 py-2 border-b border-border-subtle bg-bg-secondary"
    >
      {/* Run button */}
      <Button
        onClick={handleRun}
        disabled={isExecuting || !isInputValid}
        size="sm"
        className="gap-2 bg-indigo-600 hover:bg-indigo-700 text-white h-8"
      >
        {isExecuting ? (
          <Loader2 className="w-3.5 h-3.5 animate-spin" />
        ) : (
          <Play className="w-3.5 h-3.5" />
        )}
        {isExecuting ? t('playground.running') : t('playground.run')}
        <kbd className="hidden sm:inline-flex items-center gap-0.5 text-[10px] opacity-60 ml-1 font-mono">
          ⌘↵
        </kbd>
      </Button>

      <div className="w-px h-5 bg-border-subtle" />

      {/* Format JSON */}
      <Button
        variant="ghost"
        size="sm"
        onClick={formatJson}
        className="gap-1.5 h-8 text-text-secondary hover:text-text-primary"
        title={t('playground.format') + ' (⌘⇧F)'}
      >
        <Code2 className="w-3.5 h-3.5" />
        <span className="hidden sm:inline text-xs">{t('playground.format')}</span>
      </Button>

      {/* Reset */}
      <Button
        variant="ghost"
        size="sm"
        onClick={resetPlayground}
        className="gap-1.5 h-8 text-text-secondary hover:text-text-primary"
        title={t('playground.reset') + ' playground (⌘⇧R)'}
      >
        <RotateCcw className="w-3.5 h-3.5" />
        <span className="hidden sm:inline text-xs">{t('playground.reset')}</span>
      </Button>

      {/* Copy Link */}
      <Button
        variant="ghost"
        size="sm"
        onClick={handleCopyLink}
        className="gap-1.5 h-8 text-text-secondary hover:text-text-primary"
        title={t('playground.share') + ' (⌘⇧C)'}
      >
        {copiedLink ? (
          <Check className="w-3.5 h-3.5 text-green-500" />
        ) : (
          <Link2 className="w-3.5 h-3.5" />
        )}
        <span className="hidden sm:inline text-xs">
          {copiedLink ? t('playground.copied') : t('playground.share')}
        </span>
      </Button>

      <div className="flex-1" />

      {/* Settings dropdown */}
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            variant="ghost"
            size="sm"
            className="gap-1.5 h-8 text-text-secondary hover:text-text-primary"
          >
            <Settings2 className="w-3.5 h-3.5" />
            <span className="hidden sm:inline text-xs">{t('playground.settings')}</span>
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" className="w-52">
          <DropdownMenuLabel className="text-xs">{t('playground.playgroundSettings')}</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onClick={() => updateSettings({ autoRun: !settings.autoRun })}
            className="flex items-center justify-between text-xs"
          >
            <span>{t('playground.autoRunOnChange')}</span>
            <div
              className={`w-8 h-4 rounded-full transition-colors ${
                settings.autoRun ? 'bg-indigo-500' : 'bg-border-subtle'
              }`}
            >
              <div
                className={`w-3 h-3 rounded-full bg-white mt-0.5 transition-transform ${
                  settings.autoRun ? 'translate-x-4' : 'translate-x-0.5'
                }`}
              />
            </div>
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => updateSettings({ showTimeline: !settings.showTimeline })}
            className="flex items-center justify-between text-xs"
          >
            <span>{t('playground.showExecutionTimeline')}</span>
            <div
              className={`w-8 h-4 rounded-full transition-colors ${
                settings.showTimeline ? 'bg-indigo-500' : 'bg-border-subtle'
              }`}
            >
              <div
                className={`w-3 h-3 rounded-full bg-white mt-0.5 transition-transform ${
                  settings.showTimeline ? 'translate-x-4' : 'translate-x-0.5'
                }`}
              />
            </div>
          </DropdownMenuItem>
          <DropdownMenuItem
            onClick={() => updateSettings({ showHeaders: !settings.showHeaders })}
            className="flex items-center justify-between text-xs"
          >
            <span>{t('playground.showResponseHeaders')}</span>
            <div
              className={`w-8 h-4 rounded-full transition-colors ${
                settings.showHeaders ? 'bg-indigo-500' : 'bg-border-subtle'
              }`}
            >
              <div
                className={`w-3 h-3 rounded-full bg-white mt-0.5 transition-transform ${
                  settings.showHeaders ? 'translate-x-4' : 'translate-x-0.5'
                }`}
              />
            </div>
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Toggle sidebar */}
      <Button
        variant="ghost"
        size="sm"
        onClick={() => setSidebarOpen(!sidebarOpen)}
        className="gap-1.5 h-8 text-text-secondary hover:text-text-primary"
        title={t('playground.toggleSidebar')}
      >
        {sidebarOpen ? (
          <PanelRightClose className="w-3.5 h-3.5" />
        ) : (
          <PanelRightOpen className="w-3.5 h-3.5" />
        )}
      </Button>
    </motion.div>
  );
}
