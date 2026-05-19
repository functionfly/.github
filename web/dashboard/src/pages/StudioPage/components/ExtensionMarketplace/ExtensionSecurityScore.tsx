import { GlassCard, Badge } from "@functionfly/ui-core";
import { Shield, AlertTriangle, CheckCircle, Lock, Eye, FileSearch } from "lucide-react";
import type { Extension } from "@/api/marketplace";

interface ExtensionSecurityScoreProps {
  extension: Extension;
  detailed?: boolean;
}

interface SecurityCheck {
  id: string;
  name: string;
  description: string;
  status: "pass" | "warning" | "fail";
  score: number;
}

export function ExtensionSecurityScore({ extension, detailed = false }: ExtensionSecurityScoreProps) {
  const checks: SecurityCheck[] = [
    {
      id: "signature",
      name: "Code Signature",
      description: "Extension bundle is cryptographically signed",
      status: extension.signature ? "pass" : "warning",
      score: extension.signature ? 100 : 50,
    },
    {
      id: "verified",
      name: "Verified Publisher",
      description: "Publisher identity has been verified",
      status: extension.verified ? "pass" : "fail",
      score: extension.verified ? 100 : 0,
    },
    {
      id: "sandbox",
      name: "Sandbox Isolation",
      description: "Extension runs in isolated sandbox environment",
      status: extension.sandbox_score >= 80 ? "pass" : extension.sandbox_score >= 50 ? "warning" : "fail",
      score: extension.sandbox_score,
    },
    {
      id: "trust",
      name: "Trust Score",
      description: "Overall trust rating based on community feedback",
      status: extension.trust_score >= 80 ? "pass" : extension.trust_score >= 50 ? "warning" : "fail",
      score: extension.trust_score,
    },
    {
      id: "security",
      name: "Security Audit",
      description: "Code has passed security audit checks",
      status: extension.security_score >= 80 ? "pass" : extension.security_score >= 50 ? "warning" : "fail",
      score: extension.security_score,
    },
    {
      id: "runtime",
      name: "Runtime Safety",
      description: "Runtime environment meets safety standards",
      status: extension.runtime_score >= 80 ? "pass" : extension.runtime_score >= 50 ? "warning" : "fail",
      score: extension.runtime_score,
    },
  ];

  const passCount = checks.filter((c) => c.status === "pass").length;
  const warningCount = checks.filter((c) => c.status === "warning").length;
  const failCount = checks.filter((c) => c.status === "fail").length;
  const overallScore = Math.round(checks.reduce((acc, c) => acc + c.score, 0) / checks.length);

  const overallStatus = overallScore >= 80 ? "pass" : overallScore >= 50 ? "warning" : "fail";
  const statusColor = overallStatus === "pass" ? "text-green-400" : overallStatus === "warning" ? "text-yellow-400" : "text-red-400";
  const statusBg = overallStatus === "pass" ? "bg-green-500/20" : overallStatus === "warning" ? "bg-yellow-500/20" : "bg-red-500/20";

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <Shield className={`w-5 h-5 ${statusColor}`} />
          <h3 className="text-sm font-semibold text-white">Security Assessment</h3>
        </div>
        <Badge variant={overallStatus === "pass" ? "success" : overallStatus === "warning" ? "warning" : "destructive"}>
          {overallScore}%
        </Badge>
      </div>

      <GlassCard className="p-4">
        <div className={`p-4 rounded-lg ${statusBg} flex items-center gap-4`}>
          <div className={`w-12 h-12 rounded-full flex items-center justify-center ${
            overallStatus === "pass" ? "bg-green-500/30" : overallStatus === "warning" ? "bg-yellow-500/30" : "bg-red-500/30"
          }`}>
            <Shield className={`w-6 h-6 ${statusColor}`} />
          </div>
          <div className="flex-1">
            <div className="text-lg font-semibold text-white">
              {overallStatus === "pass" ? "Secure" : overallStatus === "warning" ? "Review Recommended" : "Potential Risk"}
            </div>
            <div className="text-sm text-white/60">
              {overallStatus === "pass"
                ? "This extension has passed all security checks"
                : overallStatus === "warning"
                ? "Some security aspects could be improved"
                : "This extension may pose security risks"}
            </div>
          </div>
          <div className="text-3xl font-bold text-white">{overallScore}%</div>
        </div>
      </GlassCard>

      <div className="grid grid-cols-3 gap-2">
        <div className="p-2 rounded-lg bg-green-500/10 text-center">
          <CheckCircle className="w-4 h-4 text-green-400 mx-auto mb-1" />
          <div className="text-lg font-semibold text-green-400">{passCount}</div>
          <div className="text-[10px] text-white/60">Passed</div>
        </div>
        <div className="p-2 rounded-lg bg-yellow-500/10 text-center">
          <AlertTriangle className="w-4 h-4 text-yellow-400 mx-auto mb-1" />
          <div className="text-lg font-semibold text-yellow-400">{warningCount}</div>
          <div className="text-[10px] text-white/60">Warnings</div>
        </div>
        <div className="p-2 rounded-lg bg-red-500/10 text-center">
          <AlertTriangle className="w-4 h-4 text-red-400 mx-auto mb-1" />
          <div className="text-lg font-semibold text-red-400">{failCount}</div>
          <div className="text-[10px] text-white/60">Failed</div>
        </div>
      </div>

      {detailed && (
        <div className="space-y-2">
          {checks.map((check) => (
            <div
              key={check.id}
              className="flex items-center gap-3 p-3 rounded-lg bg-white/5 border border-white/10"
            >
              <div className={`w-8 h-8 rounded-full flex items-center justify-center ${
                check.status === "pass" ? "bg-green-500/20 text-green-400" :
                check.status === "warning" ? "bg-yellow-500/20 text-yellow-400" :
                "bg-red-500/20 text-red-400"
              }`}>
                {check.status === "pass" ? (
                  <CheckCircle className="w-4 h-4" />
                ) : (
                  <AlertTriangle className="w-4 h-4" />
                )}
              </div>
              <div className="flex-1">
                <div className="text-sm font-medium text-white">{check.name}</div>
                <div className="text-xs text-white/60">{check.description}</div>
              </div>
              <div className="flex items-center gap-2">
                <div className="w-16 h-2 bg-white/10 rounded-full overflow-hidden">
                  <div
                    className={`h-full rounded-full ${
                      check.status === "pass" ? "bg-green-400" :
                      check.status === "warning" ? "bg-yellow-400" :
                      "bg-red-400"
                    }`}
                    style={{ width: `${check.score}%` }}
                  />
                </div>
                <span className="text-sm text-white/60 w-10 text-right">{check.score}%</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}