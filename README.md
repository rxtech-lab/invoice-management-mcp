# Invoice Management System

A comprehensive invoice management system built with Go and Next.js, featuring AI-powered search and Model Context Protocol (MCP) integration.

## Features

- 📄 **Invoice Management**: Create, read, update, and delete invoices with detailed line items
- 🏢 **Company & Receiver Management**: Organize vendors and invoice recipients
- 🏷️ **Categories & Tags**: Flexible categorization and tagging system
- 📊 **Analytics & Reporting**: Built-in analytics and statistics dashboard
- 📤 **File Upload**: S3-compatible storage for invoice attachments
- 🔍 **AI-Powered Search**: Intelligent search capabilities using AI SDK
- 🤖 **MCP Server**: Model Context Protocol integration for AI assistants
- 💱 **Currency Support**: Multi-currency with exchange rate conversion
- 🔐 **OAuth Authentication**: Secure authentication with JWT tokens
- 📱 **Responsive UI**: Modern Next.js frontend with mobile support

## Tech Stack

### Backend

- **Go 1.25+**: Main backend language
- **Fiber v2**: Fast HTTP framework
- **GORM**: ORM for database operations
- **Turso/SQLite**: Database (supports both local SQLite and Turso)
- **AWS SDK v2**: S3-compatible storage integration
- **MCP Go**: Model Context Protocol server implementation
- **OpenAPI 3.0**: API specification and code generation

### Frontend

- **Next.js 16**: React framework
- **TypeScript**: Type-safe development
- **TanStack Query**: Data fetching and caching
- **TanStack Table**: Advanced table components
- **Shadcn UI**: Component library
- **Tailwind CSS**: Utility-first styling
- **React Hook Form**: Form management with Zod validation
- **Recharts**: Data visualization
- **AI SDK**: AI-powered features

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Node.js 20 or higher
- Docker & Docker Compose (optional)
- Bun or npm package manager

### Environment Variables

Create a `.env` file in the root directory with the following variables:

```env
# Server
PORT=8080

# Database (Turso) - Leave empty to use local SQLite
TURSO_DATABASE_URL=
TURSO_AUTH_TOKEN=

# S3-compatible storage
S3_ENDPOINT=your-s3-endpoint
S3_BUCKET=your-bucket-name
S3_ACCESS_KEY=your-access-key
S3_SECRET_KEY=your-secret-key
S3_REGION=us-east-1
S3_USE_PATH_STYLE=false

# Authentication
MCPROUTER_SERVER_URL=your-mcprouter-url
MCPROUTER_SERVER_API_KEY=your-api-key
OAUTH_SERVER_URL=your-oauth-server
OAUTH_ISSUER=your-oauth-issuer
OAUTH_AUDIENCE=your-oauth-audience

# File Server (for unlinking files)
FILE_SERVER_URL=your-file-server-url
```

### Installation

#### Option 1: Using Docker Compose (Recommended)

```bash
# Start the application
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the application
docker-compose down
```

#### Option 2: Local Development

**Backend:**

```bash
# Install dependencies
make deps

# Generate OpenAPI code
make generate

# Build the binary
make build

# Run the server
make run

# Or run with hot reload (requires air)
air
```

**Frontend:**

```bash
cd frontend

# Install dependencies
bun install

# Run development server
bun dev

# Build for production
bun build
```

### Development Commands

**Backend:**

```bash
make deps           # Download and tidy dependencies
make build          # Build the binary
make run            # Run the server
make test           # Run tests
make test-e2e       # Run E2E tests
make test-coverage  # Run tests with coverage
make fmt            # Format code
make lint           # Lint code
make generate       # Generate OpenAPI code
make clean          # Clean build artifacts
make install-local  # Install binary to /usr/local/bin
```

**Frontend:**

```bash
bun dev     # Start development server
bun build   # Build for production
bun start   # Start production server
bun lint    # Run ESLint
```

## Project Structure

```
.
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── api/             # API server and handlers
│   │   ├── handlers/    # HTTP request handlers
│   │   ├── middleware/  # Middleware functions
│   │   └── generated/   # OpenAPI generated code
│   ├── assets/          # Static assets and OpenAPI spec
│   ├── mcp/             # Model Context Protocol server
│   ├── models/          # Database models
│   ├── services/        # Business logic services
│   ├── tools/           # MCP tools implementation
│   └── utils/           # Utility functions
├── frontend/            # Next.js frontend application
│   ├── app/             # Next.js app directory
│   │   ├── (auth)/      # Authentication pages
│   │   ├── (dashboard)/ # Dashboard pages
│   │   └── api/         # API routes
│   ├── components/      # React components
│   ├── hooks/           # Custom React hooks
│   └── lib/             # Utility libraries and API clients
├── e2e/                 # End-to-end tests
├── k8s/                 # Kubernetes deployment configs
└── oapi-codegen/        # OpenAPI code generation configs
```

## API Documentation

The API follows OpenAPI 3.0 specification. Key endpoints include:

- `/api/health` - Health check
- `/api/categories` - Category management
- `/api/companies` - Company/vendor management
- `/api/receivers` - Invoice receiver management
- `/api/tags` - Tag management
- `/api/invoices` - Invoice CRUD operations
- `/api/invoices/{id}/items` - Invoice line items
- `/api/upload` - File upload operations
- `/api/analytics/*` - Analytics and statistics

Full API documentation is available in [internal/assets/openapi.yaml](internal/assets/openapi.yaml).

## MCP Server

This project includes a Model Context Protocol (MCP) server that allows AI assistants to interact with the invoice management system. The MCP server provides tools for:

- Creating, reading, updating, and deleting invoices
- Managing categories, companies, receivers, and tags
- Uploading and managing files
- Generating analytics and statistics

The MCP server is automatically started alongside the main API server.

## Kubernetes Deployment

Kubernetes deployment configurations are available in the `k8s/` directory:

```bash
cd k8s

# Apply all configurations
kubectl apply -k .

# Or apply individually
kubectl apply -f namespace.yaml
kubectl apply -f secrets.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml
kubectl apply -f cluster-issuer.yaml
```

See [k8s/README.md](k8s/README.md) for detailed deployment instructions.

## Testing

**Unit Tests:**

```bash
make test
```

**E2E API Tests:**

```bash
make test-e2e
```

**Coverage Report:**

```bash
make test-coverage
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is maintained by [RxTech Lab](https://github.com/rxtech-lab).

## Support

For issues and questions, please open an issue on GitHub.

## Acknowledgments

- Built with [Fiber](https://gofiber.io/)
- Frontend powered by [Next.js](https://nextjs.org/)
- UI components from [Shadcn UI](https://ui.shadcn.com/)
- MCP integration via [mcp-go](https://github.com/mark3labs/mcp-go)
