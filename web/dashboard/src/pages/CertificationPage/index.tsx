import { motion } from 'framer-motion';
import { Award, TrendingUp, Users, BookOpen } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { PageLayout } from '@/components/layout/PageLayout';
import { PageHeader } from '@/components/layout/PageHeader';
import { TierCard } from '@/components/certification/TierCard';
import { useCertTiers, useStartExam, useMyExams } from '@/hooks/useCertification';
import { Loader2 } from 'lucide-react';

export function CertificationPage() {
  const navigate = useNavigate();
  const { data: tiersData, isLoading: tiersLoading } = useCertTiers();
  const { data: examsData } = useMyExams();
  const startExam = useStartExam();

  const activeExamsByTierId = (examsData?.exams || [])
    .filter((e) => e.status === 'in_progress')
    .reduce<Record<string, string>>((acc, e) => {
      if (e.tier_id) acc[e.tier_id] = e.id;
      return acc;
    }, {});

  const handleStartExam = (tierSlug: string) => {
    startExam.mutate(tierSlug, {
      onSuccess: (data) => {
        if (!data.checkout_url) {
          navigate(`/certification/exam/${data.exam.id}`);
        }
      },
    });
  };

  const handleResumeExam = (examId: string) => {
    navigate(`/certification/exam/${examId}`);
  };

  return (
    <PageLayout>
      <PageHeader
        title="Developer Certification"
        subtitle="Prove your expertise. Earn verifiable credentials. Stand out in the marketplace."
        badges={[{ label: 'New', variant: 'new' }]}
      />

      {/* Hero stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
        {[
          { icon: Users, label: 'Certified Developers', value: 'Growing', color: 'text-blue-500' },
          { icon: TrendingUp, label: 'Premium Pricing', value: 'Up to 40% more', color: 'text-emerald-500' },
          { icon: BookOpen, label: 'Industry Recognition', value: 'LinkedIn Ready', color: 'text-purple-500' },
        ].map((stat, i) => (
          <motion.div
            key={stat.label}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1 }}
            className="glass-card rounded-xl border border-theme bg-card p-4 flex items-center gap-3"
          >
            <stat.icon className={`h-8 w-8 ${stat.color}`} />
            <div>
              <p className="text-sm text-text-muted">{stat.label}</p>
              <p className="text-lg font-bold text-text-primary">{stat.value}</p>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Tier Cards */}
      {tiersLoading ? (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-brand-500" />
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {(tiersData?.tiers || []).map((tier) => (
            <TierCard
              key={tier.id}
              tier={tier}
              onStart={() => handleStartExam(tier.slug)}
              isLoading={startExam.isPending}
              activeExamId={activeExamsByTierId[tier.id]}
              onResume={() => handleResumeExam(activeExamsByTierId[tier.id])}
            />
          ))}
        </div>
      )}

      {/* Info section */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.5 }}
        className="mt-12 rounded-xl border border-theme bg-card p-6"
      >
        <h3 className="text-lg font-bold text-text-primary mb-4 flex items-center gap-2">
          <Award className="h-5 w-5 text-brand-500" />
          How It Works
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
          {[
            { step: '1', title: 'Choose Your Tier', desc: 'Select Associate, Professional, or Architect based on your skill level.' },
            { step: '2', title: 'Take the Exam', desc: 'Answer knowledge questions and complete hands-on practical challenges.' },
            { step: '3', title: 'Get Certified', desc: 'Pass with 70%+ to earn your verifiable credential and digital badge.' },
            { step: '4', title: 'Share & Grow', desc: 'Display your badge on LinkedIn, resumes, and your FunctionFly profile.' },
          ].map((item) => (
            <div key={item.step} className="text-center">
              <div className="mb-2 flex h-10 w-10 items-center justify-center rounded-full bg-brand-500/10 text-brand-500 font-bold mx-auto">
                {item.step}
              </div>
              <h4 className="font-medium text-text-primary mb-1">{item.title}</h4>
              <p className="text-sm text-text-muted">{item.desc}</p>
            </div>
          ))}
        </div>
      </motion.div>
    </PageLayout>
  );
}
