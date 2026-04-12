/**
 * EmptyStateOverlay Component
 * Shows when a new graph is created with empty canvas
 * Provides templates and AI prompt to get started
 */

import { useState } from 'react';
import { motion } from 'framer-motion';
import {
  Sparkles,
  Webhook,
  Brain,
  Database,
  Clock,
  ArrowRight,
  ArrowRightToLine,
  ArrowLeftFromLine,
  MessageSquare,
  Zap,
  Play,
  Timer,
  Bell,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Input as UIInput } from '@/components/ui/input';
import { Card, CardContent } from '@/components/ui/card';
import { cn } from '@/lib/utils';

interface Template {
  id: string;
  name: string;
  description: string;
  icon: React.ReactNode;
  nodeCount: number;
  nodes: Array<{
    type: 'input' | 'function' | 'ai' | 'transform' | 'output' | 'timer' | 'notify';
    label: string;
    color: string;
  }>;
}

const templates: Template[] = [
  {
    id: 'webhook-api',
    name: 'Webhook → Function → Response',
    description: 'Create an API endpoint with request handling and response',
    icon: <Webhook className="w-5 h-5" />,
    nodeCount: 3,
    nodes: [
      { type: 'input', label: 'HTTP Request', color: 'bg-blue-500' },
      { type: 'function', label: 'Process', color: 'bg-purple-500' },
      { type: 'output', label: 'Response', color: 'bg-green-500' },
    ],
  },
  {
    id: 'ai-pipeline',
    name: 'AI Pipeline',
    description: 'Input → GPT-4 → Output processing chain',
    icon: <Brain className="w-5 h-5" />,
    nodeCount: 3,
    nodes: [
      { type: 'input', label: 'User Input', color: 'bg-blue-500' },
      { type: 'ai', label: 'GPT-4', color: 'bg-pink-500' },
      { type: 'output', label: 'Response', color: 'bg-green-500' },
    ],
  },
  {
    id: 'data-processing',
    name: 'Data Processing',
    description: 'Upload → Transform → Store workflow',
    icon: <Database className="w-5 h-5" />,
    nodeCount: 3,
    nodes: [
      { type: 'input', label: 'File Upload', color: 'bg-blue-500' },
      { type: 'transform', label: 'Transform', color: 'bg-orange-500' },
      { type: 'output', label: 'Database', color: 'bg-green-500' },
    ],
  },
  {
    id: 'scheduled-job',
    name: 'Scheduled Job',
    description: 'Timer → Function → Notification workflow',
    icon: <Clock className="w-5 h-5" />,
    nodeCount: 3,
    nodes: [
      { type: 'timer', label: 'Schedule', color: 'bg-yellow-500' },
      { type: 'function', label: 'Execute', color: 'bg-purple-500' },
      { type: 'notify', label: 'Notify', color: 'bg-teal-500' },
    ],
  },
];

function NodeIcon({ type }: { type: string }) {
  switch (type) {
    case 'input':
      return <ArrowRightToLine className="w-3 h-3" />;
    case 'output':
      return <ArrowLeftFromLine className="w-3 h-3" />;
    case 'function':
    case 'execute':
      return <Play className="w-3 h-3" />;
    case 'ai':
      return <Brain className="w-3 h-3" />;
    case 'transform':
      return <Zap className="w-3 h-3" />;
    case 'timer':
      return <Timer className="w-3 h-3" />;
    case 'notify':
      return <Bell className="w-3 h-3" />;
    default:
      return <Zap className="w-3 h-3" />;
  }
}

function TemplatePreview({ nodes }: { nodes: Template['nodes'] }) {
  return (
    <div className="flex items-center justify-center gap-1 py-2">
      {nodes.map((node, index) => (
        <div key={index} className="flex items-center">
          <motion.div
            initial={{ opacity: 0, scale: 0.8 }}
            animate={{ opacity: 1, scale: 1 }}
            transition={{ delay: index * 0.1 }}
            className={cn(
              "w-8 h-8 rounded-lg flex items-center justify-center text-white text-[10px] font-medium",
              node.color
            )}
          >
            <NodeIcon type={node.type} />
          </motion.div>
          {index < nodes.length - 1 && (
            <motion.div
              initial={{ width: 0 }}
              animate={{ width: 16 }}
              transition={{ delay: index * 0.1 + 0.05 }}
              className="h-0.5 bg-[var(--border-default)] mx-1"
            />
          )}
        </div>
      ))}
    </div>
  );
}

