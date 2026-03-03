import { Shield, ArrowLeft } from 'lucide-react';
import { Link } from 'react-router-dom';

export function SecurityPageHeader() {
  return (
    <>
      {/* Navigation Bar */}
      <nav className="security-nav border-b border-border-default bg-bg-secondary/90 dark:bg-black/30 backdrop-blur-md sticky top-0 z-50 relative overflow-hidden">
        {/* Background gradient overlay */}
        <div className="absolute inset-0 bg-gradient-to-r from-green-500/5 via-transparent to-red-500/5" />
        <div className="relative container mx-auto px-4">
          <div className="flex items-center justify-between h-16">
            <Link to="/" className="flex items-center gap-2 text-text-primary hover:text-green-600 dark:hover:text-[#10b981] transition-all duration-300 group">
              <div className="p-1 rounded-lg bg-bg-tertiary dark:bg-white/5 group-hover:bg-green-500/10 transition-colors">
                <ArrowLeft className="w-4 h-4" />
              </div>
              <span className="font-medium">Back to Home</span>
            </Link>
            <div className="flex items-center gap-3">
              <div className="w-2 h-2 rounded-full bg-gradient-to-r from-[#10b981] to-[#ef4444] animate-pulse" />
              <h1 className="text-xl font-bold text-text-primary">
                Security & Compliance
              </h1>
            </div>
            <div className="w-24" /> {/* Spacer for centering */}
          </div>
        </div>
      </nav>

      {/* Header */}
      <div className="border-b border-border-default bg-gradient-to-r from-transparent via-bg-tertiary/50 dark:via-white/5 to-transparent">
        <div className="container mx-auto px-4 py-12 md:py-16">
          <div className="max-w-4xl mx-auto text-center">
            <div className="relative inline-block mb-8">
              <div className="w-20 h-20 mx-auto rounded-3xl bg-gradient-to-br from-[#10b981]/30 via-[#ef4444]/20 to-[#6366f1]/20 border border-[#10b981]/30 flex items-center justify-center backdrop-blur-sm">
                <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-[#10b981] to-[#ef4444] flex items-center justify-center">
                  <Shield className="w-8 h-8 text-white" />
                </div>
              </div>
              <div className="absolute -inset-4 bg-gradient-to-r from-[#10b981]/20 via-[#ef4444]/10 to-[#6366f1]/20 rounded-full blur-xl -z-10" />
            </div>

            <h1 className="text-5xl md:text-6xl lg:text-7xl font-bold mb-6 leading-tight">
              <span className="text-text-primary">
                Security &
              </span>
              <br />
              <span className="security-compliance-gradient bg-gradient-to-r from-[#10b981] via-[#ef4444] to-[#6366f1] bg-clip-text text-transparent">
                Compliance
              </span>
            </h1>

            <p className="text-text-secondary text-xl md:text-2xl max-w-3xl mx-auto leading-relaxed font-light">
              Enterprise-grade security measures to protect your applications and data.
              <br className="hidden md:block" />
              Built with security-first architecture and compliance standards.
            </p>

            <div className="flex items-center justify-center gap-4 mt-8">
              <div className="flex items-center gap-2 px-4 py-2 rounded-full bg-gradient-to-r from-green-500/10 to-emerald-500/10 border border-green-500/20">
                <div className="w-2 h-2 rounded-full bg-green-500 dark:bg-green-400 animate-pulse" />
                <span className="text-sm font-medium text-green-600 dark:text-green-400">SOC 2 Type II Certified</span>
              </div>
              <div className="flex items-center gap-2 px-4 py-2 rounded-full bg-gradient-to-r from-blue-500/10 to-cyan-500/10 border border-blue-500/20">
                <div className="w-2 h-2 rounded-full bg-blue-500 dark:bg-blue-400 animate-pulse" />
                <span className="text-sm font-medium text-blue-600 dark:text-blue-400">GDPR Compliant</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
}
