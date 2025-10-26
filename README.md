# Chat-Based Ecommerce Application

A modern ecommerce platform that enables users to shop entirely through conversational chat interface, while maintaining traditional catalog browsing and checkout options. Built with Golang backend, React/Vite frontend, and OpenAI GPT-4 integration.

## Features

- **Chat Shopping**: Complete purchase journey through conversational interface
- **Traditional Web Interface**: Standard catalog browsing and checkout
- **Inventory Management**: Real-time stock tracking and admin interface
- **Real-time Synchronization**: Shared cart state across all interfaces
- **Payment Processing**: Stripe integration for secure payments
- **User Authentication**: JWT-based authentication with session management

## Technology Stack

### Backend
- **Golang** with Gin framework
- **Air** for live reloading in development
- **PostgreSQL** database with GORM ORM
- **Redis** for caching and sessions
- **OpenAI GPT-4** for natural language processing
- **WebSocket** for real-time communication
- **Stripe** for payment processing

### Frontend
- **React** with TypeScript
- **Vite** for fast development and builds
- **Tailwind CSS** for styling
- **Socket.io** for WebSocket communication
- **React Router** for navigation

### DevOps
- **Docker** containerization
- **GitHub Actions** for CI/CD
- **PostgreSQL** and **Redis** services

## Quick Start

### Prerequisites
- Go 1.23+
- Node.js 18+
- Docker and Docker Compose
- PostgreSQL 15+
- Redis 7+

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd chat-ecommerce
   ```

2. **Set up environment variables**
   ```bash
   cp env.example .env
   cp frontend/env.example frontend/.env
   ```
   Update the environment variables with your actual values.

3. **Start services with Docker Compose**
   ```bash
   docker-compose up -d postgres redis
   ```

4. **Set up the backend**
   
   Option A: Using Air for live reloading (Recommended)
   ```bash
   cd backend
   go mod download
   
   # Install Air if not already installed
   go install github.com/air-verse/air@latest
   
   # Run with Air (enables live reloading)
   air
   ```
   
   Option B: Using standard Go run
   ```bash
   cd backend
   go mod download
   go run cmd/api/main.go
   ```
   
   Option C: Using Docker Compose with Air (with live reloading)
   ```bash
   # Start all services with live reloading
   docker-compose -f docker-compose.dev.yml up
   
   # OR start in background
   docker-compose -f docker-compose.dev.yml up -d
   ```

   **How it works:** Docker volumes mount your local `./backend` directory into `/app` in the container. When you edit files on your host, Docker reflects the changes in the container, and Air detects them and rebuilds. Polling is enabled for cross-platform compatibility.

5. **Set up the frontend**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

6. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - API Documentation: http://localhost:8080/docs

#### Using Air for Live Reloading

[Air](https://github.com/air-verse/air) is configured for the backend to provide automatic code reloading during development.

**Features:**
- ✅ Automatic rebuild on file changes
- ✅ No need to manually restart the server
- ✅ Faster development workflow
- ✅ Colorful log output for better debugging

**Usage:**
```bash
cd backend
air
```

Air configuration is in `backend/.air.toml`. You can customize it to:
- Add/remove files/directories to watch
- Change build commands
- Adjust delay times
- Configure logging behavior

To run Air with custom arguments:
```bash
cd backend
air -c .air.toml -- -custom-arg
```

### Production Deployment

1. **Build Docker images**
   ```bash
   docker-compose build
   ```

2. **Deploy with Docker Compose**
   ```bash
   docker-compose up -d
   ```

## Project Structure

```
chat-ecommerce/
├── backend/                 # Golang backend
│   ├── cmd/api/            # Application entry point
│   ├── internal/           # Private application code
│   │   ├── handlers/       # HTTP handlers
│   │   ├── middleware/     # HTTP middleware
│   │   ├── models/         # Data models
│   │   └── services/       # Business logic
│   ├── pkg/                # Public packages
│   │   ├── auth/           # Authentication utilities
│   │   ├── database/       # Database connection
│   │   └── websocket/      # WebSocket utilities
│   └── tests/              # Test files
├── frontend/               # React frontend
│   ├── src/
│   │   ├── components/     # React components
│   │   ├── services/       # API services
│   │   ├── hooks/          # Custom hooks
│   │   └── contexts/       # React contexts
│   └── tests/              # Test files
├── specs/                  # Project specifications
└── scripts/                # Utility scripts
```

## API Documentation

The API follows RESTful principles and includes:

- **Authentication**: JWT-based user authentication
- **Products**: Product catalog management
- **Cart**: Shopping cart operations
- **Orders**: Order processing and management
- **Chat**: Conversational interface endpoints
- **Admin**: Administrative functions

API documentation is available at `/docs` when running the backend.

## Testing

### Backend Tests
```bash
cd backend
go test -v ./...
go test -race ./...
go test -cover ./...
```

### Frontend Tests
```bash
cd frontend
npm test
npm run test:coverage
```

### End-to-End Tests
```bash
npm run test:e2e
```

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code of conduct and the process for submitting pull requests.

## Environment Variables

### Backend (.env)
- `DB_HOST`: PostgreSQL host
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `REDIS_HOST`: Redis host
- `JWT_SECRET`: JWT signing secret
- `OPENAI_API_KEY`: OpenAI API key
- `STRIPE_SECRET_KEY`: Stripe secret key

### Frontend (.env)
- `VITE_API_BASE_URL`: Backend API URL
- `VITE_WS_URL`: WebSocket URL
- `VITE_STRIPE_PUBLISHABLE_KEY`: Stripe publishable key

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support, please open an issue in the GitHub repository or contact the development team.

## Roadmap

- [ ] Mobile app development
- [ ] Multi-language support
- [ ] Advanced analytics dashboard
- [ ] Machine learning recommendations
- [ ] Voice interface integration
