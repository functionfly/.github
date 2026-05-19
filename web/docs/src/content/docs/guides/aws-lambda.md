---
title: AWS Lambda
description: Deploy FunctionFly functions to AWS Lambda
---

# AWS Lambda

Deploy your FunctionFly functions to [AWS Lambda](https://aws.amazon.com/lambda/) for enterprise-grade serverless compute with deep AWS integration.

## Features

- **Deep AWS integration** - S3, DynamoDB, SQS, and more
- **Maximum memory** - Up to 10GB
- **Longest timeout** - Up to 900 seconds
- **Pay per invocation** - No idle costs

## Prerequisites

- AWS account with Lambda permissions
- AWS CLI configured: `aws configure`

## Configuration

```jsonc
// functionfly.jsonc
{
  "provider": "aws-lambda",
  "provider_config": {
    "runtime": "nodejs18.x",
    "role": "arn:aws:iam::123456789:role/lambda-role",
    "s3_bucket": "my-functionfly-bucket",
    "vpc_config": {
      "subnet_ids": ["subnet-123", "subnet-456"],
      "security_group_ids": ["sg-789"]
    }
  }
}
```

## Deployment

```bash
# Deploy to AWS Lambda
ffly deploy --provider aws-lambda

# Deploy specific region
ffly deploy --provider aws-lambda --region us-east-1
```

## Environment Variables

```bash
# Set AWS credentials
ffly env set AWS_ACCESS_KEY_ID=AKIA... --provider aws-lambda
ffly env set AWS_SECRET_ACCESS_KEY=... --provider aws-lambda
ffly env set AWS_REGION=us-east-1 --provider aws-lambda
```

## Supported Runtimes

| Runtime | Architecture |
|---------|-------------|
| nodejs18.x | x86_64, arm64 |
| python3.11 | x86_64, arm64 |
| java17 | x86_64 |
| go1.x | x86_64 |
| provided.al2 | x86_64, arm64 |

## Limitations

- Cold start ~1 second
- Maximum 10GB memory
- Maximum 900 second timeout
- AWS-specific (not portable)