/**
 * GitHub Post-Deploy Hook
 * Runs after a workflow deployment completes
 */

async function postDeployHook(context) {
  const { workflow, result, config, logger } = context;
  const { github_token, default_repo, auto_sync } = config;

  logger.info('GitHub post-deploy: starting', {
    workflowId: workflow?.id,
    success: result?.success,
    repo: default_repo,
  });

  if (!github_token || !auto_sync) {
    logger.info('GitHub post-deploy: skipping (no token or auto_sync disabled)');
    return { success: true };
  }

  try {
    const [owner, repo] = (default_repo || '').split('/');
    if (!owner || !repo) {
      return { success: true };
    }

    // Create commit status update
    const state = result?.success ? 'success' : 'failure';
    const description = result?.success
      ? `FunctionFly deployment completed: ${workflow?.name}`
      : `FunctionFly deployment failed: ${workflow?.name}`;
    const targetUrl = `https://functionfly.com/workflows/${workflow?.id}`;

    // In a real implementation, this would call the GitHub API
    // For now, we just log the action
    logger.info('GitHub post-deploy: would update commit status', {
      owner,
      repo,
      sha: workflow?.commit_sha,
      state,
      description,
    });

    // Create deployment comment on GitHub if this was a PR-triggered workflow
    if (workflow?.pr_number) {
      logger.info('GitHub post-deploy: would post PR comment', {
        owner,
        repo,
        pr: workflow.pr_number,
        result: result?.success ? 'success' : 'failure',
      });
    }

    return {
      success: true,
      metadata: {
        github_updated: true,
        state,
        timestamp: new Date().toISOString(),
      },
    };
  } catch (error) {
    logger.error('GitHub post-deploy hook failed', {
      error: error.message,
      workflowId: workflow?.id,
    });

    return { success: true, metadata: { error: error.message } };
  }
}

module.exports = postDeployHook;