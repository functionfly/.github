import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Database, FileText, Key, Layers, Mail, Shield, Zap } from 'lucide-react';

interface EnvPreset {
  id: string;
  name: string;
  description: string;
  icon: React.ReactNode;
  category: 'database' | 'api' | 'security' | 'email' | 'storage';
  variables: Array<{
    key: string;
    value: string;
    isSecret: boolean;
    description?: string;
  }>;
}

const ENV_PRESETS: EnvPreset[] = [
  {
    id: 'database',
    name: 'Database Connection',
    description: 'Common database connection variables',
    icon: <Database className="w-4 h-4" />,
    category: 'database',
    variables: [
      { key: 'DATABASE_URL', value: '', isSecret: true, description: 'Full database connection string' },
      { key: 'DB_HOST', value: 'localhost', isSecret: false },
      { key: 'DB_PORT', value: '5432', isSecret: false },
      { key: 'DB_NAME', value: '', isSecret: false },
      { key: 'DB_USER', value: '', isSecret: false },
      { key: 'DB_PASSWORD', value: '', isSecret: true },
      { key: 'DB_SSL', value: 'true', isSecret: false },
    ],
  },
  {
    id: 'redis',
    name: 'Redis Cache',
    description: 'Redis connection configuration',
    icon: <Layers className="w-4 h-4" />,
    category: 'database',
    variables: [
      { key: 'REDIS_URL', value: 'redis://localhost:6379', isSecret: true },
      { key: 'REDIS_HOST', value: 'localhost', isSecret: false },
      { key: 'REDIS_PORT', value: '6379', isSecret: false },
      { key: 'REDIS_PASSWORD', value: '', isSecret: true },
      { key: 'REDIS_DB', value: '0', isSecret: false },
    ],
  },
  {
    id: 'stripe',
    name: 'Stripe Payments',
    description: 'Stripe API keys for payments',
    icon: <Key className="w-4 h-4" />,
    category: 'api',
    variables: [
      { key: 'STRIPE_SECRET_KEY', value: 'sk_test_', isSecret: true, description: 'Stripe secret key (starts with sk_test_ or sk_live_)' },
      { key: 'STRIPE_PUBLISHABLE_KEY', value: 'pk_test_', isSecret: false },
      { key: 'STRIPE_WEBHOOK_SECRET', value: 'whsec_', isSecret: true },
      { key: 'STRIPE_API_VERSION', value: '2023-10-16', isSecret: false },
    ],
  },
  {
    id: 'aws',
    name: 'AWS Services',
    description: 'AWS access keys and configuration',
    icon: <Zap className="w-4 h-4" />,
    category: 'api',
    variables: [
      { key: 'AWS_ACCESS_KEY_ID', value: '', isSecret: true },
      { key: 'AWS_SECRET_ACCESS_KEY', value: '', isSecret: true },
      { key: 'AWS_REGION', value: 'us-east-1', isSecret: false },
      { key: 'AWS_BUCKET_NAME', value: '', isSecret: false },
    ],
  },
  {
    id: 'sendgrid',
    name: 'SendGrid Email',
    description: 'Email sending configuration',
    icon: <Mail className="w-4 h-4" />,
    category: 'email',
    variables: [
      { key: 'SENDGRID_API_KEY', value: 'SG.', isSecret: true },
      { key: 'FROM_EMAIL', value: '', isSecret: false },
      { key: 'FROM_NAME', value: '', isSecret: false },
    ],
  },
  {
    id: 'jwt',
    name: 'JWT Auth',
    description: 'JSON Web Token configuration',
    icon: <Shield className="w-4 h-4" />,
    category: 'security',
    variables: [
      { key: 'JWT_SECRET', value: '', isSecret: true, description: 'Secret key for signing JWTs' },
      { key: 'JWT_EXPIRY', value: '3600', isSecret: false, description: 'Token expiry in seconds' },
      { key: 'JWT_ISSUER', value: '', isSecret: false },
      { key: 'JWT_AUDIENCE', value: '', isSecret: false },
    ],
  },
  {
    id: 'oauth',
    name: 'OAuth Providers',
    description: 'Social login configuration',
    icon: <Shield className="w-4 h-4" />,
    category: 'security',
    variables: [
      { key: 'GOOGLE_CLIENT_ID', value: '', isSecret: false },
      { key: 'GOOGLE_CLIENT_SECRET', value: '', isSecret: true },
      { key: 'GITHUB_CLIENT_ID', value: '', isSecret: false },
      { key: 'GITHUB_CLIENT_SECRET', value: '', isSecret: true },
      { key: 'OAUTH_REDIRECT_URL', value: '', isSecret: false },
    ],
  },
  {
    id: 's3-storage',
    name: 'S3-Compatible Storage',
    description: 'R2, MinIO, or other S3-compatible storage',
    icon: <FileText className="w-4 h-4" />,
    category: 'storage',
    variables: [
      { key: 'S3_ENDPOINT', value: '', isSecret: false, description: 'e.g., https://<account>.r2.cloudflarestorage.com' },
      { key: 'S3_ACCESS_KEY_ID', value: '', isSecret: true },
      { key: 'S3_SECRET_ACCESS_KEY', value: '', isSecret: true },
      { key: 'S3_BUCKET', value: '', isSecret: false },
      { key: 'S3_REGION', value: 'auto', isSecret: false },
    ],
  },
];

