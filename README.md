# Twitter Clone Monorepo

Full-stack Twitter/X clone with a Go-first backend and Next.js frontend.

- `twitter-next-web` (frontend)
- `twitter-go-api` (backend)
- `twitter-java-api` (legacy/alternative backend)
- `infra` (Terraform IaC)

## Stack

- Frontend: Next.js (App Router), TypeScript, Tailwind, TanStack Query, Zustand
- Primary Backend: Go, Gin, PostgreSQL + sqlc, Redis
- Legacy Backend: Java (Spring Boot), PostgreSQL

## AWS Infrastructure

- Frontend: Amplify
- Ingress: API Gateway (HTTP API)
- Backend: EC2 (Go API + Docker)
- Database: RDS (PostgreSQL)
- Media: S3 (presigned URL uploads) + CloudFront (CDN)
- IaC: Terraform
- CI/CD: GitHub Actions

## Local Development

### 1. Start infra (Postgres + Redis)

```bash
docker-compose up -d
```

### 2. Run Go API

```bash
cd twitter-go-api
make migrateup
make run
```

### 3. Run Next.js Web

```bash
cd twitter-next-web
npm install
npm run dev
```

Web: `http://localhost:3000`  
API: `http://localhost:8080`

## Validation Commands

Go API:
```bash
cd twitter-go-api
go test ./...
```

Next.js Web:
```bash
cd twitter-next-web
npx tsc --noEmit
npm run lint
```

## Deployment

Deployment is configured with GitHub Actions to AWS:

Workflows:
- `.github/workflows/deploy-go-api.yml`
  - Trigger: push to `main` when `twitter-go-api/**` changes, or manual dispatch
  - Build: Docker image → GHCR
  - Deploy: SSH to EC2, pull image, restart container
