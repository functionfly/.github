import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { ProviderIcon } from '@/components/common/ProviderIcon';
import { AlertCircle, Check, Copy, ExternalLink, Loader2, Plus, Shield } from 'lucide-react';
import { useState } from 'react';
import type { ProviderConfig } from '../constants/providerMeta';

const AWS_REGIONS = [
  { value: 'us-east-1', label: 'US East (N. Virginia)' },
  { value: 'us-east-2', label: 'US East (Ohio)' },
  { value: 'us-west-1', label: 'US West (N. California)' },
  { value: 'us-west-2', label: 'US West (Oregon)' },
  { value: 'eu-west-1', label: 'Europe (Ireland)' },
  { value: 'eu-central-1', label: 'Europe (Frankfurt)' },
  { value: 'eu-west-2', label: 'Europe (London)' },
  { value: 'eu-west-3', label: 'Europe (Paris)' },
  { value: 'eu-north-1', label: 'Europe (Stockholm)' },
  { value: 'ap-southeast-1', label: 'Asia Pacific (Singapore)' },
  { value: 'ap-southeast-2', label: 'Asia Pacific (Sydney)' },
  { value: 'ap-northeast-1', label: 'Asia Pacific (Tokyo)' },
  { value: 'ap-south-1', label: 'Asia Pacific (Mumbai)' },
  { value: 'ca-central-1', label: 'Canada (Central)' },
  { value: 'sa-east-1', label: 'South America (São Paulo)' },
];

