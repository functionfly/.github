'use client';

import { motion } from 'framer-motion';

interface SectionProps {
  title: string;
  children: React.ReactNode;
  index: number;
}

function Section({ title, children, index }: SectionProps) {
  return (
    <motion.section
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: index * 0.1 }}
      className="space-y-4"
    >
      <h2 className="text-xl font-semibold text-text-primary">{title}</h2>
      <div className="text-text-secondary space-y-4 leading-relaxed">
        {children}
      </div>
    </motion.section>
  );
}

export function TermsContent() {
  return (
    <div className="bg-card border border-border-default rounded-xl p-6 md:p-8 space-y-8">
      <Section title="1. Acceptance of Terms" index={0}>
        <p>
          By accessing or using FunctionFly (&ldquo;the Service&rdquo;), you agree to be bound by these
          Terms of Service (&ldquo;Terms&rdquo;). If you do not agree to these Terms, you may not
          access or use the Service.
        </p>
        <p>
          FunctionFly provides a platform for deploying, managing, and executing edge functions.
          These Terms govern your use of our website, services, and any associated software.
        </p>
      </Section>

      <Section title="2. Account Registration" index={1}>
        <p>
          To use certain features of the Service, you must register for an account. You agree to:
        </p>
        <ul className="list-disc list-inside space-y-2 ml-4">
          <li>Provide accurate, current, and complete information during registration</li>
          <li>Maintain and promptly update your account information</li>
          <li>Maintain the security of your account credentials</li>
          <li>Accept responsibility for all activities that occur under your account</li>
          <li>Notify us immediately of any unauthorized use of your account</li>
        </ul>
      </Section>

      <Section title="3. Acceptable Use" index={2}>
        <p>You agree not to use the Service to:</p>
        <ul className="list-disc list-inside space-y-2 ml-4">
          <li>Violate any applicable laws or regulations</li>
          <li>Infringe upon the rights of others</li>
          <li>Distribute malware, viruses, or any malicious code</li>
          <li>Conduct unauthorized data mining, scraping, or harvesting</li>
          <li>Overload or harm our infrastructure</li>
          <li>Send spam or unsolicited communications</li>
          <li>Impersonate any person or entity</li>
        </ul>
      </Section>

      <Section title="4. Intellectual Property" index={3}>
        <p>
          The Service and its original content, features, and functionality are owned by FunctionFly
          and are protected by international copyright, trademark, patent, trade secret, and other
          intellectual property laws.
        </p>
        <p>
          You retain ownership of any functions, code, or content you create and deploy using the
          Service (&ldquo;Your Content&rdquo;). By using the Service, you grant us a limited license
          to host, store, and execute Your Content solely as necessary to provide the Service.
        </p>
      </Section>

      <Section title="5. Subscription and Billing" index={4}>
        <p>
          Some features of the Service require a paid subscription. By subscribing:
        </p>
        <ul className="list-disc list-inside space-y-2 ml-4">
          <li>You agree to pay all fees associated with your subscription plan</li>
          <li>Subscriptions automatically renew unless canceled before the renewal date</li>
          <li>All fees are non-refundable unless otherwise specified</li>
          <li>We may modify pricing with advance notice to affected users</li>
        </ul>
      </Section>

      <Section title="6. Service Level Agreement" index={5}>
        <p>
          We strive to maintain high availability of the Service but do not guarantee uninterrupted
          access. We reserve the right to suspend or terminate the Service for maintenance,
          updates, or circumstances beyond our control.
        </p>
        <p>
          Our Service Level Agreement (SLA), if applicable to your plan, provides specific uptime
          guarantees and remedies for service interruptions.
        </p>
      </Section>

      <Section title="7. Data and Privacy" index={6}>
        <p>
          Your use of the Service is also governed by our Privacy Policy, which describes how we
          collect, use, and protect your personal information. By using the Service, you consent
          to our privacy practices as described in the Privacy Policy.
        </p>
      </Section>

      <Section title="8. Termination" index={7}>
        <p>
          You may terminate your account at any time through the Service settings or by
          contacting us. We may suspend or terminate your access to the Service if you violate
          these Terms or engage in prohibited activities.
        </p>
        <p>
          Upon termination, your right to use the Service immediately ceases. All provisions
          that by their nature should survive termination will survive, including ownership
          provisions, warranty disclaimers, and limitations of liability.
        </p>
      </Section>

      <Section title="9. Limitation of Liability" index={8}>
        <p>
          To the maximum extent permitted by law, FunctionFly and its affiliates, officers,
          employees, agents, and licensors shall not be liable for any indirect, incidental,
          special, consequential, or punitive damages, including loss of profits, data, or goodwill,
          arising out of or relating to your use of or inability to use the Service.
        </p>
        <p>
          Our total liability for any claim arising out of or relating to these Terms shall not
          exceed the amount you paid us for the Service during the twelve (12) months immediately
          preceding the event giving rise to the liability.
        </p>
      </Section>

      <Section title="10. Changes to Terms" index={9}>
        <p>
          We may modify these Terms at any time. We will provide notice of significant changes
          by posting the updated Terms on the Service and updating the &ldquo;Last updated&rdquo;
          date. Your continued use of the Service after such changes constitutes your acceptance
          of the revised Terms.
        </p>
      </Section>

      <Section title="11. Governing Law" index={10}>
        <p>
          These Terms shall be governed by and construed in accordance with the laws of the
          jurisdiction in which FunctionFly is established, without regard to its conflict of
          law provisions. Any legal action arising out of these Terms shall be brought in the
          courts of that jurisdiction.
        </p>
      </Section>

      <Section title="12. Contact Information" index={11}>
        <p>
          If you have any questions about these Terms, please contact us:
        </p>
        <ul className="list-disc list-inside space-y-2 ml-4">
          <li>By email: legal@functionfly.com</li>
          <li>By visiting the Contact page on our website</li>
          <li>By mail at the address provided in our Contact page</li>
        </ul>
      </Section>
    </div>
  );
}
