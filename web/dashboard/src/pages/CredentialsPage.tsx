import { Award, Loader2, FileText, ExternalLink } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import {
  PageGrid,
  Chamber,
  CornerBrace,
  TrustSeal,
  SealedButton,
  FrameButton,
  AnnotationTag,
  StatusPill,
} from '@/components/containment';
import { CredentialCard } from '@/components/certification/CredentialCard';
import { useMyCredentials, useMyExams } from '@/hooks/useCertification';

import './credentials-page.css';

export function CredentialsPage() {
  const navigate = useNavigate();
  const { data: credsData, isLoading: credsLoading } = useMyCredentials();
  const { data: examsData, isLoading: examsLoading } = useMyExams();

  const credentials = credsData?.credentials || [];
  const exams = examsData?.exams || [];

  return (
    <div className="cred-page">
      <PageGrid />

      {/* Hero */}
      <Chamber className="cred-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="CHAMBER 01" secondary="My Credentials" position="top-right" />

        <div className="cred-hero__header">
          <div className="cred-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="cred-hero__title">My Credentials</h1>
          </div>
          <p className="cred-hero__subtitle">
            Your earned FunctionFly certifications and exam history.
          </p>
        </div>
      </Chamber>

      {/* Active Credentials */}
      <Chamber className="cred-section">
        <CornerBrace position="tr" />
        <CornerBrace position="bl" />
        <AnnotationTag primary="CHAMBER 02" secondary="Active Certifications" position="top-right" />

        <div className="cred-section__header">
          <Award className="cred-section__icon" />
          <h2 className="cred-section__title">Active Certifications</h2>
        </div>

        {credsLoading ? (
          <div className="cred-loading">
            <Loader2 className="cred-loading__spinner" />
          </div>
        ) : credentials.length === 0 ? (
          <div className="cred-empty">
            <Award className="cred-empty__icon" />
            <h3 className="cred-empty__title">No Certifications Yet</h3>
            <p className="cred-empty__desc">Start your certification journey and earn verifiable credentials.</p>
            <SealedButton onClick={() => navigate('/certification')}>
              Browse Certifications
            </SealedButton>
          </div>
        ) : (
          <div className="cred-grid">
            {credentials.map((cred) => (
              <CredentialCard key={cred.id} credential={cred} />
            ))}
          </div>
        )}
      </Chamber>

      {/* Exam History */}
      <Chamber className="cred-section">
        <CornerBrace position="tl" />
        <CornerBrace position="br" />

        <div className="cred-section__header">
          <FileText className="cred-section__icon" />
          <h2 className="cred-section__title">Exam History</h2>
        </div>

        {examsLoading ? (
          <div className="cred-loading">
            <Loader2 className="cred-loading__spinner" />
          </div>
        ) : exams.length === 0 ? (
          <div className="cred-empty cred-empty--compact">
            <p className="cred-empty__text">No exam attempts yet.</p>
          </div>
        ) : (
          <div className="cred-exam-list">
            {exams.map((exam) => (
              <div key={exam.id} className="cred-exam-item">
                <div className="cred-exam-info">
                  <div className={`cred-exam-dot cred-exam-dot--${exam.status}`} />
                  <div>
                    <p className="cred-exam-name">{exam.tier_id || 'Exam'}</p>
                    <p className="cred-exam-date">
                      {new Date(exam.created_at || exam.started_at).toLocaleDateString()}
                    </p>
                  </div>
                </div>
                <div className="cred-exam-actions">
                  {exam.total_score != null && (
                    <span className="cred-exam-score">{exam.total_score.toFixed(1)}%</span>
                  )}
                  <StatusPill
                    status={exam.status === 'passed' ? 'live' : exam.status === 'failed' ? 'revoked' : 'pending'}
                    label={exam.status.replace('_', ' ')}
                  />
                  {exam.status === 'in_progress' && (
                    <SealedButton
                      size="sm"
                      onClick={() => navigate(`/certification/exam/${exam.id}`)}
                    >
                      Continue
                    </SealedButton>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </Chamber>

      {/* Verification CTA */}
      {credentials.length > 0 && (
        <Chamber className="cred-verify">
          <CornerBrace position="tr" />
          <CornerBrace position="bl" />

          <div className="cred-verify__content">
            <div className="cred-verify__info">
              <div className="cred-verify__title-row">
                <TrustSeal size="md" />
                <h2 className="cred-verify__title">Share Your Credentials</h2>
              </div>
              <p className="cred-verify__desc">Anyone can verify your certifications at your public verification URL.</p>
            </div>
            <a
              href={`/verify/${(window as any).__USERNAME__ || ''}`}
              target="_blank"
              rel="noopener noreferrer"
            >
              <SealedButton iconRight={<ExternalLink className="h-4 w-4" />}>
                Verification Page
              </SealedButton>
            </a>
          </div>
        </Chamber>
      )}
    </div>
  );
}
