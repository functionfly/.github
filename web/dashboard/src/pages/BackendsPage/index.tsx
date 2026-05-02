import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  Boxes,
  Globe,
  MapPin,
  Check,
  ChevronRight,
  Zap,
  Shield,
  Rocket,
  Cloud,
  Code,
  ExternalLink,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { BACKEND_PROVIDERS } from '@/types';
import type { BackendTemplate } from '@/types';

const PROVIDER_ICONS: Record<string, React.ReactNode> = {
  'functionfly-edge': (
    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-brand-500/20 to-purple-500/10 border border-brand-500/20 flex items-center justify-center">
      <span className="text-2xl">⬡</span>
    </div>
  ),
  workers: (
    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-orange-500/20 to-yellow-500/10 border border-orange-500/20 flex items-center justify-center">
      <span className="text-2xl">☁</span>
    </div>
  ),
  vercel: (
    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-gray-500/20 to-gray-400/10 border border-gray-500/20 flex items-center justify-center">
      <span className="text-2xl">▲</span>
    </div>
  ),
  fly: (
    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-blue-500/20 to-cyan-500/10 border border-blue-500/20 flex items-center justify-center">
      <span className="text-2xl">✈</span>
    </div>
  ),
  'deno-deploy': (
    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-teal-500/20 to-green-500/10 border border-teal-500/20 flex items-center justify-center">
      <span className="text-2xl">  </span>
    </div>
  ),
};

const CAPABILITY_COLORS: Record<string, string> = {
  http: 'bg-green-500/10 text-green-400 border-green-500/20',
  websocket: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
  grpc: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
  queue: 'bg-yellow-500/10 text-yellow-400 border-yellow-500/20',
  kv: 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20',
  tcp: 'bg-orange-500/10 text-orange-400 border-orange-500/20',
  serverless: 'bg-pink-500/10 text-pink-400 border-pink-500/20',
};

function ProviderCard({ template, index }: { template: BackendTemplate; index: number }) {
  const icon = PROVIDER_ICONS[template.provider] || (
    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-gray-500/20 to-gray-400/10 border border-gray-500/20 flex items-center justify-center">
      <Globe className="w-6 h-6 text-gray-400" />
    </div>
  );

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: index * 0.1 }}
    >
      <Card className="h-full overflow-hidden hover:border-brand-500/30 transition-all duration-300 group">
        <CardHeader className="pb-4">
          <div className="flex items-start justify-between">
            <div className="flex items-center gap-4">
              {icon}
              <div>
                <CardTitle className="text-lg font-semibold text-text-primary">
                  {template.name}
                </CardTitle>
                <p className="text-sm text-text-secondary mt-1 line-clamp-2">
                  {template.description}
                </p>
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <div className="flex items-center gap-2 mb-2">
              <Rocket className="w-4 h-4 text-brand-400" />
              <span className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Features
              </span>
            </div>
            <div className="flex flex-wrap gap-2">
              {template.features.map((feature) => (
                <Badge
                  key={feature}
                  variant="secondary"
                  className="text-xs bg-brand-500/5 text-brand-400 border-brand-500/10"
                >
                  <Check className="w-3 h-3 mr-1" />
                  {feature}
                </Badge>
              ))}
            </div>
          </div>

          <div>
            <div className="flex items-center gap-2 mb-2">
              <MapPin className="w-4 h-4 text-brand-400" />
              <span className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Regions ({template.supportedRegions.length})
              </span>
            </div>
            <div className="flex flex-wrap gap-1">
              {template.supportedRegions.slice(0, 6).map((region) => (
                <Badge
                  key={region}
                  variant="outline"
                  className="text-xs font-mono"
                >
                  {region}
                </Badge>
              ))}
              {template.supportedRegions.length > 6 && (
                <Badge variant="outline" className="text-xs">
                  +{template.supportedRegions.length - 6} more
                </Badge>
              )}
            </div>
          </div>

          <div>
            <div className="flex items-center gap-2 mb-2">
              <Zap className="w-4 h-4 text-brand-400" />
              <span className="text-xs font-medium text-text-muted uppercase tracking-wide">
                Capabilities
              </span>
            </div>
            <div className="flex flex-wrap gap-2">
              {template.capabilities.map((cap) => (
                <span
                  key={cap}
                  className={cn(
                    'px-2 py-1 rounded-md text-xs font-medium border uppercase tracking-wide',
                    CAPABILITY_COLORS[cap] || 'bg-gray-500/10 text-gray-400 border-gray-500/20'
                  )}
                >
                  {cap}
                </span>
              ))}
            </div>
          </div>

          <div className="pt-4 border-t border-border/50 flex items-center justify-between">
            <span className="text-xs text-text-muted">
              {template.supportedRegions.length} regions available
            </span>
            <Button
              variant="ghost"
              size="sm"
              className="gap-1 text-brand-400 hover:text-brand-300 hover:bg-brand-500/5 opacity-0 group-hover:opacity-100 transition-opacity"
              asChild
            >
              <Link to={`/apps?addBackend=${template.provider}`}>
                Deploy
                <ChevronRight className="w-4 h-4" />
              </Link>
            </Button>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

