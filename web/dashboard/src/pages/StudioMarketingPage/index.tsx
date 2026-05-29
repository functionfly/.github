import { Navbar } from '@/components/common/Navbar';
import { MetaTags } from '@/components/seo/MetaTags';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Footer } from '@/pages/LandingPage/components';
import '@/styles/index.css';
import { motion } from 'framer-motion';
import {
  ArrowRight,
  Bot,
  Clock,
  Code,
  GitBranch,
  Layers,
  Puzzle,
  Rocket,
  Shield,
  Sparkles,
} from 'lucide-react';
import { Link } from 'react-router-dom';

const fadeInUp = {
  initial: { opacity: 0, y: 20 },
  animate: { opacity: 1, y: 0 },
  transition: { duration: 0.6 },
};

const stagger = {
  animate: {
    transition: {
      staggerChildren: 0.1,
    },
  },
};

const features = [
  {
    icon: <Bot className="w-8 h-8 text-cyan-500" />,
    title: 'Multi-Agent Orchestration',
    description:
      'Coordinate multiple AI agents working together on complex tasks with built-in collaboration, memory sharing, and conflict resolution.',
  },
  {
    icon: <GitBranch className="w-8 h-8 text-green-500" />,
    title: 'Visual Workflow Editor',
    description:
      'Design agent workflows with an intuitive visual canvas. Define decision trees, parallel execution paths, and error handling visually.',
  },
  {
    icon: <Code className="w-8 h-8 text-yellow-500" />,
    title: 'Code Intelligence',
    description:
      'AI-powered code analysis, autocomplete, and refactoring suggestions that understand your entire codebase context.',
  },
  {
    icon: <Layers className="w-8 h-8 text-purple-500" />,
    title: 'Execution Environments',
    description:
      'Sandboxed runtime environments for each agent task with automatic isolation, resource limits, and dependency management.',
  },
  {
    icon: <Shield className="w-8 h-8 text-red-500" />,
    title: 'Enterprise Security',
    description:
      'End-to-end encryption, audit trails, and fine-grained permissions for all agent operations and data access.',
  },
  {
    icon: <Clock className="w-8 h-8 text-indigo-500" />,
    title: 'Persistent Memory',
    description:
      'Agents maintain context across sessions with vector-powered memory retrieval and semantic search capabilities.',
  },
  {
    icon: <Sparkles className="w-8 h-8 text-pink-500" />,
    title: 'Skills Marketplace',
    description:
      'Extend agent capabilities with pre-built skills for common tasks like data analysis, API integrations, and content generation.',
  },
  {
    icon: <Rocket className="w-8 h-8 text-orange-500" />,
    title: 'One-Click Deploy',
    description:
      'Deploy agent workflows to production with a single click. Automatic scaling, monitoring, and error recovery included.',
  },
  {
    icon: <Puzzle className="w-8 h-8 text-teal-500" />,
    title: 'Plugin Architecture',
    description:
      'Integrate with external services and APIs through a extensible plugin system. Build custom integrations easily.',
  },
];

const capabilities = [
  {
    title: 'Collaborative AI Development',
    description:
      'Work alongside AI agents in real-time. Agents can suggest code changes, run tests, and refactor while you maintain full control.',
    highlight: 'Pair programming at scale',
  },
  {
    title: 'Autonomous Task Execution',
    description:
      'Delegate complex multi-step tasks to AI agents that can reason, plan, and execute with minimal supervision.',
    highlight: 'Handles 80% of routine tasks',
  },
  {
    title: 'Visual Debugging',
    description:
      'Trace agent reasoning steps, inspect intermediate results, and debug issues with visual execution timelines.',
    highlight: 'Full observability',
  },
  {
    title: 'Team Memory Sharing',
    description:
      'Agents remember past decisions, learnings, and patterns across your entire team for institutional knowledge retention.',
    highlight: 'Never forget context',
  },
];

const useCases = [
  {
    title: 'Code Review & Refactoring',
    description:
      'Deploy agents to automatically review pull requests, suggest improvements, and perform refactoring tasks.',
    metrics: ['3x faster reviews', '50% less technical debt', 'Consistent quality'],
  },
  {
    title: 'Automated Testing',
    description:
      'Agents generate comprehensive test suites, identify edge cases, and continuously validate your codebase.',
    metrics: ['90% coverage', 'Instant feedback', 'Zero missed edge cases'],
  },
  {
    title: 'Documentation Generation',
    description:
      'Automatically generate and update documentation as your code evolves, keeping docs in sync with implementation.',
    metrics: ['Always up-to-date', 'Markdown & API docs', 'Code examples'],
  },
];

