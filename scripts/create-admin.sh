#!/bin/bash

# FunctionFly Admin User Creation Script
# This script generates SQL to create an admin user in your database
#
# Production: prefer `go run ./cmd/create-admin -production` with ADMIN_CREATE_PASSWORD
# so passwords are strong and not echoed. Defaults below (e.g. admin123) are dev-only.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if required tools are available
command -v psql >/dev/null 2>&1 || {
    print_error "psql command not found. Please install PostgreSQL client tools."
    exit 1
}

# SECURITY: Default password is only allowed in DEVELOPMENT mode
if [ "$DEVELOPMENT" != "true" ]; then
    if [ -z "$ADMIN_CREATE_PASSWORD" ]; then
        echo "ERROR: In production, you must set ADMIN_CREATE_PASSWORD env var for the admin password."
        echo "       Alternatively, set DEVELOPMENT=true to use default dev credentials (NOT recommended for production)."
        exit 1
    fi
    DEFAULT_PASSWORD="$ADMIN_CREATE_PASSWORD"
else
    # Only use default dev password in DEVELOPMENT mode
    DEFAULT_PASSWORD="${ADMIN_CREATE_PASSWORD:-admin123}"
fi
DEFAULT_EMAIL="admin@functionfly.com"
DEFAULT_ROLE="super_admin"
DEFAULT_DB_HOST="localhost"
DEFAULT_DB_PORT="5434"
DEFAULT_DB_USER="postgres"
DEFAULT_DB_NAME="functionfly"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --email)
            EMAIL="$2"
            shift 2
            ;;
        --password)
            PASSWORD="$2"
            shift 2
            ;;
        --role)
            ROLE="$2"
            shift 2
            ;;
        --tenant-id)
            TENANT_ID="$2"
            shift 2
            ;;
        --db-host)
            DB_HOST="$2"
            shift 2
            ;;
        --db-port)
            DB_PORT="$2"
            shift 2
            ;;
        --db-user)
            DB_USER="$2"
            shift 2
            ;;
        --db-name)
            DB_NAME="$2"
            shift 2
            ;;
        --help|-h)
            echo "FunctionFly Admin User Creation Script"
            echo ""
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --email EMAIL          Admin user email (default: $DEFAULT_EMAIL)"
            echo "  --password PASSWORD    Admin user password (default: $DEFAULT_PASSWORD)"
            echo "  --role ROLE           Admin role: super_admin, support, billing_admin, developer_admin (default: $DEFAULT_ROLE)"
            echo "  --tenant-id UUID       Tenant ID for the admin user (optional, will use first tenant if not specified)"
            echo "  --db-host HOST         Database host (default: $DEFAULT_DB_HOST)"
            echo "  --db-port PORT         Database port (default: $DEFAULT_DB_PORT)"
            echo "  --db-user USER         Database user (default: $DEFAULT_DB_USER)"
            echo "  --db-name NAME         Database name (default: $DEFAULT_DB_NAME)"
            echo "  --help, -h            Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0 --email admin@mycompany.com --password mysecurepass --role super_admin"
            echo "  $0 --tenant-id 123e4567-e89b-12d3-a456-426614174000"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Set defaults
EMAIL=${EMAIL:-$DEFAULT_EMAIL}
PASSWORD=${PASSWORD:-$DEFAULT_PASSWORD}
ROLE=${ROLE:-$DEFAULT_ROLE}
DB_HOST=${DB_HOST:-$DEFAULT_DB_HOST}
DB_PORT=${DB_PORT:-$DEFAULT_DB_PORT}
DB_USER=${DB_USER:-$DEFAULT_DB_USER}
DB_NAME=${DB_NAME:-$DEFAULT_DB_NAME}

# Validate role
case $ROLE in
    super_admin|support|billing_admin|developer_admin)
        ;;
    *)
        print_error "Invalid role: $ROLE"
        print_info "Valid roles: super_admin, support, billing_admin, developer_admin"
        exit 1
        ;;
esac

# Generate password hash: use bcrypt via htpasswd, or Go create-admin (API expects bcrypt/Argon2, not plain SHA256)
print_info "Generating password hash..."
if command -v htpasswd >/dev/null 2>&1; then
    PASSWORD_HASH=$(echo -n "$PASSWORD" | htpasswd -bnBC 10 "" | tr -d ':\n' | sed 's/$2y/$2b/')
