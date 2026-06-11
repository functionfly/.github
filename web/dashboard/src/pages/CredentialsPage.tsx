import { motion } from 'framer-motion';
import { Award, Loader2, FileText, ExternalLink } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { PageLayout } from '@/components/layout/PageLayout';
import { PageHeader } from '@/components/layout/PageHeader';
import { CredentialCard } from '@/components/certification/CredentialCard';
import { Button } from '@/components/ui/button';
import { useMyCredentials, useMyExams } from '@/hooks/useCertification';

import './credentials-page.css';

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
        className="mb-8"
      />

      {/* Active Credentials */}
      <div className="credentials-active-section">
        <div className="credentials-section-header">
          <Award className="h-5 w-5" />
          <h3>Active Certifications</h3>
        </div>

        {credsLoading ? (
          <div className="credentials-loading">
            <div className="credentials-loading-spinner" />
          </div>
        ) : credentials.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="credentials-empty-state"
          >
            <div className="credentials-empty-state-icon">
              <Award className="h-12 w-12" />
            </div>
            <h4>No Certifications Yet</h4>
            <p>Start your certification journey and earn verifiable credentials.</p>
            <button
              onClick={() => navigate('/certification')}
              className="btn-primary"
            >
              Browse Certifications
            </button>
          </motion.div>
        ) : (
          <div className="credentials-grid">
            {credentials.map((cred) => (
              <CredentialCard key={cred.id} credential={cred} />
            ))}
          </div>
        )}
      </div>

      {/* Exam History */}
      <div className="credentials-exam-section">
        <div className="credentials-section-header">
          <FileText className="h-5 w-5" />
          <h3>Exam History</h3>
        </div>

        {examsLoading ? (
          <div className="credentials-loading">
            <div className="credentials-loading-spinner" />
          </div>
        ) : exams.length === 0 ? (
          <div className="credentials-empty-state">
            <p className="text-sm text-text-muted">No exam attempts yet.</p>
          </div>
        ) : (
          <div className="credentials-exam-list">
            {exams.map((exam) => (
              <motion.div
                key={exam.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                className="credentials-exam-item"
              >
                <div className="credentials-exam-info">
                  <div className={`credentials-exam-status ${exam.status}`} />
                  <div>
                    <p className="credentials-exam-name">
                      {exam.tier_id || 'Exam'}
                    </p>
                    <p className="credentials-exam-date">
                      {new Date(exam.created_at || exam.started_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                <div className="credentials-exam-actions">
                  {exam.total_score != null && (
                    <span className="credentials-exam-score">
                      {exam.total_score.toFixed(1)}%
                    </span>
                  )}
                  <span className={`credentials-exam-status-badge ${exam.status}`}>
                    {exam.status}
                  </span>
                  {exam.status === 'in_progress' && (
                    <button
                      onClick={() => navigate(`/certification/exam/${exam.id}`)}
                      className="credentials-exam-continue"
                    >
                      Continue
                    </button>
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
          className="credentials-verification-section"
        >
          <div className="credentials-verification-content">
            <div className="credentials-verification-info">
              <h4>Share Your Credentials</h4>
              <p>Anyone can verify your certifications at your public verification URL.</p>
            </div>
            <a
              href={`/verify/${(window as any).__USERNAME__ || ''}`}
              target="_blank"
              rel="noopener noreferrer"
              className="credentials-verification-btn"
            >
              <ExternalLink className="h-4 w-4" />
              Verification Page
            </a>
          </div>
        </motion.div>
      )}
    </PageLayout>
  );
}
