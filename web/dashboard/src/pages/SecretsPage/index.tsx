import { useEffect } from "react";
import { usePageTitle } from '@/hooks';
import { SecretList } from '@/components/SecretsVault/SecretList';
import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";
import './styles.css';

interface SecretsPageProps {
  deprecated?: boolean;
}

export function SecretsPage({ deprecated }: SecretsPageProps) {
  usePageTitle('Secrets');

  useEffect(() => {
    if (deprecated) {
      const dismissed = sessionStorage.getItem("secrets-deprecation-dismissed");
      if (!dismissed) {
        sessionStorage.setItem("secrets-deprecation-dismissed", "true");
        window.location.href = "/vault";
      }
    }
  }, [deprecated]);

  return (
    <div className="secrets-page">
      {deprecated && (
        <div className="mb-4 p-3 border border-amber-500/50 bg-amber-500/10 rounded-md flex items-center gap-3">
          <AlertTriangle className="h-5 w-5 text-amber-500 shrink-0" />
          <p className="text-sm flex-1">
            This page has moved to <a href="/vault" className="underline font-medium">/vault</a>. Please update your bookmarks.
          </p>
          <Button size="sm" variant="outline" onClick={() => window.location.href = "/vault"}>
            Go to Vault
          </Button>
        </div>
      )}
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
