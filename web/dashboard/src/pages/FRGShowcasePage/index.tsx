/**
 * FRG Showcase Page
 * Public marketing page with amazing 3D scenes and graph info
 */

import { useRef, useEffect, useState } from 'react';
import { motion, useScroll, useTransform, useSpring } from 'framer-motion';
import { gsap } from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';
import { 
  Share2, 
  GitBranch, 
  Zap, 
  Globe, 
  Cpu, 
  Activity,
  Layers,
  ArrowRight,
  Sparkles,
  Play,
  Clock,
  Shield,
  Terminal,
  Network,
  GitMerge,
  Workflow,
  Boxes,
  Database,
  Server,
  RefreshCw,
  ChevronDown,
  ExternalLink,
  Code2,
  FunctionSquare,
  Route,
  GitCommit,
  Maximize2,
  Eye,
  MousePointerClick,
  Info,
  ArrowUpRight,
  Menu,
  X,
  CheckCircle,
  Disc,
} from 'lucide-react';
import { siGithub, siX } from 'simple-icons';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';
import { Separator } from '@/components/ui/separator';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

// Import 3D scenes
import { 
  GraphNetwork3D, 
  FlowingDataStream, 
  CrystalGraph, 
  AnimatedNodeCluster 
} from '@/components/3d';

gsap.registerPlugin(ScrollTrigger);

// Feature card component
interface FeatureCardProps {
  icon: React.ComponentType<{ className?: string }>;
  title: string;
  description: string;
  color: string;
  delay?: number;
}

function FeatureCard({ icon: Icon, title, description, color, delay = 0 }: FeatureCardProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 30 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ delay, duration: 0.6, ease: "easeOut" }}
    >
      <Card className="p-6 bg-black/40 backdrop-blur-sm border-white/10 hover:border-white/20 transition-all duration-300 hover:scale-[1.02] group">
        <div 
          className="w-12 h-12 rounded-xl flex items-center justify-center mb-4 transition-transform group-hover:scale-110"
          style={{ backgroundColor: `${color}20`, color }}
        >
          <Icon className="w-6 h-6" />
        </div>
        <h3 className="text-lg font-semibold text-white mb-2">{title}</h3>
        <p className="text-sm text-white/60">{description}</p>
      </Card>
    </motion.div>
  );
}

// Stat counter with animation
function AnimatedStat({ value, label, suffix = '' }: { value: number; label: string; suffix?: string }) {
  const [count, setCount] = useState(0);
  const ref = useRef<HTMLDivElement>(null);
  
  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          const duration = 2000;
          const steps = 60;
          const increment = value / steps;
          let current = 0;
          
          const timer = setInterval(() => {
            current += increment;
            if (current >= value) {
              setCount(value);
              clearInterval(timer);
            } else {
              setCount(Math.floor(current));
            }
          }, duration / steps);
          
          observer.disconnect();
        }
      },
      { threshold: 0.5 }
    );
    
    if (ref.current) observer.observe(ref.current);
    return () => observer.disconnect();
  }, [value]);

  return (
    <div ref={ref} className="text-center">
      <div className="text-4xl md:text-5xl font-bold bg-gradient-to-r from-brand-400 to-purple-400 bg-clip-text text-transparent">
        {count.toLocaleString()}{suffix}
      </div>
      <div className="text-sm text-white/60 mt-1">{label}</div>
    </div>
  );
}

// Code snippet component
function CodeSnippet() {
  const code = `{
  "graph": {
    "name": "AI Pipeline",
    "version": "1.2.0",
    "nodes": [
      {
        "id": "input",
        "type": "webhook",
        "config": { "auth": "bearer" }
      },
      {
        "id": "transform",
        "type": "function",
        "ref": "openai/gpt-4o",
        "input": { "from": "input.data" }
      },
      {
        "id": "output",
        "type": "response",
        "mapping": { "result": "transform.output" }
      }
    ],
    "edges": [
      { "from": "input", "to": "transform" },
      { "from": "transform", "to": "output" }
    ]
  }
}`;

  return (
    <motion.div
      initial={{ opacity: 0, x: 50 }}
      whileInView={{ opacity: 1, x: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.8 }}
      className="relative"
    >
      <div className="absolute -inset-1 bg-gradient-to-r from-brand-500 to-purple-600 rounded-xl blur opacity-30"></div>
      <div className="relative bg-black/80 rounded-xl border border-white/10 overflow-hidden">
        <div className="flex items-center gap-2 px-4 py-3 border-b border-white/10 bg-white/5">
          <div className="w-3 h-3 rounded-full bg-red-500/80" />
          <div className="w-3 h-3 rounded-full bg-yellow-500/80" />
          <div className="w-3 h-3 rounded-full bg-green-500/80" />
          <span className="ml-4 text-xs text-white/40">graph.json</span>
        </div>
        <pre className="p-4 text-sm text-white/80 font-mono overflow-x-auto">
          <code>{code}</code>
        </pre>
      </div>
    </motion.div>
  );
}

