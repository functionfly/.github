# Blog Post Component Guide

This guide explains how to use enhanced components in blog posts via shortcode syntax.

## Available Components

### Callout Boxes

Highlight important information with styled callout boxes.

**Types:** `tip`, `warning`, `info`, `important`

```
:::callout[tip]
This is a helpful tip for readers.
:::

:::callout[warning]
This is a warning message.
:::

:::callout[info]
This is an informational note.
:::

:::callout[important]
This is critical information.
:::
```

### Lifecycle Flow

Display a flow of stages with badges and arrows.

```
:::lifecycle draft > published > deprecated > archived :::
```

### Workflow Steps

Show numbered workflow steps with code blocks.

```
:::workflow
1. `POST /v1/functions` — Create the function
2. `POST /v1/functions/deploy` — Deploy it
3. `POST /v1/functions/{id}/execute` — Run it
[/workflow]
```

### API Comparison Table

Display side-by-side comparison of API details.

```
:::api-table
| label | Platform | Registry |
| API Endpoint | `POST /v1/functions` | `POST /v1/registry/publish` |
| Handler | `functions.HandleCreate` | `registry.HandlePublish` |
[/api-table]
```

### Decision Grid

Show two-column decision options.

```
:::decision
[platform]
Stick with creating if:
- Item 1
- Item 2
[/platform]
[registry]
Publish to registry if:
- Item A
- Item B
[/registry]
[/decision]
```

### Comparison Cards

Show two side-by-side comparison cards.

```
:::comparison
[create]
**Create a Function**
Build a private function for your own use.
[/create]
[publish]
**Publish to Registry**
Share a function publicly with the community.
[/publish]
[/comparison]
```

## Example Blog Post

```
# My Blog Post Title

Introduction text here.

## Section Title

Content here.

:::callout[tip]
Here's a helpful tip!
:::

More content.

:::api-table
| Feature | Option A | Option B |
| Speed | Fast | Slow |
[/api-table]
```
