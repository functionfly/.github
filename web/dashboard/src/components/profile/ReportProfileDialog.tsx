import { usersApi } from '@/api/users';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { useMutation } from '@tanstack/react-query';
import axios from 'axios';
import { ExternalLink, Flag, Loader2 } from 'lucide-react';
import { useEffect, useId, useState } from 'react';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';

export const PROFILE_REPORT_REASONS = [
  { value: 'tos_violation', label: 'Violates Terms of Service' },
  { value: 'harassment', label: 'Harassment or abuse' },
  { value: 'spam', label: 'Spam or misleading content' },
  { value: 'impersonation', label: 'Impersonation' },
  { value: 'other', label: 'Other (describe below)' },
] as const;

export type ProfileReportReason = (typeof PROFILE_REPORT_REASONS)[number]['value'];

export interface ReportProfileDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  username: string;
  displayName: string;
}

export function ReportProfileDialog({
  open,
  onOpenChange,
  username,
  displayName,
}: ReportProfileDialogProps) {
  const formId = useId();
  const [reason, setReason] = useState<ProfileReportReason>('tos_violation');
  const [details, setDetails] = useState('');
  const [ack, setAck] = useState(false);

  useEffect(() => {
    if (!open) {
      setReason('tos_violation');
      setDetails('');
      setAck(false);
    }
  }, [open]);

  const mutation = useMutation({
    mutationFn: () =>
      usersApi.reportProfile(username, {
        reason,
        details: details.trim(),
        acknowledged_accuracy: ack,
      }),
    onSuccess: (res) => {
      toast.success(res.message || 'Report submitted. Thank you.');
      onOpenChange(false);
    },
    onError: (err: unknown) => {
      let msg = 'Could not submit report. Please try again.';
      if (axios.isAxiosError(err)) {
        const data = err.response?.data as { error?: string } | undefined;
        if (data?.error) msg = data.error;
      } else if (err instanceof Error && err.message) msg = err.message;
      toast.error(msg);
    },
  });

  const canSubmit = ack && (reason !== 'other' || details.trim().length >= 20);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Flag className="h-5 w-5 text-red-500" aria-hidden="true" />
            Report profile
          </DialogTitle>
          <DialogDescription>
            Report <span className="font-medium text-foreground">@{username}</span>
            {displayName && displayName !== username ? (
              <span className="text-muted-foreground"> ({displayName})</span>
            ) : null}{' '}
            if you believe they violate our{' '}
            <Link
              to="/terms"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-0.5 text-brand-500 hover:underline"
            >
              Terms of Service
              <ExternalLink className="h-3 w-3" aria-hidden="true" />
            </Link>{' '}
            or community guidelines. False reports may result in action on your account.
          </DialogDescription>
        </DialogHeader>

        <form
          id={formId}
          className="space-y-4"
          onSubmit={(e) => {
            e.preventDefault();
            if (!canSubmit || mutation.isPending) return;
            mutation.mutate();
          }}
        >
          <div className="space-y-2">
            <Label htmlFor={`${formId}-reason`}>Reason</Label>
            <Select
              value={reason}
              onValueChange={(v) => setReason(v as ProfileReportReason)}
            >
              <SelectTrigger id={`${formId}-reason`}>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {PROFILE_REPORT_REASONS.map((r) => (
                  <SelectItem key={r.value} value={r.value}>
                    {r.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor={`${formId}-details`}>
              Details {reason === 'other' ? '(required, min 20 characters)' : '(optional)'}
            </Label>
            <Textarea
              id={`${formId}-details`}
              value={details}
              onChange={(e) => setDetails(e.target.value)}
              placeholder="What did you observe? Include links or context if helpful."
              rows={4}
              maxLength={4000}
              className="resize-y min-h-[100px]"
            />
            <p className="text-xs text-muted-foreground text-right">{details.length}/4000</p>
          </div>

          <div className="flex items-start gap-3 rounded-lg border border-border-subtle bg-bg-secondary/40 p-3">
            <Checkbox
              id={`${formId}-ack`}
              checked={ack}
              onCheckedChange={(v) => setAck(v === true)}
              className="mt-0.5"
            />
            <Label htmlFor={`${formId}-ack`} className="text-sm font-normal leading-snug cursor-pointer">
              I confirm this report is accurate to the best of my knowledge and submitted in good faith.
            </Label>
          </div>
        </form>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button type="button" variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button
            type="submit"
            form={formId}
            disabled={!canSubmit || mutation.isPending}
            className="bg-red-600 hover:bg-red-700 text-white"
          >
            {mutation.isPending ? (
              <>
                <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                Submitting…
              </>
            ) : (
              'Submit report'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