else
    print_warning "htpasswd not found. Using Go create-admin (creates user with Argon2 hash compatible with API)."
    PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
    export DB_HOST DB_PORT DB_USER DB_NAME
    export DB_PASSWORD="${DB_PASSWORD:-$PGPASSWORD}"
    export DB_PASSWORD="${DB_PASSWORD:-postgres}"
    if (cd "$PROJECT_ROOT" && go run ./cmd/create-admin -email "$EMAIL" -password "$PASSWORD" -role "$ROLE"); then
        print_success "Admin user created via Go (login will work with API)."
        exit 0
    else
        print_error "Go create-admin failed. Install htpasswd (apache2-utils) for bcrypt, or fix DB connection (DB_HOST=$DB_HOST DB_PORT=$DB_PORT)."
        exit 1
    fi
fi

print_info "Connecting to database..."
print_info "Host: $DB_HOST:$DB_PORT"
print_info "Database: $DB_NAME"
print_info "User: $DB_USER"

# Check database connection
if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1;" >/dev/null 2>&1; then
    print_error "Failed to connect to database. Please check your connection parameters."
    print_info "Make sure PostgreSQL is running and accessible."
    exit 1
fi

print_info "Checking for existing tenants..."

# Get tenant ID or create one
if [ -n "$TENANT_ID" ]; then
    # Verify the provided tenant exists
    TENANT_EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM tenants WHERE id = '$TENANT_ID';" | tr -d ' ')
    if [ "$TENANT_EXISTS" -eq 0 ]; then
        print_error "Tenant with ID $TENANT_ID does not exist."
        exit 1
    fi
    print_info "Using existing tenant: $TENANT_ID"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "UPDATE tenants SET plan = 'enterprise', updated_at = NOW() WHERE id = '$TENANT_ID';" >/dev/null 2>&1
else
    # Get the first tenant or create a default one
    TENANT_COUNT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM tenants;" | tr -d ' ')
    if [ "$TENANT_COUNT" -gt 0 ]; then
        TENANT_ID=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT id FROM tenants LIMIT 1;" | tr -d ' ')
        TENANT_NAME=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT name FROM tenants WHERE id = '$TENANT_ID';" | tr -d ' ')
        print_info "Using existing tenant: $TENANT_NAME ($TENANT_ID)"
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "UPDATE tenants SET plan = 'enterprise', updated_at = NOW() WHERE id = '$TENANT_ID';" >/dev/null 2>&1
    else
        # Create a default tenant
        TENANT_ID=$(uuidgen 2>/dev/null || cat /proc/sys/kernel/random/uuid 2>/dev/null || echo "$(date +%s)-$(head -c 4 /dev/urandom | od -An -tx1 | tr -d ' ')")
        print_info "Creating default tenant with ID: $TENANT_ID"

        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
            INSERT INTO tenants (id, name, plan, status, created_at, updated_at)
            VALUES ('$TENANT_ID', 'Default Tenant', 'enterprise', 'active', NOW(), NOW())
            ON CONFLICT (id) DO NOTHING;
        " >/dev/null 2>&1
    fi
fi

# Check if user already exists
USER_EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users WHERE email = '$EMAIL';" | tr -d ' ')
if [ "$USER_EXISTS" -gt 0 ]; then
    print_error "User with email $EMAIL already exists."
    exit 1
fi

# Create the admin user (email_verified = true so login works without verification flow)
print_info "Creating admin user..."
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
    INSERT INTO users (id, tenant_id, email, username, password_hash, role, email_verified, created_at, updated_at)
    VALUES (gen_random_uuid(), '$TENANT_ID', '$EMAIL', 'functionfly', '$PASSWORD_HASH', '$ROLE', true, NOW(), NOW())
    ON CONFLICT (email) DO NOTHING;
" >/dev/null 2>&1

# Verify the user was created
USER_CREATED=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT COUNT(*) FROM users WHERE email = '$EMAIL';" | tr -d ' ')
if [ "$USER_CREATED" -gt 0 ]; then
    print_success "Admin user created successfully!"
    echo ""
    print_info "User Details:"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
        SELECT id, email, username, role, tenant_id, created_at
        FROM users
        WHERE email = '$EMAIL';
    "

    echo ""
    print_success "Login Credentials:"
    echo "  Email: $EMAIL"
    echo "  Password: $PASSWORD"
    echo ""
    print_success "Admin Panel Access:"
    echo "  1. Login at: http://localhost:8080/login"
    echo "  2. Navigate to: http://localhost:8080/admin"
    echo ""
    print_warning "Remember to change the password after first login!"
else
    print_error "Failed to create admin user."
    exit 1
fi
