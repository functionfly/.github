import { motion } from 'framer-motion';
import { Check, ExternalLink, Loader2, MessageSquare, Bell, Code } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { HelpTooltip } from '@/components/ui/help-tooltip';
import { useState } from 'react';
import { toast } from 'sonner';
import { useOnboardingStore, type IntegrationConfig } from '@/stores/onboardingStore';
import { isIntegrationAvailable, type PlanTier } from '@/lib/plan-gating';

interface Integration {
  id: string;
  name: string;
  description: string;
  icon: typeof MessageSquare;
  color: string;
  category: 'notifications' | 'monitoring' | 'cicd';
}

const INTEGRATIONS: Integration[] = [
  {
    id: 'slack',
    name: 'Slack',
    description: 'Get deployment notifications and alerts in your Slack workspace',
    icon: MessageSquare,
    color: '#4A154B',
    category: 'notifications',
  },
  {
    id: 'discord',
    name: 'Discord',
    description: 'Receive real-time notifications in your Discord server',
    icon: Bell,
    color: '#5865F2',
    category: 'notifications',
  },
  {
    id: 'github',
    name: 'GitHub',
    description: 'Integrate with GitHub Actions for automated deployments',
    icon: Code,
    color: '#333333',
    category: 'cicd',
  },
];

export function IntegrationsStep() {
  const { connectedIntegrations = [], addConnectedIntegration, removeConnectedIntegration, selectedPlan = 'free' } =
    useOnboardingStore();

  const [connecting, setConnecting] = useState<string | null>(null);

  const handleConnect = async (integrationId: string) => {
    if (!isIntegrationAvailable(integrationId as any, selectedPlan as PlanTier)) {
      toast.error('This integration requires a Professional plan or higher');
      return;
    }

    setConnecting(integrationId);

    // Simulate OAuth flow
    await new Promise((resolve) => setTimeout(resolve, 1500));

    const integration: IntegrationConfig = {
      id: `${integrationId}_${Date.now()}`,
      type: integrationId as IntegrationConfig['type'],
      connectedAt: new Date().toISOString(),
    };

    addConnectedIntegration(integration);
    setConnecting(null);
    toast.success(`${integrationId} connected successfully!`);
  };

  const handleDisconnect = (integrationId: string) => {
    removeConnectedIntegration(integrationId);
    toast.info('Integration disconnected');
  };

  const isConnected = (integrationId: string) =>
    connectedIntegrations.some((i) => i.type === integrationId);

  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center space-y-4"
      >
        <div className="onboarding-step-icon w-16 h-16 rounded-2xl flex items-center justify-center mx-auto bg-indigo-500/20">
          <Bell className="w-8 h-8 text-indigo-400" />
        </div>
        <p className="text-lg text-aviation-text-secondary font-mono max-w-xl mx-auto">
          Connect your favorite tools to receive notifications and integrate with your workflow.
        </p>
      </motion.div>

      <div className="grid gap-4">
        {INTEGRATIONS.map((integration, index) => {
          const connected = isConnected(integration.id);
          const available = isIntegrationAvailable(integration.id as any, selectedPlan as PlanTier);
          const Icon = integration.icon;
          const isConnecting = connecting === integration.id;

          return (
            <motion.div
              key={integration.id}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: index * 0.1 }}
            >
              <Card
                className={`onboarding-step-card p-4 transition-all ${
                  connected
                    ? 'border-aviation-green/50 bg-aviation-green/5'
                    : !available
                      ? 'opacity-60'
                      : ''
                }`}
              >
                <div className="flex items-center gap-4">
                  <div
                    className="w-12 h-12 rounded-lg flex items-center justify-center flex-shrink-0"
                    style={{ backgroundColor: `${integration.color}20` }}
                  >
                    <Icon className="w-6 h-6" style={{ color: integration.color }} />
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="font-mono font-semibold text-aviation-text-primary">
                        {integration.name}
                      </h3>
                      {!available && (
                        <span className="text-xs font-mono bg-aviation-amber/20 text-aviation-amber px-2 py-0.5 rounded">
                          PRO
                        </span>
                      )}
                      {connected && (
                        <span className="text-xs font-mono bg-aviation-green/20 text-aviation-green px-2 py-0.5 rounded">
                          CONNECTED
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-aviation-text-secondary font-mono">
                      {integration.description}
                    </p>
                  </div>

                  <div className="flex-shrink-0">
                    {connected ? (
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => handleDisconnect(connectedIntegrations.find((i) => i.type === integration.id)?.id || '')}
                        className="font-mono text-aviation-red border-aviation-red/50 hover:bg-aviation-red/10"
                      >
                        Disconnect
                      </Button>
                    ) : (
                      <Button
                        size="sm"
                        onClick={() => handleConnect(integration.id)}
                        disabled={!available || isConnecting}
                        className="font-mono"
                      >
                        {isConnecting ? (
                          <>
                            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                            Connecting...
                          </>
                        ) : available ? (
                          'Connect'
                        ) : (
                          <>
                            <ExternalLink className="w-4 h-4 mr-2" />
                            Upgrade
                          </>
                        )}
                      </Button>
                    )}
                  </div>
                </div>

                {connected && (
                  <motion.div
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: 'auto' }}
                    className="mt-4 pt-4 border-t border-aviation-border-panel"
                  >
                    <div className="flex items-center gap-2 text-aviation-green">
                      <Check className="w-4 h-4" />
                      <span className="text-sm font-mono">
                        {integration.name} is connected and ready to receive notifications
                      </span>
                    </div>
                  </motion.div>
                )}

                {!available && (
                  <motion.div
                    initial={{ opacity: 0, height: 0 }}
                    animate={{ opacity: 1, height: 'auto' }}
                    className="mt-4 pt-4 border-t border-aviation-border-panel"
                  >
                    <div className="flex items-start gap-2">
                      <HelpTooltip
                        content="Upgrade to Professional or higher to unlock this integration"
                        side="top"
                      />
                      <p className="text-sm text-aviation-text-muted font-mono">
                        Upgrade to Professional or higher to connect {integration.name}
                      </p>
                    </div>
                  </motion.div>
                )}
              </Card>
            </motion.div>
          );
        })}
      </div>

      <div className="bg-aviation-bg-tertiary rounded-lg p-4">
        <h4 className="font-mono font-medium text-aviation-text-primary mb-2">
          Integration Benefits
        </h4>
        <ul className="space-y-2 text-sm text-aviation-text-secondary font-mono">
          <li className="flex items-start gap-2">
            <Check className="w-4 h-4 text-aviation-green mt-0.5 flex-shrink-0" />
            Get instant alerts when deployments succeed or fail
          </li>
          <li className="flex items-start gap-2">
            <Check className="w-4 h-4 text-aviation-green mt-0.5 flex-shrink-0" />
            Monitor function health and performance in your preferred tool
          </li>
          <li className="flex items-start gap-2">
            <Check className="w-4 h-4 text-aviation-green mt-0.5 flex-shrink-0" />
            Automate deployments from your CI/CD pipeline
          </li>
        </ul>
      </div>

      <p className="text-center text-sm text-aviation-text-muted font-mono">
        You can always add more integrations later from your dashboard settings.
      </p>
    </div>
  );
}
