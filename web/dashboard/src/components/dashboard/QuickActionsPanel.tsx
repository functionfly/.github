import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';
import {
  Cloud,
  Code2,
  Database,
  FileCode,
  Globe,
  Layers,
  Lock,
  Network,
  Plus,
  Rocket,
  Settings,
  Shield,
  Terminal,
  Webhook,
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

export interface QuickAction {
  id: string;
  label: string;
  description: string;
  icon: React.ReactNode;
  onClick: () => void;
  variant?: 'default' | 'primary' | 'secondary';
  disabled?: boolean;
}

export interface QuickActionsPanelProps {
  onCreateFunction?: () => void;
  onCreateApp?: () => void;
  onConnectProvider?: () => void;
  onViewSecrets?: () => void;
  onViewLogs?: () => void;
  onSettings?: () => void;
  onCreateGraph?: () => void;
  className?: string;
}

const actionIcons = {
  function: <Code2 className="w-5 h-5" />,
  app: <Layers className="w-5 h-5" />,
  provider: <Cloud className="w-5 h-5" />,
  secret: <Lock className="w-5 h-5" />,
  logs: <Terminal className="w-5 h-5" />,
  settings: <Settings className="w-5 h-5" />,
  webhook: <Webhook className="w-5 h-5" />,
  database: <Database className="w-5 h-5" />,
  template: <FileCode className="w-5 h-5" />,
  deploy: <Rocket className="w-5 h-5" />,
  graph: <Network className="w-5 h-5" />,
  domain: <Globe className="w-5 h-5" />,
  security: <Shield className="w-5 h-5" />,
};

function ActionButton({ action, index }: { action: QuickAction; index: number }) {
  const baseStyles =
    'group flex flex-col items-center justify-center gap-2 p-4 rounded-xl border transition-all duration-200 cursor-pointer text-center';

  const variantStyles = {
    default:
      'border-border bg-bg-secondary hover:border-[var(--color-aviation-amber)]/50 hover:bg-[var(--color-aviation-amber)]/5',
    primary:
      'border-[var(--color-aviation-amber)]/50 bg-[var(--color-aviation-amber)]/10 hover:bg-[var(--color-aviation-amber)]/20',
    secondary:
      'border-[var(--color-aviation-cyan)]/50 bg-[var(--color-aviation-cyan)]/10 hover:bg-[var(--color-aviation-cyan)]/20',
  };

  return (
    <motion.button
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay: index * 0.05 }}
      onClick={action.onClick}
      disabled={action.disabled}
      className={cn(
        baseStyles,
        variantStyles[action.variant || 'default'],
        action.disabled && 'opacity-50 cursor-not-allowed'
      )}
    >
      <div
        className={cn(
          'flex items-center justify-center w-11 h-11 rounded-lg transition-colors',
          action.variant === 'primary'
            ? 'bg-(--color-aviation-amber)/20 text-(--color-aviation-amber)'
            : action.variant === 'secondary'
              ? 'bg-(--color-aviation-cyan)/20 text-(--color-aviation-cyan)'
              : 'bg-bg-tertiary text-text-secondary group-hover:text-(--color-aviation-amber)'
        )}
      >
        {action.icon}
      </div>
      <div className="w-full">
        <p className="font-semibold text-text-primary text-sm">{action.label}</p>
        <p className="text-xs text-text-muted mt-0.5">{action.description}</p>
      </div>
    </motion.button>
  );
}

export function QuickActionsPanel({
  onCreateFunction,
  onCreateApp,
  onConnectProvider,
  onViewSecrets,
  onViewLogs,
  onSettings,
  onCreateGraph,
  className,
}: QuickActionsPanelProps) {
  const { t } = useTranslation();
  const actions: QuickAction[] = [
    {
      id: 'function',
      label: t('quickActions.newFunction'),
      description: t('quickActions.deployServerless'),
      icon: actionIcons.function,
      onClick: onCreateFunction || (() => {}),
      variant: 'primary',
    },
    {
      id: 'graph',
      label: t('quickActions.newGraph'),
      description: t('quickActions.createFunctionGraph'),
      icon: actionIcons.graph,
      onClick: onCreateGraph || (() => {}),
    },
    {
      id: 'app',
      label: t('quickActions.newApp'),
      description: t('quickActions.createNewApp'),
      icon: actionIcons.app,
      onClick: onCreateApp || (() => {}),
    },
    {
      id: 'provider',
      label: t('quickActions.connectProvider'),
      description: t('quickActions.linkCloudProvider'),
      icon: actionIcons.provider,
      onClick: onConnectProvider || (() => {}),
    },
    {
      id: 'secret',
      label: t('quickActions.addSecret'),
      description: t('quickActions.storeSecrets'),
      icon: actionIcons.secret,
      onClick: onViewSecrets || (() => {}),
    },
    {
      id: 'logs',
      label: t('quickActions.viewLogs'),
      description: t('quickActions.checkFunctionLogs'),
      icon: actionIcons.logs,
      onClick: onViewLogs || (() => {}),
    },
    {
      id: 'settings',
      label: t('quickActions.settings'),
      description: t('quickActions.managePreferences'),
      icon: actionIcons.settings,
      onClick: onSettings || (() => {}),
    },
  ];

  return (
    <Card className={cn('border-theme bg-card overflow-hidden', className)}>
      <CardHeader className="pb-3 pt-4">
        <div className="flex items-center justify-between">
          <CardTitle className="text-sm font-medium text-text-secondary">{t('quickActions.title')}</CardTitle>
          <button
            className="flex items-center justify-center w-7 h-7 rounded-md bg-bg-tertiary hover:bg-bg-hover transition-colors border border-border-subtle"
            aria-label={t('quickActions.addNew')}
          >
            <Plus className="w-4 h-4 text-text-muted" />
          </button>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="grid grid-cols-2 gap-3">
          {actions.map((action, index) => (
            <ActionButton key={action.id} action={action} index={index} />
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
