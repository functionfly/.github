import { SecretList } from "@/components/SecretsVault/SecretList";

export function SecretsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight text-text-primary">
          Secrets
        </h1>
        <p className="mt-1 text-text-secondary">
          Manage API keys, tokens, and other secrets in your vault. Values are encrypted at rest.
        </p>
      </div>
      <SecretList />
    </div>
  );
}