const benefits = [
  'Reduce development time by 40%',
  'Consistent code quality standards',
  '24/7 AI agent availability',
  'Seamless team collaboration',
  'Built-in security & compliance',
  'No credit card required',
];

function HeroSection() {
  return (
    <section className="relative overflow-hidden bg-gradient-to-b from-slate-50 to-white dark:from-slate-900 dark:to-slate-800 py-24 lg:py-32">
      <div className="absolute inset-0 bg-grid-pattern opacity-10" />
      <div className="absolute top-20 left-10 w-72 h-72 bg-cyan-500/20 rounded-full blur-[100px]" />
      <div className="absolute bottom-20 right-10 w-72 h-72 bg-purple-500/20 rounded-full blur-[100px]" />

      <div className="relative max-w-7xl mx-auto px-4 lg:px-6">
        <motion.div
          className="text-center max-w-4xl mx-auto"
          initial="initial"
          animate="animate"
          variants={stagger}
        >
          <motion.div variants={fadeInUp} className="mb-6">
            <span className="inline-flex items-center gap-2 px-4 py-1.5 rounded-full bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 text-sm font-medium border border-cyan-500/20">
              <Sparkles className="w-4 h-4" />
              Now in Public Beta
            </span>
          </motion.div>

          <motion.h1
            variants={fadeInUp}
            className="text-5xl lg:text-7xl font-bold tracking-tight mb-6"
          >
            <span className="text-slate-900 dark:text-white">Build AI Agents with </span>
            <span className="bg-gradient-to-r from-cyan-500 to-purple-600 bg-clip-text text-transparent">
              Studio
            </span>
          </motion.h1>

          <motion.p
            variants={fadeInUp}
            className="text-xl text-slate-600 dark:text-slate-300 mb-8 max-w-2xl mx-auto"
          >
            The visual workspace for building, deploying, and orchestrating AI agents. Design
            workflows visually, collaborate in real-time, and deploy with confidence.
          </motion.p>

          <motion.div
            variants={fadeInUp}
            className="flex flex-col sm:flex-row gap-4 justify-center"
          >
            <Link to="/studio">
              <Button
                size="lg"
                className="bg-gradient-to-r from-cyan-500 to-purple-600 hover:from-cyan-600 hover:to-purple-700 text-white shadow-lg shadow-cyan-500/25"
              >
                Launch Studio
                <ArrowRight className="w-4 h-4 ml-2" />
              </Button>
            </Link>
            <Link to="/studio">
              <Button
                variant="outline"
                size="lg"
                className="border-slate-300 dark:border-slate-600"
              >
                Watch Demo
              </Button>
            </Link>
          </motion.div>
        </motion.div>
      </div>
    </section>
  );
}

