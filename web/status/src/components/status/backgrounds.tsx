import { motion } from "framer-motion";
import { cn } from "@/lib/utils";
import type { StatusOrbitalProps } from "./types";

export function AnimatedBackground() {
  return (
    <div className="fixed inset-0 pointer-events-none overflow-hidden -z-10">
      <motion.div
        className="absolute w-[600px] h-[600px] rounded-full opacity-20 blur-[100px]"
        style={{
          background:
            "radial-gradient(circle, rgba(99, 102, 241, 0.3) 0%, transparent 70%)",
          top: "-10%",
          left: "-10%",
        }}
        animate={{
          x: [0, 100, 0],
          y: [0, 50, 0],
          scale: [1, 1.2, 1],
        }}
        transition={{
          duration: 20,
          repeat: Infinity,
          ease: "easeInOut",
        }}
      />
      <motion.div
        className="absolute w-[500px] h-[500px] rounded-full opacity-15 blur-[80px]"
        style={{
          background:
            "radial-gradient(circle, rgba(139, 92, 246, 0.3) 0%, transparent 70%)",
          bottom: "-5%",
          right: "-5%",
        }}
        animate={{
          x: [0, -80, 0],
          y: [0, -30, 0],
          scale: [1, 1.1, 1],
        }}
        transition={{
          duration: 15,
          repeat: Infinity,
          ease: "easeInOut",
          delay: 2,
        }}
      />
      <div
        className="absolute inset-0 opacity-[0.02]"
        style={{
          backgroundImage: `
            linear-gradient(rgba(255,255,255,0.1) 1px, transparent 1px),
            linear-gradient(90deg, rgba(255,255,255,0.1) 1px, transparent 1px)
          `,
          backgroundSize: "50px 50px",
        }}
      />
    </div>
  );
}

export function StatusOrbital({ status, size = "lg" }: StatusOrbitalProps) {
  const colors = {
    operational: "bg-emerald-500 text-emerald-500",
    degraded: "bg-amber-500 text-amber-500",
    maintenance: "bg-purple-500 text-purple-500",
    major_outage: "bg-red-500 text-red-500",
    partial_outage: "bg-orange-500 text-orange-500",
  };

  const sizeClasses = {
    sm: "w-3 h-3",
    md: "w-4 h-4",
    lg: "w-6 h-6",
    xl: "w-8 h-8",
  };

  return (
    <div className="relative flex items-center justify-center">
      {status === "operational" && (
        <>
          <motion.div
            className={cn(
              "absolute rounded-full border-2 border-current opacity-30",
              colors[status as keyof typeof colors],
              sizeClasses[size],
            )}
            animate={{ scale: [1, 1.8, 2.2], opacity: [0.5, 0.2, 0] }}
            transition={{ duration: 2, repeat: Infinity, ease: "easeOut" }}
          />
          <motion.div
            className={cn(
              "absolute rounded-full border border-current opacity-20",
              colors[status as keyof typeof colors],
              sizeClasses[size],
            )}
            animate={{ scale: [1, 2, 2.5], opacity: [0.3, 0.1, 0] }}
            transition={{
              duration: 2,
              repeat: Infinity,
              ease: "easeOut",
              delay: 0.5,
            }}
          />
        </>
      )}
      <motion.div
        className={cn(
          "rounded-full relative z-10",
          colors[status as keyof typeof colors].split(" ")[0],
          sizeClasses[size],
        )}
        animate={
          status === "operational"
            ? {
                boxShadow: [
                  "0 0 10px currentColor",
                  "0 0 20px currentColor",
                  "0 0 10px currentColor",
                ],
              }
            : {}
        }
        transition={{ duration: 2, repeat: Infinity }}
      />
    </div>
  );
}
