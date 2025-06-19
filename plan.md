# URL Shortener – End-to-End Plan

1. Goals
   REST API to create, resolve and delete short URLs
   Basic React UI for manual use / demos
   Automated tests (unit + integration)
   CI/CD pipeline that deploys to Google Cloud on every push to main
2. Tech Stack
   Backend (Golang)

- Go 1.22
- gin-gonic/gin (HTTP router)
- go-redis/redis v9 (storage)
- Test: Go's testing pkg + testify + httptest
- Containerisation: Docker + multi-stage build
  Frontend (React)
- React 18 + Vite
- TypeScript
- Tailwind (simple styling)
  DevOps
- GitHub Actions
- Google Cloud Run (container deploy)
- Artifact Registry (image store)
- Cloud Build / gcloud cli

3. Repository Layout (monorepo)
   /cmd/api → Go entrypoint (main.go)
   /internal → Go business logic
   /http → handlers, router
   /storage → Redis wrapper
   /id → Key generation
   /web → React app (Vite project)
   /deploy → Dockerfiles, k8s/terraform (future)
   /scripts → local helper scripts
   plan.md, todo.md → docs
   .github/workflows → CI/CD
4. Milestones & Deliverables
   a. Step 1 (Create)
   POST /api/v1/urls – idempotent shortener
   Persist {key ↔ longURL} in Redis
   400 on bad input, 201 on create/return
   b. Step 2 (Redirect)
   GET /{key} → 302 + Location header
   404 if key missing
   c. Step 3 (Delete)
   DELETE /api/v1/urls/{key}
   200 if existed, 204 if not
   d. Step 4 (UI)
   Single-page React app hitting the API
   Create + list + delete links, copy button
   Success view showing original and shortened URL
   Copy-to-clipboard functionality for shortened URLs
   Navigation to create new shortened links
   e. Step 5 (CI/CD)
   GitHub Actions: lint, test, build Docker, push to GCP
   Deploy to Cloud Run via gcloud
   PR checks + main branch auto-deploy
5. Key Generation
   Base62-encoded random 48-bit number (≈ 2.8 × 10¹⁴ combos)
   3-hour TTL per mapping (refreshed on create)
   Collision check in Redis; retry on clash
   Optional: hash(longURL) for idempotency
6. Testing Matrix
   Unit: key gen, storage wrapper, validation
   Integration: containerised Redis via Testcontainers
   E2E: Docker-compose API + Redis + cURL script
   Frontend: Vitest + React Testing Library
7. Security / Ops
   Rate-limit middleware (future)
   HTTPS only in production
   Cloud Secret Manager for Redis creds
   Monitoring: Cloud Logging + Cloud Monitoring
