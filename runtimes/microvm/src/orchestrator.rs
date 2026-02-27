//! MicroVM Orchestrator - manages Firecracker VM lifecycle

use anyhow::{Context, Result};
use crate::firecracker::{FirecrackerClient, NetworkInterface, VMConfig, VMInstance, generate_mac};
use std::collections::HashMap;
use std::path::PathBuf;
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

/// VM Pool entry
#[derive(Debug)]
struct PooledVM {
    instance: VMInstance,
    last_used: Instant,
    is_busy: bool,
    tenant_id: String,
}

/// MicroVM Orchestrator
pub struct MicroVMOrchestrator {
    /// Firecracker API client
    client: FirecrackerClient,
    /// Path to VM images
    image_path: PathBuf,
    /// Default vCPU count
    default_vcpus: u32,
    /// Default memory in MB
    default_memory_mb: u32,
    /// Maximum concurrent VMs
    max_vms: u32,
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
        let client = FirecrackerClient::new(socket_path)
            .context("Failed to create Firecracker client")?;

        // Ensure image directory exists
        let image_path = PathBuf::from(image_path);
        if !image_path.exists() {
            std::fs::create_dir_all(&image_path)
                .context("Failed to create image directory")?;
        }

        info!("MicroVM Orchestrator initialized: {} vCPUs, {}MB memory, max {} VMs",
              vcpus, memory_mb, max_vms);

        Ok(Self {
            client,
            image_path,
            default_vcpus: vcpus,
            default_memory_mb: memory_mb,
            max_vms,
            active_vms: HashMap::new(),
            warm_pool: Vec::new(),
        })
    }

    /// Execute a function in a MicroVM
    pub async fn execute(&mut self, tenant_id: &str, request: ExecutionRequest) -> Result<ExecutionResult> {
        let start = Instant::now();

        info!("Executing function for tenant {} with {} vCPUs, {}MB memory",
              tenant_id, request.vcpus, request.memory_mb);

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

        // Check if we can create a new VM
        if self.active_vms.len() >= self.max_vms as usize {
            // Wait for a VM to become available with timeout
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
        let instance: VMInstance = self.create_new_vm(vcpus, memory_mb).await?;

        let pooled_vm = PooledVM {
            instance: instance.clone(),
            last_used: Instant::now(),
            is_busy: true,
            tenant_id: tenant_id.to_string(),
        };
        self.active_vms.insert(instance.id.to_string(), pooled_vm);

        Ok(instance)
    }

    /// Create a new VM instance
    async fn create_new_vm(&self, vcpus: u32, memory_mb: u32) -> Result<VMInstance> {
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

        let config = VMConfig {
            vcpu_count: vcpus,
            mem_size_mib: memory_mb,
            kernel_image: kernel,
            initrd: None,
            root_drive: rootfs,
            network_interface: NetworkInterface {
                guest_mac: generate_mac(),
                host_dev_name: "tap0".to_string(),
            },
        };

        let instance = self.client.create_vm(&config).await?;
        self.client.start_vm(&instance.id).await?;

        info!("Created and started new VM: {}", instance.id);

        Ok(instance)
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

        // Load the function code into the VM
        client.load_function(&request.code, &request.handler, &request.packages).await
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
                // Terminate excess warm VMs
                self.terminate_vm(instance.id.to_string()).await;
            }
        }
    }

    /// Terminate a VM
    async fn terminate_vm(&mut self, vm_id: String) {
        if let Some(pooled) = self.active_vms.remove(&vm_id) {
            if let Err(e) = self.client.stop_vm(&pooled.instance.id).await {
                error!("Failed to stop VM {}: {}", vm_id, e);
            }
            info!("Terminated VM: {}", vm_id);
        }
    }

    /// Clean up idle VMs from the warm pool
    pub async fn cleanup_idle_vms(&mut self, max_idle_seconds: u64) {
        let max_idle = Duration::from_secs(max_idle_seconds);
        let now = Instant::now();

        self.warm_pool.retain(|vm| {
            let idle = now.duration_since(vm.last_used);
            if idle > max_idle {
                info!("Cleaning up idle VM: {} (idle for {:?})", vm.instance.id, idle);
                false // Remove
            } else {
                true // Keep
            }
        });
    }

    /// Shutdown the orchestrator
    pub async fn shutdown(&mut self) -> Result<()> {
        info!("Shutting down MicroVM Orchestrator...");

        // Stop all active VMs
        for (id, vm) in self.active_vms.drain() {
            if let Err(e) = self.client.stop_vm(&vm.instance.id).await {
                error!("Failed to stop VM {}: {}", id, e);
            }
        }

        // Stop all warm VMs
        for vm in self.warm_pool.drain(..) {
            if let Err(e) = self.client.stop_vm(&vm.instance.id).await {
                error!("Failed to stop warm VM {}: {}", vm.instance.id, e);
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
#[derive(Debug, Clone)]
pub struct OrchestratorStats {
    pub active_vms: u32,
    pub warm_vms: u32,
    pub max_vms: u32,
}

#[cfg(test)]
mod tests {
    use super::*;

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
