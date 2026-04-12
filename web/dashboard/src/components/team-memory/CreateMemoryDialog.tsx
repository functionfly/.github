import { useState } from 'react';
import { Brain, CheckCircle, Filter, User, Shield } from 'lucide-react';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Checkbox } from '@/components/ui/checkbox';
import { useCreateMemory } from '@/hooks/use-team-memory';
import { initTeamMemoryCrypto, getTeamMemoryCrypto } from '@/utils/team-memory-crypto';

interface CreateMemoryDialogProps {
  teamId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const memoryTypes = [
  { value: 'decision', label: 'Decision', icon: CheckCircle, description: 'Team decisions and rationale' },
  { value: 'preference', label: 'Preference', icon: Brain, description: 'Preferences and requirements' },
  { value: 'process', label: 'Process', icon: Filter, description: 'Workflows and procedures' },
  { value: 'client_context', label: 'Client Context', icon: User, description: 'Client-specific information' },
];

export function CreateMemoryDialog({ teamId, open, onOpenChange }: CreateMemoryDialogProps) {
  const [type, setType] = useState('decision');
  const [summary, setSummary] = useState('');
  const [category, setCategory] = useState('');
  const [isEncrypted, setIsEncrypted] = useState(false);
  
  // Type-specific fields
  const [decisionRationale, setDecisionRationale] = useState('');
  const [decisionMaker, setDecisionMaker] = useState('');
  const [prefSubject, setPrefSubject] = useState('');
  const [prefValue, setPrefValue] = useState('');
  const [prefContext, setPrefContext] = useState('');
  const [processName, setProcessName] = useState('');
  const [processSteps, setProcessSteps] = useState('');
  const [clientName, setClientName] = useState('');
  const [clientNotes, setClientNotes] = useState('');

  const createMemory = useCreateMemory(teamId);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    let content: Record<string, any> = {};
    
    switch (type) {
      case 'decision':
        content = {
          title: summary,
          rationale: decisionRationale,
          decision_maker: decisionMaker,
        };
        break;
      case 'preference':
        content = {
          subject: prefSubject,
          value: prefValue,
          context: prefContext,
        };
        break;
      case 'process':
        content = {
          name: processName,
          steps: processSteps.split('\n').filter(s => s.trim()),
        };
        break;
      case 'client_context':
        content = {
          client_name: clientName,
          notes: clientNotes,
        };
        break;
    }

    let encryptedData;
    
    if (isEncrypted) {
      try {
        await initTeamMemoryCrypto(teamId);
        const crypto = getTeamMemoryCrypto();
        encryptedData = await crypto.encryptForAPI(content);
        content = {}; // Clear plaintext content
      } catch (error) {
        alert('Failed to encrypt. Make sure you have set up encryption in the Encryption settings.');
        return;
      }
    }

    await createMemory.mutateAsync({
      memory_type: type,
      summary,
      category: category || undefined,
      content,
      is_encrypted: isEncrypted,
      encrypted_data: encryptedData,
    });

    onOpenChange(false);
    resetForm();
  };

