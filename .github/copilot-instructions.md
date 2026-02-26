# Fibra-Ctrl: AI Agent Instructions

## Project Overview

Fibra-Ctrl is an OLT (Optical Line Terminal) management system designed to interface with various fiber optic network equipment. The system has a Go backend API and React frontend, both containerized with Docker.

## Technology Stack & Versions

- **Go**: 1.23 (see backend/Dockerfile)
- **React**: 19.1.1 (see frontend/package.json)
- **TypeScript**: ~5.8.3 (see frontend/package.json)
- **Vite**: 7.1.2
- **TailwindCSS**: 3.4.1
- **Node.js**: 20-alpine (Docker base image)
- **PostgreSQL**: 15
- **JWT**: github.com/golang-jwt/jwt/v5
- **Gin Framework**: github.com/gin-gonic/gin

## Architecture

### Backend (Go)

- **API Layer**: RESTful service using Gin framework (`internal/api/server.go`)
- **Domain Model**: OLT management abstractions (`internal/domain/interfaces.go`)
- **Command System**: Asynchronous processing of OLT commands via queue/worker pattern
  - Commands are sent to queue (`internal/queue/queue.go`)
  - Processed by worker pool (`internal/worker/pool.go`)
- **Connection Management**: Connection pooling for OLT devices (`internal/connection/pool.go`)
- **Device Support**: Abstraction for different OLT vendors (`internal/olt/`)
  - Vendor-specific implementations (e.g., `internal/olt/nokia/fx16.go`)
- **Authentication**: JWT-based auth system (`internal/api/handlers/auth.go`)

### Frontend (React/TypeScript)

- Vite-based React SPA with TypeScript
- TailwindCSS for styling
- React Router for navigation
- Simple interface for user management and OLT operations

## Development Workflow

### Running the App

```bash
# Development mode with hot-reloading
docker-compose up

# Production mode
docker-compose -f docker-compose.prod.yml up
```

### Backend Development Patterns

- Command abstraction: OLT commands are defined as functions and mapped to vendor-specific implementations
- Repositories handle data persistence (`internal/repository/`)
- Workers process commands asynchronously to avoid blocking API requests
- Connection pooling manages OLT connections efficiently

### Adding New OLT Support

1. Create vendor-specific implementation in `internal/olt/{vendor}/`
2. Implement the `OLT` interface from `internal/olt/olt.go`
3. Register the handler in the worker system
4. Add function definitions to the repository

## Key Integration Points

- Backend-Frontend: REST API communication via `http://localhost:8080`
- Backend-OLT: SSH/Telnet connections to network devices (`internal/connection/`)
- Command Execution Flow:
  1. API receives command request (`/api/command`)
  2. Command added to queue
  3. Worker pool executes command on OLT device
  4. Result stored and made available via status endpoint (`/api/command/{id}`)

## Project Conventions

- Repository pattern for data access
- Interface-based abstractions (see `internal/domain/interfaces.go`)
- JWT for authentication
- Command pattern for OLT interactions
- Environment variables for configuration

## Common Tasks

- Adding new OLT vendor: Implement `OLT` interface in new package
- Adding command functionality: Define function in repository and implement parsing
- API endpoints: Add handlers to appropriate files in `internal/api/handlers/`
- Frontend views: Add to `frontend/src/pages/` and update router in `App.tsx`

## Configuration

- Database connection in docker-compose.yml
- JWT settings in `internal/api/handlers/auth.go`
- OLT connection details stored in repository

## Testing

- API testing can be done via tools like curl, Postman, or the frontend UI
- OLT command testing requires real device or mock implementations
