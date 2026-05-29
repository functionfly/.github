import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';
import { motion } from 'framer-motion';
import { Bot, CreditCard, FunctionSquare, Key, ShoppingBag } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { KIND_META } from '../constants';

const GHOST_ITEMS = [
  { kind: 'function' as const, title: 'author/my-function', meta: 'Function purchase' },
  { kind: 'agent' as const, title: 'Research task', meta: 'Agent hiring' },
  { kind: 'license' as const, title: 'Pro license', meta: 'License key' },
  { kind: 'subscription' as const, title: 'Creator plan', meta: 'Monthly subscription' },
];

export function PurchaseEmptyState() {
  const { t } = useTranslation();

  return (
    <motion.div
      initial={{ opacity: 0, y: 12 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, ease: 'easeOut' }}
      className="relative overflow-hidden rounded-xl border border-dashed border-aviation-border-panel bg-aviation-bg-panel/40 px-6 py-12 text-center"
    >
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.04]"
        style={{
          backgroundImage:
            'linear-gradient(var(--aviation-border-panel) 1px, transparent 1px), linear-gradient(90deg, var(--aviation-border-panel) 1px, transparent 1px)',
          backgroundSize: '24px 24px',
        }}
      />

      <motion.div
        initial={{ scale: 0.9, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        transition={{ delay: 0.1 }}
        className="relative mx-auto mb-5 flex h-14 w-14 items-center justify-center rounded-full bg-aviation-bg-instrument ring-1 ring-aviation-cyan/20"
      >
        <ShoppingBag className="h-7 w-7 text-aviation-cyan" />
      </motion.div>

      <h2 className="relative text-lg font-semibold text-aviation-text-primary">
        {t('purchasesPage.emptyTitle')}
      </h2>
      <p className="relative mx-auto mt-2 max-w-md text-sm text-aviation-text-secondary">
        {t('purchasesPage.emptyBody')}
      </p>

      <div className="relative mx-auto mt-8 grid max-w-2xl gap-3 sm:grid-cols-2">
        {GHOST_ITEMS.map((item, i) => {
          const meta = KIND_META[item.kind];
          const Icon = meta.icon;
          return (
            <motion.div
              key={item.kind}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 0.45, y: 0 }}
              transition={{ delay: 0.15 + i * 0.08 }}
              className={cn(
                'rounded-lg border border-aviation-border-instrument/60 bg-aviation-bg-instrument/30 p-3 text-left border-l-4',
                meta.accent
              )}
            >
              <div className="flex items-center gap-2">
                <div className={cn('rounded-md p-1.5', meta.iconBg)}>
                  <Icon className="h-3.5 w-3.5" />
                </div>
                <div className="min-w-0">
                  <p className="truncate font-mono text-xs text-aviation-text-primary">{item.title}</p>
                  <p className="text-[10px] text-aviation-text-dim">{item.meta}</p>
                </div>
              </div>
            </motion.div>
          );
        })}
      </div>

      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.45 }}
        className="relative mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row"
      >
        <Button asChild size="lg">
          <Link to="/marketplace/agents">
            <Bot className="mr-2 h-4 w-4" />
            {t('purchasesPage.browseAgents')}
          </Link>
        </Button>
        <Button asChild variant="outline" size="lg">
          <Link to="/functions/discovery">
            <FunctionSquare className="mr-2 h-4 w-4" />
            {t('purchasesPage.browseFunctions')}
          </Link>
        </Button>
        <Link
          to="/marketplace"
          className="text-sm text-aviation-text-secondary underline-offset-4 hover:text-aviation-cyan hover:underline"
        >
          {t('purchasesPage.browseMarketplace')}
        </Link>
      </motion.div>

      <p className="relative mt-6 flex items-center justify-center gap-3 text-[10px] uppercase tracking-widest text-aviation-text-dim">
        <CreditCard className="h-3 w-3" />
        {t('purchasesPage.emptyHint')}
        <Key className="h-3 w-3" />
      </p>
    </motion.div>
  );
}
