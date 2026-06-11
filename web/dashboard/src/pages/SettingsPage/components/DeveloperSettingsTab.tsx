import { ApiKeysSettingsSection } from "./ApiKeysSettingsSection";
import { ConnectedProvidersSection } from "./ConnectedProvidersSection";
import { DeployKeysSettingsSection } from "./DeployKeysSettingsSection";
import { WebhooksSettingsSection } from "./WebhooksSettingsSection";

export function DeveloperSettingsTab() {
  return (
    <div className="settings-page space-y-6">
      <ApiKeysSettingsSection />
      <ConnectedProvidersSection />
      <DeployKeysSettingsSection />
      <WebhooksSettingsSection />
    </div>
  );
}