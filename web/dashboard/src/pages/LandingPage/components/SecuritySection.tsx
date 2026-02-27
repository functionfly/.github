import { motion } from "framer-motion";
import { Database, Cloud, Key, Fingerprint, Server, FileLock, Lock } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { SOC2Badge, GDPRBadge, ISO27001Badge, HIPAAReadyBadge, CCPABadge, ZeroTrustBadge } from "@/components/icons/SecurityBadges";

const securityFeatures = [
  {
    icon: Database,
    title: "End-to-End Encryption",
    description:
      "All data encrypted in transit and at rest with industry-standard AES-256",
  },
  {
    icon: Cloud,
    title: "Your Data Stays Yours",
    description:
      "Your data never leaves your cloud. We don't store or access your application data",
  },
  {
    icon: Key,
    title: "Zero-Trust Architecture",
    description:
      "Every request authenticated and authorized with minimal privilege access",
  },
  {
    icon: Fingerprint,
    title: "Advanced Authentication",
    description:
      "Multi-factor authentication, API keys, and OAuth integration for secure access",
  },
  {
    icon: Server,
    title: "Isolated Environments",
    description:
      "Each tenant runs in completely isolated environments with network segmentation",
  },
  {
    icon: FileLock,
    title: "Audit Logging",
    description:
      "Comprehensive audit trails for all system activities and access attempts",
  },
];

const securityBadges = [
  {
    icon: SOC2Badge,
    title: "SOC 2 Type II",
    description: "Security, availability, and confidentiality controls",
  },
  {
    icon: GDPRBadge,
    title: "GDPR Compliant",
    description: "European data protection regulation compliance",
  },
  {
    icon: ISO27001Badge,
    title: "ISO 27001",
    description: "Information security management systems",
  },
  {
    icon: HIPAAReadyBadge,
    title: "HIPAA Ready",
    description: "Healthcare data protection standards",
  },
  {
    icon: CCPABadge,
    title: "CCPA Compliant",
    description: "California consumer privacy protection",
  },
  {
    icon: ZeroTrustBadge,
    title: "Zero Trust",
    description: "Never trust, always verify security model",
  },
];

export function SecuritySection() {
  return (
    <section className="py-20 border-t border-white/8 mesh-gradient-bg security-section-enhanced">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl font-bold text-text-primary mb-4" style={{ color: 'var(--text-primary)', fontWeight: 800 }}>
            Enterprise-grade security & compliance
          </h2>
          <p className="text-text-secondary max-w-2xl mx-auto">
            Your data and applications are protected with military-grade security standards.
          </p>
        </div>

        {/* Security Features Grid */}
        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 mb-16">
          {securityFeatures.map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
            >
              <Card className="h-full hover:border-[#6366f1]/30 transition-colors glass-card card-elevation">
                <CardContent className="p-6">
                  <div className="w-12 h-12 rounded-xl bg-linear-to-br from-emerald-500/20 to-emerald-600/20 border border-emerald-500/20 flex items-center justify-center mb-4">
                    <feature.icon className="w-6 h-6 text-emerald-400" />
                  </div>
                  <h3 className="text-lg font-semibold text-text-primary mb-2" style={{ color: 'var(--text-primary)', fontWeight: 600 }}>
                    {feature.title}
                  </h3>
                  <p className="text-text-secondary text-sm">
                    {feature.description}
                  </p>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>

        {/* Key Security Promise */}
        <div className="text-center mb-16">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
            className="inline-block"
          >
            <Card className="border-emerald-500/30 bg-emerald-500/5 max-w-2xl mx-auto glass-card card-elevation">
              <CardContent className="p-8">
                <div className="w-16 h-16 mx-auto mb-6 rounded-2xl bg-linear-to-br from-emerald-500/20 to-emerald-600/20 border border-emerald-500/20 flex items-center justify-center">
                  <Lock className="w-8 h-8 text-emerald-400" />
                </div>
                <h3 className="text-2xl font-bold text-text-primary mb-4" style={{ color: 'var(--text-primary)' }}>
                  Your Data Never Leaves Your Cloud
                </h3>
                <p className="text-text-secondary text-lg">
                  Unlike other platforms, we don't store, process, or access your application data.
                  Your functions and data remain entirely within your chosen cloud providers.
                </p>
              </CardContent>
            </Card>
          </motion.div>
        </div>

        {/* Security Compliance Badges */}
        <div>
          <div className="text-center mb-8">
            <h3 className="text-xl font-semibold text-text-primary mb-2" style={{ color: 'var(--text-primary)', fontWeight: 700 }}>
              Compliance certifications & standards
            </h3>
            <p className="text-text-secondary">
              Meet the highest industry standards for data protection and security
            </p>
          </div>

          <div className="grid grid-cols-2 lg:grid-cols-3 gap-4">
            {securityBadges.map((badge, index) => (
              <motion.div
                key={badge.title}
                initial={{ opacity: 0, scale: 0.95 }}
                whileInView={{ opacity: 1, scale: 1 }}
                viewport={{ once: true }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
                className="text-center"
              >
                <Card className="h-full hover:border-[#6366f1]/30 transition-colors glass-card card-elevation">
                  <CardContent className="p-4">
                    <div className="w-12 h-12 mx-auto mb-3 rounded-xl bg-white/5 border border-white/10 flex items-center justify-center">
                      <badge.icon />
                    </div>
                    <h4 className="text-sm font-semibold text-text-primary mb-1" style={{ color: 'var(--text-primary)' }}>{badge.title}</h4>
                    <p className="text-xs text-text-secondary">{badge.description}</p>
                  </CardContent>
                </Card>
              </motion.div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}