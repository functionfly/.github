/**
 * FunctionFly GitHub Integration Plugin
 * Main entry point
 */

const GithubAPI = require('./lib/github-api');
const WebhookHandler = require('./endpoints/webhook');
const { createIssueHandler, syncReposHandler } = require('./endpoints/handlers');

class GitHubPlugin {
  constructor(config = {}) {
    this.config = config;
    this.github = null;
    this.hooks = {};
  }

  async initialize(context) {
    const { tenantId, services } = context;

    // Initialize GitHub API client
    this.github = new GithubAPI({
      token: this.config.github_token,
      owner: this.config.default_repo?.split('/')[0],
      repo: this.config.default_repo?.split('/')[1],
    });

    // Register hooks
    this.hooks = {
      pre_deploy: this.handlePreDeploy.bind(this),
      post_deploy: this.handlePostDeploy.bind(this),
      on_error: this.handleError.bind(this),
    };

    // Setup webhook endpoint
    this.webhookHandler = new WebhookHandler({
      secret: this.config.webhook_secret,
      github: this.github,
    });

    context.logger.info('GitHub plugin initialized', {
      tenantId,
      defaultRepo: this.config.default_repo,
    });

    return { success: true };
  }

  async handlePreDeploy(context) {
    const { workflow, tenantId } = context;

    context.logger.info('GitHub pre-deploy hook', {
      workflowId: workflow?.id,
      tenantId,
    });

    // Check if GitHub Actions workflow should be triggered
    const workflowFile = await this.github.getWorkflowStatus(workflow?.id);
    if (workflowFile) {
      return {
        shouldContinue: true,
        metadata: {
          github_workflow: workflowFile.name,
          trigger_type: 'pre_deploy',
        },
      };
    }

    return { shouldContinue: true };
  }

  async handlePostDeploy(context) {
    const { workflow, result, tenantId } = context;

    context.logger.info('GitHub post-deploy hook', {
      workflowId: workflow?.id,
      tenantId,
      success: result?.success,
    });

    // Update GitHub status if configured
    if (this.config.auto_sync && result?.success) {
      await this.github.createCommitStatus({
        owner: this.config.default_repo?.split('/')[0],
        repo: this.config.default_repo?.split('/')[1],
        sha: workflow?.commit_sha,
        state: 'success',
        description: `FunctionFly deployment completed: ${workflow?.name}`,
        target_url: `https://functionfly.com/workflows/${workflow?.id}`,
      });
    }

    return { success: true };
  }

  async handleError(context) {
    const { workflow, error, tenantId } = context;

    context.logger.error('GitHub error hook', {
      workflowId: workflow?.id,
      tenantId,
      error: error?.message,
    });

    // Create GitHub issue on workflow failure
    if (this.config.auto_sync && error) {
      try {
        await this.github.createIssue({
          owner: this.config.default_repo?.split('/')[0],
          repo: this.config.default_repo?.split('/')[1],
          title: `[FunctionFly] Workflow Failed: ${workflow?.name}`,
          body: `## Workflow Error\n\n**Workflow**: ${workflow?.name}\n**ID**: ${workflow?.id}\n**Error**: ${error.message}\n\n[View in FunctionFly](https://functionfly.com/workflows/${workflow?.id})`,
          labels: ['functionfly', 'automated'],
        });
      } catch (err) {
        context.logger.error('Failed to create GitHub issue', { error: err.message });
      }
    }

    return { handled: true };
  }

  async handleWebhook(payload, headers) {
    return this.webhookHandler.handle(payload, headers);
  }

  async syncRepos() {
    return syncReposHandler(this.github);
  }

  async createIssue(data) {
    return createIssueHandler(this.github, data);
  }

  // Cleanup on shutdown
  async destroy() {
    this.github = null;
    this.hooks = {};
  }
}

module.exports = GitHubPlugin;