  const resetForm = () => {
    setType('decision');
    setSummary('');
    setCategory('');
    setIsEncrypted(false);
    setDecisionRationale('');
    setDecisionMaker('');
    setPrefSubject('');
    setPrefValue('');
    setPrefContext('');
    setProcessName('');
    setProcessSteps('');
    setClientName('');
    setClientNotes('');
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add Team Memory</DialogTitle>
          <DialogDescription>
            Capture knowledge for your team&apos;s shared brain. This will be
            accessible to all team members and agents.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          <Tabs value={type} onValueChange={setType}>
            <TabsList className="grid grid-cols-4">
              {memoryTypes.map((t) => (
                <TabsTrigger key={t.value} value={t.value} className="gap-2">
                  <t.icon className="h-4 w-4" />
                  <span className="hidden sm:inline">{t.label}</span>
                </TabsTrigger>
              ))}
            </TabsList>

            {memoryTypes.map((t) => (
              <TabsContent key={t.value} value={t.value}>
                <p className="text-sm text-muted-foreground mb-4">
                  {t.description}
                </p>
              </TabsContent>
            ))}
          </Tabs>

          <div className="space-y-4">
            <div>
              <Label htmlFor="summary">Summary *</Label>
              <Input
                id="summary"
                placeholder="Brief description of this memory"
                value={summary}
                onChange={(e) => setSummary(e.target.value)}
                required
              />
            </div>

            <div>
              <Label htmlFor="category">Category (optional)</Label>
              <Input
                id="category"
                placeholder="e.g., client:acme-corp, process:onboarding"
                value={category}
                onChange={(e) => setCategory(e.target.value)}
              />
            </div>

            {type === 'decision' && (
              <>
                <div>
                  <Label htmlFor="rationale">Rationale</Label>
                  <Textarea
                    id="rationale"
                    placeholder="Why was this decision made?"
                    value={decisionRationale}
                    onChange={(e) => setDecisionRationale(e.target.value)}
                    rows={4}
                  />
                </div>
                <div>
                  <Label htmlFor="decision-maker">Decision Maker</Label>
                  <Input
                    id="decision-maker"
                    placeholder="Who made this decision?"
                    value={decisionMaker}
                    onChange={(e) => setDecisionMaker(e.target.value)}
                  />
                </div>
              </>
            )}

            {type === 'preference' && (
              <>
                <div>
                  <Label htmlFor="subject">Subject *</Label>
                  <Input
                    id="subject"
                    placeholder="e.g., email_style, meeting_times"
                    value={prefSubject}
                    onChange={(e) => setPrefSubject(e.target.value)}
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="value">Value *</Label>
                  <Input
                    id="value"
                    placeholder="e.g., short, mornings_only"
                    value={prefValue}
                    onChange={(e) => setPrefValue(e.target.value)}
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="pref-context">Context</Label>
                  <Input
                    id="pref-context"
                    placeholder="e.g., for_client_emails"
                    value={prefContext}
                    onChange={(e) => setPrefContext(e.target.value)}
                  />
                </div>
              </>
            )}

            {type === 'process' && (
              <>
                <div>
                  <Label htmlFor="process-name">Process Name *</Label>
                  <Input
                    id="process-name"
                    placeholder="e.g., Weekly Review, Client Onboarding"
                    value={processName}
                    onChange={(e) => setProcessName(e.target.value)}
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="process-steps">Steps (one per line)</Label>
                  <Textarea
                    id="process-steps"
                    placeholder="1. First step&#10;2. Second step&#10;3. Third step"
                    value={processSteps}
                    onChange={(e) => setProcessSteps(e.target.value)}
                    rows={4}
                  />
                </div>
              </>
            )}

            {type === 'client_context' && (
              <>
                <div>
                  <Label htmlFor="client-name">Client Name *</Label>
                  <Input
                    id="client-name"
                    value={clientName}
                    onChange={(e) => setClientName(e.target.value)}
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="client-notes">Notes</Label>
                  <Textarea
                    id="client-notes"
                    placeholder="Important context about this client..."
                    value={clientNotes}
                    onChange={(e) => setClientNotes(e.target.value)}
                    rows={4}
                  />
                </div>
              </>
            )}

            <div className="flex items-center space-x-2 pt-4 border-t">
              <Checkbox
                id="encrypt"
                checked={isEncrypted}
                onCheckedChange={(checked) => setIsEncrypted(checked as boolean)}
              />
              <Label htmlFor="encrypt" className="flex items-center gap-2 cursor-pointer">
                <Shield className="h-4 w-4" />
                Encrypt this memory (client-side encryption)
              </Label>
            </div>
            {isEncrypted && (
              <p className="text-sm text-amber-600 bg-amber-50 p-2 rounded">
                Encrypted memories require the team passphrase to view. Make sure
                team members have the passphrase.
              </p>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={createMemory.isPending}>
              {createMemory.isPending ? 'Creating...' : 'Create Memory'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
