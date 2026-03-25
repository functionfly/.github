//! MicroVM Orchestrator - manages Firecracker VM lifecycle

use anyhow::{Context, Result};
use crate::firecracker::{FirecrackerClient, NetworkInterface, VMConfig, VMInstance, generate_mac};
use crate::firecracker_spawn;
use std::collections::HashMap;
use std::path::PathBuf;
use std::process::Child;
use std::time::{Duration, Instant};
use tracing::{debug, error, info, warn};

use crate::vsock::VsockClient;

/// Execution request for a MicroVM
#[derive(Debug, Clone)]
pub struct ExecutionRequest {
    /// Function code to execute
    pub code: String,
    /// Input data for the function
    pub input: String,
    /// Function handler name
    pub handler: String,
    /// Python packages to install (optional)
    pub packages: Vec<String>,
    /// Memory limit in MB
    pub memory_mb: u32,
    /// vCPU count
    pub vcpus: u32,
    /// Timeout in milliseconds
    pub timeout_ms: u64,
    /// Allowed outbound hostnames for the guest. Empty list = no network egress.
    pub network_whitelist: Vec<String>,
    /// When true, connections to hosts not on the whitelist are hard-rejected (not just logged).
    pub strict_network_whitelist: bool,
    /// Enable per-tenant Python package caching inside the VM.
    pub package_caching_enabled: bool,
}

/// Execution result from a MicroVM
#[derive(Debug, Clone)]
pub struct ExecutionResult {
    /// Output from the function
    pub output: String,
    /// Whether execution was successful
    pub success: bool,
    /// Error message if failed
    pub error: Option<String>,
    /// Execution time in milliseconds
    pub execution_time_ms: u64,
    /// Memory used in MB
    pub memory_used_mb: u32,
}

/// VM Pool entry — one Firecracker process + API client per VM (dedicated Unix socket).
struct PooledVM {
    instance: VMInstance,
    /// HTTP client for this VM's Firecracker API socket
    api: FirecrackerClient,
    /// OS child process for `firecracker --api-sock …` (None only if legacy / tests)
    fc_proc: Option<Child>,
    last_used: Instant,
    is_busy: bool,
    tenant_id: String,
}

/// Stop Firecracker via API, then kill the child process.
async fn shutdown_pooled(mut vm: PooledVM) -> Result<()> {
    if let Err(e) = vm.api.stop_vm(&vm.instance.id).await {
        warn!("Firecracker stop API failed for {}: {}", vm.instance.id, e);
    }
    if let Some(c) = vm.fc_proc.take() {
        firecracker_spawn::kill_firecracker(c);
    }
    Ok(())
}

/// MicroVM Orchestrator
pub struct MicroVMOrchestrator {
    /// Path to VM images
    image_path: PathBuf,
    /// Default vCPU count
    _default_vcpus: u32,
    /// Default memory in MB
    _default_memory_mb: u32,
    /// Maximum concurrent VMs across all tenants
    max_vms: u32,
    /// Maximum **concurrent** (active + in warm-pool) VMs per tenant.
    /// Controlled by `FUNCTIONFLY_MICROVM_MAX_VMS_PER_TENANT` (default: 10).
    max_vms_per_tenant: u32,
    /// When true, run user code with host CPython (no Firecracker). For local integration tests only.
    dev_mode: bool,
    /// Active VM instances
    active_vms: HashMap<String, PooledVM>,
    /// Warm VM pool
    warm_pool: Vec<PooledVM>,
}

impl MicroVMOrchestrator {
    /// Create a new MicroVM Orchestrator
    pub async fn new(
        socket_path: String,
        image_path: String,
        vcpus: u32,
        memory_mb: u32,
        max_vms: u32,
    ) -> Result<Self> {
        let _ = socket_path; // retained for CLI compatibility; each VM uses its own socket

        // Ensure image directory exists
        let image_path = PathBuf::from(image_path);
        if !image_path.exists() {
            std::fs::create_dir_all(&image_path)
                .context("Failed to create image directory")?;
        }

        let dev_mode = std::env::var("FUNCTIONFLY_MICROVM_DEV_MODE")
            .map(|v| v == "1" || v.eq_ignore_ascii_case("true"))
            .unwrap_or(false);

        let max_vms_per_tenant = std::env::var("FUNCTIONFLY_MICROVM_MAX_VMS_PER_TENANT")
            .ok()
            .and_then(|v| v.parse::<u32>().ok())
            .unwrap_or(10);

        if dev_mode {
            warn!("FUNCTIONFLY_MICROVM_DEV_MODE is set: executing Python on the host (no Firecracker)");
        } else {
            info!("MicroVM Orchestrator initialized: {} vCPUs, {}MB memory, max {} VMs ({} per tenant)",
                  vcpus, memory_mb, max_vms, max_vms_per_tenant);
        }

        Ok(Self {
            image_path,
            _default_vcpus: vcpus,
            _default_memory_mb: memory_mb,
            max_vms,
            max_vms_per_tenant,
            dev_mode,
            active_vms: HashMap::new(),
            warm_pool: Vec::new(),
        })
    }

