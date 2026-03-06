# ==========================================
# Stage 1: Builder
# ==========================================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# ติดตั้งแพ็กเกจที่จำเป็นสำหรับระบบและ SSL
RUN apk update && apk add --no-cache git tzdata ca-certificates

# Copy เฉพาะไฟล์จัดการ Dependencies ก่อนเพื่อใช้ประโยชน์จาก Docker Cache
COPY go.mod go.sum ./
RUN go mod download

# Copy Source Code ทั้งหมดในโฟลเดอร์ twitter-go-api
COPY . .

# Build Go Binary แบบรีดไขมัน (ตัด Debug info ทิ้ง)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api

# ==========================================
# Stage 2: Production (Minimal & Secure)
# ==========================================
FROM alpine:latest

WORKDIR /app

# ก๊อปปี้ SSL Certs และ Timezone มาจาก Builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# ก๊อปปี้ Binary และไฟล์ Migration
COPY --from=builder /app/main .
COPY --from=builder /app/db/migration ./db/migration

# 🛡️ SECURITY: สร้าง User ธรรมดาเพื่อรันแอป (ไม่ใช้ root)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8080

CMD ["./main"]