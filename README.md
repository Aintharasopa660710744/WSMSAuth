# Auth Service

Go microservice สำหรับ Authentication ใช้ JWT

## Tech Stack
- **Go** + Gin framework
- **PostgreSQL** (ผ่าน pgx/v5)
- **JWT** (golang-jwt/jwt v5) — Access + Refresh token
- **bcrypt** สำหรับ hash password

---

## API Endpoints

### Public

#### `POST /auth/register`
```json
// Request
{
  "email": "user@example.com",
  "password": "password123",
  "name": "Somchai"
}

// Response 201
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "Somchai",
    "role": "user"
  }
}
```

#### `POST /auth/login`
```json
// Request
{
  "email": "user@example.com",
  "password": "password123"
}

// Response 200 — same shape as register
```

#### `POST /auth/refresh`
```json
// Request
{
  "refresh_token": "eyJ..."
}

// Response 200 — new access_token + refresh_token
```

### Protected (ต้องใส่ Bearer Token)

#### `GET /auth/me`
```
Authorization: Bearer <access_token>
```
```json
// Response 200
{
  "user_id": "uuid",
  "email": "user@example.com",
  "role": "user"
}
```

#### `POST /auth/validate`
ใช้โดย microservice อื่น ๆ เพื่อ verify token
```
Authorization: Bearer <access_token>
```
Returns user info ถ้า token valid

---

## Token Flow

```
Login/Register
     │
     ▼
Access Token (15 นาที)  ←──── ใช้ทุก API call
Refresh Token (7 วัน)   ←──── ใช้ขอ Access Token ใหม่

POST /auth/refresh
     │
     ▼
Access Token ใหม่ + Refresh Token ใหม่
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8081` | Port ที่ listen |
| `GIN_MODE` | `debug` | `debug` หรือ `release` |
| `DATABASE_DSN` | postgres://... | PostgreSQL connection string |
| `JWT_ACCESS_SECRET` | - | Secret สำหรับ sign access token |
| `JWT_REFRESH_SECRET` | - | Secret สำหรับ sign refresh token |
| `JWT_ACCESS_EXPIRY_MINUTES` | `15` | อายุ access token (นาที) |
| `JWT_REFRESH_EXPIRY_DAYS` | `7` | อายุ refresh token (วัน) |

---

## Run

```bash
# Local dev
docker compose up

# Manual
go run ./cmd/server
```

## วิธีที่ Services อื่น Verify Token

Service อื่น (customer, admin, dashboard) สามารถ:
1. Copy `pkg/jwt` package มาใช้แล้ว validate เอง
2. หรือ call `POST /auth/validate` พร้อม Bearer token แล้วรับ user_id/role กลับมา
