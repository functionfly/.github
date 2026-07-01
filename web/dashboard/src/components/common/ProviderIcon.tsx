import { Terminal, Zap } from "lucide-react";
import { CloudflareIcon, VercelIcon, FlyIoIcon, FunctionFlyEdgeIcon, AwsIcon } from "@/pages/LandingPage/components/icons";
import { OpenAIIcon, AnthropicIcon, MiMoIcon, StepFunIcon, TogetherIcon, FireworksIcon, GroqIcon, OpenRouterIcon, DeepInfraIcon, MiniMaxIcon } from "@/components/icons/AiProviderIcons";
import { cn } from "@/lib/utils";

interface ProviderIconProps {
  provider: "workers" | "vercel" | "fly" | "deno" | "functionfly-edge" | "aws-lambda" | string;
  size?: "sm" | "md" | "lg" | "xl";
  className?: string;
}

const sizeConfig = {
  sm: "w-4 h-4",
  md: "w-5 h-5",
  lg: "w-6 h-6",
  xl: "w-10 h-10",
};

const providerConfig = {
  workers: {
    Icon: CloudflareIcon,
    color: "#f48120",
    label: "Cloudflare Workers",
  },
  vercel: {
    Icon: VercelIcon,
    color: "#000000",
    label: "Vercel",
  },
  fly: {
    Icon: FlyIoIcon,
    color: "#7b68ee",
    label: "Fly.io",
  },
  deno: {
    Icon: Terminal,
    color: "#ffffff",
    label: "Deno Deploy",
  },
  "functionfly-edge": {
    Icon: FunctionFlyEdgeIcon,
    color: "#6366f1",
    label: "FunctionFly Edge",
  },
  "aws-lambda": {
    Icon: AwsIcon,
    color: "#FF9900",
    label: "AWS Lambda",
  },
  openai: {
    Icon: OpenAIIcon,
    color: "#10a37f",
    label: "OpenAI",
  },
  anthropic: {
    Icon: AnthropicIcon,
    color: "#d97757",
    label: "Anthropic",
  },
  mimo: {
    Icon: MiMoIcon,
    color: "#ff6900",
    label: "Xiaomi MiMo",
  },
  stepfun: {
    Icon: StepFunIcon,
    color: "#000000",
    label: "StepFun",
  },
  together: {
    Icon: TogetherIcon,
    color: "#6e25c0",
    label: "Together AI",
  },
  fireworks: {
    Icon: FireworksIcon,
    color: "#6720FF",
    label: "Fireworks AI",
  },
  groq: {
    Icon: GroqIcon,
    color: "#f55036",
    label: "Groq",
  },
  openrouter: {
    Icon: OpenRouterIcon,
    color: "#c084fc",
    label: "OpenRouter",
  },
  deepinfra: {
    Icon: DeepInfraIcon,
    color: "#0ea5e9",
    label: "DeepInfra",
  },
  minimax: {
    Icon: MiniMaxIcon,
    color: "#E2167E",
    label: "MiniMax",
  },
  "mimo-token-plan": {
    Icon: MiMoIcon,
    color: "#ff6900",
    label: "MiMo Token Plan",
  },
  "minimax-token-plan": {
    Icon: MiniMaxIcon,
    color: "#ED3465",
    label: "MiniMax Token Plan",
  },
};

export function ProviderIcon({ provider, size = "md", className }: ProviderIconProps) {
  const config = providerConfig[provider as keyof typeof providerConfig] || {
    Icon: Zap,
    color: "#a0a0b0",
    label: provider,
  };

  const Icon = config.Icon;

  return (
    <div
      className={cn(
        "flex items-center justify-center rounded-md shrink-0",
        sizeConfig[size],
        className
      )}
      style={{ color: config.color }}
      title={config.label}
    >
      <Icon className="w-full h-full text-inherit" />
    </div>
  );
}
