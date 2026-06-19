import { useCallback, useEffect, useState } from 'react';
import { toast } from 'sonner';
import { DRAFT_KEY } from '../constants';
import type { DraftData, StoredDraft } from '../types';

interface UseComposerDraftOptions {
  description: string;
  constraints: string;
  runtime: string;
  onRestore: (draft: DraftData) => void;
}

export function useComposerDraft({
  description,
  constraints,
  runtime,
  onRestore,
}: UseComposerDraftOptions) {
  const [lastSaved, setLastSaved] = useState<Date | null>(null);
  const [hasDraft, setHasDraft] = useState(false);

  useEffect(() => {
    try {
      const draftJson = localStorage.getItem(DRAFT_KEY);
      if (draftJson) {
        const draft: DraftData = JSON.parse(draftJson);
        const maxAge = 7 * 24 * 60 * 60 * 1000;
        if (Date.now() - draft.timestamp < maxAge) {
          onRestore(draft);
          setHasDraft(true);
          setLastSaved(new Date(draft.timestamp));
          toast.info('Restored draft from ' + new Date(draft.timestamp).toLocaleDateString());
        } else {
          localStorage.removeItem(DRAFT_KEY);
        }
      }
    } catch (error) {
      console.error('Failed to load draft:', error);
    }
    // Only restore on mount
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    const saveDraft = () => {
      if (description || constraints) {
        const draft: StoredDraft = {
          description,
          constraints,
          runtime,
          timestamp: Date.now(),
        };
        try {
          localStorage.setItem(DRAFT_KEY, JSON.stringify(draft));
          setLastSaved(new Date());
          setHasDraft(true);
        } catch (error) {
          console.error('Failed to save draft:', error);
        }
      }
    };

    const timeout = setTimeout(saveDraft, 1000);
    return () => clearTimeout(timeout);
  }, [description, constraints, runtime]);

  const clearDraft = useCallback(() => {
    localStorage.removeItem(DRAFT_KEY);
    setHasDraft(false);
    setLastSaved(null);
    toast.success('Draft cleared');
  }, []);

  return { lastSaved, hasDraft, clearDraft };
}
