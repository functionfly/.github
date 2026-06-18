data_dir  = "/opt/nomad/data"
bind_addr = "0.0.0.0"

log_level = "INFO"

server {
  enabled          = true
  bootstrap_expect = 3
  server_affinity  = true
  num_schedulers   = 4
}

client {
  enabled = true

  node_pool = "functionfly"

  host_volume "functionfly-secrets" {
    path      = "/mnt/secrets/functionfly"
    read_only = true
  }

  host_volume "agent-config" {
    path      = "/mnt/config/agent"
    read_only = true
  }

  host_volume "ai-gateway-config" {
    path      = "/mnt/config/ai-gateway"
    read_only = true
  }

  host_volume "model-cache" {
    path      = "/mnt/storage/model-cache"
    read_only = false
  }

  host_volume "model-cache-us-west" {
    path      = "/mnt/storage-us-west/model-cache"
    read_only = false
  }

  host_volume "model-cache-eu" {
    path      = "/mnt/storage-eu/model-cache"
    read_only = false
  }

  options {
    "driver.raw_exec.enable" = "1"
    "nvidia.enabled"         = "true"
    "nvidia.visible_devices" = "all"
  }

  meta {
    "environment" = "production"
    "role"        = "worker"
  }
}

plugin "docker" {
  config {
    volumes_enabled = true
    allow_privileged = true

    gc {
      image = true
      container = true
    }

    logs {
      max_files = 4
      max_file_size = 10
    }
  }
}

plugin "nvidia" {
  config {
    minecraft = false
  }
}

advertise {
  http = "{{ GetInterfaceIP \"eth0\" }}:4646"
  rpc  = "{{ GetInterfaceIP \"eth0\" }}:4647"
  serf = "{{ GetInterfaceIP \"eth0\" }}:4648"
}

enable_debug = false

skip_leave_on_interrupt = true
leave_on_interrupt = false