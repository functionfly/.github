import { Link } from 'react-router-dom';
import { DollarSign, TrendingUp, Wallet } from 'lucide-react';
import { useTranslation } from 'react-i18next';

export function PurchaseCrossLinks() {
  const { t } = useTranslation();

  return (
    <div className="flex flex-wrap items-center gap-x-4 gap-y-2 border-t border-aviation-border-panel pt-4 text-sm">
      <span className="text-aviation-text-dim">{t('purchasesPage.relatedLinks')}</span>
      <Link
        to="/wallet"
        className="inline-flex items-center gap-1.5 text-aviation-cyan hover:underline"
      >
        <Wallet className="h-3.5 w-3.5" />
        {t('purchasesPage.viewWallet')}
      </Link>
      <Link
        to="/marketplace-economy"
        className="inline-flex items-center gap-1.5 text-aviation-amber hover:underline"
      >
        <TrendingUp className="h-3.5 w-3.5" />
        {t('purchasesPage.viewCreatorEarnings')}
      </Link>
      <Link
        to="/marketplace"
        className="inline-flex items-center gap-1.5 text-aviation-text-secondary hover:text-aviation-text-primary hover:underline"
      >
        <DollarSign className="h-3.5 w-3.5" />
        {t('purchasesPage.browseMarketplace')}
      </Link>
    </div>
  );
}
