# bootdev_httpserver

## About

A simplified Twitter API implementation built with Go, following a bootdev course project to learn HTTP webserver architecture and database utilization.

## Why This Project

A small-scale example demonstrating how to set up a webserver with multiple endpoints, database integration, user authentication, and webhook handling.

## Installation

1. Clone the repository
2. Install Go (1.16+)
3. Set up environment variables:
   - `DB_URL`: PostgreSQL connection string
   - `JWT_SECRET`: Secret key for JWT tokens
   - `POLKA_KEY`: API key for Polka webhooks
   - `PLATFORM`: Set to "dev" for development mode
4. Create the PostgreSQL database using the schema in `sql/`
5. Run `go run .` to start the server on port 8080

## Usage

### Endpoints

**Chirps (Twitter-like posts)**
- `POST /api/chirps` - Create a new chirp (requires auth)
- `GET /api/chirps` - Get all chirps (accepts optional query parameters author=ID and sort=desc)
- `GET /api/chirps/{chirpID}` - Get chirp by ID
- `DELETE /api/chirps/{chirpID}` - Delete chirp (requires auth)

**Users**
- `POST /api/users` - Create a new user
- `PUT /api/users` - Update user (requires auth)
- `POST /api/login` - Login and receive JWT token

**Authentication**
- `POST /api/refresh` - Refresh JWT token
- `POST /api/revoke` - Revoke token

**Admin**
- `GET /api/healthz` - Health check
- `GET /admin/metrics` - View server metrics
- `POST /admin/resetmetrics` - Reset metrics counters
- `POST /admin/reset` - Reset database (dev mode only, set via .env)

**Webhooks**
- `POST /api/polka/webhooks` - Polka (fake Stripe) payment webhook handler

**Static Files**
- `GET /app/*` - Serve static files from project root
