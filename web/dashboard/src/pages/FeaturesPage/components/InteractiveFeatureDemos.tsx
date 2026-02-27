import { motion } from "framer-motion";
import { useState, useEffect } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { ArrowRightLeft, Network, Brain, Globe, Activity, CheckCircle, AlertTriangle, Server } from "lucide-react";

// Predictive Routing Demo Component
const PredictiveRoutingDemo = () => {
  const [trafficFlow, setTrafficFlow] = useState([
    { id: 1, provider: "Vercel", status: "healthy", load: 75, active: true },
    { id: 2, provider: "Fly.io", status: "healthy", load: 45, active: false },
    { id: 3, provider: "Cloudflare", status: "healthy", load: 60, active: false },
  ]);

  const [currentRequest, setCurrentRequest] = useState(0);
  const [isRouting, setIsRouting] = useState(false);

  useEffect(() => {
    const interval = setInterval(() => {
      if (!isRouting) {
        setIsRouting(true);
        setCurrentRequest(prev => prev + 1);

        // Simulate AI routing decision
        setTimeout(() => {
          setTrafficFlow(prev => prev.map(provider => ({
            ...provider,
            active: false
          })));

          // Choose provider with lowest load (simulating AI decision)
          const bestProvider = [...trafficFlow].sort((a, b) => a.load - b.load)[0];
          setTrafficFlow(prev => prev.map(provider =>
            provider.id === bestProvider.id
              ? { ...provider, active: true, load: Math.min(provider.load + 10, 95) }
              : provider
          ));

          setTimeout(() => setIsRouting(false), 1000);
        }, 500);
      }
    }, 3000);

    return () => clearInterval(interval);
  }, [isRouting, trafficFlow]);

  return (
    <Card className="h-full">
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center gap-2 text-lg">
          <Brain className="w-5 h-5 text-[#10b981]" />
          Predictive Routing Demo
        </CardTitle>
        <p className="text-sm text-text-secondary">
          Watch AI-powered traffic routing in real-time
        </p>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="flex items-center justify-between">
          <span className="text-sm font-medium">Requests Processed</span>
          <Badge variant="secondary">{currentRequest}</Badge>
        </div>

        <div className="space-y-3">
          {trafficFlow.map((provider) => (
            <motion.div
              key={provider.id}
              className={`p-3 rounded-lg border transition-all duration-300 ${
                provider.active
                  ? 'border-[#10b981] bg-[#10b981]/5'
                  : 'border-white/10 bg-white/5'
              }`}
              animate={provider.active ? { scale: 1.02 } : { scale: 1 }}
            >
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  <Server className="w-4 h-4" />
                  <span className="font-medium">{provider.provider}</span>
                  <div className={`w-2 h-2 rounded-full ${
                    provider.status === 'healthy' ? 'bg-green-500' : 'bg-red-500'
                  }`} />
                </div>
                <span className="text-sm text-text-secondary">{provider.load}% load</span>
              </div>
              <div className="w-full bg-white/10 rounded-full h-2">
                <motion.div
                  className={`h-2 rounded-full ${
                    provider.active ? 'bg-[#10b981]' : 'bg-white/30'
                  }`}
                  initial={{ width: 0 }}
                  animate={{ width: `${provider.load}%` }}
                  transition={{ duration: 0.5 }}
                />
              </div>
            </motion.div>
          ))}
        </div>

        <div className="text-xs text-text-secondary text-center">
          AI automatically routes to the healthiest provider
        </div>
      </CardContent>
    </Card>
  );
};

