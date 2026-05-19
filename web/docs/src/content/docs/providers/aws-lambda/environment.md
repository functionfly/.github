---
title: AWS Lambda Environment
description: Environment variables specific to AWS Lambda deployment.
---

Deploy FunctionFly functions to AWS Lambda.

## Provider Variables

| Variable | Description |
|----------|-------------|
| `AWS_LAMBDA_FUNCTION_NAME` | Lambda function name |
| `AWS_LAMBDA_FUNCTION_VERSION` | Function version |
| `AWS_LAMBDA_FUNCTION_MEMORY` | Memory limit in MB |
| `AWS_LAMBDA_FUNCTION_INITIALIZATION_TYPE` | Initialization type |
| `AWS_REGION` | AWS region |
| `AWS_DEFAULT_REGION` | Default AWS region |
| `AWS_ACCOUNT_ID` | AWS account ID |
| `AWS_EXECUTION_ENV` | Execution environment |

## Request Context

| Variable | Description |
|----------|-------------|
| `AWS_REQUEST_ID` | Unique request ID |
| `AWS_TRACING_TRACE_ID` | X-Ray trace ID |
| `AWS_TRACING_SPAN_ID` | X-Ray span ID |

## S3 Integration

| Variable | Description |
|----------|-------------|
| `S3_BUCKET_NAME` | Default S3 bucket |
| `S3_OBJECT_PREFIX` | S3 object key prefix |

## DynamoDB Integration

| Variable | Description |
|----------|-------------|
| `DYNAMODB_TABLE_NAME` | Default DynamoDB table |
| `DYNAMODB_ENDPOINT` | DynamoDB endpoint (local) |

## SQS Integration

| Variable | Description |
|----------|-------------|
| `SQS_QUEUE_URL` | SQS queue URL |
| `SQS_MAX_RECEIVE_COUNT` | Maximum receive count |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `LAMBDA_MEMORY_SIZE` | `128` | Memory in MB |
| `LAMBDA_TIMEOUT_SEC` | `3` | Timeout in seconds (max 900) |
| `LAMBDA_MAX_LAYERS` | `5` | Maximum number of layers |
| `LAMBDA_RUNTIME` | `nodejs20.x` | Lambda runtime |

## Cold Start

| Metric | Value |
|--------|-------|
| Cold start (provisioned) | < 100ms |
| Cold start (on-demand) | 200-500ms |
| Memory | Up to 10GB |
| Timeout | Up to 900 seconds |

## Secrets

Store secrets in AWS Secrets Manager:

```javascript
const secrets = JSON.parse(process.env.VAULT_SECRETS);
```

Or use Parameter Store:

```bash
aws ssm get-parameter --name /functionfly/api-key
```

## IAM Role

Required IAM permissions:

```json
{
  "Effect": "Allow",
  "Action": [
    "lambda:InvokeFunction",
    "s3:GetObject",
    "dynamodb:GetItem"
  ],
  "Resource": "*"
}
```

## Example Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "aws-lambda",
  "environment": {
    "LAMBDA_MEMORY_SIZE": "512",
    "LAMBDA_TIMEOUT_SEC": "30"
  },
  "runtime": "nodejs20.x"
}
```

## VPC Configuration

```javascript
// functionfly.jsonc
{
  "provider": "aws-lambda",
  "vpc": {
    "subnet_ids": ["subnet-abc123", "subnet-def456"],
    "security_group_ids": ["sg-123abc"]
  }
}
```