// 3D Scene section with tabs
function SceneShowcase() {
  const [activeScene, setActiveScene] = useState('network');
  
  const scenes = {
    network: {
      title: 'Graph Network',
      description: 'Visualize your workflow as a living, breathing network of interconnected functions.',
      component: <GraphNetwork3D />,
    },
    stream: {
      title: 'Data Streams',
      description: 'Watch data flow through your pipeline in real-time with glowing streams.',
      component: <FlowingDataStream />,
    },
    crystal: {
      title: 'Crystal Graph',
      description: 'Explore the crystalline structure of your composed functions.',
      component: <CrystalGraph />,
    },
    swarm: {
      title: 'Node Swarm',
      description: 'See your functions as an organic swarm of intelligent agents.',
      component: <AnimatedNodeCluster />,
    },
  };

  return (
    <section className="py-24 px-4 md:px-8 lg:px-16 relative">
      <div className="max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-12"
        >
          <Badge variant="outline" className="mb-4 border-brand-500/50 text-brand-400">
            <Eye className="w-3 h-3 mr-1" />
            Interactive 3D
          </Badge>
          <h2 className="text-3xl md:text-5xl font-bold text-white mb-4">
            Visualize Your Logic
          </h2>
          <p className="text-white/60 max-w-2xl mx-auto">
            Experience your workflows in stunning 3D. Drag, zoom, rotate, and watch your graph come alive.
          </p>
        </motion.div>

        <Tabs value={activeScene} onValueChange={setActiveScene} className="w-full">
          <TabsList className="grid grid-cols-4 w-full max-w-2xl mx-auto mb-8 bg-white/5">
            <TabsTrigger value="network" className="text-xs data-[state=active]:bg-brand-500">
              Network
            </TabsTrigger>
            <TabsTrigger value="stream" className="text-xs data-[state=active]:bg-brand-500">
              Stream
            </TabsTrigger>
            <TabsTrigger value="crystal" className="text-xs data-[state=active]:bg-brand-500">
              Crystal
            </TabsTrigger>
            <TabsTrigger value="swarm" className="text-xs data-[state=active]:bg-brand-500">
              Swarm
            </TabsTrigger>
          </TabsList>
          
          {Object.entries(scenes).map(([key, scene]) => (
            <TabsContent key={key} value={key} className="mt-0">
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                animate={{ opacity: 1, scale: 1 }}
                transition={{ duration: 0.5 }}
                className="relative rounded-2xl overflow-hidden border border-white/10 bg-black/50"
                style={{ height: '600px' }}
              >
                <div className="absolute inset-0">
                  {scene.component}
                </div>
                <div className="absolute bottom-4 left-4 right-4 flex items-end justify-between">
                  <div className="bg-black/60 backdrop-blur-sm rounded-lg p-4 max-w-sm">
                    <h3 className="text-lg font-semibold text-white mb-1">{scene.title}</h3>
                    <p className="text-sm text-white/60">{scene.description}</p>
                  </div>
                  <div className="flex gap-2">
                    <Button variant="secondary" size="sm" className="bg-white/10 backdrop-blur-sm">
                      <MousePointerClick className="w-4 h-4 mr-2" />
                      Interactive
                    </Button>
                    <Button size="sm" className="bg-brand-500 hover:bg-brand-600">
                      <Maximize2 className="w-4 h-4 mr-2" />
                      Fullscreen
                    </Button>
                  </div>
                </div>
              </motion.div>
            </TabsContent>
          ))}
        </Tabs>
      </div>
    </section>
  );
}

