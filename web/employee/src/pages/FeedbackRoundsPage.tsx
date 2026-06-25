import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { feedbackRoundsApi, type FeedbackRound } from '@/api/feedback_rounds';
import { MessageCircle, Plus, Play, BarChart3, Trash2 } from 'lucide-react';

const statusColors: Record<string, string> = {
  draft: 'bg-gray-500/20 text-gray-400',
  active: 'bg-green-500/20 text-green-400',
  collecting: 'bg-blue-500/20 text-blue-400',
  completed: 'bg-purple-500/20 text-purple-400',
};

const questionTypeLabels: Record<string, string> = {
  text: 'Text',
  rating: 'Rating (1-5)',
  boolean: 'Yes/No',
};

export function FeedbackRoundsPage() {
  const queryClient = useQueryClient();
  const [showCreate, setShowCreate] = useState(false);
  const [showResults, setShowResults] = useState<string | null>(null);
  const [form, setForm] = useState({
    name: '',
    description: '',
    review_period: '',
    round_type: 'peer',
    start_date: '',
    end_date: '',
  });
  const [questions, setQuestions] = useState<{ question: string; type: string; required: boolean }[]>([
    { question: '', type: 'text', required: true },
  ]);

  const { data, isLoading } = useQuery({
    queryKey: ['feedback-rounds'],
    queryFn: () => feedbackRoundsApi.list(),
  });

  const { data: resultsData } = useQuery({
    queryKey: ['feedback-rounds-results', showResults],
    queryFn: () => feedbackRoundsApi.getResults(showResults!),
    enabled: !!showResults,
  });

  const createMutation = useMutation({
    mutationFn: (data: Partial<FeedbackRound>) => feedbackRoundsApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['feedback-rounds'] });
      resetForm();
    },
  });

  const startMutation = useMutation({
    mutationFn: (id: string) => feedbackRoundsApi.start(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['feedback-rounds'] }),
  });

  function resetForm() {
    setShowCreate(false);
    setForm({ name: '', description: '', review_period: '', round_type: 'peer', start_date: '', end_date: '' });
    setQuestions([{ question: '', type: 'text', required: true }]);
  }

  function addQuestion() {
    setQuestions([...questions, { question: '', type: 'text', required: true }]);
  }

  function removeQuestion(index: number) {
    setQuestions(questions.filter((_, i) => i !== index));
  }

  function updateQuestion(index: number, field: string, value: string | boolean) {
    const updated = [...questions];
    (updated[index] as Record<string, unknown>)[field] = value;
    setQuestions(updated);
  }

  function handleCreate() {
    const validQuestions = questions.filter((q) => q.question.trim());
    createMutation.mutate({
      name: form.name,
      description: form.description || undefined,
      review_period: form.review_period,
      round_type: form.round_type,
      start_date: form.start_date,
      end_date: form.end_date,
      questions: validQuestions,
    });
  }

  const rounds = data?.data?.rounds || [];
  const results = resultsData?.data?.results || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <MessageCircle className="h-6 w-6 text-blue-400" />
          <h1 className="text-2xl font-bold">360 Feedback Rounds</h1>
        </div>
        <button
          onClick={() => setShowCreate(true)}
          className="flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          <Plus className="h-4 w-4" />
          Create Round
        </button>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : rounds.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <MessageCircle className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">No feedback rounds yet</p>
        </div>
      ) : (
        <div className="space-y-3">
          {rounds.map((round) => (
            <div key={round.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="font-medium text-gray-100">{round.name}</h3>
                  <div className="mt-1 flex items-center gap-2 text-xs text-gray-500">
                    <span className={`rounded-full px-2 py-0.5 ${statusColors[round.status] || ''}`}>{round.status}</span>
                    <span>{round.round_type}</span>
                    <span>{round.review_period}</span>
                    <span>{new Date(round.start_date).toLocaleDateString()} - {new Date(round.end_date).toLocaleDateString()}</span>
                  </div>
                  {round.description && <p className="mt-1 text-sm text-gray-400">{round.description}</p>}
                  <p className="mt-1 text-xs text-gray-500">{round.questions.length} question{round.questions.length !== 1 ? 's' : ''}</p>
                </div>
                <div className="flex items-center gap-2">
                  {round.status === 'draft' && (
                    <button
                      onClick={() => startMutation.mutate(round.id)}
                      className="flex items-center gap-2 rounded-lg bg-green-600/10 px-3 py-2 text-sm text-green-400 hover:bg-green-600/20"
                    >
                      <Play className="h-4 w-4" />
                      Start
                    </button>
                  )}
                  {(round.status === 'completed' || round.status === 'collecting') && (
                    <button
                      onClick={() => setShowResults(showResults === round.id ? null : round.id)}
                      className="flex items-center gap-2 rounded-lg bg-gray-800 px-3 py-2 text-sm text-gray-300 hover:bg-gray-700"
                    >
                      <BarChart3 className="h-4 w-4" />
                      Results
                    </button>
                  )}
                </div>
              </div>

              {showResults === round.id && (
                <div className="mt-4 space-y-3 border-t border-gray-800 pt-4">
                  <h4 className="text-sm font-medium text-gray-300">Results</h4>
                  {results.length === 0 ? (
                    <p className="text-sm text-gray-500">No results available yet</p>
                  ) : (
                    results.map((r, i) => (
                      <div key={i} className="rounded-lg bg-gray-800 p-3">
                        <p className="text-sm font-medium text-gray-200">Reviewee: {r.reviewee_id}</p>
                        <div className="mt-2 flex flex-wrap gap-3">
                          {Object.entries(r.avg_ratings).map(([q, rating]) => (
                            <div key={q} className="rounded bg-gray-700 px-2 py-1 text-xs">
                              <span className="text-gray-400">{q}: </span>
                              <span className="font-medium text-blue-400">{rating.toFixed(1)}</span>
                            </div>
                          ))}
                        </div>
                        {r.comments.length > 0 && (
                          <div className="mt-2">
                            <p className="text-xs text-gray-500">Comments:</p>
                            {r.comments.map((c, ci) => (
                              <p key={ci} className="mt-1 text-sm text-gray-400">&ldquo;{c}&rdquo;</p>
                            ))}
                          </div>
                        )}
                      </div>
                    ))
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {showCreate && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Create Feedback Round</h2>
            <input
              type="text"
              placeholder="Round name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <textarea
              placeholder="Description (optional)"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={2}
            />
            <input
              type="text"
              placeholder="Review period (e.g. Q1 2026)"
              value={form.review_period}
              onChange={(e) => setForm({ ...form, review_period: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <select
              value={form.round_type}
              onChange={(e) => setForm({ ...form, round_type: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="peer">Peer Review</option>
              <option value="upward">Upward Review</option>
              <option value="self">Self Review</option>
              <option value="360">360 Review</option>
            </select>
            <div className="mb-3 grid grid-cols-2 gap-3">
              <div>
                <label className="mb-1 block text-xs text-gray-500">Start Date</label>
                <input
                  type="date"
                  value={form.start_date}
                  onChange={(e) => setForm({ ...form, start_date: e.target.value })}
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
                />
              </div>
              <div>
                <label className="mb-1 block text-xs text-gray-500">End Date</label>
                <input
                  type="date"
                  value={form.end_date}
                  onChange={(e) => setForm({ ...form, end_date: e.target.value })}
                  className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
                />
              </div>
            </div>

            <div className="mb-4">
              <div className="mb-2 flex items-center justify-between">
                <label className="text-sm font-medium text-gray-300">Questions</label>
                <button onClick={addQuestion} className="text-xs text-blue-400 hover:text-blue-300">+ Add Question</button>
              </div>
              {questions.map((q, i) => (
                <div key={i} className="mb-2 flex items-start gap-2">
                  <div className="flex-1 space-y-1">
                    <input
                      type="text"
                      placeholder={`Question ${i + 1}`}
                      value={q.question}
                      onChange={(e) => updateQuestion(i, 'question', e.target.value)}
                      className="w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
                    />
                    <div className="flex items-center gap-2">
                      <select
                        value={q.type}
                        onChange={(e) => updateQuestion(i, 'type', e.target.value)}
                        className="rounded border border-gray-700 bg-gray-800 px-2 py-1 text-xs text-gray-300"
                      >
                        <option value="text">Text</option>
                        <option value="rating">Rating (1-5)</option>
                        <option value="boolean">Yes/No</option>
                      </select>
                      <label className="flex items-center gap-1 text-xs text-gray-500">
                        <input
                          type="checkbox"
                          checked={q.required}
                          onChange={(e) => updateQuestion(i, 'required', e.target.checked)}
                          className="rounded"
                        />
                        Required
                      </label>
                    </div>
                  </div>
                  {questions.length > 1 && (
                    <button onClick={() => removeQuestion(i)} className="mt-2 text-gray-500 hover:text-red-400">
                      <Trash2 className="h-4 w-4" />
                    </button>
                  )}
                </div>
              ))}
            </div>

            <div className="flex justify-end gap-3">
              <button onClick={resetForm} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={handleCreate}
                disabled={!form.name.trim() || !form.review_period.trim() || !form.start_date || !form.end_date}
                className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
              >
                Create
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
