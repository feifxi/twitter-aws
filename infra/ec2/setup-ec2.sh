#!/bin/bash
# 1. อัปเดตระบบและติดตั้ง Docker & Compose
sudo dnf update -y
sudo dnf install docker -y
sudo dnf install docker-compose-plugin -y

# 2. เปิดใช้งาน Docker
sudo systemctl start docker
sudo systemctl enable docker

# 3. ให้สิทธิ์ ec2-user ใช้ Docker ได้โดยไม่ต้องพิมพ์ sudo
sudo usermod -aG docker ec2-user

# 4. สร้างโฟลเดอร์สำหรับโปรเจกต์
mkdir -p /home/ec2-user/app

echo "✅ Setup Complete! Please type 'exit' to logout and SSH again to apply docker permissions."