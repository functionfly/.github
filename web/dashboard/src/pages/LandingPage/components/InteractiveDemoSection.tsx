import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Play, Pause, RotateCcw, Code, BarChart3, Calculator, Activity, AlertTriangle, CheckCircle, Zap } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
// Using native input for slider functionality
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, AreaChart, Area } from 'recharts';
import { useSpring, animated } from '@react-spring/web';

// Mock data for status dashboard
const generateMetricsData = () => {
  const data = [];
  for (let i = 0; i < 24; i++) {
    data.push({
      time: `${i}:00`,
      invocations: Math.floor(Math.random() * 1000) + 500,
      errors: Math.floor(Math.random() * 50),
      latency: Math.floor(Math.random() * 200) + 50,
      status: Math.random() > 0.95 ? 'error' : 'healthy'
    });
  }
  return data;
};

// Code sandbox with failover simulation
function CodeSandboxDemo() {
  const [isRunning, setIsRunning] = useState(false);
  const [hasFailover, setHasFailover] = useState(false);
  const [logs, setLogs] = useState<string[]>([]);

  const sampleCode = `// FunctionFly Serverless Function
export async function handler(event, context) {
  console.log('Processing request...', event);

  // Simulate processing
  await new Promise(resolve => setTimeout(resolve, 100));

  return {
    statusCode: 200,
    body: JSON.stringify({
      message: 'Hello from FunctionFly!',
      timestamp: new Date().toISOString()
    })
  };
}`;

  const simulateFailover = () => {
    setHasFailover(true);
    setLogs(prev => [...prev,
      '[INFO] Primary region experiencing high latency',
      '[WARN] Switching to failover region...',
      '[SUCCESS] Failover completed in 2.3s',
      '[INFO] Traffic redirected to backup instances'
    ]);
  };

  const runSimulation = () => {
    setIsRunning(true);
    setLogs([]);

    const events = [
      '[START] Function initialized',
      '[INFO] Connecting to database...',
      '[SUCCESS] Database connection established',
      '[INFO] Processing 100 concurrent requests...',
      '[METRICS] Avg latency: 45ms',
      '[SUCCESS] All requests processed successfully'
    ];

    events.forEach((event, index) => {
      setTimeout(() => {
        setLogs(prev => [...prev, event]);
      }, index * 800);
    });

    setTimeout(() => {
      setIsRunning(false);
      if (Math.random() > 0.7) {
        simulateFailover();
      }
    }, events.length * 800 + 1000);
  };

  return (
    <Card className="border-white/8 bg-white/5 card-elevation glass-card shine-effect">
      <CardHeader>
        <CardTitle className="text-text-primary flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
          <Code className="w-5 h-5" />
          Code Sandbox - Failover Simulation
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="bg-black/50 rounded-lg p-4 font-mono text-sm text-green-400 min-h-[200px]">
          <div className="text-gray-400 mb-2">// Sample FunctionFly Function</div>
          {sampleCode.split('\n').map((line, i) => (
            <div key={i} className="leading-relaxed">
              <span className="text-gray-500 mr-4">{(i + 1).toString().padStart(2, ' ')}</span>
              {line}
            </div>
          ))}
        </div>

        <div className="flex gap-2">
          <Button
            onClick={runSimulation}
            disabled={isRunning}
            className="bg-[#6366f1] hover:bg-[#6366f1]/80"
          >
            {isRunning ? <Pause className="w-4 h-4 mr-2" /> : <Play className="w-4 h-4 mr-2" />}
            {isRunning ? 'Running...' : 'Run Simulation'}
          </Button>
          <Button
            variant="outline"
            onClick={() => { setLogs([]); setHasFailover(false); }}
            className="interactive-demo-reset-btn border-border-default text-text-primary hover:bg-bg-hover" style={{ color: 'var(--text-primary)', borderColor: 'var(--border-default)' }}
          >
            <RotateCcw className="w-4 h-4 mr-2" />
            Reset
          </Button>
        </div>

        <div className="bg-black/30 rounded-lg p-4 min-h-[150px] font-mono text-xs">
          <div className="text-gray-400 mb-2">Console Output:</div>
          <div className="space-y-1 max-h-[120px] overflow-y-auto">
            {logs.map((log, i) => (
              <div key={i} className={`${
                log.includes('ERROR') || log.includes('WARN') ? 'text-yellow-400' :
                log.includes('SUCCESS') ? 'text-green-400' : 'text-blue-400'
              }`}>
                {log}
              </div>
            ))}
            {logs.length === 0 && !isRunning && (
              <div className="text-gray-500 italic">Click "Run Simulation" to see FunctionFly in action</div>
            )}
          </div>
        </div>

        {hasFailover && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="bg-yellow-500/10 border border-yellow-500/20 rounded-lg p-3"
          >
            <div className="flex items-center gap-2 text-yellow-400 text-sm">
              <AlertTriangle className="w-4 h-4" />
              Failover Event Detected - Automatic recovery initiated
            </div>
          </motion.div>
        )}
      </CardContent>
    </Card>
  );
}

