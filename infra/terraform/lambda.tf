# ── IAM for Lambda (Consumer) ───────────────────────

resource "aws_iam_role" "lambda_embedding" {
  name = "${var.project_name}-lambda-embedding-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = { Name = "${var.project_name}-lambda-embedding-role" }
}

# Standard execution & VPC access
resource "aws_iam_role_policy_attachment" "lambda_vpc_access" {
  role       = aws_iam_role.lambda_embedding.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"
}

# SQS permissions
resource "aws_iam_role_policy" "lambda_sqs" {
  name = "${var.project_name}-lambda-sqs-policy"
  role = aws_iam_role.lambda_embedding.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Resource = aws_sqs_queue.tweet_embedding.arn
      },
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameter",
          "ssm:GetParameters",
          "ssm:GetParametersByPath"
        ]
        Resource = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/chmtwt/prod/*"
      }
    ]
  })
}

# ── Build Step: Install Python Dependencies ─────────

resource "null_resource" "lambda_pip_install" {
  triggers = {
    requirements = filemd5("${path.module}/../lambda/tweet-embedding/requirements.txt")
    source       = filemd5("${path.module}/../lambda/tweet-embedding/lambda_function.py")
  }

  provisioner "local-exec" {
    command     = "pip install -r requirements.txt -t . --upgrade --quiet"
    working_dir = "${path.module}/../lambda/tweet-embedding"
  }
}

# ── Lambda Function ─────────────────────────────────

data "archive_file" "lambda_embedding_zip" {
  depends_on = [null_resource.lambda_pip_install]

  type        = "zip"
  source_dir  = "${path.module}/../lambda/tweet-embedding"
  output_path = "${path.module}/.terraform/archive/tweet_embedding.zip"

  excludes = [
    "requirements.txt",
    "__pycache__",
    "*.pyc",
  ]
}

resource "aws_lambda_function" "tweet_embedding" {
  filename         = data.archive_file.lambda_embedding_zip.output_path
  source_code_hash = data.archive_file.lambda_embedding_zip.output_base64sha256
  function_name    = "${var.project_name}-tweet-embedding"
  role             = aws_iam_role.lambda_embedding.arn
  handler          = "lambda_function.lambda_handler"
  runtime          = "python3.12"
  timeout          = 30
  memory_size      = 256

  vpc_config {
    subnet_ids         = [aws_subnet.private_1.id, aws_subnet.private_2.id]
    security_group_ids = [aws_security_group.lambda_embedding.id]
  }

  environment {
    variables = {
      ENVIRONMENT    = "production"
      GEMINI_API_KEY = var.gemini_api_key
      DATABASE_URL   = "postgresql://${var.db_username}:${var.db_password}@${aws_db_instance.this.endpoint}/${var.db_name}?sslmode=require"
    }
  }

  tags = { Name = "${var.project_name}-tweet-embedding-lambda" }

  depends_on = [aws_instance.nat]
}

# ── SQS to Lambda Trigger ───────────────────────────

resource "aws_lambda_event_source_mapping" "sqs_to_lambda" {
  event_source_arn        = aws_sqs_queue.tweet_embedding.arn
  function_name           = aws_lambda_function.tweet_embedding.arn
  batch_size              = 10
  enabled                 = true
  function_response_types = ["ReportBatchItemFailures"]
}
