import { Button } from '@/components/ui/button';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Clock, DollarSign, Zap, Info, TrendingUp, Calendar } from 'lucide-react';
import { useState } from 'react';

interface Bundle {
  id: string;
  slug: string;
  display_name: string;
  description: string;
}

interface FounderModeModalProps {
  isOpen: boolean;
  onClose: () => void;
  bundle: Bundle | null;
  onSubmit: (modeType: string, freeDays: number, mrrThreshold: number) => void;
  loading: boolean;
}

type FounderModeType = 'time_based' | 'revenue_based' | 'hybrid';

export function FounderModeModal({
  isOpen,
  onClose,
  bundle,
  onSubmit,
  loading,
}: FounderModeModalProps) {
  const [modeType, setModeType] = useState<FounderModeType>('hybrid');

  const handleSubmit = () => {
    if (!bundle) return;

    let freeDays = 90;
    let mrrThreshold = 1000; // $1000

    if (modeType === 'time_based') {
      freeDays = 90;
      mrrThreshold = 0;
    } else if (modeType === 'revenue_based') {
      freeDays = 0;
      mrrThreshold = 1000;
    } else {
      // hybrid
      freeDays = 90;
      mrrThreshold = 1000;
    }

    onSubmit(modeType, freeDays, mrrThreshold);
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-lg">
        <DialogHeader>
          <div className="flex items-center gap-2 mb-2">
            <div className="w-10 h-10 bg-gradient-to-r from-violet-500 to-fuchsia-500 rounded-lg flex items-center justify-center">
              <Zap className="w-5 h-5 text-white" />
            </div>
            <div>
              <DialogTitle className="text-xl">Founder Mode</DialogTitle>
              <DialogDescription>
                {bundle?.display_name || 'Bundle'}
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Intro text */}
          <div className="bg-gradient-to-r from-violet-50 to-fuchsia-50 dark:from-violet-900/20 dark:to-fuchsia-900/20 rounded-lg p-4">
            <p className="text-sm text-slate-700 dark:text-slate-300">
              Start building <strong>completely free</strong>. We'll only start billing when you hit any of these triggers:
            </p>
            <ul className="mt-2 space-y-1 text-sm text-slate-600 dark:text-slate-400">
              <li className="flex items-center gap-2">
                <TrendingUp className="w-4 h-4 text-violet-500" />
                100 users signed up
              </li>
              <li className="flex items-center gap-2">
                <DollarSign className="w-4 h-4 text-green-500" />
                $1,000 MRR reached
              </li>
              <li className="flex items-center gap-2">
                <Calendar className="w-4 h-4 text-blue-500" />
                3 months elapsed
              </li>
            </ul>
          </div>

          {/* Mode selection */}
          <RadioGroup
            value={modeType}
            onValueChange={(value) => setModeType(value as FounderModeType)}
            className="space-y-3"
          >
            <div
              className={`flex items-start gap-3 p-4 rounded-lg border-2 cursor-pointer transition-colors ${
                modeType === 'hybrid'
                  ? 'border-violet-500 bg-violet-50 dark:bg-violet-900/20'
                  : 'border-slate-200 dark:border-slate-700 hover:border-violet-200'
              }`}
              onClick={() => setModeType('hybrid')}
            >
              <RadioGroupItem value="hybrid" id="hybrid" className="mt-1" />
              <div className="flex-1">
                <Label htmlFor="hybrid" className="font-semibold cursor-pointer flex items-center gap-2">
                  <Zap className="w-4 h-4 text-violet-500" />
                  Hybrid (Recommended)
                </Label>
                <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
                  Free for 3 months OR until $1K MRR, whichever comes first
                </p>
              </div>
            </div>

            <div
              className={`flex items-start gap-3 p-4 rounded-lg border-2 cursor-pointer transition-colors ${
                modeType === 'time_based'
                  ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                  : 'border-slate-200 dark:border-slate-700 hover:border-blue-200'
              }`}
              onClick={() => setModeType('time_based')}
            >
              <RadioGroupItem value="time_based" id="time_based" className="mt-1" />
              <div className="flex-1">
                <Label htmlFor="time_based" className="font-semibold cursor-pointer flex items-center gap-2">
                  <Clock className="w-4 h-4 text-blue-500" />
                  Fixed Time
                </Label>
                <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
                  Free for exactly 3 months, then billing starts
                </p>
              </div>
            </div>

            <div
              className={`flex items-start gap-3 p-4 rounded-lg border-2 cursor-pointer transition-colors ${
                modeType === 'revenue_based'
                  ? 'border-green-500 bg-green-50 dark:bg-green-900/20'
                  : 'border-slate-200 dark:border-slate-700 hover:border-green-200'
              }`}
              onClick={() => setModeType('revenue_based')}
            >
              <RadioGroupItem value="revenue_based" id="revenue_based" className="mt-1" />
              <div className="flex-1">
                <Label htmlFor="revenue_based" className="font-semibold cursor-pointer flex items-center gap-2">
                  <DollarSign className="w-4 h-4 text-green-500" />
                  Revenue-Based
                </Label>
                <p className="text-sm text-slate-600 dark:text-slate-400 mt-1">
                  Free until you hit $1K MRR, no time limit
                </p>
              </div>
            </div>
          </RadioGroup>

          {/* Info box */}
          <div className="flex items-start gap-2 text-sm text-slate-500 dark:text-slate-400 bg-slate-50 dark:bg-slate-800 p-3 rounded-lg">
            <Info className="w-4 h-4 flex-shrink-0 mt-0.5" />
            <p>
              When you hit 80% of any threshold, we'll email you a warning. At 100%, you get a 7-day grace period to add payment info. Your data is never deleted.
            </p>
          </div>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onClose} disabled={loading}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            disabled={loading}
            className="bg-gradient-to-r from-violet-600 to-fuchsia-600 hover:from-violet-700 hover:to-fuchsia-700"
          >
            {loading ? 'Starting...' : '🚀 Start Building Free'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
