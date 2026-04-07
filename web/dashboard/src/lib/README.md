# FunctionFly Authentication Client

Complete client-side authentication system with JWT, CSRF, and HMAC signing support.

## Features

- ✅ JWT token management with automatic refresh
- ✅ CSRF token handling for mutations
- ✅ HMAC signing for admin operations (server-side only)
- ✅ React hooks for easy integration
- ✅ Automatic token storage and validation

## Setup

### 1. Install Dependencies

```bash
npm install crypto-js  # Required for HMAC signing in Node.js environments
```

### 2. Configure API Secret

For HMAC signing, set the API shared secret:

```typescript
// In your app configuration
import { auth } from './lib/auth';

// Configure with API shared secret (from admin)
auth.setSharedSecret('your-api-shared-secret-here');
```

## Usage

### Basic Authentication

```typescript
import { auth } from './lib/auth';

// Login
try {
  const tokens = await auth.login('admin@example.com', 'password');
  console.log('Logged in:', tokens);
} catch (error) {
  console.error('Login failed:', error);
}

// Make authenticated requests
const response = await auth.get('/users/me');
const user = await response.json();
```

### React Hook Usage

```typescript
import { useAuth, useAdminAPI } from './hooks/useAuth';

function AdminDashboard() {
  const { user, isAuthenticated, login, logout } = useAuth();
  const { adminRequest, getSignupInvites, createSignupInvite } = useAdminAPI();

  if (!isAuthenticated) {
    return <LoginForm onLogin={login} />;
  }

  const handleCreateInvite = async () => {
    try {
      const invite = await createSignupInvite({
        label: 'New User Invite',
        max_uses: 5
      });
      console.log('Created invite:', invite);
    } catch (error) {
      console.error('Failed to create invite:', error);
    }
  };

  return (
    <div>
      <h1>Welcome, {user?.username}!</h1>
      <button onClick={handleCreateInvite}>Create Signup Invite</button>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
```

### Manual Request Signing

```typescript
import { auth } from './lib/auth';

// Prepare a signed request
const signedRequest = await auth.prepareRequest({
  method: 'POST',
  path: '/v1/admin/signup-invites',
  body: { label: 'Test Invite', max_uses: 10 }
});

// Make the request
const response = await fetch(`${auth.baseURL}${signedRequest.path}`, {
  method: signedRequest.method,
  headers: signedRequest.headers,
  body: JSON.stringify(signedRequest.body)
});
```

## API Endpoints

### Authentication
- `POST /auth/login` - Login with email/password
- `POST /auth/refresh` - Refresh JWT token
- `GET /v1/admin/csrf` - Get CSRF token

### Admin Operations (require authentication + CSRF + HMAC)
- `GET /v1/admin/signup-invites` - List signup invites
- `POST /v1/admin/signup-invites` - Create signup invite
- `POST /v1/admin/signup-invites/{id}/revoke` - Revoke invite

## Security Features

1. **JWT Tokens**: Short-lived access tokens with refresh capability
2. **CSRF Protection**: Double-submit pattern with token validation
3. **HMAC Signing**: Request signing for admin operations (server-side)
4. **Automatic Refresh**: Seamless token renewal on expiry
5. **Secure Storage**: Local storage with plans for secure alternatives

## Configuration

### Environment Variables
```bash
# API Base URL
VITE_API_URL=https://api.functionfly.com

# API Shared Secret (for HMAC signing - server-side only)
FFLY_SHARED_SECRET=your-secret-here
```

### Browser Support
- HMAC signing requires Node.js environment or crypto-js polyfill
- Automatic fallback to unsigned requests in browser environments
- CSRF tokens are always included for mutations

## Error Handling

```typescript
try {
  const result = await adminAPI.createSignupInvite(data);
} catch (error) {
  if (error.message.includes('Session expired')) {
    // Redirect to login
    navigate('/login');
  } else if (error.message.includes('Invalid signature')) {
    // HMAC signing failed - check API_SHARED_SECRET
    console.error('HMAC signing failed');
  } else {
    // Other error
    console.error('Request failed:', error);
  }
}
```

## Troubleshooting

### 403 Forbidden on POST requests
- Check that CSRF token is fresh (expires every hour)
- Verify HMAC signing is configured (requires API_SHARED_SECRET)
- Ensure JWT token is valid and not expired

### 401 Unauthorized
- JWT token expired - automatic refresh should handle this
- Check login credentials
- Verify admin session creation (server-side)

### HMAC Signing Issues
- HMAC signing only works in Node.js environments
- Requires `API_SHARED_SECRET` environment variable
- Browser environments skip HMAC signing automatically