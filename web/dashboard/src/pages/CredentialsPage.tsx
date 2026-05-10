import { motion } from 'framer-motion';
import { Award, Loader2, FileText, ExternalLink } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { PageLayout } from '@/components/layout/PageLayout';
import { PageHeader } from '@/components/layout/PageHeader';
import { CredentialCard } from '@/components/certification/CredentialCard';
import { Button } from '@/components/ui/button';
import { useMyCredentials, useMyExams } from '@/hooks/useCertification';

export function CredentialsPage() {
  const navigate = useNavigate();
  const { data: credsData, isLoading: credsLoading } = useMyCredentials();
  const { data: examsData, isLoading: examsLoading } = useMyExams();

  const credentials = credsData?.credentials || [];
  const exams = examsData?.exams || [];

  return (
    <PageLayout>
      <PageHeader
        title="My Credentials"
        subtitle="Your earned FunctionFly certifications and exam history."
      />

      {/* Active Credentials */}
      <div className="mb-8">
        <h3 className="text-lg font-bold text-text-primary mb-4 flex items-center gap-2">
          <Award className="h-5 w-5 text-brand-500" />
          Active Certifications
        </h3>

        {credsLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-brand-500" />
          </div>
        ) : credentials.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="rounded-xl border border-theme bg-card p-8 text-center"
          >
            <Award className="h-12 w-12 text-text-muted mx-auto mb-4" />
            <h4 className="text-lg font-medium text-text-primary mb-2">No Certifications Yet</h4>
            <p className="text-sm text-text-muted mb-4">
              Start your certification journey and earn verifiable credentials.
            </p>
            <Button
              onClick={() => navigate('/certification')}
              className="bg-gradient-to-r from-brand-500 to-purple-500 text-white"
            >
              Browse Certifications
            </Button>
          </motion.div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {credentials.map((cred) => (
              <CredentialCard key={cred.id} credential={cred} />
            ))}
          </div>
        )}
      </div>

      {/* Exam History */}
      <div>
        <h3 className="text-lg font-bold text-text-primary mb-4 flex items-center gap-2">
          <FileText className="h-5 w-5 text-brand-500" />
          Exam History
        </h3>

        {examsLoading ? (
          <div className="flex items-center justify-center py-12">
            <Loader2 className="h-8 w-8 animate-spin text-brand-500" />
          </div>
        ) : exams.length === 0 ? (
          <div className="rounded-xl border border-theme bg-card p-6 text-center">
            <p className="text-sm text-text-muted">No exam attempts yet.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {exams.map((exam) => (
              <motion.div
                key={exam.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                className="flex items-center justify-between rounded-lg border border-theme bg-card p-4"
              >
                <div className="flex items-center gap-3">
                  <div className={`h-3 w-3 rounded-full ${
                    exam.status === 'passed' ? 'bg-emerald-500' :
                    exam.status === 'failed' ? 'bg-red-500' :
                    exam.status === 'in_progress' ? 'bg-amber-500' :
                    'bg-gray-400'
                  }`} />
                  <div>
                    <p className="text-sm font-medium text-text-primary">
                      {exam.tier_id || 'Exam'}
                    </p>
                    <p className="text-xs text-text-muted">
                      {new Date(exam.created_at || exam.started_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  {exam.total_score != null && (
                    <span className="text-sm font-medium text-text-secondary">
                      {exam.total_score.toFixed(1)}%
                    </span>
                  )}
                  <span className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
                    exam.status === 'passed' ? 'bg-emerald-500/10 text-emerald-500' :
                    exam.status === 'failed' ? 'bg-red-500/10 text-red-500' :
                    exam.status === 'in_progress' ? 'bg-amber-500/10 text-amber-500' :
                    'bg-gray-500/10 text-gray-500'
                  }`}>
                    {exam.status}
                  </span>
                  {exam.status === 'in_progress' && (
                    <Button size="sm" variant="outline" onClick={() => navigate(`/certification/exam/${exam.id}`)}>
                      Continue
                    </Button>
                  )}
                </div>
              </motion.div>
            ))}
          </div>
        )}
      </div>

      {/* Verification link */}
      {credentials.length > 0 && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.3 }}
          className="mt-8 rounded-xl border border-theme bg-card p-6 flex items-center justify-between"
        >
          <div>
            <h4 className="font-medium text-text-primary">Share Your Credentials</h4>
            <p className="text-sm text-text-muted">
              Anyone can verify your certifications at your public verification URL.
            </p>
          </div>
          <Button variant="outline" asChild>
            <a href={`/verify/${(window as any).__USERNAME__ || ''}`} target="_blank" rel="noopener noreferrer">
              <ExternalLink className="h-4 w-4" />
              Verification Page
            </a>
          </Button>
        </motion.div>
      )}
    </PageLayout>
  );
}
