'use client';

import { motion } from 'framer-motion';
import { Link } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { FileText, Radar, Award, Shield, AlertTriangle, HelpCircle, Library, Phone } from 'lucide-react';
import { Footer } from '@/pages/LandingPage/components/Footer';
import { useSecurityState } from './hooks';
import {
  SecurityPageHeader,
  SecurityHero,
  RealTimeSecurityStatus,
  ComplianceFrameworks,
  SecurityMeasures,
  IncidentResponse,
  SecurityFAQ,
  SecurityResources,
  SecurityContactInfo,
  CollapsibleSection
} from './components';

export function SecurityPage() {
  const {
    expandedSection,
    securityScore,
    lastUpdated,
    serviceStatus,
    recentIncidents,
    sslCertificates,
    loading,
    error,
    toggleSection,
    retry
  } = useSecurityState();

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 relative overflow-hidden">
        {/* Background decorative elements */}
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgba(16,185,129,0.08),transparent_50%)] pointer-events-none" />
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_70%_80%,rgba(239,68,68,0.08),transparent_50%)] pointer-events-none" />

        <SecurityPageHeader />
        <div className="relative z-10 container mx-auto px-4 py-16">
          <div className="max-w-6xl mx-auto">
            <div className="flex items-center justify-center py-24">
              <div className="text-center">
                <div className="relative mb-8">
                  <div className="w-16 h-16 mx-auto rounded-2xl bg-gradient-to-br from-[#10b981]/30 to-[#ef4444]/20 border border-[#10b981]/30 flex items-center justify-center backdrop-blur-sm">
                    <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-[#10b981] to-[#ef4444] flex items-center justify-center">
                      <div className="animate-spin rounded-full h-8 w-8 border-2 border-white border-t-transparent"></div>
                    </div>
                  </div>
                  <div className="absolute -inset-4 bg-gradient-to-r from-[#10b981]/20 via-[#ef4444]/10 to-[#6366f1]/20 rounded-full blur-xl -z-10" />
                </div>
                <h2 className="text-2xl font-bold text-white mb-4">Loading Security Data</h2>
                <p className="text-text-secondary">Initializing security monitoring and compliance checks...</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 relative overflow-hidden">
        {/* Background decorative elements */}
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgba(239,68,68,0.08),transparent_50%)] pointer-events-none" />

        <SecurityPageHeader />
        <div className="relative z-10 container mx-auto px-4 py-16">
          <div className="max-w-6xl mx-auto">
            <div className="flex items-center justify-center py-24">
              <div className="text-center">
                <div className="relative mb-8">
                  <div className="w-20 h-20 mx-auto rounded-3xl bg-gradient-to-br from-red-500/30 to-orange-500/20 border border-red-500/30 flex items-center justify-center backdrop-blur-sm">
                    <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-red-500 to-orange-500 flex items-center justify-center">
                      <svg className="h-9 w-9 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z" />
                      </svg>
                    </div>
                  </div>
                  <div className="absolute -inset-4 bg-gradient-to-r from-red-500/20 via-orange-500/10 to-red-500/20 rounded-full blur-xl -z-10" />
                </div>
                <h2 className="text-3xl font-bold text-white mb-4">Failed to Load Security Data</h2>
                <p className="text-text-secondary text-lg mb-8 max-w-md mx-auto leading-relaxed">{error}</p>
                <Button
                  onClick={retry}
                  size="lg"
                  className="bg-gradient-to-r from-red-600 to-orange-600 hover:from-red-700 hover:to-orange-700 text-white font-semibold px-8 py-4 shadow-lg shadow-red-500/25 hover:shadow-red-500/40 transition-all duration-300 transform hover:scale-105"
                >
                  Try Again
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 relative overflow-hidden">
      {/* Background decorative elements */}
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_30%_20%,rgba(16,185,129,0.08),transparent_50%)] pointer-events-none" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_70%_80%,rgba(239,68,68,0.08),transparent_50%)] pointer-events-none" />
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_50%_50%,rgba(99,102,241,0.05),transparent_50%)] pointer-events-none" />
      <div className="absolute inset-0 bg-[linear-gradient(45deg,transparent_25%,rgba(255,255,255,0.005)_25%,rgba(255,255,255,0.005)_50%,transparent_50%,transparent_75%,rgba(255,255,255,0.005)_75%)] bg-[length:20px_20px] opacity-30 pointer-events-none" />

      <SecurityPageHeader />

      <div className="relative z-10 container mx-auto px-4 py-8 md:py-12">
        <div className="max-w-7xl mx-auto space-y-6 md:space-y-12">
          <SecurityHero
            securityScore={securityScore}
            lastUpdated={lastUpdated}
          />

          <CollapsibleSection
            title="Real-Time Security Status"
            description="Live monitoring and threat detection"
            icon={<Radar className="h-5 w-5 text-blue-600" />}
            defaultExpanded={true}
          >
            <RealTimeSecurityStatus
              serviceStatus={serviceStatus}
              sslCertificates={sslCertificates}
              recentIncidents={recentIncidents}
            />
          </CollapsibleSection>

          <CollapsibleSection
            title="Compliance Certifications"
            description="Security standards and regulatory compliance"
            icon={<Award className="h-5 w-5 text-purple-600" />}
            defaultExpanded={false}
          >
            <ComplianceFrameworks />
          </CollapsibleSection>

          <CollapsibleSection
            title="Security Measures"
            description="Active security controls and protections"
            icon={<Shield className="h-5 w-5 text-emerald-600" />}
            defaultExpanded={false}
          >
            <SecurityMeasures />
          </CollapsibleSection>

          <CollapsibleSection
            title="Incident Response"
            description="How we handle and respond to security incidents"
            icon={<AlertTriangle className="h-5 w-5 text-orange-600" />}
            defaultExpanded={false}
          >
            <IncidentResponse />
          </CollapsibleSection>

          <CollapsibleSection
            title="Security FAQ"
            description="Common questions about our security practices"
            icon={<HelpCircle className="h-5 w-5 text-indigo-600" />}
            defaultExpanded={false}
          >
            <SecurityFAQ
              expandedSection={expandedSection}
              toggleSection={toggleSection}
            />
          </CollapsibleSection>

          <CollapsibleSection
            title="Security Resources"
            description="Documentation, guides, and security tools"
            icon={<Library className="h-5 w-5 text-green-600" />}
            defaultExpanded={false}
          >
            <SecurityResources />
          </CollapsibleSection>

          <CollapsibleSection
            title="Contact Information"
            description="How to reach our security team"
            icon={<Phone className="h-5 w-5 text-cyan-600" />}
            defaultExpanded={false}
          >
            <SecurityContactInfo />
          </CollapsibleSection>

          {/* Back to Home */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.6, delay: 0.2 }}
            className="text-center pt-8"
          >
            <Link to="/">
              <Button
                variant="outline"
                size="lg"
                className="border-white/30 hover:border-white/50 hover:bg-white/10 text-white font-semibold px-8 py-4 backdrop-blur-sm transition-all duration-300 transform hover:scale-105"
              >
                <FileText className="h-5 w-5 mr-3" />
                Back to Home
              </Button>
            </Link>
          </motion.div>
        </div>
      </div>

      <Footer />
    </div>
  );
}