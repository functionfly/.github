import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  performanceApi,
  type PerformanceGoal,
  type PerformanceReview,
  type PeerFeedback,
} from '@/api/performance';
import { useAuthStore } from '@/stores/authStore';
import { Target, Award, MessageSquare, Plus, Star, CheckCircle, Clock } from 'lucide-react';
import { formatDate } from '@/lib/utils';

const goalStatusColors: Record<string, string> = {
  not_started: 'bg-gray-500/20 text-gray-400',
  in_progress: 'bg-blue-500/20 text-blue-400',
  completed: 'bg-green-500/20 text-green-400',
  cancelled: 'bg-red-500/20 text-red-400',
};

const reviewStatusColors: Record<string, string> = {
  draft: 'bg-gray-500/20 text-gray-400',
  submitted: 'bg-blue-500/20 text-blue-400',
  in_review: 'bg-yellow-500/20 text-yellow-400',
  completed: 'bg-green-500/20 text-green-400',
};

export function PerformancePage() {
  const queryClient = useQueryClient();
  const { user } = useAuthStore();
  const [tab, setTab] = useState<'goals' | 'reviews' | 'feedback'>('goals');
  const [showGoalModal, setShowGoalModal] = useState(false);
  const [showReviewModal, setShowReviewModal] = useState(false);
  const [showFeedbackModal, setShowFeedbackModal] = useState(false);
  const [goalFilter, setGoalFilter] = useState('');

  const [newGoal, setNewGoal] = useState({ title: '', description: '', category: 'professional', priority: 'medium', target_date: '' });
  const [newReview, setNewReview] = useState({ review_period: '', review_type: 'self', comments: '' });
  const [newFeedback, setNewFeedback] = useState({ to_employee_id: '', feedback_text: '', rating: 0, is_anonymous: false });

  const { data: goalsData } = useQuery({
    queryKey: ['performance', 'goals'],
    queryFn: () => performanceApi.listGoals(),
  });

  const { data: reviewsData } = useQuery({
    queryKey: ['performance', 'reviews'],
    queryFn: () => performanceApi.listReviews(),
  });

  const { data: feedbackData } = useQuery({
    queryKey: ['performance', 'feedback', user?.id],
    queryFn: () => performanceApi.listFeedback(user!.id),
    enabled: !!user?.id,
  });

  const createGoalMutation = useMutation({
    mutationFn: (data: Partial<PerformanceGoal>) => performanceApi.createGoal(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['performance', 'goals'] });
      setShowGoalModal(false);
      setNewGoal({ title: '', description: '', category: 'professional', priority: 'medium', target_date: '' });
    },
  });

  const updateGoalMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<PerformanceGoal> }) => performanceApi.updateGoal(id, data),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['performance', 'goals'] }),
  });

  const createReviewMutation = useMutation({
    mutationFn: (data: Partial<PerformanceReview>) => performanceApi.createReview(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['performance', 'reviews'] });
      setShowReviewModal(false);
      setNewReview({ review_period: '', review_type: 'self', comments: '' });
    },
  });

  const submitReviewMutation = useMutation({
    mutationFn: (id: string) => performanceApi.submitReview(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['performance', 'reviews'] }),
  });

  const giveFeedbackMutation = useMutation({
    mutationFn: (data: Partial<PeerFeedback>) => performanceApi.giveFeedback(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['performance', 'feedback'] });
      setShowFeedbackModal(false);
      setNewFeedback({ to_employee_id: '', feedback_text: '', rating: 0, is_anonymous: false });
    },
  });

  const goals = goalsData?.data?.goals || [];
  const reviews = reviewsData?.data?.reviews || [];
  const feedback = feedbackData?.data?.feedback || [];

  const filteredGoals = goalFilter ? goals.filter((g) => g.status === goalFilter) : goals;

  const tabs = [
    { id: 'goals' as const, label: 'Goals', icon: Target },
    { id: 'reviews' as const, label: 'Reviews', icon: Award },
    { id: 'feedback' as const, label: 'Feedback', icon: MessageSquare },
  ];

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Performance</h1>

      <div className="flex gap-1 rounded-lg bg-gray-900 p-1">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`flex flex-1 items-center justify-center gap-2 rounded-md px-4 py-2 text-sm font-medium ${
              tab === t.id ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-gray-200'
            }`}
          >
            <t.icon className="h-4 w-4" />
            {t.label}
          </button>
        ))}
      </div>

      {/* Goals Tab */}
      {tab === 'goals' && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <select
              value={goalFilter}
              onChange={(e) => setGoalFilter(e.target.value)}
              className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="">All Statuses</option>
              <option value="not_started">Not Started</option>
              <option value="in_progress">In Progress</option>
              <option value="completed">Completed</option>
              <option value="cancelled">Cancelled</option>
            </select>
            <button
              onClick={() => setShowGoalModal(true)}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            >
              <Plus className="h-4 w-4" />
              New Goal
            </button>
          </div>

          {filteredGoals.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
              <Target className="mb-4 h-12 w-12 text-gray-600" />
              <p className="text-gray-400">No goals found</p>
            </div>
          ) : (
            <div className="space-y-3">
              {filteredGoals.map((goal) => (
                <div key={goal.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
                  <div className="mb-2 flex items-start justify-between">
                    <div>
                      <h3 className="font-semibold text-gray-100">{goal.title}</h3>
                      {goal.description && <p className="mt-1 text-sm text-gray-400">{goal.description}</p>}
                    </div>
                    <span className={`rounded-full px-2 py-0.5 text-xs ${goalStatusColors[goal.status] || ''}`}>
                      {goal.status.replace('_', ' ')}
                    </span>
                  </div>
                  <div className="mb-2 flex items-center gap-4 text-xs text-gray-500">
                    <span className="capitalize">{goal.category}</span>
                    <span className="capitalize">{goal.priority}</span>
                    {goal.target_date && <span>Target: {formatDate(goal.target_date)}</span>}
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="flex-1 h-2 rounded-full bg-gray-800">
                      <div
                        className="h-full rounded-full bg-blue-600 transition-all"
                        style={{ width: `${goal.progress_pct}%` }}
                      />
                    </div>
                    <span className="text-sm font-medium text-blue-400">{goal.progress_pct}%</span>
                  </div>
                  {goal.status !== 'completed' && goal.status !== 'cancelled' && (
                    <div className="mt-3 flex gap-2">
                      {[25, 50, 75, 100].map((pct) => (
                        <button
                          key={pct}
                          onClick={() => updateGoalMutation.mutate({ id: goal.id, data: { progress_pct: pct, status: pct === 100 ? 'completed' : goal.status } })}
                          className="rounded border border-gray-700 bg-gray-800 px-2 py-1 text-xs text-gray-300 hover:bg-gray-700"
                        >
                          {pct}%
                        </button>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Reviews Tab */}
      {tab === 'reviews' && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <button
              onClick={() => setShowReviewModal(true)}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            >
              <Plus className="h-4 w-4" />
              New Review
            </button>
          </div>

          {reviews.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
              <Award className="mb-4 h-12 w-12 text-gray-600" />
              <p className="text-gray-400">No reviews yet</p>
            </div>
          ) : (
            <div className="space-y-3">
              {reviews.map((review) => (
                <div key={review.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
                  <div className="mb-2 flex items-start justify-between">
                    <div>
                      <h3 className="font-semibold text-gray-100">{review.review_period}</h3>
                      <p className="text-sm text-gray-500 capitalize">{review.review_type.replace('_', ' ')}</p>
                    </div>
                    <span className={`rounded-full px-2 py-0.5 text-xs ${reviewStatusColors[review.status] || ''}`}>
                      {review.status.replace('_', ' ')}
                    </span>
                  </div>
                  {review.overall_rating && (
                    <div className="mb-2 flex items-center gap-1">
                      {[1, 2, 3, 4, 5].map((star) => (
                        <Star
                          key={star}
                          className={`h-4 w-4 ${star <= review.overall_rating! ? 'fill-yellow-400 text-yellow-400' : 'text-gray-600'}`}
                        />
                      ))}
                      <span className="ml-2 text-sm text-gray-400">{review.overall_rating}/5</span>
                    </div>
                  )}
                  {review.strengths && (
                    <p className="text-sm text-gray-400"><span className="text-green-400">Strengths:</span> {review.strengths}</p>
                  )}
                  {review.areas_for_improvement && (
                    <p className="text-sm text-gray-400"><span className="text-orange-400">Areas:</span> {review.areas_for_improvement}</p>
                  )}
                  {review.status === 'draft' && (
                    <button
                      onClick={() => submitReviewMutation.mutate(review.id)}
                      className="mt-3 flex items-center gap-2 rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-blue-700"
                    >
                      <CheckCircle className="h-3 w-3" />
                      Submit
                    </button>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Feedback Tab */}
      {tab === 'feedback' && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <button
              onClick={() => setShowFeedbackModal(true)}
              className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
            >
              <Plus className="h-4 w-4" />
              Give Feedback
            </button>
          </div>

          {feedback.length === 0 ? (
            <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
              <MessageSquare className="mb-4 h-12 w-12 text-gray-600" />
              <p className="text-gray-400">No feedback received yet</p>
            </div>
          ) : (
            <div className="space-y-3">
              {feedback.map((fb) => (
                <div key={fb.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
                  <div className="mb-2 flex items-center justify-between">
                    <span className="text-sm text-gray-400">
                      {fb.is_anonymous ? 'Anonymous' : `From: ${fb.from_employee_id}`}
                    </span>
                    {fb.rating && (
                      <div className="flex items-center gap-1">
                        {[1, 2, 3, 4, 5].map((star) => (
                          <Star
                            key={star}
                            className={`h-3 w-3 ${star <= fb.rating! ? 'fill-yellow-400 text-yellow-400' : 'text-gray-600'}`}
                          />
                        ))}
                      </div>
                    )}
                  </div>
                  <p className="text-sm text-gray-200">{fb.feedback_text}</p>
                  <p className="mt-2 text-xs text-gray-500">{formatDate(fb.created_at)}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Goal Modal */}
      {showGoalModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Create Goal</h2>
            <input
              type="text"
              placeholder="Goal title"
              value={newGoal.title}
              onChange={(e) => setNewGoal({ ...newGoal, title: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              autoFocus
            />
            <textarea
              placeholder="Description (optional)"
              value={newGoal.description}
              onChange={(e) => setNewGoal({ ...newGoal, description: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              rows={3}
            />
            <div className="mb-3 grid grid-cols-2 gap-3">
              <select
                value={newGoal.category}
                onChange={(e) => setNewGoal({ ...newGoal, category: e.target.value })}
                className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              >
                <option value="professional">Professional</option>
                <option value="personal">Personal</option>
                <option value="technical">Technical</option>
                <option value="leadership">Leadership</option>
              </select>
              <select
                value={newGoal.priority}
                onChange={(e) => setNewGoal({ ...newGoal, priority: e.target.value })}
                className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>
            <input
              type="date"
              value={newGoal.target_date}
              onChange={(e) => setNewGoal({ ...newGoal, target_date: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowGoalModal(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => createGoalMutation.mutate(newGoal)}
                disabled={!newGoal.title.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Review Modal */}
      {showReviewModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Create Review</h2>
            <input
              type="text"
              placeholder="Review period (e.g. Q1 2026)"
              value={newReview.review_period}
              onChange={(e) => setNewReview({ ...newReview, review_period: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              autoFocus
            />
            <select
              value={newReview.review_type}
              onChange={(e) => setNewReview({ ...newReview, review_type: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="self">Self Review</option>
              <option value="peer">Peer Review</option>
              <option value="manager">Manager Review</option>
            </select>
            <textarea
              placeholder="Comments (optional)"
              value={newReview.comments}
              onChange={(e) => setNewReview({ ...newReview, comments: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              rows={3}
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowReviewModal(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => createReviewMutation.mutate(newReview)}
                disabled={!newReview.review_period.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Feedback Modal */}
      {showFeedbackModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Give Feedback</h2>
            <input
              type="text"
              placeholder="Employee ID"
              value={newFeedback.to_employee_id}
              onChange={(e) => setNewFeedback({ ...newFeedback, to_employee_id: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              autoFocus
            />
            <textarea
              placeholder="Your feedback..."
              value={newFeedback.feedback_text}
              onChange={(e) => setNewFeedback({ ...newFeedback, feedback_text: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
              rows={4}
            />
            <div className="mb-3 flex items-center gap-2">
              <span className="text-sm text-gray-400">Rating:</span>
              {[1, 2, 3, 4, 5].map((star) => (
                <button key={star} onClick={() => setNewFeedback({ ...newFeedback, rating: star })}>
                  <Star className={`h-5 w-5 ${star <= newFeedback.rating ? 'fill-yellow-400 text-yellow-400' : 'text-gray-600'}`} />
                </button>
              ))}
            </div>
            <label className="mb-4 flex items-center gap-2 text-sm text-gray-400">
              <input
                type="checkbox"
                checked={newFeedback.is_anonymous}
                onChange={(e) => setNewFeedback({ ...newFeedback, is_anonymous: e.target.checked })}
                className="rounded border-gray-700 bg-gray-800"
              />
              Submit anonymously
            </label>
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowFeedbackModal(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => giveFeedbackMutation.mutate(newFeedback)}
                disabled={!newFeedback.to_employee_id.trim() || !newFeedback.feedback_text.trim()}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Submit
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
