import type { CreateDecisionRequest, TeamDecision, UpdateDecisionRequest } from '@/api/decisions';
import { Badge } from '@/components/ui/badge';
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
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Plus, X } from 'lucide-react';
import { useState } from 'react';

interface DecisionFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateDecisionRequest | UpdateDecisionRequest) => void;
  decision?: TeamDecision;
  isLoading?: boolean;
}

export function DecisionForm({
  open,
  onOpenChange,
  onSubmit,
  decision,
  isLoading = false,
}: DecisionFormProps) {
  const [title, setTitle] = useState(decision?.title || '');
  const [description, setDescription] = useState(decision?.description || '');
  const [rationale, setRationale] = useState(decision?.rationale || '');
  const [outcome, setOutcome] = useState(decision?.outcome || '');
  const [alternatives, setAlternatives] = useState<string[]>(decision?.alternatives || []);
  const [tags, setTags] = useState<string[]>(decision?.tags || []);
  const [newAlternative, setNewAlternative] = useState('');
  const [newTag, setNewTag] = useState('');
  const [importanceScore, setImportanceScore] = useState(decision?.importance_score || 0.5);

  const isEditing = !!decision;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (isEditing) {
      const updateData: UpdateDecisionRequest = {};
      if (title !== decision.title) updateData.title = title;
      if (description !== decision.description) updateData.description = description;
      if (rationale !== decision.rationale) updateData.rationale = rationale;
      if (outcome !== decision.outcome) updateData.outcome = outcome;
      if (JSON.stringify(alternatives) !== JSON.stringify(decision.alternatives))
        updateData.alternatives = alternatives;
      if (JSON.stringify(tags) !== JSON.stringify(decision.tags)) updateData.tags = tags;
      if (importanceScore !== decision.importance_score)
        updateData.importance_score = importanceScore;
      onSubmit(updateData);
    } else {
      const createData: CreateDecisionRequest = {
        title,
        description,
        rationale,
        outcome,
        alternatives,
        tags,
        importance_score: importanceScore,
      };
      onSubmit(createData);
    }
  };

  const addAlternative = () => {
    if (newAlternative.trim()) {
      setAlternatives([...alternatives, newAlternative.trim()]);
      setNewAlternative('');
    }
  };

  const removeAlternative = (index: number) => {
    setAlternatives(alternatives.filter((_, i) => i !== index));
  };

  const addTag = () => {
    if (newTag.trim() && !tags.includes(newTag.trim())) {
      setTags([...tags, newTag.trim()]);
      setNewTag('');
    }
  };

  const removeTag = (tag: string) => {
    setTags(tags.filter((t) => t !== tag));
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      // Reset form on close
      setTitle(decision?.title || '');
      setDescription(decision?.description || '');
      setRationale(decision?.rationale || '');
      setOutcome(decision?.outcome || '');
      setAlternatives(decision?.alternatives || []);
      setTags(decision?.tags || []);
      setImportanceScore(decision?.importance_score || 0.5);
      setNewAlternative('');
      setNewTag('');
    }
    onOpenChange(newOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>{isEditing ? 'Edit Decision' : 'Record a New Decision'}</DialogTitle>
            <DialogDescription>
              {isEditing
                ? 'Update the details of this decision.'
                : 'Capture what was decided, why, and who approved it.'}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {/* Title */}
            <div className="grid gap-2">
              <Label htmlFor="title">
                Decision Title <span className="text-destructive">*</span>
              </Label>
              <Input
                id="title"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
                placeholder="e.g., Choose Stripe as our payment processor"
                required
                minLength={3}
                maxLength={500}
              />
            </div>

            {/* Description */}
            <div className="grid gap-2">
              <Label htmlFor="description">Summary</Label>
              <Textarea
                id="description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="A brief summary of what this decision is about..."
                rows={3}
              />
            </div>

            {/* Rationale */}
            <div className="grid gap-2">
              <Label htmlFor="rationale">Why This Decision?</Label>
              <Textarea
                id="rationale"
                value={rationale}
                onChange={(e) => setRationale(e.target.value)}
                placeholder="Explain the reasoning behind this decision..."
                rows={3}
              />
            </div>

            {/* Outcome */}
            <div className="grid gap-2">
              <Label htmlFor="outcome">What Was Decided</Label>
              <Textarea
                id="outcome"
                value={outcome}
                onChange={(e) => setOutcome(e.target.value)}
                placeholder="What specifically was decided or agreed upon..."
                rows={2}
              />
            </div>

            {/* Alternatives */}
            <div className="grid gap-2">
              <Label>Alternatives Considered</Label>
              <div className="flex gap-2">
                <Input
                  value={newAlternative}
                  onChange={(e) => setNewAlternative(e.target.value)}
                  placeholder="e.g., PayPal, Square"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      addAlternative();
                    }
                  }}
                />
                <Button type="button" variant="outline" onClick={addAlternative}>
                  <Plus className="h-4 w-4" />
                </Button>
              </div>
              {alternatives.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {alternatives.map((alt, i) => (
                    <Badge key={i} variant="secondary" className="pr-1">
                      {alt}
                      <button
                        type="button"
                        onClick={() => removeAlternative(i)}
                        className="ml-1 hover:text-destructive"
                      >
                        <X className="h-3 w-3" />
                      </button>
                    </Badge>
                  ))}
                </div>
              )}
            </div>

            {/* Tags */}
            <div className="grid gap-2">
              <Label>Tags</Label>
              <div className="flex gap-2">
                <Input
                  value={newTag}
                  onChange={(e) => setNewTag(e.target.value)}
                  placeholder="e.g., billing, infrastructure"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      e.preventDefault();
                      addTag();
                    }
                  }}
                />
                <Button type="button" variant="outline" onClick={addTag}>
                  <Plus className="h-4 w-4" />
                </Button>
              </div>
              {tags.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-2">
                  {tags.map((tag) => (
                    <Badge key={tag} variant="outline" className="pr-1">
                      {tag}
                      <button
                        type="button"
                        onClick={() => removeTag(tag)}
                        className="ml-1 hover:text-destructive"
                      >
                        <X className="h-3 w-3" />
                      </button>
                    </Badge>
                  ))}
                </div>
              )}
            </div>

            {/* Importance Score */}
            <div className="grid gap-2">
              <Label htmlFor="importance">
                Importance:{' '}
                {importanceScore < 0.3 ? 'Low' : importanceScore < 0.7 ? 'Medium' : 'High'}
              </Label>
              <input
                id="importance"
                type="range"
                min="0"
                max="1"
                step="0.1"
                value={importanceScore}
                onChange={(e) => setImportanceScore(parseFloat(e.target.value))}
                className="w-full"
              />
            </div>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={() => handleOpenChange(false)}>
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading || !title.trim()}>
              {isLoading ? 'Saving...' : isEditing ? 'Save Changes' : 'Record Decision'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
