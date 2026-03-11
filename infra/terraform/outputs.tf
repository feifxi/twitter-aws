output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.this.id
}

output "ec2_public_ip" {
  description = "Public IP of the API server"
  value       = aws_instance.api.public_ip
}

output "rds_endpoint" {
  description = "RDS connection endpoint (host:port)"
  value       = aws_db_instance.this.endpoint
}

output "rds_database_url" {
  description = "Full PostgreSQL connection string"
  value       = "postgresql://${var.db_username}:${var.db_password}@${aws_db_instance.this.endpoint}/${var.db_name}?sslmode=require"
  sensitive   = true
}

output "s3_bucket_name" {
  description = "S3 media bucket name"
  value       = aws_s3_bucket.media.id
}

output "cloudfront_domain" {
  description = "CloudFront distribution domain for media"
  value       = aws_cloudfront_distribution.media.domain_name
}

output "api_gateway_url" {
  description = "API Gateway invoke URL"
  value       = aws_apigatewayv2_api.this.api_endpoint
}

output "amplify_app_url" {
  description = "Amplify Frontend URL"
  value       = "https://${aws_amplify_branch.main.branch_name}.${aws_amplify_app.this.default_domain}"
}

output "allowed_frontend_origins" {
  description = "Aggregated list of allowed frontend origins"
  value       = local.frontend_origins
}
