import { motion } from 'framer-motion';
import { Award, TrendingUp, Users, BookOpen } from 'lucide-react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { PageLayout } from '@/components/layout/PageLayout';
import { PageHeader } from '@/components/layout/PageHeader';
import { TierCard } from '@/components/certification/TierCard';
import { certificationKeys, useCertTiers, useStartExam, useMyExams, useExam } from '@/hooks/useCertification';
import { useExamStatusStream } from '@/hooks/useExamStatusStream';
import { useQueryClient } from '@tanstack/react-query';
import { Loader2 } from 'lucide-react';
import { useEffect } from 'react';

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
        const checkoutUrl = typeof data === 'object' && data && 'checkout_url' in (data as Record<string, unknown>) ? (data as Record<string, string>).checkout_url : undefined;
        if (checkoutUrl) {
          window.location.href = checkoutUrl;
        } else {
          navigate(`/certification/exam/${data.exam.id}`);
        }
      },
    });
  };

  const handleContinueToCheckout = (tierSlug: string) => {
    startExam.mutate(tierSlug, {
      onSuccess: (data) => {
        const checkoutUrl = typeof data === 'object' && data && 'checkout_url' in (data as Record<string, unknown>) ? (data as Record<string, string>).checkout_url : undefined;
        if (checkoutUrl) {
          window.location.href = checkoutUrl;
        }
      },
    });
  };

  return (
    <PageLayout>
      <PageHeader
        title="Developer Certification"
        subtitle="Prove your expertise. Earn verifiable credentials. Stand out in the marketplace."
        badges={[{ label: 'New', variant: 'new' }]}
        className="mb-8"
      />

      {/* Hero stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        {[
          { icon: Users, label: 'Certified Developers', value: 'Growing', color: 'users' },
          { icon: TrendingUp, label: 'Premium Pricing', value: 'Up to 40% more', color: 'trending' },
          { icon: BookOpen, label: 'Industry Recognition', value: 'LinkedIn Ready', color: 'book' },
        ].map((stat, i) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1 }}
            className="certification-stat-card"
          >
            <div className={`certification-stat-icon ${stat.color}`}>
              <stat.icon className="h-6 w-6" />
            </div>
            <div className="flex flex-col gap-1">
              <p className="certification-stat-label">{stat.label}</p>
              <p className="certification-stat-value">{stat.value}</p>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Tier Cards */}
      {tiersLoading ? (
        <div className="certification-loading">
          <div className="certification-loading-spinner" />
        </div>
      ) : (
        <div className="certification-tier-grid">
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

      {/* Info section */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.5 }}
        className="certification-how-it-works"
      >
        <div className="certification-how-it-works-header">
          <Award className="h-5 w-5" />
          <h3>How It Works</h3>
        </div>
        <div className="certification-steps-grid">
          {[
            { step: '1', title: 'Choose Your Tier', desc: 'Select Associate, Professional, or Architect based on your skill level.' },
            { step: '2', title: 'Take the Exam', desc: 'Answer knowledge questions and complete hands-on practical challenges.' },
            { step: '3', title: 'Get Certified', desc: 'Pass with 70%+ to earn your verifiable credential and digital badge.' },
            { step: '4', title: 'Share & Grow', desc: 'Display your badge on LinkedIn, resumes, and your FunctionFly profile.' },
          ].map((item) => (
            <div key={item.step} className="certification-step">
              <div className="certification-step-number">
                {item.step}
              </div>
              <h4 className="certification-step-title">{item.title}</h4>
              <p className="certification-step-description">{item.desc}</p>
            </div>
          ))}
        </div>
      </motion.div>
    </PageLayout>
  );
}
