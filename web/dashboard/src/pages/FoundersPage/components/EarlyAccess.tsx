import { Zap } from 'lucide-react';
import type { EarlyAccessFeature } from '../hooks/useFounderConsole';

interface EarlyAccessProps {
  features: EarlyAccessFeature[];
  claimFeature: (slug: string) => Promise<void>;
}

export function EarlyAccess({ features, claimFeature }: EarlyAccessProps) {
  if (features.length === 0) return null;

  return (
    <section className="founders-section">
      <div className="founders-section__title">
        <Zap size={14} />
        Early Access
      </div>
      <div className="founders-grid">
        {features.map((feature) => (
          <div key={feature.slug} className="founders-chamber founders-chamber--medium">
            <div className="feature-card">
              <div className="feature-card__info">
                <h3 className="feature-card__name">{feature.name}</h3>
                {feature.description && (
                  <p className="feature-card__description">{feature.description}</p>
                )}
              </div>
              {feature.is_claimed ? (
                <span className="status-pill status-pill--live">
                  <span className="status-pill__dot" />
                  Claimed
                </span>
              ) : (
                <button
                  className="sealed-button sealed-button--sm"
                  onClick={() => claimFeature(feature.slug)}
                >
                  Claim
                </button>
              )}
            </div>
          </div>
        ))}
      </div>
    </section>
  );
}
