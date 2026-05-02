---
title: CI/CD Integration
description: Automate deployments with GitHub Actions and other CI/CD pipelines.
sidebar:
  order: 7
---

Integrate FunctionFly into your continuous integration and deployment workflows for automated, reliable function deployments.

## GitHub Actions

### Quick Setup

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to FunctionFly

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup FunctionFly CLI
        uses: functionfly/setup-fly@v1
        with:
          version: latest
      
      - name: Deploy to FunctionFly
        env:
          FLY_API_KEY: ${{ secrets.FLY_API_KEY }}
        run: |
          ffly deploy --environment production
```

### Multi-Environment Deployment

```yaml
name: Deploy to FunctionFly

on:
  push:
    branches:
      - main
      - staging

jobs:
  deploy-staging:
    if: github.ref == 'refs/heads/staging'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: functionfly/setup-fly@v1
      - name: Deploy to Staging
        env:
          FLY_API_KEY: ${{ secrets.FLY_API_KEY_STAGING }}
        run: ffly deploy --environment staging

  deploy-production:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: functionfly/setup-fly@v1
      - name: Deploy to Production
        env:
          FLY_API_KEY: ${{ secrets.FLY_API_KEY_PRODUCTION }}
        run: ffly deploy --environment production
```

### Testing Before Deploy

```yaml
name: Test and Deploy

on: [push]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      
      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          pip install pytest
      
      - name: Run tests
        run: pytest

  deploy:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: functionfly/setup-fly@v1
      - name: Deploy
        env:
          FLY_API_KEY: ${{ secrets.FLY_API_KEY }}
        run: ffly deploy
```

## GitLab CI

### Basic Configuration

Create `.gitlab-ci.yml`:

```yaml
stages:
  - test
  - deploy

test:
  stage: test
  image: python:3.11
  script:
    - pip install -r requirements.txt
    - pip install pytest
    - pytest

deploy_staging:
  stage: deploy
  image: functionfly/fly-cli:latest
  script:
    - ffly deploy --environment staging
  only:
    - staging
  variables:
    FLY_API_KEY: $FLY_API_KEY_STAGING

deploy_production:
  stage: deploy
  image: functionfly/fly-cli:latest
  script:
    - ffly deploy --environment production
  only:
    - main
  variables:
    FLY_API_KEY: $FLY_API_KEY_PRODUCTION
```

## CircleCI

### Configuration

Create `.circleci/config.yml`:

```yaml
version: 2.1

orbs:
  python: circleci/python@2.1

jobs:
  test:
    docker:
      - image: cimg/python:3.11
    steps:
      - checkout
      - python/install-packages:
          pkg-manager: pip
      - run:
          name: Run tests
          command: pytest

  deploy:
    docker:
      - image: functionfly/fly-cli:latest
    steps:
      - checkout
      - run:
          name: Deploy to FunctionFly
          command: ffly deploy

workflows:
  test_and_deploy:
    jobs:
      - test
      - deploy:
          requires:
            - test
          filters:
            branches:
              only: main
```

## Jenkins

### Pipeline Script

```groovy
pipeline {
    agent any
    
    stages {
        stage('Test') {
            steps {
                sh '''
                    pip install -r requirements.txt
                    pip install pytest
                    pytest
                '''
            }
        }
        
        stage('Deploy') {
            when {
                branch 'main'
            }
            steps {
                sh 'ffly deploy --environment production'
            }
        }
    }
    
    environment {
        FLY_API_KEY = credentials('fly-api-key')
    }
}
```

## Azure DevOps

### Pipeline Configuration

Create `azure-pipelines.yml`:

```yaml
trigger:
  - main
  - staging

pool:
  vmImage: ubuntu-latest

steps:
- task: UsePythonVersion@0
  inputs:
    versionSpec: '3.11'
  displayName: 'Use Python 3.11'

- script: |
    pip install -r requirements.txt
    pip install pytest
    pytest
  displayName: 'Run tests'

- script: |
    curl -fsSL https://cli.functionfly.com/install.sh | sh
    ffly deploy --environment $(environment)
  displayName: 'Deploy to FunctionFly'
  env:
    FLY_API_KEY: $(FLY_API_KEY)
```

## Terraform Integration

### Infrastructure as Code

```hcl
# main.tf
terraform {
  required_providers {
    functionfly = {
      source = "functionfly/functionfly"
      version = "~> 1.0"
    }
  }
}

