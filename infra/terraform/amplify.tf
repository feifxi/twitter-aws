# ── IAM Role for Amplify ───────
resource "aws_iam_role" "amplify" {
  name = "${var.project_name}-amplify-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "amplify.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "amplify_admin" {
  role       = aws_iam_role.amplify.name
  policy_arn = "arn:aws:iam::aws:policy/AdministratorAccess-Amplify"
}

# ── Amplify App ──────────────────────────────────────

resource "aws_amplify_app" "this" {
  name       = var.project_name
  repository = var.gh_repo_url

  # GitHub Personal Access Token
  access_token = var.gh_token

  iam_service_role_arn = aws_iam_role.amplify.arn

  # Build settings for Next.js in a sub-directory
  build_spec = <<-EOT
    version: 1
    applications:
      - appRoot: twitter-next-web
        frontend:
          phases:
            preBuild:
              commands:
                - npm ci
                - export NEXT_PUBLIC_API_URL=$(aws ssm get-parameter --name "/chmtwt/prod/NEXT_PUBLIC_API_URL" --with-decryption --query "Parameter.Value" --output text --region ap-southeast-1)
                - export NEXT_PUBLIC_GOOGLE_CLIENT_ID=$(aws ssm get-parameter --name "/chmtwt/prod/GOOGLE_CLIENT_ID" --with-decryption --query "Parameter.Value" --output text --region ap-southeast-1)
                - echo "NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL" >> .env.production
                - echo "NEXT_PUBLIC_GOOGLE_CLIENT_ID=$NEXT_PUBLIC_GOOGLE_CLIENT_ID" >> .env.production
            build:
              commands:
                - npm run build
          artifacts:
            baseDirectory: .next
            files:
              - '**/*'
          cache:
            paths:
              - node_modules/**/*
    EOT

  # Environment variables for the frontend
  # Using the API Gateway URL for the backend communication
  environment_variables = {
    AMPLIFY_MONOREPO_APP_ROOT    = "twitter-next-web"
    NEXT_TELEMETRY_DISABLED      = "1"
  }

  platform = "WEB_COMPUTE"

  # Auto-branch creation settings
  enable_auto_branch_creation = true
  auto_branch_creation_patterns = [
    "main",
    "master"
  ]

  auto_branch_creation_config {
    enable_auto_build = true
  }

  tags = { Name = "${var.project_name}-amplify-app" }
}

# ── Amplify Branch ───────────────────────────────────

resource "aws_amplify_branch" "main" {
  app_id      = aws_amplify_app.this.id
  branch_name = var.gh_branch

  framework         = "Next.js - SSR"
  stage             = "PRODUCTION"
  enable_auto_build = true
  tags = { Name = "${var.project_name}-amplify-branch-${var.gh_branch}" }
}

resource "aws_iam_role_policy" "amplify_ssm" {
  name = "${var.project_name}-amplify-ssm-policy"
  role = aws_iam_role.amplify.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "ssm:GetParameter",
        "ssm:GetParameters"
      ]
      Resource = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/chmtwt/prod/*"
    }]
  })
}
