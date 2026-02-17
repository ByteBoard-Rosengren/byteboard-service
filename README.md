# ByteBoard Backend Service

### [Click Here to Checkout the Website](https://byteboard.win)

A RESTful API backend for ByteBoard - a social platform for developers to share posts, comments, and profiles. Built with Go, featuring JWT authentication, role-based authorization, and PostgreSQL.

## Features

- JWT authentication with HMAC-SHA512 signing
- Role-based authorization (admin/user roles)
- Bcrypt password hashing
- Automatic profile creation on registration
- CORS support
- Structured logging with Zerolog
- Middleware chain (recovery, logging, CORS, auth)
- PostgreSQL with cascading deletes

## Tech Stack

- **Language**: Go 1.25.0
- **Router**: [Gorilla Mux](https://github.com/gorilla/mux)
- **Database**: PostgreSQL with [lib/pq](https://github.com/lib/pq)
- **JWT**: [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
- **Password Hashing**: [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)
- **Logging**: [Zerolog](https://github.com/rs/zerolog)
- **Config**: [godotenv](https://github.com/joho/godotenv) + [caarlos0/env](https://github.com/caarlos0/env)

## Project Structure

```
byte-board-service/
├── cmd/server/
├───── main.go                   # Entry point & routing
├── internal/
│   ├── appconfig/               # Configuration
├──────── config.go
│   ├── auth/                    # JWT & password utilities
├──────── jwt.go
├──────── password.go
│   ├── handler/                 # HTTP handlers
├──────── auth.go
├──────── handlers.go
│   ├── middleware/              # Auth, CORS, logging, recovery
├──────── auth.go
├──────── cors.go
├──────── logging.go
├──────── recovery.go
│   ├── model/                   # Data models
├──────── errors.go
├──────── models.go
├──────── user.go
│   ├── repository/              # Database operations
├──────── database.go
│   └── service/                 # Business logic
├──────── auth_service.go
├── database.sql                 # Schema & seed data
├── .env                         # Environment variables
└── secrets/                     # Sensitive files
```

## Quick Start

### Prerequisites
- Go 1.25+
- PostgreSQL

### Setup

1. **Clone and install dependencies**
   ```bash
   SSH
   - git clone git@github.com:krosengr4/byte-board-service.git
   HTTPS
   - git clone https://github.com/krosengr4/byte-board-service.git

   cd byte-board-service
   go mod download
   ```

2. **Setup database**
   ```bash
   createdb byteboard_db
   psql -U your-user -d byteboard_db -f database.sql
   ```

3. **Configure environment** (create `.env` file)
   ```env
   PORT=8080
   POSTGRES_HOST=localhost
   POSTGRES_PORT=5432
   POSTGRES_DB=byteboard_db
   POSTGRES_USER=your-user
   POSTGRES_PASSWORD_FILE=postgres-password
   POSTGRES_SSL_MODE=disable
   JWT_SECRET=your-secret-key-change-in-production
   JWT_EXPIRATION_HOURS=30
   ALLOWED_ORIGINS=http://localhost:3000
   SECRETS_PATH=./secrets
   ```

4. **Create password file**
   ```bash
   mkdir -p secrets
   echo "your-postgres-password" > secrets/postgres-password
   ```

5. **Run the server**
   ```bash
   go run cmd/server/main.go
   ```

Server starts on `http://localhost:8080`

## API Overview

### Account registration and login
- `POST /api/register` - Create account
- `POST /api/login` - Get JWT token

### Public endpoints
- `GET /api/posts` - View posts
- `GET /api/comments` - View comments
- `GET /api/profiles` - View profiles

### Protected Endpoints (JWT required)
- `GET /api/auth/me` - Current user info
- `DELETE /api/auth/account` - Delete own account

### Admin Endpoints (JWT + admin role)
- `GET /api/admin/users` - View all users
- `GET /api/admin/users/{userId}` - Get user by ID
- `GET /api/admin/users/username/{username}` - Get user by username

### POST endpoints
- `POST /api/posts` - Create a post (As a Verified User)
- `POST /api/comments` - Create a comment (As a Verified User)

### PUT endpoints
- `PUT /api/post/{postId}` - Update your post
- `PUT /api/profiles` - Update your profile

## Usage Examples

### Register
```bash
curl -X POST http://localhost:8080/api/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "johndoe",
    "password": "password123"
  }'
```

### Access Protected Endpoint
```bash
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## Database Schema

- **users** - Authentication (username, hashed_password, role)
- **profiles** - User info (name, email, github, location)
- **posts** - User posts (title, content, author)
- **comments** - Post comments (content, author)

All tables use cascading deletes (delete user → deletes their profile, posts, comments).

## Security

- Passwords hashed with bcrypt (cost factor 10)
- JWT tokens signed with HMAC-SHA512
- Token expiration (default 30 hours)
- Role-based access control
- Hashed passwords never exposed in API responses
- Minimum password length: 8 characters

## Development

**Hot reload with Air:**
```bash
go install github.com/cosmtrek/air@latest
air
```

**Run tests:**
```bash
go test ./...
```

**Build for production:**
```bash
go build -o bin/server cmd/server/main.go
```

## Authentication Flow

1. User registers → Account + profile created (role: "user")
2. User logs in → Receives JWT token
3. Token includes username and role
4. Protected endpoints validate token via middleware
5. Admin endpoints also check role from token
6. Token expires after 30 hours (configurable)

**Important:** Role changes in database require re-login to get new token with updated role.

## Error Codes

- `400` - Bad request (missing fields, invalid input)
- `401` - Unauthorized (invalid credentials, missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `409` - Conflict (username already exists)
- `500` - Internal server error

## Roadmap

- [ ] POST/PUT/DELETE for posts and comments
- [ ] Profile update endpoint
- [ ] Password change functionality
- [ ] Token refresh mechanism
- [ ] Rate limiting
- [ ] API documentation (Swagger)
- [ ] Unit tests
- [ ] Docker support

## License
For more information about the license, please follow this link: [MIT License](https://opensource.org/license/mit/)

## Contributing

### Please contribute to this project:
- [Submit Bugs and Request Features you'd like to see Implemented](https://github.com/krosengr4/byte-board-service/issues)

## Questions
- [Link to my Github Profile](https://github.com/krosengr4)

- For any additional questions, email me at rosenkev4@gmail.com
