# Structured Logging for Fly.io

## Real-Time Logs (Free)

```bash
# Stream all logs
fly logs -a functionfly-orchestrator -f

# Filter errors
fly logs -a functionfly-orchestrator | grep -i error

# Filter JSON by level
fly logs -a functionfly-orchestrator | jq 'select(.level == "error")'

# Recent logs
fly logs -a functionfly-orchestrator | tail -100
```

No setup needed - works out of the box.