const IAM_POLICY = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "lambda:CreateFunction",
        "lambda:UpdateFunctionCode",
        "lambda:UpdateFunctionConfiguration",
        "lambda:GetFunction",
        "lambda:ListFunctions",
        "lambda:DeleteFunction",
        "lambda:InvokeFunction",
        "lambda:PublishVersion",
        "lambda:CreateAlias",
        "lambda:UpdateAlias",
        "lambda:GetAccountSettings",
        "lambda:CreateFunctionUrlConfig",
        "lambda:GetFunctionUrlConfig",
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "iam:PassRole"
      ],
      "Resource": "*"
    }
  ]
}`;

interface ConnectAWSDialogProps {
  provider: ProviderConfig;
  accent: { border: string; text: string };
  onConnect: (providerId: string, apiKey?: string) => Promise<void>;
}

export function ConnectAWSDialog({ provider, accent, onConnect }: ConnectAWSDialogProps) {
  const [accessKeyId, setAccessKeyId] = useState('');
  const [secretAccessKey, setSecretAccessKey] = useState('');
  const [region, setRegion] = useState('us-east-1');
  const [roleArn, setRoleArn] = useState('');
  const [isConnecting, setIsConnecting] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const [validationError, setValidationError] = useState<string | null>(null);
  const [showPolicy, setShowPolicy] = useState(false);
  const [copiedPolicy, setCopiedPolicy] = useState(false);

  const handleConnect = async () => {
    setValidationError(null);

    if (!accessKeyId.trim()) {
      setValidationError('Access Key ID is required');
      return;
    }
    if (!accessKeyId.startsWith('AKIA') || accessKeyId.length !== 20) {
      setValidationError('Access Key ID must start with AKIA and be 20 characters');
      return;
    }
    if (!secretAccessKey.trim()) {
      setValidationError('Secret Access Key is required');
      return;
    }
    if (secretAccessKey.length !== 40) {
      setValidationError('Secret Access Key must be 40 characters');
      return;
    }
    if (!region) {
      setValidationError('Region is required');
      return;
    }
    if (roleArn && !/^arn:aws:iam::\d{12}:role\/.+/.test(roleArn)) {
      setValidationError('Invalid IAM Role ARN format (expected: arn:aws:iam::<account-id>:role/<name>)');
      return;
    }

    const credentials = [accessKeyId.trim(), secretAccessKey.trim(), region, roleArn.trim()]
      .filter(Boolean)
      .join('|');

    setIsConnecting(true);
    try {
      await onConnect(provider.id, credentials);
      setIsOpen(false);
      resetForm();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to connect AWS Lambda';
      setValidationError(errorMessage);
    } finally {
      setIsConnecting(false);
    }
  };

  const resetForm = () => {
    setAccessKeyId('');
    setSecretAccessKey('');
    setRegion('us-east-1');
    setRoleArn('');
    setValidationError(null);
    setShowPolicy(false);
  };

  const handleClose = () => {
    setIsOpen(false);
    resetForm();
  };

  const copyPolicy = async () => {
    await navigator.clipboard.writeText(IAM_POLICY);
    setCopiedPolicy(true);
    setTimeout(() => setCopiedPolicy(false), 2000);
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && handleClose()}>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          className="w-full gap-2 border-border-default hover:border-border-subtle hover:bg-bg-secondary transition-all duration-200"
          onClick={() => setIsOpen(true)}
        >
          <Plus className="w-4 h-4" />
          Connect
        </Button>
      </DialogTrigger>
      <DialogContent className="bg-bg-tertiary border-border-subtle sm:max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <div className="flex items-center gap-3 mb-2">
            <div
              className="w-10 h-10 rounded-xl flex items-center justify-center"
              style={{ backgroundColor: `${accent.border}15` }}
            >
              <ProviderIcon provider={provider.id} size="md" />
            </div>
            <div>
              <DialogTitle className="text-text-primary text-lg">
                Connect AWS Lambda
              </DialogTitle>
            </div>
          </div>
          <DialogDescription className="text-text-secondary">
            Enter your AWS IAM credentials to connect Lambda. Credentials are encrypted with AES-256-GCM.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 pt-2">
          {validationError && (
            <div className="p-3 rounded-lg bg-error/10 border border-error/20 animate-in slide-in-from-top-1 duration-200">
              <div className="flex items-start gap-2">
                <AlertCircle className="w-4 h-4 text-error mt-0.5 shrink-0" />
                <p className="text-sm text-error">{validationError}</p>
              </div>
            </div>
          )}

          <div className="space-y-2">
            <Label htmlFor="aws-access-key" className="text-text-primary">
              Access Key ID
            </Label>
            <Input
              id="aws-access-key"
              type="text"
              placeholder="AKIAIOSFODNN7EXAMPLE"
              value={accessKeyId}
              onChange={(e) => {
                setAccessKeyId(e.target.value);
                if (validationError) setValidationError(null);
              }}
              className="bg-bg-secondary border-border-subtle focus:border-border-default font-mono text-sm"
              disabled={isConnecting}
              maxLength={20}
              autoComplete="off"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="aws-secret-key" className="text-text-primary">
              Secret Access Key
            </Label>
            <Input
              id="aws-secret-key"
              type="password"
              placeholder="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
              value={secretAccessKey}
              onChange={(e) => {
                setSecretAccessKey(e.target.value);
                if (validationError) setValidationError(null);
              }}
              className="bg-bg-secondary border-border-subtle focus:border-border-default font-mono text-sm"
              disabled={isConnecting}
              maxLength={40}
              autoComplete="off"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="aws-region" className="text-text-primary">
              Region
            </Label>
            <select
              id="aws-region"
              value={region}
              onChange={(e) => setRegion(e.target.value)}
              className="w-full px-3 py-2 rounded-md bg-bg-secondary border border-border-subtle text-text-primary text-sm focus:border-border-default focus:outline-none"
              disabled={isConnecting}
            >
              {AWS_REGIONS.map((r) => (
                <option key={r.value} value={r.value}>
                  {r.label}
                </option>
              ))}
            </select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="aws-role-arn" className="text-text-primary">
              Execution Role ARN <span className="text-text-tertiary font-normal">(optional)</span>
            </Label>
            <Input
              id="aws-role-arn"
              type="text"
              placeholder="arn:aws:iam::123456789012:role/lambda-execution-role"
              value={roleArn}
              onChange={(e) => {
                setRoleArn(e.target.value);
                if (validationError) setValidationError(null);
              }}
              className="bg-bg-secondary border-border-subtle focus:border-border-default font-mono text-sm"
              disabled={isConnecting}
              autoComplete="off"
            />
            <p className="text-xs text-text-tertiary">
              Required for deploying new functions. Leave empty if only invoking existing functions.
            </p>
          </div>

          <div className="p-3 rounded-lg bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800/50">
            <div className="flex items-start gap-2">
              <Shield className="w-4 h-4 text-amber-600 dark:text-amber-400 mt-0.5 shrink-0" />
              <div className="text-xs text-amber-950 dark:text-amber-100">
                <p className="font-medium mb-1">Minimum IAM permissions required:</p>
                <p className="text-amber-800 dark:text-amber-200">
                  Lambda full access, CloudWatch Logs, and IAM PassRole for the execution role.
                </p>
                <button
                  type="button"
                  onClick={() => setShowPolicy(!showPolicy)}
                  className="mt-2 text-amber-700 dark:text-amber-300 underline hover:no-underline"
                >
                  {showPolicy ? 'Hide' : 'View'} IAM policy
                </button>
                {showPolicy && (
                  <div className="mt-2 relative">
                    <pre className="p-2 rounded bg-amber-100 dark:bg-amber-900/50 text-[10px] leading-tight overflow-x-auto">
                      {IAM_POLICY}
                    </pre>
                    <button
                      type="button"
                      onClick={copyPolicy}
                      className="absolute top-1 right-1 p-1 rounded bg-amber-200 dark:bg-amber-800 hover:bg-amber-300 dark:hover:bg-amber-700"
                      title="Copy to clipboard"
                    >
                      {copiedPolicy ? (
                        <Check className="w-3 h-3 text-green-700" />
                      ) : (
                        <Copy className="w-3 h-3 text-amber-700 dark:text-amber-300" />
                      )}
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2 text-xs text-text-tertiary">
            <ExternalLink className="w-3 h-3" />
            <a
              href="https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html"
              target="_blank"
              rel="noopener noreferrer"
              className="hover:text-text-secondary underline"
            >
              How to create AWS access keys
            </a>
          </div>

          <button
            type="button"
            onClick={handleConnect}
            disabled={!accessKeyId.trim() || !secretAccessKey.trim() || isConnecting}
            className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-md font-medium transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed hover:opacity-90 active:scale-[0.98]"
            style={{
              backgroundColor: accent.border,
              color: 'white',
            }}
          >
            {isConnecting ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin" />
                Validating & Connecting...
              </>
            ) : (
              <>
                <Check className="w-4 h-4" />
                Connect AWS Lambda
              </>
            )}
          </button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
