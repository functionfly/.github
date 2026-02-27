import { motion } from "framer-motion";
import { GradientButton } from "./GradientButton";
import { FadeInOnScroll } from "./FadeInOnScroll";

interface CTASectionProps {
  title: string;
  subtitle?: string;
  primaryButton?: {
    text: string;
    href?: string;
    onClick?: () => void;
  };
  secondaryButton?: {
    text: string;
    href?: string;
    onClick?: () => void;
  };
  className?: string;
}

export function CTASection({
  title,
  subtitle,
  primaryButton,
  secondaryButton,
  className = ""
}: CTASectionProps) {
  return (
    <motion.div
      className={`text-center py-12 ${className}`}
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6 }}
      viewport={{ once: true }}
    >
      <FadeInOnScroll>
        <h2 className="text-3xl lg:text-4xl font-bold text-slate-900 dark:text-white mb-4">
          {title}
        </h2>
      </FadeInOnScroll>

      {subtitle && (
        <FadeInOnScroll delay={0.2}>
          <p className="text-xl text-slate-600 dark:text-text-secondary mb-8 max-w-2xl mx-auto">
            {subtitle}
          </p>
        </FadeInOnScroll>
      )}

      <FadeInOnScroll delay={0.4}>
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          {primaryButton && (
            <GradientButton
              size="lg"
              href={primaryButton.href}
              onClick={primaryButton.onClick}
            >
              {primaryButton.text}
            </GradientButton>
          )}
          {secondaryButton && (
            <GradientButton
              variant="secondary"
              size="lg"
              href={secondaryButton.href}
              onClick={secondaryButton.onClick}
            >
              {secondaryButton.text}
            </GradientButton>
          )}
        </div>
      </FadeInOnScroll>
    </motion.div>
  );
}