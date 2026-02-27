import { Alert, AlertDescription } from '@/components/ui/alert';
import { CheckCircle, AlertCircle } from 'lucide-react';

interface SubmitStatusAlertsProps {
  submitStatus: 'idle' | 'success' | 'error';
}

export function SubmitStatusAlerts({ submitStatus }: SubmitStatusAlertsProps) {
  if (submitStatus === 'idle') return null;

  if (submitStatus === 'success') {
    return (
      <Alert className="border-success/20 bg-success/5">
        <CheckCircle className="h-4 w-4 text-success" />
        <AlertDescription className="text-success">
          Your message has been sent successfully! We'll get back to you within 24 hours.
        </AlertDescription>
      </Alert>
    );
  }

  return (
    <Alert variant="destructive">
      <AlertCircle className="h-4 w-4" />
      <AlertDescription>
        There was an error sending your message. Please try again or contact us directly.
      </AlertDescription>
    </Alert>
  );
}