#!/bin/bash

# FunctionFly Enterprise Test Account Creation Script
# Creates a tenant with plan=enterprise and a test user for testing enterprise features (SLA, audit, support).

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_info()    { echo -e "${BLUE}ℹ️  $1${NC}"; }
print_success() { echo -e "${GREEN}✅ $1${NC}"; }
print_warning() { echo -e "${YELLOW}⚠️  $1${NC}"; }
print_error()   { echo -e "${RED}❌ $1${NC}"; }

command -v psql >/dev/null 2>&1 || {
    print_error "psql not found. Install PostgreSQL client tools."
    exit 1
}

command -v htpasswd >/dev/null 2>&1 || {
    print_error "htpasswd not found. Install apache2-utils (e.g. apt install apache2-utils) for bcrypt password hashing."
    exit 1
}

DEFAULT_EMAIL="enterprise-test@functionfly.local"
DEFAULT_PASSWORD="enterprise123"
DEFAULT_DB_HOST="localhost"
DEFAULT_DB_PORT="5432"
DEFAULT_DB_USER="postgres"
DEFAULT_DB_NAME="functionfly"
TENANT_NAME="Enterprise Test"

EMAIL=""
PASSWORD=""
DB_HOST=""
DB_PORT=""
DB_USER=""
DB_NAME=""
TENANT_ID=""
REUSE_TENANT=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --email)       EMAIL="$2"; shift 2 ;;
        --password)    PASSWORD="$2"; shift 2 ;;
        --tenant-id)   TENANT_ID="$2"; REUSE_TENANT=1; shift 2 ;;
        --db-host)     DB_HOST="$2"; shift 2 ;;
        --db-port)     DB_PORT="$2"; shift 2 ;;
        --db-user)     DB_USER="$2"; shift 2 ;;
        --db-name)     DB_NAME="$2"; shift 2 ;;
        --help|-h)
            echo "FunctionFly Enterprise Test Account"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Creates an enterprise-plan tenant and a test user for testing enterprise features."
            echo ""
            echo "Options:"
            echo "  --email EMAIL          User email (default: $DEFAULT_EMAIL)"
            echo "  --password PASSWORD    User password (default: $DEFAULT_PASSWORD)"
            echo "  --tenant-id UUID       Use existing tenant and set plan to enterprise (optional)"
            echo "  --db-host HOST         Database host (default: $DEFAULT_DB_HOST)"
            echo "  --db-port PORT         Database port (default: $DEFAULT_DB_PORT)"
            echo "  --db-user USER         Database user (default: $DEFAULT_DB_USER)"
            echo "  --db-name NAME         Database name (default: $DEFAULT_DB_NAME)"
            echo "  --help, -h             Show this help"
            echo ""
            echo "Examples:"
            echo "  $0"
            echo "  $0 --email ent@example.com --password secret"
            echo "  $0 --tenant-id \$(psql -t -c \"SELECT id FROM tenants LIMIT 1\" ...)"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage."
            exit 1
            ;;
    esac
done

EMAIL="${EMAIL:-$DEFAULT_EMAIL}"
PASSWORD="${PASSWORD:-$DEFAULT_PASSWORD}"
DB_HOST="${DB_HOST:-$DEFAULT_DB_HOST}"
DB_PORT="${DB_PORT:-$DEFAULT_DB_PORT}"
DB_USER="${DB_USER:-$DEFAULT_DB_USER}"
DB_NAME="${DB_NAME:-$DEFAULT_DB_NAME}"

export PGPASSWORD="${PGPASSWORD:-$DB_PASSWORD}"
export PGPASSWORD="${PGPASSWORD:-postgres}"

print_info "Generating password hash..."
# htpasswd -n outputs "user:hash"; use a dummy user and strip it. -B = bcrypt, -C 10 = cost.
PASSWORD_HASH=$(htpasswd -nbB -C 10 _ "$PASSWORD" | sed 's/^_://')
if [ -z "$PASSWORD_HASH" ] || [ "${PASSWORD_HASH#\$2}" = "$PASSWORD_HASH" ]; then
    print_error "Failed to generate bcrypt hash (htpasswd may not support -B -C)."
    exit 1
fi

print_info "Connecting to database $DB_HOST:$DB_PORT/$DB_NAME ..."
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
    print_error "Cannot connect to database. Check DB_* or --db-* and that PostgreSQL is running."
    exit 1
fi

# Escape single quotes for SQL
EMAIL_SQL="${EMAIL//\'/\'\'}"
USER_EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users WHERE email = '$EMAIL_SQL';" | tr -d ' ')
if [ "$USER_EXISTS" -gt 0 ]; then
    print_error "User already exists: $EMAIL"
    exit 1
fi

if [ -n "$REUSE_TENANT" ] && [ -n "$TENANT_ID" ]; then
    TENANT_EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM tenants WHERE id = '$TENANT_ID';" | tr -d ' ')
    if [ "$TENANT_EXISTS" -eq 0 ]; then
        print_error "Tenant not found: $TENANT_ID"
        exit 1
    fi
    print_info "Using existing tenant $TENANT_ID, setting plan to enterprise."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "UPDATE tenants SET plan = 'enterprise', updated_at = NOW() WHERE id = '$TENANT_ID';" >/dev/null 2>&1
else
    TENANT_ID=$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid 2>/dev/null || true)
    if [ -z "$TENANT_ID" ]; then
        print_error "Could not generate UUID (uuidgen / proc/sys/kernel/random/uuid unavailable)."
        exit 1
    fi
    print_info "Creating tenant: $TENANT_NAME (plan=enterprise)"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        INSERT INTO tenants (id, name, plan, status, created_at, updated_at)
        VALUES ('$TENANT_ID', '$TENANT_NAME', 'enterprise', 'active', NOW(), NOW())
        ON CONFLICT (id) DO UPDATE SET plan = 'enterprise', updated_at = NOW();
    " >/dev/null 2>&1
fi

# Username: strip domain from email for display
USERNAME="enterprise-test"
[[ "$EMAIL" == *@* ]] && USERNAME="${EMAIL%%@*}"
USERNAME_SQL="${USERNAME//\'/\'\'}"

print_info "Creating user in enterprise tenant..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
    INSERT INTO users (id, tenant_id, email, username, password_hash, role, email_verified, created_at, updated_at)
    VALUES (gen_random_uuid(), '$TENANT_ID', '$EMAIL_SQL', '$USERNAME_SQL', '$PASSWORD_HASH', 'owner', true, NOW(), NOW());
" >/dev/null 2>&1

USER_CREATED=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users WHERE email = '$EMAIL_SQL';" | tr -d ' ')
if [ "$USER_CREATED" -eq 0 ]; then
    print_error "Failed to create user."
    exit 1
fi

print_success "Enterprise test account created."
echo ""
print_info "User:"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
    SELECT id, email, username, role, tenant_id, created_at
    FROM users
    WHERE email = '$EMAIL_SQL';
"
echo ""
print_success "Login:"
echo "  Email:    $EMAIL"
echo "  Password: $PASSWORD"
echo ""
print_info "Enterprise features to test:"
echo "  SLA:      /enterprise/sla"
echo "  Audit:    /enterprise/audit"
echo "  Support:  /enterprise/support"
echo ""
print_info "Login at your dashboard (e.g. http://localhost:3000 or http://localhost:8080) then open the paths above."
