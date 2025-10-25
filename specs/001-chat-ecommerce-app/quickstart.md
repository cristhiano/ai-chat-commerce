# Quickstart Guide: Chat-Based Ecommerce Application

**Date:** 2024-12-19  
**Project:** Chat-Based Ecommerce Application  
**Technology Stack:** Golang, React/Vite, PostgreSQL, OpenAI GPT-4  
**Version:** 1.0

## Overview

This guide provides a quick start to understanding and implementing the chat-based ecommerce application using Golang for the backend, React with Vite for the frontend, PostgreSQL for persistence, and OpenAI GPT-4 for natural language processing.

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React/Vite    │    │   React/Vite    │    │   React/Vite    │
│   Chat Interface│    │ Traditional UI  │    │ Inventory Mgmt  │
│                 │    │                 │    │                 │
│ • OpenAI GPT-4  │    │ • Catalog Browse│    │ • Product Admin │
│ • WebSocket     │    │ • Search/Filter │    │ • Stock Tracking│
│ • Cart Management│    │ • Checkout Flow│    │ • Reports       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │  Golang Backend │
                    │                 │
                    │ • Gin/Echo API  │
                    │ • GORM Models   │
                    │ • JWT Auth      │
                    │ • WebSocket     │
                    │ • OpenAI Client │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │   PostgreSQL    │
                    │                 │
                    │ • Product Data  │
                    │ • User Accounts │
                    │ • Orders        │
                    │ • Chat Sessions │
                    └─────────────────┘
```

## Technology Stack Details

### Backend (Golang)
- **Framework**: Gin or Echo for HTTP API
- **ORM**: GORM for database operations
- **Authentication**: JWT tokens with refresh rotation
- **Real-time**: WebSocket connections
- **AI Integration**: OpenAI GPT-4 API client
- **Database**: PostgreSQL with Redis caching

### Frontend (React/Vite)
- **Framework**: React with TypeScript
- **Build Tool**: Vite for fast development and optimized builds
- **Styling**: Tailwind CSS for utility-first styling
- **State Management**: React Context with useReducer
- **HTTP Client**: Axios for API communication
- **WebSocket**: Socket.io-client for real-time features

### Database (PostgreSQL)
- **Primary Database**: PostgreSQL for ACID compliance
- **Caching**: Redis for session data and API responses
- **Migrations**: GORM AutoMigrate for schema management
- **Indexing**: Strategic indexes for performance optimization

## Key Features

### 1. Chat Shopping Experience
- **OpenAI Integration**: GPT-4 processes natural language queries
- **Anonymous Shopping**: Users can start shopping without authentication
- **Context Awareness**: Maintains conversation history and shopping preferences
- **Cart Management**: Add, modify, remove items through conversation
- **Checkout Integration**: Complete purchase flow via chat

### 2. Traditional Web Interface
- **React Components**: Modular, reusable UI components
- **Vite Optimization**: Fast development server and optimized production builds
- **Product Catalog**: Browse, search, filter products with TypeScript safety
- **Shopping Cart**: Standard ecommerce cart functionality
- **User Accounts**: Registration, login, order history

### 3. Inventory Management
- **Admin Dashboard**: React-based administrative interface
- **Real-time Stock**: Live inventory tracking and updates
- **Low Stock Alerts**: Automated notifications for restocking
- **Bulk Operations**: Import/export product data

## Development Setup

### Prerequisites
- Go 1.21+ and Go modules
- Node.js 18+ and npm/yarn
- PostgreSQL 14+
- Redis 6+
- OpenAI API key
- Stripe API keys

### Backend Setup (Golang)

```bash
# Create Go module
mkdir ecommerce-chat-backend
cd ecommerce-chat-backend
go mod init github.com/yourorg/ecommerce-chat-backend

# Install dependencies
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get github.com/golang-jwt/jwt/v5
go get github.com/google/uuid
go get github.com/gorilla/websocket
go get github.com/sashabaranov/go-openai
go get github.com/go-redis/redis/v8

# Project structure
mkdir -p cmd/api
mkdir -p internal/{models,handlers,middleware,services,config}
mkdir -p pkg/{database,websocket,auth}
mkdir -p migrations
mkdir -p tests
```

### Frontend Setup (React/Vite)

```bash
# Create Vite project
npm create vite@latest ecommerce-chat-frontend -- --template react-ts
cd ecommerce-chat-frontend

# Install dependencies
npm install
npm install axios
npm install socket.io-client
npm install @tailwindcss/forms
npm install react-router-dom
npm install @headlessui/react
npm install @heroicons/react

# Install dev dependencies
npm install -D @types/node
npm install -D tailwindcss
npm install -D autoprefixer
npm install -D postcss
npm install -D @testing-library/react
npm install -D @testing-library/jest-dom
npm install -D vitest
npm install -D jsdom
```

### Database Setup

```bash
# Start PostgreSQL and Redis with Docker
docker-compose up -d postgres redis

