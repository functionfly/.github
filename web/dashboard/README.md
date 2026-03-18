# FunctionFly Dashboard

React + TypeScript + Vite dashboard for the FunctionFly serverless platform.

## Overview

The FunctionFly Dashboard is a single-page application (SPA) built with React, TypeScript, and Vite. It provides a web interface for managing functions, viewing analytics, and configuring platform settings.

## Tech Stack

- **Framework**: React 18+ with TypeScript
- **Build Tool**: Vite
- **Routing**: React Router v6
- **State Management**: React Context + hooks
- **Styling**: CSS Modules / CSS Variables
- **HTTP Client**: Fetch API
- **Package Manager**: bun

## Getting Started

### Prerequisites

- Node.js 18+
- bun package manager

### Installation

```bash
# Install dependencies
bun install

# Or from workspace root
cd ../..
bun install
```

### Development

```bash
# Start development server
bun run dev

# Or with custom API URL
VITE_API_URL=http://localhost:8080 bun run dev
```

The dashboard will be available at `http://localhost:3000`. The dev server proxies API requests to `http://localhost:8080`.

### Production Build

```bash
# Build for production
bun run build

# Preview production build
bun run preview
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_URL` | API base URL | `http://localhost:8080` |
| `VITE_WS_URL` | WebSocket URL | `ws://localhost:8080` |

## Project Structure

```
src/
├── pages/           # Page components
│   ├── Dashboard.tsx
│   ├── Functions.tsx
│   ├── FunctionDetail.tsx
│   ├── Deployments.tsx
│   ├── Analytics.tsx
│   ├── Settings.tsx
│   └── Login.tsx
├── components/      # Reusable components
│   ├── Header.tsx
│   ├── Sidebar.tsx
│   ├── FunctionCard.tsx
│   ├── LogViewer.tsx
│   └── ...
├── hooks/           # Custom React hooks
│   ├── useAuth.ts
│   ├── useFunctions.ts
│   └── useApi.ts
├── lib/             # Utilities and helpers
│   ├── api.ts
│   ├── auth.ts
│   └── utils.ts
├── contexts/        # React contexts
│   ├── AuthContext.tsx
│   └── ThemeContext.tsx
├── types/           # TypeScript type definitions
└── App.tsx          # Main application component
```

## Features

### Function Management
- Create, edit, and delete functions
- Deploy functions to the registry
- View function versions and history
- Monitor function performance

### Analytics
- Usage statistics and metrics
- Execution history and logs
- Performance graphs

### User Settings
- Profile management
- API key management
- Notification preferences

### Authentication
- OAuth (GitHub, Google)
- Email/password login
- Session management

## API Integration

The dashboard communicates with the FunctionFly API at `VITE_API_URL`. All API endpoints are documented in [`docs/API.md`](../../docs/API.md).

### Authentication Flow

1. User logs in via OAuth or email/password
2. API returns JWT access token and refresh token
3. Tokens are stored in localStorage
4. Tokens are included in Authorization header for API requests

## Testing

```bash
# Run unit tests
bun run test

# Run tests with coverage
bun run test:coverage

# Run tests in watch mode
bun run test:watch
```

## Linting and Formatting

```bash
# Run ESLint
bun run lint

# Format code with Prettier
bun run format
```

## Troubleshooting

### Dashboard can't connect to API

1. Check `VITE_API_URL` environment variable
2. Verify the API server is running on port 8080
3. Check browser console for CORS errors

### Login issues

1. Clear localStorage and sessionStorage
2. Check browser console for authentication errors
3. Verify OAuth credentials are configured

### Build failures

1. Clear node_modules and reinstall
2. Check for TypeScript errors: `bun run typecheck`

## Contributing

See [`CONTRIBUTING.md`](../../CONTRIBUTING.md) for contribution guidelines.

## License

MIT License - see [`LICENSE`](../../LICENSE) for details.

## Support

- GitHub Issues: https://github.com/functionfly/functionfly/issues
- Discord: https://discord.gg/functionfly