provider "functionfly" {
  api_key = var.fly_api_key
}

resource "functionfly_function" "api" {
  name    = "my-api"
  runtime = "python"
  source  = "./src"
  
  environment_variables = {
    LOG_LEVEL = "info"
  }
  
  secrets = [
    "DATABASE_URL",
    "API_KEY"
  ]
}
```

## Environment Management

### Staging vs Production

```bash
# Deploy to staging on PR
ffly deploy --environment staging \
  --tag "pr-${GITHUB_PR_NUMBER}"

# Deploy to production on merge
ffly deploy --environment production \
  --tag "v${GITHUB_SHA:0:7}"
```

### Feature Flags

```yaml
# function.yaml
deployments:
  staging:
    environment: staging
    traffic: 100
  
  production:
    environment: production
    traffic: 95
    canary:
      traffic: 5
      metric: error_rate
      threshold: 0.01
```

## Rollback Strategies

### Automatic Rollback

```yaml
# .github/workflows/deploy.yml
- name: Deploy with rollback
  run: |
    ffly deploy || \
    (echo "Deployment failed, rolling back..." && \
     ffly rollback --environment production)
```

### Blue-Green Deployment

```bash
# Deploy to green environment
ffly deploy --environment production-green

# Run smoke tests
./scripts/smoke-tests.sh https://green-api.functionfly.com

# Switch traffic
ffly traffic --environment production-green --percentage 100

# Keep blue for 24h, then remove
ffly schedule-cleanup --environment production-blue --after 24h
```

## Secrets in CI/CD

### GitHub Actions

```yaml
env:
  FLY_API_KEY: ${{ secrets.FLY_API_KEY }}
  DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

### GitLab CI

```yaml
variables:
  FLY_API_KEY: $FLY_API_KEY
  DATABASE_URL: $DATABASE_URL
```

### Best Practices

1. **Use environment-specific keys**: Separate staging and production credentials
2. **Rotate regularly**: Set up automatic key rotation
3. **Least privilege**: CI keys should only have deploy permission
4. **Audit access**: Review which pipelines can access production

## Testing Deployments

### Smoke Tests

```python
# test/smoke_test.py
import requests
import sys

def test_health_endpoint():
    response = requests.get(f"{BASE_URL}/health")
    assert response.status_code == 200
    assert response.json()['status'] == 'ok'

def test_api_endpoint():
    response = requests.post(
        f"{BASE_URL}/api/v1/data",
        json={'test': 'data'}
    )
    assert response.status_code == 200

if __name__ == '__main__':
    BASE_URL = sys.argv[1]
    test_health_endpoint()
    test_api_endpoint()
    print("Smoke tests passed!")
```

### Integration in CI

```yaml
- name: Run smoke tests
  run: |
    ffly deploy --environment staging
    python test/smoke_test.py https://staging-api.functionfly.com
```

## Monitoring Deployments

### Deployment Notifications

```yaml
- name: Notify deployment
  run: |
    curl -X POST ${{ secrets.SLACK_WEBHOOK }} \
      -H 'Content-Type: application/json' \
      -d '{
        "text": "Deployed to production",
        "commit": "${{ github.sha }}",
        "author": "${{ github.actor }}"
      }'
```

### Deployment Metrics

Track in your CI pipeline:

- Deployment duration
- Success/failure rate
- Time to first request
- Error rate post-deployment

## Best Practices

1. **Test before deploy**: Always run tests before deploying
2. **Environment parity**: Keep staging as close to production as possible
3. **Gradual rollout**: Use canary deployments for production
4. **Quick rollback**: Have rollback procedures ready
5. **Audit trail**: Log all deployments with commit info
6. **Secret management**: Use CI-native secret storage
7. **Notifications**: Alert team on deployment status

## Troubleshooting

### Common Issues

**Authentication failures:**
- Verify FLY_API_KEY is set correctly
- Check key has deploy permissions
- Ensure key hasn't expired

**Deployment timeouts:**
- Increase timeout in CI configuration
- Check function size and dependencies
- Verify network connectivity

**Test failures blocking deploy:**
- Run tests locally first
- Use `continue-on-error` for non-critical tests
- Consider flaky test detection

## Next Steps

- Learn about [monitoring and analytics](/analytics/)
- Explore [rate limiting](/guides/rate-limiting/) for production protection
- Set up [webhooks](/guides/webhooks/) for deployment notifications
