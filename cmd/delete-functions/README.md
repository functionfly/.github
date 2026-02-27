# Delete Functions Script

This script permanently deletes ALL functions from the FunctionFly database.

## WARNING

⚠️ **This operation is irreversible!** It will delete:

- All user-created functions
- All function deployments
- All function execution logs
- All registry functions and versions
- All function ratings and reviews
- All execution records

**Always backup your database before running this script!**

## Usage

1. Set the `DATABASE_URL` environment variable:

   ```bash
   export DATABASE_URL="postgres://user:password@localhost/dbname?sslmode=disable"
   ```

2. Run the script:

   ```bash
   ./delete-functions
   ```

3. Confirm the operation by typing "yes" when prompted.

## Skip Confirmation

To skip the confirmation prompt (useful for automation):

```bash
./delete-functions --yes
```

## What Gets Deleted

The script deletes data from these tables in the correct order to respect foreign key constraints:

1. `registry_function_approval_comments`
2. `registry_function_approvals`
3. `registry_function_malware_scans`
4. `registry_function_signatures`
5. `registry_function_ratings`
6. `registry_executions_public`
7. `registry_function_executions`
8. `registry_function_versions`
9. `registry_functions`
10. `function_logs`
11. `function_deployments`
12. `functions`

## Building

```bash
cd cmd/delete-functions
go mod tidy
go build -o delete-functions main.go
```