# Create database
createdb ecommerce_chat

# Run migrations
go run cmd/migrate/main.go
```

### Environment Variables

**Backend (.env)**
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=ecommerce_chat
REDIS_URL=redis://localhost:6379

# Authentication
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRES_IN=24h

# External APIs
OPENAI_API_KEY=your-openai-api-key
STRIPE_SECRET_KEY=your-stripe-secret-key
STRIPE_PUBLISHABLE_KEY=your-stripe-publishable-key

# Application
GIN_MODE=debug
PORT=8080
CORS_ORIGIN=http://localhost:5173
```

**Frontend (.env)**
```env
VITE_API_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws
VITE_STRIPE_PUBLISHABLE_KEY=your-stripe-publishable-key
```

## Quick Start Commands

### Backend Development
```bash
# Start development server
go run cmd/api/main.go

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Build for production
go build -o bin/api cmd/api/main.go
```

### Frontend Development
```bash
# Start development server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Run tests
npm run test

# Run tests with coverage
npm run test:coverage
```

## Project Structure

### Backend Structure
```
ecommerce-chat-backend/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── models/
│   │   ├── product.go
│   │   ├── user.go
│   │   ├── order.go
│   │   └── chat.go
│   ├── handlers/
│   │   ├── product.go
│   │   ├── cart.go
│   │   ├── order.go
│   │   └── chat.go
│   ├── services/
│   │   ├── product_service.go
│   │   ├── chat_service.go
│   │   └── order_service.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logging.go
│   └── config/
│       └── config.go
├── pkg/
│   ├── database/
│   │   └── connection.go
│   ├── websocket/
│   │   └── hub.go
│   └── auth/
│       └── jwt.go
└── tests/
    ├── integration/
    └── unit/
```

### Frontend Structure
```
ecommerce-chat-frontend/
├── src/
│   ├── components/
│   │   ├── chat/
│   │   ├── product/
│   │   ├── cart/
│   │   └── admin/
│   ├── pages/
│   │   ├── Home.tsx
│   │   ├── ProductList.tsx
│   │   ├── ProductDetail.tsx
│   │   ├── Cart.tsx
│   │   ├── Checkout.tsx
│   │   └── Admin.tsx
│   ├── hooks/
│   │   ├── useAuth.ts
│   │   ├── useCart.ts
│   │   └── useWebSocket.ts
│   ├── services/
│   │   ├── api.ts
│   │   ├── auth.ts
│   │   └── websocket.ts
│   ├── types/
│   │   ├── product.ts
│   │   ├── user.ts
│   │   └── order.ts
│   ├── utils/
│   │   ├── constants.ts
│   │   └── helpers.ts
│   └── App.tsx
├── public/
└── tests/
```

## Key Implementation Examples

### Golang API Handler
```go
// internal/handlers/product.go
func (h *ProductHandler) GetProducts(c *gin.Context) {
    var query ProductQuery
    if err := c.ShouldBindQuery(&query); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    products, total, err := h.productService.GetProducts(query)
    if err != nil {
        c.JSON(500, gin.H{"error": "Failed to fetch products"})
        return
    }
    
    c.JSON(200, gin.H{
        "products": products,
        "pagination": Pagination{
            Page:  query.Page,
            Limit: query.Limit,
            Total: total,
            Pages: int(math.Ceil(float64(total) / float64(query.Limit))),
        },
    })
}
```

### React Component
```tsx
// src/components/chat/ChatInterface.tsx
import { useState, useEffect } from 'react';
import { useWebSocket } from '../../hooks/useWebSocket';

export const ChatInterface: React.FC = () => {
    const [message, setMessage] = useState('');
    const [messages, setMessages] = useState<ChatMessage[]>([]);
    const { sendMessage, lastMessage } = useWebSocket();
    
    useEffect(() => {
        if (lastMessage) {
            setMessages(prev => [...prev, lastMessage]);
        }
    }, [lastMessage]);
    
    const handleSendMessage = async () => {
        if (!message.trim()) return;
        
        const userMessage: ChatMessage = {
            id: Date.now().toString(),
            type: 'user',
            content: message,
            timestamp: new Date(),
        };
        
        setMessages(prev => [...prev, userMessage]);
        sendMessage(message);
        setMessage('');
    };
    
    return (
        <div className="flex flex-col h-full">
            <div className="flex-1 overflow-y-auto p-4">
                {messages.map(msg => (
                    <div key={msg.id} className={`mb-4 ${msg.type === 'user' ? 'text-right' : 'text-left'}`}>
                        <div className={`inline-block p-3 rounded-lg ${
                            msg.type === 'user' 
                                ? 'bg-blue-500 text-white' 
                                : 'bg-gray-200 text-gray-800'
                        }`}>
                            {msg.content}
                        </div>
                    </div>
                ))}
            </div>
            <div className="p-4 border-t">
                <div className="flex gap-2">
                    <input
                        type="text"
                        value={message}
                        onChange={(e) => setMessage(e.target.value)}
                        onKeyPress={(e) => e.key === 'Enter' && handleSendMessage()}
                        className="flex-1 p-2 border rounded-lg"
                        placeholder="Type your message..."
                    />
                    <button
                        onClick={handleSendMessage}
                        className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                    >
                        Send
                    </button>
                </div>
            </div>
        </div>
    );
};
```