    /// Execute a function in a MicroVM
    pub async fn execute(&mut self, tenant_id: &str, request: ExecutionRequest) -> Result<ExecutionResult> {
        let start = Instant::now();

        info!("Executing function for tenant {} with {} vCPUs, {}MB memory",
              tenant_id, request.vcpus, request.memory_mb);

        if self.dev_mode {
            let workdir = std::env::temp_dir().to_string_lossy().to_string();
            let executor = crate::executor::PythonExecutor::new(workdir);
            let wall = Duration::from_millis(request.timeout_ms.max(1));
            let run = executor.execute(
                &request.code,
                &request.input,
                &request.handler,
                &request.packages,
                Some(request.timeout_ms),
            );
            let exec_out = tokio::time::timeout(wall, run)
                .await
                .map_err(|_| anyhow::anyhow!("execution timed out after {}ms", request.timeout_ms))?;
            let execution_time = start.elapsed().as_millis() as u64;
            return match exec_out {
                Ok(stdout) => Ok(parse_host_python_response(&stdout, execution_time, request.memory_mb)),
                Err(e) => Ok(ExecutionResult {
                    output: String::new(),
                    success: false,
                    error: Some(e.to_string()),
                    execution_time_ms: execution_time,
                    memory_used_mb: 0,
                }),
            };
        }

        // Get or create a VM
        let mut vm = self.get_or_create_vm(tenant_id, request.vcpus, request.memory_mb).await?;

        // Send the execution request to the VM
        let result = self.send_execution(&mut vm, &request).await;

        // Return VM to pool or terminate
        let execution_time = start.elapsed().as_millis() as u64;

        if result.is_ok() {
            self.return_to_pool(vm, tenant_id).await;
        } else {
            self.terminate_vm(vm.id.to_string()).await;
        }

        match result {
            Ok(output) => Ok(ExecutionResult {
                output,
                success: true,
                error: None,
                execution_time_ms: execution_time,
                memory_used_mb: request.memory_mb,
            }),
            Err(e) => Ok(ExecutionResult {
                output: String::new(),
                success: false,
                error: Some(e.to_string()),
                execution_time_ms: execution_time,
                memory_used_mb: 0,
            }),
        }
    }

    /// Get a VM from the pool or create a new one
    async fn get_or_create_vm(&mut self, tenant_id: &str, vcpus: u32, memory_mb: u32) -> Result<VMInstance> {
        // Try to get a warm VM from the pool
        if let Some(index) = self.warm_pool.iter().position(|vm| {
            !vm.is_busy && vm.tenant_id == tenant_id
        }) {
            let mut pooled_vm = self.warm_pool.remove(index);
            pooled_vm.last_used = Instant::now();
            pooled_vm.is_busy = true;
            let instance = pooled_vm.instance.clone();
            self.active_vms.insert(instance.id.to_string(), pooled_vm);
            debug!("Reusing warm VM {} for tenant {}", instance.id, tenant_id);
            return Ok(instance);
        }

        // Per-tenant quota check: count active + warm VMs owned by this tenant.
        let tenant_count = self.active_vms.values()
            .filter(|v| v.tenant_id == tenant_id)
            .count()
            + self.warm_pool.iter()
            .filter(|v| v.tenant_id == tenant_id)
            .count();
        if tenant_count >= self.max_vms_per_tenant as usize {
            return Err(anyhow::anyhow!(
                "tenant {} has reached the per-tenant VM limit of {}",
                tenant_id, self.max_vms_per_tenant
            ));
        }

        // Global pool capacity check.
        if self.active_vms.len() >= self.max_vms as usize {
            warn!("VM pool exhausted, waiting for available VM");
            let mut attempts = 0;
            let max_attempts = 50; // 5 seconds max wait
            while self.active_vms.len() >= self.max_vms as usize && attempts < max_attempts {
                tokio::time::sleep(Duration::from_millis(100)).await;
                attempts += 1;
            }
            if self.active_vms.len() >= self.max_vms as usize {
                return Err(anyhow::anyhow!("VM pool exhausted, could not acquire VM within timeout"));
            }
        }

        // Create a new VM
        let mut pooled = self.create_new_vm(vcpus, memory_mb).await?;
        pooled.tenant_id = tenant_id.to_string();
        let instance = pooled.instance.clone();
        self.active_vms.insert(instance.id.to_string(), pooled);

        Ok(instance)
    }

