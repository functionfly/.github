import { useStudioDevOps } from "@/hooks/useStudioDevops";
import {
  CloudRegionSelector,
  DeploymentPipeline,
  EnvironmentManager,
} from "@functionfly/ui-devops";
import { Activity, Rocket, Zap } from "lucide-react";

export function DevOpsPanel() {
  const {
    stats,
    pipelines,
    environments,
    regions,
    isLoading,
    updateStage,
    retryStage,
    createEnvironment,
    updateEnvironment,
    deleteEnvironment,
    addVariable,
    addSecret,
  } = useStudioDevOps();

  // Transform API data to UI format
  const pipelineData = pipelines.length > 0 ? pipelines[0] : null;

  const handleStageSelect = (stage: any) => {
    console.log("Selected stage:", stage);
  };

  const handleStageRetry = (stageId: string) => {
    if (pipelineData) {
      retryStage(pipelineData.id, stageId);
    }
  };

  const handlePipelinePause = () => {
    console.log("Pause pipeline");
  };

  const handlePipelineResume = () => {
    console.log("Resume pipeline");
  };

  const handleEnvironmentSelect = (env: any) => {
    console.log("Selected env:", env);
  };

  const handleEnvironmentCreate = (env: any) => {
    createEnvironment({
      name: env.name || "New Environment",
      type: env.type || "development",
      color: env.color || "#06b6d4",
      variables: env.variables || {},
      replicas: env.replicas || 1,
      auto_scale: env.autoScale || false,
      region: env.region || "",
    });
  };

  const handleEnvironmentUpdate = (id: string, updates: any) => {
    updateEnvironment(id, updates);
  };

  const handleEnvironmentDelete = (id: string) => {
    deleteEnvironment(id);
  };

  const handleVariableAdd = (envId: string, key: string, value: string) => {
    addVariable(envId, key, value);
  };

  const handleSecretAdd = (envId: string, key: string) => {
    addSecret(envId, key);
  };

  const handleRegionSelect = (region: any) => {
    console.log("Selected region:", region);
  };

  const handleProviderFilter = (provider: string) => {
    console.log("Filter provider:", provider);
  };

  if (isLoading && !pipelineData) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-orange-500" />
      </div>
    );
  }

  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <h3 className="text-sm font-medium mb-1">DevOps Pipeline</h3>
        <p className="text-xs text-text-muted">Monitor and manage your deployment pipeline</p>
      </div>

      <div className="grid grid-cols-3 gap-2 mb-4">
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Rocket className="size-5 text-brand-400 mx-auto mb-1" />
          <div className="text-lg font-semibold">{stats?.pipelines ?? 0}</div>
          <div className="text-[10px] text-text-muted">Pipelines</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Activity className="size-5 text-success mx-auto mb-1" />
          <div className="text-lg font-semibold">{stats?.success_rate ?? 0}%</div>
          <div className="text-[10px] text-text-muted">Success Rate</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Zap className="size-5 text-warning mx-auto mb-1" />
          <div className="text-lg font-semibold">{((stats?.avg_cold_start_ms ?? 0) / 1000).toFixed(1)}s</div>
          <div className="text-[10px] text-text-muted">Avg Cold Start</div>
        </div>
      </div>

      {pipelineData ? (
        <DeploymentPipeline
          pipeline={{
            id: pipelineData.id,
            name: pipelineData.name,
            version: pipelineData.version,
            status: pipelineData.status,
            stages: pipelineData.stages.map((s) => ({
              id: s.id,
              name: s.name,
              status: s.status,
              duration: s.duration,
              startedAt: s.started_at,
              completedAt: s.completed_at,
              tasks: s.tasks.map((t) => ({
                id: t.id,
                name: t.name,
                status: t.status,
                duration: t.duration,
              })),
            })),
            triggeredBy: pipelineData.triggered_by,
            triggeredAt: pipelineData.triggered_at,
            branch: pipelineData.branch,
            commitSha: pipelineData.commit_sha,
            source: pipelineData.source as 'manual' | 'webhook' | 'scheduled' | 'api',
          }}
          onStageSelect={handleStageSelect}
          onStageRetry={handleStageRetry}
          onPipelinePause={handlePipelinePause}
          onPipelineResume={handlePipelineResume}
        />
      ) : (
        <div className="flex items-center justify-center h-32 border border-border-subtle rounded-lg">
          <p className="text-sm text-text-muted">No pipelines yet. Create one to get started.</p>
        </div>
      )}

      <div className="border-t border-border-subtle pt-4 space-y-4">
        <h4 className="text-xs font-medium mb-2">Environments</h4>
        <EnvironmentManager
          environments={environments.map((env) => ({
            id: env.id,
            name: env.name,
            type: env.type,
            color: env.color,
            variables: env.variables,
            secrets: env.secrets.map((s) => ({
              key: s.key,
              masked: s.masked,
              lastUpdated: s.last_updated,
            })),
            replicas: env.replicas,
            autoScale: env.auto_scale,
            region: env.region || "",
          }))}
          activeEnvironmentId={environments.length > 0 ? environments[0].id : undefined}
          onEnvironmentSelect={handleEnvironmentSelect}
          onEnvironmentCreate={handleEnvironmentCreate}
          onEnvironmentUpdate={(id, updates) => handleEnvironmentUpdate(id, updates)}
          onEnvironmentDelete={handleEnvironmentDelete}
          onVariableAdd={handleVariableAdd}
          onVariableUpdate={(id, key, value) => handleVariableAdd(id, key, value)}
          onVariableDelete={(id, key) => console.log("Delete var:", id, key)}
          onSecretAdd={handleSecretAdd}
          onSecretDelete={(id, key) => console.log("Delete secret:", id, key)}
        />

        <h4 className="text-xs font-medium mb-2">Cloud Regions</h4>
        <CloudRegionSelector
          regions={regions.map((r) => ({
            id: r.id,
            name: r.name,
            provider: r.provider,
            zone: r.zone,
            zoneName: r.zone_name,
            location: r.location,
            country: r.country,
            coordinates: r.coordinates,
            isAvailable: r.is_available,
            isRecommended: r.is_recommended || false,
            specs: r.specs ? {
              compute: r.specs.compute,
              memory: r.specs.memory,
              storage: r.specs.storage,
              gpu: r.specs.gpu,
            } : undefined,
          }))}
          selectedRegionId={regions.length > 0 ? regions[0].id : undefined}
          onRegionSelect={handleRegionSelect}
          onProviderFilter={handleProviderFilter}
        />
      </div>
    </div>
  );
}
