# FunctionFly Admin User Setup

This guide explains how to create admin users for your FunctionFly platform.

## Prerequisites

1. **Running Database**: You need access to a PostgreSQL database with the FunctionFly schema installed.
2. **Database Connection**: Make sure you can connect to your database using `psql` or similar tools.

## Quick Start

### Option 1: Automated Script (Recommended)

Use the automated bash script to create an admin user:

```bash
# Basic usage with defaults
./scripts/create-admin.sh

# Custom configuration
./scripts/create-admin.sh \
  --email admin@yourcompany.com \
  --password yoursecurepassword \
  --role super_admin \
  --db-host localhost \
  --db-port 5434 \
  --db-user postgres \
  --db-name functionfly
```

**Parameters:**
- `--email`: Admin user email address (default: admin@example.com)
- `--password`: Admin user password (default: admin123)
- `--role`: Admin role - `super_admin`, `support`, `billing_admin`, `developer_admin` (default: super_admin)
- `--tenant-id`: Specific tenant ID (optional - uses first available tenant if not specified)
- `--db-host`: Database host (default: localhost)
- `--db-port`: Database port (default: 5434 for Docker Postgres)
- `--db-user`: Database user (default: postgres)
- `--db-name`: Database name (default: functionfly)

### Option 2: Manual SQL Execution

If you prefer to run SQL directly:

1. Edit `scripts/create-admin-user.sql` and replace the placeholder values
2. Run the script against your database:

```bash
psql -h localhost -U postgres -d functionfly -f scripts/create-admin-user.sql
```

## Admin Roles

Choose the appropriate role for your admin user:

- **`super_admin`**: Full access to everything (tenants, users, billing, system)
- **`support`**: Limited access for support operations (tenant/app status, basic management)
- **`billing_admin`**: Billing management only (pricing, subscriptions, invoices)
- **`developer_admin`**: Development operations (apps, backends, deployments, routing)

## Accessing the Admin Panel

After creating an admin user:

1. **Login**: Go to `http://localhost:8080/login`
2. **Enter credentials**: Use the email and password you created
3. **Access admin panel**: Navigate to `http://localhost:8080/admin`

## Available Admin Sections

The admin panel includes:

- **`/admin/tenants`**: Manage tenants (create, suspend, view details)
- **`/admin/users`**: Manage users (list, invite, assign roles)
- **`/admin/billing`**: Billing management (pricing tiers, subscriptions, invoices, coupons)
- **`/admin/audit`**: View audit logs and events
- **`/admin/system`**: System health monitoring
- **`/admin/feedback`**: Manage user feedback
- **`/admin/content`**: Content management (blog posts, changelog)
- **`/admin/redirects`**: URL redirects management
- **`/admin/newsletter`**: Newsletter management
- **`/admin/content-calendar`**: Content calendar

## Security Notes

- **Change passwords**: Always change the default password after first login
- **MFA required**: Admin accounts require multi-factor authentication
- **Audit logging**: All admin actions are logged for security and compliance
- **Role-based access**: Each admin endpoint enforces specific permission checks

## Troubleshooting

### Database Connection Issues

Make sure your database is running and accessible:

```bash
# Check if PostgreSQL is running
psql -h localhost -U postgres -d functionfly -c "SELECT 1;"

# For Docker Compose setup
docker compose up -d postgres
```

### Login returns 401 Unauthorized

The API returns 401 when credentials are wrong or the account is not verified. If you created the admin with an older version of the script, the user may have `email_verified = false`. Fix it with:

```sql
-- Mark the admin as verified so login succeeds
UPDATE users SET email_verified = true WHERE email = 'admin@functionfly.local';
```

Replace the email with the one you use to log in. Then try logging in again with the same password.

Alternatively, create the admin with the **Go binary** (sets `email_verified = true`):

```bash
./create-admin -email admin@functionfly.local -password yourpassword -role super_admin
```

### User Already Exists

If you get an error that the user already exists:

```sql
-- Check existing users
SELECT id, email, role FROM users WHERE email = 'your-email@example.com';

-- Update existing user to admin role
UPDATE users SET role = 'super_admin' WHERE email = 'your-email@example.com';
```

### No Tenants Available

If no tenants exist, the script will create a default tenant automatically. Or create one manually:

```sql
INSERT INTO tenants (id, name, plan, status, created_at, updated_at)
VALUES (gen_random_uuid(), 'Default Tenant', 'free', 'active', NOW(), NOW());
```

## Development Setup

For local development with Docker:

```bash
# Start the database
docker compose up -d postgres

# Wait for it to be ready
sleep 5

# Create admin user
./scripts/create-admin.sh --email dev-admin@example.com --password devpass123

# Start the API server
go run cmd/server/main.go
```

## Production Considerations

- **Secure passwords**: Use strong, unique passwords
- **Environment variables**: Set database connection via environment variables
- **Separate admin subdomain**: Consider hosting admin panel on `admin.yourdomain.com`
- **Zero-trust access**: Implement Cloudflare Access or similar for admin panel protection
- **Regular rotation**: Rotate admin credentials regularly
