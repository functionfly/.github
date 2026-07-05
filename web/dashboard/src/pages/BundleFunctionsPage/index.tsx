import { useSearchParams, useNavigate, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  Code,
  Webhook,
  Mail,
  Store,
  MessageSquare,
  Cpu,
  Database,
  ArrowRight,
  Loader2,
  ExternalLink,
  CheckCircle,
  Settings,
  Rocket,
  Shield,
  CreditCard,
  Brain,
  Sparkles,
  Zap,
  type LucideIcon,
} from 'lucide-react';
import { apiClient } from '@/api/client';
import { usePageTitle } from '@/hooks';
import { getBundleCatalog, type BundleCatalogItem } from '@/api/billing';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  SealedButton,
  FrameButton,
  StatusPill,
  AnnotationTag,
} from '@/components/containment';
import './styles.css';

const ICON_MAP: Record<string, LucideIcon> = {
  Rocket,
  Shield,
  CreditCard,
  Brain,
  Sparkles,
  Zap,
  Code,
  Webhook,
  Mail,
  Store,
  MessageSquare,
  Cpu,
  Database,
  CheckCircle,
  Settings,
  ExternalLink,
  ArrowRight,
  Loader2,
};

function getIcon(name: string): LucideIcon {
  return ICON_MAP[name] || Code;
}

