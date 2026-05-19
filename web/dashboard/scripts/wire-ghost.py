with open('src/pages/StudioPage.tsx', 'r') as f:
    content = f.read()

# Add back the useGhostMode import
content = content.replace(
    'import {\n  ExtensionManager,\n  ExtensionDetailPanel,\n  HookSystemVisualizer,\n  SandboxMonitor,\n  ExtensionSDKDebugger,\n} from "@functionfly/ui-extensibility";',
    'import {\n  ExtensionManager,\n  ExtensionDetailPanel,\n  HookSystemVisualizer,\n  SandboxMonitor,\n  ExtensionSDKDebugger,\n} from "@functionfly/ui-extensibility";\nimport { useGhostMode } from "@/services/ghost-api";'
)

# Replace the stub GhostModeView with a proper one
old = '''// Ghost Mode View - Autonomous building (STUB)
function GhostModeView() {
  return (
    <div className="flex-1 overflow-y-auto p-4">
      <div className="text-text-muted">Ghost Mode - Backend API not connected</div>
    </div>
  );
}'''

new = '''// Ghost Mode View - Autonomous building (STUB)
function GhostModeView() {
  const { loading, error, createBuild, pauseBuild, resumeBuild, cancelBuild,
          submitApproval, pauseBuild: pause, resumeBuild: resume, getBuild,
          pollBuild } = useGhostMode();
  const [build, setBuild] = React.useState({
    id: "build-001",
    goal: "Build a REST API for user management with auth, CRUD operations, and rate limiting",
    description: "Auto-generating full-stack user management microservice",
    phase: "building",
    progress: 0.67,
    startedAt: new Date(Date.now() - 300000).toISOString(),
    updatedAt: new Date().toISOString(),
    estimatedCompletion: new Date(Date.now() + 180000).toISOString(),
    currentTaskId: "task-4",
    humanApprovalRequired: false,
    approvalType: "schema",
    tasks: [
      { id: "task-1", title: "Design database schema", description: "Create user table with roles and permissions", status: "completed", phase: "planning", startedAt: new Date(Date.now() - 280000).toISOString(), completedAt: new Date(Date.now() - 240000).toISOString(), duration: 40000, logs: [], agentId: "arch-agent" },
      { id: "task-2", title: "Provision infrastructure", description: "Create Docker containers and networking", status: "completed", phase: "provisioning", startedAt: new Date(Date.now() - 230000).toISOString(), completedAt: new Date(Date.now() - 180000).toISOString(), duration: 50000, logs: [], agentId: "infra-agent" },
      { id: "task-3", title: "Implement auth endpoints", description: "JWT-based authentication with refresh tokens", status: "completed", phase: "building", startedAt: new Date(Date.now() - 170000).toISOString(), completedAt: new Date(Date.now() - 100000).toISOString(), duration: 70000, logs: [], agentId: "backend-agent" },
      { id: "task-4", title: "Generate CRUD handlers", description: "User create/read/update/delete operations", status: "in_progress", phase: "building", startedAt: new Date(Date.now() - 90000).toISOString(), logs: [{ timestamp: new Date().toISOString(), level: "info", message: "Parsing OpenAPI spec..." }, { timestamp: new Date().toISOString(), level: "info", message: "Generating handler code..." }], agentId: "backend-agent" },
      { id: "task-5", title: "Add rate limiting middleware", description: "Token bucket algorithm for API protection", status: "pending", phase: "building", dependencies: ["task-4"] },
      { id: "task-6", title: "Write unit tests", description: "80% code coverage requirement", status: "pending", phase: "building", dependencies: ["task-4"] },
      { id: "task-7", title: "Deploy to staging", description: "Blue-green deployment to staging environment", status: "pending", phase: "deploying", dependencies: ["task-5", "task-6"] },
    ],
  });

  // Poll for build updates when active
  React.useEffect(() => {
    if (build && build.id !== "build-001" && build.phase !== "complete") {
      const interval = setInterval(async () => {
        try {
          const updated = await pollBuild(build.id);
          setBuild(updated);
        } catch { clearInterval(interval); }
      }, 5000);
      return () => clearInterval(interval);
    }
  }, [build.id, build.phase, pollBuild]);

  const handleCreateBuild = async () => {
    try {
      const res = await createBuild({ goal: build.goal, description: build.description });
      setBuild(prev => ({ ...prev, id: res.build_id }));
    } catch {}
  };

  return (
    <div className="flex-1 overflow-y-auto">
      <GhostModeOrchestrator
        build={build}
        onPause={async () => { try { await pauseBuild(build.id); } catch {} }}
        onResume={async () => { try { await resumeBuild(build.id); } catch {} }}
        onCancel={async () => { try { await cancelBuild(build.id); } catch {} }}
        onApprove={async (type) => {
          try {
            await submitApproval({ build_id: build.id, approval_type: type || "schema", decision: "approve" });
          } catch {}
        }}
        onTaskClick={async (taskId, action) => {
          if (action === "approve" && build.approvalType) {
            try {
              await submitApproval({ build_id: build.id, approval_type: build.approvalType, decision: "approve" });
            } catch {}
          }
        }}
      />
      {loading && <div className="fixed inset-0 bg-black/20 flex items-center justify-center z-50"><div className="text-text-primary">Loading...</div></div>}
      {error && <div className="fixed bottom-4 right-4 bg-error/10 text-error border border-error/30 rounded-lg p-4">Error: {error}</div>}
    </div>
  );
}'''

if old not in content:
    print("WARNING: old GhostModeView not found in file, using manual replacement")

content = content.replace(old, new)

with open('src/pages/StudioPage.tsx', 'w') as f:
    f.write(content)

print("Done - wired useGhostMode hook")