const CATEGORIES_KEYS = [
  { id: 'all', labelKey: 'funcEditor.allCategory' as const },
  { id: 'database', labelKey: 'funcEditor.databaseCategory' as const },
  { id: 'api', labelKey: 'funcEditor.apisCategory' as const },
  { id: 'security', labelKey: 'funcEditor.securityCategory' as const },
  { id: 'email', labelKey: 'funcEditor.emailCategory' as const },
  { id: 'storage', labelKey: 'funcEditor.storageCategory' as const },
];

interface EnvPresetsPickerProps {
  onSelect: (preset: EnvPreset) => void;
  children: React.ReactNode;
}

import { Button } from '@/components/ui/button';
import { ScrollArea } from '@/components/ui/scroll-area';
import { useState } from 'react';
import { useTranslation } from 'react-i18next';

export function EnvPresetsPicker({ onSelect, children }: EnvPresetsPickerProps) {
  const { t } = useTranslation();
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [open, setOpen] = useState(false);

  const filteredPresets =
    selectedCategory === 'all'
      ? ENV_PRESETS
      : ENV_PRESETS.filter((p) => p.category === selectedCategory);

  const handleSelect = (preset: EnvPreset) => {
    onSelect(preset);
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent
        className="sm:max-w-2xl max-h-[80vh]"
        style={{ background: 'var(--bg-secondary)' }}
      >
        <DialogHeader>
            <DialogTitle className="flex items-center gap-2 text-base font-display">
              <Layers className="w-5 h-5 text-[#FF6B35]" />
              {t('funcEditor.envPresets')}
            </DialogTitle>
        </DialogHeader>

        {/* Category filters */}
        <div className="flex flex-wrap gap-2 mt-2">
          {CATEGORIES_KEYS.map((cat) => (
            <button
              key={cat.id}
              onClick={() => setSelectedCategory(cat.id)}
              className={`px-3 py-1.5 rounded-full text-xs transition-colors ${
                selectedCategory === cat.id
                  ? 'bg-[#FF6B35] text-white'
                  : 'bg-bg-tertiary text-text-secondary hover:bg-bg-hover'
              }`}
            >
              {t(cat.labelKey)}
            </button>
          ))}
        </div>

        {/* Presets list */}
        <ScrollArea className="h-[400px] mt-2">
          <div className="space-y-2 pr-4">
            {filteredPresets.map((preset) => (
              <div
                key={preset.id}
                className="flex items-start gap-3 p-4 rounded-lg border border-border-subtle/30 bg-bg-tertiary hover:border-border-default transition-colors"
              >
                <div className="p-2 rounded-lg bg-bg-secondary text-[#FF6B35]">{preset.icon}</div>
                <div className="flex-1 min-w-0">
                  <h4 className="font-medium text-sm text-text-primary">{preset.name}</h4>
                  <p className="text-xs text-text-muted mt-0.5">{preset.description}</p>

                  {/* Variable preview */}
                  <div className="flex flex-wrap gap-1 mt-2">
                    {preset.variables.slice(0, 4).map((v) => (
                      <span
                        key={v.key}
                        className={`text-[10px] px-1.5 py-0.5 rounded ${
                          v.isSecret
                            ? 'bg-amber-500/10 text-amber-500'
                            : 'bg-bg-secondary text-text-muted'
                        }`}
                      >
                        {v.isSecret ? '🔒' : ''}
                        {v.key}
                      </span>
                    ))}
                    {preset.variables.length > 4 && (
                      <span className="text-[10px] px-1.5 py-0.5 rounded bg-bg-secondary text-text-muted">
                        +{preset.variables.length - 4}
                      </span>
                    )}
                  </div>
                </div>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => handleSelect(preset)}
                  className="shrink-0"
                >
                  {t('funcEditor.add')}
                </Button>
              </div>
            ))}
          </div>
        </ScrollArea>

        <p className="text-xs text-text-muted mt-2">
          {t('funcEditor.secretsMarkedInfo')}
        </p>
      </DialogContent>
    </Dialog>
  );
}

export type { EnvPreset, EnvPresetsPickerProps };
export { ENV_PRESETS };
