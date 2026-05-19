---
title: Registry Guide
description: Learn how to browse, install, and publish functions in the FunctionFly public registry.
sidebar:
  order: 101
---

The FunctionFly Registry is a public marketplace for serverless functions. Discover, share, and reuse functions built by the community.

## Quick Start

```bash
# Search the registry
ffly search <query>

# Install a function
ffly install <author>/<function-name>

# Invoke an installed function
ffly run <author>/<function-name> --data '{"key": "value"}'
```

## Key Features

- **Discover**: Browse thousands of community-built functions
- **Install**: Add registry functions to your projects with one command
- **Publish**: Share your functions with the community
- **Trust Scores**: Evaluate function quality and security

## Browsing

Visit [functionfly.com/registry](https://functionfly.com/registry) to:
- Search by name, description, or tags
- Filter by runtime (Python, JavaScript, Go, etc.)
- Sort by popularity, trust score, or recently added
- View documentation and source code

## Installing Functions

```bash
# Install latest version
ffly install <author>/<function-name>

# Install specific version
ffly install <author>/<function-name>@1.2.0

# Install to specific directory
ffly install <author>/<function-name> --dir ./my-functions/
```

## Using Installed Functions

```bash
# Invoke from CLI
ffly run <function-name> --data '{"key": "value"}'
```

Or in code:

```javascript
import { invoke } from '@functionfly/sdk';

const result = await invoke('author/function-name', {
  name: 'FunctionFly'
});
```

## Testing in Playground

Before invoking via code, test functions in the Playground at `/playground/{author}/{function-name}`.

## Publishing Functions

```bash
# Prepare for publishing
ffly publish --dry-run

# Publish
ffly publish

# Version with semantic versioning
ffly publish --bump patch  # bug fixes
ffly publish --bump minor  # new features
ffly publish --bump major # breaking changes
```

Requirements:
- Passes security scanning
- Includes documentation (README.md)
- Valid function.yaml configuration

## Trust Scores

Functions are rated 0-100% based on:
- **Code Quality**: Static analysis and security scanning
- **Documentation**: Completeness of README and examples
- **Maintenance**: Update frequency and issue response
- **Usage**: Number of installs and success rate
- **Verification**: Author identity and domain verification

Use functions with scores above 70% for production workloads.

## Managing Published Functions

```bash
# List your functions
ffly list --mine

# Update
ffly update <function-name>

# Deprecate
ffly deprecate <function-name>
```

## Next Steps

- [Secrets Vault](/secrets-vault/) - Securely manage sensitive configuration
- [Authentication](/guides/authentication/) - Secure your functions
- [CI/CD Integration](/guides/ci-cd/) - Automate deployments