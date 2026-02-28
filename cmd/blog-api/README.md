# Blog API Service

The Blog API is a standalone NestJS service that powers the FunctionFly blog and changelog. It runs as a **separate service** from the main Go API orchestrator.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    FunctionFly Services                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────┐    ┌──────────────────────────────┐   │
│  │  Go Orchestrator │    │  Blog API (NestJS)            │   │
│  │  :8080           │    │  :3001                        │   │
│  │                  │    │                               │   │
│  │  - Auth          │    │  - Blog posts (CRUD)          │   │
│  │  - Functions     │    │  - Authors                    │   │
│  │  - Registry      │    │  - Categories                 │   │
│  │  - State         │    │  - Drizzle ORM + PostgreSQL   │   │
│  └──────────────────┘    └──────────────────────────────┘   │
│           │                          │                       │
│           └──────────┬───────────────┘                      │
│                      ▼                                       │
│              ┌──────────────┐                               │
│              │  PostgreSQL  │                               │
│              │  (shared or  │                               │
│              │   separate)  │                               │
│              └──────────────┘                               │
└─────────────────────────────────────────────────────────────┘
```

## Quick Start

### Development

```bash
cd cmd/blog-api
cp .env.example .env
# Edit .env with your database credentials

bun install
bun run db:migrate
bun run start:dev
```

### Docker Compose (Recommended)

The blog API has its own Docker Compose profile. Use the `blog` profile:

```bash
# Start everything including blog API
docker compose --profile blog up

# Start only the blog API
docker compose --profile blog up blog-api

# Start main stack without blog API (default)
docker compose up
```

### Production Deployment

```bash
# Build the Docker image
docker build -t functionfly-blog-api ./cmd/blog-api

# Run with environment variables
docker run -p 3001:3001 \
  -e DATABASE_URL="postgresql://user:pass@host:5432/blog_db" \
  -e JWT_SECRET="your-jwt-secret" \
  functionfly-blog-api
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | ✅ | PostgreSQL connection string |
| `JWT_SECRET` | ✅ | JWT signing secret (must match main API) |
| `PORT` | ❌ | HTTP port (default: 3001) |
| `NODE_ENV` | ❌ | Environment (development/production) |

## Database

The blog API uses its own schema managed by [Drizzle ORM](https://orm.drizzle.team/).

```bash
# Generate a new migration
bun run db:generate

# Apply migrations
bun run db:migrate

# Open Drizzle Studio (visual DB browser)
bun run db:studio
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/blog/posts` | List published posts |
| `GET` | `/blog/posts/:slug` | Get post by slug |
| `POST` | `/blog/posts` | Create post (admin) |
| `PUT` | `/blog/posts/:id` | Update post (admin) |
| `DELETE` | `/blog/posts/:id` | Delete post (admin) |
| `GET` | `/blog/authors` | List authors |
| `GET` | `/blog/categories` | List categories |
| `GET` | `/health` | Health check |

## Seeding Default Content

```bash
bun run seed
```

## Integration with Dashboard

The React dashboard fetches blog content via `api/blog.ts`. Configure the API URL:

```env
# web/dashboard/.env
VITE_BLOG_API_URL=http://localhost:3001
```

## Operational Notes

- **Separate deployment**: The blog API can be deployed independently from the main Go orchestrator.
- **Shared JWT**: Uses the same `JWT_SECRET` as the main API for admin authentication.
- **Database isolation**: For production, consider a separate PostgreSQL database for the blog.
- **Health check**: Exposes `/health` for load balancer health checks.
