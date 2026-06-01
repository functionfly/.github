import { useParams, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Loader2, CheckCircle2, XCircle, Trophy, ArrowLeft, Award } from 'lucide-react';
import { PageLayout } from '@/components/layout/PageLayout';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useExam } from '@/hooks/useCertification';

export function ExamResultsPage() {
  const { examId } = useParams<{ examId: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useExam(examId || '');

  const exam = data?.exam;

  if (isLoading) {
    return (
      <PageLayout>
        <div className="flex items-center justify-center py-24">
          <Loader2 className="h-8 w-8 animate-spin text-brand-500" />
        </div>
      </PageLayout>
    );
  }

  if (!exam) {
    return (
      <PageLayout>
        <div className="flex flex-col items-center justify-center py-24 gap-4">
          <p className="text-text-muted">Exam results not found.</p>
          <Button variant="outline" onClick={() => navigate('/certification')}>
            Back to Certification
          </Button>
        </div>
      </PageLayout>
    );
  }

  const isPassed = exam.passed === true;
  const isGraded = exam.status === 'passed' || exam.status === 'failed';

  return (
    <PageLayout>
      <div className="max-w-2xl mx-auto">
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          className="glass-card glow rounded-xl border border-theme bg-card p-8 text-center"
        >
          {/* Status icon */}
          <motion.div
            initial={{ scale: 0 }}
            animate={{ scale: 1 }}
            transition={{ delay: 0.2, type: 'spring', stiffness: 200 }}
            className="mb-6"
          >
            {!isGraded ? (
              <div className="flex flex-col items-center gap-3">
                <Loader2 className="h-16 w-16 animate-spin text-brand-500" />
                <h2 className="text-2xl font-bold text-text-primary">Grading in Progress</h2>
                <p className="text-text-muted">
                  Your knowledge questions are being graded. Practical challenges will be evaluated shortly.
                </p>
              </div>
            ) : isPassed ? (
              <div className="flex flex-col items-center gap-3">
                <div className="flex h-20 w-20 items-center justify-center rounded-full bg-emerald-500/10">
                  <Trophy className="h-10 w-10 text-emerald-500" />
                </div>
                <h2 className="text-3xl font-bold text-emerald-500">Congratulations!</h2>
                <p className="text-lg text-text-primary">You passed the certification exam!</p>
              </div>
            ) : (
              <div className="flex flex-col items-center gap-3">
                <div className="flex h-20 w-20 items-center justify-center rounded-full bg-red-500/10">
                  <XCircle className="h-10 w-10 text-red-500" />
                </div>
                <h2 className="text-3xl font-bold text-red-500">Not Passed</h2>
                <p className="text-lg text-text-primary">You didn't meet the passing threshold this time.</p>
              </div>
            )}
          </motion.div>

          {/* Scores */}
          {isGraded && (
            <div className="grid grid-cols-3 gap-4 mb-6">
              {[
                { label: 'Knowledge', value: exam.knowledge_score, max: 100 },
                { label: 'Practical', value: exam.practical_score, max: 100 },
                { label: 'Total', value: exam.total_score, max: 100 },
              ].map((score) => (
                <div key={score.label} className="rounded-lg border border-theme bg-bg-secondary p-4">
                  <p className="text-sm text-text-muted mb-1">{score.label}</p>
                  <p className="text-2xl font-bold text-text-primary">
                    {score.value != null ? `${score.value.toFixed(1)}%` : '—'}
                  </p>
                </div>
              ))}
            </div>
          )}

          {/* Actions */}
          <div className="flex items-center justify-center gap-3">
<Button variant="outline" onClick={() => navigate('/certification')} className="border-brand-500 text-brand-500 hover:bg-brand-500/10 hover:scale-105 transition-all duration-200">
            <ArrowLeft className="h-4 w-4" />
            Back to Certification
          </Button>
            {isPassed && (
              <Button
                onClick={() => navigate('/credentials')}
                className="bg-gradient-to-r from-brand-500 to-purple-500 text-white hover:from-brand-600 hover:to-purple-600 hover:scale-105 hover:shadow-lg hover:shadow-brand-500/25 transition-all duration-200"
              >
                <Award className="h-4 w-4" />
                View Credentials
              </Button>
            )}
            {!isPassed && isGraded && (
              <Button
                onClick={() => navigate('/certification')}
                className="bg-gradient-to-r from-brand-500 to-purple-500 text-white hover:from-brand-600 hover:to-purple-600 hover:scale-105 hover:shadow-lg hover:shadow-brand-500/25 transition-all duration-200"
              >
                Try Again
              </Button>
            )}
          </div>
        </motion.div>
      </div>
    </PageLayout>
  );
}
