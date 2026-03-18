/**
 * Hello World TypeScript Function
 *
 * A basic example demonstrating the FunctionFly TypeScript handler format.
 * This function accepts a name and returns a personalized greeting.
 *
 * @module hello-world
 */

// Import the FunctionFly types (local type definitions for this example)
import type {
  Context,
  Env,
  Handler,
  Request,
  Response,
} from "./types/functionfly";

/**
 * Input interface for the hello-world function
 */
interface HelloWorldInput {
  name?: string;
}

/**
 * Output interface for the hello-world function
 */
interface HelloWorldOutput {
  message: string;
}

/**
 * Main handler function for the hello-world function.
 *
 * This function demonstrates:
 * - Accessing request body data
 * - Using environment variables
 * - Returning a typed JSON response
 *
 * @param request - The incoming HTTP request
 * @param env - Environment variables
 * @param context - Function execution context
 * @returns Response object with greeting message
 */
const handler: Handler = async (
  request: Request,
  env: Env,
  _context: Context,
): Promise<Response> => {
  // Parse the request body
  let name = "World";

  try {
    const body = await request.json<HelloWorldInput>();
    if (body.name) {
      name = body.name;
    }
  } catch {
    // If body is not valid JSON, try to get it as text
    try {
      const text = await request.text();
      if (text) {
        name = text;
      }
    } catch {
      // Use default name
    }
  }

  // Get greeting from environment variable with fallback
  const greeting = env.GREETING ?? "Hello";

  // Create the response
  const output: HelloWorldOutput = {
    message: `${greeting}, ${name}!`,
  };

  return {
    status: 200,
    headers: {
      "Content-Type": "application/json",
      "X-Greeting-From": "FunctionFly TypeScript",
    },
    body: output,
  };
};

// Export the handler
export { handler };

// Also support the default export for easier importing
export default handler;
