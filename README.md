# Twitter Clone

Full-stack Twitter/X clone with a Next.js frontend, Go backend, and AWS infrastructure managed by Terraform.

## Features

- **Authentication** — Google OAuth with JWT access & refresh tokens
- **Tweets** — Create, delete, like, retweet, and threaded replies
- **Feeds** — "For You" global feed with **time-decay gravity algorithm** (weighing likes, retweets, and replies), following-only feed, and per-user profile feed
- **Follow / Unfollow** — Follow users with follower & following lists
- **Media Uploads** — S3 presigned-URL uploads served via CloudFront CDN
- **Real-time Notifications** — Server-Sent Events (SSE) streaming with unread count
- **Direct Messages** — Real-time conversations over WebSocket
- **Search** — Search users, tweets, and hashtags
- **Trending & Discovery** — Trending hashtags and suggested users

## Architecture

![AWS Architecture](docs/assets/Chanom_Twitter_AWS_Architecture.jpg)

## Tech Stack

| Layer | Technologies |
|---|---|
| Frontend | Next.js (App Router), TypeScript, Tailwind CSS, TanStack Query, Zustand |
| Backend | Go, Gin, PostgreSQL + sqlc, Redis |
| Infrastructure | AWS (Amplify, API Gateway, EC2, RDS, S3, CloudFront), Terraform, GitHub Actions |

## Prerequisites

| Tool | Purpose |
|---|---|
| [Go](https://go.dev/dl/) | Backend API development |
| [Node.js](https://nodejs.org/) (v18+) | Frontend development |
| [Docker](https://www.docker.com/) | Local Postgres + Redis, and production deployment |
| [AWS CLI v2](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html) | AWS resource management & SSH tunneling |
| [Terraform](https://developer.hashicorp.com/terraform/install) (v1.5+) | Infrastructure provisioning |
| [Make](https://www.gnu.org/software/make/) | Running Go API tasks (migrations, etc.) |

**Install via Homebrew (macOS):**

```bash
brew install go node docker awscli hashicorp/tap/terraform make
```

## AWS Infrastructure

| Service | Purpose |
|---|---|
| **Amplify** | Hosts & auto-deploys the Next.js frontend (SSR) |
| **API Gateway** | HTTP API proxy to the backend EC2 instance |
| **EC2** | Runs the Go API in Docker via `docker compose` |
| **RDS** | Managed PostgreSQL database (private subnet) |
| **S3** | Media storage with presigned-URL uploads |
| **CloudFront** | CDN for serving S3 media over HTTPS |
| **SSM Parameter Store** | Securely manages, stores, and injects runtime configuration into the Go API |
| **Terraform** | Infrastructure as Code for all resources |
| **GitHub Actions** | CI/CD pipeline pulling from GHCR and deploying via SSM Run Command |

## AWS Setup

> You need an [AWS account](https://aws.amazon.com/free/) to deploy this project.

### 1. Create an IAM user for Terraform

1. Go to the [IAM Console](https://console.aws.amazon.com/iam/) → Users → **Create user**
2. Attach the **AdministratorAccess** policy (for a learning project; use least-privilege in production)
3. Create an **Access Key** (select "CLI" use case) and save the Access Key ID and Secret

### 2. Configure AWS CLI

```bash
aws configure
```

Enter your Access Key ID, Secret, region (`ap-southeast-1`), and output format (`json`).

### 3. Provision infrastructure

```bash
cd infra/terraform
cp terraform.tfvars.example terraform.tfvars  # Fill in your values
terraform init
terraform plan     # Review what will be created
terraform apply    # Create all resources (~5 min)
```

To tear down all resources and **stop all AWS charges**:

```bash
terraform destroy
```

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

## Deployment

### Frontend — AWS Amplify

The frontend is deployed automatically via **AWS Amplify**, configured through Terraform ([`amplify.tf`](infra/terraform/amplify.tf)).

**How it works:**

1. Amplify connects to the GitHub repository using a Personal Access Token
2. On every push to `main`, Amplify auto-builds the Next.js app:
   - Runs `npm ci` and `npm run build` inside `twitter-next-web/`
   - Deploys as a Server-Side Rendered (SSR) application (`WEB_COMPUTE` platform)
3. Environment variables (e.g. `NEXT_PUBLIC_API_URL`) are injected by Terraform from the API Gateway stage URL

**Amplify default URL pattern:** `https://<branch>.<app-id>.amplifyapp.com`

### Backend — GitHub Actions → EC2

The Go API is deployed via the [`deploy-go-api.yml`](.github/workflows/deploy-go-api.yml) GitHub Actions workflow.

**Trigger:** Push to `main` when `twitter-go-api/**` changes, or manual dispatch.

**Pipeline:**

1. **Test** — Sets up Go and runs `go test ./...`
2. **Build & Push** — Builds a Docker image and pushes it to GitHub Container Registry (GHCR)
3. **Deploy** — Sends commands to EC2 via **AWS SSM** (no SSH key needed), pulls the latest image, and restarts the container via `docker compose`

**Required GitHub Secrets:**

| Secret | Description |
|---|---|
| `AWS_ACCESS_KEY_ID` | IAM access key for SSM commands |
| `AWS_SECRET_ACCESS_KEY` | IAM secret key |
| `AWS_REGION` | AWS region (e.g. `ap-southeast-1`) |
| `EC2_INSTANCE_ID` | EC2 instance ID (e.g. `i-0e83a1b99...`) |
| `GHCR_PAT` | GitHub PAT with `read:packages` scope for pulling from GHCR |

## AWS Access

### SSH into EC2

Connect to the EC2 instance using **EC2 Instance Connect Endpoint** (no `.pem` key needed):

```bash
aws ec2-instance-connect ssh --instance-id <INSTANCE_ID>
```

### Connect to RDS (via SSH tunnel)

RDS is in a private subnet and cannot be accessed directly. Use the EC2 instance as a bastion host to create an SSH tunnel:

```bash
aws ec2-instance-connect ssh \
    --instance-id <INSTANCE_ID> \
    --local-forwarding 5433:<RDS_ENDPOINT>:5432
```

This forwards your local port `5433` to RDS port `5432` through the EC2 instance. Port `5433` is used to avoid conflicts with any local PostgreSQL running on `5432`.

**Keep the terminal window open**, then connect with your database tool (DBeaver, TablePlus, etc.):

| Setting | Value |
|---|---|
| Host | `127.0.0.1` |
| Port | `5433` |
| Database | `twitter_db` |
| Username | _(from `terraform.tfvars`)_ |
| Password | _(from `terraform.tfvars`)_ |

> **Tip:** Add `-- -N` at the end to open the tunnel without dropping into the EC2 shell:
> ```bash
> aws ec2-instance-connect ssh \
>     --instance-id <INSTANCE_ID> \
>     --local-forwarding 5433:<RDS_ENDPOINT>:5432 \
>     -- -N
> ```

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

## Project Structure

```
├── twitter-next-web/    # Next.js frontend
├── twitter-go-api/      # Go backend (primary)
├── twitter-java-api/    # Java backend (legacy/alternative)
├── infra/
│   ├── terraform/       # AWS infrastructure (Terraform)
│   └── ec2/             # EC2 setup scripts & docker-compose
├── docs/
│   └── assets/          # Architecture diagrams
└── .github/workflows/   # CI/CD pipelines
```
