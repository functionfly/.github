/**
 * Default blog post: Getting Started Tutorial
 * Practical guide for new FunctionFly users
 */
import { ContentStatus } from "../../modules/blog/dto/blog-post.dto";

export const slug = "getting-started-deploy-your-first-functionfly-function";

const body = [
  {
    type: "paragraph",
    children: [
      {
        text: "Ready to start building on FunctionFly? This tutorial will walk you through deploying your first serverless function in under 10 minutes. We'll build a simple API endpoint that processes text and returns AI-powered insights.",
      },
    ],
  },
  {
    type: "heading",
    level: 2,
    children: [{ text: "Prerequisites" }],
  },
  {
    type: "paragraph",
    children: [{ text: "Before we start, make sure you have:" }],
  },
  {
    type: "paragraph",
    children: [
      { text: "• A FunctionFly account (sign up at functionfly.com)" },
    ],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "• Go 1.25+ (for go install) or a downloaded fly binary from GitHub Releases",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "• Basic familiarity with JavaScript/TypeScript" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "Don't worry if you're new to serverless—we'll explain everything step by step!",
      },
    ],
  },
  {
    type: "heading",
    level: 2,
    children: [{ text: "Step 1: Install the FunctionFly CLI" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "The FunctionFly CLI (`fly`) ships as a Go binary. Install with go install or download a release archive from GitHub:",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```bash" }],
  },
  {
    type: "paragraph",
    children: [
      { text: "go install github.com/functionfly/functionfly/cmd/fly@latest" },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "paragraph",
    children: [{ text: "Verify the installation worked:" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```bash" }],
  },
  {
    type: "paragraph",
    children: [{ text: "fly --version" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "heading",
    level: 2,
    children: [{ text: "Step 2: Authenticate with FunctionFly" }],
  },
  {
    type: "paragraph",
    children: [{ text: "Login to your FunctionFly account:" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```bash" }],
  },
  {
    type: "paragraph",
    children: [{ text: "fly auth login" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "This will open your browser for authentication. Once you're logged in, the CLI will be ready to use.",
      },
    ],
  },
  {
    type: "heading",
    level: 2,
    children: [{ text: "Step 3: Create Your First Function" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "Let's create a new function that analyzes text sentiment. Create a new directory and initialize a FunctionFly project:",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```bash" }],
  },
  {
    type: "paragraph",
    children: [{ text: "mkdir sentiment-analyzer && cd sentiment-analyzer" }],
  },
  {
    type: "paragraph",
    children: [{ text: "fly init" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: 'Choose "JavaScript" as your language and "HTTP API" as the template. This creates a basic function structure.',
      },
    ],
  },
  {
    type: "heading",
    level: 2,
    children: [{ text: "Step 4: Implement the Function Logic" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "Open the generated `index.js` file and replace its contents with our sentiment analyzer:",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```javascript" }],
  },
  {
    type: "paragraph",
    children: [{ text: "// Sentiment analysis function" }],
  },
  {
    type: "paragraph",
    children: [
      { text: "export default async function handler(request, context) {" },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "  // Parse the request body" }],
  },
  {
    type: "paragraph",
    children: [{ text: "  const { text } = await request.json();" }],
  },
  {
    type: "paragraph",
    children: [{ text: "  " }],
  },
  {
    type: "paragraph",
    children: [{ text: "  // Validate input" }],
  },
  {
    type: "paragraph",
    children: [{ text: "  if (!text || typeof text !== 'string') {" }],
  },
  {
    type: "paragraph",
    children: [{ text: "    return new Response(JSON.stringify({" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "      error: 'Please provide a \"text\" field in the request body'",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "    }), {" }],
  },
  {
    type: "paragraph",
    children: [{ text: "      status: 400," }],
  },
  {
    type: "paragraph",
    children: [
      { text: "      headers: { 'Content-Type': 'application/json' }" },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "    });" }],
  },
  {
    type: "paragraph",
    children: [{ text: "  }" }],
  },
  {
    type: "paragraph",
    children: [{ text: "  " }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "  // Simple sentiment analysis (you could integrate with OpenAI, Anthropic, etc.)",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "  const sentiment = analyzeSentiment(text);" }],
  },
  {
    type: "paragraph",
    children: [{ text: "  " }],
  },
  {
    type: "paragraph",
    children: [{ text: "  // Return the analysis" }],
  },
  {
    type: "paragraph",
    children: [{ text: "  return new Response(JSON.stringify({" }],
  },
  {
    type: "paragraph",
    children: [{ text: "    text: text," }],
  },
  {
    type: "paragraph",
    children: [{ text: "    sentiment: sentiment," }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "    confidence: Math.random() * 0.3 + 0.7 // Mock confidence score",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "  }), {" }],
  },
  {
    type: "paragraph",
    children: [{ text: "    headers: { 'Content-Type': 'application/json' }" }],
  },
  {
    type: "paragraph",
    children: [{ text: "  });" }],
  },
  {
    type: "paragraph",
    children: [{ text: "}" }],
  },
  {
    type: "paragraph",
    children: [{ text: "" }],
  },
  {
    type: "paragraph",
    children: [{ text: "// Simple sentiment analysis function" }],
  },
  {
    type: "paragraph",
    children: [{ text: "function analyzeSentiment(text) {" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "  const positiveWords = ['good', 'great', 'excellent', 'amazing', 'wonderful', 'fantastic'];",
      },
    ],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "  const negativeWords = ['bad', 'terrible', 'awful', 'horrible', 'disappointing', 'worst'];",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "  " }],
  },
  {
    type: "paragraph",
    children: [{ text: "  const lowerText = text.toLowerCase();" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "  const positiveCount = positiveWords.filter(word => lowerText.includes(word)).length;",
      },
    ],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "  const negativeCount = negativeWords.filter(word => lowerText.includes(word)).length;",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "  " }],
  },
  {
    type: "paragraph",
    children: [
      { text: "  if (positiveCount > negativeCount) return 'positive';" },
    ],
  },
  {
    type: "paragraph",
    children: [
      { text: "  if (negativeCount > positiveCount) return 'negative';" },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "  return 'neutral';" }],
  },
  {
    type: "paragraph",
    children: [{ text: "}" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "heading",
    level: 2,
    children: [{ text: "Step 5: Test Locally" }],
  },
  {
    type: "paragraph",
    children: [{ text: "Before deploying, let's test our function locally:" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```bash" }],
  },
  {
    type: "paragraph",
    children: [{ text: "fly dev" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "This starts a local development server. In another terminal, test your function:",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```bash" }],
  },
  {
    type: "paragraph",
    children: [{ text: "curl -X POST http://localhost:3000 \\" }],
  },
  {
    type: "paragraph",
    children: [{ text: '  -H "Content-Type: application/json" \\' }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: '  -d \'{"text": "This product is absolutely amazing and wonderful!"}\'',
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "paragraph",
    children: [{ text: "You should see a response like:" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```json" }],
  },
  {
    type: "paragraph",
    children: [{ text: "{" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: '  "text": "This product is absolutely amazing and wonderful!",',
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: '  "sentiment": "positive",' }],
  },
  {
    type: "paragraph",
    children: [{ text: '  "confidence": 0.85' }],
  },
  {
    type: "paragraph",
    children: [{ text: "}" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "heading",
    level: 2,
    children: [{ text: "Step 6: Deploy to Production" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "Once you're happy with your function, deploy it to FunctionFly:",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```bash" }],
  },
  {
    type: "paragraph",
    children: [{ text: "fly deploy" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "paragraph",
    children: [{ text: "This will:" }],
  },
  {
    type: "paragraph",
    children: [{ text: "1. Build your function into a deterministic capsule" }],
  },
  {
    type: "paragraph",
    children: [{ text: "2. Upload it to the FunctionFly registry" }],
  },
  {
    type: "paragraph",
    children: [{ text: "3. Generate a unique URL for your function" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "When deployment completes, you'll see a URL like `https://fx.run/your-username/sentiment-analyzer`.",
      },
    ],
  },
  {
    type: "heading",
    level: 2,
    children: [{ text: "Step 7: Test Your Live Function" }],
  },
  {
    type: "paragraph",
    children: [
      { text: "Test your deployed function using the production URL:" },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```bash" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "curl -X POST https://fx.run/your-username/sentiment-analyzer \\",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: '  -H "Content-Type: application/json" \\' }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: '  -d \'{"text": "This is the worst product I have ever used"}\'',
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "paragraph",
    children: [{ text: "Your function is now live and accessible worldwide!" }],
  },
  {
    type: "heading",
    level: 2,
    children: [{ text: "Step 8: Explore Advanced Features" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "Now that you have a working function, let's explore some of FunctionFly's powerful features:",
      },
    ],
  },
  {
    type: "heading",
    level: 3,
    children: [{ text: "Add State with State Fabric" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "Make your function stateful by storing data between invocations:",
      },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```javascript" }],
  },
  {
    type: "paragraph",
    children: [{ text: "// Store analysis history" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "const history = await context.state.get('analysis-history') || [];",
      },
    ],
  },
  {
    type: "paragraph",
    children: [
      { text: "history.push({ text, sentiment, timestamp: Date.now() });" },
    ],
  },
  {
    type: "paragraph",
    children: [
      { text: "await context.state.set('analysis-history', history);" },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "heading",
    level: 3,
    children: [{ text: "Secure Secrets with Secrets Vault" }],
  },
  {
    type: "paragraph",
    children: [{ text: "Store API keys and sensitive data securely:" }],
  },
  {
    type: "paragraph",
    children: [{ text: "```javascript" }],
  },
  {
    type: "paragraph",
    children: [{ text: "// Access secrets securely" }],
  },
  {
    type: "paragraph",
    children: [
      { text: "const apiKey = await context.secrets.get('openai-api-key');" },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "// Use the key for AI-powered analysis" }],
  },
  {
    type: "paragraph",
    children: [
      { text: "const aiAnalysis = await analyzeWithAI(text, apiKey);" },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "heading",
    level: 3,
    children: [{ text: "Participate in Flywheel Network" }],
  },
  {
    type: "paragraph",
    children: [
      { text: "Share your function in the community and get feedback:" },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```bash" }],
  },
  {
    type: "paragraph",
    children: [
      { text: 'fly share --description "Simple sentiment analysis API"' },
    ],
  },
  {
    type: "paragraph",
    children: [{ text: "```" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "This creates a Flywheel thread where other developers and AI agents can discuss, improve, and fork your function.",
      },
    ],
  },
  {
    type: "heading",
    level: 2,
    children: [{ text: "Next Steps" }],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "Congratulations! You've deployed your first FunctionFly function. Here are some ideas for what to build next:",
      },
    ],
  },
  {
    type: "paragraph",
    children: [
      { text: "• **Image Processing API**: Analyze images with AI models" },
    ],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "• **Data Transformation Service**: Convert between different data formats",
      },
    ],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "• **Webhook Handler**: Process webhooks from third-party services",
      },
    ],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "• **AI Agent Endpoint**: Create an API that AI agents can call",
      },
    ],
  },
  {
    type: "paragraph",
    children: [
      {
        text: "Check out our documentation for more examples and advanced patterns. Happy building!",
      },
    ],
  },
];

export const tutorialPost = {
  title: "Getting Started: Deploy Your First FunctionFly Function",
  slug,
  description:
    "Learn to deploy your first serverless function on FunctionFly in under 10 minutes. Build a sentiment analysis API with step-by-step guidance.",
  body,
  tags: [
    "tutorial",
    "getting-started",
    "javascript",
    "serverless",
    "api",
    "beginners",
    "functionfly",
  ],
  status: ContentStatus.PUBLISHED,
  publishedAt: new Date().toISOString(),
  seoTitle: "Getting Started with FunctionFly | First Function Tutorial",
  seoDescription:
    "Deploy your first serverless function on FunctionFly in under 10 minutes. Step-by-step tutorial for building a sentiment analysis API.",
  keywords: [
    "functionfly tutorial",
    "serverless functions",
    "getting started",
    "javascript",
    "API development",
    "sentiment analysis",
  ],
  canonicalUrl:
    "https://functionfly.com/blog/getting-started-deploy-your-first-functionfly-function",
} as const;

export type TutorialPostPayload = typeof tutorialPost;
