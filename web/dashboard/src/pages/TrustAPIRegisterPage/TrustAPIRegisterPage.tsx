import {
  createTrustAPIPartner,
  createTrustAPICheckout,
  getTrustAPITierPricing,
  formatCurrency,
  getTrustAPIErrorMessage,
  type TrustAPIPartnerCreateRequest,
  type TrustAPITierPricing,
} from '@/api/trustapi';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { usePageTitle } from '@/hooks/usePageTitle';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useAuthStore } from '@/stores/authStore';
import { ArrowLeft, CheckCircle2, CreditCard, Shield } from 'lucide-react';
import { useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

function slugify(value: string): string {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-|-$/g, '');
}

function tierLabel(tier: string): string {
  const labels: Record<string, string> = {
    developer: 'Developer',
    payg: 'Pay-as-you-Go',
    startup: 'Startup',
    business: 'Business',
    enterprise: 'Enterprise',
  };
  return labels[tier] ?? tier.charAt(0).toUpperCase() + tier.slice(1);
}

function tierPriceLabel(tier: TrustAPITierPricing): string {
  if (tier.monthly_price_usd === 0) return 'Free';
  return `${formatCurrency(tier.monthly_price_usd)}/mo`;
}

function tierRequestLabel(tier: TrustAPITierPricing): string {
  if (tier.monthly_request_limit >= 1000000) {
    return `${(tier.monthly_request_limit / 1000000).toFixed(0)}M req/mo`;
  }
  if (tier.monthly_request_limit >= 1000) {
    return `${(tier.monthly_request_limit / 1000).toFixed(0)}K req/mo`;
  }
  return `${tier.monthly_request_limit} req/mo`;
}

