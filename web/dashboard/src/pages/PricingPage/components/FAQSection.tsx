import { motion } from "framer-motion";
import { Card, CardContent } from "@/components/ui/card";
import { faqs } from "../data";
import { useScrollAnimation } from "../hooks";

// FAQ Section with scroll animations
export function FAQSection() {
  const { ref, inView } = useScrollAnimation(0.1, false);

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 40 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 40 }}
      transition={{ duration: 0.8, ease: "easeOut" }}
      className="mb-20"
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={inView ? { opacity: 1, scale: 1 } : { opacity: 0, scale: 0.95 }}
        transition={{ duration: 0.6, delay: 0.2 }}
        className="text-center mb-12"
      >
        <h2 className="text-3xl font-bold text-white mb-4">Frequently Asked Questions</h2>
        <p className="text-text-secondary max-w-2xl mx-auto">
          Everything you need to know about pricing and billing
        </p>
      </motion.div>

      <div className="grid md:grid-cols-2 gap-8 max-w-4xl mx-auto">
        {faqs.map((faq, index) => (
          <motion.div
            key={index}
            initial={{ opacity: 0, x: index % 2 === 0 ? -30 : 30, y: 20 }}
            animate={inView ? { opacity: 1, x: 0, y: 0 } : { opacity: 0, x: index % 2 === 0 ? -30 : 30, y: 20 }}
            transition={{
              duration: 0.6,
              delay: 0.4 + index * 0.1,
              ease: [0.25, 0.46, 0.45, 0.94]
            }}
          >
            <Card className="border-white/8 bg-white/5 h-full">
              <CardContent className="p-6">
                <h3 className="text-lg font-semibold text-white mb-3">{faq.question}</h3>
                <p className="text-text-secondary leading-relaxed">{faq.answer}</p>
              </CardContent>
            </Card>
          </motion.div>
        ))}
      </div>
    </motion.div>
  );
}