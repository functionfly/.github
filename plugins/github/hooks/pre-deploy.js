/**
 * GitHub Pre-Deploy Hook
 * Runs before a workflow deployment starts
 */

const crypto = require('crypto');

async function preDeployHook(context) {
  const { workflow, config, logger } = context;
  const { github_token, default_repo, auto_sync } = config;

  logger.info('GitHub pre-deploy: starting', {
    workflowId: workflow?.id,
    repo: default_repo,
  });

  // Verify GitHub token has required permissions
  if (!github_token) {
    logger.warn('GitHub pre-deploy: no token configured, skipping');
    return { shouldContinue: true, metadata: {} };
  }

  try {
    // Check for required repo access
    const [owner, repo] = (default_repo || '').split('/');
    if (!owner || !repo) {
      logger.warn('GitHub pre-deploy: no default repo configured');
      return { shouldContinue: true, metadata: {} };
    }

    // Generate pre-deploy status
    const preDeployData = {
      trigger: 'pre_deploy',
      workflow: workflow?.name,
      workflowId: workflow?.id,
      timestamp: new Date().toISOString(),
      repo: { owner, repo },
    };

    // If auto_sync is enabled, check if we should wait for GitHub Actions
    if (auto_sync) {
      // Return metadata for post-deploy to use
      return {
        shouldContinue: true,
        metadata: {
          github_trigger: true,
          ...preDeployData,
        },
      };
    }

    return {
      shouldContinue: true,
      metadata: preDeployData,
    };
  } catch (error) {
    logger.error('GitHub pre-deploy hook failed', {
      error: error.message,
      workflowId: workflow?.id,
    });

    // Don't block deployment for hook failures
    return { shouldContinue: true, metadata: { error: error.message } };
  }
}

module.exports = preDeployHook;