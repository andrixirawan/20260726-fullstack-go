# Fullstack Go REST API & Docker Setup

Repository ini berisi boilerplate aplikasi fullstack modern dengan backend Go, database PostgreSQL, serta frontend Next.js dan React Router 7. Seluruh environment dideploy menggunakan Docker Compose dengan integrasi hot-reloading untuk mempermudah proses development.

---

## 🚀 Tech Stack

### Backend (`server/`)
- **Go 1.26+** dengan standard folder structure (`cmd/`, `internal/`, `migrations/`).
- **go-chi** (`github.com/go-chi/chi/v5`) - Router HTTP yang kompatibel dengan standard library.
- **pgx** (`github.com/jackc/pgx/v5`) - Driver PostgreSQL performa tinggi dengan connection pool (`pgxpool`).
- **golang-migrate** (`github.com/golang-migrate/migrate/v4`) - Pengelolaan skema database & migrasi otomatis.
- **slog** (standard library) - Structured logging dengan format JSON.
- **golang-jwt** (`github.com/golang-jwt/jwt/v5`) - Otentikasi berbasis JWT Token.
- **bcrypt** (`golang.org/x/crypto/bcrypt`) - Hashing password yang aman dan teruji.
- **validator** (`github.com/go-playground/validator/v10`) - Validasi request body menggunakan struct tags.
- **Air** (`github.com/air-verse/air`) - Hot reloading aplikasi Go secara real-time.

### Frontend
- **Next.js 16** (`next/`) - Framework React modern dengan App Router.
- **React Router 7** (`react/`) - Framework single page app berbasis React & Vite.
- **pnpm** - Package manager yang cepat, efisien, dan hemat disk space.

### Infrastructure
- **PostgreSQL 16** - Database relational utama.
- **Docker Compose** - Menjalankan semua service (`db`, `server`, `next`, `react`) secara harmonis dalam satu virtual network.

**Port Mapping:**
| Service | Host Port | Container Port |
|---------|-----------|----------------|
| Go API | 8080 | 8080 |
| Next.js | 3001 | 3000 |
| React | 5173 | 5173 |
| PostgreSQL | 5433 | 5432 |

---

## 📁 Struktur Folder Project

```
├── .env                       # Environment variables global
├── .gitignore                 # File ignore git global
├── docker-compose.yml         # Konfigurasi Docker Compose multi-service
├── README.md                  # Dokumentasi ini
├── next/                      # Next.js Frontend scaffold (port 3000)
├── react/                     # React Router 7 Frontend scaffold (port 5173)
└── server/                    # Go Backend REST API (port 8080)
    ├── cmd/api/main.go        # Entry point aplikasi
    ├── internal/              # Core logic tertutup (handler, service, repo, dll.)
    ├── migrations/            # File SQL migrasi database
    ├── uploads/               # Direktori local penyimpanan file upload
    ├── .air.toml              # Konfigurasi Hot-Reloading Air
    └── Dockerfile             # Multi-stage build (dev & prod)
```

---

## 🛠️ Cara Menjalankan Aplikasi

### 1. Prasyarat (Prerequisites)
Pastikan Anda sudah menginstal:
- Docker & Docker Compose
- *Catatan: Jika Anda mendapatkan error permission denied saat menjalankan Docker, pastikan user Anda sudah dimasukkan ke group `docker` (`sudo usermod -aG docker $USER && newgrp docker`).*

### 2. Konfigurasi Environment (`.env`)
File `.env` sudah diinisialisasi secara otomatis di root. Jika ingin melakukan kustomisasi port atau credential database, silakan ubah isi file `.env`.

### 3. Menjalankan Semua Service
Jalankan perintah berikut di terminal root:
```bash
docker compose up --build
```
Perintah ini akan melakukan beberapa hal:
1. Menarik image PostgreSQL, Node.js Alpine, dan Golang Alpine.
2. Membangun Go container dengan hot-reload enabled menggunakan **Air**.
3. Menjalankan container frontend Next.js dan React (masing-masing melakukan auto-install dependencies menggunakan `pnpm`).
4. Menjalankan migrasi database SQL secara otomatis saat backend terhubung ke DB.

---

## 🌐 Endpoint REST API

| Method | Path | Deskripsi | Auth Required |
|--------|------|-----------|---------------|
| `GET` | `/api/v1/health` | Status kesehatan server & koneksi database | ❌ |
| `POST` | `/api/v1/auth/register` | Mendaftarkan user baru | ❌ |
| `POST` | `/api/v1/auth/login` | Login user untuk mendapatkan JWT Token | ❌ |
| `GET` | `/api/v1/auth/me` | Mendapatkan profil user yang sedang login | ✅ |
| `POST` | `/api/v1/upload` | Upload file (avatar/dokumen) | ✅ |

---

## 🧪 Alur Verifikasi & Pengujian (Testing Flow)

Anda dapat menguji API menggunakan perintah `curl` atau API client seperti Postman / Bruno.

### 1. Cek Health Server
```bash
curl http://localhost:8080/api/v1/health
```
**Response Sukses:**
```json
{
  "database": "up",
  "status": "ok"
}
```

### 2. Registrasi User Baru (`/register`)
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "developer@example.com",
    "password": "supersecretpassword",
    "full_name": "Antigravity Dev"
  }'
```
**Response Sukses (201 Created):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "f8a7e3d1...",
    "email": "developer@example.com",
    "full_name": "Antigravity Dev",
    "avatar_url": "",
    "is_active": true,
    "created_at": "...",
    "updated_at": "..."
  }
}
```

### 3. Login User (`/login`)
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "developer@example.com",
    "password": "supersecretpassword"
  }'
```
*Gunakan token JWT yang dihasilkan dari response login ini untuk langkah selanjutnya.*

### 4. Dapatkan Profil Login (`/me`)
Ganti `<TOKEN>` dengan token JWT yang Anda dapatkan dari login:
```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer <TOKEN>"
```

### 5. Upload File (`/upload`)
Ganti `<TOKEN>` dengan token JWT aktif Anda:
```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer <TOKEN>" \
  -F "file=@/path/ke/file/gambar.jpg"
```
**Response Sukses (201 Created):**
```json
{
  "filename": "f304192b-8a8b-4a55-8025-a1b70df84c31.jpg",
  "url": "/uploads/f304192b-8a8b-4a55-8025-a1b70df84c31.jpg",
  "size": 102400
}
```
*Anda dapat mengakses file yang diupload secara langsung melalui browser di: `http://localhost:8080/uploads/<filename>`*