// Multi-Provider Deployment Demo Component
const MultiProviderDeploymentDemo = () => {
  const [deploymentStatus, setDeploymentStatus] = useState({
    vercel: { status: 'idle', progress: 0 },
    fly: { status: 'idle', progress: 0 },
    cloudflare: { status: 'idle', progress: 0 },
  });

  const [isDeploying, setIsDeploying] = useState(false);

  const startDeployment = () => {
    if (isDeploying) return;

    setIsDeploying(true);
    setDeploymentStatus({
      vercel: { status: 'deploying', progress: 0 },
      fly: { status: 'pending', progress: 0 },
      cloudflare: { status: 'pending', progress: 0 },
    });

    // Simulate staggered deployment
    const providers = ['vercel', 'fly', 'cloudflare'];
    let currentIndex = 0;

    const deployNext = () => {
      if (currentIndex >= providers.length) {
        setIsDeploying(false);
        return;
      }

      const provider = providers[currentIndex];
      setDeploymentStatus(prev => ({
        ...prev,
        [provider]: { status: 'deploying', progress: 0 }
      }));

      // Simulate deployment progress
      const progressInterval = setInterval(() => {
        setDeploymentStatus(prev => {
          const currentProgress = prev[provider].progress;
          if (currentProgress >= 100) {
            clearInterval(progressInterval);
            setDeploymentStatus(prev => ({
              ...prev,
              [provider]: { status: 'completed', progress: 100 }
            }));
            currentIndex++;
            setTimeout(deployNext, 500);
            return prev;
          }
          return {
            ...prev,
            [provider]: { status: 'deploying', progress: currentProgress + 8 }
          };
        });
      }, 50);
    };

    deployNext();
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="w-4 h-4 text-green-500" />;
      case 'deploying':
        return <Activity className="w-4 h-4 text-blue-500 animate-pulse" />;
      case 'failed':
        return <AlertTriangle className="w-4 h-4 text-red-500" />;
      default:
        return <div className="w-4 h-4 rounded-full border-2 border-white/30" />;
    }
  };

  return (
    <Card className="h-full">
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center gap-2 text-lg">
          <Network className="w-5 h-5 text-[#8b5cf6]" />
          Multi-Provider Deployment
        </CardTitle>
        <p className="text-sm text-text-secondary">
          Deploy to multiple providers simultaneously
        </p>
      </CardHeader>
      <CardContent className="space-y-4">
        <Button
          onClick={startDeployment}
          disabled={isDeploying}
          className="w-full bg-[#8b5cf6] hover:bg-[#8b5cf6]/90"
        >
          {isDeploying ? 'Deploying...' : 'Start Multi-Provider Deployment'}
        </Button>

        <div className="space-y-3">
          {Object.entries(deploymentStatus).map(([provider, { status, progress }]) => (
            <div key={provider} className="p-3 rounded-lg border border-white/10 bg-white/5">
              <div className="flex items-center justify-between mb-2">
                <div className="flex items-center gap-2">
                  {getStatusIcon(status)}
                  <span className="font-medium capitalize">{provider}</span>
                </div>
                <Badge
                  variant="secondary"
                  className={`text-xs ${
                    status === 'completed' ? 'bg-green-500/20 text-green-400' :
                    status === 'deploying' ? 'bg-blue-500/20 text-blue-400' :
                    'bg-white/20'
                  }`}
                >
                  {status}
                </Badge>
              </div>
              <div className="w-full bg-white/10 rounded-full h-2">
                <motion.div
                  className={`h-2 rounded-full ${
                    status === 'completed' ? 'bg-green-500' :
                    status === 'deploying' ? 'bg-blue-500' :
                    'bg-white/30'
                  }`}
                  animate={{ width: `${progress}%` }}
                  transition={{ duration: 0.1 }}
                />
              </div>
              <div className="text-xs text-text-secondary mt-1">
                {status === 'deploying' ? `${progress}% complete` :
                 status === 'completed' ? 'Ready' : 'Waiting'}
              </div>
            </div>
          ))}
        </div>

        <div className="text-xs text-text-secondary text-center">
          Zero-downtime deployments across all providers
        </div>
      </CardContent>
    </Card>
  );
};

// Global Edge Network Demo Component
const GlobalEdgeNetworkDemo = () => {
  const [activeNodes, setActiveNodes] = useState<Set<number>>(new Set());
  const [totalRequests, setTotalRequests] = useState(0);

  // Simulate global edge locations
  const edgeLocations = [
    { id: 1, name: "North America", region: "NA", latency: 12, requests: 0 },
    { id: 2, name: "Europe", region: "EU", latency: 45, requests: 0 },
    { id: 3, name: "Asia Pacific", region: "APAC", latency: 120, requests: 0 },
    { id: 4, name: "South America", region: "SA", latency: 85, requests: 0 },
    { id: 5, name: "Africa", region: "AF", latency: 95, requests: 0 },
    { id: 6, name: "Middle East", region: "ME", latency: 75, requests: 0 },
  ];

  const [locations, setLocations] = useState(edgeLocations);

  useEffect(() => {
    const interval = setInterval(() => {
      // Simulate requests hitting different edge locations
      const randomLocation = Math.floor(Math.random() * locations.length);
      setLocations(prev => prev.map((loc, index) =>
        index === randomLocation
          ? { ...loc, requests: loc.requests + 1 }
          : loc
      ));

      setActiveNodes(prev => {
        const newSet = new Set(prev);
        newSet.add(randomLocation + 1);
        setTimeout(() => {
          setActiveNodes(current => {
            const updated = new Set(current);
            updated.delete(randomLocation + 1);
            return updated;
          });
        }, 1000);
        return newSet;
      });

      setTotalRequests(prev => prev + 1);
    }, 800);

    return () => clearInterval(interval);
  }, [locations.length]);

  return (
    <Card className="h-full">
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center gap-2 text-lg">
          <Globe className="w-5 h-5 text-[#f97316]" />
          Global Edge Network
        </CardTitle>
        <p className="text-sm text-text-secondary">
          Real-time traffic distribution across 200+ locations
        </p>
      </CardHeader>
      <CardContent className="space-y-4">
        <div className="grid grid-cols-2 gap-2">
          {locations.map((location) => (
            <motion.div
              key={location.id}
              className={`p-2 rounded border text-center transition-all duration-300 ${
                activeNodes.has(location.id)
                  ? 'border-[#f97316] bg-[#f97316]/10'
                  : 'border-white/10 bg-white/5'
              }`}
              animate={activeNodes.has(location.id) ? { scale: 1.05 } : { scale: 1 }}
            >
              <div className="text-xs font-medium">{location.region}</div>
              <div className="text-lg font-bold text-[#f97316]">{location.requests}</div>
              <div className="text-xs text-text-secondary">{location.latency}ms</div>
            </motion.div>
          ))}
        </div>

        <div className="flex items-center justify-between text-sm">
          <span>Total Requests</span>
          <Badge variant="secondary">{totalRequests}</Badge>
        </div>

        <div className="text-xs text-text-secondary text-center">
          Requests automatically routed to nearest edge location
        </div>
      </CardContent>
    </Card>
  );
};

