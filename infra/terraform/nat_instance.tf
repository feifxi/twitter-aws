# ── NAT Instance (Cost-Effective Internet Access for Private Subnets) ──
#
# Lambda functions in private subnets need internet access to call external
# APIs (e.g., Google Gemini). A NAT Instance (~$3.50/month) is used instead
# of a NAT Gateway (~$32/month) for cost optimization.

# Find the latest Amazon Linux 2023 AMI for NAT
# Note: AL2023 doesn't have a dedicated NAT AMI, so we use a regular instance
# with IP forwarding enabled via user_data.

resource "aws_security_group" "nat" {
  name        = "${var.project_name}-nat-sg"
  description = "Allow traffic through NAT instance"
  vpc_id      = aws_vpc.this.id

  tags = { Name = "${var.project_name}-nat-sg" }
}

resource "aws_vpc_security_group_ingress_rule" "nat_from_private" {
  security_group_id = aws_security_group.nat.id
  description       = "All traffic from VPC private subnets"
  ip_protocol       = "-1"
  cidr_ipv4         = var.vpc_cidr
}

resource "aws_vpc_security_group_egress_rule" "nat_all" {
  security_group_id = aws_security_group.nat.id
  ip_protocol       = "-1"
  cidr_ipv4         = "0.0.0.0/0"
}

resource "aws_instance" "nat" {
  ami                         = data.aws_ami.amazon_linux.id
  instance_type               = "t3.nano"
  subnet_id                   = aws_subnet.public_1.id
  vpc_security_group_ids      = [aws_security_group.nat.id]
  source_dest_check           = false # Required for NAT
  associate_public_ip_address = true  # <--- [เพิ่ม] บังคับให้รับ Public IP เพื่อออกเน็ต

  user_data = <<-EOF
    #!/bin/bash
    # Enable IP forwarding
    echo 1 > /proc/sys/net/ipv4/ip_forward
    echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf

    # Get the default network interface name dynamically (usually ens5 on t3)
    INTERFACE=$(ip route | awk '/^default/ {print $5}')

    # Configure iptables for NAT
    yum install -y iptables-services
    iptables -t nat -A POSTROUTING -o $INTERFACE -s ${var.vpc_cidr} -j MASQUERADE
    iptables-save > /etc/sysconfig/iptables
    systemctl enable iptables
    systemctl start iptables
  EOF

  tags = { Name = "${var.project_name}-nat-instance" }
}

# Route private subnet traffic through NAT instance
resource "aws_route" "private_1_nat" {
  route_table_id         = aws_route_table.private_1.id
  destination_cidr_block = "0.0.0.0/0"
  network_interface_id   = aws_instance.nat.primary_network_interface_id
}

resource "aws_route" "private_2_nat" {
  route_table_id         = aws_route_table.private_2.id
  destination_cidr_block = "0.0.0.0/0"
  network_interface_id   = aws_instance.nat.primary_network_interface_id
}
