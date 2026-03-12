variable "project_name" {
  description = "Prefix for all resource names"
  type        = string
  default     = "chmtwt"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-southeast-1"
}

variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

# ── EC2 ──────────────────────────────────────────────

variable "ec2_instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}



# ── RDS ──────────────────────────────────────────────

variable "db_name" {
  description = "Initial database name"
  type        = string
  default     = "twitter_db"
}

variable "db_username" {
  description = "Master username for RDS"
  type        = string
  sensitive   = true
}

variable "db_password" {
  description = "Master password for RDS"
  type        = string
  sensitive   = true
}

variable "rds_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t4g.micro"
}

# ── API Gateway ──────────────────────────────────────

variable "gateway_secret" {
  description = "Secret header value that API Gateway injects into X-Gateway-Secret"
  type        = string
  sensitive   = true
}

# ── Go API General Configuration ────────────────────────

variable "db_max_conns" {
  description = "Max open connections to the database"
  type        = string
  default     = "25"
}

variable "db_min_conns" {
  description = "Min open connections to the database"
  type        = string
  default     = "0"
}

variable "db_max_conn_lifetime_minutes" {
  description = "Max lifetime of a database connection in minutes"
  type        = string
  default     = "5"
}

variable "max_media_bytes" {
  description = "Maximum size in bytes for media uploads"
  type        = string
  default     = "104857600" # 100 MB
}

variable "max_avatar_bytes" {
  description = "Maximum size in bytes for avatar uploads"
  type        = string
  default     = "5242880" # 5 MB
}

variable "cookie_domain" {
  description = "Domain for authentication cookies (leave empty for localhost)"
  type        = string
  default     = ""
}

variable "cookie_same_site" {
  description = "SameSite policy for authentication cookies"
  type        = string
  default     = "Lax"
}

variable "cookie_secure" {
  description = "Secure policy for authentication cookies (should be true in production)"
  type        = string
  default     = "true"
}

variable "token_duration_minutes" {
  description = "Duration in minutes before the JWT token expires"
  type        = string
  default     = "15"
}

variable "refresh_token_duration_days" {
  description = "Duration in days before the refresh token expires"
  type        = string
  default     = "30"
}

variable "redis_address" {
  description = "Redis address (leave empty to disable Redis)"
  type        = string
  default     = ""
}

variable "redis_password" {
  description = "Redis password (leave empty if none)"
  type        = string
  default     = ""
  sensitive   = true
}

# ── Go API Secrets ───────────────────────────────────

variable "token_symmetric_key" {
  description = "32+ char secret key for JWT token signing"
  type        = string
  sensitive   = true
}

variable "google_client_id" {
  description = "Google OAuth Web client ID"
  type        = string
}

# ── Frontend / CORS ──────────────────────────────────

variable "frontend_url" {
  description = "Frontend URL for CORS (comma-separated if multiple)"
  type        = string
  default     = "http://localhost:3000"
}

# ── GitHub / Amplify ──────────────────────────────────

variable "gh_repo_url" {
  description = "GitHub repository URL for Amplify"
  type        = string
  default     = "https://github.com/feifxi/twitter-aws"
}

variable "gh_branch" {
  description = "GitHub branch to deploy"
  type        = string
  default     = "main"
}

variable "gh_token" {
  description = "GitHub Personal Access Token (PAT)"
  type        = string
  sensitive   = true
}
