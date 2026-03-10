// Example FunctionFly function for testing watch mode
export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    // Health check endpoint
    if (url.pathname === '/healthz') {
      return new Response('OK', {
        status: 200,
        headers: { 'Content-Type': 'text/plain' }
      });
    }

    // Main application logic
    return new Response(`Hello from FunctionFly with Hot Reload!
Environment: ${env.NODE_ENV || 'development'}
Timestamp: ${new Date().toISOString()}
Method: ${request.method}
Path: ${url.pathname}
`, {
      status: 200,
      headers: {
        'Content-Type': 'text/plain',
        'X-FunctionFly-Version': '1.0.0'
      }
    });
  }
};