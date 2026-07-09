import { BunIcon } from '@/components/icons/BunIcon';
import { DenoIcon } from '@/components/icons/DenoIcon';
import { GoIcon } from '@/components/icons/GoIcon';
import { NodeIcon } from '@/components/icons/NodeIcon';
import { PythonIcon } from '@/components/icons/PythonIcon';
import { RustIcon } from '@/components/icons/RustIcon';
import type { LucideIcon } from 'lucide-react';
import {
  Activity,
  ArrowLeftRight,
  Coins,
  Cpu,
  ExternalLink,
  FileCode2,
  Shield,
  Zap,
} from 'lucide-react';
import type { FC } from 'react';

type IconComponent = LucideIcon | FC<{ className?: string }>;

export const RUNTIMES: { value: string; label: string; icon: IconComponent }[] = [
  { value: 'python', label: 'Python 3.11', icon: PythonIcon },
  { value: 'python-light', label: 'Python (light)', icon: PythonIcon },
  { value: 'nodejs', label: 'Node.js 20', icon: NodeIcon },
  { value: 'go', label: 'Go 1.21', icon: GoIcon },
  { value: 'rust', label: 'Rust', icon: RustIcon },
  { value: 'deno', label: 'Deno', icon: DenoIcon },
  { value: 'bun', label: 'Bun', icon: BunIcon },
];

export const COMPLEXITY_COLORS = {
  simple: {
    bg: 'bg-green-500',
    text: 'text-green-700 dark:text-green-300',
    border: 'border-green-500/30',
    label: 'Simple',
  },
  moderate: {
    bg: 'bg-yellow-500',
    text: 'text-yellow-700 dark:text-yellow-300',
    border: 'border-yellow-500/30',
    label: 'Moderate',
  },
  complex: {
    bg: 'bg-red-500',
    text: 'text-red-700 dark:text-red-300',
    border: 'border-red-500/30',
    label: 'Complex',
  },
};

export const CAPABILITY_INFO: Record<string, { description: string; icon: IconComponent }> = {
  http: {
    description: 'Allows the function to make outbound HTTP requests',
    icon: ExternalLink,
  },
  filesystem: {
    description: 'Provides temporary filesystem access for file processing',
    icon: FileCode2,
  },
  crypto: {
    description: 'Enables cryptographic operations and hashing',
    icon: Shield,
  },
  database: {
    description: 'Allows database connections and queries',
    icon: Coins,
  },
  streaming: {
    description: 'Supports streaming input/output for large data',
    icon: Activity,
  },
  gpu: {
    description: 'Provides GPU acceleration for compute-intensive tasks',
    icon: Cpu,
  },
  cache: {
    description: 'Enables access to distributed caching layer',
    icon: Zap,
  },
  queue: {
    description: 'Allows message queue operations',
    icon: ArrowLeftRight,
  },
};

export const TOKEN_COST_USD = {
  prompt: 0.0015,
  completion: 0.002,
};

export const QUICK_REFINEMENTS = [
  {
    label: 'Add error handling',
    prompt: 'Add comprehensive error handling with try-catch blocks and proper error messages',
  },
  { label: 'Add pagination', prompt: 'Make it handle pagination for large datasets' },
  {
    label: 'Add input validation',
    prompt: 'Add input validation with clear error messages for invalid inputs',
  },
  { label: 'Add logging', prompt: 'Add logging statements for debugging and monitoring' },
  {
    label: 'Optimize performance',
    prompt: 'Optimize for better performance and reduce memory usage',
  },
  { label: 'Add comments', prompt: 'Add detailed inline comments explaining the code' },
];

export const DRAFT_KEY = 'ai-composer-draft';
