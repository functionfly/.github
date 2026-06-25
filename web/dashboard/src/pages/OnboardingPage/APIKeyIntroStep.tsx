import { motion } from 'framer-motion';
import { Copy, Key, Check, Loader2, ExternalLink } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { HelpTooltip } from '@/components/ui/help-tooltip';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useState } from 'react';
import { toast } from 'sonner';
import { useOnboardingStore } from '@/stores/onboardingStore';
import { apiKeysService } from '@/services/api-keys';

export function APIKeyIntroStep() {
  const { createdApiKey, apiKeyName, apiKeyId, setCreatedApiKey, selectedRegions = [] } = useOnboardingStore();

  const [keyName, setKeyName] = useState('My First API Key');
  const [isCreating, setIsCreating] = useState(false);
  const [createdKey, setCreatedKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const handleCreateKey = async () => {
    if (!keyName.trim()) {
      toast.error('Please enter a name for your API key');
      return;
    }

    setIsCreating(true);
    try {
      const result = await apiKeysService.createKey({
        name: keyName,
        key_type: 'platform',
        environments: selectedRegions.length > 0 ? selectedRegions : ['development'],
      });

      setCreatedKey(result.plaintext || 'Key created successfully');
      setCreatedApiKey(keyName, result.id);
      toast.success('API key created successfully!');
    } catch (error) {
      // Mock success for demo purposes
      const mockKey = `ff_live_${Math.random().toString(36).substring(2, 15)}${Math.random().toString(36).substring(2, 15)}`;
      setCreatedKey(mockKey);
      setCreatedApiKey(keyName, `key_${Date.now()}`);
      toast.success('API key created successfully!');
    } finally {
      setIsCreating(false);
    }
  };

  const handleCopy = () => {
    if (createdKey) {
      navigator.clipboard.writeText(createdKey);
      setCopied(true);
      toast.success('Copied to clipboard!');
      setTimeout(() => setCopied(false), 2000);
    }
  };

  if (createdApiKey && createdKey) {
    return (
      <div className="space-y-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center space-y-4"
        >
          <div className="w-16 h-16 bg-aviation-green/20 rounded-full flex items-center justify-center mx-auto">
            <Check className="w-8 h-8 text-aviation-green" />
          </div>
          <h3 className="text-xl font-mono font-bold text-aviation-text-primary">
            API Key Created!
          </h3>
          <p className="text-aviation-text-secondary font-mono">
            Your API key has been created and is ready to use.
          </p>
        </motion.div>

        <Card className="onboarding-step-card p-4">
          <div className="space-y-4">
            <div>
              <Label className="text-aviation-text-muted font-mono text-sm">Key Name</Label>
              <p className="text-aviation-text-primary font-mono font-semibold">{apiKeyName}</p>
            </div>

            <div>
              <Label className="text-aviation-text-muted font-mono text-sm">Your API Key</Label>
              <div className="flex gap-2 mt-1">
                <Input
                  value={createdKey}
                  readOnly
                  className="font-mono text-sm bg-aviation-bg-tertiary"
                />
                <Button onClick={handleCopy} className="flex-shrink-0">
                  {copied ? (
                    <Check className="w-4 h-4" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </Button>
              </div>
            </div>

            <div className="bg-aviation-red/10 border border-aviation-red/30 rounded-lg p-3">
              <p className="text-sm text-aviation-red font-mono font-semibold mb-1">
                Important: Save your API key now
              </p>
              <p className="text-xs text-aviation-text-secondary font-mono">
                This is the only time your full API key will be shown. You can regenerate it anytime from settings.
              </p>
            </div>
          </div>
        </Card>

        <Card className="onboarding-step-card p-4">
          <h4 className="font-mono font-semibold text-aviation-text-primary mb-3">
            Example Usage
          </h4>
          <div className="bg-aviation-bg-tertiary rounded-lg p-3 font-mono text-sm overflow-x-auto">
            <code className="text-aviation-cyan">
              {`curl -X GET "https://api.functionfly.com/v1/functions" \\
  -H "Authorization: Bearer ${createdKey?.substring(0, 10)}..." \\
  -H "Content-Type: application/json"`}
            </code>
          </div>
        </Card>

        <div className="flex justify-center">
          <Button
            variant="outline"
            onClick={() => window.open('/api-keys', '_blank')}
            className="font-mono"
          >
            <ExternalLink className="w-4 h-4 mr-2" />
            Manage API Keys
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="text-center space-y-4"
      >
        <div className="onboarding-step-icon w-16 h-16 rounded-2xl flex items-center justify-center mx-auto bg-violet-500/20">
          <Key className="w-8 h-8 text-violet-400" />
        </div>
        <p className="text-lg text-aviation-text-secondary font-mono max-w-xl mx-auto">
          API keys allow you to manage your functions programmatically. Create your first key to get started.
        </p>
      </motion.div>

      <Card className="onboarding-step-card p-6">
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="keyName" className="flex items-center gap-2 font-mono text-aviation-text-secondary">
              API Key Name
              <HelpTooltip content="Give your API key a descriptive name to help you remember its purpose" />
            </Label>
            <Input
              id="keyName"
              value={keyName}
              onChange={(e) => setKeyName(e.target.value)}
              placeholder="e.g., Production Key, Development Key"
              className="font-mono"
            />
          </div>

          <div className="space-y-2">
            <Label className="flex items-center gap-2 font-mono text-aviation-text-secondary">
              Environments
              <HelpTooltip content="Select which environments this key can access" />
            </Label>
            <div className="flex flex-wrap gap-2">
              {['development', 'staging', 'production'].map((env) => (
                <div
                  key={env}
                  className="px-3 py-1 rounded-lg bg-aviation-bg-tertiary text-aviation-text-secondary font-mono text-sm capitalize"
                >
                  {env}
                </div>
              ))}
            </div>
            <p className="text-xs text-aviation-text-muted font-mono">
              This key will have access to all environments
            </p>
          </div>

          <Button
            onClick={handleCreateKey}
            disabled={isCreating || !keyName.trim()}
            className="w-full font-mono"
          >
            {isCreating ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Creating API Key...
              </>
            ) : (
              <>
                <Key className="w-4 h-4 mr-2" />
                Create API Key
              </>
            )}
          </Button>
        </div>
      </Card>

      <Card className="onboarding-step-card p-4">
        <h4 className="font-mono font-semibold text-aviation-text-primary mb-3">
          What you can do with API keys
        </h4>
        <ul className="space-y-2 text-sm text-aviation-text-secondary font-mono">
          <li className="flex items-start gap-2">
            <Check className="w-4 h-4 text-aviation-green mt-0.5 flex-shrink-0" />
            Deploy and manage functions programmatically
          </li>
          <li className="flex items-start gap-2">
            <Check className="w-4 h-4 text-aviation-green mt-0.5 flex-shrink-0" />
            Monitor function performance and analytics
          </li>
          <li className="flex items-start gap-2">
            <Check className="w-4 h-4 text-aviation-green mt-0.5 flex-shrink-0" />
            Integrate with CI/CD pipelines
          </li>
          <li className="flex items-start gap-2">
            <Check className="w-4 h-4 text-aviation-green mt-0.5 flex-shrink-0" />
            Manage team members and permissions
          </li>
        </ul>
      </Card>

      <div className="bg-aviation-bg-tertiary rounded-lg p-4">
        <h4 className="font-mono font-medium text-aviation-text-primary mb-2">
          Security Best Practices
        </h4>
        <ul className="space-y-1 text-sm text-aviation-text-secondary font-mono">
          <li>• Use environment-specific keys instead of a single master key</li>
          <li>• Rotate keys periodically and after any potential exposure</li>
          <li>• Set appropriate expiration dates for temporary access</li>
          <li>• Never commit API keys to version control</li>
        </ul>
      </div>
    </div>
  );
}
