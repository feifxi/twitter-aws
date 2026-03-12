locals {
  # The Amplify default domain follows this pattern: https://branch-name.d12345.amplifyapp.com
  amplify_url = "https://${var.gh_branch}.${aws_amplify_app.this.default_domain}"

  # Combine development URLs (if any) and the production Amplify URL
  frontend_origins = var.frontend_url != "" ? concat(split(",", var.frontend_url), [local.amplify_url]) : [local.amplify_url]
}
