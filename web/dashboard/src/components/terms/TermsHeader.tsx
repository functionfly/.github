'use client';

import { motion } from 'framer-motion';
import { FileText } from 'lucide-react';

export function TermsHeader() {
  return (
    <div className="relative overflow-hidden bg-gradient-to-br from-[#6366f1]/20 to-transparent pt-24 pb-12">
      {/* Decorative Elements */}
      <div className="absolute inset-0 bg-grid-pattern opacity-[0.02]" />
      <div className="absolute top-1/4 right-1/4 w-64 h-64 bg-[#6366f1]/20 rounded-full blur-[100px]" />
      <div className="absolute bottom-0 left-1/4 w-48 h-48 bg-[#8b5cf6]/20 rounded-full blur-[80px]" />

      <div className="container mx-auto px-4 relative z-10">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="text-center max-w-2xl mx-auto"
        >
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-[#6366f1]/10 mb-6">
            <FileText className="w-8 h-8 text-[#6366f1]" />
          </div>
          <h1 className="text-3xl md:text-4xl font-bold text-text-primary mb-4">
            Terms of Service
          </h1>
          <p className="text-text-secondary">
            Last updated: {new Date().toLocaleDateString('en-US', { month: 'long', year: 'numeric' })}
          </p>
        </motion.div>
      </div>
    </div>
  );
}
