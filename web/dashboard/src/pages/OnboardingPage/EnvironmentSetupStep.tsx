import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { HelpTooltip } from '@/components/ui/help-tooltip';
import { Input } from '@/components/ui/input';
import { useOnboardingStore, type Environment } from '@/stores/onboardingStore';
import { motion } from 'framer-motion';
import { Check, Code, Globe, Plus, Server } from 'lucide-react';
import { useState } from 'react';
import { toast } from 'sonner';

const DEFAULT_ENVIRONMENTS: Array<{
  id: Environment;
  name: string;
  description: string;
  icon: typeof Server;
  color: string;
}> = [
  {
    id: 'development',
    name: 'Development',
    description: 'For local development and testing',
    icon: Code,
    color: '#10b981',
  },
  {
    id: 'staging',
    name: 'Staging',
    description: 'For pre-production testing',
    icon: Globe,
    color: '#f59e0b',
  },
  {
    id: 'production',
    name: 'Production',
    description: 'Live environment for end users',
    icon: Server,
    color: '#ef4444',
  },
];

export function EnvironmentSetupStep() {
  const {
    environments = ['development', 'staging', 'production'],
    activeEnvironment = 'development',
    customEnvironments = [],
    setEnvironments,
    addCustomEnvironment,
    selectedPlan,
  } = useOnboardingStore();

  const [showAddCustom, setShowAddCustom] = useState(false);
  const [customEnvName, setCustomEnvName] = useState('');

  const allEnvironments = [
    ...DEFAULT_ENVIRONMENTS,
    ...customEnvironments.map((name) => ({
      id: name as Environment,
      name,
      description: 'Custom environment',
      icon: Server,
      color: '#8b5cf6',
    })),
  ];

  const canAddCustom = selectedPlan !== 'free';

  const handleSelect = (envId: Environment) => {
    setEnvironments([...environments.filter((e) => e !== envId), envId]);
  };

  const handleAddCustom = () => {
    const name = customEnvName.trim().toLowerCase().replace(/\s+/g, '-');
    if (!name) {
      toast.error('Please enter a valid environment name');
      return;
    }
    if (allEnvironments.some((e) => e.name.toLowerCase() === name)) {
      toast.error('An environment with this name already exists');
      return;
    }
    addCustomEnvironment(name);
    setCustomEnvName('');
    setShowAddCustom(false);
    toast.success(`Environment "${name}" created`);
  };

  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center space-y-4"
      >
        <div className="onboarding-step-icon w-16 h-16 rounded-2xl flex items-center justify-center mx-auto bg-emerald-500/20">
          <Server className="w-8 h-8 text-emerald-400" />
        </div>
        <p className="text-lg text-aviation-text-secondary font-mono max-w-xl mx-auto">
          Set up your deployment environments. Each environment isolates your functions and
          configuration.
        </p>
      </motion.div>

      <div className="grid gap-3">
        {allEnvironments.map((env, index) => {
          const isActive = activeEnvironment === env.id;
          const isDefault = DEFAULT_ENVIRONMENTS.some((e) => e.id === env.id);
          const Icon = env.icon;

          return (
            <motion.div
              key={env.id}
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: index * 0.1 }}
            >
              <Card
                className={`onboarding-step-card p-4 cursor-pointer transition-all ${
                  isActive
                    ? 'border-aviation-amber shadow-lg shadow-aviation-amber/20'
                    : 'hover:border-aviation-amber/50'
                }`}
                onClick={() => handleSelect(env.id)}
              >
                <div className="flex items-center gap-4">
                  <div
                    className="w-12 h-12 rounded-lg flex items-center justify-center flex-shrink-0"
                    style={{ backgroundColor: `${env.color}20` }}
                  >
                    <Icon className="w-6 h-6" style={{ color: env.color }} />
                  </div>

                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <h3 className="font-mono font-semibold text-aviation-text-primary">
                        {env.name}
                      </h3>
                      {isDefault && (
                        <span className="text-xs font-mono bg-aviation-bg-tertiary text-aviation-text-muted px-2 py-0.5 rounded">
                          DEFAULT
                        </span>
                      )}
                      {isActive && (
                        <span className="text-xs font-mono bg-aviation-amber/20 text-aviation-amber px-2 py-0.5 rounded">
                          ACTIVE
                        </span>
                      )}
                    </div>
                    <p className="text-sm text-aviation-text-secondary font-mono">
                      {env.description}
                    </p>
                  </div>

                  <div
                    className={`w-6 h-6 rounded-full border-2 flex items-center justify-center flex-shrink-0 ${
                      isActive
                        ? 'border-aviation-amber bg-aviation-amber'
                        : 'border-aviation-border-panel'
                    }`}
                  >
                    {isActive && <Check className="w-4 h-4 text-aviation-bg-primary" />}
                  </div>
                </div>
              </Card>
            </motion.div>
          );
        })}
      </div>

      {canAddCustom ? (
        <div className="space-y-3">
          {showAddCustom ? (
            <motion.div
              initial={{ opacity: 0, height: 0 }}
              animate={{ opacity: 1, height: 'auto' }}
              className="flex gap-3 items-end"
            >
              <div className="flex-1 space-y-1">
                <label className="text-sm font-mono text-aviation-text-secondary">
                  Custom Environment Name
                </label>
                <Input
                  placeholder="e.g., qa, uat, staging-2"
                  value={customEnvName}
                  onChange={(e) => setCustomEnvName(e.target.value)}
                  className="font-mono"
                  onKeyDown={(e) => e.key === 'Enter' && handleAddCustom()}
                />
              </div>
              <Button onClick={handleAddCustom} className="font-mono">
                <Plus className="w-4 h-4 mr-2" />
                Add
              </Button>
              <Button
                variant="ghost"
                onClick={() => {
                  setShowAddCustom(false);
                  setCustomEnvName('');
                }}
                className="font-mono"
              >
                Cancel
              </Button>
            </motion.div>
          ) : (
            <Button
              variant="outline"
              onClick={() => setShowAddCustom(true)}
              className="w-full font-mono"
            >
              <Plus className="w-4 h-4 mr-2" />
              Add Custom Environment
            </Button>
          )}
        </div>
      ) : (
        <div className="bg-aviation-bg-tertiary rounded-lg p-4">
          <div className="flex items-start gap-3">
            <HelpTooltip
              content="Custom environments allow you to create additional stages like QA, UAT, or stage-specific setups. Upgrade to Starter or higher to unlock this feature."
              side="top"
            />
            <div>
              <h4 className="font-mono font-medium text-aviation-text-primary mb-1">
                Need More Environments?
              </h4>
              <p className="text-sm text-aviation-text-secondary font-mono">
                Upgrade to Starter or higher to create custom environments like QA, UAT, or regional
                stages.
              </p>
            </div>
          </div>
        </div>
      )}

      <div className="bg-aviation-bg-tertiary rounded-lg p-4">
        <h4 className="font-mono font-medium text-aviation-text-primary mb-2">Environment Tips</h4>
        <ul className="space-y-2 text-sm text-aviation-text-secondary font-mono">
          <li>• Development: Use for local testing and iteration</li>
          <li>• Staging: Mirror production configuration for final testing</li>
          <li>• Production: Live environment serving your users</li>
        </ul>
      </div>
    </div>
  );
}
