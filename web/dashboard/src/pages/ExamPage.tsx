import { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';
import { Loader2, Send, AlertTriangle } from 'lucide-react';
import { PageLayout } from '@/components/layout/PageLayout';
import { ExamQuestion } from '@/components/certification/ExamQuestion';
import { ExamTimer } from '@/components/certification/ExamTimer';
import { ExamProgress } from '@/components/certification/ExamProgress';
import { PracticalChallenge } from '@/components/certification/PracticalChallenge';
import { Button } from '@/components/ui/button';
import { useExam, useSubmitAnswer, useSubmitExam, useAbandonExam } from '@/hooks/useCertification';
import { useExamStore } from '@/stores/examStore';
import { toast } from 'sonner';

export function ExamPage() {
  const { examId } = useParams<{ examId: string }>();
  const navigate = useNavigate();
  const { data, isLoading, error } = useExam(examId || '');
  const submitAnswer = useSubmitAnswer();
  const submitExam = useSubmitExam();
  const abandonExam = useAbandonExam();
  const { answers, setAnswer, currentQuestionIndex, setCurrentQuestionIndex, reset } = useExamStore();
  const [showSubmitConfirm, setShowSubmitConfirm] = useState(false);

  const exam = data?.exam;
  const questions = exam?.questions || [];
  const questionIds = questions.map((q) => q.id);

  // Sync answers from server on load
  useEffect(() => {
    if (exam?.answers && Object.keys(answers).length === 0) {
      useExamStore.getState().setAnswers(exam.answers as Record<string, string>);
    }
  }, [exam?.answers]);

  // Redirect if exam is no longer in progress
  useEffect(() => {
    if (exam && exam.status !== 'in_progress') {
      navigate(`/certification/exam/${examId}/results`, { replace: true });
    }
  }, [exam, examId, navigate]);

  const handleAnswer = useCallback(
    (questionId: string, answer: string) => {
      setAnswer(questionId, answer);
      // Persist to server (debounced via mutation)
      if (examId) {
        submitAnswer.mutate({ examId, questionId, answer });
      }
    },
    [examId, setAnswer, submitAnswer]
  );

  const handleSubmit = useCallback(() => {
    if (!examId) return;
    submitExam.mutate(
      { examId, answers },
      {
        onSuccess: () => {
          reset();
          navigate(`/certification/exam/${examId}/results`, { replace: true });
        },
      }
    );
  }, [examId, answers, submitExam, reset, navigate]);

  const handleExpire = useCallback(() => {
    toast.warning('Time is up! Submitting your exam...');
    handleSubmit();
  }, [handleSubmit]);

  const handleAbandon = useCallback(() => {
    if (!examId) return;
    abandonExam.mutate(examId, {
      onSuccess: () => {
        reset();
        navigate('/certification', { replace: true });
      },
    });
  }, [examId, abandonExam, reset, navigate]);

  if (isLoading) {
    return (
      <PageLayout>
        <div className="flex items-center justify-center py-24">
          <Loader2 className="h-8 w-8 animate-spin text-brand-500" />
        </div>
      </PageLayout>
    );
  }

  if (error || !exam) {
    return (
      <PageLayout>
        <div className="flex flex-col items-center justify-center py-24 gap-4">
          <AlertTriangle className="h-12 w-12 text-red-500" />
          <p className="text-text-muted">Failed to load exam. It may have expired.</p>
          <Button variant="outline" onClick={() => navigate('/certification')} className="border-brand-500 text-brand-500 hover:bg-brand-500/10 hover:scale-105 transition-all duration-200">
            Back to Certification
          </Button>
        </div>
      </PageLayout>
    );
  }

  const currentQuestion = questions[currentQuestionIndex];
  const answeredCount = Object.keys(answers).length;

  return (
    <PageLayout>
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Sidebar — Progress + Timer */}
        <div className="lg:col-span-1 space-y-4">
          <div className="sticky top-24 space-y-4">
            {/* Timer */}
            <div className="rounded-xl border border-theme bg-card p-4">
              <h4 className="text-sm font-medium text-text-muted mb-3">Time Remaining</h4>
              <ExamTimer expiresAt={exam.expires_at} onExpire={handleExpire} />
            </div>

            {/* Progress */}
            <div className="rounded-xl border border-theme bg-card p-4">
              <h4 className="text-sm font-medium text-text-muted mb-3">Progress</h4>
              <ExamProgress
                currentIndex={currentQuestionIndex}
                totalQuestions={questions.length}
                answeredCount={answeredCount}
                answers={answers}
                questionIds={questionIds}
                onNavigate={setCurrentQuestionIndex}
              />
            </div>

            {/* Submit */}
            <div className="rounded-xl border border-theme bg-card p-4">
              {showSubmitConfirm ? (
                <div className="space-y-3">
                  <p className="text-sm text-text-primary font-medium">
                    Submit {answeredCount} of {questions.length} answers?
                  </p>
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      onClick={handleSubmit}
                      disabled={submitExam.isPending}
                      className="flex-1 bg-gradient-to-r from-brand-500 to-purple-500 text-white"
                    >
                      {submitExam.isPending ? (
                        <Loader2 className="h-4 w-4 animate-spin" />
                      ) : (
                        <>
                          <Send className="h-4 w-4" />
                          Confirm
                        </>
                      )}
                    </Button>
                    <Button size="sm" variant="outline" onClick={() => setShowSubmitConfirm(false)}>
                      Cancel
                    </Button>
                  </div>
                </div>
              ) : (
                <div className="space-y-2">
                  <Button
                    onClick={() => setShowSubmitConfirm(true)}
                    className="w-full bg-gradient-to-r from-brand-500 to-purple-500 text-white"
                  >
                    <Send className="h-4 w-4" />
                    Submit Exam
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleAbandon}
                    className="w-full text-text-muted hover:text-red-500"
                  >
                    Abandon Exam
                  </Button>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Main — Question */}
        <div className="lg:col-span-3">
          <div className="rounded-xl border border-theme bg-card p-6">
            <AnimatePresence mode="wait">
              {currentQuestion && (
                <motion.div
                  key={currentQuestion.id}
                  initial={{ opacity: 0, x: 20 }}
                  animate={{ opacity: 1, x: 0 }}
                  exit={{ opacity: 0, x: -20 }}
                  transition={{ duration: 0.2 }}
                >
                  <ExamQuestion
                    question={currentQuestion}
                    selectedAnswer={answers[currentQuestion.id]}
                    onAnswer={(answer) => handleAnswer(currentQuestion.id, answer)}
                    questionNumber={currentQuestionIndex + 1}
                    totalQuestions={questions.length}
                    onNext={() => setCurrentQuestionIndex(currentQuestionIndex + 1)}
                    onPrevious={() => setCurrentQuestionIndex(currentQuestionIndex - 1)}
                  />
                </motion.div>
              )}
            </AnimatePresence>
          </div>

          {/* Practical Challenges */}
          {exam.practical_challenges && exam.practical_challenges.length > 0 && (
            <div className="mt-6 space-y-4">
              <h3 className="text-lg font-bold text-text-primary">Practical Challenges</h3>
              {exam.practical_challenges.map((challenge) => (
                <PracticalChallenge key={challenge.id} challenge={challenge} examId={exam.id} />
              ))}
            </div>
          )}
        </div>
      </div>
    </PageLayout>
  );
}
