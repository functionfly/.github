import {
  Chamber,
  SealedButton,
  FrameButton,
  StatusPill,
} from '@/components/containment';
import type { StateFabricAddOnDTO } from '@/api/billing';
import { Check, Package, Zap } from 'lucide-react';

interface AddOnsTabProps {
  addOnCatalog: StateFabricAddOnDTO[];
  entitledAddOnIds: string[];
  onPurchase: (addonId: string) => Promise<void>;
  isLoading: boolean;
}

export function AddOnsTab({ addOnCatalog, entitledAddOnIds, onPurchase, isLoading }: AddOnsTabProps) {
  if (isLoading) {
    return (
      <div className="sc-billing-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
        <Chamber nested>
          <div style={{ height: 24, width: 192, background: 'var(--panel)', borderRadius: 'var(--radius)', marginBottom: 'var(--space-4)' }} />
          {[1, 2, 3, 4].map((i) => (
            <div key={i} style={{ height: 128, background: 'var(--panel)', borderRadius: 'var(--radius)', marginBottom: 'var(--space-3)' }} />
          ))}
        </Chamber>
      </div>
    );
  }

  if (addOnCatalog.length === 0) {
    return (
      <div className="sc-billing-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
        <Chamber nested>
          <div className="sc-billing-card-title" style={{ marginBottom: 'var(--space-2)' }}>
            <Package style={{ width: 14, height: 14 }} />
            State Fabric Add-ons
          </div>
          <div className="sc-billing-card-description" style={{ marginBottom: 'var(--space-5)' }}>
            Enhance your State Fabric experience with premium add-ons
          </div>
          <div className="empty-state">
            <div style={{ textAlign: 'center' }}>
              <Package style={{ width: 48, height: 48, color: 'var(--text-faint)', margin: '0 auto var(--space-3)' }} />
              <p style={{ color: 'var(--text-faint)' }}>No add-ons available at the moment</p>
            </div>
          </div>
        </Chamber>
      </div>
    );
  }

  return (
    <div className="sc-billing-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--space-6)' }}>
      <Chamber nested>
        <div className="sc-billing-card-header" style={{ margin: 'calc(-1 * var(--space-5))', marginBottom: 'var(--space-5)', padding: 'var(--space-4) var(--space-5)' }}>
          <div className="sc-billing-card-title">
            <Package style={{ width: 14, height: 14 }} />
            State Fabric Add-ons
          </div>
          <div className="sc-billing-card-description">Premium add-ons to enhance your State Fabric experience</div>
        </div>

        <div className="sc-billing-grid sc-billing-grid-2">
          {addOnCatalog.map((addon) => {
            const isEntitled = entitledAddOnIds.includes(addon.id);
            return (
              <div
                key={addon.id}
                style={{
                  padding: 'var(--space-4)',
                  borderRadius: 'var(--radius)',
                  border: `1px solid ${isEntitled ? 'rgba(143, 255, 208, 0.3)' : 'var(--panel-edge)'}`,
                  background: isEntitled ? 'rgba(143, 255, 208, 0.03)' : 'var(--panel)',
                  transition: 'border-color var(--duration-fast) var(--ease-out)',
                }}
              >
                <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', marginBottom: 'var(--space-3)' }}>
                  <div>
                    <h4 style={{ fontFamily: 'var(--font-display)', fontSize: 15, fontWeight: 600, color: 'var(--text)' }}>{addon.name}</h4>
                    <p style={{ fontFamily: 'var(--font-display)', fontSize: 28, fontWeight: 700, color: 'var(--status-ok)', marginTop: 'var(--space-1)' }}>
                      ${addon.price}
                      <span style={{ fontSize: 13, fontWeight: 400, color: 'var(--text-dim)' }}>/{addon.period}</span>
                    </p>
                  </div>
                  {isEntitled && <StatusPill status="live" label="Active" />}
                </div>

                <p style={{ fontSize: 13, color: 'var(--text-dim)', lineHeight: 1.5, marginBottom: 'var(--space-4)' }}>{addon.description}</p>

                {isEntitled ? (
                  <div className="sc-billing-info">
                    <Check style={{ width: 14, height: 14 }} />
                    <div className="sc-billing-info-content">
                      <span className="sc-billing-info-text">You have access to {addon.name}</span>
                    </div>
                  </div>
                ) : (
                  <FrameButton
                    style={{ width: '100%' }}
                    onClick={() => onPurchase(addon.id)}
                    iconLeft={<Zap style={{ width: 14, height: 14 }} />}
                  >
                    Subscribe
                  </FrameButton>
                )}
              </div>
            );
          })}
        </div>
      </Chamber>
    </div>
  );
}
