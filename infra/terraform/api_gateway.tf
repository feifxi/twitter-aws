# ── HTTP API Gateway ─────────────────────────────────

resource "aws_apigatewayv2_api" "this" {
  name          = "${var.project_name}-api-gateway"
  protocol_type = "HTTP"

  tags = { Name = "${var.project_name}-api-gateway" }
  
  cors_configuration {
    allow_origins     = local.frontend_origins
    allow_methods     = ["GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"]
    allow_headers     = ["Origin", "Content-Type", "Accept", "Authorization"]
    allow_credentials = true
    max_age           = 3600
  }
}

# ── Integration (HTTP proxy to EC2) ──────────────────

resource "aws_apigatewayv2_integration" "ec2" {
  api_id             = aws_apigatewayv2_api.this.id
  integration_type   = "HTTP_PROXY"
  integration_method = "ANY"
  integration_uri    = "http://${aws_instance.api.public_ip}:8080/{proxy}"

  request_parameters = {
    "overwrite:header.X-Gateway-Secret" = var.gateway_secret
  }
}

# ── Catch-all route ──────────────────────────────────

resource "aws_apigatewayv2_route" "proxy" {
  api_id    = aws_apigatewayv2_api.this.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.ec2.id}"
}

# ── Default stage (auto-deploy) ──────────────────────

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.this.id
  name        = "$default"
  auto_deploy = true

  default_route_settings {
    throttling_burst_limit = 100
    throttling_rate_limit  = 50
  }

  tags = { Name = "${var.project_name}-api-stage" }
}