// Status dashboard preview
function StatusDashboardDemo() {
  const [metricsData, setMetricsData] = useState(generateMetricsData());
  const [isLive, setIsLive] = useState(true);

  useEffect(() => {
    if (!isLive) return;

    const interval = setInterval(() => {
      setMetricsData(prev => {
        const newData = [...prev.slice(1)];
        newData.push({
          time: new Date().toLocaleTimeString(),
          invocations: Math.floor(Math.random() * 1000) + 500,
          errors: Math.floor(Math.random() * 50),
          latency: Math.floor(Math.random() * 200) + 50,
          status: Math.random() > 0.95 ? 'error' : 'healthy'
        });
        return newData;
      });
    }, 3000);

    return () => clearInterval(interval);
  }, [isLive]);

  const currentMetrics = metricsData[metricsData.length - 1] || {
    invocations: 0, errors: 0, latency: 0, status: 'healthy'
  };

  return (
    <Card className="border-white/8 bg-white/5 card-elevation glass-card shine-effect">
      <CardHeader>
        <CardTitle className="text-text-primary flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
          <BarChart3 className="w-5 h-5" />
          Live Status Dashboard
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Key Metrics */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <motion.div
            whileHover={{ scale: 1.05 }}
            className="bg-white/5 rounded-lg p-4 text-center"
          >
            <div className="text-2xl font-bold text-[#6366f1]">{currentMetrics.invocations}</div>
            <div className="text-xs text-text-secondary">Invocations/min</div>
          </motion.div>
          <motion.div
            whileHover={{ scale: 1.05 }}
            className="bg-white/5 rounded-lg p-4 text-center"
          >
            <div className="text-2xl font-bold text-green-400">{currentMetrics.latency}ms</div>
            <div className="text-xs text-text-secondary">Avg Latency</div>
          </motion.div>
          <motion.div
            whileHover={{ scale: 1.05 }}
            className="bg-white/5 rounded-lg p-4 text-center"
          >
            <div className="text-2xl font-bold text-red-400">{currentMetrics.errors}</div>
            <div className="text-xs text-text-secondary">Errors</div>
          </motion.div>
          <motion.div
            whileHover={{ scale: 1.05 }}
            className="bg-white/5 rounded-lg p-4 text-center"
          >
            <div className={`text-2xl font-bold flex items-center justify-center gap-1 ${
              currentMetrics.status === 'healthy' ? 'text-green-400' : 'text-red-400'
            }`}>
              <Activity className="w-4 h-4" />
              {currentMetrics.status === 'healthy' ? 'Healthy' : 'Issues'}
            </div>
            <div className="text-xs text-text-secondary">Status</div>
          </motion.div>
        </div>

        {/* Charts */}
        <div className="grid md:grid-cols-2 gap-6">
          <div>
            <h4 className="text-text-primary font-medium mb-3">Invocations Over Time</h4>
            <ResponsiveContainer width="100%" height={150}>
              <AreaChart data={metricsData}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" />
                <XAxis dataKey="time" stroke="var(--text-muted)" fontSize={12} />
                <YAxis stroke="var(--text-muted)" fontSize={12} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'var(--bg-elevated)',
                    border: '1px solid var(--border-default)',
                    borderRadius: '6px',
                    color: 'var(--text-primary)'
                  }}
                />
                <Area
                  type="monotone"
                  dataKey="invocations"
                  stroke="var(--brand-500)"
                  fill="var(--brand-500)"
                  fillOpacity={0.3}
                  isAnimationActive={false}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>

          <div>
            <h4 className="text-text-primary font-medium mb-3">Latency Trends</h4>
            <ResponsiveContainer width="100%" height={150}>
              <LineChart data={metricsData}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--border-subtle)" />
                <XAxis dataKey="time" stroke="var(--text-muted)" fontSize={12} />
                <YAxis stroke="var(--text-muted)" fontSize={12} />
                <Tooltip
                  contentStyle={{
                    backgroundColor: 'var(--bg-elevated)',
                    border: '1px solid var(--border-default)',
                    borderRadius: '6px',
                    color: 'var(--text-primary)'
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="latency"
                  stroke="var(--success)"
                  strokeWidth={2}
                  dot={false}
                  isAnimationActive={false}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="flex justify-center">
          <Button
            onClick={() => setIsLive(!isLive)}
            variant="outline"
            className="border-border-default text-text-primary hover:bg-bg-hover" style={{ color: 'var(--text-primary)', borderColor: 'var(--border-default)' }}
          >
            {isLive ? <Pause className="w-4 h-4 mr-2" /> : <Play className="w-4 h-4 mr-2" />}
            {isLive ? 'Pause Live Data' : 'Resume Live Data'}
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}

// Interactive pricing calculator
function PricingCalculatorDemo() {
  const [invocations, setInvocations] = useState([1000000]);
  const [plan, setPlan] = useState<'free' | 'developer' | 'professional' | 'enterprise'>('professional');

  const plans = {
    free: { name: 'Free', basePrice: 0, includedInvocations: 500 },
    developer: { name: 'Developer', basePrice: 9, includedInvocations: 1000000 },
    professional: { name: 'Professional', basePrice: 29, includedInvocations: 5000000 },
    enterprise: { name: 'Enterprise', basePrice: 99, includedInvocations: 25000000 }
  };

  const currentPlan = plans[plan];
  const totalInvocations = invocations[0];
  const overageInvocations = Math.max(0, totalInvocations - currentPlan.includedInvocations);
  const overageCost = overageInvocations * 0.00002; // $0.02 per 1000 invocations overage
  const totalCost = currentPlan.basePrice + overageCost;

  const springProps = useSpring({
    totalCost: totalCost,
    config: { tension: 300, friction: 20 }
  });

  return (
    <Card className="border-white/8 bg-white/5 card-elevation glass-card shine-effect">
      <CardHeader>
        <CardTitle className="text-text-primary flex items-center gap-2" style={{ color: 'var(--text-primary)' }}>
          <Calculator className="w-5 h-5" />
          Interactive Pricing Calculator
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-6">
        {/* Plan Selection */}
        <div>
          <label className="text-text-primary font-medium mb-3 block" style={{ color: 'var(--text-primary)' }}>Select Plan</label>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
            {Object.entries(plans).map(([key, planData]) => (
              <Button
                key={key}
                variant={plan === key ? "default" : "outline"}
                size="sm"
                onClick={() => setPlan(key as typeof plan)}
                className={plan === key ?
                  "bg-[#6366f1] border-[#6366f1]" :
                  "border-border-default text-text-primary hover:bg-bg-hover"}
                style={plan === key ? {} : { color: 'var(--text-primary)', borderColor: 'var(--border-default)' }}
              >
                {planData.name}
              </Button>
            ))}
          </div>
        </div>

        {/* Invocations Slider */}
        <div>
          <label className="text-text-primary font-medium mb-3 block" style={{ color: 'var(--text-primary)' }}>
            Monthly Invocations: {totalInvocations.toLocaleString()}
          </label>
          <input
            type="range"
            min={100000}
            max={50000000}
            step={100000}
            value={invocations[0]}
            onChange={(e) => setInvocations([parseInt(e.target.value)])}
            className="w-full h-2 bg-white/20 rounded-lg appearance-none cursor-pointer slider"
          />
          <div className="flex justify-between text-xs text-text-secondary mt-1">
            <span>100K</span>
            <span>50M</span>
          </div>
        </div>

        {/* Cost Breakdown */}
        <div className="bg-white/5 rounded-lg p-4 space-y-3">
          <div className="flex justify-between text-text-primary" style={{ color: 'var(--text-primary)' }}>
            <span>{currentPlan.name} Plan</span>
            <span>${currentPlan.basePrice}/month</span>
          </div>

          {overageInvocations > 0 && (
            <div className="flex justify-between text-yellow-400 text-sm">
              <span>Overage ({(overageInvocations / 1000).toFixed(0)}K invocations)</span>
              <span>${overageCost.toFixed(2)}</span>
            </div>
          )}

          <div className="border-t border-white/10 pt-3">
            <div className="flex justify-between text-text-primary font-bold text-lg" style={{ color: 'var(--text-primary)' }}>
              <span>Total Monthly Cost</span>
              <animated.span>
                {springProps.totalCost.to(val => `$${val.toFixed(2)}`)}
              </animated.span>
            </div>
          </div>
        </div>

        {/* Included Features */}
        <div className="grid md:grid-cols-2 gap-4 text-sm">
          <div className="space-y-2">
            <h4 className="text-text-primary font-medium" style={{ color: 'var(--text-primary)' }}>What's Included:</h4>
            <ul className="space-y-1 text-text-secondary">
              <li className="flex items-center gap-2">
                <CheckCircle className="w-3 h-3 text-green-400" />
                {currentPlan.includedInvocations.toLocaleString()} invocations
              </li>
              <li className="flex items-center gap-2">
                <CheckCircle className="w-3 h-3 text-green-400" />
                Real-time monitoring
              </li>
              <li className="flex items-center gap-2">
                <CheckCircle className="w-3 h-3 text-green-400" />
                Automatic failover
              </li>
            </ul>
          </div>

          <div className="space-y-2">
            <h4 className="text-text-primary font-medium" style={{ color: 'var(--text-primary)' }}>Additional Perks:</h4>
            <ul className="space-y-1 text-text-secondary">
              <li className="flex items-center gap-2">
                <Zap className="w-3 h-3 text-yellow-400" />
                99.9% uptime SLA
              </li>
              <li className="flex items-center gap-2">
                <Zap className="w-3 h-3 text-yellow-400" />
                24/7 support
              </li>
              <li className="flex items-center gap-2">
                <Zap className="w-3 h-3 text-yellow-400" />
                Advanced analytics
              </li>
            </ul>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function InteractiveDemoSection() {
  return (
    <section className="py-20 border-t border-white/8 aurora-bg interactive-demo-enhanced">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        <div className="text-center mb-16">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
          >
            <div className="w-16 h-16 mx-auto mb-6 rounded-2xl bg-linear-to-br from-[#6366f1]/30 to-[#8b5cf6]/30 border border-[#6366f1]/40 flex items-center justify-center glow">
              <Activity className="w-8 h-8 text-white drop-shadow-lg" />
            </div>
            <h2 className="text-3xl font-bold text-text-primary mb-4 relative z-10" style={{ fontWeight: 800 }}>
              Experience FunctionFly Live
            </h2>
            <p className="text-text-secondary max-w-2xl mx-auto relative z-10" style={{ fontWeight: 500 }}>
              Interact with real FunctionFly features. Test failover scenarios, monitor live metrics,
              and calculate your perfect pricing plan - all in your browser.
            </p>
          </motion.div>
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 0.5, delay: 0.2 }}
        >
          <Tabs defaultValue="code" className="w-full">
            <TabsList className="interactive-demo-tabs grid w-full grid-cols-3 bg-white/5 border border-white/8">
              <TabsTrigger
                value="code"
                className="interactive-demo-tab data-[state=active]:bg-[#6366f1] data-[state=active]:text-white text-text-secondary"
              >
                <Code className="w-4 h-4 mr-2" />
                Code Sandbox
              </TabsTrigger>
              <TabsTrigger
                value="dashboard"
                className="interactive-demo-tab data-[state=active]:bg-[#6366f1] data-[state=active]:text-white text-text-secondary"
              >
                <BarChart3 className="w-4 h-4 mr-2" />
                Status Dashboard
              </TabsTrigger>
              <TabsTrigger
                value="pricing"
                className="interactive-demo-tab data-[state=active]:bg-[#6366f1] data-[state=active]:text-white text-text-secondary"
              >
                <Calculator className="w-4 h-4 mr-2" />
                Pricing Calculator
              </TabsTrigger>
            </TabsList>

            <div className="mt-8">
              <TabsContent value="code">
                <CodeSandboxDemo />
              </TabsContent>

              <TabsContent value="dashboard">
                <StatusDashboardDemo />
              </TabsContent>

              <TabsContent value="pricing">
                <PricingCalculatorDemo />
              </TabsContent>
            </div>
          </Tabs>
        </motion.div>
      </div>
    </section>
  );
}
