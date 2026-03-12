# ── DB Subnet Group ──────────────────────────────────

resource "aws_db_subnet_group" "this" {
  name       = "${var.project_name}-rds-subnets"
  subnet_ids = [aws_subnet.private_1.id, aws_subnet.private_2.id]

  tags = { Name = "${var.project_name}-rds-subnets" }
}

# ── RDS PostgreSQL ──────────────────────────────────

resource "aws_db_instance" "this" {
  identifier     = "${var.project_name}-pgrds"
  engine         = "postgres"
  engine_version = "16"
  instance_class = var.rds_instance_class

  db_name  = var.db_name
  username = var.db_username
  password = var.db_password

  allocated_storage = 20
  storage_type      = "gp2"

  db_subnet_group_name   = aws_db_subnet_group.this.name
  vpc_security_group_ids = [aws_security_group.rds.id]

  publicly_accessible = false
  multi_az            = false
  skip_final_snapshot = true

  # Low-cost settings (disable for learning project)
  backup_retention_period      = 0
  storage_encrypted            = false
  performance_insights_enabled = false
  monitoring_interval          = 0

  tags = { Name = "${var.project_name}-pgrds" }
}
