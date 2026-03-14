# ── AWS Systems Manager (SSM) Parameter Store ────────

locals {
  ssm_parameters = {
    "/chmtwt/prod/HTTP_SERVER_ADDRESS"          = "0.0.0.0:8080"
    "/chmtwt/prod/DATABASE_URL"                 = "postgresql://${var.db_username}:${var.db_password}@${aws_db_instance.this.endpoint}/${var.db_name}?sslmode=require"
    "/chmtwt/prod/DB_MAX_CONNS"                 = var.db_max_conns
    "/chmtwt/prod/DB_MIN_CONNS"                 = var.db_min_conns
    "/chmtwt/prod/DB_MAX_CONN_LIFETIME_MINUTES" = var.db_max_conn_lifetime_minutes
    "/chmtwt/prod/MAX_MEDIA_BYTES"              = var.max_media_bytes
    "/chmtwt/prod/MAX_AVATAR_BYTES"             = var.max_avatar_bytes
    "/chmtwt/prod/MAX_BANNER_BYTES"             = var.max_banner_bytes
    "/chmtwt/prod/FRONTEND_URL"                 = join(",", local.frontend_origins)
    "/chmtwt/prod/COOKIE_DOMAIN"                = var.cookie_domain
    "/chmtwt/prod/COOKIE_SAME_SITE"             = var.cookie_same_site
    "/chmtwt/prod/COOKIE_SECURE"                = var.cookie_secure
    "/chmtwt/prod/TOKEN_SYMMETRIC_KEY"          = var.token_symmetric_key
    "/chmtwt/prod/TOKEN_DURATION_MINUTES"       = var.token_duration_minutes
    "/chmtwt/prod/REFRESH_TOKEN_DURATION_DAYS"  = var.refresh_token_duration_days
    "/chmtwt/prod/GOOGLE_CLIENT_ID"             = var.google_client_id
    "/chmtwt/prod/S3_BUCKET_NAME"               = aws_s3_bucket.media.id
    "/chmtwt/prod/S3_REGION"                    = var.aws_region
    "/chmtwt/prod/CLOUDFRONT_DOMAIN"            = aws_cloudfront_distribution.media.domain_name
    "/chmtwt/prod/GATEWAY_SECRET"               = var.gateway_secret
    "/chmtwt/prod/REDIS_ADDRESS"                = var.redis_address
    "/chmtwt/prod/REDIS_PASSWORD"               = var.redis_password
    "/chmtwt/prod/NEXT_PUBLIC_API_URL"          = "${aws_apigatewayv2_stage.default.invoke_url}api/v1"
  }
}

resource "aws_ssm_parameter" "config" {
  for_each = local.ssm_parameters

  name  = each.key
  type  = "SecureString"
  value = each.value == "" ? "N/A" : each.value

  tags = { Name = "${var.project_name}-ssm-${replace(each.key, "/", "-")}" }
}
