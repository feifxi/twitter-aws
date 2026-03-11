# ── EC2 Security Group ───────────────────────────────

resource "aws_security_group" "ec2" {
  name        = "${var.project_name}-ec2-sg"
  description = "Allow SSH + API traffic to EC2"
  vpc_id      = aws_vpc.this.id

  tags = { Name = "${var.project_name}-ec2-sg" }
}

resource "aws_vpc_security_group_ingress_rule" "ec2_ssh_myip" {
  security_group_id = aws_security_group.ec2.id
  description       = "SSH from my IP"
  ip_protocol       = "tcp"
  from_port         = 22
  to_port           = 22
  cidr_ipv4         = var.my_ip
}

resource "aws_vpc_security_group_ingress_rule" "ec2_ssh_eice" {
  security_group_id            = aws_security_group.ec2.id
  description                  = "SSH from EICE"
  ip_protocol                  = "tcp"
  from_port                    = 22
  to_port                      = 22
  referenced_security_group_id = aws_security_group.eice.id
}

resource "aws_vpc_security_group_ingress_rule" "ec2_api" {
  security_group_id = aws_security_group.ec2.id
  description       = "API traffic (protected by X-Gateway-Secret)"
  ip_protocol       = "tcp"
  from_port         = 8080
  to_port           = 8080
  cidr_ipv4         = "0.0.0.0/0"
}

resource "aws_vpc_security_group_egress_rule" "ec2_all" {
  security_group_id = aws_security_group.ec2.id
  ip_protocol       = "-1"
  cidr_ipv4         = "0.0.0.0/0"
}

# ── EICE Security Group ─────────────────────────────

resource "aws_security_group" "eice" {
  name        = "${var.project_name}-eice-sg"
  description = "EC2 Instance Connect Endpoint bridge"
  vpc_id      = aws_vpc.this.id

  tags = { Name = "${var.project_name}-eice-sg" }
}

resource "aws_vpc_security_group_egress_rule" "eice_ssh" {
  security_group_id            = aws_security_group.eice.id
  description                  = "SSH to EC2 instances"
  ip_protocol                  = "tcp"
  from_port                    = 22
  to_port                      = 22
  referenced_security_group_id = aws_security_group.ec2.id
}

# ── RDS Security Group ──────────────────────────────

resource "aws_security_group" "rds" {
  name        = "${var.project_name}-rds-sg"
  description = "Allow Postgres from EC2 only"
  vpc_id      = aws_vpc.this.id

  tags = { Name = "${var.project_name}-rds-sg" }
}

resource "aws_vpc_security_group_ingress_rule" "rds_postgres" {
  security_group_id            = aws_security_group.rds.id
  description                  = "Postgres from EC2"
  ip_protocol                  = "tcp"
  from_port                    = 5432
  to_port                      = 5432
  referenced_security_group_id = aws_security_group.ec2.id
}

resource "aws_vpc_security_group_egress_rule" "rds_all" {
  security_group_id = aws_security_group.rds.id
  ip_protocol       = "-1"
  cidr_ipv4         = "0.0.0.0/0"
}
