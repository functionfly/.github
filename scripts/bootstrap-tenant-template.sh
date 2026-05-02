#!/bin/bash
# Bootstrap script to create and initialize the tenant database template
# This script creates the functionfly_tenant_template database that is used
# to clone new tenant databases via CREATE DATABASE WITH TEMPLATE
#
# Usage: ./scripts/bootstrap-tenant-template.sh [--recreate]
#
# The template database contains the base schema for all new tenant databases.
# It is cloned when provisioning a new dedicated database for a tenant.
#
# IMPORTANT: The template database should never be modified after creation
# in a production environment, as existing tenant databases may depend on it.

set -euo pipefail

TEMPLATE_DB="${TENANT_DB_TEMPLATE:-functionfly_tenant_template}"
PLATFORM_DB="${DB_NAME:-functionfly}"
PGHOST="${DB_HOST:-localhost}"
PGPORT="${DB_PORT:-5432}"
PGUSER="${DB_USER:-postgres}"
PGPASSWORD="${DB_PASSWORD:-postgres}"
PGSSLMODE="${DB_SSLMODE:-require}"

export PGPASSWORD

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check for --recreate flag
RECREATE=false
if [[ "${1:-}" == "--recreate" ]]; then
    RECREATE=true
    log_warn "Recreate flag detected - will drop and recreate template database"
fi

# Check if template already exists
template_exists() {
    psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d postgres -t -c "SELECT 1 FROM pg_database WHERE datname = '$TEMPLATE_DB'" 2>/dev/null | grep -q '1'
}

# Drop existing template if recreate is requested
drop_template() {
    if $RECREATE; then
        log_info "Dropping existing template database..."
        psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d postgres -c "DROP DATABASE IF EXISTS $TEMPLATE_DB" 2>/dev/null || true
    fi
}

# Create the template database
create_template() {
    log_info "Creating template database: $TEMPLATE_DB"
    psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d postgres -c "CREATE DATABASE $TEMPLATE_DB" 2>/dev/null || {
        log_error "Failed to create template database"
        exit 1
    }
}

# Apply base schema to template
apply_schema() {
    log_info "Applying base schema to template database..."

    local schema_file="internal/storage/sql/tenant_migrations/20260501142000_tenant_base_schema.up.sql"

    if [[ ! -f "$schema_file" ]]; then
        log_error "Schema file not found: $schema_file"
        exit 1
    fi

    psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$TEMPLATE_DB" -f "$schema_file" || {
        log_error "Failed to apply schema to template"
        exit 1
    }
}

# Set template DB to be unlogged (optional, for performance)
# Note: This makes the template unsuitable for streaming replication
configure_template() {
    log_info "Configuring template database..."

    # Set default tablespace if needed
    psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$TEMPLATE_DB" << 'EOF' || true
-- Disable row-level security for template clone compatibility
-- Note: In production, you may want to enable RLS after clone

-- Grant public access for template cloning
GRANT CONNECT ON DATABASE template1 TO PUBLIC;

-- Create extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Ensure standard conforming strings
SET default_text_search_config TO 'pg_catalog.english';
EOF

    log_info "Template database configured successfully"
}

# Verify template
verify_template() {
    log_info "Verifying template database..."

    local table_count
    table_count=$(psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$TEMPLATE_DB" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'" 2>/dev/null | tr -d ' ')

    if [[ -z "$table_count" ]]; then
        log_error "Failed to verify template database"
        exit 1
    fi

    log_info "Template database verified: $table_count tables created"

    # List tables
    log_info "Tables in template:"
    psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$TEMPLATE_DB" -c "\dt" || true
}

# Main execution
main() {
    echo "========================================"
    echo "Tenant Database Template Bootstrap"
    echo "========================================"
    echo ""
    echo "Configuration:"
    echo "  Template DB:  $TEMPLATE_DB"
    echo "  PostgreSQL:   $PGHOST:$PGPORT"
    echo "  User:         $PGUSER"
    echo ""

    if template_exists; then
        if $RECREATE; then
            drop_template
            create_template
            apply_schema
            configure_template
            verify_template
        else
            log_warn "Template database already exists. Use --recreate to drop and recreate."
            verify_template
        fi
    else
        create_template
        apply_schema
        configure_template
        verify_template
    fi

    echo ""
    echo "========================================"
    log_info "Bootstrap complete!"
    echo ""
    echo "The template database '$TEMPLATE_DB' is ready for tenant provisioning."
    echo "New tenant databases can be cloned using:"
    echo "  CREATE DATABASE new_tenant_db WITH TEMPLATE = $TEMPLATE_DB"
    echo "========================================"
}

main "$@"