function FeaturesSection() {
  return (
    <section className="py-24 bg-white dark:bg-slate-800">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <h2 className="text-4xl lg:text-5xl font-bold text-slate-900 dark:text-white mb-4">
            Everything You Need to Build Agents
          </h2>
          <p className="text-xl text-slate-600 dark:text-slate-300 max-w-2xl mx-auto">
            A complete development environment designed specifically for AI agent workflows
          </p>
        </motion.div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <motion.div
              key={feature.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: index * 0.1 }}
            >
              <Card className="h-full bg-slate-50 dark:bg-slate-700/50 border-slate-200 dark:border-slate-600 hover:border-cyan-500/50 transition-colors">
                <CardHeader>
                  <div className="w-12 h-12 rounded-lg bg-slate-100 dark:bg-slate-600 flex items-center justify-center mb-4">
                    {feature.icon}
                  </div>
                  <CardTitle className="text-xl text-slate-900 dark:text-white">
                    {feature.title}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-slate-600 dark:text-slate-300">{feature.description}</p>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

function CapabilitiesSection() {
  return (
    <section className="py-24 bg-slate-50 dark:bg-slate-900">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <h2 className="text-4xl lg:text-5xl font-bold text-slate-900 dark:text-white mb-4">
            Power to the Developer
          </h2>
          <p className="text-xl text-slate-600 dark:text-slate-300 max-w-2xl mx-auto">
            Studio gives you the tools to build sophisticated AI agents without complexity
          </p>
        </motion.div>

        <div className="grid md:grid-cols-2 gap-8">
          {capabilities.map((cap, index) => (
            <motion.div
              key={cap.title}
              initial={{ opacity: 0, x: index % 2 === 0 ? -20 : 20 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ delay: index * 0.15 }}
            >
              <Card className="h-full bg-white dark:bg-slate-800 border-slate-200 dark:border-slate-700">
                <CardContent className="pt-6">
                  <span className="text-sm font-medium text-cyan-600 dark:text-cyan-400 mb-2 block">
                    {cap.highlight}
                  </span>
                  <h3 className="text-2xl font-bold text-slate-900 dark:text-white mb-3">
                    {cap.title}
                  </h3>
                  <p className="text-slate-600 dark:text-slate-300">{cap.description}</p>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

function UseCasesSection() {
  return (
    <section className="py-24 bg-white dark:bg-slate-800">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <h2 className="text-4xl lg:text-5xl font-bold text-slate-900 dark:text-white mb-4">
            Built for Real-World Tasks
          </h2>
          <p className="text-xl text-slate-600 dark:text-slate-300 max-w-2xl mx-auto">
            See how teams are using Studio to automate complex development workflows
          </p>
        </motion.div>

        <div className="grid md:grid-cols-3 gap-8">
          {useCases.map((useCase, index) => (
            <motion.div
              key={useCase.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: index * 0.1 }}
            >
              <Card className="h-full bg-gradient-to-br from-slate-50 to-white dark:from-slate-700 dark:to-slate-800 border-slate-200 dark:border-slate-600">
                <CardHeader>
                  <CardTitle className="text-lg text-slate-900 dark:text-white">
                    {useCase.title}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <p className="text-slate-600 dark:text-slate-300 mb-4">{useCase.description}</p>
                  <div className="flex flex-wrap gap-2">
                    {useCase.metrics.map((metric) => (
                      <span
                        key={metric}
                        className="px-2 py-1 bg-cyan-100 dark:bg-cyan-900/30 text-cyan-700 dark:text-cyan-300 text-xs font-medium rounded"
                      >
                        {metric}
                      </span>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

function CTASection() {
  return (
    <section className="py-24 bg-gradient-to-r from-cyan-500 to-purple-600 relative overflow-hidden">
      <div className="absolute inset-0 bg-grid-pattern opacity-10" />
      <div className="relative max-w-4xl mx-auto px-4 lg:px-6 text-center">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
        >
          <h2 className="text-4xl lg:text-5xl font-bold text-white mb-6">Ready to Build?</h2>
          <p className="text-xl text-white/80 mb-8 max-w-2xl mx-auto">
            Join thousands of developers building the next generation of AI-powered applications
            with Studio.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link to="/studio">
              <Button size="lg" className="bg-white text-purple-600 hover:bg-white/90 shadow-lg">
                Start Building Free
                <ArrowRight className="w-4 h-4 ml-2" />
              </Button>
            </Link>
            <Link to="/contact">
              <Button
                variant="outline"
                size="lg"
                className="border-white text-white hover:bg-white/10"
              >
                Contact Sales
              </Button>
            </Link>
          </div>
        </motion.div>
      </div>
    </section>
  );
}

export function StudioMarketingPage() {
  return (
    <div className="studio-marketing-page">
      <MetaTags
        title="Studio | FunctionFly - Visual AI Agent Development"
        description="Build, deploy, and orchestrate AI agents with Studio - the visual workspace for AI agent development. Design workflows, collaborate in real-time, deploy with confidence."
        keywords={[
          'AI agent studio',
          'agent development',
          'visual workflow editor',
          'multi-agent orchestration',
          'AI development tools',
        ]}
        url="/products/studio"
      />

      <Navbar />

      <HeroSection />
      <FeaturesSection />
      <CapabilitiesSection />
      <UseCasesSection />
      <CTASection />

      <Footer />
    </div>
  );
}

export default StudioMarketingPage;
