import { Loader2, FileText } from 'lucide-react';

const LoadingState = () => {
  return (
    <div className="min-h-screen bg-gradient-radial relative overflow-hidden flex items-center justify-center">
      {/* Background Animation Effects */}
      <div className="absolute inset-0 aurora-bg opacity-30"></div>
      <div className="absolute inset-0 gradient-shift-bg opacity-20"></div>

      <div className="text-center relative z-10 space-y-8">
        <div className="flex justify-center">
          <div className="glass-card p-8 rounded-3xl glow animate-float">
            <div className="relative">
              <div className="absolute inset-0 bg-gradient-to-r from-brand-500/20 to-purple-500/20 rounded-2xl blur-xl animate-pulse"></div>
              <div className="relative p-6 bg-bg-glass/80 backdrop-blur-sm rounded-2xl">
                <FileText className="h-16 w-16 text-brand-500 animate-pulse-glow mx-auto mb-4" />
                <Loader2 className="h-8 w-8 animate-spin text-brand-500 mx-auto" />
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-4">
          <h2 className="text-2xl lg:text-3xl font-bold text-gradient animate-fade-in">
            Loading Changelog
          </h2>
          <p className="text-text-secondary text-lg animate-fade-in-delayed">
            Fetching the latest updates and release information...
          </p>
          <div className="flex justify-center space-x-1 animate-fade-in-delayed">
            <div className="w-2 h-2 bg-brand-500 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
            <div className="w-2 h-2 bg-purple-500 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
            <div className="w-2 h-2 bg-pink-500 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default LoadingState;