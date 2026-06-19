job "database-backup" {
  datacenters = ["functionfly"]
  type        = "batch"

  # Daily full backup at 2 AM UTC
  periodic {
    cron       = "0 2 * * *"
    persist    = true
    time_zone  = "UTC"
  }

  group "full-backup" {
    count = 1

    restart {
      attempts = 2
      interval = "4h"
      delay    = "30s"
      mode     = "fail"
    }

    volume "secrets" {
      type      = "host"
      source    = "functionfly-secrets"
      read_only = true
    }

    task "backup" {
      driver = "docker"

      config {
        image = "functionfly/orchestrator:latest"
        command = "/bin/sh"
        args    = ["-c", "echo 'Starting full backup' && ./bin/database-backup --mode=full --upload --retention=30 && echo 'Backup completed'"]
      }

      env {
        DB_HOST            = "pgbouncer.service.functionfly"
        DB_PORT            = "6432"
        DB_USER            = "postgres"
        DB_PASSWORD        = "${file("/secrets/db-password")}"
        DB_NAME            = "functionfly"
        DB_SSLMODE         = "require"
        R2_ACCOUNT_ID      = "${file("/secrets/r2-account-id")}"
        R2_ACCESS_KEY_ID   = "${file("/secrets/r2-access-key-id")}"
        R2_SECRET_ACCESS_KEY = "${file("/secrets/r2-secret-access-key")}"
        R2_BACKUP_BUCKET   = "functionfly-db-backups"
        ENABLE_WAL_ARCHIVING = "true"
        WAL_ARCHIVE_BUCKET = "functionfly-wal-archives"
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
        destination = "secrets/r2-account-id"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/r2-account-id\" }}"
      }

      template {
        destination = "secrets/r2-access-key-id"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/r2-access-key-id\" }}"
      }

      template {
        destination = "secrets/r2-secret-access-key"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/r2-secret-access-key\" }}"
      }
    }
  }
}

job "wal-archive" {
  datacenters = ["functionfly"]
  type        = "batch"

  # Every 15 minutes
  periodic {
    cron       = "*/15 * * * *"
    persist    = true
    time_zone  = "UTC"
  }

  group "wal-archive" {
    count = 1

    restart {
      attempts = 3
      interval = "1h"
      delay    = "30s"
      mode     = "fail"
    }

    volume "secrets" {
      type      = "host"
      source    = "functionfly-secrets"
      read_only = true
    }

    task "wal-archive" {
      driver = "docker"

      config {
        image = "functionfly/orchestrator:latest"
        command = "/bin/sh"
        args    = ["-c", "echo 'Starting WAL archive' && ./bin/database-backup --mode=wal-archive --upload && echo 'WAL archive completed'"]
      }

      env {
        DB_HOST              = "postgres-primary.service.functionfly"
        DB_PORT              = "5432"
        DB_USER              = "postgres"
        DB_PASSWORD          = "${file("/secrets/db-password")}"
        DB_NAME              = "functionfly"
        DB_SSLMODE           = "require"
        R2_ACCESS_KEY_ID     = "${file("/secrets/r2-access-key-id")}"
        R2_SECRET_ACCESS_KEY = "${file("/secrets/r2-secret-access-key")}"
        WAL_ARCHIVE_BUCKET   = "functionfly-wal-archives"
      }

      resources {
        cpu    = 100
        memory = 128
      }

      template {
        destination = "secrets/db-password"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/db-password\" }}"
      }

      template {
        destination = "secrets/r2-access-key-id"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/r2-access-key-id\" }}"
      }

      template {
        destination = "secrets/r2-secret-access-key"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/r2-secret-access-key\" }}"
      }
    }
  }
}

job "backup-verify" {
  datacenters = ["functionfly"]
  type        = "batch"

  # Weekly on Sunday at 4 AM UTC
  periodic {
    cron       = "0 4 * * 0"
    persist    = true
    time_zone  = "UTC"
  }

  group "verify" {
    count = 1

    restart {
      attempts = 2
      interval = "4h"
      delay    = "30s"
      mode     = "fail"
    }

    volume "secrets" {
      type      = "host"
      source    = "functionfly-secrets"
      read_only = true
    }

    task "verify" {
      driver = "docker"

      config {
        image = "functionfly/orchestrator:latest"
        command = "/bin/sh"
        args    = ["-c", "echo 'Starting backup verification' && ./bin/database-backup --mode=list --verify && echo 'Verification completed'"]
      }

      env {
        R2_ACCOUNT_ID        = "${file("/secrets/r2-account-id")}"
        R2_ACCESS_KEY_ID     = "${file("/secrets/r2-access-key-id")}"
        R2_SECRET_ACCESS_KEY = "${file("/secrets/r2-secret-access-key")}"
        R2_BACKUP_BUCKET     = "functionfly-db-backups"
      }

      resources {
        cpu    = 100
        memory = 128
      }

      template {
        destination = "secrets/r2-account-id"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/r2-account-id\" }}"
      }

      template {
        destination = "secrets/r2-access-key-id"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/r2-access-key-id\" }}"
      }

      template {
        destination = "secrets/r2-secret-access-key"
        env         = false
        change_mode = "restart"
        contents    = "{{ key \"functionfly/r2-secret-access-key\" }}"
      }
    }
  }
}
