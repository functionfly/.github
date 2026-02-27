import React from "react";
import { motion } from "framer-motion";
import { Icon } from "@iconify/react";
// @ts-ignore - react-aws-icons doesn't have TypeScript definitions
import { default as AwsLambda } from "react-aws-icons/dist/aws/logo/Lambda";
import { RailwayIcon, DjangoIcon, RenderIcon, FlyIoIcon, HerokuIcon, VercelIcon, NetlifyIcon, FastifyIcon, AzureIcon, NextJsIcon, ExpressIcon } from "./icons";
import { Globe } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

type IconType = React.ComponentType<any> | string | { hex: string; title: string };

interface PartnerItem {
  name: string;
  icon: IconType;
  category: string;
}

interface IntegrationPartners {
  cloudProviders: PartnerItem[];
  frameworks: PartnerItem[];
  deploymentPlatforms: PartnerItem[];
}

const integrationPartners: IntegrationPartners = {
  cloudProviders: [
    {
      name: "AWS Lambda",
      icon: AwsLambda,
      category: "Serverless"
    },
    {
      name: "Google Cloud Functions",
      icon: "logos:google-cloud-functions",
      category: "Serverless"
    },
    {
      name: "Azure Functions",
      icon: AzureIcon,
      category: "Serverless"
    },
    {
      name: "Cloudflare Workers",
      icon: "simple-icons:cloudflareworkers",
      category: "Edge"
    },
    {
      name: "Vercel Functions",
      icon: VercelIcon,
      category: "Edge"
    },
    {
      name: "Netlify Functions",
      icon: NetlifyIcon,
      category: "Edge"
    },
  ],
  frameworks: [
    {
      name: "React",
      icon: "logos:react",
      category: "Frontend"
    },
    {
      name: "Next.js",
      icon: NextJsIcon,
      category: "React Framework"
    },
    {
      name: "Vue.js",
      icon: "logos:vue",
      category: "Frontend"
    },
    {
      name: "Svelte",
      icon: "logos:svelte-icon",
      category: "Frontend"
    },
    {
      name: "Express.js",
      icon: ExpressIcon,
      category: "Backend"
    },
    {
      name: "Fastify",
      icon: FastifyIcon,
      category: "Backend"
    },
    {
      name: "NestJS",
      icon: "logos:nestjs",
      category: "Backend"
    },
    {
      name: "Django",
      icon: DjangoIcon,
      category: "Backend"
    },
  ],
  deploymentPlatforms: [
    {
      name: "Vercel",
      icon: VercelIcon,
      category: "Deployment"
    },
    {
      name: "Netlify",
      icon: NetlifyIcon,
      category: "Deployment"
    },
    {
      name: "Railway",
      icon: RailwayIcon,
      category: "Deployment"
    },
    {
      name: "Render",
      icon: RenderIcon,
      category: "Deployment"
    },
    {
      name: "Fly.io",
      icon: FlyIoIcon,
      category: "Deployment"
    },
    {
      name: "Heroku",
      icon: HerokuIcon,
      category: "Deployment"
    },
  ],
};