export function TrustAPIRegisterPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);
  const queryClient = useQueryClient();
  const [searchParams, setSearchParams] = useSearchParams();

  usePageTitle(t('trustAPIRegister.pageTitle'));

  const [name, setName] = useState('');
  const [slug, setSlug] = useState('');
  const [slugEdited, setSlugEdited] = useState(false);
  const [description, setDescription] = useState('');
  const [contactEmail, setContactEmail] = useState(user?.email ?? '');
  const [contactName, setContactName] = useState('');
  const [websiteUrl, setWebsiteUrl] = useState('');
  const [tier, setTier] = useState('developer');
  const [submitting, setSubmitting] = useState(false);
  const [created, setCreated] = useState(false);
  const [createdPartner, setCreatedPartner] = useState<{
    name: string;
    slug: string;
    tier: string;
    status: string;
  } | null>(null);

  const {
    data: tiersData,
    isLoading: tiersLoading,
    error: tiersError,
  } = useQuery({
    queryKey: ['trustapi', 'tiers'],
    queryFn: getTrustAPITierPricing,
    retry: false,
    staleTime: 5 * 60 * 1000,
  });

  const tiers = tiersData?.tiers ?? [];

  useEffect(() => {
    const checkoutStatus = searchParams.get('trustApiCheckout');
    if (!checkoutStatus) return;

    const next = new URLSearchParams(searchParams);
    next.delete('trustApiCheckout');
    setSearchParams(next, { replace: true });

    if (checkoutStatus === 'success') {
      toast.success(t('trustAPIRegister.paymentSuccess'), {
        description: t('trustAPIRegister.paymentSuccessDesc'),
      });
      queryClient.invalidateQueries({ queryKey: ['trustapi'] });
      navigate('/settings#trust-api', { replace: true });
    } else if (checkoutStatus === 'cancel') {
      toast.info(t('trustAPIRegister.checkoutCancelled'), {
        description: t('trustAPIRegister.checkoutCancelledDesc'),
      });
      navigate('/settings#trust-api', { replace: true });
    }
  }, [searchParams, setSearchParams, queryClient, navigate, t]);

  const handleNameChange = (value: string) => {
    setName(value);
    if (!slugEdited) {
      setSlug(slugify(value));
    }
  };

  const handleSlugChange = (value: string) => {
    setSlugEdited(true);
    setSlug(slugify(value));
  };

  const validate = (): boolean => {
    if (name.trim().length < 2) {
      toast.error(t('trustAPIRegister.nameTooShort'));
      return false;
    }
    if (slug.trim().length < 2) {
      toast.error(t('trustAPIRegister.slugTooShort'));
      return false;
    }
    if (!/^[a-z0-9][a-z0-9-]*[a-z0-9]$/.test(slug) && slug.length > 1) {
      toast.error(t('trustAPIRegister.slugInvalid'));
      return false;
    }
    if (!contactEmail.trim() || !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(contactEmail)) {
      toast.error(t('trustAPIRegister.emailRequired'));
      return false;
    }
    if (websiteUrl && !/^https?:\/\/.+/.test(websiteUrl)) {
      toast.error(t('trustAPIRegister.websiteInvalid'));
      return false;
    }
    return true;
  };

  const initiateCheckout = async (partnerTier: string) => {
    const origin = window.location.origin;
    const successUrl = `${origin}/trust-api/register?trustApiCheckout=success`;
    const cancelUrl = `${origin}/trust-api/register?trustApiCheckout=cancel`;

    try {
      const { url } = await createTrustAPICheckout(partnerTier, successUrl, cancelUrl);
      window.location.href = url;
    } catch (checkoutErr) {
      toast.error(getTrustAPIErrorMessage(checkoutErr, t('trustAPIRegister.checkoutFailed')));
      navigate('/settings#trust-api');
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    setSubmitting(true);
    try {
      const payload: TrustAPIPartnerCreateRequest = {
        name: name.trim(),
        slug: slug.trim(),
        contact_email: contactEmail.trim(),
        tier,
      };
      if (description.trim()) payload.description = description.trim();
      if (contactName.trim()) payload.contact_name = contactName.trim();
      if (websiteUrl.trim()) payload.website_url = websiteUrl.trim();

      const partner = await createTrustAPIPartner(payload);

      const isPaidTier = tier !== 'developer';
      if (isPaidTier) {
        toast.success(t('trustAPIRegister.partnerCreated'), {
          description: t('trustAPIRegister.redirectingToStripe'),
        });
        await initiateCheckout(tier);
        return;
      }

      setCreatedPartner({
        name: partner.name,
        slug: partner.slug,
        tier: partner.tier,
        status: partner.status,
      });
      setCreated(true);
      toast.success(t('trustAPIRegister.partnerCreatedSuccess'));
    } catch (err) {
      toast.error(getTrustAPIErrorMessage(err, t('trustAPIRegister.createFailed')));
    } finally {
      setSubmitting(false);
    }
  };

  if (created && createdPartner) {
    return (
      <div className="min-h-screen flex items-center justify-center p-4">
        <Card className="ff-card-velocity max-w-lg w-full">
          <CardHeader className="text-center">
            <div className="mx-auto mb-3 flex h-14 w-14 items-center justify-center rounded-full bg-green-500/10">
              <CheckCircle2 className="h-8 w-8 text-green-500" />
            </div>
            <CardTitle className="font-display text-2xl">
              {t('trustAPIRegister.registrationComplete')}
            </CardTitle>
            <CardDescription className="text-text-secondary">
              {t('trustAPIRegister.freeTierActive')}
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2 rounded-lg bg-bg-secondary border border-border-default p-4">
              <div className="flex items-center justify-between text-sm">
                <span className="text-text-muted">{t('trustAPIRegister.orgLabel')}</span>
                <span className="font-medium text-text-primary">{createdPartner.name}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-text-muted">{t('trustAPIRegister.slugLabel')}</span>
                <span className="font-mono text-text-primary">{createdPartner.slug}</span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-text-muted">{t('trustAPIRegister.tierLabel')}</span>
                <span className="font-medium text-text-primary">
                  {tierLabel(createdPartner.tier)}
                </span>
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-text-muted">{t('trustAPIRegister.statusLabel')}</span>
                <Badge variant="success" className="ff-badge-success">
                  {createdPartner.status}
                </Badge>
              </div>
            </div>
            <p className="text-sm text-text-muted text-center">
              {t('trustAPIRegister.freeTierDesc')}
            </p>
            <div className="flex gap-3">
              <Button
                variant="outline"
                className="flex-1"
                onClick={() => navigate('/settings#trust-api')}
              >
                <ArrowLeft className="w-4 h-4" />
                {t('trustAPIRegister.trustAPISettings')}
              </Button>
              <Button className="flex-1" onClick={() => navigate('/overview')}>
                {t('trustAPIRegister.goToDashboard')}
              </Button>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  const selectedTierData = tiers.find((t) => t.tier.toLowerCase() === tier);
  const isPaidTier = tier !== 'developer';

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="ff-card-velocity max-w-xl w-full">
        <CardHeader>
          <div className="flex items-center gap-3 mb-2">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-brand-500/10">
              <Shield className="h-5 w-5 text-brand-500" />
            </div>
            <div>
              <CardTitle className="font-display text-xl">
                {t('trustAPIRegister.cardTitle')}
              </CardTitle>
              <CardDescription className="text-text-secondary">
                {t('trustAPIRegister.cardDescription')}
              </CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-5">
            <div className="space-y-2">
              <Label htmlFor="name">
                {t('trustAPIRegister.orgName')} <span className="text-red-400">*</span>
              </Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => handleNameChange(e.target.value)}
                placeholder="Acme Corp"
                maxLength={255}
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="slug">
                {t('trustAPIRegister.slug')} <span className="text-red-400">*</span>
              </Label>
              <Input
                id="slug"
                value={slug}
                onChange={(e) => handleSlugChange(e.target.value)}
                placeholder="acme-corp"
                maxLength={100}
                className="font-mono"
                required
              />
              <p className="text-xs text-text-muted">{t('trustAPIRegister.slugHelp')}</p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="contactEmail">
                {t('trustAPIRegister.contactEmail')} <span className="text-red-400">*</span>
              </Label>
              <Input
                id="contactEmail"
                type="email"
                value={contactEmail}
                onChange={(e) => setContactEmail(e.target.value)}
                placeholder="api@acme.com"
                required
              />
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="contactName">{t('trustAPIRegister.contactName')}</Label>
                <Input
                  id="contactName"
                  value={contactName}
                  onChange={(e) => setContactName(e.target.value)}
                  placeholder="Jane Smith"
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="websiteUrl">{t('trustAPIRegister.websiteUrl')}</Label>
                <Input
                  id="websiteUrl"
                  value={websiteUrl}
                  onChange={(e) => setWebsiteUrl(e.target.value)}
                  placeholder="https://acme.com"
                />
              </div>
            </div>

            <div className="space-y-2">
              <Label htmlFor="tier">{t('trustAPIRegister.planTier')}</Label>
              <Select value={tier} onValueChange={setTier} disabled={tiersLoading}>
                <SelectTrigger id="tier">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {tiers.length > 0
                    ? tiers.map((t) => (
                        <SelectItem key={t.tier.toLowerCase()} value={t.tier.toLowerCase()}>
                          <div className="flex items-center justify-between w-full gap-3">
                            <span>{tierLabel(t.tier)}</span>
                            <span className="text-xs text-text-muted">
                              {tierPriceLabel(t)} &middot; {tierRequestLabel(t)}
                            </span>
                          </div>
                        </SelectItem>
                      ))
                    : (
                      <>
                        <SelectItem value="developer">Developer &mdash; Free</SelectItem>
                        <SelectItem value="startup">Startup</SelectItem>
                        <SelectItem value="business">Business</SelectItem>
                        <SelectItem value="enterprise">Enterprise</SelectItem>
                      </>
                    )}
                </SelectContent>
              </Select>
              {tiersError && (
                <p className="text-xs text-amber-400">
                  {t('trustAPIRegister.tiersLoadFailed')}
                </p>
              )}
              {selectedTierData && (
                <div className="rounded-lg bg-bg-secondary border border-border-default p-3 space-y-1">
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-text-muted">{t('trustAPIRegister.monthlyPrice')}</span>
                    <span className="font-medium text-text-primary">
                      {selectedTierData.monthly_price_usd === 0
                        ? t('trustAPISettings.free')
                        : formatCurrency(selectedTierData.monthly_price_usd)}
                    </span>
                  </div>
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-text-muted">{t('trustAPIRegister.includedRequests')}</span>
                    <span className="font-medium text-text-primary">
                      {selectedTierData.monthly_request_limit.toLocaleString()}/mo
                    </span>
                  </div>
                  {selectedTierData.has_overage_billing && (
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-text-muted">{t('trustAPIRegister.overage')}</span>
                      <span className="font-medium text-text-primary">
                        {formatCurrency(selectedTierData.overage_price_per_1000)}/1K req
                      </span>
                    </div>
                  )}
                  <div className="flex items-center justify-between text-sm">
                    <span className="text-text-muted">{t('trustAPIRegister.rateLimit')}</span>
                    <span className="font-medium text-text-primary">
                      {selectedTierData.rate_limit_per_minute}{t('trustAPIRegister.perMin')}
                    </span>
                  </div>
                  {selectedTierData.description && (
                    <p className="text-xs text-text-muted pt-1 border-t border-border-default mt-1">
                      {selectedTierData.description}
                    </p>
                  )}
                </div>
              )}
              {isPaidTier && (
                <div className="flex items-start gap-2 rounded-lg bg-brand-500/5 border border-brand-500/20 p-3">
                  <CreditCard className="h-4 w-4 text-brand-500 mt-0.5 shrink-0" />
                  <p className="text-xs text-text-secondary">
                    {t('trustAPIRegister.checkoutNotice')}
                  </p>
                </div>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="description">{t('trustAPIRegister.description')}</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder={t('trustAPIRegister.descriptionPlaceholder')}
                rows={3}
              />
            </div>

            <div className="flex gap-3 pt-2">
              <Button
                type="button"
                variant="outline"
                className="flex-1"
                onClick={() => navigate(-1)}
              >
                {t('trustAPIRegister.cancel')}
              </Button>
              <Button type="submit" htmlType="submit" className="flex-1" isLoading={submitting}>
                {isPaidTier ? (
                  <>
                    <CreditCard className="w-4 h-4" />
                    {t('trustAPIRegister.createAndSubscribe')}
                  </>
                ) : (
                  t('trustAPIRegister.createAccount')
                )}
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