export default function BundleFunctionsPage() {
  usePageTitle('Bundle Functions');
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const bundleSlug = searchParams.get('bundle') || 'saas-starter';

  const { data: catalogData } = useQuery({
    queryKey: ['bundle-catalog'],
    queryFn: getBundleCatalog,
  });

  const bundle: BundleCatalogItem | undefined = catalogData?.bundles.find((b) => b.slug === bundleSlug);
  const Icon = bundle ? getIcon(bundle.icon) : Rocket;

  const bundleFunctionNames = bundle ? Object.keys(bundle.functions) : [];

  const { data: bundleFunctions, isLoading } = useQuery({
    queryKey: ['bundle-registry-functions', bundleSlug],
    queryFn: async () => {
      const results: Record<string, { id: string; status: string; region: string }> = {};
      for (const fnName of bundleFunctionNames) {
        try {
          const fn = await apiClient.get<any>(`/v1/functions/functionfly/${fnName}`);
          if (fn?.id) {
            results[fnName] = { id: fn.id, status: fn.status || 'active', region: fn.region || 'us-east-1' };
          }
        } catch {
          // function not published yet
        }
      }
      return results;
    },
    refetchInterval: (query) => {
      const data = query.state.data;
      if (!data) return false;
      return Object.keys(data).length < bundleFunctionNames.length ? 5000 : false;
    },
  });

  const resolvedFunctions = bundleFunctions ?? {};
  const resolvedCount = Object.keys(resolvedFunctions).length;
  const allPresent = bundleFunctionNames.length === resolvedCount;

  return (
    <div className="bp-page">
      <PageGrid />

      {/* Header */}
      <Chamber>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="BUNDLE FUNCTIONS" secondary={bundle.name} />
        <div className="sc-bundle-fns__header">
          <div className="sc-bundle-fns__header-icon">
            <Icon size={28} />
          </div>
          <div className="sc-bundle-fns__header-content">
            <h1 className="sc-bundle-fns__title">Explore your functions</h1>
            <p className="sc-bundle-fns__tagline">
              Your bundle came with pre-configured function templates. Customize them for your use case.
            </p>
          </div>
        </div>

        {/* Stepper */}
        <div className="sc-bundle-fns__stepper">
          {bundle.provisioning_steps.map((step, i) => (
            <div
              key={step.label}
              className={`sc-bundle-fns__step ${i === 0 ? 'sc-bundle-fns__step--active' : ''}`}
            >
              <div className="sc-bundle-fns__step-num">{i + 1}</div>
              <div className="sc-bundle-fns__step-info">
                <span className="sc-bundle-fns__step-label">{step.label}</span>
                <span className="sc-bundle-fns__step-desc">{step.description}</span>
              </div>
              {i < bundle.provisioning_steps.length - 1 && (
                <div className="sc-bundle-fns__step-line" />
              )}
            </div>
          ))}
        </div>
      </Chamber>

      {/* Function Cards */}
      {isLoading ? (
        <Chamber nested className="sc-bundle-fns__loading">
          <Loader2 size={20} className="sc-community-spinner" />
          <span className="sc-bundle-fns__loading-text">Loading your functions...</span>
        </Chamber>
      ) : (
        <div className="sc-bundle-fns__grid">
          {bundleFunctionNames.map((fnName) => {
            const meta = bundle.functions[fnName];
            const fn = resolvedFunctions[fnName];
            const FnIcon = getIcon(meta.icon);

            return (
              <Chamber key={fnName}>
                <CornerBrace position="tl" />
                <CornerBrace position="br" />
                <div className="sc-bundle-fns__card">
                  <div className="sc-bundle-fns__card-header">
                    <div className="sc-bundle-fns__card-icon">
                      <FnIcon size={20} />
                    </div>
                    <div className="sc-bundle-fns__card-info">
                      <span className="sc-bundle-fns__card-name">{fnName}</span>
                      {fn ? (
                        <StatusPill
                          status={fn.status === 'active' || fn.status === 'deployed' ? 'live' : 'pending'}
                          label={fn.status}
                        />
                      ) : (
                        <StatusPill status="pending" label="provisioning" />
                      )}
                    </div>
                  </div>

                  <p className="sc-bundle-fns__card-desc">{meta.description}</p>

                  <div className="sc-bundle-fns__card-caps">
                    {meta.capabilities.map((cap) => (
                      <span key={cap} className="sc-bundle-fns__cap">{cap}</span>
                    ))}
                    {fn?.region && (
                      <span className="sc-bundle-fns__cap sc-bundle-fns__cap--region">{fn.region}</span>
                    )}
                  </div>

                  <div className="sc-bundle-fns__card-actions">
                    {fn ? (
                      <Link to={`/fx/functionfly/${fnName}`}>
                        <SealedButton size="sm" iconLeft={<Code size={14} />}>
                          Customize
                        </SealedButton>
                      </Link>
                    ) : (
                      <SealedButton size="sm" disabled iconLeft={<Loader2 size={14} />}>
                        Provisioning...
                      </SealedButton>
                    )}
                  </div>
                </div>
              </Chamber>
            );
          })}
        </div>
      )}

      {/* Continue */}
      <Chamber nested className="sc-bundle-fns__continue">
        <div className="sc-bundle-fns__continue-inner">
          <div className="sc-bundle-fns__continue-info">
            {allPresent ? (
              <>
                <CheckCircle size={18} className="sc-bundle-fns__continue-icon--ok" />
                <div>
                  <span className="sc-bundle-fns__continue-title">All functions provisioned</span>
                  <span className="sc-bundle-fns__continue-desc">
                    {resolvedCount} function{resolvedCount !== 1 ? 's' : ''} ready to customize
                  </span>
                </div>
              </>
            ) : (
              <>
                <Loader2 size={18} className="sc-community-spinner" />
                <div>
                  <span className="sc-bundle-fns__continue-title">Provisioning functions...</span>
                  <span className="sc-bundle-fns__continue-desc">
                    {resolvedCount} of {bundleFunctionNames.length} ready
                  </span>
                </div>
              </>
            )}
          </div>
          <div className="sc-bundle-fns__continue-actions">
            <FrameButton
              size="sm"
              onClick={() => navigate(`/bundles/overview?bundle=${bundleSlug}`)}
            >
              Back to Overview
            </FrameButton>
            <SealedButton
              size="sm"
              iconRight={<ArrowRight size={14} />}
              onClick={() => navigate(`/bundles/integrations?bundle=${bundleSlug}`)}
            >
              Continue to Integrations
            </SealedButton>
          </div>
        </div>
      </Chamber>
    </div>
  );
}
