# Function Embeds Documentation

The FunctionFly Embed feature allows you to embed your functions directly into any website, enabling serverless function execution from client-side applications.

## Overview

Function embeddings provide a JavaScript snippet that can be included in any HTML page to execute functions remotely. This enables:

- **Serverless execution** - Run functions without managing infrastructure
- **Cross-origin requests** - Execute functions from browser-based applications
- **UI Widgets** - Optional interactive UI for function input/output
- **Analytics tracking** - Monitor embed usage by origin domain

## Quick Start

### 1. Enable Embedding

Navigate to your function's settings page in the dashboard and click the **Embed** tab. Toggle "Enable Embed" to allow external websites to embed your function.

### 2. Configure Allowed Origins

By default, embeds are open to all origins. For security, specify allowed domains:

```
example.com
*.example.com
https://app.example.com
```

### 3. Copy Embed Code

The dashboard generates a ready-to-use `<script>` tag:

```html
<script src="https://functionfly.com/embed/author/functionname.js"></script>
```

Add this to your HTML page's `<head>` or body:

```html
<!DOCTYPE html>
<html>
<head>
  <title>My App</title>
  <script src="https://functionfly.com/embed/acme/summarize.js"></script>
</head>
<body>
  <!-- Your content -->
</body>
</html>
```

## JavaScript API

After loading the embed script, a global `ff` object is available (or custom namespace if specified).

### `ff.run(input, options?)`

Execute the function with input data.

```javascript
// Basic call
const result = await ff.run({ text: "Hello world" });

// With options
const result = await ff.run(
  { text: "Hello world" },
  {
    version: "1.0.0",        // Pin to specific version
    apiKey: "your-api-key", // Provide API key if required
  }
);
```

**Parameters:**
- `input` (any): Input data to pass to the function
- `options` (object, optional):
  - `version` (string): Specific version to call
  - `apiKey` (string): API key for authentication
  - `timeout` (number): Timeout in milliseconds

**Returns:** Promise resolving to function output

### `ff.form(options?)`

Open an interactive form widget for the function.

```javascript
// Open default form
ff.form();

// Open with custom options
ff.form({
  theme: "dark",
  title: "My Function",
  onSubmit: (input) => console.log("Submitting:", input),
  onResult: (output) => console.log("Result:", output),
});
```

**Parameters:**
- `options` (object, optional):
  - `theme` (string): "light", "dark", or "auto"
  - `title` (string): Custom title for the widget
  - `onSubmit` (function): Callback when form is submitted
  - `onResult` (function): Callback with function result

### `ff.widget(options?)`

Show an inline widget with input form and output display.

```javascript
ff.widget({
  container: "#my-container",
  theme: "auto",
  showOutput: true,
});
```

**Parameters:**
- `options` (object, optional):
  - `container` (string): CSS selector for container element
  - `theme` (string): "light", "dark", or "auto"
  - `showOutput` (boolean): Show output area

## Configuration Options

### Embed Script Query Parameters

The embed script URL supports several query parameters:

| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `namespace` | string | Global variable name | `ff` |
| `autoload` | boolean | Auto-initialize on load | `true` |
| `ui` | boolean | Enable UI widget | `false` |
| `theme` | string | Theme: light/dark/auto | `auto` |

Example with all options:
```html
<script src="https://functionfly.com/embed/author/functionname.js?namespace=myFunc&ui=true&theme=dark"></script>
```

### Server-Side Configuration

Configure embed behavior in the dashboard:

| Setting | Description |
|---------|-------------|
| **Enable Embed** | Toggle embed functionality on/off |
| **Allowed Origins** | Comma-separated list of allowed domains |
| **Require API Key** | Force API key authentication |
| **UI Widget** | Enable interactive widget |
| **UI Theme** | Default theme for widgets |
| **Rate Limit** | Max requests per hour per origin |

## Security Considerations

### CORS & Origin Validation

By default, embeds accept requests from any origin. For production:

1. **Restrict allowed origins** - Only permit trusted domains
2. **Use wildcard carefully** - `*.example.com` matches subdomains
3. **Enable API key** - Require authentication for sensitive functions

### API Key Management

If "Require API Key" is enabled:

```javascript
// Include API key in requests
ff.run(input, { apiKey: "fk_live_xxxxxxxxxxxx" });
```

Get your API key from the dashboard (Settings → API Keys).

### Rate Limiting

Default rate limit is 1000 requests/hour per origin. Adjust based on expected traffic. Set to 0 for unlimited.

## Examples

### Simple Text Processing

```html
<script src="https://functionfly.com/embed/acme/text-length.js"></script>
<script>
  const length = await ff.run({ text: "Hello world" });
  console.log(`Length: ${length}`); // Length: 11
</script>
```

### With Custom Namespace

```html
<script src="https://functionfly.com/embed/acme/summarize.js?namespace= summarizer"></script>
<script>
  const summary = await summarizer.run({ text: longText });
</script>
```

### Interactive Form

```javascript
// Open form in modal
ff.form({
  title: "Sentiment Analysis",
  onResult: (sentiment) => {
    console.log("Sentiment:", sentiment.label, sentiment.score);
  }
});
```

### Error Handling

```javascript
try {
  const result = await ff.run({ data: "test" });
  console.log("Success:", result);
} catch (error) {
  if (error.code === "RATE_LIMITED") {
    console.log("Too many requests, try again later");
  } else if (error.code === "UNAUTHORIZED") {
    console.log("Invalid API key");
  } else {
    console.error("Function error:", error.message);
  }
}
```

## API Reference

### Embed Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/registry/functions/{author}/{name}/embed` | Get embed config |
| PUT | `/v1/registry/functions/{author}/{name}/embed` | Update embed config |
| GET | `/v1/registry/functions/{author}/{name}/embed/snippet` | Get code snippet |
| GET | `/v1/registry/functions/{author}/{name}/embed/analytics` | Get usage analytics |

### Embed Script URL Format

```
https://functionfly.com/embed/{author}/{name}[@{version}].js[?params]
```

### Response Format

Function execution returns:

```json
{
  "ok": true,
  "data": { "result": "output" },
  "duration_ms": 45
}
```

Error response:

```json
{
  "ok": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message"
  }
}
```

## Troubleshooting

### Embed Not Loading

1. Check browser console for errors
2. Verify the function exists and is published
3. Ensure the origin is in allowed origins list

### CORS Errors

If you see CORS errors:
- Add your domain to allowed origins
- Check that the origin matches exactly (including protocol)

### Rate Limiting

If rate limited:
- Wait before making more requests
- Contact support for limit increases
- Implement exponential backoff

### Widget Not Appearing

1. Ensure `ui=true` in script URL
2. Check that UI Widget is enabled in dashboard
3. Verify container element exists (for inline widget)
