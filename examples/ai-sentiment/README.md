# AI Sentiment Analysis Example

This example demonstrates how to use the `ai` capability in FunctionFly functions.

## Overview

The AI capability allows functions to perform machine learning inference using pre-loaded models. This example shows sentiment analysis, but the capability supports various AI/ML tasks.

## Usage

```json
{
  "text": "I love this product! It's amazing and works perfectly."
}
```

## Response

```json
{
  "status": "success",
  "model": "sentiment",
  "text": "I love this product! It's amazing and works perfectly.",
  "sentiment": "positive",
  "confidence": 0.75,
  "analysis": {
    "positive_indicators": 3,
    "negative_indicators": 0,
    "text_length": 8
  }
}
```

## Supported Models

- `sentiment`: Sentiment analysis (positive/negative/neutral)
- `classify`: Text classification by length/content

## Security Notes

- AI inference is restricted to declared capabilities only
- Models are pre-approved and loaded by the platform
- Input size limits apply to prevent abuse