### OpenAI Integration
```go
// internal/services/chat_service.go
func (s *ChatService) ProcessMessage(ctx context.Context, sessionID string, message string) (*ChatResponse, error) {
    // Get conversation context
    context, err := s.getConversationContext(sessionID)
    if err != nil {
        return nil, err
    }
    
    // Prepare OpenAI request
    messages := []openai.ChatCompletionMessage{
        {
            Role:    openai.ChatMessageRoleSystem,
            Content: s.getSystemPrompt(),
        },
    }
    
    // Add conversation history
    for _, msg := range context.ConversationHistory {
        messages = append(messages, openai.ChatCompletionMessage{
            Role:    msg.Role,
            Content: msg.Content,
        })
    }
    
    // Add current message
    messages = append(messages, openai.ChatCompletionMessage{
        Role:    openai.ChatMessageRoleUser,
        Content: message,
    })
    
    // Call OpenAI API
    resp, err := s.openaiClient.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model:    openai.GPT4,
        Messages: messages,
        Functions: s.getFunctionDefinitions(),
    })
    
    if err != nil {
        return nil, err
    }
    
    // Process response and update context
    return s.processOpenAIResponse(resp, sessionID)
}
```

## Testing Strategy

### Backend Testing (Go)
```go
// tests/integration/product_test.go
func TestGetProducts(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Create test products
    products := createTestProducts(t, db)
    
    // Test API endpoint
    w := httptest.NewRecorder()
    req := httptest.NewRequest("GET", "/api/v1/products", nil)
    
    router := setupTestRouter(db)
    router.ServeHTTP(w, req)
    
    assert.Equal(t, 200, w.Code)
    
    var response ProductListResponse
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Len(t, response.Products, len(products))
}
```

### Frontend Testing (Vitest)
```tsx
// tests/components/ChatInterface.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { ChatInterface } from '../../src/components/chat/ChatInterface';

describe('ChatInterface', () => {
    it('should send message when user types and presses enter', async () => {
        render(<ChatInterface />);
        
        const input = screen.getByPlaceholderText('Type your message...');
        const sendButton = screen.getByText('Send');
        
        fireEvent.change(input, { target: { value: 'Hello' } });
        fireEvent.click(sendButton);
        
        expect(screen.getByText('Hello')).toBeInTheDocument();
    });
});
```

## Deployment

### Docker Configuration

**Backend Dockerfile**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

**Frontend Dockerfile**
```dockerfile
FROM node:18-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Docker Compose
```yaml
version: '3.8'
services:
  postgres:
    image: postgres:14
    environment:
      POSTGRES_DB: ecommerce_chat
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"

  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_URL=redis://redis:6379
    depends_on:
      - postgres
      - redis

  frontend:
    build: ./frontend
    ports:
      - "80:80"
    depends_on:
      - backend

volumes:
  postgres_data:
```

## Performance Optimization

### Backend Optimizations
- **Connection Pooling**: Configure PostgreSQL connection pool
- **Query Optimization**: Use GORM's preloading and select optimization
- **Caching**: Redis caching for frequently accessed data
- **Compression**: Gzip compression for API responses

### Frontend Optimizations
- **Vite Build**: Tree-shaking and code splitting
- **Lazy Loading**: React.lazy for route-based code splitting
- **Image Optimization**: WebP format and lazy loading
- **Bundle Analysis**: Regular bundle size monitoring

## Monitoring and Observability

### Backend Monitoring
- **Logging**: Structured JSON logs with Go's log package
- **Metrics**: Prometheus metrics with Gin middleware
- **Tracing**: OpenTelemetry for distributed tracing
- **Health Checks**: Kubernetes-ready health endpoints

### Frontend Monitoring
- **Error Tracking**: Sentry integration for error reporting
- **Performance**: Web Vitals monitoring
- **Analytics**: User behavior tracking
- **Bundle Analysis**: Regular performance audits

---

This quickstart guide provides the foundation for implementing the chat-based ecommerce application with Golang, React/Vite, PostgreSQL, and OpenAI GPT-4. The architecture ensures high performance, maintainability, and scalability while providing an excellent user experience.
