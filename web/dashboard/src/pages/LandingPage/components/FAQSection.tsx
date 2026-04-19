import { Card, CardContent } from '@/components/ui/card';
import { AnimatePresence, motion } from 'framer-motion';
import { ChevronDown, HelpCircle } from 'lucide-react';
import { useState } from 'react';

interface FAQItem {
  question: string;
  answer: string;
}

const faqItems: FAQItem[] = [
  {
    question: 'How does pricing work?',
    answer:
      'FunctionFly offers transparent, usage-based pricing. Start with our free tier for up to 100,000 function invocations per month. Paid plans start at $29/month for the Starter plan (1M invocations) and scale up to $99/month for Professional (10M invocations). Enterprise plans offer custom pricing with unlimited requests and dedicated support. You only pay for what you use - no hidden fees, no minimum commitments.',
  },
  {
    question: "What's the setup process?",
    answer:
      'Getting started takes less than 5 minutes. Simply sign up, connect your cloud provider (AWS, Vercel, Google Cloud, etc.), and deploy our lightweight agent. FunctionFly automatically discovers your functions and starts monitoring. No code changes required - works with your existing infrastructure.',
  },
  {
    question: 'Can I migrate existing apps?',
    answer:
      "Absolutely! FunctionFly is designed to work with your existing applications without any changes. Whether you're using serverless functions on AWS Lambda, Google Cloud Functions, Vercel, or Netlify, our agent integrates seamlessly. Migration typically takes just a few minutes and doesn't require downtime.",
  },
  {
    question: 'What happens during failover?',
    answer:
      "During a failover, FunctionFly automatically redirects traffic to healthy instances while maintaining full monitoring. Our intelligent routing ensures zero data loss and minimal latency impact. You'll receive real-time alerts and detailed incident reports, with automatic recovery when services are restored.",
  },
  {
    question: 'What are the support response times?',
    answer:
      'We provide 24/7 support with different response times based on your plan. Free tier: Community support via Discord and GitHub. Starter plan ($29/mo): Email support within 24 hours. Professional plan ($99/mo): Priority email support within 4 hours. Enterprise plan: Phone and chat support within 1 hour, with dedicated account manager.',
  },
];

export function FAQSection() {
  const [openIndex, setOpenIndex] = useState<number | null>(null);

  const toggleFAQ = (index: number) => {
    setOpenIndex(openIndex === index ? null : index);
  };

  return (
    <section className="py-20 border-t border-white/8 aurora-bg faq-section-enhanced">
      <div className="max-w-4xl mx-auto px-4 lg:px-6">
        <div className="text-center mb-16">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
          >
            <div className="w-16 h-16 mx-auto mb-6 rounded-2xl bg-linear-to-br from-[#6366f1]/30 to-[#8b5cf6]/30 border border-[#6366f1]/40 flex items-center justify-center glow">
              <HelpCircle className="w-8 h-8 text-white drop-shadow-lg" />
            </div>
            <h2 className="text-3xl font-bold text-white mb-4">Frequently Asked Questions</h2>
            <p className="text-text-secondary max-w-2xl mx-auto">
              Everything you need to know about getting started with FunctionFly and maximizing your
              serverless infrastructure.
            </p>
          </motion.div>
        </div>

        <div className="space-y-4">
          {faqItems.map((item, index) => (
            <motion.div
              key={index}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
            >
              <Card className="border-white/8 bg-white/5 hover:bg-white/10 transition-colors card-elevation glass-card shine-effect">
                <CardContent className="p-0">
                  <button
                    onClick={() => toggleFAQ(index)}
                    className="w-full p-6 text-left flex items-center justify-between hover:bg-white/5 transition-colors rounded-lg"
                  >
                    <h3 className="text-lg font-semibold text-white pr-4">{item.question}</h3>
                    <motion.div
                      animate={{ rotate: openIndex === index ? 180 : 0 }}
                      transition={{ duration: 0.2 }}
                    >
                      <ChevronDown className="w-5 h-5 text-text-secondary shrink-0" />
                    </motion.div>
                  </button>

                  <AnimatePresence>
                    {openIndex === index && (
                      <motion.div
                        initial={{ height: 0, opacity: 0 }}
                        animate={{ height: 'auto', opacity: 1 }}
                        exit={{ height: 0, opacity: 0 }}
                        transition={{ duration: 0.3 }}
                        className="overflow-hidden"
                      >
                        <div className="px-6 pb-6">
                          <p className="text-text-secondary leading-relaxed">{item.answer}</p>
                        </div>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>

        {/* Call to action */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          whileInView={{ opacity: 1, scale: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.5 }}
          className="text-center mt-16"
        >
          <Card className="border-[#6366f1]/30 bg-[#6366f1]/5 max-w-lg mx-auto card-elevation glass-card shine-effect glow">
            <CardContent className="p-8">
              <h3 className="text-xl font-bold text-white mb-3">Still have questions?</h3>
              <p className="text-text-secondary mb-4">
                Our team is here to help you get the most out of FunctionFly.
              </p>
              <a
                href="mailto:support@functionfly.com"
                className="inline-flex items-center px-4 py-2 rounded-lg bg-[#6366f1] hover:bg-[#6366f1]/80 text-white font-medium transition-colors glow hover-lift"
              >
                Contact Support
              </a>
            </CardContent>
          </Card>
        </motion.div>
      </div>
    </section>
  );
}
