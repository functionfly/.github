import { Award, TrendingUp, Users, BookOpen, Clock, Wrench } from 'lucide-react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  StatusPill,
  SealedButton,
  FrameButton,
  GaugeStrip,
  Gauge,
  AnnotationTag,
} from '@/components/containment';
import { TierCard } from '@/components/certification/TierCard';
import { certificationKeys, useCertTiers, useStartExam, useMyExams, useExam } from '@/hooks/useCertification';
import { useExamStatusStream } from '@/hooks/useExamStatusStream';
import { Loader2 } from 'lucide-react';

import './certification-page.css';

export function CertificationPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const queryClient = useQueryClient();
  const { data: tiersData, isLoading: tiersLoading } = useCertTiers();
  const { data: examsData } = useMyExams();
  const startExam = useStartExam();

  const paidExamId = searchParams.get('exam');
  const isPaidRedirect = searchParams.get('paid') === 'true';

  const { data: paidExamData } = useExam(paidExamId || '');

  useExamStatusStream(isPaidRedirect ? paidExamId || undefined : undefined);

  const paidExam = isPaidRedirect && paidExamId && paidExamData?.exam ? paidExamData.exam : null;

  useEffect(() => {
    if (paidExam?.id && paidExam.status === 'in_progress') {
      navigate(`/certification/exam/${paidExam.id}`);
    }
  }, [paidExam, navigate]);

  const activeExamsByTierId = (examsData?.exams || [])
    .filter((e) => e.status === 'in_progress')
    .reduce<Record<string, string>>((acc, e) => {
      if (e.tier_id) acc[e.tier_id] = e.id;
      return acc;
    }, {});

  const pendingExamsByTierId = (examsData?.exams || [])
    .filter((e) => e.status === 'pending_payment')
    .reduce<Record<string, string>>((acc, e) => {
      if (e.tier_id) acc[e.tier_id] = e.id;
      return acc;
    }, {});

  const handleStartExam = (tierSlug: string) => {
    startExam.mutate(tierSlug, {
      onSuccess: (data) => {
        const dataObj = data as { checkout_url?: string; exam?: { id: string } };
        const checkoutUrl = dataObj?.checkout_url;
        if (checkoutUrl) {
          window.location.href = checkoutUrl;
        } else if (dataObj?.exam?.id) {
          navigate(`/certification/exam/${dataObj.exam.id}`);
        }
      },
    });
  };

  const handleContinueToCheckout = (tierSlug: string) => {
    startExam.mutate(tierSlug, {
      onSuccess: (data) => {
        const dataObj = data as { checkout_url?: string };
        const checkoutUrl = dataObj?.checkout_url;
        if (checkoutUrl) {
          window.location.href = checkoutUrl;
        }
      },
    });
  };

  return (
    <div className="cert-page">
      <PageGrid />

      {/* Hero Chamber */}
      <Chamber className="cert-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="CHAMBER 01" secondary="Developer Certification" position="top-right" />

        <div className="cert-hero__header">
          <div className="cert-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="cert-hero__title">Developer Certification</h1>
          </div>
          <p className="cert-hero__subtitle">
            Prove your expertise. Earn verifiable credentials. Stand out in the marketplace.
          </p>
          <div className="cert-hero__badges">
            <StatusPill status="live" label="Active" />
          </div>
        </div>

        {/* Stats Gauge Strip */}
        <GaugeStrip>
          <Gauge
            isFirst
            data={{ value: 'Growing', label: 'Certified Developers' }}
          />
          <Gauge
            data={{ value: '+40%', label: 'Premium Pricing' }}
          />
          <Gauge
            data={{ value: 'Ready', label: 'LinkedIn Verified' }}
          />
        </GaugeStrip>
      </Chamber>

      {/* Tier Cards Section */}
      <div className="cert-section">
        <div className="cert-section__header">
          <h2 className="cert-section__title">Certification Tiers</h2>
          <p className="cert-section__subtitle">
            Choose your path. Each tier validates progressively deeper expertise.
          </p>
        </div>

        {tiersLoading ? (
          <div className="cert-loading">
            <Loader2 className="cert-loading__spinner" />
          </div>
        ) : (
          <div className="cert-tier-grid">
            {(tiersData?.tiers || []).map((tier) => {
              const pendingId = pendingExamsByTierId[tier.id] || (isPaidRedirect ? paidExamId || undefined : undefined);
              const hasPending = !!pendingId;
              return (
                <TierCard
                  key={tier.id}
                  tier={tier}
                  onStart={() => handleStartExam(tier.slug)}
                  isLoading={startExam.isPending}
                  activeExamId={activeExamsByTierId[tier.id]}
                  onResume={() => navigate(`/certification/exam/${activeExamsByTierId[tier.id]}`)}
                  pendingExamId={pendingId}
                  onBuyNow={() => {
                    if (hasPending) {
                      handleContinueToCheckout(tier.slug);
                    } else {
                      handleStartExam(tier.slug);
                    }
                  }}
                  paymentConfirmed={isPaidRedirect}
                />
              );
            })}
          </div>
        )}
      </div>

      {/* How It Works Chamber */}
      <Chamber className="cert-steps-chamber">
        <CornerBrace position="tr" />
        <CornerBrace position="bl" />
        <AnnotationTag primary="CHAMBER 02" secondary="Process Flow" position="top-right" />

        <div className="cert-steps__header">
          <Award className="cert-steps__icon" />
          <h2 className="cert-steps__title">How It Works</h2>
        </div>

        <div className="cert-steps-grid">
          {[
            { step: '01', title: 'Choose Your Tier', desc: 'Select Associate, Professional, or Architect based on your skill level.' },
            { step: '02', title: 'Take the Exam', desc: 'Answer knowledge questions and complete hands-on practical challenges.' },
            { step: '03', title: 'Get Certified', desc: 'Pass with 70%+ to earn your verifiable credential and digital badge.' },
            { step: '04', title: 'Share & Grow', desc: 'Display your badge on LinkedIn, resumes, and your FunctionFly profile.' },
          ].map((item) => (
            <div key={item.step} className="cert-step">
              <div className="cert-step__number">{item.step}</div>
              <h3 className="cert-step__title">{item.title}</h3>
              <p className="cert-step__desc">{item.desc}</p>
            </div>
          ))}
        </div>
      </Chamber>
    </div>
  );
}
