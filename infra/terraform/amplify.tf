# ── Amplify App ──────────────────────────────────────

resource "aws_amplify_app" "this" {
  name       = var.project_name
  repository = var.gh_repo_url
  
  # GitHub Personal Access Token
  access_token = var.gh_token

  # Build settings for Next.js in a sub-directory
  build_spec = <<-EOT
    version: 1
    applications:
      - frontend:
          phases:
            preBuild:
              commands:
                - cd twitter-next-web
                - npm ci
            build:
              commands:
                - npm run build
          artifacts:
            baseDirectory: twitter-next-web/.next
            files:
              - '**/*'
          cache:
            paths:
              - twitter-next-web/node_modules/**/*
    EOT

  # Environment variables for the frontend
  # Using the API Gateway URL for the backend communication
  environment_variables = {
    NEXT_PUBLIC_API_URL = "${aws_apigatewayv2_stage.default.invoke_url}api/v1"
    NEXT_TELEMETRY_DISABLED = "1"
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

  framework     = "Next.js - SSR"
  stage         = "PRODUCTION"
  enable_auto_build = true

  tags = { Name = "${var.project_name}-amplify-branch-${var.gh_branch}" }
}