export function IntegrationsSection() {
  return (
    <section className="py-20 border-t border-white/8 gradient-shift-bg integrations-section-enhanced">
      <div className="max-w-7xl mx-auto px-4 lg:px-6">
        <div className="text-center mb-16">
          <h2 className="text-3xl font-bold text-text-primary mb-4" style={{ color: 'var(--text-primary)', fontWeight: 800 }}>
            Works with your favorite platforms
          </h2>
          <p className="text-text-secondary max-w-2xl mx-auto" style={{ color: 'var(--text-secondary)' }}>
            FunctionFly integrates seamlessly with the tools and platforms you already use.
            Deploy once, run everywhere across multiple providers.
          </p>
        </div>

        {/* Cloud Providers */}
        <div className="mb-16">
          <div className="text-center mb-8">
            <h3 className="text-2xl font-semibold text-text-primary mb-2" style={{ color: 'var(--text-primary)', fontWeight: 700 }}>
              Cloud Providers
            </h3>
            <p className="text-text-secondary" style={{ color: 'var(--text-secondary)' }}>
              Multi-cloud deployment with automatic failover across all major providers
            </p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-6">
            {integrationPartners.cloudProviders.map((provider, index) => {
              const isReactIcon = typeof provider.icon === 'function';
              const isIconifyIcon = typeof provider.icon === 'string';
              const isSimpleIcon = typeof provider.icon === 'object' && provider.icon !== null && 'hex' in provider.icon;

              return (
                <motion.div
                  key={provider.name}
                  initial={{ opacity: 0, scale: 0.9 }}
                  whileInView={{ opacity: 1, scale: 1 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.5, delay: index * 0.05 }}
                  className="group"
                >
                  <Card className="h-full hover:border-[#6366f1]/30 transition-all duration-300 hover:shadow-lg hover:shadow-[#6366f1]/10 glass-card card-elevation">
                    <CardContent className="p-6 text-center">
                      <div className={`w-12 h-12 mx-auto mb-3 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform duration-300 shadow-sm ${
                        provider.name === "AWS Lambda"
                          ? "bg-orange-500/20 border border-orange-500/30"
                          : provider.name === "Google Cloud Functions"
                          ? "bg-blue-500/20 border border-blue-500/30"
                          : provider.name === "Cloudflare Workers"
                          ? "bg-orange-500/20 border border-orange-500/30"
                          :                         provider.name === "Netlify Functions"
                          ? "bg-green-500/20 border border-green-500/30"
                          : provider.name === "Vercel Functions"
                          ? "bg-black/10 border border-black/20"
                          : "bg-gray-500/20 border border-gray-500/30"
                      }`}>
                        {isReactIcon ? (
                          React.createElement(provider.icon as React.ComponentType<any>, { className: "w-8 h-8 text-orange-500" })
                        ) : isIconifyIcon ? (
                          <Icon
                            icon={provider.icon as string}
                            className={`w-8 h-8 ${
                              provider.name === "Google Cloud Functions"
                                ? "text-blue-400"
                                : provider.name === "Cloudflare Workers"
                                ? "text-orange-400"
                                : provider.name === "Netlify Functions"
                                ? "text-green-400"
                                : provider.name === "Vercel Functions"
                                ? "text-black"
                                : "text-gray-400"
                            }`}
                          />
                        ) : isSimpleIcon ? (
                          <div
                            style={{
                              backgroundColor: `#${(provider.icon as any).hex}`,
                              color: 'white',
                              fontWeight: 'bold',
                              fontSize: '14px',
                              width: '100%',
                              height: '100%',
                              borderRadius: '6px',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center'
                            }}
                          >
                            {(provider.icon as any).title.charAt(0)}
                          </div>
                        ) : null}
                      </div>
                      <h4 className="text-sm font-semibold text-text-primary mb-1" style={{ color: 'var(--text-primary)' }}>
                        {provider.name}
                      </h4>
                      <p className="text-xs text-text-secondary">
                        {provider.category}
                      </p>
                    </CardContent>
                  </Card>
                </motion.div>
              );
            })}
          </div>
        </div>

        {/* Frameworks */}
        <div className="mb-16">
          <div className="text-center mb-8">
            <h3 className="text-2xl font-semibold text-text-primary mb-2" style={{ color: 'var(--text-primary)', fontWeight: 700 }}>
              Framework Integrations
            </h3>
            <p className="text-text-secondary" style={{ color: 'var(--text-secondary)' }}>
              Native support for popular frameworks and runtimes
            </p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-8 gap-4">
            {integrationPartners.frameworks.map((framework, index) => {
              const isReactIcon = typeof framework.icon === 'function';
              const isIconifyIcon = typeof framework.icon === 'string';
              const isSimpleIcon = typeof framework.icon === 'object' && framework.icon !== null && 'hex' in framework.icon;

              return (
                <motion.div
                  key={framework.name}
                  initial={{ opacity: 0, y: 20 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.5, delay: index * 0.05 }}
                  className="group"
                >
                  <Card className="h-full hover:border-[#6366f1]/30 transition-all duration-300 hover:shadow-lg hover:shadow-[#6366f1]/10 glass-card card-elevation">
                    <CardContent className="p-4 text-center">
                      <div className={`w-10 h-10 mx-auto mb-2 rounded-md flex items-center justify-center group-hover:scale-110 transition-transform duration-300 shadow-sm ${
                        framework.name === "React"
                          ? "bg-cyan-500/20 border border-cyan-500/30"
                          : framework.name === "Next.js"
                          ? "bg-white/10 border border-white/20"
                          : framework.name === "Vue.js"
                          ? "bg-green-500/20 border border-green-500/30"
                          : framework.name === "Django"
                          ? "bg-gray-500/20 border border-gray-500/30"
                          : framework.name === "Svelte"
                          ? "bg-orange-500/20 border border-orange-500/30"
                          : framework.name === "Express.js"
                          ? "bg-blue-500/20 border border-blue-500/30"
                          : framework.name === "Fastify"
                          ? "bg-pink-500/20 border border-pink-500/30"
                          : framework.name === "NestJS"
                          ? "bg-red-500/20 border border-red-500/30"
                          : "bg-gray-500/20 border border-gray-500/30"
                      }`}>
                        {isReactIcon ? (
                          React.createElement(framework.icon as React.ComponentType<any>, { className: "w-8 h-8 text-gray-500" })
                        ) : isIconifyIcon ? (
                          <Icon
                            icon={framework.icon as string}
                            className={`w-8 h-8 ${
                              framework.name === "React"
                                ? "text-cyan-400"
                                : framework.name === "Next.js"
                                ? "text-gray-100"
                                : framework.name === "Vue.js"
                                ? "text-green-400"
                                : framework.name === "Django"
                                ? "text-green-200"
                                : framework.name === "Svelte"
                                ? "text-orange-400"
                                : framework.name === "Express.js"
                                ? "text-blue-400"
                                : framework.name === "Fastify"
                                ? "text-pink-400"
                                : framework.name === "NestJS"
                                ? "text-red-400"
                                : "text-gray-400"
                            }`}
                          />
                        ) : isSimpleIcon ? (
                          <div
                            style={{
                              backgroundColor: `#${(framework.icon as any).hex}`,
                              color: 'white',
                              fontWeight: 'bold',
                              fontSize: '12px',
                              width: '100%',
                              height: '100%',
                              borderRadius: '4px',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center'
                            }}
                          >
                            {(framework.icon as any).title.charAt(0)}
                          </div>
                        ) : null}
                      </div>
                      <h4 className="text-xs font-semibold text-text-primary mb-1" style={{ color: 'var(--text-primary)' }}>
                        {framework.name}
                      </h4>
                      <p className="text-xs text-text-muted">
                        {framework.category}
                      </p>
                    </CardContent>
                  </Card>
                </motion.div>
              );
            })}
          </div>
        </div>

        {/* Deployment Platforms */}
        <div>
          <div className="text-center mb-8">
            <h3 className="text-2xl font-semibold text-text-primary mb-2" style={{ color: 'var(--text-primary)', fontWeight: 700 }}>
              Deployment Platforms
            </h3>
            <p className="text-text-secondary" style={{ color: 'var(--text-secondary)' }}>
              One-click deployment to your preferred hosting platform
            </p>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-6">
            {integrationPartners.deploymentPlatforms.map((platform, index) => {
              const isReactIcon = typeof platform.icon === 'function';
              const isIconifyIcon = typeof platform.icon === 'string';
              const isSimpleIcon = typeof platform.icon === 'object' && platform.icon !== null && 'hex' in platform.icon;

              return (
                <motion.div
                  key={platform.name}
                  initial={{ opacity: 0, scale: 0.9 }}
                  whileInView={{ opacity: 1, scale: 1 }}
                  viewport={{ once: true }}
                  transition={{ duration: 0.5, delay: index * 0.05 }}
                  className="group"
                >
                  <Card className="h-full hover:border-[#6366f1]/30 transition-all duration-300 hover:shadow-lg hover:shadow-[#6366f1]/10 glass-card card-elevation">
                    <CardContent className="p-6 text-center">
                      <div className={`w-12 h-12 mx-auto mb-3 rounded-lg flex items-center justify-center group-hover:scale-110 transition-transform duration-300 shadow-sm ${
                        platform.name === "Netlify"
                          ? "bg-green-500/20 border border-green-500/30"
                          : platform.name === "Railway"
                          ? "bg-purple-500/20 border border-purple-500/30"
                          : platform.name === "Fly.io"
                          ? "bg-purple-600/20 border border-purple-600/30"
                          : platform.name === "Heroku"
                          ? "bg-purple-700/20 border border-purple-700/30"
                          : platform.name === "Vercel"
                          ? "bg-black/10 border border-black/20"
                          : "bg-gray-500/20 border border-gray-500/30"
                      }`}>
                        {isReactIcon ? (
                          React.createElement(platform.icon as React.ComponentType<any>, { className: "w-8 h-8 text-gray-500" })
                        ) : isIconifyIcon ? (
                          <Icon
                            icon={platform.icon as string}
                            className={`w-8 h-8 ${
                              platform.name === "Netlify"
                                ? "text-green-400"
                                : platform.name === "Railway"
                                ? "text-purple-400"
                                : platform.name === "Fly.io"
                                ? "text-purple-500"
                                : platform.name === "Heroku"
                                ? "text-purple-600"
                                : platform.name === "Vercel"
                                ? "text-black"
                                : "text-gray-400"
                            }`}
                          />
                        ) : isSimpleIcon ? (
                          <div
                            style={{
                              backgroundColor: `#${(platform.icon as any).hex}`,
                              color: 'white',
                              fontWeight: 'bold',
                              fontSize: '14px',
                              width: '100%',
                              height: '100%',
                              borderRadius: '6px',
                              display: 'flex',
                              alignItems: 'center',
                              justifyContent: 'center'
                            }}
                          >
                            {(platform.icon as any).title.charAt(0)}
                          </div>
                        ) : null}
                      </div>
                      <h4 className="text-sm font-semibold text-text-primary mb-1" style={{ color: 'var(--text-primary)' }}>
                        {platform.name}
                      </h4>
                      <p className="text-xs text-text-secondary">
                        {platform.category}
                      </p>
                    </CardContent>
                  </Card>
                </motion.div>
              );
            })}
          </div>
        </div>

        {/* Integration Promise */}
        <div className="text-center mt-16">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5 }}
          >
            <Card className="border-[#6366f1]/30 bg-[#6366f1]/5 max-w-2xl mx-auto glass-card card-elevation">
              <CardContent className="p-8">
                <div className="w-16 h-16 mx-auto mb-6 rounded-2xl bg-linear-to-br from-[#6366f1]/20 to-[#8b5cf6]/20 border border-[#6366f1]/20 flex items-center justify-center">
                  <Globe className="w-8 h-8 text-[#6366f1]" />
                </div>
                <h3 className="text-2xl font-bold text-text-primary mb-4" style={{ color: 'var(--text-primary)' }}>
                  One Platform, Infinite Possibilities
                </h3>
                <p className="text-text-secondary text-lg">
                  Whether you're using AWS Lambda, Vercel Functions, or deploying to multiple clouds simultaneously,
                  FunctionFly adapts to your infrastructure choices while providing unmatched reliability.
                </p>
              </CardContent>
            </Card>
          </motion.div>
        </div>
      </div>
    </section>
  );
}