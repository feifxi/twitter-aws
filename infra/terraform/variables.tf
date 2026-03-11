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

variable "ec2_key_pair_name" {
  description = "Name of an existing EC2 Key Pair (leave empty to skip)"
  type        = string
  default     = ""
}

variable "my_ip" {
  description = "Your public IP in CIDR notation for SSH access (e.g. 1.2.3.4/32)"
  type        = string
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

# ── Frontend / CORS ──────────────────────────────────

variable "frontend_url" {
  description = "Frontend URL for CORS (comma-separated if multiple)"
  type        = string
  default     = "http://localhost:3000"
}
