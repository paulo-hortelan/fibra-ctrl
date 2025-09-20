# fibra-ctrl

Fibra-Ctrl is a management system for optical line terminals (OLTs).

## Project Structure

- `/internal` - Internal application packages
- `/cmd` - Main applications for this project
- `/frontend` - React frontend application

## Docker Setup

This project includes Docker configurations for both the backend Go service and a React frontend.

### Prerequisites

- Docker
- Docker Compose

### Running the Application

1. Clone the repository:
```bash
git clone https://github.com/your-username/fibra-ctrl.git
cd fibra-ctrl
```

2. Start the services with Docker Compose:
```bash
docker-compose up -d
```

3. Access the services:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

### Development

#### Backend

The backend is written in Go and uses the Gin framework.

To build and run the backend locally:
```bash
go mod download
go run cmd/main.go
```

#### Frontend

The frontend is a React application.

To run the frontend locally:
```bash
cd frontend
npm install
npm start
```

## API Endpoints

- `POST /api/commands` - Submit a command to an OLT
- `GET /api/commands/:id` - Get the status of a command

## License

[Include your license information here]
