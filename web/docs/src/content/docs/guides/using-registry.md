---
title: Using the Registry
description: Browse, install, and publish functions to the public FunctionFly registry.
sidebar:
  order: 2
---

The FunctionFly Registry is a public marketplace for serverless functions. Discover, share, and reuse functions built by the community.

## Browsing the Registry

### Web Interface

Visit [functionfly.com/registry](https://functionfly.com/registry) to browse available functions. You can:

- Search by name, description, or tags
- Filter by runtime (Python, JavaScript, Go, etc.)
- Sort by popularity, trust score, or recently added
- View function documentation and source code

### CLI Commands

```bash
# Search the registry
fly search <query>

# List all functions
fly list

# Get detailed information about a function
fly info <function-name>
```

## Installing Functions

Install functions from the registry to use in your projects:

```bash
# Install a function
fly install <author>/<function-name>

# Install a specific version
fly install <author>/<function-name>@1.2.0

# Install to a specific directory
fly install <author>/<function-name> --dir ./my-functions/
```

## Using Installed Functions

Once installed, you can invoke functions directly:

```bash
# Invoke an installed function
fly run <function-name> --data '{"key": "value"}'

# Invoke with a file
fly run <function-name> --file input.json
```

Or use them in your code:

```javascript
import { invoke } from '@functionfly/sdk';

const result = await invoke('author/function-name', {
  name: 'FunctionFly'
});
```

## Publishing Functions

Share your functions with the community:

### Requirements

- Function must pass security scanning
- Must include documentation (README.md)
- Must have a valid function.yaml configuration
- Code should follow best practices

### Publishing Steps

```bash
# Prepare your function for publishing
fly publish --dry-run

# Publish to the registry
fly publish

# Publish with a specific visibility
fly publish --public    # Visible to everyone
fly publish --unlisted  # Accessible but not listed
```

### Versioning

Follow semantic versioning when publishing:

```bash
# Bump patch version (bug fixes)
fly publish --bump patch

# Bump minor version (new features)
fly publish --bump minor

# Bump major version (breaking changes)
fly publish --bump major
```

## Trust Scores

Each function in the registry has a trust score based on:

- **Code Quality**: Static analysis and security scanning
- **Documentation**: Completeness of README and examples
- **Maintenance**: Update frequency and issue response
- **Usage**: Number of installs and success rate
- **Verification**: Author identity and domain verification

Trust scores range from 0-100%. We recommend using functions with scores above 70% for production workloads.

## Managing Your Published Functions

```bash
# List your published functions
fly list --mine

# Update a published function
fly update <function-name>

# Deprecate a function
fly deprecate <function-name>

# Unpublish a function
fly unpublish <function-name>
```

## Best Practices

1. **Document thoroughly**: Include usage examples, input/output schemas, and environment variables
2. **Handle errors gracefully**: Return meaningful error messages
3. **Test extensively**: Include unit tests and integration tests
4. **Version properly**: Use semantic versioning for breaking changes
5. **Keep dependencies minimal**: Reduce cold start times and security surface area

## Next Steps

- Learn about [authentication](/guides/authentication/) for securing your functions
- Explore [rate limiting](/guides/rate-limiting/) to manage usage
- Read about [CI/CD integration](/guides/ci-cd/) for automated publishing
