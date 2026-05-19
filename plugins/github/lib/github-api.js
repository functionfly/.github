/**
 * GitHub API Client
 * Simplified wrapper around GitHub REST API
 */

const https = require('https');
const http = require('http');

class GithubAPI {
  constructor(options = {}) {
    this.token = options.token;
    this.owner = options.owner;
    this.repo = options.repo;
    this.baseUrl = 'api.github.com';
  }

  async request(method, path, data = null) {
    const url = new URL(`https://${this.baseUrl}${path}`);
    const options = {
      hostname: url.hostname,
      port: 443,
      path: url.pathname + url.search,
      method,
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Accept': 'application/vnd.github+json',
        'X-GitHub-Api-Version': '2022-11-28',
        'User-Agent': 'FunctionFly-GitHub-Plugin/1.0',
        'Content-Type': 'application/json',
      },
    };

    return new Promise((resolve, reject) => {
      const req = https.request(options, (res) => {
        let body = '';
        res.on('data', chunk => body += chunk);
        res.on('end', () => {
          try {
            const parsed = body ? JSON.parse(body) : {};
            if (res.statusCode >= 200 && res.statusCode < 300) {
              resolve(parsed);
            } else {
              reject(new Error(`GitHub API error: ${res.statusCode} - ${parsed.message || body}`));
            }
          } catch (e) {
            reject(new Error(`Failed to parse response: ${e.message}`));
          }
        });
      });

      req.on('error', reject);

      if (data) {
        req.write(JSON.stringify(data));
      }

      req.end();
    });
  }

  async getWorkflowStatus(workflowId) {
    if (!this.token || !this.owner || !this.repo) {
      return null;
    }

    try {
      // Check if workflow file exists in the repo
      const contents = await this.request('GET', `/repos/${this.owner}/${this.repo}/contents/.github/workflows`);
      return contents;
    } catch (error) {
      // Workflows directory might not exist
      return null;
    }
  }

  async getRepositories() {
    return this.request('GET', '/user/repos?sort=updated&per_page=30');
  }

  async createIssue({ owner, repo, title, body, labels = [] }) {
    return this.request('POST', `/repos/${owner}/${repo}/issues`, {
      title,
      body,
      labels,
    });
  }

  async createCommitStatus({ owner, repo, sha, state, description, target_url }) {
    return this.request('POST', `/repos/${owner}/${repo}/statuses/${sha}`, {
      state,
      description: description.substring(0, 140),
      target_url,
    });
  }

  async getPullRequests({ owner, repo, state = 'open' }) {
    return this.request('GET', `/repos/${owner}/${repo}/pulls?state=${state}`);
  }

  async createPullRequestComment({ owner, repo, prNumber, body }) {
    return this.request('POST', `/repos/${owner}/${repo}/issues/${prNumber}/comments`, {
      body,
    });
  }

  async triggerWorkflow({ owner, repo, workflowId, ref, inputs = {} }) {
    return this.request('POST', `/repos/${owner}/${repo}/actions/workflows/${workflowId}/dispatches`, {
      ref,
      inputs,
    });
  }

  async getWorkflowRuns({ owner, repo, workflowId }) {
    return this.request('GET', `/repos/${owner}/${repo}/actions/workflows/${workflowId}/runs`);
  }
}

module.exports = GithubAPI;