# Contributing to Chat-Based Ecommerce Application

Thank you for your interest in contributing to the Chat-Based Ecommerce Application! This document provides guidelines and information for contributors.

## Code of Conduct

This project follows a code of conduct that we expect all contributors to follow. Please be respectful, inclusive, and constructive in all interactions.

## Getting Started

### Prerequisites
- Go 1.23+
- Node.js 18+
- Docker and Docker Compose
- Git

### Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/your-username/chat-ecommerce.git
   cd chat-ecommerce
   ```

2. **Set up environment variables**
   ```bash
   cp env.example .env
   cp frontend/env.example frontend/.env
   ```

3. **Start development services**
   ```bash
   docker-compose up -d postgres redis
   ```

4. **Install dependencies**
   ```bash
   # Backend
   cd backend
   go mod download
   
   # Frontend
   cd ../frontend
   npm install
   ```

## Development Guidelines

### Code Style

#### Backend (Go)
- Follow Go best practices and conventions
- Use `gofmt` for formatting
- Follow the project's `.golangci.yml` configuration
- Write comprehensive tests for new features
- Document public functions and types

#### Frontend (React/TypeScript)
- Use TypeScript for type safety
- Follow React best practices and hooks patterns
- Use Tailwind CSS for styling
- Write tests for components and utilities
- Follow ESLint and Prettier configurations

### Commit Messages

Use conventional commit format:
```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:
```
feat(chat): add OpenAI GPT-4 integration
fix(auth): resolve JWT token expiration issue
docs(api): update authentication endpoints
```

### Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, well-documented code
   - Add tests for new functionality
   - Update documentation as needed

3. **Run tests and linting**
   ```bash
   # Backend
   cd backend
   go test ./...
   golangci-lint run
   
   # Frontend
   cd frontend
   npm test
   npm run lint
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat(scope): your commit message"
   ```

5. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request**
   - Provide a clear description of changes
   - Reference any related issues
   - Include screenshots for UI changes
   - Ensure all CI checks pass

### Testing Requirements

#### Backend Tests
- Unit tests for all new functions and methods
- Integration tests for API endpoints
- Test coverage should be above 80%
- Use table-driven tests where appropriate

#### Frontend Tests
- Unit tests for React components
- Integration tests for user workflows
- Test coverage should be above 80%
- Use React Testing Library best practices

### Documentation

- Update README.md for significant changes
- Document new API endpoints
- Add JSDoc comments for complex functions
- Update environment variable documentation

## Issue Reporting

When reporting issues:

1. **Check existing issues** to avoid duplicates
2. **Use the issue template** provided
3. **Include relevant information**:
   - OS and version
   - Browser version (for frontend issues)
   - Steps to reproduce
   - Expected vs actual behavior
   - Screenshots or error logs

## Feature Requests

For feature requests:

1. **Check the roadmap** to see if it's already planned
2. **Provide detailed description** of the feature
3. **Explain the use case** and benefits
4. **Consider implementation complexity**

## Development Workflow

### Branch Strategy
- `main`: Production-ready code
- `develop`: Integration branch for features
- `feature/*`: Feature development branches
- `bugfix/*`: Bug fix branches
- `hotfix/*`: Critical production fixes

### Release Process
1. Features are developed in feature branches
2. Features are merged into `develop` after review
3. Releases are created from `develop` to `main`
4. Hotfixes can be applied directly to `main`

## Code Review Process

### For Contributors
- Ensure your code follows project standards
- Write clear commit messages
- Respond to review feedback promptly
- Test your changes thoroughly

### For Reviewers
- Be constructive and respectful
- Focus on code quality and functionality
- Check for security issues
- Verify tests and documentation

## Getting Help

- **Documentation**: Check the README and inline documentation
- **Issues**: Search existing issues or create a new one
- **Discussions**: Use GitHub Discussions for questions
- **Code Review**: Ask questions in pull request comments

## Recognition

Contributors will be recognized in:
- CONTRIBUTORS.md file
- Release notes
- Project documentation

Thank you for contributing to the Chat-Based Ecommerce Application!
