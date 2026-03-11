# ── S3 Bucket ────────────────────────────────────────

resource "aws_s3_bucket" "media" {
  bucket = "${var.project_name}-media-s3"

  tags = { Name = "${var.project_name}-media-s3" }
}

resource "aws_s3_bucket_public_access_block" "media" {
  bucket = aws_s3_bucket.media.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_s3_bucket_cors_configuration" "media" {
  bucket = aws_s3_bucket.media.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["PUT"]
    allowed_origins = local.frontend_origins
    expose_headers  = ["ETag"]
    max_age_seconds = 3600
  }
}

# Bucket policy is set by cloudfront.tf via OAC.
