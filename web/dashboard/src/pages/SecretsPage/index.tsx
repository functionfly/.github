import { usePageTitle } from '@/hooks';
import { SecretList } from '@/components/SecretsVault/SecretList';
import './styles.css';

export function SecretsPage() {
  usePageTitle('Secrets');
  return (
    <div className="secrets-page">
      <div className="secrets-header">
        <h1 className="secrets-title">Secrets</h1>
        <p className="secrets-subtitle">
          Manage API keys, tokens, and other secrets in your vault. Values are encrypted at rest.
        </p>
      </div>
      <SecretList />
      {/* Aviation corner accents */}
      <div className="aviation-corner aviation-corner-tl" />
      <div className="aviation-corner aviation-corner-tr" />
      <div className="aviation-corner aviation-corner-bl" />
      <div className="aviation-corner aviation-corner-br" />
    </div>
  );
}
