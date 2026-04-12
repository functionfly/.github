import { useToastSettingsStore, type ToastPosition } from '@/stores/toastSettingsStore';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { Button } from '@/components/ui/button';
import { toast } from 'sonner';

const POSITION_OPTIONS: { value: ToastPosition; label: string }[] = [
  { value: 'top-left', label: 'Top Left' },
  { value: 'top-center', label: 'Top Center' },
  { value: 'top-right', label: 'Top Right' },
  { value: 'bottom-left', label: 'Bottom Left' },
  { value: 'bottom-center', label: 'Bottom Center' },
  { value: 'bottom-right', label: 'Bottom Right' },
];

const DURATION_OPTIONS: { value: number; label: string }[] = [
  { value: 3000, label: '3 seconds' },
  { value: 5000, label: '5 seconds' },
  { value: 7000, label: '7 seconds' },
  { value: 10000, label: '10 seconds' },
  { value: 15000, label: '15 seconds' },
  { value: 30000, label: '30 seconds' },
];

export function ToastSettingsPanel() {
  const {
    position,
    duration,
    richColors,
    closeButton,
    setPosition,
    setDuration,
    setRichColors,
    setCloseButton,
    resetToDefaults,
  } = useToastSettingsStore();

  const handleTestToast = () => {
    toast.success('Test notification', {
      description: 'This is how notifications will appear',
    });
  };

  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="toast-position">Position</Label>
        <Select value={position} onValueChange={(v) => setPosition(v as ToastPosition)}>
          <SelectTrigger id="toast-position" className="w-full sm:w-[240px]">
            <SelectValue placeholder="Select position" />
          </SelectTrigger>
          <SelectContent>
            {POSITION_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <p className="text-sm text-text-secondary">
          Choose where notifications appear on your screen
        </p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="toast-duration">Display Duration</Label>
        <Select 
          value={duration.toString()} 
          onValueChange={(v) => setDuration(parseInt(v, 10))}
        >
          <SelectTrigger id="toast-duration" className="w-full sm:w-[240px]">
            <SelectValue placeholder="Select duration" />
          </SelectTrigger>
          <SelectContent>
            {DURATION_OPTIONS.map((opt) => (
              <SelectItem key={opt.value} value={opt.value.toString()}>
                {opt.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <p className="text-sm text-text-secondary">
          How long notifications stay visible before auto-dismissing
        </p>
      </div>

      <div className="flex items-center justify-between">
        <div className="space-y-0.5">
          <Label htmlFor="rich-colors">Rich Colors</Label>
          <p className="text-sm text-text-secondary">
            Use color-coded notifications (green for success, red for errors)
          </p>
        </div>
        <Switch
          id="rich-colors"
          checked={richColors}
          onCheckedChange={setRichColors}
        />
      </div>

      <div className="flex items-center justify-between">
        <div className="space-y-0.5">
          <Label htmlFor="close-button">Show Close Button</Label>
          <p className="text-sm text-text-secondary">
            Display a dismiss button on notifications
          </p>
        </div>
        <Switch
          id="close-button"
          checked={closeButton}
          onCheckedChange={setCloseButton}
        />
      </div>

      <div className="flex gap-3 pt-4 border-t border-border-subtle">
        <Button variant="outline" onClick={handleTestToast}>
          Test Notification
        </Button>
        <Button variant="ghost" onClick={resetToDefaults}>
          Reset to Defaults
        </Button>
      </div>
    </div>
  );
}
