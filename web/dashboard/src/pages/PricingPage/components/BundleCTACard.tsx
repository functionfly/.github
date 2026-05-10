import { motion } from 'framer-motion';
import { Rocket, Sparkles, ArrowRight, Zap, ShoppingCart, Brain } from 'lucide-react';
import { Link } from 'react-router-dom';

export function BundleCTACard() {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: 0.1 }}
      className="relative overflow-hidden rounded-2xl bg-gradient-to-r from-violet-600 via-purple-600 to-fuchsia-600 p-1 mb-12"
    >
      <div className="relative bg-slate-900/90 backdrop-blur-sm rounded-xl p-6 md:p-8">
        {/* Background glow */}
        <div className="absolute top-0 right-0 w-64 h-64 bg-violet-500/20 rounded-full blur-3xl -translate-y-1/2 translate-x-1/2" />
        <div className="absolute bottom-0 left-0 w-48 h-48 bg-fuchsia-500/20 rounded-full blur-3xl translate-y-1/2 -translate-x-1/2" />

        <div className="relative flex flex-col md:flex-row items-center gap-6">
          {/* Icon stack */}
          <div className="flex-shrink-0">
            <div className="flex -space-x-3">
              <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-blue-500 to-cyan-500 flex items-center justify-center border-2 border-slate-900 shadow-lg">
                <Rocket className="w-7 h-7 text-white" />
              </div>
              <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-purple-500 to-pink-500 flex items-center justify-center border-2 border-slate-900 shadow-lg">
                <ShoppingCart className="w-7 h-7 text-white" />
              </div>
              <div className="w-14 h-14 rounded-xl bg-gradient-to-br from-orange-500 to-amber-500 flex items-center justify-center border-2 border-slate-900 shadow-lg">
                <Brain className="w-7 h-7 text-white" />
              </div>
            </div>
          </div>

          {/* Content */}
          <div className="flex-1 text-center md:text-left">
            <div className="inline-flex items-center gap-2 text-violet-400 text-sm font-semibold mb-2">
              <Sparkles className="w-4 h-4" />
              NEW: Backend-in-a-Box Pricing
            </div>
            <h2 className="text-2xl md:text-3xl font-bold text-white mb-2">
              One Click → Full Backend
            </h2>
            <p className="text-slate-400 max-w-xl">
              Skip the decision fatigue. Get pre-configured bundles with Auth, Payments, 
              User DB, and more. Start free until you hit 100 users or $1K MRR.
            </p>
          </div>

          {/* CTA */}
          <div className="flex-shrink-0">
            <Link
              to="/bundles"
              className="group inline-flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-violet-500 to-fuchsia-500 hover:from-violet-600 hover:to-fuchsia-600 text-white font-semibold rounded-xl transition-all shadow-lg shadow-violet-500/25 hover:shadow-violet-500/40"
            >
              <Zap className="w-5 h-5" />
              Explore Bundles
              <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
            </Link>
          </div>
        </div>

        {/* Trust badges */}
        <div className="relative mt-6 pt-6 border-t border-slate-800/50 flex flex-wrap justify-center md:justify-start gap-6 text-sm text-slate-500">
          <span className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-green-500" />
            SaaS Starter: $29/mo
          </span>
          <span className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-purple-500" />
            Marketplace: $49/mo
          </span>
          <span className="flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-orange-500" />
            AI App: $39/mo
          </span>
          <span className="flex items-center gap-2 text-violet-400 font-medium">
            <Sparkles className="w-4 h-4" />
            Founder Mode: $0 for 3 months
          </span>
        </div>
      </div>
    </motion.div>
  );
}
