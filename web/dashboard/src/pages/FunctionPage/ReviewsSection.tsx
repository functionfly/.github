import { registryApi, type RegistryFunctionReview } from '@/api/registry';
import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { useAuthStore } from '@/stores/authStore';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { formatDistanceToNow } from 'date-fns';
import { motion } from 'framer-motion';
import { Loader2, Star } from 'lucide-react';
import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';

const REVIEWS_PER_PAGE = 20;

export function ReviewsSection() {
  const { author, name } = useParams<{ author: string; name: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [stars, setStars] = useState(5);
  const [title, setTitle] = useState('');
  const [body, setBody] = useState('');
  const [offset, setOffset] = useState(0);

  const { data: reviewsResp, isLoading, isFetching } = useQuery({
    queryKey: ['function-reviews', author, name, offset],
    queryFn: async () => {
      if (!author || !name) return null;
      return registryApi.listReviews(author, name, { limit: REVIEWS_PER_PAGE, offset });
    },
    enabled: !!author && !!name,
  });

  const reviews: RegistryFunctionReview[] = reviewsResp?.reviews ?? [];
  const total = reviewsResp?.total ?? reviews.length;
  const hasMore = reviews.length < total;

  const submitMutation = useMutation({
    mutationFn: async () => {
      if (!author || !name) throw new Error('Missing function');
      return registryApi.submitReview(author, name, { stars, title, body });
    },
    onSuccess: () => {
      toast.success('Review submitted');
      setDialogOpen(false);
      setStars(5);
      setTitle('');
      setBody('');
      setOffset(0);
      queryClient.invalidateQueries({ queryKey: ['function-reviews', author, name] });
      registryApi
        .submitRating(author!, name!, {
          overall_score: stars,
          reliability_score: 0,
          latency_score: 0,
          documentation_score: 0,
        })
        .catch(() => {});
    },
    onError: (e) => {
      const msg = (e as Error)?.message || 'Failed to submit review';
      if (msg.toLowerCase().includes('execute this function')) {
        toast.info('Run it once to unlock reviews', {
          action: {
            label: 'Try it now',
            onClick: () => navigate(`/run/${author}/${name}`),
          },
        });
        return;
      }
      toast.error(msg);
    },
  });

  const handleLoadMore = () => {
    setOffset((prev) => prev + REVIEWS_PER_PAGE);
  };

  return (
    <>
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="function-page-section function-page-section--delayed-3"
      >
        <div className="function-page-reviews-header">
          <div>
            <h2 className="function-page-reviews-title">Community reviews</h2>
            <p className="function-page-reviews-count">
              {total} {total === 1 ? 'review' : 'reviews'}
            </p>
          </div>
          <Button
            variant="outline"
            onClick={() => {
              if (!isAuthenticated) {
                toast.info('Sign in to leave a review');
                navigate(`/login?redirect=${encodeURIComponent(window.location.pathname)}`);
                return;
              }
              setDialogOpen(true);
            }}
            className="gap-2"
          >
            <Star className="h-4 w-4" />
            Write a review
          </Button>
        </div>

        <div className="function-page-reviews-list">
          {isLoading ? (
            <div className="function-page-loading">Loading reviews…</div>
          ) : reviews.length === 0 ? (
            <div className="function-page-reviews-empty">
              <p className="function-page-reviews-empty-text">No reviews yet. Be the first.</p>
            </div>
          ) : (
            <>
              {reviews.map((r) => (
                <div key={r.id} className="function-page-review-card">
                  <div className="function-page-review-header">
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <div className="function-page-review-stars">
                          {Array.from({ length: 5 }).map((_, i) => (
                            <Star
                              key={i}
                              className={`function-page-review-star ${
                                i < (r.stars ?? 0)
                                  ? 'function-page-review-star--filled'
                                  : 'function-page-review-star--empty'
                              }`}
                            />
                          ))}
                        </div>
                        <span className="function-page-review-date">
                          {formatDistanceToNow(new Date(r.created_at), { addSuffix: true })}
                        </span>
                      </div>
                      {r.title && <div className="function-page-review-title">{r.title}</div>}
                      {r.body && <div className="function-page-review-body">{r.body}</div>}
                    </div>
                  </div>
                </div>
              ))}
              {hasMore && (
                <div className="flex justify-center mt-4">
                  <Button
                    variant="outline"
                    onClick={handleLoadMore}
                    disabled={isFetching}
                    className="gap-2"
                  >
                    {isFetching && <Loader2 className="h-4 w-4 animate-spin" />}
                    Load more reviews
                  </Button>
                </div>
              )}
            </>
          )}
        </div>
      </motion.div>

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>Write a review</DialogTitle>
            <DialogDescription>
              Share your experience. Your star rating will also contribute to the function's
              community score.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <div className="text-sm font-medium text-foreground mb-2">Rating</div>
              <div className="flex items-center gap-1">
                {Array.from({ length: 5 }).map((_, i) => {
                  const val = i + 1;
                  const active = val <= stars;
                  return (
                    <button
                      key={val}
                      type="button"
                      onClick={() => setStars(val)}
                      className="p-1 rounded-md hover:bg-bg-tertiary transition-colors"
                      aria-label={`${val} star${val === 1 ? '' : 's'}`}
                    >
                      <Star
                        className={`h-5 w-5 ${
                          active ? 'fill-amber-400 text-amber-400' : 'text-muted-foreground/50'
                        }`}
                      />
                    </button>
                  );
                })}
                <span className="ml-2 text-sm text-muted-foreground">{stars}/5</span>
              </div>
            </div>

            <div className="space-y-2">
              <div className="text-sm font-medium text-foreground">Title (optional)</div>
              <Input
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="Short summary"
                maxLength={120}
              />
            </div>

            <div className="space-y-2">
              <div className="text-sm font-medium text-foreground">Review (optional)</div>
              <Textarea
                value={body}
                onChange={(e) => setBody(e.target.value)}
                placeholder="What worked well? Anything to watch out for?"
                maxLength={5000}
              />
            </div>
          </div>

          <DialogFooter className="gap-2">
            <Button variant="outline" onClick={() => setDialogOpen(false)} className="gap-2">
              Cancel
            </Button>
            <Button
              onClick={() => submitMutation.mutate()}
              disabled={submitMutation.isPending}
              className="gap-2"
            >
              {submitMutation.isPending ? (
                <>
                  <Loader2 className="h-4 w-4 animate-spin" />
                  Submitting…
                </>
              ) : (
                'Submit'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}