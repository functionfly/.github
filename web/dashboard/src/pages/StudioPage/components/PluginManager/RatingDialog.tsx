import { useState, useEffect } from "react";
import { GlassCard, Button, Spinner } from "@functionfly/ui-core";
import { Star, X } from "lucide-react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { marketplaceApi, type Extension, type Rating } from "@/api/marketplace";
import { toast } from "sonner";
import { cn } from "@/lib/utils";

interface RatingDialogProps {
  extension: Extension;
  onClose: () => void;
}

export function RatingDialog({ extension, onClose }: RatingDialogProps) {
  const [rating, setRating] = useState(0);
  const [hoverRating, setHoverRating] = useState(0);
  const [review, setReview] = useState("");
  const queryClient = useQueryClient();

  const { data: myRatingData } = useQuery({
    queryKey: ["marketplace-my-rating", extension.id],
    queryFn: () => marketplaceApi.getMyRating(extension.id),
  });

  useEffect(() => {
    if (myRatingData?.rating) {
      setRating(myRatingData.rating.rating);
      setReview(myRatingData.rating.review || "");
    }
  }, [myRatingData]);

  const rateMutation = useMutation({
    mutationFn: () => marketplaceApi.rate(extension.id, rating, review),
    onSuccess: () => {
      toast.success(`Rated "${extension.name}" ${rating} stars`);
      queryClient.invalidateQueries({ queryKey: ["marketplace-my-rating", extension.id] });
      queryClient.invalidateQueries({ queryKey: ["marketplace-extensions"] });
      onClose();
    },
    onError: (error: Error) => {
      toast.error(`Failed to submit rating: ${error.message}`);
    },
  });

  const handleSubmit = () => {
    if (rating < 1 || rating > 5) {
      toast.error("Please select a rating between 1 and 5 stars");
      return;
    }
    rateMutation.mutate();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
      <GlassCard className="w-full max-w-md p-6 space-y-4">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold text-white">Rate {extension.name}</h3>
            <p className="text-sm text-white/60">v{extension.version}</p>
          </div>
          <button onClick={onClose} className="p-1 rounded text-white/60 hover:text-white hover:bg-white/10">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="flex flex-col items-center gap-2 py-4">
          <div className="flex items-center gap-1">
            {[1, 2, 3, 4, 5].map((star) => (
              <button
                key={star}
                type="button"
                onClick={() => setRating(star)}
                onMouseEnter={() => setHoverRating(star)}
                onMouseLeave={() => setHoverRating(0)}
                className="p-1 transition-transform hover:scale-110"
              >
                <Star
                  className={cn(
                    "w-8 h-8 transition-colors",
                    (hoverRating || rating) >= star
                      ? "fill-yellow-400 text-yellow-400"
                      : "text-white/30"
                  )}
                />
              </button>
            ))}
          </div>
          <p className="text-sm text-white/60">
            {rating === 0 ? "Click to rate" : `${rating} of 5 stars`}
          </p>
        </div>

        <div>
          <label className="block text-sm font-medium text-white/80 mb-2">
            Review (optional)
          </label>
          <textarea
            value={review}
            onChange={(e) => setReview(e.target.value)}
            placeholder="Share your experience with this plugin..."
            className="w-full h-24 bg-white/5 border border-white/10 rounded-lg px-3 py-2 text-white text-sm resize-none focus:outline-none focus:border-white/30"
            maxLength={1000}
          />
          <p className="text-xs text-white/40 mt-1">{review.length}/1000</p>
        </div>

        <div className="flex items-center justify-end gap-2 pt-2">
          <Button variant="outline" onClick={onClose} disabled={rateMutation.isPending}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={rateMutation.isPending || rating === 0}
            className="bg-gradient-to-r from-orange-500 to-red-500 hover:from-orange-400 hover:to-red-400"
          >
            {rateMutation.isPending ? <Spinner className="w-4 h-4" /> : "Submit Rating"}
          </Button>
        </div>
      </GlassCard>
    </div>
  );
}
