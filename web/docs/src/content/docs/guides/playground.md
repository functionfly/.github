---
title: Function Playground
description: Test and debug your functions interactively using the FunctionFly playground.
sidebar:
  order: 3
---

The Function Playground is an interactive environment where you can test your functions before deploying them to production. It provides a real-time execution environment with support for streaming responses, variable substitution, and detailed execution history.

## Accessing the Playground

You can access the playground in two ways:

1. **From the Registry**: Visit any function's page and click the "Try It" button
2. **Direct URL**: Navigate to `/playground/{author}/{function-name}`

## Interface Overview

The playground consists of several panels:

### Input Panel

The input panel accepts JSON input for your function. You can switch between:

- **Form Mode**: A structured form view based on the function's input schema
- **JSON Mode**: Direct JSON editor with syntax highlighting
- **Examples Mode**: Select from pre-configured examples or the function's default input

### Output Panel

The output panel displays the execution results:

- **Response Tab**: The function's returned data
- **Headers Tab**: Response headers (useful for debugging)
- **Timeline Tab**: Execution timeline visualization
- **Diff Tab**: Compare outputs from different executions

### Sidebar

The sidebar provides additional tools:

- **History**: View past executions and compare results
- **Variables**: Define reusable variables with `{{variable_name}}` syntax
- **Schema**: View the function's input/output schema
- **Snippets**: Code snippets for calling the function from various SDKs
- **Share**: Generate shareable links with pre-filled input
- **Info**: Function metadata, pricing, and reliability scores

## Variable Substitution

Use variables to reuse values across multiple executions:

```
{{api_key}}
{{user_id}}
{{base_url}}
```

Variables are defined in the sidebar and persist across sessions.

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd/Ctrl + Enter` | Execute function |
| `Cmd/Ctrl + Shift + F` | Format JSON |
| `Cmd/Ctrl + Shift + R` | Reset playground |
| `Cmd/Ctrl + Shift + C` | Copy shareable link |

## Sharing and Embedding

### Shareable Links

Click "Share" in the sidebar to generate a URL that encodes:

- Function author and name
- Input data
- Selected version

### Embed Snippets

Get embeddable code snippets for:

- cURL
- JavaScript (Fetch)
- Python
- Go

## Adding Examples to Your Function

To populate the Examples dropdown, add an `examples` array to your function's manifest:

```json
{
  "name": "my-function",
  "version": "1.0.0",
  "runtime": "nodejs",
  "input": {
    "type": "object",
    "properties": {
      "name": { "type": "string" },
      "count": { "type": "integer" }
    }
  },
  "examples": [
    {
      "name": "Basic Example",
      "input": {
        "name": "World",
        "count": 42
      },
      "description": "A simple example demonstrating basic usage"
    },
    {
      "name": "Empty Count",
      "input": {
        "name": "Test"
      },
      "description": "Example with default count"
    }
  ]
}
```

Each example must include:

- `name`: Display name shown in the dropdown
- `input`: Valid JSON input matching the function's schema
- `description` (optional): Additional context about the example

## Best Practices

1. **Use descriptive example names**: Make it clear what each example demonstrates
2. **Include edge cases**: Add examples for boundary conditions and error handling
3. **Keep inputs realistic**: Examples should reflect real-world usage patterns
4. **Document variable usage**: If using variables, include comments explaining their purpose

## Next Steps

- Learn about [secrets vault](/guides/secrets-vault/) for secure credential storage
- Explore [state fabric](/guides/statefabric/) for persistent state management
- Read the [functions API reference](/api/functions/) for programmatic access