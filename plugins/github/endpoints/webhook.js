/**
 * GitHub Webhook Handler
 */

const crypto = require('crypto');

class WebhookHandler {
  constructor(options = {}) {
    this.secret = options.secret;
    this.github = options.github;
  }

  verifySignature(payload, headers) {
    if (!this.secret) {
      return true; // Skip verification if no secret configured
    }

    const signature = headers['x-hub-signature-256'] || headers['X-Hub-Signature-256'];
    if (!signature) {
      return false;
    }

    const expectedSig = 'sha256=' + crypto
      .createHmac('sha256', this.secret)
      .update(JSON.stringify(payload))
      .digest('hex');

    return crypto.timingSafeEqual(
      Buffer.from(signature),
      Buffer.from(expectedSig)
    );
  }

  async handle(payload, headers) {
    const event = headers['x-github-event'] || headers['X-GitHub-Event'];

    // Verify webhook signature
    if (!this.verifySignature(payload, headers)) {
      return {
        error: 'Invalid signature',
        status: 401,
      };
    }

    switch (event) {
      case 'push':
        return this.handlePush(payload);
      case 'pull_request':
        return this.handlePullRequest(payload);
      case 'workflow_run':
        return this.handleWorkflowRun(payload);
      case 'workflow_job':
        return this.handleWorkflowJob(payload);
      default:
        return { handled: true, event };
    }
  }

  async handlePush(payload) {
    const { repository, ref, commits, pusher } = payload;

    return {
      handled: true,
      event: 'push',
      data: {
        repo: repository?.full_name,
        ref,
        commits_count: commits?.length || 0,
        pusher: pusher?.name,
      },
    };
  }

  async handlePullRequest(payload) {
    const { action, pull_request, repository } = payload;

    return {
      handled: true,
      event: 'pull_request',
      action,
      data: {
        repo: repository?.full_name,
        pr_number: pull_request?.number,
        pr_title: pull_request?.title,
        pr_state: pull_request?.state,
        merged: pull_request?.merged,
      },
    };
  }

  async handleWorkflowRun(payload) {
    const { action, workflow_run, repository } = payload;

    return {
      handled: true,
      event: 'workflow_run',
      action,
      data: {
        repo: repository?.full_name,
        workflow_name: workflow_run?.name,
        run_id: workflow_run?.id,
        conclusion: workflow_run?.conclusion,
        status: workflow_run?.status,
      },
    };
  }

  async handleWorkflowJob(payload) {
    const { action, workflow_job, repository } = payload;

    return {
      handled: true,
      event: 'workflow_job',
      action,
      data: {
        repo: repository?.full_name,
        job_name: workflow_job?.name,
        job_status: workflow_job?.status,
        conclusion: workflow_job?.conclusion,
      },
    };
  }
}

module.exports = WebhookHandler;