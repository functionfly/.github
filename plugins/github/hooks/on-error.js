/**
 * GitHub On-Error Hook
 * Runs when a workflow encounters an error
 */

async function onErrorHook(context) {
  const { workflow, error, config, logger } = context;
  const { github_token, default_repo, auto_sync } = config;

  logger.error('GitHub on-error: starting', {
    workflowId: workflow?.id,
    error: error?.message,
    repo: default_repo,
  });

  if (!github_token || !auto_sync) {
    logger.info('GitHub on-error: skipping (no token or auto_sync disabled)');
    return { handled: true };
  }

  try {
    const [owner, repo] = (default_repo || '').split('/');
    if (!owner || !repo) {
      return { handled: true };
    }

    // Create GitHub issue for the error
    const issueTitle = `[FunctionFly] Workflow Failed: ${workflow?.name || 'Unknown'}`;
    const issueBody = [
      `## Workflow Error Report`,
      ``,
      `**Workflow**: ${workflow?.name || 'Unknown'}`,
      `**ID**: ${workflow?.id || 'N/A'}`,
      `**Time**: ${new Date().toISOString()}`,
      ``,
      `### Error Details`,
      ``,
      `\`\`\``,
      error?.message || 'Unknown error',
      `\`\`\``,
      ``,
      `### Context`,
      ``,
      `- Workflow Version: ${workflow?.version || 'N/A'}`,
      `- Triggered By: ${workflow?.triggered_by || 'Unknown'}`,
      ``,
      `[View in FunctionFly](https://functionfly.com/workflows/${workflow?.id}) | [Contact Support](https://functionfly.com/support)`,
    ].join('\n');

    const labels = ['functionfly', 'automated', 'bug'];

    logger.info('GitHub on-error: would create issue', {
      owner,
      repo,
      title: issueTitle,
      labels,
    });

    // Also update the commit status to error
    logger.info('GitHub on-error: would update commit status to error', {
      owner,
      repo,
      sha: workflow?.commit_sha,
    });

    return {
      handled: true,
      metadata: {
        github_issue_created: false, // Would be true in real implementation
        github_status_updated: false,
        timestamp: new Date().toISOString(),
      },
    };
  } catch (err) {
    logger.error('GitHub on-error hook failed', {
      error: err.message,
      workflowId: workflow?.id,
    });

    // Still return handled: true so the error doesn't propagate
    return { handled: true, metadata: { hook_error: err.message } };
  }
}

module.exports = onErrorHook;