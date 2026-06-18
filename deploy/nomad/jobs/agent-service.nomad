job "agent-service" {
  datacenters = ["functionfly"]
  type        = "service"

  update {
    max_parallel     = 1
    min_healthy_time = "10s"
    healthy_deadline = "2m"
    auto_revert      = true
  }

  migrate {
    max_parallel     = 1
    min_healthy_time = "30s"
    health_check     = "checks"
    healthy_deadline = "5m"
  }

  group "agents" {
    count = 3

    restart {
      attempts = 3
      interval = "5m"
      delay    = "15s"
      mode     = "fail"
    }

    network {
      port "http" {
        static = 8080
        to     = 8080
      }
      port "metrics" {
        static = 9090
        to     = 9090
      }
    }

    service {
      name = "agent-service"
      port = "http"
      tags = ["http", "api"]

      check {
        name     = "http-health"
        type     = "http"
        path     = "/health"
        interval = "10s"
        timeout  = "5s"
      }
    }

    service {
      name = "agent-service-metrics"
      port = "metrics"
      tags = ["metrics"]
    }

    volume "config" {
      type      = "host"
      source    = "agent-config"
      read_only = true
    }

    task "agent-service" {
      driver = "docker"

      config {
        image = "functionfly/agent-service:latest"
        ports = ["http", "metrics"]

        volumes = [
          "config:/app/config"
        ]
      }

      env {
        DATABASE_URL               = "postgres://postgres:${file(\"/secrets/db-password\")}@${attr.unique.network.ip-address}:5432/functionfly?sslmode=disable"
        REDIS_ADDR                 = "localhost:6379"
        AGENT_WALLET_LOW_BALANCE_USD = "5.00"
        SKIP_MIGRATION_VALIDATION  = "true"
        VERIFICATION_ENABLED       = "false"
        CLUSTER_MANAGER_ENABLED    = "true"
        CLUSTER_REFRESH_INTERVAL   = "30"
      }

      resources {
        cpu    = 500
        memory = 512
      }

      template {
        destination = "secrets/db-password"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/db-password\" }}"
      }

      template {
        destination = "config/runpod.env"
        env         = true
        change_mode = "restart"
        contents    = "{{ key \"agent/runpod\" }}"
      }
    }
  }

  group "agent-scaler" {
    count = 1

    restart {
      attempts = 2
      interval = "5m"
      delay    = "30s"
      mode     = "fail"
    }

    network {
      port "http" {
        static = 8081
        to     = 8081
      }
    }

    task "agent-scaler" {
      driver = "docker"

      config {
        image = "functionfly/agent-scaler:latest"
        ports = ["http"]
      }

      env {
        REDIS_ADDR               = "localhost:6379"
        AGENT_SERVICE_URL        = "http://agent-service.service.functionfly:8080"
        MIN_REPLICAS             = "2"
        MAX_REPLICAS             = "10"
        TARGET_CPU_UTILIZATION   = "70"
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}