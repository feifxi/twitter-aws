# ── IAM Role for EC2 (SSM + S3) ─────────────────────

resource "aws_iam_role" "ec2" {
  name = "${var.project_name}-ec2-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "ec2.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })

  tags = { Name = "${var.project_name}-ec2-role" }
}

resource "aws_iam_role_policy_attachment" "ec2_ssm" {
  role       = aws_iam_role.ec2.name
  policy_arn = "arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore"
}

resource "aws_iam_role_policy" "ec2_s3" {
  name = "${var.project_name}-ec2-s3-policy"
  role = aws_iam_role.ec2.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
        "s3:PutObject",
        "s3:DeleteObject",
        "s3:GetObject",
        "s3:ListBucket"
      ]

      Resource = [
        aws_s3_bucket.media.arn,
        "${aws_s3_bucket.media.arn}/*"
      ]
      }, {
      Effect = "Allow"
      Action = [
        "ssm:GetParameter",
        "ssm:GetParameters",
        "ssm:GetParametersByPath"
      ]
      Resource = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/chmtwt/prod/*"
    }]
  })
}


resource "aws_iam_instance_profile" "ec2" {
  name = "${var.project_name}-ec2-profile"
  role = aws_iam_role.ec2.name
}

# ── EC2 Instance ────────────────────────────────────

resource "aws_instance" "api" {
  ami                    = data.aws_ami.amazon_linux.id
  instance_type          = var.ec2_instance_type
  subnet_id              = aws_subnet.public_1.id
  vpc_security_group_ids = [aws_security_group.ec2.id]
  iam_instance_profile   = aws_iam_instance_profile.ec2.name


  root_block_device {
    volume_size = 30
    volume_type = "gp3"
  }

  metadata_options {
    http_endpoint               = "enabled"
    http_tokens                 = "required" # Enforce IMDSv2
    http_put_response_hop_limit = 2          # Allow Docker bridge network (2 hops)
  }

  tags = { Name = "${var.project_name}-ec2" }

  user_data = <<-EOF
    #!/bin/bash
    # Create app directory
    mkdir -p /home/ec2-user/app
    
    # Write docker-compose.yml
    cat << 'DOCKER_COMPOSE_EOF' > /home/ec2-user/app/docker-compose.yml
    ${file("${path.module}/../ec2/docker-compose.yml")}
    DOCKER_COMPOSE_EOF
    
    # Execute setup script
    ${file("${path.module}/../ec2/setup-ec2.sh")}
    
    # Final ownership fix
    chown -R ec2-user:ec2-user /home/ec2-user/app
  EOF
}

# ── EC2 Instance Connect Endpoint ───────────────────

resource "aws_ec2_instance_connect_endpoint" "this" {
  subnet_id          = aws_subnet.public_1.id
  security_group_ids = [aws_security_group.eice.id]

  tags = { Name = "${var.project_name}-eice" }
}
