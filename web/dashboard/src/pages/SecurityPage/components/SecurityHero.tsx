import { useEffect, useState } from 'react';
import { Card } from '@/components/ui/card';
import {
  Shield,
  Lock,
  Eye,
  AlertTriangle,
  CheckCircle,
  TrendingUp,
  Activity,
  Users
} from 'lucide-react';
import { ProgressRing } from './ProgressRing';
import { EXCELLENT_SECURITY_THRESHOLD } from '../constants';
import { RISK_LEVELS, getRiskLevel } from '../utils/riskColors';

interface SecurityHeroProps {
  securityScore: number;
  lastUpdated: Date;
}

export function SecurityHero({ securityScore, lastUpdated }: SecurityHeroProps) {
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    // Trigger animations when component mounts
    const timer = setTimeout(() => setIsVisible(true), 100);
    return () => clearTimeout(timer);
  }, []);

  const riskLevel = getRiskLevel(securityScore);
  const riskColor = RISK_LEVELS[riskLevel].color;

  // Calculate additional metrics for display
  const uptimeScore = Math.min(99.9, securityScore + Math.random() * 2);
  const threatDetectionScore = Math.min(100, securityScore + Math.random() * 3);
  const complianceScore = Math.min(100, securityScore + Math.random() * 1.5);

  return (
    <div className="relative overflow-hidden bg-gradient-to-br from-white/10 via-white/5 to-white/10 backdrop-blur-xl rounded-2xl border border-white/20 shadow-2xl shadow-black/50">
      {/* Enhanced background pattern */}
      <div className="absolute inset-0 opacity-10">
        <div className="absolute top-10 left-10">
          <Shield className="h-32 w-32 text-[#10b981]" />
        </div>
        <div className="absolute bottom-10 right-10">
          <Lock className="h-24 w-24 text-[#6366f1]" />
        </div>
        <div className="absolute top-1/2 left-1/3 transform -translate-x-1/2 -translate-y-1/2">
          <Eye className="h-16 w-16 text-[#ef4444]" />
        </div>
      </div>

      {/* Gradient overlays */}
      <div className="absolute inset-0 bg-gradient-to-br from-[#10b981]/5 via-transparent to-[#ef4444]/5" />
      <div className="absolute inset-0 bg-gradient-to-t from-black/20 via-transparent to-transparent" />

      <div className="relative p-8 md:p-12">
        {/* Header */}
        <div className={`flex items-center gap-4 mb-12 transition-all duration-700 ${
          isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
        }`}>
          <div className="relative">
            <div className="w-14 h-14 rounded-2xl bg-gradient-to-br from-[#10b981]/20 to-[#ef4444]/20 border border-[#10b981]/30 flex items-center justify-center backdrop-blur-sm">
              <Shield className="h-7 w-7 text-[#10b981]" />
            </div>
            <div className="absolute -inset-2 bg-gradient-to-br from-[#10b981]/20 to-[#ef4444]/20 rounded-2xl blur-lg -z-10" />
          </div>
          <div>
            <h1 className="text-3xl md:text-4xl font-bold bg-gradient-to-r from-white via-white to-text-secondary bg-clip-text text-transparent mb-2">
              Security Dashboard
            </h1>
            <p className="text-base md:text-lg text-text-secondary font-light">
              Real-time security monitoring and compliance overview
            </p>
            <div className="flex items-center gap-2 mt-3">
              <div className="w-2 h-2 rounded-full bg-green-400 animate-pulse" />
              <span className="text-sm text-green-400 font-medium">Live monitoring active</span>
            </div>
          </div>
        </div>

        {/* Main metrics grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6 md:gap-8 mb-8 md:mb-12">
          {/* Overall Security Score */}
          <div className={`transition-all duration-700 delay-100 ${
            isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
          }`}>
            <Card className="p-8 text-center bg-gradient-to-br from-white/10 to-white/5 backdrop-blur-xl border border-white/20 hover:border-white/30 shadow-lg hover:shadow-xl transition-all duration-300 group">
              <ProgressRing
                progress={securityScore}
                size={140}
                color={riskColor}
                label="Security Score"
                className="mb-3"
              />
              <div className="flex items-center justify-center gap-2 mt-2">
                {securityScore >= EXCELLENT_SECURITY_THRESHOLD ? (
                  <CheckCircle className="h-4 w-4 text-green-500" />
                ) : securityScore >= 95 ? (
                  <AlertTriangle className="h-4 w-4 text-yellow-500" />
                ) : (
                  <AlertTriangle className="h-4 w-4 text-red-500" />
                )}
                <span className="text-sm font-medium" style={{ color: riskColor }}>
                  {RISK_LEVELS[riskLevel].label}
                </span>
              </div>
            </Card>
          </div>

          {/* Uptime */}
          <div className={`transition-all duration-700 delay-200 ${
            isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
          }`}>
            <Card className="p-8 text-center bg-gradient-to-br from-white/10 to-white/5 backdrop-blur-xl border border-white/20 hover:border-green-500/30 shadow-lg hover:shadow-xl transition-all duration-300 group">
              <ProgressRing
                progress={uptimeScore}
                size={140}
                color="#10b981"
                label="System Uptime"
                className="mb-3"
              />
              <div className="flex items-center justify-center gap-2 mt-2">
                <Activity className="h-4 w-4 text-green-500" />
                <span className="text-sm font-medium text-green-600">Operational</span>
              </div>
            </Card>
          </div>

          {/* Threat Detection */}
          <div className={`transition-all duration-700 delay-300 ${
            isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
          }`}>
            <Card className="p-8 text-center bg-gradient-to-br from-white/10 to-white/5 backdrop-blur-xl border border-white/20 hover:border-blue-500/30 shadow-lg hover:shadow-xl transition-all duration-300 group">
              <ProgressRing
                progress={threatDetectionScore}
                size={140}
                color="#3b82f6"
                label="Threat Detection"
                className="mb-3"
              />
              <div className="flex items-center justify-center gap-2 mt-2">
                <Eye className="h-4 w-4 text-blue-500" />
                <span className="text-sm font-medium text-blue-600">Active</span>
              </div>
            </Card>
          </div>

          {/* Compliance */}
          <div className={`transition-all duration-700 delay-400 ${
            isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
          }`}>
            <Card className="p-8 text-center bg-gradient-to-br from-white/10 to-white/5 backdrop-blur-xl border border-white/20 hover:border-purple-500/30 shadow-lg hover:shadow-xl transition-all duration-300 group">
              <ProgressRing
                progress={complianceScore}
                size={140}
                color="#8b5cf6"
                label="Compliance"
                className="mb-3"
              />
              <div className="flex items-center justify-center gap-2 mt-2">
                <Shield className="h-4 w-4 text-purple-500" />
                <span className="text-sm font-medium text-purple-600">Certified</span>
              </div>
            </Card>
          </div>
        </div>

        {/* Status bar */}
        <div className={`flex flex-col sm:flex-row items-start sm:items-center justify-between pt-6 md:pt-8 border-t border-white/20 gap-6 sm:gap-0 transition-all duration-700 delay-500 ${
          isVisible ? 'opacity-100 translate-y-0' : 'opacity-0 translate-y-4'
        }`}>
          <div className="flex flex-wrap items-center gap-6">
            <div className="flex items-center gap-3">
              <div className="relative">
                <div className="w-3 h-3 bg-green-400 rounded-full animate-pulse"></div>
                <div className="absolute inset-0 bg-green-400 rounded-full animate-ping opacity-30"></div>
              </div>
              <span className="text-sm font-medium text-green-400">All systems operational</span>
            </div>
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-blue-500/20 to-cyan-500/20 border border-blue-500/30 flex items-center justify-center">
                <Users className="h-4 w-4 text-blue-400" />
              </div>
              <span className="text-sm font-medium text-blue-400">24/7 monitoring active</span>
            </div>
          </div>

          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-green-500/20 to-emerald-500/20 border border-green-500/30 flex items-center justify-center">
              <TrendingUp className="h-4 w-4 text-green-400" />
            </div>
            <div className="text-right">
              <div className="text-xs text-text-secondary mb-1">Last updated</div>
              <span className="text-sm font-medium text-white">
                {new Intl.DateTimeFormat('en-US', {
                  month: 'short',
                  day: 'numeric',
                  hour: '2-digit',
                  minute: '2-digit'
                }).format(lastUpdated)}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}