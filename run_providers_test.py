#!/usr/bin/env python3
"""
Run e2e test for Providers page with both backend and frontend servers
"""
import subprocess
import sys

# Run playwright test with both servers
# Backend on 8080, Dashboard on 3000

cmd = [
    'python', '/home/micro/.kilocode/skills/webapp-testing/scripts/with_server.py',
    '--server', 'source .env && export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=functionfly DB_SSLMODE=disable REDIS_ADDR=localhost:6379 DEVELOPMENT=true SKIP_MIGRATION_VALIDATION=true VERIFICATION_ENABLED=false && ./bin/orchestrator-api --skip-migrations',
    '--port', '8080',
    '--server', 'cd web/dashboard && npx vite --host 0.0.0.0 --port 3000',
    '--port', '3000',
    '--timeout', '30',
    '--',
    'bash', '-c',
    'cd web/dashboard && source ../../.venv/bin/activate && npx playwright test e2e/providers-theme.spec.ts --project=chromium --headed=false'
]

result = subprocess.run(cmd, cwd='/home/micro/projects/functionfly')
sys.exit(result.returncode)
