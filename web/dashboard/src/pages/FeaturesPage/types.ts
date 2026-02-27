import { ComponentType } from "react";

// Feature type definition
export interface Feature {
  icon: ComponentType<{ className?: string }>;
  title: string;
  description: string;
  category: string;
  highlights: string[];
  color: string;
  bgGradient: string;
  borderColor: string;
  illustration?: ComponentType;
  cta?: {
    text: string;
    action: string;
    link?: string;
  };
  detailedContent?: {
    overview: string;
    technicalDetails: string[];
    benefits: string[];
    implementation: string;
  };
}

// Category color schemes
export const categoryColors: Record<string, {
  color: string;
  bgGradient: string;
  borderColor: string;
}> = {
  Reliability: {
    color: "#f59e0b", // amber
    bgGradient: "from-[#f59e0b]/20 to-[#d97706]/20",
    borderColor: "#f59e0b",
  },
  Deployment: {
    color: "#8b5cf6", // violet
    bgGradient: "from-[#8b5cf6]/20 to-[#7c3aed]/20",
    borderColor: "#8b5cf6",
  },
  Intelligence: {
    color: "#10b981", // emerald
    bgGradient: "from-[#10b981]/20 to-[#059669]/20",
    borderColor: "#10b981",
  },
  "Developer Tools": {
    color: "#3b82f6", // blue
    bgGradient: "from-[#3b82f6]/20 to-[#2563eb]/20",
    borderColor: "#3b82f6",
  },
  Collaboration: {
    color: "#ec4899", // pink
    bgGradient: "from-[#ec4899]/20 to-[#db2777]/20",
    borderColor: "#ec4899",
  },
  Monitoring: {
    color: "#06b6d4", // cyan
    bgGradient: "from-[#06b6d4]/20 to-[#0891b2]/20",
    borderColor: "#06b6d4",
  },
  Configuration: {
    color: "#84cc16", // lime
    bgGradient: "from-[#84cc16]/20 to-[#65a30d]/20",
    borderColor: "#84cc16",
  },
  Infrastructure: {
    color: "#f97316", // orange
    bgGradient: "from-[#f97316]/20 to-[#ea580c]/20",
    borderColor: "#f97316",
  },
};