export function BackendsPage() {
  const { t } = useTranslation();

  return (
    <div className="min-h-screen bg-bg-primary relative overflow-hidden">
      <div className="absolute inset-0 opacity-30">
        <div className="absolute top-1/4 left-1/4 w-96 h-96 bg-brand-500/20 rounded-full blur-3xl" />
        <div className="absolute bottom-1/4 right-1/4 w-96 h-96 bg-info/20 rounded-full blur-3xl" />
      </div>

      <div className="relative z-10 max-w-7xl mx-auto px-4 py-12">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="mb-12"
        >
          <div className="flex items-center gap-3 mb-4">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-brand-500/20 to-purple-500/10 border border-brand-500/20 flex items-center justify-center">
              <Boxes className="w-5 h-5 text-brand-400" />
            </div>
            <Badge variant="outline" className="bg-brand-500/5 text-brand-400 border-brand-500/20">
              Available Providers
            </Badge>
          </div>
          <h1 className="text-4xl font-bold text-text-primary mb-4">
            Backend <span className="text-brand-400">Providers</span>
          </h1>
          <p className="text-lg text-text-secondary max-w-2xl">
            Choose from multiple deployment providers to host your serverless functions.
            Each provider offers unique features, regions, and capabilities.
          </p>
        </motion.div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {BACKEND_PROVIDERS.map((provider, index) => (
            <ProviderCard key={provider.provider} template={provider} index={index} />
          ))}
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.5 }}
          className="mt-16 p-8 rounded-2xl bg-gradient-to-br from-brand-500/5 via-purple-500/5 to-transparent border border-brand-500/10"
        >
          <div className="flex flex-col md:flex-row items-start md:items-center gap-6">
            <div className="flex-1">
              <h3 className="text-xl font-semibold text-text-primary mb-2">
                Add Your Own Provider
              </h3>
              <p className="text-text-secondary">
                Don't see your preferred provider? Our platform supports custom backend configurations.
                Contact us to add your infrastructure to the list.
              </p>
            </div>
            <Button variant="outline" className="gap-2" asChild>
              <Link to="/enterprise">
                Contact Sales
                <ExternalLink className="w-4 h-4" />
              </Link>
            </Button>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.6 }}
          className="mt-8 text-center"
        >
          <div className="flex items-center justify-center gap-6 text-sm text-text-muted">
            <div className="flex items-center gap-2">
              <Shield className="w-4 h-4" />
              <span>Enterprise-grade security</span>
            </div>
            <div className="flex items-center gap-2">
              <Cloud className="w-4 h-4" />
              <span>99.9% uptime SLA</span>
            </div>
            <div className="flex items-center gap-2">
              <Code className="w-4 h-4" />
              <span>Unified API</span>
            </div>
          </div>
        </motion.div>
      </div>
    </div>
  );
}

export default BackendsPage;