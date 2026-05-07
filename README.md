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