    /// Create a new VM instance (spawn Firecracker + configure + start on a dedicated API socket).
    async fn create_new_vm(&self, vcpus: u32, memory_mb: u32) -> Result<PooledVM> {
        let kernel = self.image_path.join("vmlinux");
        let rootfs = self.image_path.join("python311.ext4");

        // Check if images exist
        if !kernel.exists() {
            return Err(anyhow::anyhow!(
                "Kernel image not found at {}. Please run the image builder first.",
                kernel.display()
            ));
        }

        if !rootfs.exists() {
            return Err(anyhow::anyhow!(
                "Root filesystem not found at {}. Please run the image builder first.",
                rootfs.display()
            ));
        }

        let api_sock = firecracker_spawn::allocate_api_socket_path();
        let fc_proc = firecracker_spawn::spawn_firecracker(&api_sock)?;
        firecracker_spawn::wait_for_api_socket(&api_sock, 15_000).await?;

        let api = FirecrackerClient::new(api_sock.to_string_lossy().to_string())
            .context("Failed to create Firecracker API client")?;

        // Tap device: configurable via FIRECRACKER_TAP_DEVICE (default "tap0").
        // On Kubernetes the DaemonSet initContainer creates these devices.
        // Set FIRECRACKER_SKIP_NETWORK=1 to skip the network PUT entirely (no-net VMs).
        let skip_network = std::env::var("FIRECRACKER_SKIP_NETWORK")
            .map(|v| v == "1" || v.eq_ignore_ascii_case("true"))
            .unwrap_or(false);
        let tap_device = std::env::var("FIRECRACKER_TAP_DEVICE")
            .unwrap_or_else(|_| "tap0".to_string());

        let config = VMConfig {
            vcpu_count: vcpus,
            mem_size_mib: memory_mb,
            kernel_image: kernel,
            initrd: None,
            root_drive: rootfs,
            skip_network,
            network_interface: NetworkInterface {
                guest_mac: generate_mac(),
                host_dev_name: tap_device,
            },
        };

        let instance = api.create_vm(&config).await?;
        api.start_vm(&instance.id).await?;

        info!("Created and started new VM: {}", instance.id);

        Ok(PooledVM {
            instance,
            api,
            fc_proc: Some(fc_proc),
            last_used: Instant::now(),
            is_busy: true,
            tenant_id: String::new(),
        })
    }

    /// Send execution request to a VM
    async fn send_execution(&self, vm: &VMInstance, request: &ExecutionRequest) -> Result<String> {
        // Use the VM's assigned CID for vsock communication
        let cid = vm.cid;
        let port = 1234; // Standard port for VM communication

        let client = VsockClient::new(cid, port);

        // Wait for VM to be ready
        info!("Waiting for VM {} to be ready...", vm.id);
        let mut attempts = 0;
        let max_attempts = 30; // 30 seconds max wait

        while attempts < max_attempts {
            if client.ping().await.unwrap_or(false) {
                break;
            }
            tokio::time::sleep(Duration::from_millis(1000)).await;
            attempts += 1;
        }

        if attempts >= max_attempts {
            return Err(anyhow::anyhow!("VM {} did not become ready within timeout", vm.id));
        }

        info!("VM {} is ready, loading function code", vm.id);

        // Load the function code into the VM, forwarding network/cache policy.
        client.load_function(
            &request.code,
            &request.handler,
            &request.packages,
            &request.network_whitelist,
            request.strict_network_whitelist,
            request.package_caching_enabled,
        ).await
            .context("Failed to load function code into VM")?;

        info!("Function code loaded, executing with input for VM {}", vm.id);

        // Execute the function with input
        let input_value: serde_json::Value = serde_json::from_str(&request.input)
            .unwrap_or_else(|_| serde_json::Value::String(request.input.clone()));

        let result = client.execute_function(input_value, &[]).await
            .context("Failed to execute function in VM")?;

        // Convert result back to string
        let result_str = serde_json::to_string(&result)
            .context("Failed to serialize execution result")?;

        Ok(result_str)
    }

    /// Return a VM to the warm pool
    async fn return_to_pool(&mut self, instance: VMInstance, tenant_id: &str) {
        if let Some(mut pooled) = self.active_vms.remove(&instance.id.to_string()) {
            pooled.is_busy = false;
            pooled.last_used = Instant::now();

            // Only keep a limited number of warm VMs per tenant
            let warm_count = self.warm_pool.iter()
                .filter(|vm| vm.tenant_id == tenant_id)
                .count();

            if warm_count < 3 {
                self.warm_pool.push(pooled);
                debug!("Returned VM {} to warm pool", instance.id);
            } else {
                // Terminate excess warm VMs (pooled was already removed from active_vms)
                if let Err(e) = shutdown_pooled(pooled).await {
                    error!("Failed to shut down VM {}: {}", instance.id, e);
                }
            }
        }
    }

