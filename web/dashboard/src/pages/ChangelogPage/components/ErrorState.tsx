import { AlertTriangle, RefreshCw } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface ErrorStateProps {
  error: string;
}

const ErrorState = ({ error }: ErrorStateProps) => {
  return (
    <div className="min-h-screen bg-gradient-radial relative overflow-hidden flex items-center justify-center">
      {/* Background Animation Effects */}
      <div className="absolute inset-0 aurora-bg opacity-30"></div>
      <div className="absolute inset-0 gradient-shift-bg opacity-20"></div>

      <div className="text-center relative z-10 space-y-8 max-w-lg mx-auto px-4">
        <div className="flex justify-center">
          <div className="glass-card p-8 rounded-3xl glow animate-shake border-error/30 bg-gradient-to-br from-error/5 to-red-500/5">
            <div className="relative">
              <div className="absolute inset-0 bg-gradient-to-r from-error/20 to-red-500/20 rounded-2xl blur-xl animate-pulse"></div>
              <div className="relative p-6 bg-bg-glass/80 backdrop-blur-sm rounded-2xl">
                <AlertTriangle className="h-16 w-16 text-error animate-pulse-glow mx-auto" />
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-4">
          <h2 className="text-2xl lg:text-3xl font-bold text-gradient animate-fade-in">
            Oops! Something went wrong
          </h2>
          <p className="text-text-secondary text-lg animate-fade-in-delayed leading-relaxed">
            We encountered an issue while loading the changelog. This might be a temporary problem.
          </p>
          <div className="bg-error/10 border border-error/20 rounded-lg p-4 animate-fade-in-delayed">
            <p className="text-error text-sm font-medium">{error}</p>
          </div>
        </div>

        <div className="flex flex-col sm:flex-row gap-4 justify-center animate-fade-in-delayed">
          <Button
            onClick={() => window.location.reload()}
            className="btn-secondary hover:scale-105 transition-all duration-300"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Try Again
          </Button>
          <Button
            variant="outline"
            onClick={() => window.location.href = '/'}
            className="btn-secondary hover:scale-105 transition-all duration-300"
          >
            Go Home
          </Button>
        </div>
      </div>
    </div>
  );
};

export default ErrorState;