// Architecture diagram component
function ArchitectureSection() {
  const steps = [
    { 
      icon: Share2, 
      title: 'Input',
      desc: 'Webhooks, APIs, schedules, or events',
      color: '#3b82f6'
    },
    { 
      icon: Network, 
      title: 'Graph Engine',
      desc: 'Orchestrates function execution',
      color: '#8b5cf6'
    },
    { 
      icon: Boxes, 
      title: 'Functions',
      desc: 'Composble, reusable logic blocks',
      color: '#10b981'
    },
    { 
      icon: Database, 
      title: 'State',
      desc: 'Persistent, versioned state fabric',
      color: '#f59e0b'
    },
    { 
      icon: Zap, 
      title: 'Output',
      desc: 'Responses, triggers, or side effects',
      color: '#ec4899'
    },
  ];

  return (
    <section className="py-24 px-4 md:px-8 lg:px-16 relative overflow-hidden">
      <div className="absolute inset-0 bg-gradient-to-b from-transparent via-brand-500/5 to-transparent" />
      
      <div className="max-w-7xl mx-auto relative">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <Badge variant="outline" className="mb-4 border-purple-500/50 text-purple-400">
            <Layers className="w-3 h-3 mr-1" />
            Architecture
          </Badge>
          <h2 className="text-3xl md:text-5xl font-bold text-white mb-4">
            How It Works
          </h2>
          <p className="text-white/60 max-w-2xl mx-auto">
            From input to output, every step is visual, debuggable, and optimized.
          </p>
        </motion.div>

        <div className="grid md:grid-cols-5 gap-4 relative">
          {/* Connecting line */}
          <div className="hidden md:block absolute top-12 left-0 right-0 h-0.5 bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 opacity-30" />
          
          {steps.map((step, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 30 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1 }}
              className="relative"
            >
              <div className="flex flex-col items-center text-center">
                <div 
                  className="w-24 h-24 rounded-2xl flex items-center justify-center mb-4 relative z-10 border-2 backdrop-blur-sm"
                  style={{ 
                    backgroundColor: `${step.color}10`,
                    borderColor: `${step.color}40`,
                    color: step.color
                  }}
                >
                  <step.icon className="w-10 h-10" />
                </div>
                <h3 className="text-lg font-semibold text-white mb-2">{step.title}</h3>
                <p className="text-sm text-white/60">{step.desc}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

// Main page component
export default function FRGShowcasePage() {
  const containerRef = useRef<HTMLDivElement>(null);
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end end"]
  });
  
  const y = useTransform(scrollYProgress, [0, 1], ["0%", "50%"]);
  const opacity = useTransform(scrollYProgress, [0, 0.2], [1, 0]);

  const features = [
    {
      icon: Workflow,
      title: 'Visual Graph Editor',
      description: 'Build complex workflows by dragging and connecting functions. No code required.',
      color: '#8b5cf6',
    },
    {
      icon: GitBranch,
      title: 'Versioned & Forkable',
      description: 'Every graph is versioned like code. Fork, modify, and merge with confidence.',
      color: '#3b82f6',
    },
    {
      icon: Zap,
      title: 'Live Execution',
      description: 'Watch data flow through your graph in real-time. Debug, pause, step through.',
      color: '#f59e0b',
    },
    {
      icon: Globe,
      title: 'Universal Runtime',
      description: 'Deploy anywhere. Edge, cloud, on-premise. Same graph, any environment.',
      color: '#10b981',
    },
    {
      icon: Cpu,
      title: 'AI-Powered',
      description: 'Get suggestions, auto-optimize, and let AI build graphs from descriptions.',
      color: '#ec4899',
    },
    {
      icon: Shield,
      title: 'DRE Verified',
      description: 'Deterministic Replay Execution. Reproduce any execution, any time, anywhere.',
      color: '#6366f1',
    },
  ];

  return (
    <div ref={containerRef} className="min-h-screen bg-black text-white overflow-x-hidden">
      {/* Background gradient */}
      <div className="fixed inset-0 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-purple-900/20 via-black to-black pointer-events-none" />
      
      {/* Navigation */}
      <nav className="fixed top-0 left-0 right-0 z-50 border-b border-white/10 bg-black/50 backdrop-blur-lg">
        <div className="max-w-7xl mx-auto px-4 md:px-8 h-16 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-brand-500 to-purple-600 flex items-center justify-center">
              <Share2 className="w-5 h-5 text-white" />
            </div>
            <span className="font-bold text-lg">FunctionFly</span>
            <Badge variant="secondary" className="ml-2 text-xs hidden sm:inline">FRG</Badge>
          </div>
          
          <div className="hidden md:flex items-center gap-6">
            <a href="#features" className="text-sm text-white/60 hover:text-white transition-colors">Features</a>
            <a href="#scenes" className="text-sm text-white/60 hover:text-white transition-colors">3D Scenes</a>
            <a href="#architecture" className="text-sm text-white/60 hover:text-white transition-colors">Architecture</a>
            <a href="#code" className="text-sm text-white/60 hover:text-white transition-colors">Code</a>
          </div>
          
          <div className="flex items-center gap-3">
            <Button variant="ghost" size="sm" className="hidden sm:flex">
              Documentation
            </Button>
            <Button size="sm" className="bg-brand-500 hover:bg-brand-600">
              Start Building
              <ArrowRight className="w-4 h-4 ml-2" />
            </Button>
            <Button 
              variant="ghost" 
              size="icon" 
              className="md:hidden"
              onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            >
              {mobileMenuOpen ? <X className="w-5 h-5" /> : <Menu className="w-5 h-5" />}
            </Button>
          </div>
        </div>
        
        {/* Mobile menu */}
        {mobileMenuOpen && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            className="md:hidden border-t border-white/10 bg-black/90 backdrop-blur-lg"
          >
            <div className="px-4 py-4 space-y-2">
              <a href="#features" className="block py-2 text-white/60 hover:text-white">Features</a>
              <a href="#scenes" className="block py-2 text-white/60 hover:text-white">3D Scenes</a>
              <a href="#architecture" className="block py-2 text-white/60 hover:text-white">Architecture</a>
              <a href="#code" className="block py-2 text-white/60 hover:text-white">Code</a>
            </div>
          </motion.div>
        )}
      </nav>

      {/* Hero Section */}
      <section className="relative min-h-screen flex items-center justify-center pt-16">
        <motion.div style={{ y, opacity }} className="absolute inset-0 z-0">
          <div className="absolute inset-0 opacity-40">
            <GraphNetwork3D />
          </div>
        </motion.div>
        
        <div className="relative z-10 max-w-5xl mx-auto px-4 text-center">
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
          >
            <Badge className="mb-6 bg-brand-500/20 text-brand-400 border-brand-500/30 hover:bg-brand-500/30">
              <Sparkles className="w-3 h-3 mr-1" />
              Now in Public Beta
            </Badge>
            <h1 className="text-5xl md:text-7xl font-bold mb-6 bg-clip-text text-transparent bg-gradient-to-r from-white via-brand-200 to-purple-200">
              Function Runtime Graph
            </h1>
            <p className="text-xl md:text-2xl text-white/60 max-w-3xl mx-auto mb-8">
              Build, visualize, and execute complex workflows as living, breathing 3D graphs. 
              The future of serverless is visual.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Button size="lg" className="bg-brand-500 hover:bg-brand-600 text-lg px-8">
                <Play className="w-5 h-5 mr-2" />
                Launch Editor
              </Button>
              <Button size="lg" variant="outline" className="text-lg px-8 border-white/20 hover:bg-white/10">
                <Code2 className="w-5 h-5 mr-2" />
                View on GitHub
              </Button>
            </div>
          </motion.div>

          {/* Stats */}
          <motion.div
            initial={{ opacity: 0, y: 40 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.3, duration: 0.8 }}
            className="grid grid-cols-2 md:grid-cols-4 gap-8 mt-20"
          >
            <AnimatedStat value={50000} label="Graphs Created" suffix="+" />
            <AnimatedStat value={2500000} label="Executions" suffix="+" />
            <AnimatedStat value={99.9} label="Uptime" suffix="%" />
            <AnimatedStat value={150} label="Functions" suffix="+" />
          </motion.div>
        </div>

        {/* Scroll indicator */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 1 }}
          className="absolute bottom-8 left-1/2 -translate-x-1/2"
        >
          <ChevronDown className="w-8 h-8 text-white/40 animate-bounce" />
        </motion.div>
      </section>

      {/* Features Section */}
      <section id="features" className="py-24 px-4 md:px-8 lg:px-16 relative">
        <div className="max-w-7xl mx-auto">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            className="text-center mb-16"
          >
            <Badge variant="outline" className="mb-4 border-purple-500/50 text-purple-400">
              <Zap className="w-3 h-3 mr-1" />
              Capabilities
            </Badge>
            <h2 className="text-3xl md:text-5xl font-bold text-white mb-4">
              Everything You Need
            </h2>
            <p className="text-white/60 max-w-2xl mx-auto">
              From visual building to live debugging, version control to AI assistance.
            </p>
          </motion.div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            {features.map((feature, i) => (
              <FeatureCard key={i} {...feature} delay={i * 0.1} />
            ))}
          </div>
        </div>
      </section>

      {/* 3D Scenes Section */}
      <section id="scenes">
        <SceneShowcase />
      </section>

      {/* Architecture Section */}
      <section id="architecture">
        <ArchitectureSection />
      </section>

      {/* Code + Visual Section */}
      <section id="code" className="py-24 px-4 md:px-8 lg:px-16 relative">
        <div className="max-w-7xl mx-auto">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            <motion.div
              initial={{ opacity: 0, x: -50 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.8 }}
            >
              <Badge variant="outline" className="mb-4 border-green-500/50 text-green-400">
                <Terminal className="w-3 h-3 mr-1" />
                Code-First
              </Badge>
              <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
                Code Meets Visual
              </h2>
              <p className="text-white/60 mb-6 text-lg">
                Define your graphs as JSON, YAML, or use our visual editor. 
                Changes sync bi-directionally. Your workflow, your way.
              </p>
              <ul className="space-y-3">
                {[
                  'JSON/YAML definitions',
                  'Git-based version control',
                  'CI/CD integration',
                  'Type-safe schemas',
                  'Import/export any format',
                ].map((item, i) => (
                  <li key={i} className="flex items-center gap-3 text-white/80">
                    <CheckCircle className="w-5 h-5 text-green-500" />
                    {item}
                  </li>
                ))}
              </ul>
              <div className="flex gap-3 mt-8">
                <Button className="bg-brand-500 hover:bg-brand-600">
                  <ArrowRight className="w-4 h-4 mr-2" />
                  Try the Editor
                </Button>
                <Button variant="outline" className="border-white/20">
                  <ExternalLink className="w-4 h-4 mr-2" />
                  Read Docs
                </Button>
              </div>
            </motion.div>
            
            <CodeSnippet />
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-24 px-4 md:px-8 lg:px-16 relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-r from-brand-500/10 via-purple-500/10 to-brand-500/10" />
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="max-w-4xl mx-auto text-center relative"
        >
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-4">
            Ready to Build Your First Graph?
          </h2>
          <p className="text-xl text-white/60 mb-8">
            Join thousands of developers building the future of serverless workflows.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button size="lg" className="bg-white text-black hover:bg-white/90 text-lg px-8">
              Get Started Free
              <ArrowUpRight className="w-5 h-5 ml-2" />
            </Button>
            <Button size="lg" variant="outline" className="border-white/30 text-lg px-8 hover:bg-white/10">
              <Play className="w-5 h-5 mr-2" />
              Watch Demo
            </Button>
          </div>
        </motion.div>
      </section>

      {/* Footer */}
      <footer className="border-t border-white/10 py-12 px-4 md:px-8 lg:px-16">
        <div className="max-w-7xl mx-auto">
          <div className="grid md:grid-cols-4 gap-8 mb-8">
            <div>
              <div className="flex items-center gap-2 mb-4">
                <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-brand-500 to-purple-600 flex items-center justify-center">
                  <Share2 className="w-5 h-5 text-white" />
                </div>
                <span className="font-bold">FunctionFly</span>
              </div>
              <p className="text-sm text-white/40">
                The visual platform for building, deploying, and executing function graphs.
              </p>
            </div>
            
            <div>
              <h4 className="font-semibold mb-4">Product</h4>
              <ul className="space-y-2 text-sm text-white/40">
                <li><a href="#" className="hover:text-white transition-colors">Graph Editor</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Marketplace</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Enterprise</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Pricing</a></li>
              </ul>
            </div>
            
            <div>
              <h4 className="font-semibold mb-4">Resources</h4>
              <ul className="space-y-2 text-sm text-white/40">
                <li><a href="#" className="hover:text-white transition-colors">Documentation</a></li>
                <li><a href="#" className="hover:text-white transition-colors">API Reference</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Tutorials</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Blog</a></li>
              </ul>
            </div>
            
            <div>
              <h4 className="font-semibold mb-4">Company</h4>
              <ul className="space-y-2 text-sm text-white/40">
                <li><a href="#" className="hover:text-white transition-colors">About</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Careers</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Contact</a></li>
                <li><a href="#" className="hover:text-white transition-colors">Privacy</a></li>
              </ul>
            </div>
          </div>
          
          <Separator className="bg-white/10 mb-8" />
          
          <div className="flex flex-col md:flex-row items-center justify-between gap-4">
            <p className="text-sm text-white/40">
              © 2024 FunctionFly. All rights reserved.
            </p>
            <div className="flex gap-4">
              <Button variant="ghost" size="icon" className="text-white/40 hover:text-white">
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                  <path d={siGithub.path} />
                </svg>
              </Button>
              <Button variant="ghost" size="icon" className="text-white/40 hover:text-white">
                <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 24 24">
                  <path d={siX.path} />
                </svg>
              </Button>
              <Button variant="ghost" size="icon" className="text-white/40 hover:text-white">
                <Disc className="w-5 h-5" />
              </Button>
            </div>
          </div>
        </div>
      </footer>
    </div>
  );
}
