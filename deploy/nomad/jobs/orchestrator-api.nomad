job "orchestrator-api" {
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

  group "api" {
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
    }

    service {
      name = "orchestrator-api"
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

    volume "secrets" {
      type      = "host"
      source    = "functionfly-secrets"
      read_only = true
    }

    task "orchestrator-api" {
      driver = "docker"

      config {
        image = "functionfly/orchestrator-api:latest"
        ports = ["http"]

        volumes = [
          "secrets:/secrets"
        ]
      }

      env {
        DB_HOST     = "${attr.unique.network.ip-address}"
        DB_PORT     = "5432"
        DB_USER     = "postgres"
        DB_PASSWORD = "${file(\"/secrets/db-password\")}"
        DB_NAME     = "functionfly"
        DB_SSLMODE  = "disable"
        REDIS_ADDR  = "localhost:6379"
        JWT_SECRET  = "${file(\"/secrets/jwt-secret\")}"
      }

      resources {
        cpu    = 250
        memory = 256
      }

      template {
        destination = "secrets/db-password"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/db-password\" }}"
      }

      template {
        destination = "secrets/jwt-secret"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/jwt-secret\" }}"
      }
    }

    task "health-monitor" {
      driver = "docker"

      config {
        image = "functionfly/health-monitor:latest"
        ports = ["http"]
      }

      env {
        DB_HOST     = "${attr.unique.network.ip-address}"
        DB_PORT     = "5432"
        DB_USER     = "postgres"
        DB_PASSWORD = "${file(\"/secrets/db-password\")}"
        DB_NAME     = "functionfly"
      }

      resources {
        cpu    = 100
        memory = 128
      }
    }
  }
}