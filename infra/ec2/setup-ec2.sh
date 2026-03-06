#!/bin/bash

# 1. อัปเดตระบบและติดตั้ง Docker
sudo dnf update -y
sudo dnf install docker -y

# 2. ติดตั้ง Docker Compose V2 (วิธีมาตรฐานสำหรับ Amazon Linux 2023)
sudo mkdir -p /usr/local/libexec/docker/cli-plugins
sudo curl -SL "https://github.com/docker/compose/releases/latest/download/docker-compose-linux-$(uname -m)" -o /usr/local/libexec/docker/cli-plugins/docker-compose
sudo chmod +x /usr/local/libexec/docker/cli-plugins/docker-compose

# 3. เปิดใช้งานและตั้งให้ Docker ทำงานอัตโนมัติตอนเปิดเครื่อง
sudo systemctl start docker
sudo systemctl enable docker

# 4. ให้สิทธิ์ ec2-user ใช้ Docker ได้โดยไม่ต้องพิมพ์ sudo
sudo usermod -aG docker ec2-user

# 5. สร้างโฟลเดอร์สำหรับโปรเจกต์และให้ ec2-user เป็นเจ้าของ
sudo mkdir -p /home/ec2-user/app
sudo chown -R ec2-user:ec2-user /home/ec2-user/app

echo "✅ Setup Complete! Docker and Docker Compose V2 are ready."
echo "⚠️ Please type 'exit' to logout and SSH again to apply docker permissions."