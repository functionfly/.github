job "pgbouncer" {
  datacenters = ["functionfly"]
  type        = "service"

  update {
    max_parallel     = 1
    min_healthy_time = "10s"
    healthy_deadline = "2m"
    auto_revert      = true
  }

  group "pooler" {
    count = 2

    restart {
      attempts = 3
      interval = "5m"
      delay    = "15s"
      mode     = "fail"
    }

    network {
      port "pgbouncer" {
        static = 6432
        to     = 6432
      }
    }

    service {
      name = "pgbouncer"
      port = "pgbouncer"
      tags = ["pgbouncer", "connection-pooler"]

      check {
        name     = "pgbouncer-health"
        type     = "tcp"
        interval = "10s"
        timeout  = "5s"
      }
    }

    volume "secrets" {
      type      = "host"
      source    = "functionfly-secrets"
      read_only = true
    }

    task "pgbouncer" {
      driver = "docker"

      config {
        image = "edoburu/pgbouncer:latest"
        ports = ["pgbouncer"]

        env {
          DB_HOST     = "postgres-primary.service.functionfly"
          DB_PORT     = "5432"
          DB_USER     = "postgres"
          DB_PASSWORD = "${file("/secrets/db-password")}"
          DB_NAME     = "functionfly"
        }
      }

      env {
        POOL_MODE            = "transaction"
        MAX_CLIENT_CONN      = "1000"
        DEFAULT_POOL_SIZE    = "20"
        MIN_POOL_SIZE        = "5"
        RESERVE_POOL_SIZE    = "5"
        RESERVE_POOL_TIMEOUT = "3"
        QUERY_WAIT_TIMEOUT   = "120"
        IDLE_TRANSACTION_TIMEOUT = "300"
        SERVER_IDLE_TIMEOUT  = "600"
        SERVER_LIFETIME      = "3600"
      }

      resources {
        cpu    = 100
        memory = 256
      }

      template {
        destination = "secrets/db-password"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/db-password\" }}"
      }

      template {
        destination = "local/pgbouncer.ini"
        env         = false
        change_mode = "restart"
        contents    = <<-EOF
          [databases]
          functionfly = host=${env "attr.unique.network.ip-address"} port=5432 dbname=functionfly

          [pgbouncer]
          pool_mode = transaction
          max_client_conn = 1000
          default_pool_size = 20
          min_pool_size = 5
          reserve_pool_size = 5
          reserve_pool_timeout = 3
          query_wait_timeout = 120
          idle_transaction_timeout = 300
          server_idle_timeout = 600
          server_lifetime = 3600
          auth_type = md5
          auth_query = SELECT usename, passwd FROM pg_shadow WHERE usename = $1
          server_reset_query = DISCARD ALL
          log_connections = 0
          log_disconnections = 0
          log_pooler_errors = 1
          server_check_delay = 30
          server_fast_close = 1
          listen_addr = 0.0.0.0
          listen_port = 6432
        EOF
      }
    }
  }
}