// Fast Failover Demo Component
const FastFailoverDemo = () => {
  const [currentProvider, setCurrentProvider] = useState(1);
  const [isFailing, setIsFailing] = useState(false);
  const [failoverCount, setFailoverCount] = useState(0);

  const providers = [
    { id: 1, name: "Primary", status: "active", color: "#10b981" },
    { id: 2, name: "Secondary", status: "standby", color: "#f59e0b" },
    { id: 3, name: "Tertiary", status: "standby", color: "#ef4444" },
  ];

  const triggerFailover = () => {
    if (isFailing) return;

    setIsFailing(true);
    const nextProvider = (currentProvider % providers.length) + 1;

    // Simulate failure
    setTimeout(() => {
      setCurrentProvider(nextProvider);
      setFailoverCount(prev => prev + 1);
      setIsFailing(false);
    }, 150); // 150ms failover time
  };

  return (
    <Card className="h-full">
      <CardHeader className="pb-4">
        <CardTitle className="flex items-center gap-2 text-lg">
          <ArrowRightLeft className="w-5 h-5 text-[#f59e0b]" />
          Fast Failover Demo
        </CardTitle>
        <p className="text-sm text-text-secondary">
          Sub-second provider switching simulation
        </p>
      </CardHeader>
      <CardContent className="space-y-4">
        <Button
          onClick={triggerFailover}
          disabled={isFailing}
          className="w-full bg-[#f59e0b] hover:bg-[#f59e0b]/90"
        >
          {isFailing ? 'Failing Over...' : 'Trigger Failover'}
        </Button>

        <div className="space-y-2">
          {providers.map((provider) => (
            <motion.div
              key={provider.id}
              className={`p-3 rounded-lg border transition-all duration-300 ${
                provider.id === currentProvider
                  ? 'border-[#f59e0b] bg-[#f59e0b]/10'
                  : 'border-white/10 bg-white/5'
              }`}
              animate={{
                scale: provider.id === currentProvider ? 1.02 : 1,
                opacity: provider.id === currentProvider ? 1 : 0.7
              }}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <div
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: provider.color }}
                  />
                  <span className="font-medium">{provider.name}</span>
                </div>
                <Badge
                  variant="secondary"
                  className={provider.id === currentProvider ? 'bg-[#f59e0b]/20 text-[#f59e0b]' : ''}
                >
                  {provider.id === currentProvider ? 'Active' : 'Standby'}
                </Badge>
              </div>
            </motion.div>
          ))}
        </div>

        <div className="flex items-center justify-between text-sm">
          <span>Failovers Completed</span>
          <Badge variant="secondary">{failoverCount}</Badge>
        </div>

        <div className="text-xs text-text-secondary text-center">
          Failover completed in {isFailing ? 'processing...' : '<150ms'}
        </div>
      </CardContent>
    </Card>
  );
};

export function InteractiveFeatureDemos() {
  return (
    <motion.section
      initial={{ opacity: 0, y: 20 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true }}
      transition={{ duration: 0.6 }}
      className="py-20 border-t border-white/8"
    >
      <div className="text-center mb-16">
        <h2 className="text-3xl font-bold text-white mb-4">
          Experience Our Features in Action
        </h2>
        <p className="text-text-secondary max-w-2xl mx-auto">
          Interactive demonstrations showing how our key features work in real-time.
          Click the buttons to see intelligent routing, multi-provider deployments,
          and global edge network performance.
        </p>
      </div>

      <div className="grid md:grid-cols-2 gap-8">
        <PredictiveRoutingDemo />
        <MultiProviderDeploymentDemo />
        <GlobalEdgeNetworkDemo />
        <FastFailoverDemo />
      </div>
    </motion.section>
  );
}