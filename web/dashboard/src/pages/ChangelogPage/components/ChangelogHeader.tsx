import { Link } from 'react-router-dom';
import { FileText, Rss } from 'lucide-react';
import { Navbar } from '@/components/common/Navbar';

const ChangelogHeader = () => {
  return (
    <>
      {/* Navbar */}
      <Navbar variant="landing" />

      {/* Header */}
      <div className="border-b border-border-subtle/50 pt-16 relative">
        <div className="absolute inset-0 bg-gradient-to-b from-transparent via-bg-primary/50 to-transparent"></div>
        <div className="container mx-auto px-4 py-12 relative z-10">
          <div className="text-center space-y-6">
            <div className="flex items-center justify-center gap-4 mb-6">
              <div className="glass-card p-4 rounded-2xl glow animate-float">
                <FileText className="h-10 w-10 text-brand-500 animate-pulse-glow" />
              </div>
              <h1 className="text-4xl lg:text-5xl font-bold text-gradient animate-fade-in">Changelog</h1>
            </div>
            <p className="text-text-secondary text-xl max-w-3xl mx-auto leading-relaxed animate-fade-in-delayed">
              Stay up to date with the latest features, improvements, and bug fixes in FunctionFly.
              Track our journey of innovation and continuous improvement.
            </p>
            <div className="flex justify-center">
              <Link
                to="/changelog/rss"
                className="inline-flex items-center gap-2 px-4 py-2 text-sm text-text-secondary hover:text-brand-500 bg-bg-glass rounded-lg border border-border-subtle/50 hover:border-brand-500/50 transition-all duration-300 hover:glow-sm group"
              >
                <Rss className="h-4 w-4 group-hover:animate-pulse" />
                RSS Feed
                <span className="text-xs opacity-60 group-hover:opacity-100 transition-opacity">Subscribe for updates</span>
              </Link>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default ChangelogHeader;