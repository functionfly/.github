/**
 * GitHub API Endpoint Handlers
 */

async function syncReposHandler(github) {
  try {
    const repos = await github.getRepositories();

    return {
      success: true,
      data: {
        repos: repos.map(r => ({
          id: r.id,
          name: r.name,
          full_name: r.full_name,
          private: r.private,
          default_branch: r.default_branch,
          updated_at: r.updated_at,
        })),
        count: repos.length,
      },
    };
  } catch (error) {
    return {
      success: false,
      error: error.message,
    };
  }
}

async function createIssueHandler(github, data) {
  const { owner, repo, title, body, labels } = data;

  if (!title) {
    return {
      success: false,
      error: 'Title is required',
    };
  }

  try {
    const issue = await github.createIssue({ owner, repo, title, body, labels });

    return {
      success: true,
      data: {
        id: issue.id,
        number: issue.number,
        title: issue.title,
        url: issue.html_url,
      },
    };
  } catch (error) {
    return {
      success: false,
      error: error.message,
    };
  }
}

async function createPullRequestCommentHandler(github, data) {
  const { owner, repo, prNumber, body } = data;

  if (!prNumber || !body) {
    return {
      success: false,
      error: 'PR number and body are required',
    };
  }

  try {
    const comment = await github.createPullRequestComment({ owner, repo, prNumber, body });

    return {
      success: true,
      data: {
        id: comment.id,
        body: comment.body,
      },
    };
  } catch (error) {
    return {
      success: false,
      error: error.message,
    };
  }
}

module.exports = {
  syncReposHandler,
  createIssueHandler,
  createPullRequestCommentHandler,
};