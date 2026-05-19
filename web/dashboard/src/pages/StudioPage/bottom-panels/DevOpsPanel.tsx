import React from "react";
import {
  DeploymentPipeline,
  EnvironmentManager,
  CloudRegionSelector,
} from "@functionfly/ui-devops";
import { Rocket, Activity, Zap } from "lucide-react";

export function DevOpsPanel() {
  return (
    <div className="p-3 space-y-4">
      <div className="border-b border-border-subtle pb-3">
        <h3 className="text-sm font-medium mb-1">DevOps Pipeline</h3>
        <p className="text-xs text-text-muted">Monitor and manage your deployment pipeline</p>
      </div>

      <div className="grid grid-cols-3 gap-2 mb-4">
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Rocket className="size-5 text-brand-400 mx-auto mb-1" />
          <div className="text-lg font-semibold">12</div>
          <div className="text-[10px] text-text-muted">Pipelines</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Activity className="size-5 text-success mx-auto mb-1" />
          <div className="text-lg font-semibold">98%</div>
          <div className="text-[10px] text-text-muted">Success Rate</div>
        </div>
        <div className="bg-bg-primary rounded-lg border border-border-subtle p-3 text-center">
          <Zap className="size-5 text-warning mx-auto mb-1" />
          <div className="text-lg font-semibold">2.1s</div>
          <div className="text-[10px] text-text-muted">Avg Cold Start</div>
        </div>
      </div>

      <DeploymentPipeline
        pipeline={{
          id: "pipeline-1",
          name: "Production Deploy",
          version: "1.2.0",
          status: "active",
          stages: [
            {
              id: "build",
              name: "Build",
              status: "completed",
              duration: 45000,
              startedAt: Date.now() - 120000,
              completedAt: Date.now() - 75000,
              tasks: [
                { id: "t1", name: "Compile", status: "completed", duration: 12000 },
                { id: "t2", name: "Lint", status: "completed", duration: 8000 },
              ],
            },
            {
              id: "test",
              name: "Test",
              status: "running",
              duration: 30000,
              startedAt: Date.now() - 30000,
              tasks: [
                { id: "t3", name: "Unit Tests", status: "completed", duration: 15000 },
                { id: "t4", name: "Integration Tests", status: "running", duration: 12000 },
              ],
            },
            {
              id: "deploy",
              name: "Deploy",
              status: "waiting",
              tasks: [{ id: "t5", name: "Push to Prod", status: "pending" }],
            },
          ],
          triggeredBy: "system",
          triggeredAt: Date.now() - 120000,
          branch: "main",
          commitSha: "a1b2c3d4e5f6",
          source: "webhook",
        }}
        onStageSelect={(stage) => console.log("Selected stage:", stage)}
        onStageRetry={(stageId) => console.log("Retry stage:", stageId)}
        onPipelinePause={() => console.log("Pause pipeline")}
        onPipelineResume={() => console.log("Resume pipeline")}
      />

      <div className="border-t border-border-subtle pt-4 space-y-4">
        <h4 className="text-xs font-medium mb-2">Environments</h4>
        <EnvironmentManager
          environments={[
            {
              id: "env-1",
              name: "Production",
              type: "production",
              color: "#ef4444",
              variables: { NODE_ENV: "production", API_URL: "https://api.functionfly.io" },
              secrets: [{ key: "DB_PASSWORD", masked: true, lastUpdated: Date.now() }],
              replicas: 5,
              autoScale: true,
              region: "us-east-1",
            },
            {
              id: "env-2",
              name: "Staging",
              type: "staging",
              color: "#f59e0b",
              variables: { NODE_ENV: "staging", API_URL: "https://staging.functionfly.io" },
              secrets: [{ key: "DB_PASSWORD", masked: true, lastUpdated: Date.now() }],
              replicas: 2,
              autoScale: false,
              region: "us-west-2",
            },
            {
              id: "env-3",
              name: "Development",
              type: "development",
              color: "#22c55e",
              variables: { NODE_ENV: "development", API_URL: "http://localhost:8080" },
              secrets: [],
              replicas: 1,
              autoScale: false,
              region: "local",
            },
          ]}
          activeEnvironmentId="env-1"
          onEnvironmentSelect={(env) => console.log("Selected env:", env)}
          onEnvironmentCreate={(env) => console.log("Create env:", env)}
          onEnvironmentUpdate={(id, updates) => console.log("Update env:", id, updates)}
        />

        <h4 className="text-xs font-medium mb-2">Cloud Regions</h4>
        <CloudRegionSelector
          regions={[
            {
              id: "us-east-1",
              name: "US East (N. Virginia)",
              provider: "aws",
              zone: "us-east-1a",
              zoneName: "N. Virginia",
              location: "Virginia",
              country: "USA",
              coordinates: { lat: 37.23, lng: -78.66 },
              isAvailable: true,
              isRecommended: true,
              specs: { compute: 64, memory: 256, storage: 1024, gpu: false },
            },
            {
              id: "eu-west-1",
              name: "EU (Ireland)",
              provider: "aws",
              zone: "eu-west-1a",
              zoneName: "Ireland",
              location: "Dublin",
              country: "Ireland",
              coordinates: { lat: 53.35, lng: -6.26 },
              isAvailable: true,
              specs: { compute: 32, memory: 128, storage: 512, gpu: false },
            },
            {
              id: "ap-south-1",
              name: "Asia Pacific (Mumbai)",
              provider: "aws",
              zone: "ap-south-1a",
              zoneName: "Mumbai",
              location: "Mumbai",
              country: "India",
              coordinates: { lat: 19.08, lng: 72.88 },
              isAvailable: true,
              specs: { compute: 16, memory: 64, storage: 256, gpu: false },
            },
          ]}
          selectedRegionId="us-east-1"
          onRegionSelect={(region) => console.log("Selected region:", region)}
          onProviderFilter={(provider) => console.log("Filter provider:", provider)}
        />
      </div>
    </div>
  );
}