function TemplateCard({
  template,
  onSelect,
  index,
}: {
  template: Template;
  onSelect: (templateId: string) => void;
  index: number;
}) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.1 }}
      whileHover={{ scale: 1.02, y: -4 }}
      whileTap={{ scale: 0.98 }}
      onClick={() => onSelect(template.id)}
    >
      <Card className="cursor-pointer border-[var(--border-subtle)] bg-[var(--bg-secondary)] hover:border-[var(--border-focus)] hover:bg-[var(--bg-tertiary)] transition-all duration-200 overflow-hidden group">
        <CardContent className="p-4">
          {/* Header */}
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-brand-500/20 to-purple-500/20 flex items-center justify-center text-brand-500 group-hover:from-brand-500 group-hover:to-purple-500 group-hover:text-white transition-all duration-200">
              {template.icon}
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="text-sm font-semibold text-[var(--text-primary)] group-hover:text-brand-400 transition-colors">
                {template.name}
              </h3>
              <p className="text-xs text-[var(--text-secondary)] mt-0.5">
                {template.description}
              </p>
            </div>
            <ArrowRight className="w-4 h-4 text-[var(--text-muted)] opacity-0 group-hover:opacity-100 group-hover:translate-x-1 transition-all" />
          </div>

          {/* Visual Preview */}
          <div className="mt-3 pt-3 border-t border-[var(--border-subtle)]">
            <TemplatePreview nodes={template.nodes} />
          </div>

          {/* Footer */}
          <div className="flex items-center justify-between mt-3">
            <span className="text-xs text-[var(--text-muted)]">
              {template.nodeCount} nodes
            </span>
            <Button variant="ghost" size="sm" className="h-7 text-xs opacity-0 group-hover:opacity-100 transition-opacity">
              Use Template
            </Button>
          </div>
        </CardContent>
      </Card>
    </motion.div>
  );
}

interface EmptyStateOverlayProps {
  onTemplateSelect: (templateId: string) => void;
  onAIPrompt: (prompt: string) => void;
}

export function EmptyStateOverlay({ onTemplateSelect, onAIPrompt }: EmptyStateOverlayProps) {
  const [aiPrompt, setAiPrompt] = useState('');

  const handleAISubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (aiPrompt.trim()) {
      onAIPrompt(aiPrompt);
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="absolute inset-0 z-20 flex items-center justify-center p-8 bg-[var(--bg-primary)]/80 backdrop-blur-sm"
    >
      <div className="max-w-4xl w-full space-y-8">
        {/* Welcome Header */}
        <motion.div
          initial={{ opacity: 0, y: -20 }}
          animate={{ opacity: 1, y: 0 }}
          className="text-center"
        >
          <div className="inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-gradient-to-br from-brand-500/20 to-purple-500/20 mb-4">
            <Zap className="w-8 h-8 text-brand-500" />
          </div>
          <h2 className="text-2xl font-bold text-[var(--text-primary)] mb-2">
            Build your first workflow
          </h2>
          <p className="text-sm text-[var(--text-secondary)] max-w-md mx-auto">
            Choose a template to get started quickly, or describe what you want to build and let AI create it for you
          </p>
        </motion.div>

        {/* Template Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {templates.map((template, index) => (
            <TemplateCard
              key={template.id}
              template={template}
              onSelect={onTemplateSelect}
              index={index}
            />
          ))}
        </div>

        {/* AI Prompt Section */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
          className="relative"
        >
          <div className="absolute inset-0 bg-gradient-to-r from-brand-500/10 via-purple-500/10 to-pink-500/10 rounded-xl blur-xl" />
          <Card className="relative border-[var(--border-subtle)] bg-[var(--bg-secondary)]/80 backdrop-blur-sm">
            <CardContent className="p-4">
              <form onSubmit={handleAISubmit} className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-lg bg-gradient-to-br from-pink-500 to-purple-500 flex items-center justify-center shrink-0">
                  <Sparkles className="w-5 h-5 text-white" />
                </div>
                <div className="flex-1">
                  <UIInput
                    placeholder="Describe what you want to build... (e.g., 'A webhook that processes payments and sends confirmation emails')"
                    value={aiPrompt}
                    onChange={(e) => setAiPrompt(e.target.value)}
                    className="border-[var(--border-subtle)] bg-[var(--bg-tertiary)] focus:border-brand-500 focus:ring-brand-500/20"
                  />
                </div>
                <Button
                  type="submit"
                  disabled={!aiPrompt.trim()}
                  className="bg-gradient-to-r from-brand-500 to-purple-500 disabled:opacity-50"
                >
                  <Sparkles className="w-4 h-4 mr-2" />
                  Generate
                </Button>
              </form>
              <div className="flex items-center gap-2 mt-3 px-2">
                <span className="text-xs text-[var(--text-muted)]">Examples:</span>
                {[
                  'Customer support bot',
                  'Data ETL pipeline',
                  'Image processing API',
                ].map((example) => (
                  <button
                    key={example}
                    onClick={() => setAiPrompt(`Create a ${example.toLowerCase()}`)}
                    className="text-xs text-[var(--text-secondary)] hover:text-brand-400 transition-colors"
                  >
                    {example}
                  </button>
                ))}
              </div>
            </CardContent>
          </Card>
        </motion.div>

        {/* Additional Options */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 0.5 }}
          className="flex items-center justify-center gap-4 text-sm text-[var(--text-muted)]"
        >
          <button className="hover:text-[var(--text-secondary)] transition-colors flex items-center gap-1">
            <MessageSquare className="w-4 h-4" />
            Start with blank canvas
          </button>
          <span>•</span>
          <button className="hover:text-[var(--text-secondary)] transition-colors flex items-center gap-1">
            <Database className="w-4 h-4" />
            Import from library
          </button>
        </motion.div>
      </div>
    </motion.div>
  );
}

export default EmptyStateOverlay;