    /// Terminate a VM
    async fn terminate_vm(&mut self, vm_id: String) {
        if let Some(pooled) = self.active_vms.remove(&vm_id) {
            if let Err(e) = shutdown_pooled(pooled).await {
                error!("Failed to shut down VM {}: {}", vm_id, e);
            }
            info!("Terminated VM: {}", vm_id);
        }
    }

    /// Clean up idle VMs from the warm pool
    pub async fn cleanup_idle_vms(&mut self, max_idle_seconds: u64) {
        let max_idle = Duration::from_secs(max_idle_seconds);
        let now = Instant::now();

        let mut keep = Vec::new();
        for vm in self.warm_pool.drain(..) {
            let idle = now.duration_since(vm.last_used);
            if idle > max_idle {
                info!("Cleaning up idle VM: {} (idle for {:?})", vm.instance.id, idle);
                if let Err(e) = shutdown_pooled(vm).await {
                    error!("Failed to shut down idle VM: {}", e);
                }
            } else {
                keep.push(vm);
            }
        }
        self.warm_pool = keep;
    }

    /// Shutdown the orchestrator
    pub async fn shutdown(&mut self) -> Result<()> {
        info!("Shutting down MicroVM Orchestrator...");

        for (_id, vm) in self.active_vms.drain() {
            if let Err(e) = shutdown_pooled(vm).await {
                error!("Failed to stop VM: {}", e);
            }
        }

        for vm in self.warm_pool.drain(..) {
            let id = vm.instance.id;
            if let Err(e) = shutdown_pooled(vm).await {
                error!("Failed to stop warm VM {}: {}", id, e);
            }
        }

        Ok(())
    }

    /// Get orchestrator statistics
    pub fn stats(&self) -> OrchestratorStats {
        OrchestratorStats {
            active_vms: self.active_vms.len() as u32,
            warm_vms: self.warm_pool.len() as u32,
            max_vms: self.max_vms,
        }
    }
}

/// Orchestrator statistics
#[derive(Debug, Clone, serde::Serialize, serde::Deserialize)]
pub struct OrchestratorStats {
    pub active_vms: u32,
    pub warm_vms: u32,
    pub max_vms: u32,
}

/// Match dev-mode host Python stdout to the same output shape as vsock (JSON string of `result` only).
fn parse_host_python_response(stdout: &str, execution_time_ms: u64, memory_mb: u32) -> ExecutionResult {
    let trimmed = stdout.trim();
    if let Ok(v) = serde_json::from_str::<serde_json::Value>(trimmed) {
        let ok = v.get("success").and_then(|x| x.as_bool()).unwrap_or(false);
        if ok {
            let inner = v.get("result").cloned().unwrap_or(serde_json::Value::Null);
            let output = serde_json::to_string(&inner).unwrap_or_else(|_| trimmed.to_string());
            return ExecutionResult {
                output,
                success: true,
                error: None,
                execution_time_ms,
                memory_used_mb: memory_mb,
            };
        }
        let err = v
            .get("error")
            .and_then(|x| x.as_str())
            .map(str::to_string)
            .unwrap_or_else(|| "Python execution failed".to_string());
        return ExecutionResult {
            output: String::new(),
            success: false,
            error: Some(err),
            execution_time_ms,
            memory_used_mb: 0,
        };
    }
    ExecutionResult {
        output: trimmed.to_string(),
        success: true,
        error: None,
        execution_time_ms,
        memory_used_mb: memory_mb,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_host_python_response_ok() {
        let r = parse_host_python_response(
            r#"{"success": true, "result": {"x": 1}}"#,
            10,
            512,
        );
        assert!(r.success);
        assert_eq!(r.output, r#"{"x":1}"#);
    }

    #[test]
    fn test_parse_host_python_response_err() {
        let r = parse_host_python_response(
            r#"{"success": false, "error": "boom"}"#,
            10,
            512,
        );
        assert!(!r.success);
        assert_eq!(r.error.as_deref(), Some("boom"));
    }

    #[tokio::test]
    #[ignore] // Requires Firecracker to be running
    async fn test_orchestrator_creation() {
        let orch = MicroVMOrchestrator::new(
            "/var/run/firecracker.sock".to_string(),
            "/tmp/test-images".to_string(),
            2,
            512,
            10,
        ).await;

        assert!(orch.is_ok());
    }
}
