import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { conversationsApi } from "@/api/conversations";
import { toast } from "sonner";

export interface BountyAttachModalProps {
  username: string;
  conversationId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function BountyAttachModal({
  username,
  conversationId,
  open,
  onOpenChange,
}: BountyAttachModalProps) {
  const queryClient = useQueryClient();
  const [amountReputation, setAmountReputation] = useState(50);
  const [amountCents, setAmountCents] = useState(0);
  const [securityWeight, setSecurityWeight] = useState(1);

  const createBounty = useMutation({
    mutationFn: () =>
      conversationsApi.createBounty(username, conversationId, {
        amount_reputation: amountReputation,
        amount_cents: amountCents || undefined,
        security_weight_multiplier: securityWeight,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["conversation", username, conversationId] });
      queryClient.invalidateQueries({ queryKey: ["conversation-bounties", username, conversationId] });
      toast.success("Bounty attached");
      onOpenChange(false);
    },
    onError: (err: Error) => {
      toast.error(err.message || "Failed to attach bounty");
    },
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Attach reputation bounty</DialogTitle>
          <DialogDescription>
            Offer reputation or a cash bounty to incentivize help. Claimed when the conversation is
            resolved.
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4 py-2">
          <div>
            <Label>Reputation points</Label>
            <Input
              type="number"
              min={0}
              value={amountReputation}
              onChange={(e) => setAmountReputation(Number(e.target.value) || 0)}
            />
          </div>
          <div>
            <Label>Amount ($ cents, optional)</Label>
            <Input
              type="number"
              min={0}
              value={amountCents || ""}
              onChange={(e) => setAmountCents(Number(e.target.value) || 0)}
              placeholder="0"
            />
          </div>
          <div>
            <Label>Security weight multiplier</Label>
            <Input
              type="number"
              min={0.1}
              step={0.1}
              value={securityWeight}
              onChange={(e) => setSecurityWeight(Number(e.target.value) || 1)}
            />
          </div>
          <Button
            onClick={() => createBounty.mutate()}
            disabled={createBounty.isPending || amountReputation < 0}
          >
            {createBounty.isPending ? "Attaching…" : "Attach bounty"}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
