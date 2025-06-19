# URL Shortener Service

A modern, production-ready URL shortening service built with Go and Redis. Features include URL creation, resolution, and management with a clean REST API and web interface.

## Features

- Create shortened URLs with automatic key generation
- Resolve and redirect to original URLs
- Delete shortened URLs
- TTL-based expiration (3 hours, refreshed on access)
- Base62 encoding for short, readable keys
- Redis-backed storage for high performance
- Modern React web interface (coming soon)
- Production-ready with Docker support

## Quick Start

### Prerequisites

- Go 1.22 or later
- Redis 6.x or later
- Docker (optional)

### Running Locally

1. Clone the repository:

   ```bash
   git clone https://github.com/prayushdave/url-shortener.git
   cd url-shortener
   ```

2. Start Redis (if not already running):

   ```bash
   docker run -d -p 6379:6379 redis:6
   ```

3. Build and run the service:
   ```bash
   go build -o urlshortener ./cmd/api
   ./urlshortener
   ```

The service will be available at `http://localhost:8080`.

### Using Docker

Build and run using Docker:

```bash
docker build -t url-shortener .
docker run -p 8080:8080 --network host url-shortener
```

Note: `--network host` is used to allow the container to access Redis on localhost.

## API Usage

### Create a Short URL

```bash
curl -X POST http://localhost:8080/api/v1/urls \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com/very/long/url"}'
```

Response:

```json
{
  "short_key": "Ab3Kd9x2",
  "url": "https://example.com/very/long/url"
}
```

### Resolve a Short URL

```bash
curl -i http://localhost:8080/{short_key}
```

Response:

```http
HTTP/1.1 302 Found
Location: https://example.com/very/long/url
```

### Delete a Short URL

```bash
curl -X DELETE http://localhost:8080/api/v1/urls/{short_key}
```

Response:

```http
HTTP/1.1 200 OK
```

## Configuration

The service can be configured using environment variables:

- `REDIS_ADDR`: Redis server address (default: "localhost:6379")
- `REDIS_PASSWORD`: Redis password (default: "")
- `REDIS_DB`: Redis database number (default: 0)
- `SERVER_PORT`: HTTP server port (default: 8080)
- `BASE_URL`: Base URL for shortened links (default: "http://localhost:8080")

## Development

### Project Structure

```
/
├── cmd/api/          # Application entrypoint
├── internal/         # Internal packages
│   ├── http/        # HTTP handlers and routing
│   ├── storage/     # Redis storage implementation
│   └── id/          # Key generation
├── web/             # React frontend (coming soon)
└── deploy/          # Deployment configurations
```

### Running Tests

Run all tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

For comprehensive test coverage information, see [TESTING.md](TESTING.md).

**Current Test Coverage: 85.2%**

- HTTP Handlers: 76.7%
- ID Generator: 94.1%
- Storage Layer: 92.9%

### Key Features Implementation

1. **Key Generation**

   - Base62 encoding (0-9, a-z, A-Z)
   - 8-character keys (≈ 2.8 × 10¹⁴ combinations)
   - Cryptographically secure random generation
   - Collision handling with retries

2. **Storage**

   - Redis-backed for high performance
   - 3-hour TTL, refreshed on access
   - Atomic operations for concurrent safety
   - Error handling for connection issues

3. **API Design**
   - RESTful endpoints
   - Input validation
   - Proper status codes
   - Error handling
   - Rate limiting (coming soon)

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Deploying to Google Cloud Platform (Free Tier)

This guide will help you deploy the URL shortener to Google Cloud Platform using the free tier resources.

### Prerequisites

1. Create a Google Cloud account and set up billing (required for free tier)
2. Install the [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
3. Create a new GCP project
4. Enable required APIs:
   ```bash
   gcloud services enable \
     run.googleapis.com \
     artifactregistry.googleapis.com \
     redis.googleapis.com \
     cloudresourcemanager.googleapis.com
   ```

### Setup Steps

1. **Set up GitHub Secrets**

   Add the following secrets to your GitHub repository:

   - `GCP_PROJECT_ID`: Your Google Cloud project ID
   - `GCP_WORKLOAD_IDENTITY_PROVIDER`: The Workload Identity Provider
   - `GCP_SA_EMAIL`: Service account email
   - `REDIS_ADDR`: Redis instance address
   - `REDIS_PASSWORD`: Redis password
   - `BACKEND_URL`: The URL where your backend will be deployed

2. **Set up Workload Identity Federation**

   ```bash
   # Create a service account
   gcloud iam service-accounts create github-actions

   # Create Workload Identity Pool
   gcloud iam workload-identity-pools create "github-actions-pool" \
     --location="global"

   # Create Workload Identity Provider
   gcloud iam workload-identity-pools providers create-oidc "github-provider" \
     --location="global" \
     --workload-identity-pool="github-actions-pool" \
     --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" \
     --issuer-uri="https://token.actions.githubusercontent.com"

   # Add IAM policy binding
   gcloud projects add-iam-policy-binding $PROJECT_ID \
     --member="serviceAccount:github-actions@$PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/run.admin"
   ```

3. **Set up Redis**

   Create a Redis instance using Google Cloud Memorystore:

   ```bash
   gcloud redis instances create urlshortener-redis \
     --size=1 \
     --region=us-central1 \
     --redis-version=redis_6_x
   ```

4. **Deploy**

   Push your code to the main branch, and GitHub Actions will automatically deploy both frontend and backend to Cloud Run.

### Free Tier Limits

The deployment is configured to stay within GCP's free tier limits:

- **Cloud Run**:

  - Memory: 256MB per instance
  - CPU: 1 vCPU
  - Max instances: 2
  - Monthly requests: Up to 2 million (free tier limit)

- **Redis**:
  - Basic tier with 1GB memory
  - Sufficient for URL shortener use case

### Monitoring

Monitor your usage in the Google Cloud Console to ensure you stay within free tier limits:

1. Go to Cloud Run services to monitor request counts and performance
2. Check Memorystore for Redis metrics
3. Set up budget alerts to avoid unexpected charges

### Troubleshooting

1. If deployments fail, check:

   - GitHub Actions logs
   - Cloud Run service logs
   - IAM permissions

2. If the application is slow:
   - Check Redis connection
   - Monitor Cloud Run instance metrics
   - Verify network configuration

For more detailed logs and metrics, use the Google Cloud Console or run:

```bash
gcloud run services describe urlshortener-backend
gcloud run services describe urlshortener-frontend
```
