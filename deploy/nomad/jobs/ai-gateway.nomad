job "ai-gateway" {
  datacenters = ["functionfly"]
  type        = "service"

  update {
    max_parallel     = 1
    min_healthy_time = "45s"
    healthy_deadline = "3m"
    auto_revert      = true
  }

  migrate {
    max_parallel     = 1
    min_healthy_time = "60s"
    health_check     = "checks"
    healthy_deadline = "10m"
  }

  group "gateway" {
    count = 3

    restart {
      attempts = 3
      interval = "10m"
      delay    = "30s"
      mode     = "fail"
    }

    spread {
      attribute = "${node.datacenter}"
    }

    spread {
      attribute = "${attr.unique.network.ip-address}"
    }

    network {
      port "http" {
        static = 8082
        to     = 8082
      }
      port "metrics" {
        static = 9090
        to     = 9090
      }
    }

    service {
      name = "ai-gateway"
      port = "http"
      tags = ["http", "api", "ai"]

      check {
        name     = "http-health"
        type     = "http"
        path     = "/health"
        interval = "15s"
        timeout  = "5s"
        initial_status = "passing"
      }
    }

    service {
      name = "ai-gateway-metrics"
      port = "metrics"
      tags = ["metrics"]
    }

    volume "model-cache" {
      type      = "host"
      source    = "model-cache"
      read_only = false
    }

    volume "config" {
      type      = "host"
      source    = "ai-gateway-config"
      read_only = true
    }

    task "ai-gateway" {
      driver = "docker"

      config {
        image = "functionfly/ai-gateway:v1.0.0"
        ports = ["http", "metrics"]

        volumes = [
          "model-cache:/tmp/model-cache",
          "config:/app/config"
        ]

        ulimit {
          nofile = "65536"
        }
      }

      env {
        INFERENCE_ENGINE           = "onnx"
        MODEL_CACHE_DIR            = "/tmp/model-cache"
        MAX_BATCH_SIZE             = "16"
        INFERENCE_TIMEOUT_SECONDS  = "180"
        CLUSTER_ENABLED            = "true"
        CLUSTER_REFRESH_INTERVAL   = "15"
        LOG_LEVEL                  = "WARNING"
        LOG_FORMAT                 = "json"
        RUNPOD_API_BASE_URL        = "https://api.runpod.io"
        RUNPOD_CLUSTER_NAME        = "production"
        DEFAULT_MAX_TOKENS         = "4096"
        DEFAULT_TEMPERATURE        = "0.7"
        DEFAULT_TOP_P              = "0.9"
        RATE_LIMIT_REQUESTS_PER_MINUTE = "500"
        RATE_LIMIT_REQUESTS_PER_HOUR   = "10000"
        REQUEST_TIMEOUT            = "180"
        CONNECT_TIMEOUT            = "10"
        READ_TIMEOUT               = "180"
      }

      resources {
        cpu    = 1000
        memory = 4096

        device "nvidia.com/gpu" {
          count = 1
        }
      }

      template {
        destination = "secrets/runpod-api-key"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"ai-gateway/runpod-api-key\" }}"
      }

      template {
        destination = "config/runpod.env"
        env         = true
        change_mode = "restart"
        contents    = <<-EOF
          RUNPOD_API_KEY={{ key "ai-gateway/runpod-api-key" }}
        EOF
      }

      constraint {
        attribute = "${attr.nvidia.com/gpu}"
        operator  = ">"
        value     = "0"
      }

      kill_timeout = "60s"
      shutdown_delay = "30s"
    }
  }

  group "gateway-us-west" {
    count = 2

    constraint {
      attribute = "${node.datacenter}"
      value     = "us-west-2"
    }

    restart {
      attempts = 3
      interval = "10m"
      delay    = "30s"
      mode     = "fail"
    }

    network {
      port "http" {
        static = 8082
        to     = 8082
      }
    }

    service {
      name = "ai-gateway-us-west"
      port = "http"
      tags = ["http", "api", "ai", "region-us-west"]
    }

    volume "model-cache" {
      type      = "host"
      source    = "model-cache-us-west"
      read_only = false
    }

    task "ai-gateway-us-west" {
      driver = "docker"

      config {
        image = "functionfly/ai-gateway:v1.0.0"
        ports = ["http"]

        volumes = [
          "model-cache:/tmp/model-cache"
        ]
      }

      env {
        INFERENCE_ENGINE           = "onnx"
        MODEL_CACHE_DIR            = "/tmp/model-cache"
        CLUSTER_ENABLED            = "true"
        CLUSTER_REFRESH_INTERVAL   = "15"
        RUNPOD_API_BASE_URL        = "https://api.runpod.io"
        RUNPOD_CLUSTER_NAME        = "production"
      }

      resources {
        cpu    = 1000
        memory = 4096

        device "nvidia.com/gpu" {
          count = 1
        }
      }

      constraint {
        attribute = "${attr.nvidia.com/gpu}"
        operator  = ">"
        value     = "0"
      }
    }
  }

  group "gateway-eu" {
    count = 1

    constraint {
      attribute = "${node.datacenter}"
      value     = "eu-west-1"
    }

    restart {
      attempts = 3
      interval = "10m"
      delay    = "30s"
      mode     = "fail"
    }

    network {
      port "http" {
        static = 8082
        to     = 8082
      }
    }

    service {
      name = "ai-gateway-eu"
      port = "http"
      tags = ["http", "api", "ai", "region-eu"]
    }

    volume "model-cache" {
      type      = "host"
      source    = "model-cache-eu"
      read_only = false
    }

    task "ai-gateway-eu" {
      driver = "docker"

      config {
        image = "functionfly/ai-gateway:v1.0.0"
        ports = ["http"]

        volumes = [
          "model-cache:/tmp/model-cache"
        ]
      }

      env {
        INFERENCE_ENGINE           = "onnx"
        MODEL_CACHE_DIR            = "/tmp/model-cache"
        CLUSTER_ENABLED            = "true"
        CLUSTER_REFRESH_INTERVAL   = "15"
        RUNPOD_API_BASE_URL        = "https://api.runpod.io"
        RUNPOD_CLUSTER_NAME        = "production-eu"
      }

      resources {
        cpu    = 1000
        memory = 4096

        device "nvidia.com/gpu" {
          count = 1
        }
      }

      constraint {
        attribute = "${attr.nvidia.com/gpu}"
        operator  = ">"
        value     = "0"
      }
    }
  }
}